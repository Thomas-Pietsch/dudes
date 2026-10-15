package humangate

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// A seal is checked from git alone (F89), by whoever holds the repository: no engine state, no
// config, no network. Verify makes four checks and reports each:
//
//	signature    git tag -v accepts the tag — its own verifier: gpg's keyring, or the
//	             gpg.ssh.allowedSignersFile of the one who verifies
//	seal commit  the tag points at the human's seal commit (SealSubject)
//	sealed by    that commit carries a Sealed-by trailer naming the sealer
//	run state    the run's state note on that commit says sealed
//
// The signature covers the tag and so the commit it points at, with its trailer. The state note is
// the engine's record: git notes are not signed, and not cloned unless fetched.

// ErrNoSealTag is returned when no tag on a seal commit is reachable from HEAD.
var ErrNoSealTag = errors.New("no seal tag is reachable from HEAD (a tag on a " + SealSubject + " commit)")

// LatestSealTag returns the newest seal tag reachable from HEAD in dir: the newest by tag date of
// those that point at a seal commit. What the tag is called does not matter — dudes/seal/<product>/<run>,
// <product>-vX.Y.Z — and a tag on any other commit is none: an instance's own release tag, set after
// the seal, is not taken for it.
func LatestSealTag(dir string) (string, error) {
	if head, _ := gitx.HeadFull(dir); head == "" {
		return "", fmt.Errorf("no git history in %s (not a repository, or no commit yet)", dir)
	}
	tags, err := gitx.TagsMerged(dir)
	if err != nil {
		return "", err
	}
	for _, t := range tags {
		if t.Subject == SealSubject {
			return t.Name, nil
		}
	}
	return "", ErrNoSealTag
}

// Check is one thing Verify looked at.
type Check struct {
	Name   string
	OK     bool
	Detail string // what was found; for a failure, why it fails — may span lines
}

// Report is what Verify found about one seal tag.
type Report struct {
	Tag    string
	Commit string // the commit the tag points at; empty when the tag leads to none
	Checks []Check
}

// OK reports whether every check holds.
func (r Report) OK() bool {
	for _, c := range r.Checks {
		if !c.OK {
			return false
		}
	}
	return len(r.Checks) > 0
}

// String is the plain report: one line per check, then the verdict.
func (r Report) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "verify %s", r.Tag)
	if r.Commit != "" {
		fmt.Fprintf(&b, " (%s)", short(r.Commit))
	}
	b.WriteString("\n")
	failed := 0
	for _, c := range r.Checks {
		verdict := "ok"
		if !c.OK {
			verdict, failed = "FAIL", failed+1
		}
		// a detail of several lines continues under the first
		detail := strings.ReplaceAll(strings.TrimSpace(c.Detail), "\n", "\n"+strings.Repeat(" ", 20))
		fmt.Fprintf(&b, "  %-12s %-4s %s\n", c.Name, verdict, detail)
	}
	if r.OK() {
		fmt.Fprintf(&b, "verified: %s\n", r.Tag)
	} else {
		fmt.Fprintf(&b, "NOT verified: %s (%d failed)\n", r.Tag, failed)
	}
	return b.String()
}

// Verify checks the seal tag (its name, with or without refs/tags/) in dir. Every check runs, so the
// report says everything that is wrong at once.
func Verify(dir, tag string) Report {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "refs/tags/")
	r := Report{Tag: tag}
	commit, ok := gitx.TagCommit(dir, tag)
	if !ok {
		r.Checks = []Check{{"tag", false, fmt.Sprintf("there is no tag %q here that leads to a commit", tag)}}
		return r
	}
	r.Commit = commit
	r.Checks = []Check{
		verifySignature(dir, tag),
		verifySealCommit(dir, commit),
		verifySealer(dir, commit),
		verifyRunState(dir, commit),
	}
	return r
}

func verifySignature(dir, tag string) Check {
	said, ok, err := gitx.VerifyTag(dir, tag)
	switch {
	case err != nil:
		return Check{"signature", false, err.Error()}
	case ok:
		return Check{"signature", true, verdictLines(said)}
	case strings.Contains(said, "no signature found"):
		return Check{"signature", false, "the tag is not signed (git tag -v: no signature found)"}
	case said == "":
		return Check{"signature", false, "git tag -v refused the tag"}
	}
	return Check{"signature", false, said}
}

// verdictLines is what a good verification says: the line that names the signature, and any warning
// (a gpg key that is not certified) — not the whole of gpg's chatter.
func verdictLines(said string) string {
	var good, warn []string
	for _, l := range strings.Split(said, "\n") {
		switch l = strings.TrimSpace(l); {
		case strings.Contains(l, "Good "):
			good = append(good, l)
		case strings.Contains(l, "WARNING"):
			warn = append(warn, l)
		}
	}
	if lines := append(good, warn...); len(lines) > 0 {
		return strings.Join(lines, "\n")
	}
	if said != "" {
		return said
	}
	return "accepted by git tag -v"
}

func verifySealCommit(dir, commit string) Check {
	msg, err := gitx.Message(dir, commit)
	if err != nil {
		return Check{"seal commit", false, err.Error()}
	}
	subject, _, _ := strings.Cut(strings.TrimLeft(msg, "\r\n"), "\n")
	if subject = strings.TrimSpace(subject); subject != SealSubject {
		return Check{"seal commit", false, fmt.Sprintf("%s is %q, not a seal commit (%s)", short(commit), subject, SealSubject)}
	}
	return Check{"seal commit", true, short(commit) + " " + subject}
}

func verifySealer(dir, commit string) Check {
	values, err := gitx.Trailers(dir, commit, SealedByKey)
	if err != nil {
		return Check{"sealed by", false, err.Error()}
	}
	if len(values) == 0 {
		return Check{"sealed by", false, fmt.Sprintf("the commit has no %s trailer", SealedByKey)}
	}
	if sealer := values[0]; !strings.Contains(sealer, "<") || !strings.HasSuffix(sealer, ">") {
		return Check{"sealed by", false, fmt.Sprintf("%s is %q, not `Name <email>`", SealedByKey, sealer)}
	}
	return Check{"sealed by", true, values[0]}
}

// verifyRunState reads the run's state note on the seal commit. The next run — or a reset — saves its
// state on the newest commit, which is the seal commit until another one is made, and replaces the
// note; the notes ref keeps the save that said sealed, so that one counts, and the report says so.
func verifyRunState(dir, commit string) Check {
	versions, err := gitx.NoteVersions(dir, runstate.NotesRef, commit)
	if err != nil {
		return Check{"run state", false, err.Error()}
	}
	if len(versions) == 0 {
		return Check{"run state", false, "no run-state note on the seal commit (refs/notes/" + runstate.NotesRef + ")\n" +
			"git does not clone notes: git fetch origin 'refs/notes/dudes/*:refs/notes/dudes/*'"}
	}
	var newest *runstate.State // the note on the commit now; nil when it is no run state
	var newestErr error
	for i, text := range versions {
		var st runstate.State
		if err := json.Unmarshal([]byte(text), &st); err != nil {
			if i == 0 {
				newestErr = err
			}
			continue
		}
		if i == 0 {
			newest = &st
		}
		switch {
		case st.Status != runstate.StatusSealed:
		case i == 0:
			return Check{"run state", true, st.RunID + " sealed"}
		default:
			return Check{"run state", true, st.RunID + " sealed (a later save replaced the note on this commit; read from the notes history)"}
		}
	}
	if newest == nil {
		return Check{"run state", false, "the note on the seal commit is not a run state: " + newestErr.Error()}
	}
	return Check{"run state", false, fmt.Sprintf("the run's state note says %q, not %q", newest.Status, runstate.StatusSealed)}
}

func short(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

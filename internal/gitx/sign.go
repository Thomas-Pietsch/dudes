package gitx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// The git side of the verifiable seal (F89): who the sealer is, whether git can sign, the signed
// tag, and what git's own verifier says about one. The standard library and the git CLI only: the
// signing and the verifying are git's — gpg, ssh-keygen or gpgsm behind it, as the sealer set up.

// Config returns the value git has for key in dir, from every scope (system, global, the
// repository's own); empty when it is unset.
func Config(dir, key string) string {
	out, _, err := run(dir, "config", "--get", key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// configBool returns the boolean git has for key in dir; false when it is unset or no boolean.
func configBool(dir, key string) bool {
	out, _, err := run(dir, "config", "--type=bool", "--get", key)
	return err == nil && strings.TrimSpace(out) == "true"
}

// Identity returns who git commits as in dir — `Name <email>`, from user.name and user.email (or the
// GIT_COMMITTER_* variables), exactly what a commit made now would record as its committer. The
// error says what git lacks.
func Identity(dir string) (string, error) {
	out, _, err := run(dir, "var", "GIT_COMMITTER_IDENT")
	if err != nil {
		return "", fmt.Errorf("git has no identity here — set user.name and user.email: %w", err)
	}
	// "Name <email> <unix time> <zone>": the identity ends at the closing bracket
	ident := strings.TrimSpace(out)
	end := strings.LastIndex(ident, ">")
	if end < 0 {
		return "", fmt.Errorf("git var GIT_COMMITTER_IDENT: unexpected %q", ident)
	}
	return ident[:end+1], nil
}

// Signing says whether git can sign a tag in dir, and how.
type Signing struct {
	OK     bool   // git is set up to sign: a signing key, in a format git knows
	Format string // openpgp (git's default), ssh or x509
	Why    string // when !OK: why a tag would be unsigned, and what makes git sign it, in plain words
}

// CanSign reports whether the sealer's git can sign a tag: user.signingkey is set, and gpg.format is
// openpgp (the default), ssh or x509. A sealer whose git signs every tag (tag.gpgSign) with the
// default key of their identity counts too: that was the advice before the seal was signed by the
// engine, and git would sign their seal tag whatever the engine did. Whether the key works is git's
// to find out when it signs.
func CanSign(dir string) Signing {
	format := Config(dir, "gpg.format")
	if format == "" {
		format = "openpgp"
	}
	switch format {
	case "openpgp", "ssh", "x509":
	default:
		return Signing{Format: format, Why: fmt.Sprintf("gpg.format %q is none of openpgp, ssh, x509", format)}
	}
	if Config(dir, "user.signingkey") == "" && !configBool(dir, "tag.gpgsign") {
		return Signing{Format: format, Why: "set user.signingkey to sign seals"}
	}
	return Signing{OK: true, Format: format}
}

// SignedTag creates a signed annotated tag at HEAD: git tag -s, with user.signingkey in the format
// gpg.format names. It fails, and leaves no tag, when git cannot sign.
func SignedTag(dir, name, message string) error {
	return tagAtHead(dir, name, message, true)
}

// ResetSoft moves HEAD back to rev and leaves the index and the work tree as they are.
func ResetSoft(dir, rev string) error {
	if _, _, err := run(dir, "reset", "--soft", "-q", rev); err != nil {
		return fmt.Errorf("git reset --soft %s: %w", rev, err)
	}
	return nil
}

// TagCommit returns the commit the tag name points at (an annotated tag peeled). ok is false when
// there is no such tag, or it does not lead to a commit.
func TagCommit(dir, name string) (sha string, ok bool) {
	out, _, err := run(dir, "rev-parse", "--verify", "-q", "refs/tags/"+name+"^{commit}")
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(out), true
}

// Tag is a tag and the subject of the commit it leads to.
type Tag struct {
	Name    string
	Subject string // empty when the tag leads to no commit
}

// TagsMerged lists the tags reachable from HEAD, newest first by their date (an annotated tag's own;
// tags of the same second come in name order), each with the subject of the commit it leads to.
func TagsMerged(dir string) ([]Tag, error) {
	const us, rs = "%1f", "%1e" // git's own escapes for the unit and the record separator
	format := "%(refname:strip=2)" + us + "%(objecttype)" + us + "%(subject)" + us + "%(*objecttype)" + us + "%(*subject)" + rs
	out, _, err := run(dir, "for-each-ref", "--merged", "HEAD", "--sort=-creatordate", "--format="+format, "refs/tags")
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	var tags []Tag
	for _, rec := range strings.Split(out, "\x1e") {
		f := strings.Split(strings.TrimLeft(rec, "\r\n"), "\x1f")
		if len(f) != 5 || f[0] == "" {
			continue
		}
		t := Tag{Name: f[0]}
		switch {
		case f[1] == "commit": // a lightweight tag: the ref is the commit
			t.Subject = f[2]
		case f[1] == "tag" && f[3] == "commit": // an annotated tag: the commit it points at
			t.Subject = f[4]
		}
		tags = append(tags, t)
	}
	return tags, nil
}

// VerifyTag runs git tag -v on the tag name: git's own verifier checks the signature — gpg's keyring
// and trust, or gpg.ssh.allowedSignersFile for an SSH key — so who counts as trusted is the git
// config of the one who verifies. ok is true only when git accepts the tag; said is what the
// verifier printed, either way, in English whatever the locale. err is set only when git could not
// be run at all.
func VerifyTag(dir, name string) (said string, ok bool, err error) {
	if strings.HasPrefix(name, "-") {
		return "", false, fmt.Errorf("%q is not a tag name", name)
	}
	cmd := exec.Command("git", "tag", "-v", name)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "LC_ALL=C")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr // the verdict is on stderr; stdout is the tag's own text
	runErr := cmd.Run()
	said = strings.TrimSpace(strings.ReplaceAll(stderr.String(), "\r\n", "\n"))
	if runErr == nil {
		return said, true, nil
	}
	if _, refused := runErr.(*exec.ExitError); refused {
		return said, false, nil
	}
	return "", false, runErr
}

// Trailers returns the values of the trailers named key (matched without case) in the message of
// commit rev, as git parses them — the message's last paragraph, `Key: value` lines.
func Trailers(dir, rev, key string) ([]string, error) {
	out, _, err := run(dir, "log", "-1", "--format=%(trailers:key="+key+",valueonly,unfold)", rev)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	var values []string
	for _, v := range strings.Split(out, "\n") {
		if v = strings.TrimSpace(v); v != "" {
			values = append(values, v)
		}
	}
	return values, nil
}

// NoteVersions returns every version the note attached to commit has had under refs/notes/<ref>,
// newest first. A later save replaces the note — the run state on the seal commit is replaced by the
// next run's — and the notes ref keeps each save as a commit of its own, so the earlier versions are
// still there. None: no notes ref, or the commit never had a note; any other failure of git is an error.
func NoteVersions(dir, ref, commit string) ([]string, error) {
	notes := "refs/notes/" + ref
	if _, code, err := run(dir, "rev-parse", "--verify", "-q", notes); err != nil {
		if code == 1 { // -q: no such ref, and nothing else wrong
			return nil, nil
		}
		return nil, fmt.Errorf("git rev-parse %s: %w", notes, err)
	}
	if len(commit) < 5 {
		return nil, fmt.Errorf("%q is not a commit hash", commit)
	}
	// The note's path in the notes tree is the commit's hash, split by the fan-out git chose for the
	// size of the tree: none, one directory or two (a third level comes with millions of notes).
	paths := []string{commit, commit[:2] + "/" + commit[2:], commit[:2] + "/" + commit[2:4] + "/" + commit[4:]}
	saves, _, err := run(dir, append([]string{"log", "--format=%H", notes, "--"}, paths...)...)
	if err != nil {
		return nil, fmt.Errorf("git log %s: %w", notes, err)
	}
	var versions []string
	for _, save := range strings.Fields(saves) {
		for _, p := range paths {
			if text, _, err := run(dir, "cat-file", "blob", save+":"+p); err == nil {
				versions = append(versions, strings.TrimSpace(text))
				break
			}
		}
	}
	return versions, nil
}

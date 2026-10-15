package humangate_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/humangate"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// adaSealer is who the fixture repository's git identity is.
const adaSealer = "Ada Sealer <ada@example.com>"

func gitIn(t *testing.T, dir string, env []string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// sealRepo is a repository with one commit, kept apart from the developer's own git configuration
// (a signing key of theirs would sign the tags below).
func sealRepo(t *testing.T) string {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "")
		os.Unsetenv(k)
	}
	dir := t.TempDir()
	gitIn(t, dir, nil, "init", "-q", "-b", "main")
	gitIn(t, dir, nil, "config", "user.email", "ada@example.com")
	gitIn(t, dir, nil, "config", "user.name", "Ada Sealer")
	commit(t, dir, "a.md", "first")
	return dir
}

func commit(t *testing.T, dir, name, message string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{name}, message); err != nil {
		t.Fatal(err)
	}
	sha, _ := gitx.HeadFull(dir)
	return sha
}

// sealState is the run state a seal leaves on the seal commit.
func sealState() *runstate.State {
	return &runstate.State{RunID: "run-1", Phase: phase.Sealed, Status: runstate.StatusSealed}
}

// seal makes what dudes seal makes, except the signature: a seal commit with the message given, the
// run's state on it (none when nil) and an annotated tag. It returns the commit. Each tag gets a seal
// file of its own, so several seals fit in one repository.
func seal(t *testing.T, dir, tag, message string, st *runstate.State) string {
	t.Helper()
	sha := commit(t, dir, strings.ReplaceAll(tag, "/", "-")+"_SEAL.md", message)
	if st != nil {
		if err := runstate.Save(dir, st); err != nil {
			t.Fatal(err)
		}
	}
	if err := gitx.AnnotatedTag(dir, tag, "DUDES human seal for demo (run-1)"); err != nil {
		t.Fatal(err)
	}
	return sha
}

func check(t *testing.T, r humangate.Report, name string) humangate.Check {
	t.Helper()
	for _, c := range r.Checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no %q check in %+v", name, r.Checks)
	return humangate.Check{}
}

// A seal made as the engine makes it holds on every check but the signature, which is not there: git
// says the tag is unsigned, and the report says it in words.
func TestVerifyUnsignedSeal(t *testing.T) {
	dir := sealRepo(t)
	sha := seal(t, dir, "dudes/seal/demo/run-1", humangate.SealMessage("Ada Sealer <ada@example.com>"), sealState())

	r := humangate.Verify(dir, "dudes/seal/demo/run-1")
	if r.OK() || r.Commit != sha {
		t.Fatalf("an unsigned seal verified: %+v", r)
	}
	if c := check(t, r, "signature"); c.OK || !strings.Contains(c.Detail, "not signed") {
		t.Errorf("signature: %+v", c)
	}
	if c := check(t, r, "seal commit"); !c.OK || !strings.Contains(c.Detail, "chore(phase): seal [🔹human]") {
		t.Errorf("seal commit: %+v", c)
	}
	if c := check(t, r, "sealed by"); !c.OK || c.Detail != "Ada Sealer <ada@example.com>" {
		t.Errorf("sealed by: %+v", c)
	}
	if c := check(t, r, "run state"); !c.OK || c.Detail != "run-1 sealed" {
		t.Errorf("run state: %+v", c)
	}
	// refs/tags/ in front is the same tag
	if again := humangate.Verify(dir, "refs/tags/dudes/seal/demo/run-1"); again.Tag != r.Tag || again.Commit != sha {
		t.Errorf("refs/tags/ form: %+v", again)
	}
	out := r.String()
	for _, want := range []string{"verify dudes/seal/demo/run-1 (" + sha[:7] + ")", "signature    FAIL", "seal commit  ok", "sealed by    ok", "run state    ok", "NOT verified: dudes/seal/demo/run-1 (1 failed)"} {
		if !strings.Contains(out, want) {
			t.Errorf("the report misses %q:\n%s", want, out)
		}
	}
}

// A tag on any other commit is no seal, however it is signed: the commit must be the seal commit, name
// its sealer and carry the run's state.
func TestVerifyChecksTheCommit(t *testing.T) {
	sealer := adaSealer
	dir, n := sealRepo(t), 0
	for _, tc := range []struct {
		name    string
		message string
		state   *runstate.State
		fails   map[string]string // check → what its detail must say
	}{
		{"not a seal commit", "docs: a note", sealState(),
			map[string]string{"seal commit": "docs: a note", "sealed by": "no Sealed-by trailer"}},
		{"no sealer", humangate.SealSubject, sealState(),
			map[string]string{"sealed by": "no Sealed-by trailer"}},
		{"a sealer that is no identity", humangate.SealSubject + "\n\nSealed-by: nobody", sealState(),
			map[string]string{"sealed by": "not `Name <email>`"}},
		{"the sealer in the body, not the trailers", humangate.SealSubject + "\n\nSealed-by: " + sealer + "\n\nthe rest", sealState(),
			map[string]string{"sealed by": "no Sealed-by trailer"}},
		{"the run is not sealed", humangate.SealMessage(sealer), &runstate.State{RunID: "run-1", Phase: phase.HumanDecision, Status: runstate.StatusWaiting},
			map[string]string{"run state": `says "waiting_human"`}},
		{"no state note", humangate.SealMessage(sealer), nil,
			map[string]string{"run state": "git fetch origin 'refs/notes/dudes/*:refs/notes/dudes/*'"}},
	} {
		// one seal commit and tag per case, in one repository
		tag := "dudes/seal/demo/case-" + strconv.Itoa(n)
		n++
		seal(t, dir, tag, tc.message, tc.state)
		r := humangate.Verify(dir, tag)
		if r.OK() {
			t.Errorf("%s: verified", tc.name)
		}
		for _, name := range []string{"seal commit", "sealed by", "run state"} {
			c := check(t, r, name)
			want, fails := tc.fails[name]
			switch {
			case fails && (c.OK || !strings.Contains(c.Detail, want)):
				t.Errorf("%s: %s should fail saying %q, got %+v", tc.name, name, want, c)
			case !fails && !c.OK:
				t.Errorf("%s: %s should hold, got %+v", tc.name, name, c)
			}
		}
	}

	// a note that is no run state
	seal(t, dir, "dudes/seal/demo/no-state", humangate.SealMessage(sealer), nil)
	if err := gitx.AddNote(dir, runstate.NotesRef, "reset\n"); err != nil {
		t.Fatal(err)
	}
	if c := check(t, humangate.Verify(dir, "dudes/seal/demo/no-state"), "run state"); c.OK || !strings.Contains(c.Detail, "not a run state") {
		t.Errorf("a note that is no run state: %+v", c)
	}
}

// A tag that is not there, and a tag that is not a tag object (git cannot verify it): both stop at a
// plain refusal.
func TestVerifyRefusesWhatIsNoSeal(t *testing.T) {
	dir := sealRepo(t)
	r := humangate.Verify(dir, "dudes/seal/demo/run-9")
	if r.OK() || r.Commit != "" || len(r.Checks) != 1 || r.Checks[0].Name != "tag" || !strings.Contains(r.String(), "NOT verified: dudes/seal/demo/run-9 (1 failed)") {
		t.Fatalf("a tag that is not there: %+v\n%s", r, r)
	}
	seal(t, dir, "dudes/seal/demo/run-1", humangate.SealMessage("Ada Sealer <ada@example.com>"), sealState())
	gitIn(t, dir, nil, "tag", "dudes/seal/demo/light") // a lightweight tag, at the same commit
	r = humangate.Verify(dir, "dudes/seal/demo/light")
	if r.OK() || check(t, r, "signature").OK {
		t.Fatalf("a lightweight tag verified: %s", r)
	}
}

func TestLatestSealTag(t *testing.T) {
	dir := sealRepo(t)
	// no commit yet: no history to look in, which is not the same as no seal
	empty := t.TempDir()
	gitIn(t, empty, nil, "init", "-q")
	if _, err := humangate.LatestSealTag(empty); err == nil || errors.Is(err, humangate.ErrNoSealTag) {
		t.Fatalf("no history: %v", err)
	}

	gitIn(t, dir, nil, "tag", "-a", "v9.9.9", "-m", "not a seal")
	if _, err := humangate.LatestSealTag(dir); !errors.Is(err, humangate.ErrNoSealTag) {
		t.Fatalf("no seal tag yet: %v", err)
	}
	// sealAs makes a seal commit and an annotated tag on it, dated
	sealAs := func(name, date string) {
		t.Helper()
		commit(t, dir, strings.NewReplacer("/", "-", ".", "-").Replace(name)+".md", humangate.SealMessage(adaSealer))
		tagAs(t, dir, name, date)
	}
	sealAs("dudes/seal/demo/run-1", "2026-01-01T10:00:00Z")
	commit(t, dir, "release.md", "chore(release): changelog")
	tagAs(t, dir, "demo-v1.0.0", "2026-02-01T10:00:00Z") // newer, but on a commit that is no seal
	if got, err := humangate.LatestSealTag(dir); err != nil || got != "dudes/seal/demo/run-1" {
		t.Fatalf("got %q, %v, want the seal, not the release tag set after it", got, err)
	}

	gitIn(t, dir, nil, "checkout", "-q", "-b", "side")
	sealAs("dudes/seal/demo/on-a-branch", "2026-04-01T10:00:00Z") // the newest, but HEAD does not contain it
	gitIn(t, dir, nil, "checkout", "-q", "main")
	if got, err := humangate.LatestSealTag(dir); err != nil || got != "dudes/seal/demo/run-1" {
		t.Fatalf("got %q, %v, want the newest seal HEAD contains", got, err)
	}

	// a release seal is a seal like any other, and the newest wins
	sealAs("demo-v1.1.0", "2026-05-01T10:00:00Z")
	if got, err := humangate.LatestSealTag(dir); err != nil || got != "demo-v1.1.0" {
		t.Fatalf("got %q, %v, want the release seal", got, err)
	}
}

// tagAs makes an annotated tag at HEAD with a tag date of its own: tags made in one second have no order.
func tagAs(t *testing.T, dir, name, date string) {
	t.Helper()
	gitIn(t, dir, []string{"GIT_COMMITTER_DATE=" + date}, "tag", "-a", name, "-m", "m")
}

// The next run saves its state on the newest commit — the seal commit, until another is made — and
// that replaces the note. The seal is still a seal: the notes ref kept the save that said sealed.
func TestVerifyRunStateSurvivesLaterSaves(t *testing.T) {
	const tag = "dudes/seal/demo/run-1"
	dir := sealRepo(t)
	seal(t, dir, tag, humangate.SealMessage(adaSealer), sealState())
	state := func(dir string) humangate.Check { return check(t, humangate.Verify(dir, tag), "run state") }
	if c := state(dir); !c.OK || c.Detail != "run-1 sealed" {
		t.Fatalf("the seal as it was made: %+v", c)
	}

	// dudes reset: the note becomes a reset note
	if err := runstate.Clear(dir); err != nil {
		t.Fatal(err)
	}
	if c := state(dir); !c.OK || !strings.HasPrefix(c.Detail, "run-1 sealed (a later save replaced the note") {
		t.Fatalf("after a reset: %+v", c)
	}
	// the next run, started on the same commit
	if err := runstate.Save(dir, runstate.NewRun("run-2", "next", phase.Default())); err != nil {
		t.Fatal(err)
	}
	if c := state(dir); !c.OK || !strings.HasPrefix(c.Detail, "run-1 sealed (a later save replaced the note") {
		t.Fatalf("after the next run started: %+v", c)
	}

	// a commit that was never sealed does not become a seal by what is saved on it later
	other := sealRepo(t)
	seal(t, other, tag, humangate.SealMessage(adaSealer), &runstate.State{RunID: "run-1", Phase: phase.HumanDecision, Status: runstate.StatusWaiting})
	if err := runstate.Save(other, runstate.NewRun("run-2", "next", phase.Default())); err != nil {
		t.Fatal(err)
	}
	if c := state(other); c.OK || !strings.Contains(c.Detail, `says "running"`) {
		t.Fatalf("a run that never said sealed: %+v", c)
	}
}

func TestReportString(t *testing.T) {
	r := humangate.Report{Tag: "demo-v1.0.0", Commit: "0123456789abcdef", Checks: []humangate.Check{
		{Name: "signature", OK: true, Detail: "Good \"git\" signature\nWARNING: not certified"},
		{Name: "run state", OK: true, Detail: "run-1 sealed"},
	}}
	if !r.OK() {
		t.Fatal("every check holds")
	}
	want := "verify demo-v1.0.0 (0123456)\n" +
		"  signature    ok   Good \"git\" signature\n" +
		strings.Repeat(" ", 20) + "WARNING: not certified\n" +
		"  run state    ok   run-1 sealed\n" +
		"verified: demo-v1.0.0\n"
	if got := r.String(); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if (humangate.Report{Tag: "x"}).OK() {
		t.Fatal("a report that checked nothing verified")
	}
}

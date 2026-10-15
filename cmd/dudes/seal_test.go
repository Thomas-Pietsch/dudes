package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/bench"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/humangate"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// adaSealer is who the fixture's repositories seal as.
const adaSealer = "Ada Sealer <ada@example.com>"

// The human seal names the sealer, is signed by the sealer's key when git has one, and can be checked
// from git alone (F89). These tests run `dudes seal` and `dudes verify` as a user types them, in
// repositories whose stub run waits at the human gate; the stub lifecycle takes seconds, so it is
// built once and copied.
func TestHumanSeal(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the stub lifecycle up to the human gate")
	}
	isolateGit(t)
	fx := newSealFixture(t)
	seal := func(cfg string, args ...string) (code int, out, errOut string) {
		return captured(func() int { return run(append([]string{"-config", cfg, "seal"}, args...)) })
	}
	verify := func(cfg string, args ...string) (code int, out string) {
		code, out, errOut := captured(func() int { return run(append([]string{"-config", cfg, "verify"}, args...)) })
		if errOut != "" {
			t.Logf("verify wrote to stderr: %s", errOut)
		}
		return code, out
	}

	t.Run("an unsigned seal works when no signature is required, and carries the trailer", func(t *testing.T) {
		ws, cfg := fx.copy(t, false)
		code, out, errOut := seal(cfg)
		if code != 0 {
			t.Fatalf("seal: exit %d\n%s\n%s", code, out, errOut)
		}
		if !strings.Contains(out, "unsigned: set user.signingkey to sign seals") {
			t.Errorf("the seal should say it is unsigned, and how to sign:\n%s", out)
		}
		tag := onlyTag(t, ws)
		if !strings.HasPrefix(tag, "dudes/seal/hello/run-") {
			t.Errorf("tag %q", tag)
		}
		if kind := gitIn(t, ws, "cat-file", "-t", tag); kind != "tag" {
			t.Errorf("the seal tag is a %s, want an annotated tag", kind)
		}
		if at, head := gitIn(t, ws, "rev-parse", tag+"^{commit}"), gitIn(t, ws, "rev-parse", "HEAD"); at != head {
			t.Errorf("the tag is at %s, the seal commit is %s", at, head)
		}
		if subject := gitIn(t, ws, "log", "-1", "--format=%s"); subject != "chore(phase): seal [🔹human]" {
			t.Errorf("seal commit subject %q", subject)
		}
		if got := gitIn(t, ws, "log", "-1", "--format=%(trailers:key=Sealed-by,valueonly)"); got != adaSealer {
			t.Errorf("Sealed-by trailer %q, want %q", got, adaSealer)
		}
		if st, ok, err := runstate.Load(ws); err != nil || !ok || st.Status != runstate.StatusSealed {
			t.Errorf("run state after the seal: %+v ok=%v err=%v", st, ok, err)
		}

		// verify: the record holds, the signature is what is missing
		code, out = verify(cfg)
		if code != 1 {
			t.Fatalf("verify of an unsigned seal: exit %d\n%s", code, out)
		}
		for name, verdict := range map[string]string{"signature": "FAIL", "seal commit": "ok", "sealed by": "ok", "run state": "ok"} {
			if l := reportLine(out, name); !strings.Contains(l, " "+verdict+" ") {
				t.Errorf("%s: want %s, got %q\n%s", name, verdict, l, out)
			}
		}
		if !strings.Contains(out, "not signed") || !strings.Contains(out, "NOT verified: "+tag) {
			t.Errorf("the report should name the tag and say it is not signed:\n%s", out)
		}
	})

	t.Run("require_signature refuses an unsigned seal and leaves no commit and no tag", func(t *testing.T) {
		ws, cfg := fx.copy(t, true)
		head := gitIn(t, ws, "rev-parse", "HEAD")
		code, out, errOut := seal(cfg)
		if code != 1 {
			t.Fatalf("seal: exit %d, want 1\n%s\n%s", code, out, errOut)
		}
		for _, want := range []string{"cannot seal", "seal.require_signature", "unsigned", "user.signingkey"} {
			if !strings.Contains(errOut, want) {
				t.Errorf("the refusal should say %q:\n%s", want, errOut)
			}
		}
		if got := gitIn(t, ws, "rev-parse", "HEAD"); got != head {
			t.Errorf("a commit was made: HEAD %s, was %s", got, head)
		}
		if tags := gitIn(t, ws, "tag", "--list"); tags != "" {
			t.Errorf("a tag was made: %s", tags)
		}
		if _, err := os.Stat(filepath.Join(ws, "hello_SEAL.md")); !os.IsNotExist(err) {
			t.Errorf("the seal file was written (%v)", err)
		}
		if st, _, _ := runstate.Load(ws); st == nil || st.Phase != phase.HumanDecision || st.Status == runstate.StatusSealed {
			t.Errorf("the run is no longer waiting at the human gate: %+v", st)
		}
		// the run is still there to be sealed: the sealer turns the demand off, and seals a release —
		// a seal tag named <product>-vX.Y.Z, which verify finds as it finds the other kind
		cfg = fx.config(t, ws, false)
		if code, out, errOut := seal(cfg, "1.0.0"); code != 0 {
			t.Fatalf("seal 1.0.0 after the refusal: exit %d\n%s\n%s", code, out, errOut)
		}
		if tag := onlyTag(t, ws); tag != "hello-v1.0.0" {
			t.Fatalf("tag %q", tag)
		}
		if _, out := verify(cfg); !strings.Contains(out, "verify hello-v1.0.0 (") {
			t.Errorf("verify should pick the release tag:\n%s", out)
		}
	})

	t.Run("a taken tag name is refused before anything is made", func(t *testing.T) {
		ws, cfg := fx.copy(t, false)
		st, _, err := runstate.Load(ws)
		if err != nil {
			t.Fatal(err)
		}
		taken := gitx.SealTagName("hello", st.RunID)
		gitIn(t, ws, "tag", taken)
		head := gitIn(t, ws, "rev-parse", "HEAD")
		code, out, errOut := seal(cfg)
		if code != 1 || !strings.Contains(errOut, taken+" exists") {
			t.Fatalf("seal: exit %d, want 1 and a word about the tag\n%s\n%s", code, out, errOut)
		}
		if got := gitIn(t, ws, "rev-parse", "HEAD"); got != head {
			t.Errorf("a commit was made: HEAD %s, was %s", got, head)
		}
		if _, err := os.Stat(filepath.Join(ws, "hello_SEAL.md")); !os.IsNotExist(err) {
			t.Errorf("the seal file was written (%v)", err)
		}
		if st, _, _ := runstate.Load(ws); st == nil || st.Status == runstate.StatusSealed {
			t.Errorf("the run is sealed: %+v", st)
		}
	})

	t.Run("a run state that cannot be saved aborts the seal and takes the seal commit back", func(t *testing.T) {
		ws, cfg := fx.copy(t, false)
		head := gitIn(t, ws, "rev-parse", "HEAD")
		// a lock on the ref the state is saved in: git cannot update it until the lock is gone
		lock := filepath.Join(ws, ".git", "refs", "notes", "dudes", "run.lock")
		if err := os.MkdirAll(filepath.Dir(lock), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(lock, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		code, out, errOut := seal(cfg)
		if code != 1 || !strings.Contains(errOut, "the run state could not be saved") {
			t.Fatalf("seal: exit %d, want 1 and a word about the run state\n%s\n%s", code, out, errOut)
		}
		if got := gitIn(t, ws, "rev-parse", "HEAD"); got != head {
			t.Errorf("the seal commit is still there: HEAD %s, was %s", got, head)
		}
		if tags := gitIn(t, ws, "tag", "--list"); tags != "" {
			t.Errorf("a tag was made: %s", tags)
		}
		if err := os.Remove(lock); err != nil {
			t.Fatal(err)
		}
		if code, out, errOut := seal(cfg); code != 0 {
			t.Fatalf("seal after the lock is gone: exit %d\n%s\n%s", code, out, errOut)
		}
		_, out = verify(cfg)
		if l := reportLine(out, "run state"); !strings.Contains(l, " ok ") {
			t.Errorf("the seal that followed:\n%s", out)
		}
	})

	t.Run("a key that cannot sign aborts the seal and takes the seal commit back", func(t *testing.T) {
		ws, cfg := fx.copy(t, false)
		gitIn(t, ws, "config", "gpg.format", "ssh")
		gitIn(t, ws, "config", "user.signingkey", filepath.ToSlash(filepath.Join(t.TempDir(), "no-such-key")))
		head := gitIn(t, ws, "rev-parse", "HEAD")
		code, out, errOut := seal(cfg)
		if code != 1 || !strings.Contains(errOut, "could not sign") {
			t.Fatalf("seal: exit %d, want 1 and a word about signing\n%s\n%s", code, out, errOut)
		}
		if got := gitIn(t, ws, "rev-parse", "HEAD"); got != head {
			t.Errorf("the seal commit is still there: HEAD %s, was %s", got, head)
		}
		if tags := gitIn(t, ws, "tag", "--list"); tags != "" {
			t.Errorf("a tag was made: %s", tags)
		}
		if status := gitIn(t, ws, "status", "--porcelain", "--", "hello_SEAL.md"); status != "" {
			t.Errorf("the seal file is left behind: %s", status)
		}
		if st, _, _ := runstate.Load(ws); st == nil || st.Status == runstate.StatusSealed {
			t.Errorf("the run is sealed: %+v", st)
		}
		// the sealer fixes it — no key, so no signing — and seals once: one seal commit in the history
		gitIn(t, ws, "config", "user.signingkey", "")
		if code, out, errOut := seal(cfg); code != 0 {
			t.Fatalf("seal after the fix: exit %d\n%s\n%s", code, out, errOut)
		}
		if n := strings.Count(gitIn(t, ws, "log", "--format=%s"), "chore(phase): seal [🔹human]"); n != 1 {
			t.Errorf("%d seal commits in the history, want 1", n)
		}
	})

	t.Run("a signed seal verifies; an unsigned, a tampered or a foreign tag does not", func(t *testing.T) {
		ws, cfg := fx.copy(t, true) // a signed seal meets the demand
		key, pub := sshKey(t, "ada@example.com")
		allowed := filepath.ToSlash(filepath.Join(t.TempDir(), "allowed_signers"))
		if err := os.WriteFile(allowed, []byte("ada@example.com "+pub+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		gitIn(t, ws, "config", "gpg.format", "ssh")
		gitIn(t, ws, "config", "user.signingkey", key)
		gitIn(t, ws, "config", "gpg.ssh.allowedSignersFile", allowed)

		code, out, errOut := seal(cfg)
		if code != 0 {
			t.Fatalf("seal: exit %d\n%s\n%s", code, out, errOut)
		}
		tag := onlyTag(t, ws)
		if !strings.Contains(out, "tag: "+tag+" (signed, ssh)") || strings.Contains(out, "unsigned") {
			t.Errorf("the seal should say it signed the tag:\n%s", out)
		}
		if raw := gitRaw(t, ws, "cat-file", "tag", tag); !strings.Contains(raw, "BEGIN SSH SIGNATURE") {
			t.Fatalf("the seal tag carries no signature:\n%s", raw)
		}

		// the signed seal verifies: by default, by name, by its full ref
		for _, args := range [][]string{nil, {tag}, {"refs/tags/" + tag}} {
			code, out := verify(cfg, args...)
			if code != 0 {
				t.Fatalf("verify %v: exit %d\n%s", args, code, out)
			}
			for _, name := range []string{"signature", "seal commit", "sealed by", "run state"} {
				if l := reportLine(out, name); !strings.Contains(l, " ok ") {
					t.Errorf("verify %v: %s: %q", args, name, l)
				}
			}
			if !strings.Contains(out, "verified: "+tag) || !strings.Contains(out, adaSealer) {
				t.Errorf("verify %v should name the tag and the sealer:\n%s", args, out)
			}
			if args == nil {
				t.Logf("a verified seal:\n%s", out)
			}
		}

		// The next run, or a reset, saves its state on the newest commit — this one — and replaces the
		// note there: the seal is still a seal, and the report says where it read the state from.
		if code, out, errOut := captured(func() int { return run([]string{"-config", cfg, "reset", "-yes"}) }); code != 0 {
			t.Fatalf("reset: exit %d\n%s\n%s", code, out, errOut)
		}
		code, out = verify(cfg)
		if l := reportLine(out, "run state"); code != 0 || !strings.Contains(l, "sealed (a later save replaced the note") {
			t.Errorf("a seal after dudes reset: exit %d\n%s", code, out)
		}

		// An extra tag is named in the command and taken away again: with no tag named, verify takes the
		// newest seal tag, which would be the extra one.

		// an annotated tag at the same commit, with no signature
		const unsigned = "dudes/seal/hello/unsigned"
		gitIn(t, ws, "tag", "-a", unsigned, "-m", "DUDES human seal for hello (unsigned)")
		code, out = verify(cfg, unsigned)
		if l := reportLine(out, "signature"); code != 1 || !strings.Contains(l, "FAIL") || !strings.Contains(l, "not signed") {
			t.Errorf("an unsigned tag: exit %d\n%s", code, out)
		}
		gitIn(t, ws, "tag", "-d", unsigned)

		// the signed tag's text changed afterwards, its signature kept
		raw := gitRaw(t, ws, "cat-file", "tag", tag)
		forged := strings.Replace(raw, "DUDES human seal for hello", "DUDES human seal for another", 1)
		if forged == raw {
			t.Fatal("nothing to tamper with")
		}
		file := filepath.Join(t.TempDir(), "tag")
		if err := os.WriteFile(file, []byte(forged), 0o644); err != nil {
			t.Fatal(err)
		}
		genuine := gitIn(t, ws, "rev-parse", tag)
		gitIn(t, ws, "update-ref", "refs/tags/"+tag, gitIn(t, ws, "hash-object", "-t", "tag", "-w", file))
		for _, args := range [][]string{nil, {tag}} {
			code, out = verify(cfg, args...)
			if l := reportLine(out, "signature"); code != 1 || !strings.Contains(l, "FAIL") || !strings.Contains(out, "NOT verified: "+tag) {
				t.Errorf("a tampered tag, verify %v: exit %d\n%s", args, code, out)
			}
		}
		t.Logf("a tampered seal:\n%s", out)
		gitIn(t, ws, "update-ref", "refs/tags/"+tag, genuine)
		if code, out = verify(cfg); code != 0 {
			t.Fatalf("the genuine tag, put back: exit %d\n%s", code, out)
		}

		// a good signature on a commit that is no seal commit: the key signed it, but there is no seal there
		const elsewhere = "dudes/seal/hello/elsewhere"
		gitIn(t, ws, "tag", "-s", elsewhere, "-m", "DUDES human seal for hello (elsewhere)", "HEAD~1")
		code, out = verify(cfg, elsewhere)
		if code != 1 || !strings.Contains(reportLine(out, "signature"), " ok ") ||
			!strings.Contains(reportLine(out, "seal commit"), "FAIL") || !strings.Contains(reportLine(out, "sealed by"), "FAIL") {
			t.Errorf("a signed tag on another commit: exit %d\n%s", code, out)
		}
		gitIn(t, ws, "tag", "-d", elsewhere)

		// who counts as trusted is the verifying git's setting: a key that is no allowed signer seals nothing
		_, eve := sshKey(t, "eve@example.com")
		if err := os.WriteFile(allowed, []byte("eve@example.com "+eve+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if code, out = verify(cfg); code != 1 || !strings.Contains(reportLine(out, "signature"), "FAIL") {
			t.Errorf("a key that is not an allowed signer: exit %d\n%s", code, out)
		}
		// and a verifier with no allowed signers at all cannot tell a seal from a forgery: git says so
		gitIn(t, ws, "config", "--unset", "gpg.ssh.allowedSignersFile")
		if code, out = verify(cfg); code != 1 || !strings.Contains(reportLine(out, "signature"), "FAIL") || !strings.Contains(out, "allowedSignersFile") {
			t.Errorf("no allowed signers: exit %d\n%s", code, out)
		}
	})

	t.Run("verify says so when there is nothing to verify", func(t *testing.T) {
		ws, cfg := fx.copy(t, false)
		code, out := verify(cfg)
		if code != 1 || !strings.Contains(out, "NOT verified: no seal tag is reachable from HEAD") {
			t.Errorf("no seal yet: exit %d\n%s", code, out)
		}
		code, out = verify(cfg, "dudes/seal/hello/none")
		if code != 1 || !strings.Contains(out, "NOT verified: dudes/seal/hello/none") {
			t.Errorf("a tag that is not there: exit %d\n%s", code, out)
		}
		if code, _, errOut := captured(func() int { return run([]string{"-config", cfg, "verify", "a", "b"}) }); code != 1 || !strings.Contains(errOut, "usage: dudes verify [tag]") {
			t.Errorf("two tags: exit %d\n%s", code, errOut)
		}
		if tags := gitIn(t, ws, "tag", "--list"); tags != "" {
			t.Errorf("verify made a tag: %s", tags)
		}
	})
}

// The seal is a commit of its own even when the seal file is already as committed: the tag goes on it
// and the Sealed-by trailer is in it.
func TestCommitSealOfAnUnchangedSealFile(t *testing.T) {
	isolateGit(t)
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q", "-b", "main")
	gitIn(t, ws, "config", "user.name", "Ada Sealer")
	gitIn(t, ws, "config", "user.email", "ada@example.com")
	if err := os.WriteFile(filepath.Join(ws, "hello_SEAL.md"), []byte("# Seal\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	message := humangate.SealMessage(adaSealer)
	if before, err := commitSeal(ws, "hello_SEAL.md", message); err != nil || before != "" { // a new file, in a repository without a commit
		t.Fatalf("first seal commit: before %q, %v", before, err)
	}
	first := gitIn(t, ws, "rev-parse", "HEAD")

	if before, err := commitSeal(ws, "hello_SEAL.md", message); err != nil || before != first {
		t.Fatalf("second seal commit: before %q, want %s, %v", before, first, err)
	}
	if gitIn(t, ws, "rev-parse", "HEAD") == first {
		t.Fatal("the seal file was as committed and no seal commit was made")
	}
	if got := gitIn(t, ws, "log", "-1", "--format=%s%n%(trailers:key=Sealed-by,valueonly)"); got != "chore(phase): seal [🔹human]\n"+adaSealer {
		t.Errorf("the seal commit: %q", got)
	}
	if gitIn(t, ws, "rev-parse", "HEAD^{tree}") != gitIn(t, ws, "rev-parse", "HEAD~1^{tree}") {
		t.Error("the empty seal commit changed the tree")
	}
}

// isolateGit keeps the developer's own git configuration out of a test: a global signing key,
// commit.gpgsign or core.hooksPath of theirs would change what a seal does.
func isolateGit(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "gitconfig"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_AUTHOR_EMAIL", "GIT_COMMITTER_NAME", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "") // registers the restore; the variable itself goes for the test
		os.Unsetenv(k)
	}
}

// sealFixture is a repository whose stub run waits at the human gate.
type sealFixture struct{ template string }

func newSealFixture(t *testing.T) *sealFixture {
	t.Helper()
	dir := t.TempDir()
	if err := bench.Prepare(dir, "hello"); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := captured(func() int {
		return cmdRun(filepath.Join(dir, "dudes.json"), "hello CLI that greets", true, false, true, "")
	})
	if code != 0 {
		t.Fatalf("the stub run did not reach the human gate: exit %d\n%s\n%s", code, out, errOut)
	}
	if st, ok, err := runstate.Load(dir); err != nil || !ok || st.Phase != phase.HumanDecision {
		t.Fatalf("run state %+v ok=%v err=%v", st, ok, err)
	}
	return &sealFixture{template: dir}
}

// copy makes a repository of its own from the fixture, with Ada Sealer as its git identity, and
// returns it and its dudes.json (require is seal.require_signature).
func (f *sealFixture) copy(t *testing.T, require bool) (ws, cfg string) {
	t.Helper()
	ws = t.TempDir()
	err := filepath.WalkDir(f.template, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(f.template, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(ws, rel), 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(ws, rel), data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, ws, "config", "user.name", "Ada Sealer")
	gitIn(t, ws, "config", "user.email", "ada@example.com")
	return ws, f.config(t, ws, require)
}

// config writes the repository's dudes.json.
func (f *sealFixture) config(t *testing.T, ws string, require bool) string {
	t.Helper()
	cfg := filepath.Join(ws, "dudes.json")
	body := fmt.Sprintf(`{"workspace": %q, "product": "hello", "test_command": ["go", "test", "./..."], "seal": {"require_signature": %v}}`+"\n", filepath.ToSlash(ws), require)
	if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return cfg
}

// sshKey makes an SSH key pair without a passphrase in a folder of its own and returns the private
// key's path and the public line. The test is skipped where this machine cannot sign with SSH keys.
func sshKey(t *testing.T, comment string) (key, pub string) {
	t.Helper()
	if _, err := exec.LookPath("ssh-keygen"); err != nil {
		t.Skip("no ssh-keygen on PATH")
	}
	if !gitAtLeast(t, 2, 34) {
		t.Skip("this git cannot sign with SSH keys (2.34)")
	}
	key = filepath.ToSlash(filepath.Join(t.TempDir(), "key"))
	if out, err := exec.Command("ssh-keygen", "-q", "-t", "ed25519", "-N", "", "-C", comment, "-f", key).CombinedOutput(); err != nil {
		t.Fatalf("ssh-keygen: %v\n%s", err, out)
	}
	line, err := os.ReadFile(key + ".pub")
	if err != nil {
		t.Fatal(err)
	}
	return key, strings.TrimSpace(string(line))
}

// captured runs fn with stdout and stderr taken over, and returns what it printed.
func captured(fn func() int) (code int, stdout, stderr string) {
	oldOut, oldErr := os.Stdout, os.Stderr
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr }()
	ro, wo, _ := os.Pipe()
	re, we, _ := os.Pipe()
	os.Stdout, os.Stderr = wo, we
	var outBuf, errBuf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); _, _ = io.Copy(&outBuf, ro) }()
	go func() { defer wg.Done(); _, _ = io.Copy(&errBuf, re) }()
	code = fn()
	wo.Close()
	we.Close()
	wg.Wait()
	return code, outBuf.String(), errBuf.String()
}

// reportLine is the line of dudes verify's report for the check name.
func reportLine(report, name string) string {
	for _, l := range strings.Split(report, "\n") {
		if strings.HasPrefix(l, "  "+name+" ") {
			return l
		}
	}
	return ""
}

// onlyTag is the one tag the repository has.
func onlyTag(t *testing.T, ws string) string {
	t.Helper()
	tags := strings.Fields(gitIn(t, ws, "tag", "--list"))
	if len(tags) != 1 {
		t.Fatalf("tags %v, want exactly one", tags)
	}
	return tags[0]
}

// gitRaw is the output of git as it is, without trimming or the error stream.
func gitRaw(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, stderr.String())
	}
	return string(out)
}

// gitAtLeast reports whether the git on this machine is major.minor or newer.
func gitAtLeast(t *testing.T, major, minor int) bool {
	t.Helper()
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		t.Fatal(err)
	}
	var maj, min int
	if _, err := fmt.Sscanf(strings.TrimPrefix(strings.TrimSpace(string(out)), "git version "), "%d.%d", &maj, &min); err != nil {
		t.Fatalf("git --version: %q: %v", out, err)
	}
	return maj > major || (maj == major && min >= minor)
}

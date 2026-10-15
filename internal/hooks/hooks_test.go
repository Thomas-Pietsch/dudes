package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The scope of pre-commit (F78): staged paths stay inside the patterns of .dudes/writable; a `!`
// line excludes — the tests and hooks a red-first task wrote are read-only for the code task (F88).
func TestPreCommitScope(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "pre-commit"))
	if err != nil {
		t.Fatal(err)
	}
	const scope = "# Scope of the bound work item\n# item: X-001\n"
	code := "pkg/*\n!pkg/x_test.go\n!pkg/hook.go\n"
	cases := []struct {
		name, writable, file string
		ok                   bool
	}{
		{"no scope file", "", "anywhere/a.go", true},
		{"only comments: no path scope", scope, "anywhere/a.go", true},
		{"inside a pattern", scope + "pkg/*\n", "pkg/a.go", true},
		{"outside every pattern", scope + "pkg/*\n", "other/a.go", false},
		{"inside, not excluded", scope + code, "pkg/new.go", true},
		{"excluded test", scope + code, "pkg/x_test.go", false},
		{"excluded hook", scope + code, "pkg/hook.go", false},
		{"outside, with exclusions present", scope + code, "other/a.go", false},
		{"only exclusions: anything else", scope + "!pkg/x_test.go\n", "other/a.go", true},
		{"only exclusions: the excluded", scope + "!pkg/x_test.go\n", "pkg/x_test.go", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws := t.TempDir()
			git := func(args ...string) {
				t.Helper()
				if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
			}
			git("init", "-q")
			git("config", "user.email", "dudes@test")
			git("config", "user.name", "dudes-test")
			git("config", "core.hooksPath", filepath.Join(ws, "no-hooks"))
			if c.writable != "" {
				if err := os.MkdirAll(filepath.Join(ws, ".dudes"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(ws, ".dudes", "writable"), []byte(c.writable), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.MkdirAll(filepath.Dir(filepath.Join(ws, filepath.FromSlash(c.file))), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(ws, filepath.FromSlash(c.file)), []byte("x\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			git("add", c.file)
			cmd := exec.Command("sh", hook)
			cmd.Dir = ws
			out, err := cmd.CombinedOutput()
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
		})
	}
}

// The branch-role matrix of commit-msg (F78, DUDES §12): a spec repo's release branch takes only
// the QA Engineer and only the QR; a dev repo's release branch is the patch line, open to every
// role; development refuses a QR commit. The first instance's QA runs hit the spec-repo rule on a
// dev repo, and the role-play committed code fixes as QA Engineer to get past it, until the hook
// was fixed — checked by hand then, by nothing since.
func TestCommitMsgBranchRoles(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name         string
		spec         bool
		branch, file string
		msg          string
		ok           bool
	}{
		{"spec repo release, QA, the QR", true, "release/x-v1.0", "X_QR.md", "docs(qa): run 3 [🔹QA Engineer]", true},
		{"spec repo release, QA, code", true, "release/x-v1.0", "main.go", "fix(x): a bug [🔹QA Engineer]", false},
		{"spec repo release, PM, the QR", true, "release/x-v1.0", "X_QR.md", "docs(pm): a note [🔹Product Manager]", false},
		{"dev repo release, Backend Developer, code", false, "release/x-v1.0", "main.go", "fix(x): a bug [🔹Backend Developer]", true},
		{"dev repo release, QA, code", false, "release/x-v1.0", "main.go", "test(x): a case [🔹QA Engineer]", true},
		{"development, QA, a QR commit", false, "development", "X_QR.md", "docs(qa): QR run 3 [🔹QA Engineer]", false},
		{"development, Backend Developer, code", false, "development", "main.go", "feat(x): a route [🔹Backend Developer]", true},
		{"no role tag", false, "development", "main.go", "feat(x): a route", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws := t.TempDir()
			git := func(args ...string) {
				t.Helper()
				if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
					t.Fatalf("git %v: %v\n%s", args, err, out)
				}
			}
			write := func(rel, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(ws, rel), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q")
			git("config", "user.email", "dudes@test")
			git("config", "user.name", "dudes-test")
			git("config", "core.hooksPath", filepath.Join(ws, "no-hooks")) // the setup commit runs no hook
			base := "README.md"
			if c.spec {
				base = "X_PRD.md" // a *_PRD.md at the root makes it a spec repo
			}
			write(base, "base\n")
			git("add", base)
			git("commit", "-q", "-m", "base")
			git("checkout", "-q", "-b", c.branch)
			write(c.file, "change\n")
			git("add", c.file)
			msg := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
			if err := os.WriteFile(msg, []byte(c.msg+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("sh", hook, msg)
			cmd.Dir = ws
			out, err := cmd.CombinedOutput()
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
		})
	}
}

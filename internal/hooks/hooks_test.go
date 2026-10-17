package hooks

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		repo         string // spec, dev, one, one-merge
		branch, file string
		msg          string
		ok           bool
	}{
		{"spec repo release, QA, the QR", "spec", "release/x-v1.0", "X_QR.md", "docs(qa): run 3 [🔹QA Engineer]", true},
		{"spec repo release, QA, code", "spec", "release/x-v1.0", "main.go", "fix(x): a bug [🔹QA Engineer]", false},
		{"spec repo release, PM, the QR", "spec", "release/x-v1.0", "X_QR.md", "docs(pm): a note [🔹Product Manager]", false},
		{"dev repo release, Backend Developer, code", "dev", "release/x-v1.0", "main.go", "fix(x): a bug [🔹Backend Developer]", true},
		{"dev repo release, QA, code", "dev", "release/x-v1.0", "main.go", "test(x): a case [🔹QA Engineer]", true},
		{"development, QA, a QR commit", "dev", "development", "X_QR.md", "docs(qa): QR run 3 [🔹QA Engineer]", false},
		{"development, QA, the QR without the word in the message", "dev", "development", "X_QR.md", "docs(qa): verdict of run 1 [🔹QA Engineer]", false},
		{"development, QA, code with the word in the message", "dev", "development", "main.go", "test(x): cover the QR export [🔹QA Engineer]", true},
		{"development, PM, the QR as the quality contract", "dev", "development", "X_QR.md", "docs(x): exit set of the release [🔹Product Manager]", true},
		{"development, Backend Developer, code", "dev", "development", "main.go", "feat(x): a route [🔹Backend Developer]", true},
		{"no role tag", "dev", "development", "main.go", "feat(x): a route", false},
		// The subject prefix (§13): <type>(<scope>): <text>; git's own Merge and Revert subjects pass
		{"no type prefix", "dev", "development", "main.go", "a route [🔹Backend Developer]", false},
		{"a type without the colon", "dev", "development", "main.go", "feat a route [🔹Backend Developer]", false},
		{"a capital type", "dev", "development", "main.go", "Feat(x): a route [🔹Backend Developer]", false},
		{"no scope, a breaking mark", "dev", "development", "main.go", "feat!: a route [🔹Backend Developer]", true},
		{"a merge subject", "dev", "development", "main.go", "Merge branch 'line' [🔹Developer]", true},
		{"a revert subject", "dev", "development", "main.go", "Revert \"feat(x): a route\" [🔹Developer]", true},
		{"a free subject in a merge in progress", "one-merge", "release/x-v1.0", "main.go", "development into the line [🔹QA Engineer]", true},
		// One repo (the starter's AGENTS.md names it; the PRD sits under spec/, not at the root)
		{"one repo release, QA, the QR", "one", "release/x-v1.0", "X_QR.md", "docs(qa): run 1 [🔹QA Engineer]", true},
		{"one repo release, QA, code", "one", "release/x-v1.0", "main.go", "fix(x): a bug [🔹QA Engineer]", false},
		{"one repo release, Developer, the QR", "one", "release/x-v1.0", "X_QR.md", "docs(dev): a note [🔹Developer]", false},
		{"one repo release, Developer, code", "one", "release/x-v1.0", "main.go", "fix(x): a bug [🔹Developer]", false},
		{"one repo development, Developer, code", "one", "development", "main.go", "feat(x): a route [🔹Developer]", true},
		{"one repo development, QA, a QR commit", "one", "development", "X_QR.md", "docs(qa): QR run 1 [🔹QA Engineer]", false},
		// QA merges development into the release branch before a re-run (§6): the merge is not refused
		{"one repo release, QA, the merge of development", "one-merge", "release/x-v1.0", "main.go", "merge: development [🔹QA Engineer]", true},
		{"one repo release, Developer, the merge of development", "one-merge", "release/x-v1.0", "main.go", "merge: development [🔹Developer]", false},
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
			if c.repo == "spec" {
				base = "X_PRD.md" // a *_PRD.md at the root makes it a spec repo
			}
			write(base, "base\n")
			git("add", base)
			if strings.HasPrefix(c.repo, "one") {
				// the starter's row: the repo type, named where the roles read it
				write("AGENTS.md", "| Product | x |\n| Repo type | Spec and dev in one repo: the spec level and the dev level |\n")
				git("add", "AGENTS.md")
			}
			git("commit", "-q", "-m", "base")
			if c.repo == "one-merge" {
				git("checkout", "-q", "-b", "line")
				write(c.file, "change\n")
				git("add", c.file)
				git("commit", "-q", "-m", "line")
				git("checkout", "-q", "-b", c.branch, "HEAD~1")
				git("merge", "-q", "--no-ff", "--no-commit", "line")
			} else {
				git("checkout", "-q", "-b", c.branch)
				write(c.file, "change\n")
				git("add", c.file)
			}
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

// The role boundary of commit-msg (§14): with a `Role paths` row in the product's AGENTS.md, a
// commit whose staged paths belong to another role than its tag is refused, naming the path and
// its owner. No row, a tag naming no role of the row, and a merge in progress keep the old behaviour.
func TestCommitMsgRolePaths(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	const row = "| Role paths | `QA Engineer: spec/*_QR.md; PM: spec/*; QA Engineer: *_test.* tests/* */tests/*; QA Engineer, Developer: Makefile; Developer: *` |\n"
	const typeRow = "| Repo type | Spec and dev in one repo: the spec level and the dev level |\n"
	cases := []struct {
		name    string
		agents  string   // AGENTS.md of the repo
		files   []string // staged
		msg     string
		merge   bool
		ok      bool
		mention string // what a refusal names
	}{
		{"QA, a test file", row, []string{"pkg/a_test.go"}, "test(x): a case [🔹QA Engineer]", false, true, ""},
		{"QA, a test under tests/", row, []string{"tests/a.sh"}, "test(x): a case [🔹QA Engineer]", false, true, ""},
		{"QA, a source file", row, []string{"pkg/a.go"}, "fix(x): a bug [🔹QA Engineer]", false, false, "pkg/a.go  (owner: Developer)"},
		{"QA, a test file and a source file", row, []string{"pkg/a_test.go", "pkg/a.go"}, "fix(x): a bug [🔹QA Engineer]", false, false, "pkg/a.go"},
		{"Developer, a source file", row, []string{"pkg/a.go"}, "feat(x): a route [🔹Developer]", false, true, ""},
		{"Backend Developer, a source file", row, []string{"pkg/a.go"}, "feat(x): a route [🔹Backend Developer]", false, true, ""},
		{"Developer, a test file", row, []string{"pkg/a_test.go"}, "feat(x): a route [🔹Developer]", false, false, "pkg/a_test.go  (owner: QA Engineer)"},
		{"Developer, the spec", row, []string{"spec/X_PRD.md"}, "docs(x): a note [🔹Developer]", false, false, "spec/X_PRD.md  (owner: PM)"},
		{"PM, the spec", row, []string{"spec/X_PRD.md"}, "docs(x): a requirement [🔹PM & Software Architect]", false, true, ""},
		{"PM, the QR", row, []string{"spec/X_QR.md"}, "docs(x): a note [🔹PM & Software Architect]", false, false, "spec/X_QR.md  (owner: QA Engineer)"},
		{"QA, the QR", row, []string{"spec/X_QR.md"}, "docs(x): run 1 [🔹QA Engineer]", false, true, ""},
		{"QA, the Makefile", row, []string{"Makefile"}, "test(x): the TEST variable [🔹QA Engineer]", false, true, ""},
		{"Developer, the Makefile", row, []string{"Makefile"}, "feat(x): a goal [🔹Developer]", false, true, ""},
		{"PM, the Makefile", row, []string{"Makefile"}, "docs(x): a note [🔹PM & Software Architect]", false, false, "Makefile"},
		{"no row, QA, a source file", "", []string{"pkg/a.go"}, "fix(x): a bug [🔹QA Engineer]", false, true, ""},
		{"no row, Developer, a test file", typeRow, []string{"pkg/a_test.go"}, "feat(x): a route [🔹Developer]", false, true, ""},
		{"a tag naming no role of the row", row, []string{"pkg/a_test.go", "spec/X_PRD.md"}, "chore(x): hooks [🔹AI-driven DevOps Engineer]", false, true, ""},
		{"the human's own commit", row, []string{"pkg/a.go"}, "chore(x): a fix [🔹human]", false, true, ""},
		{"a merge in progress", row, []string{"pkg/a_test.go"}, "merge: line [🔹Developer]", true, true, ""},
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
				p := filepath.Join(ws, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q")
			git("config", "user.email", "dudes@test")
			git("config", "user.name", "dudes-test")
			git("config", "core.hooksPath", filepath.Join(ws, "no-hooks"))
			git("checkout", "-q", "-b", branchOf(c.files))
			write("README.md", "base\n")
			git("add", "README.md")
			if c.agents != "" {
				write("AGENTS.md", c.agents)
				git("add", "AGENTS.md")
			}
			git("commit", "-q", "-m", "base")
			if c.merge {
				git("checkout", "-q", "-b", "line")
				for _, f := range c.files {
					write(f, "change\n")
					git("add", f)
				}
				git("commit", "-q", "-m", "line")
				git("checkout", "-q", "development")
				git("merge", "-q", "--no-ff", "--no-commit", "line")
			} else {
				for _, f := range c.files {
					write(f, "change\n")
					git("add", f)
				}
			}
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
			if c.mention != "" && !strings.Contains(string(out), c.mention) {
				t.Fatalf("the refusal does not name %q\n%s", c.mention, out)
			}
		})
	}
}

// branchOf is the branch a commit of these files sits on: QA's QR belongs on a release branch (§12),
// anything else on development.
func branchOf(files []string) string {
	if len(files) == 0 {
		return "development"
	}
	for _, f := range files {
		if !strings.HasSuffix(f, "_QR.md") {
			return "development"
		}
	}
	return "release/x-v1.0"
}

// TestCommitMsgSpecGate: a PM commit that stages a spec document runs `dudes item check` where a
// dudes is on PATH and the repo has the profile; any finding refuses it. Everything else passes
// without running the check.
func TestCommitMsgSpecGate(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	const pm = "docs(x): a requirement [🔹PM & Software Architect]"
	cases := []struct {
		name    string
		fake    string // "" = no dudes on PATH; else the fake's output line, then its exit code after a space
		profile bool
		files   []string
		msg     string
		ok      bool
		ran     bool // the check was called
	}{
		{"findings refuse a PM spec commit", "consistency: 3 finding(s) 1", true, []string{"spec/X_PRD.md"}, pm, false, true},
		{"no findings accept it", "consistency: 0 finding(s) 0", true, []string{"spec/X_PRD.md"}, pm, true, true},
		{"no dudes on PATH accepts it", "", true, []string{"spec/X_PRD.md"}, pm, true, false},
		{"no profile accepts it", "consistency: 3 finding(s) 1", false, []string{"spec/X_PRD.md"}, pm, true, false},
		{"a check that could not run accepts it", "dudes: no config 2", true, []string{"spec/X_PRD.md"}, pm, true, true},
		{"not PM: not run", "consistency: 3 finding(s) 1", true, []string{"spec/X_PRD.md"}, "docs(x): a note [🔹Developer]", true, false},
		{"PM, no spec document: not run", "consistency: 3 finding(s) 1", true, []string{"pkg/a.go"}, pm, true, false},
		{"PM, a QR only: not run", "consistency: 3 finding(s) 1", true, []string{"spec/X_QR.md"}, pm, true, false},
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
				p := filepath.Join(ws, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			git("init", "-q")
			git("config", "user.email", "dudes@test")
			git("config", "user.name", "dudes-test")
			git("config", "core.hooksPath", filepath.Join(ws, "no-hooks"))
			git("checkout", "-q", "-b", "development")
			write("README.md", "base\n")
			git("add", "README.md")
			git("commit", "-q", "-m", "base")
			if c.profile {
				write(".dudes/profile.json", "{}\n")
			}
			for _, f := range c.files {
				write(f, "change\n")
				git("add", f)
			}
			// PATH without any real dudes, then the fake in front of it
			keep := withoutDudes()
			_, copied := cloneOf(t, hook)
			called := filepath.Join(t.TempDir(), "called")
			if c.fake != "" {
				i := strings.LastIndex(c.fake, " ")
				bin := t.TempDir()
				script := "#!/bin/sh\necho called > '" + filepath.ToSlash(called) + "'\necho '" + c.fake[:i] + "'\nexit " + c.fake[i+1:] + "\n"
				if err := os.WriteFile(filepath.Join(bin, "dudes"), []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
				keep = append([]string{bin}, keep...)
			}
			msg := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
			if err := os.WriteFile(msg, []byte(c.msg+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command("sh", copied, msg)
			cmd.Dir = ws
			cmd.Env = append(os.Environ(), "PATH="+strings.Join(keep, string(os.PathListSeparator)))
			out, err := cmd.CombinedOutput()
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
			if _, e := os.Stat(called); (e == nil) != c.ran {
				t.Fatalf("check ran=%v, want %v\n%s", e == nil, c.ran, out)
			}
			if !c.ok && !(strings.Contains(string(out), "3 open consistency finding") && strings.Contains(string(out), "dudes item check")) {
				t.Fatalf("the refusal does not name the count and the command\n%s", out)
			}
		})
	}
}

// withoutDudes is PATH's directories minus any that hold a dudes, so a test places its own.
func withoutDudes() []string {
	var keep []string
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if _, e := os.Stat(filepath.Join(d, "dudes")); e == nil {
			continue
		}
		if _, e := os.Stat(filepath.Join(d, "dudes.exe")); e == nil {
			continue
		}
		keep = append(keep, d)
	}
	return keep
}

// cloneOf copies the hook into the hooks/ of a fresh clone and returns the clone's root and the copy,
// so the engine the hook finds in bin/ beside it is the test's, never one built in this repo.
func cloneOf(t *testing.T, hook string) (clone, copied string) {
	t.Helper()
	body, err := os.ReadFile(hook)
	if err != nil {
		t.Fatal(err)
	}
	clone = t.TempDir()
	copied = filepath.Join(clone, "hooks", "commit-msg")
	if err := os.MkdirAll(filepath.Dir(copied), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copied, body, 0o755); err != nil {
		t.Fatal(err)
	}
	return clone, copied
}

// hookRepo makes a repo on branch with one commit (README.md, and AGENTS.md when agents is not empty)
// that ran no hook. git runs a command in it and returns its output; write puts a file in its work tree.
func hookRepo(t *testing.T, branch, agents string) (ws string, git func(args ...string) string, write func(rel, body string)) {
	t.Helper()
	ws = t.TempDir()
	git = func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write = func(rel, body string) {
		t.Helper()
		p := filepath.Join(ws, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	git("config", "core.hooksPath", filepath.Join(ws, "no-hooks"))
	git("checkout", "-q", "-b", branch)
	write("README.md", "base\n")
	git("add", "README.md")
	if agents != "" {
		write("AGENTS.md", agents)
		git("add", "AGENTS.md")
	}
	git("commit", "-q", "-m", "base")
	return ws, git, write
}

// runHook runs the hook at path in ws on the message and returns its output and error.
func runHook(t *testing.T, hook, ws, msg string, env []string) (string, error) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
	if err := os.WriteFile(file, []byte(msg+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("sh", hook, file)
	cmd.Dir = ws
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// The required branch of commit-msg (§12): a role of the Role paths row commits on development or a
// release/* branch, nowhere else, and the refusal names the role, the branch and §12. A repo without
// the row, a tag naming no role of it, a merge in progress and a detached HEAD are not judged.
func TestCommitMsgRoleBranch(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	const row = "| Role paths | `QA Engineer: spec/*_QR.md; Developer: *` |\n"
	cases := []struct {
		name    string
		agents  string
		branch  string // "" = a detached HEAD
		file    string
		merge   bool // a merge in progress
		msg     string
		ok      bool
		mention string
	}{
		{"a role on a feature branch", row, "feature/x", "pkg/a.go", false, "feat(x): a route [🔹Developer]", false, "role [🔹Developer] works on development, the QR on release/*; this commit is on feature/x"},
		{"QA on a feature branch", row, "feature/x", "spec/X_QR.md", false, "docs(x): run 1 [🔹QA Engineer]", false, "this commit is on feature/x"},
		{"QA, the QR on release/*", row, "release/x-v1.0", "spec/X_QR.md", false, "docs(x): run 1 [🔹QA Engineer]", true, ""},
		{"a role on development", row, "development", "pkg/a.go", false, "feat(x): a route [🔹Developer]", true, ""},
		{"a role on the patch line", row, "release/x-v1.0", "pkg/a.go", false, "fix(x): a bug [🔹Developer]", true, ""},
		{"no row, a feature branch", "", "feature/x", "pkg/a.go", false, "feat(x): a route [🔹Developer]", true, ""},
		{"a tag naming no role of the row", row, "feature/x", "pkg/a.go", false, "chore(x): hooks [🔹AI-driven DevOps Engineer]", true, ""},
		{"the human's own commit", row, "feature/x", "pkg/a.go", false, "chore(x): a fix [🔹human]", true, ""},
		{"a merge in progress", row, "feature/x", "pkg/a.go", true, "Merge branch 'line' [🔹Developer]", true, ""},
		{"a detached HEAD", row, "", "pkg/a.go", false, "feat(x): a route [🔹Developer]", true, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			branch := c.branch
			if branch == "" {
				branch = "development"
			}
			ws, git, write := hookRepo(t, branch, c.agents)
			if c.branch == "" {
				git("checkout", "-q", "--detach")
			}
			if c.merge {
				write(".git/MERGE_HEAD", git("rev-parse", "HEAD")+"\n")
			}
			write(c.file, "change\n")
			git("add", c.file)
			out, err := runHook(t, hook, ws, c.msg, nil)
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
			if !c.ok && !strings.Contains(out, "§12") {
				t.Fatalf("the refusal does not name §12\n%s", out)
			}
			if c.mention != "" && !strings.Contains(out, c.mention) {
				t.Fatalf("the refusal does not name %q\n%s", c.mention, out)
			}
		})
	}
}

// The guard rows of commit-msg (§14): a role of the Role paths row stages no change to the `Role paths`
// and `Repo type` rows of AGENTS.md nor to the .dudes-framework pin; the human does. The row the hook
// applies is the index's, so an unstaged edit cannot disarm it, and a commit that drops the row is
// held to the one HEAD has.
func TestCommitMsgGuardRows(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	const (
		product = "| Product | x |\n"
		typeRow = "| Repo type | Spec and dev in one repo: the spec level and the dev level |\n"
		row     = "| Role paths | `QA Engineer: *_test.*; PM, Developer: AGENTS.md; Developer: *` |\n"
		dev     = "feat(x): a route [🔹Developer]"
	)
	cases := []struct {
		name     string
		staged   string // AGENTS.md as staged; "" = not staged
		unstaged string // AGENTS.md as edited in the work tree and left out of the index; "" = not edited
		pin      bool   // the .dudes-framework pin is staged
		files    []string
		msg      string
		ok       bool
		mention  string
	}{
		{"a role changes the Role paths row", product + typeRow + "| Role paths | `Developer: *` |\n", "", false, nil, dev, false, "the Role paths row"},
		{"the human changes the Role paths row", product + typeRow + "| Role paths | `Developer: *` |\n", "", false, nil, "chore(x): the guard [🔹human]", true, ""},
		{"a role changes another row", "| Product | y |\n" + typeRow + row, "", false, nil, dev, true, ""},
		{"a role changes the Repo type row", product + "| Repo type | Spec and dev apart |\n" + row, "", false, nil, dev, false, "the Repo type row"},
		{"a role drops the Role paths row", product + typeRow, "", false, nil, dev, false, "the Role paths row"},
		{"a role stages the framework pin", "", "", true, nil, dev, false, ".dudes-framework"},
		{"the human stages the framework pin", "", "", true, nil, "chore(x): the pin [🔹human]", true, ""},
		{"a tag naming no role changes the row", product + typeRow + "| Role paths | `Developer: *` |\n", "", false, nil, "chore(x): hooks [🔹AI-driven DevOps Engineer]", true, ""},
		{"an unstaged edit does not disarm the row", "", product + typeRow, false, []string{"pkg/a_test.go"}, dev, false, "pkg/a_test.go  (owner: QA Engineer)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws, git, write := hookRepo(t, "development", product+typeRow+row)
			if c.staged != "" {
				write("AGENTS.md", c.staged)
				git("add", "AGENTS.md")
			}
			if c.pin {
				git("update-index", "--add", "--cacheinfo", "160000,"+git("rev-parse", "HEAD")+",.dudes-framework")
			}
			for _, f := range c.files {
				write(f, "change\n")
				git("add", f)
			}
			if c.unstaged != "" {
				write("AGENTS.md", c.unstaged)
			}
			out, err := runHook(t, hook, ws, c.msg, nil)
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
			if c.mention != "" && !strings.Contains(out, c.mention) {
				t.Fatalf("the refusal does not name %q\n%s", c.mention, out)
			}
			if !c.ok && (c.staged != "" || c.pin) && !strings.Contains(out, "[🔹human]") {
				t.Fatalf("the refusal does not say the human changes them\n%s", out)
			}
		})
	}
}

// The spec gate finds its engine (F78): the dudes in bin/ of the clone the hook runs from, with no
// dudes on PATH at all; a clone without one and a PATH without one pass without running anything.
func TestCommitMsgSpecGateClone(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}
	hook, err := filepath.Abs(filepath.Join("..", "..", "hooks", "commit-msg"))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		bin  string // the engine's file name in the clone's bin/; "" = none
		fake string // the fake's output line, then its exit code after a space
		ok   bool
		ran  bool
	}{
		{"the clone's engine finds, refuses", "dudes", "consistency: 3 finding(s) 1", false, true},
		{"the clone's engine finds none, accepts", "dudes", "consistency: 0 finding(s) 0", true, true},
		{"the clone's dudes.exe finds, refuses", "dudes.exe", "consistency: 3 finding(s) 1", false, true},
		{"no engine in the clone or on PATH", "", "", true, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ws, git, write := hookRepo(t, "development", "")
			write(".dudes/profile.json", "{}\n")
			write("spec/X_PRD.md", "change\n")
			git("add", "spec/X_PRD.md")
			clone, copied := cloneOf(t, hook)
			called := filepath.Join(t.TempDir(), "called")
			if c.bin != "" {
				i := strings.LastIndex(c.fake, " ")
				script := "#!/bin/sh\necho called > '" + filepath.ToSlash(called) + "'\necho '" + c.fake[:i] + "'\nexit " + c.fake[i+1:] + "\n"
				if err := os.MkdirAll(filepath.Join(clone, "bin"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(clone, "bin", c.bin), []byte(script), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			out, err := runHook(t, copied, ws, "docs(x): a requirement [🔹PM & Software Architect]", []string{"PATH=" + strings.Join(withoutDudes(), string(os.PathListSeparator))})
			if (err == nil) != c.ok {
				t.Fatalf("accepted=%v, want %v\n%s", err == nil, c.ok, out)
			}
			if _, e := os.Stat(called); (e == nil) != c.ran {
				t.Fatalf("check ran=%v, want %v\n%s", e == nil, c.ran, out)
			}
		})
	}
}

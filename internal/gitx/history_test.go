package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// A task's commits on an item branch carry trailers (F88): the walk returns them per commit, newest
// first, along the first parents only.
func TestFirstParents(t *testing.T) {
	dir, base := repoWithCommit(t)
	commit := func(name, msg string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitx.AddCommit(dir, []string{name}, msg); err != nil {
			t.Fatal(err)
		}
	}
	commit("a.go", "feat(X-001a): code, pass 1 [🔹engine]\n\nItem: X-001\nTask: x-001a-code\nSpec-Source: X_PRD.md abc")
	commit("b.go", "feat(X-001a): code, pass 2 [🔹engine]\n\nA body line: no trailer here.\n\nItem: X-001\nTask: x-001a-code")
	got, err := gitx.FirstParents(dir, 10)
	if err != nil || len(got) != 3 {
		t.Fatalf("walk %+v, %v", got, err)
	}
	if got[0].Subject != "feat(X-001a): code, pass 2 [🔹engine]" || got[0].Trailer("task") != "x-001a-code" || got[0].Trailer("Spec-Source") != "" {
		t.Errorf("newest %+v", got[0])
	}
	if got[1].Trailer("Spec-Source") != "X_PRD.md abc" || got[1].Trailer("Item") != "X-001" {
		t.Errorf("second %+v", got[1])
	}
	if got[2].SHA != base || len(got[2].Trailers) != 0 {
		t.Errorf("the base %+v", got[2])
	}
	if got, _ := gitx.FirstParents(dir, 1); len(got) != 1 {
		t.Errorf("n=1 gave %d commits", len(got))
	}
	if refs := gitx.Branch(dir); refs != "refs/heads/main" {
		t.Errorf("branch %q", refs)
	}
	if got, err := gitx.FirstParents(t.TempDir(), 5); err != nil || got != nil {
		t.Errorf("no repository: %v %v", got, err)
	}
	// the files a commit holds under a folder, from the root wherever git runs
	if err := os.MkdirAll(filepath.Join(dir, "dudes", "p"), 0o755); err != nil {
		t.Fatal(err)
	}
	commit("dudes/p/profile.json", "p")
	for _, rel := range []string{"dudes", "dudes/"} {
		if got, err := gitx.Files(filepath.Join(dir, "dudes"), "HEAD", rel); err != nil || strings.Join(got, " ") != "dudes/config.json dudes/p/profile.json" {
			t.Errorf("Files(%q) = %v, %v", rel, got, err)
		}
	}
	if got, err := gitx.Files(dir, base, "dudes/p"); err != nil || len(got) != 0 {
		t.Errorf("a folder the commit lacks: %v, %v", got, err)
	}
}

// Squash lands a task's passes as one commit: the same tree, the new message, the notes of the last
// pass on it — through git commit, so a hook can refuse it, and a refusal changes nothing.
func TestSquash(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir, base := repoWithCommit(t)
	hooks := filepath.Join(t.TempDir(), "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	// a commit-msg hook that wants a role tag, as hooks/commit-msg does
	if err := os.WriteFile(filepath.Join(hooks, "commit-msg"), []byte("#!/bin/sh\ngrep -q 'engine]' \"$1\" || { echo 'no role tag'; exit 1; }\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "config", "core.hooksPath", filepath.ToSlash(hooks))
	for i, body := range []string{"one\n", "two\n", "three\n"} {
		if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitx.AddCommit(dir, []string{"x.go"}, "pass "+string(rune('1'+i))+" [🔹engine]"); err != nil {
			t.Fatal(err)
		}
	}
	head := gitOut(t, dir, "rev-parse", "HEAD")
	if err := gitx.AddNote(dir, "dudes/test", "# ok: true\n"); err != nil {
		t.Fatal(err)
	}
	refs, err := gitx.NotesRefs(dir, "dudes")
	if err != nil || len(refs) != 1 || refs[0] != "dudes/test" {
		t.Fatalf("notes refs %v, %v", refs, err)
	}

	// refused by the hook: HEAD and the history stay
	if _, err := gitx.Squash(dir, base, "no tag", refs); err == nil || !strings.Contains(err.Error(), "no role tag") {
		t.Fatalf("a message the hook refuses: %v", err)
	}
	if gitOut(t, dir, "rev-parse", "HEAD") != head || gitOut(t, dir, "status", "--porcelain") != "" {
		t.Fatal("a refused squash moved HEAD or left changes")
	}

	sha, err := gitx.Squash(dir, base, "feat(X-001a): code [🔹engine]\n\nTask: x-001a-code", refs)
	if err != nil {
		t.Fatal(err)
	}
	if gitOut(t, dir, "rev-parse", "HEAD~1") != base || gitOut(t, dir, "rev-parse", "HEAD^{tree}") != gitOut(t, dir, "rev-parse", head+"^{tree}") {
		t.Error("the squash is not one commit on the base with the last pass's tree")
	}
	if got := gitOut(t, dir, "notes", "--ref", "dudes/test", "show", sha); got != "# ok: true" {
		t.Errorf("the test note did not follow: %q", got)
	}

	// nothing staged may join it, and commits that change nothing are no task
	if err := os.WriteFile(filepath.Join(dir, "staged.go"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", "staged.go")
	if _, err := gitx.Squash(dir, base, "x [🔹engine]", nil); err == nil || !strings.Contains(err.Error(), "staged") {
		t.Errorf("a staged change: %v", err)
	}
	gitOut(t, dir, "rm", "-q", "--cached", "staged.go")
	if _, err := gitx.Squash(dir, sha, "x [🔹engine]", nil); err == nil || !strings.Contains(err.Error(), "change nothing") {
		t.Errorf("no change against the base: %v", err)
	}
	if gitOut(t, dir, "rev-parse", "HEAD") != sha {
		t.Error("a refused squash moved HEAD")
	}

	// a commit another branch holds is shared
	gitOut(t, dir, "branch", "other", sha)
	if got, err := gitx.Containing(dir, sha); err != nil || strings.Join(got, " ") != "refs/heads/main refs/heads/other" {
		t.Errorf("containing %v, %v", got, err)
	}
}

// The passes of a landed task stay under a local ref: made once, listed by folder, never a branch.
func TestKeepRefs(t *testing.T) {
	dir, base := repoWithCommit(t)
	const ref = "refs/dudes/passes/x-001a/x-001a-code/1"
	if err := gitx.CreateRef(dir, ref, base); err != nil {
		t.Fatal(err)
	}
	if err := gitx.CreateRef(dir, ref, base); err == nil {
		t.Error("a ref that exists was made again")
	}
	if got, err := gitx.Refs(dir, "refs/dudes/passes/x-001a/x-001a-code"); err != nil || strings.Join(got, " ") != ref {
		t.Errorf("refs %v, %v", got, err)
	}
	if got, _ := gitx.Containing(dir, base); strings.Join(got, " ") != "refs/heads/main" {
		t.Errorf("the kept ref counts as a branch: %v", got)
	}
	for i := 0; i < 2; i++ { // the second time it is gone already
		if err := gitx.DeleteRef(dir, ref); err != nil {
			t.Errorf("delete %d: %v", i, err)
		}
	}
	if got, _ := gitx.Refs(dir, "refs/dudes/passes"); len(got) != 0 {
		t.Errorf("left %v", got)
	}
}

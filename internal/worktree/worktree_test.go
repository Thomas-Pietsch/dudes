package worktree_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/worktree"
)

func TestWant(t *testing.T) {
	if !worktree.Want(true, nil) || !worktree.Want(false, []string{"edit", "worktree"}) {
		t.Fatal("want")
	}
	if worktree.Want(false, []string{"edit"}) {
		t.Fatal("not requested")
	}
}

func TestAddRemove(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=dudes-test",
			"GIT_AUTHOR_EMAIL=dudes@test",
			"GIT_COMMITTER_NAME=dudes-test",
			"GIT_COMMITTER_EMAIL=dudes@test",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "dudes@test")
	run("config", "user.name", "dudes-test")
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{"a.md"}, "init"); err != nil {
		t.Fatal(err)
	}
	abs, branch, err := worktree.Add(dir, "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if branch != "dudes/run-1" {
		t.Fatalf("branch=%s", branch)
	}
	if _, err := os.Stat(abs); err != nil {
		t.Fatal(err)
	}
	if err := worktree.Remove(dir, "run-1"); err != nil {
		t.Fatal(err)
	}
}

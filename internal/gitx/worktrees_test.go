package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func repoWithCommit(t *testing.T) (dir, sha string) {
	t.Helper()
	dir = t.TempDir()
	gitOut(t, dir, "init", "-q", "-b", "main")
	gitOut(t, dir, "config", "user.email", "dudes@test")
	gitOut(t, dir, "config", "user.name", "dudes-test")
	gitOut(t, dir, "config", "core.autocrlf", "false")
	if err := os.MkdirAll(filepath.Join(dir, "dudes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dudes", "config.json"), []byte("{\"a\":1}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", "-A")
	gitOut(t, dir, "commit", "-qm", "first")
	return dir, gitOut(t, dir, "rev-parse", "HEAD")
}

// Item handoff reads a repository as a commit holds it (F88): refs resolve to commits, an option is no ref.
func TestCommitAndShowFile(t *testing.T) {
	dir, sha := repoWithCommit(t)
	gitOut(t, dir, "tag", "v1")
	for _, ref := range []string{"HEAD", "main", "v1", sha} {
		if got, err := gitx.Commit(dir, ref); err != nil || got != sha {
			t.Errorf("Commit(%q) = %q, %v", ref, got, err)
		}
	}
	for _, ref := range []string{"", "nope", "--upload-pack=x"} {
		if _, err := gitx.Commit(dir, ref); err == nil {
			t.Errorf("Commit(%q) resolved", ref)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "dudes", "config.json"), []byte("{\"a\":2}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if data, err := gitx.ShowFile(dir, sha, "dudes/config.json"); err != nil || string(data) != "{\"a\":1}\n" {
		t.Errorf("ShowFile = %q, %v: want the committed file, not the edited one", data, err)
	}
	if _, err := gitx.ShowFile(dir, sha, "dudes/none.json"); !os.IsNotExist(err) {
		t.Errorf("a file the commit lacks: %v", err)
	}
	if !gitx.BranchExists(dir, "main") || gitx.BranchExists(dir, "v1") {
		t.Error("BranchExists")
	}
	if _, ok := gitx.Upstream(dir); ok {
		t.Error("an upstream where none is set")
	}
}

// A worktree on a new branch that tracks nothing, one detached at a commit, and their removal.
func TestWorktrees(t *testing.T) {
	dir, sha := repoWithCommit(t)
	feature := filepath.Join(dir, ".worktrees", "f")
	if err := gitx.AddWorktree(dir, feature, "feature/f", sha); err != nil {
		t.Fatal(err)
	}
	if got := gitOut(t, feature, "branch", "--show-current"); got != "feature/f" {
		t.Errorf("branch %q", got)
	}
	if up, ok := gitx.Upstream(feature); ok {
		t.Errorf("the item's branch tracks %q", up)
	}
	pin := filepath.Join(dir, ".worktrees", "pin")
	if err := gitx.AddWorktree(dir, pin, "", sha); err != nil {
		t.Fatal(err)
	}
	if got := gitOut(t, pin, "rev-parse", "--abbrev-ref", "HEAD"); got != "HEAD" {
		t.Errorf("the pin is on branch %q, want detached", got)
	}
	if err := gitx.AddWorktree(dir, feature, "feature/g", sha); err == nil {
		t.Error("a worktree on a path that exists")
	}
	for _, p := range []string{feature, pin} {
		if err := gitx.RemoveWorktree(dir, p); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s remains", p)
		}
	}
	if err := gitx.DeleteBranch(dir, "feature/f"); err != nil || gitx.BranchExists(dir, "feature/f") {
		t.Errorf("DeleteBranch: %v", err)
	}
}

// The engine's folders stay out of git status through the repository's local exclude file — once, and
// not at all where git already ignores them.
func TestExclude(t *testing.T) {
	dir, _ := repoWithCommit(t)
	if err := os.MkdirAll(filepath.Join(dir, ".worktrees", "x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".worktrees", "x", "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if gitx.Ignored(dir, ".worktrees/x/f") || !strings.Contains(gitOut(t, dir, "status", "--porcelain"), ".worktrees") {
		t.Fatal("the fixture is not untracked")
	}
	for i := 0; i < 2; i++ {
		if err := gitx.Exclude(dir, "/.worktrees/", ".worktrees/x/f"); err != nil {
			t.Fatal(err)
		}
	}
	if !gitx.Ignored(dir, ".worktrees/x/f") || gitOut(t, dir, "status", "--porcelain") != "" {
		t.Error("the folder still shows in git status")
	}
	if data, err := os.ReadFile(filepath.Join(dir, ".git", "info", "exclude")); err != nil || strings.Count(string(data), "/.worktrees/") != 1 {
		t.Errorf("info/exclude %q (%v): want the pattern once", data, err)
	}
	other, _ := repoWithCommit(t) // ignored by the tracked .gitignore already: the local file is left alone
	if err := os.WriteFile(filepath.Join(other, ".gitignore"), []byte(".dudes/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.Exclude(other, "/.dudes/", ".dudes/x"); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(filepath.Join(other, ".git", "info", "exclude")); strings.Contains(string(data), ".dudes") {
		t.Errorf("an ignored folder was excluded again: %q", data)
	}
}

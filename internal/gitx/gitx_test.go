package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

func TestAddCommit(t *testing.T) {
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

	ok, err := gitx.IsRepo(dir)
	if err != nil || !ok {
		t.Fatalf("IsRepo: %v %v", ok, err)
	}

	path := filepath.Join(dir, "note.md")
	if err := os.WriteFile(path, []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{"note.md"}, "chore(phase): idea [🔹engine]"); err != nil {
		t.Fatal(err)
	}
	// second commit with no changes should no-op
	if err := gitx.AddCommit(dir, []string{"note.md"}, "chore(phase): idea [🔹engine]"); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "extra.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(dir, ".dudes"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, ".dudes", "state.json"), []byte("{}"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, ".aider.chat.history.md"), []byte("x\n"), 0o644)
	paths, err := gitx.ChangedPaths(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, p := range paths {
		if p == "extra.md" {
			found = true
		}
		if p == ".dudes/state.json" || p == ".dudes" || p == ".aider.chat.history.md" {
			t.Fatalf("ChangedPaths should omit .dudes and .aider*: %v", paths)
		}
	}
	if !found {
		t.Fatalf("expected extra.md in %v", paths)
	}
}

func TestSealTagName(t *testing.T) {
	got := gitx.SealTagName("demo", "run-1")
	if got != "dudes/seal/demo/run-1" {
		t.Fatalf("got %q", got)
	}
	got = gitx.SealTagName("demo!", "run 2")
	if got != "dudes/seal/demo-/run-2" {
		t.Fatalf("got %q", got)
	}
}

func TestAnnotatedTag(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
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
	tag := gitx.SealTagName("demo", "run-9")
	if err := gitx.AnnotatedTag(dir, tag, "seal msg"); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "tag", "-l", tag)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), tag) {
		t.Fatalf("tag missing: %q err=%v", out, err)
	}
	if !gitx.TagExists(dir, tag) || gitx.TagExists(dir, "demo-v9.9.9") {
		t.Fatal("TagExists: want the seal tag, not an unknown one")
	}
}

func TestNotesFollowHistory(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	if _, _, ok, err := gitx.LatestNote(dir, "dudes/test", 10); ok || err != nil {
		t.Fatalf("no commit yet: ok=%v err=%v", ok, err)
	}
	write := func(name string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitx.AddCommit(dir, []string{name}, "add "+name); err != nil {
			t.Fatal(err)
		}
	}
	write("a.md")
	first, _ := gitx.HeadFull(dir)
	if err := gitx.AddNote(dir, "dudes/test", "# ok: false\n"); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddNote(dir, "dudes/test", "# ok: true\nPASS TestA\n"); err != nil {
		t.Fatal(err) // replaces the note on the same commit
	}
	write("b.md") // a commit without a note: the newest note in history still counts
	c, text, ok, err := gitx.LatestNote(dir, "dudes/test", 10)
	if err != nil || !ok || c != first || text != "# ok: true\nPASS TestA\n" {
		t.Fatalf("got commit=%s ok=%v err=%v text=%q", c, ok, err, text)
	}
}

func TestReleaseTagName(t *testing.T) {
	if tag, err := gitx.ReleaseTagName("sample_app", "1.1.0"); err != nil || tag != "sample_app-v1.1.0" {
		t.Fatalf("got %q, %v", tag, err)
	}
	for _, bad := range []string{"1.1", "v1.1.0", "1.1.0-rc1", ""} {
		if _, err := gitx.ReleaseTagName("p", bad); err == nil {
			t.Fatalf("%q accepted, want X.Y.Z only", bad)
		}
	}
}

// Discard restores a file HEAD knows and removes a new one, staged or not.
func TestDiscard(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "dudes@test"}, {"config", "user.name", "dudes-test"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(name, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module keep\n")
	if err := gitx.AddCommit(dir, []string{"go.mod"}, "init"); err != nil {
		t.Fatal(err)
	}
	write("go.mod", "module broken\n")
	write("db/new.go", "package db\n")
	if out, err := exec.Command("git", "-C", dir, "add", "db/new.go").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	write("db/new.go", "package db // changed after staging\n")
	write("loose.txt", "x\n")
	if err := gitx.Discard(dir, []string{"go.mod", "db/new.go", "loose.txt"}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "go.mod")); strings.ReplaceAll(string(b), "\r", "") != "module keep\n" { // autocrlf on Windows
		t.Fatalf("go.mod not restored: %q", b)
	}
	if paths, err := gitx.ChangedPaths(dir); err != nil || len(paths) != 0 {
		t.Fatalf("left changes %v %v", paths, err)
	}
}

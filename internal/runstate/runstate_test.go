package runstate_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// repo returns a git repository without commits.
func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{{"init"}, {"config", "user.email", "dudes@test"}, {"config", "user.name", "dudes-test"}} {
		git(t, dir, args...)
	}
	return dir
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

func TestSaveLoadRoundTripStartsTheRecord(t *testing.T) {
	dir := repo(t)
	st := runstate.NewRun("run-1", "build a todo app", phase.Default())
	if err := runstate.Save(dir, st); err != nil {
		t.Fatal(err) // no commit yet: Save makes the start commit
	}
	got, ok, err := runstate.Load(dir)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.RunID != "run-1" || got.Phase != phase.Idea || got.Prompt != "build a todo app" || got.Status != runstate.StatusRunning {
		t.Fatalf("got %+v", got)
	}
	if head, _ := gitx.Head(dir); head == "" {
		t.Fatal("want the start commit")
	}
}

func TestLoadMissing(t *testing.T) {
	if _, ok, err := runstate.Load(repo(t)); err != nil || ok {
		t.Fatalf("repo without run: ok=%v err=%v", ok, err)
	}
	if _, ok, err := runstate.Load(t.TempDir()); err != nil || ok {
		t.Fatalf("no repo: ok=%v err=%v", ok, err)
	}
}

func TestSaveOutsideGitFails(t *testing.T) {
	if err := runstate.Save(t.TempDir(), runstate.NewRun("run-x", "", phase.Default())); err == nil {
		t.Fatal("want an error: run state lives in git")
	}
}

func TestClearAndStartAgain(t *testing.T) {
	dir := repo(t)
	if err := runstate.Save(dir, runstate.NewRun("run-x", "p", phase.Default())); err != nil {
		t.Fatal(err)
	}
	if err := runstate.Clear(dir); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := runstate.Load(dir); err != nil || ok {
		t.Fatalf("after clear ok=%v err=%v", ok, err)
	}
	if err := runstate.Clear(dir); err != nil {
		t.Fatal(err) // idempotent
	}
	if err := runstate.Save(dir, runstate.NewRun("run-y", "", phase.Default())); err != nil {
		t.Fatal(err)
	}
	if got, ok, _ := runstate.Load(dir); !ok || got.RunID != "run-y" {
		t.Fatalf("new run after clear: %+v", got)
	}
}

// The state follows the history: reset the branch and the run state goes back with it.
func TestStateFollowsHistory(t *testing.T) {
	dir := repo(t)
	reg := phase.Default()
	st := runstate.NewRun("run-2", "", reg)
	if err := runstate.Save(dir, st); err != nil {
		t.Fatal(err)
	}
	start, _ := gitx.HeadFull(dir)

	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{"a.md"}, "work"); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := runstate.Load(dir); got.Phase != phase.Idea {
		t.Fatalf("a commit without a new note keeps the state: %s", got.Phase)
	}
	next, _ := reg.Next(st.Phase)
	st.Phase = next
	if err := runstate.Save(dir, st); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := runstate.Load(dir); got.Phase != phase.Strategy {
		t.Fatalf("advanced: %s", got.Phase)
	}

	git(t, dir, "reset", "-q", "--hard", start)
	if got, _, _ := runstate.Load(dir); got.Phase != phase.Idea {
		t.Fatalf("after reset to the start: %s", got.Phase)
	}
}

func TestMigratesStateJSON(t *testing.T) {
	dir := repo(t)
	legacy := filepath.Join(dir, ".dudes", "state.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(`{"run_id":"run-old","phase":"spec","status":"running"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok, err := runstate.Load(dir)
	if err != nil || !ok || got.RunID != "run-old" || got.Phase != phase.Spec {
		t.Fatalf("migrate: %+v ok=%v err=%v", got, ok, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("state.json should be gone after the migration")
	}
	if got, ok, _ := runstate.Load(dir); !ok || got.RunID != "run-old" {
		t.Fatal("migrated state should load from git")
	}
}

func TestApproveWaitingDecision(t *testing.T) {
	st := runstate.NewRun("run-d", "", phase.Default())
	st.Status = runstate.StatusWaitingDecision
	st.TaskID = "scope-call"
	st.Approve("scope-call", "keep field evidence")
	if st.Status != runstate.StatusRunning || !st.HasApproved("scope-call") {
		t.Fatalf("%+v", st)
	}
	st.MarkCompleted("scope-call")
	if !st.HasCompleted("scope-call") {
		t.Fatal("completed")
	}
}

// A finished task sent back runs again; only its worker gets the redo note, and the note clears
// when the task completes again.
func TestReopen(t *testing.T) {
	st := &runstate.State{Status: runstate.StatusWaitingDecision, Feedback: "run-level"}
	st.MarkCompleted("test-add")
	st.Approve("test-add", "")
	st.Reopen("test-add", "the fixture path is relative to tests/")
	if st.HasCompleted("test-add") || st.HasApproved("test-add") || st.Status != runstate.StatusRunning {
		t.Fatalf("not reopened: %+v", st)
	}
	if fb := st.TaskFeedback(); !strings.Contains(fb, "run-level") || !strings.Contains(fb, "relative to tests/") {
		t.Fatalf("feedback %q", fb)
	}
	st.TaskID = "impl-add"
	if fb := st.TaskFeedback(); strings.Contains(fb, "relative to tests/") {
		t.Fatalf("another task got the redo note: %q", fb)
	}
	st.MarkCompleted("impl-add")
	st.Reopen("impl-add", "print the line") // a second task sent back keeps the first one's note
	st.TaskID = "test-add"
	if fb := st.TaskFeedback(); !strings.Contains(fb, "relative to tests/") {
		t.Fatalf("the first note was lost: %q", fb)
	}
	st.MarkCompleted("test-add")
	if _, ok := st.Redo["test-add"]; ok || st.Redo["impl-add"] != "print the line" {
		t.Fatalf("redo notes after completion: %+v", st.Redo)
	}
}

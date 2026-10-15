package events_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestAppendRead(t *testing.T) {
	dir := t.TempDir()
	if err := events.Append(dir, events.Event{RunID: "r1", Type: "phase_leave", Phase: phase.Idea, Detail: "ok"}); err != nil {
		t.Fatal(err)
	}
	evs, err := events.ReadAll(dir)
	if err != nil || len(evs) != 1 || evs[0].Type != "phase_leave" {
		t.Fatalf("%v %v", evs, err)
	}
}

// A worker call's cost and an answer's decider read back as they were written (F88); an event without
// them stays as before, with no empty fields.
func TestUsageAndDeciderRoundTrip(t *testing.T) {
	dir := t.TempDir()
	u := &events.Usage{Task: "bk-001a-code", Worker: "aider", Model: "ollama/m", PromptTokens: 9800, CompletionTokens: 1200, Counted: "aider", DurationMs: 95000}
	for _, ev := range []events.Event{
		{RunID: "r", Type: "worker", Phase: phase.Implementation, Usage: u},
		{RunID: "r", Type: "redo", Detail: "bk-001a-code a swallowed error", By: "dudes-test <dudes@test>", Role: "Developer"},
		{RunID: "r", Type: "task_enter", Detail: "bk-001a-code"},
	} {
		if err := events.Append(dir, ev); err != nil {
			t.Fatal(err)
		}
	}
	evs, err := events.ReadAll(dir)
	if err != nil || len(evs) != 3 {
		t.Fatalf("%v %v", evs, err)
	}
	if evs[0].Usage == nil || *evs[0].Usage != *u {
		t.Errorf("usage %+v", evs[0].Usage)
	}
	if evs[1].By != "dudes-test <dudes@test>" || evs[1].Role != "Developer" {
		t.Errorf("decider %+v", evs[1])
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(events.Path())))
	if err != nil {
		t.Fatal(err)
	}
	last := strings.TrimSpace(string(data))
	last = last[strings.LastIndex(last, "\n")+1:]
	for _, absent := range []string{"usage", "by", "role"} {
		if strings.Contains(last, `"`+absent+`"`) {
			t.Errorf("a plain event carries %q: %s", absent, last)
		}
	}
}

// The ledger's events — a worker call, a decide, a redo — go into git as well, as lines of the note on
// HEAD; the other types stay in the file alone. Read back from git they come oldest first, over every
// note, and the log file need not be there (F88).
func TestLedgerInGit(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	git("commit", "-q", "--allow-empty", "-m", "start")
	t0 := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	for i, ev := range []events.Event{
		{Type: "task_enter", Detail: "bk-001a-tests"},
		{Type: "worker", Detail: "bk-001a-tests aider", Usage: &events.Usage{Task: "bk-001a-tests", Worker: "aider", PromptTokens: 2100, CompletionTokens: 302, Counted: "aider", DurationMs: 61000}},
		{Type: "redo", Detail: "bk-001a-tests the path is relative", By: "dudes-test <dudes@test>", Role: "Developer"},
	} {
		ev.RunID, ev.Time = "r", t0.Add(time.Duration(i)*time.Minute)
		if err := events.Append(dir, ev); err != nil {
			t.Fatal(err)
		}
	}
	note := git("notes", "--ref", events.LedgerRef, "show", "HEAD")
	if !strings.Contains(note, `"type":"worker"`) || !strings.Contains(note, `"type":"redo"`) || strings.Contains(note, "task_enter") {
		t.Errorf("the note on HEAD:\n%s", note)
	}
	git("commit", "-q", "--allow-empty", "-m", "a pass")
	if err := events.Append(dir, events.Event{RunID: "r", Type: "decide", Detail: "bk-001a-tests ", By: "dudes-test <dudes@test>", Time: t0.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, filepath.FromSlash(events.Path()))); err != nil {
		t.Fatal(err)
	}
	if events.Logged(dir) {
		t.Fatal("Logged without the file")
	}
	got, err := events.Ledger(dir)
	if err != nil || len(got) != 3 {
		t.Fatalf("ledger %+v %v", got, err)
	}
	if got[0].Type != "worker" || got[0].Usage == nil || got[0].Usage.PromptTokens != 2100 || got[1].Type != "redo" || got[1].Role != "Developer" || got[2].Type != "decide" {
		t.Errorf("ledger, oldest first: %+v", got)
	}
}

func TestTail(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < 5; i++ {
		if err := events.Append(dir, events.Event{RunID: "r", Type: "x", Detail: string(rune('a' + i))}); err != nil {
			t.Fatal(err)
		}
	}
	tail, err := events.Tail(dir, 2)
	if err != nil || len(tail) != 2 {
		t.Fatalf("%v %v", tail, err)
	}
	if tail[0].Detail != "d" || tail[1].Detail != "e" {
		t.Fatalf("got %+v", tail)
	}
}

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// dudes status -json sums the worker events of the run it reports: how many calls, and the prompt and
// completion tokens of their usage. A worker event of another run and an event of another type add
// nothing; a run without calls reports zeros.
func TestStatusSumsWorkerCalls(t *testing.T) {
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	gitIn(t, ws, "config", "core.autocrlf", "false")
	cfgPath := filepath.Join(ws, "dudes.json")
	cfg := `{"workspace": "` + filepath.ToSlash(ws) + `", "product": "hello"}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	st := &runstate.State{RunID: "run-a", Phase: phase.ID("implementation"), Status: runstate.StatusRunning}
	if err := runstate.Save(ws, st); err != nil {
		t.Fatal(err)
	}
	status := func() map[string]any {
		t.Helper()
		code, out, errOut := captured(func() int { return run([]string{"-config", cfgPath, "status", "-json"}) })
		var doc map[string]any
		if err := json.Unmarshal([]byte(out), &doc); code != 0 || err != nil {
			t.Fatalf("status -json: exit %d, %v\n%s\n%s", code, err, out, errOut)
		}
		run, _ := doc["run"].(map[string]any)
		return run
	}
	want := func(run map[string]any, calls, prompt, completion float64) {
		t.Helper()
		got := [3]any{run["worker_calls"], run["prompt_tokens"], run["completion_tokens"]}
		if got != [3]any{calls, prompt, completion} {
			t.Errorf("worker_calls, prompt_tokens, completion_tokens = %v, want %v %v %v", got, calls, prompt, completion)
		}
	}

	want(status(), 0, 0, 0)

	worker := func(runID string, prompt, completion int) events.Event {
		return events.Event{RunID: runID, Type: "worker", Usage: &events.Usage{Worker: "aider", PromptTokens: prompt, CompletionTokens: completion}}
	}
	for _, ev := range []events.Event{
		worker("run-a", 1200, 300),
		worker("run-b", 7000, 900),
		{RunID: "run-a", Type: "step", Detail: "not a worker call"},
		worker("run-a", 800, 50),
	} {
		if err := events.Append(ws, ev); err != nil {
			t.Fatal(err)
		}
	}
	want(status(), 2, 2000, 350)
}

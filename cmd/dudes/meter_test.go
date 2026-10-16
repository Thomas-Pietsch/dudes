package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/events"
)

// meterRepo is a git workspace with one commit and a config for a fake Ollama answering reply with
// the token counts 21 and 5.
func meterRepo(t *testing.T, reply string, extra string) (ws, cfgPath string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"response": reply, "done": true, "prompt_eval_count": 21, "eval_count": 5})
	}))
	t.Cleanup(srv.Close)
	ws = t.TempDir()
	gitIn(t, ws, "init", "-q")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	gitIn(t, ws, "config", "core.autocrlf", "false")
	if err := os.WriteFile(filepath.Join(ws, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, ws, "add", "-A")
	gitIn(t, ws, "commit", "-qm", "init")
	cfgPath = filepath.Join(ws, "dudes.json")
	cfg := `{"workspace": "` + filepath.ToSlash(ws) + `", "product": "hello", ` + extra + `"ollama": {"base_url": "` + srv.URL + `", "model": "m"}}`
	if err := os.WriteFile(cfgPath, []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	return ws, cfgPath
}

// workerCalls are the worker events of the run ledger.
func workerCalls(t *testing.T, ws string) []events.Event {
	t.Helper()
	evs, err := events.ReadAll(ws)
	if err != nil {
		t.Fatal(err)
	}
	var calls []events.Event
	for _, ev := range evs {
		if ev.Type == "worker" {
			calls = append(calls, ev)
		}
	}
	return calls
}

// A call outside a run still lands in the ledger, with its tokens and time and with no run, phase or
// task; and no reply is kept, since there is no task it belongs to.
func TestRecordCallWithoutRun(t *testing.T) {
	ws := t.TempDir()
	u := events.Usage{Worker: "ollama", Model: "m", PromptTokens: 9, CompletionTokens: 3, Counted: "ollama"}
	recordCall(config.Config{Workspace: ws}, nil, u, 1500*time.Millisecond, "a reply", nil)
	calls := workerCalls(t, ws)
	if len(calls) != 1 {
		t.Fatalf("worker events %+v", calls)
	}
	ev := calls[0]
	if ev.RunID != "" || ev.Phase != "" || ev.Usage == nil {
		t.Fatalf("event %+v", ev)
	}
	if got := *ev.Usage; got.Task != "" || got.PromptTokens != 9 || got.CompletionTokens != 3 || got.DurationMs != 1500 || got.Counted != "ollama" {
		t.Errorf("usage %+v", got)
	}
	if !strings.HasPrefix(ev.Detail, "ollama: 9 prompt + 3 completion tokens (ollama), ") {
		t.Errorf("detail %q", ev.Detail)
	}
	if _, err := os.Stat(filepath.Join(ws, ".dudes", "replies")); err == nil {
		t.Error("a reply was kept without a task")
	}
	// A failed call is recorded too.
	recordCall(config.Config{Workspace: ws}, nil, events.Usage{Worker: "ollama", Model: "m"}, time.Second, "", errors.New("down"))
	if calls = workerCalls(t, ws); len(calls) != 2 || !strings.HasSuffix(calls[1].Detail, ", failed") {
		t.Errorf("failed call %+v", calls)
	}
}

// `dudes do -pick` calls the text model: that call is a worker event with Ollama's counts.
func TestIntakePickIsMetered(t *testing.T) {
	ws, cfgPath := meterRepo(t, "NONE", "")
	if code, out, errOut := captured(func() int { return cmdDo(cfgPath, []string{"-pick", "sort", "versions"}) }); code != 2 {
		t.Fatalf("pick: %d\n%s%s", code, out, errOut)
	}
	calls := workerCalls(t, ws)
	if len(calls) != 1 || calls[0].Usage == nil {
		t.Fatalf("worker events %+v", calls)
	}
	if u := *calls[0].Usage; u.Worker != "ollama" || u.Model != "m" || u.PromptTokens != 21 || u.CompletionTokens != 5 || u.Counted != "ollama" {
		t.Errorf("usage %+v", u)
	}
}

// `dudes test -fix` calls the text model on a failing test: that call is a worker event with Ollama's
// counts, and the suggestion is still written.
func TestTestFixIsMetered(t *testing.T) {
	ws, cfgPath := meterRepo(t, "Change the return value.", `"test_command": ["sh", "-c", "echo boom; exit 1"], `)
	if code, out, errOut := captured(func() int { return cmdTest(cfgPath, true) }); code != 1 {
		t.Fatalf("test -fix: %d\n%s%s", code, out, errOut)
	}
	if fix, err := os.ReadFile(filepath.Join(ws, "hello_FIX.md")); err != nil || !strings.Contains(string(fix), "Change the return value.") {
		t.Fatalf("the suggestion: %q %v", fix, err)
	}
	calls := workerCalls(t, ws)
	if len(calls) != 1 || calls[0].Usage == nil {
		t.Fatalf("worker events %+v", calls)
	}
	if u := *calls[0].Usage; u.Worker != "ollama" || u.PromptTokens != 21 || u.CompletionTokens != 5 || u.Counted != "ollama" {
		t.Errorf("usage %+v", u)
	}
}

// Package events appends a durable run event log (F29-lite). The ledger's events — a worker call's
// cost, an answer to a pause — go into git as well (F88).
package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const relPath = ".dudes/events.jsonl"

// LedgerRef holds the ledger in git: each ledger event as a line of the note on the commit HEAD was
// on when it happened. A notes ref is one for every worktree of a repository, so the ledger outlives
// the workspace's .dudes/ — a slice's worktree, removed once the slice lands (F88).
const LedgerRef = "dudes/ledger"

// inLedger are the event types that go into git too: what a slice cost and who decided what.
var inLedger = map[string]bool{"worker": true, "decide": true, "redo": true}

// Event is one auditable engine action (not LLM chat).
type Event struct {
	Time   time.Time `json:"time"`
	RunID  string    `json:"run_id,omitempty"`
	Type   string    `json:"type"`
	Phase  phase.ID  `json:"phase,omitempty"`
	Detail string    `json:"detail,omitempty"`
	// By and Role name who answered a pause (dudes decide, dudes redo): the git identity of whoever
	// ran the command, and the role they named with -by (F88).
	By   string `json:"by,omitempty"`
	Role string `json:"role,omitempty"`
	// Usage is what a worker call cost (type "worker", F88).
	Usage *Usage `json:"usage,omitempty"`
}

// Usage is one worker call's cost in local tokens and time (F88).
type Usage struct {
	Task   string `json:"task,omitempty"` // the contract task the call worked on
	Worker string `json:"worker"`         // aider, pi, ollama
	Model  string `json:"model,omitempty"`
	// PromptTokens and CompletionTokens are the tokens of the prompt and of the reply, as Counted
	// says.
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	// Counted is whose count the tokens are: "ollama" — the model host's own (prompt_eval_count,
	// eval_count); "aider" — Aider's report, the sum of its "Tokens: … sent, … received." lines, which
	// it rounds to 0.1k under 10k and to 1k above. Empty: the worker reported no count, the tokens are 0.
	Counted    string `json:"counted,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// Path returns the log relative path.
func Path() string { return relPath }

// Append writes one JSON line under the workspace; a ledger event also to the note on HEAD under
// LedgerRef, when the workspace has a commit to attach it to.
func Append(workspaceRoot string, ev Event) error {
	if ev.Time.IsZero() {
		ev.Time = time.Now().UTC()
	}
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	line, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(abs, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return err
	}
	if inLedger[ev.Type] {
		if head, _ := gitx.Head(workspaceRoot); head != "" {
			if err := gitx.AppendNote(workspaceRoot, LedgerRef, string(line)+"\n"); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
	}
	return nil
}

// Logged reports whether the workspace's log file exists.
func Logged(workspaceRoot string) bool {
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return false
	}
	_, err = os.Stat(abs)
	return err == nil
}

// Ledger reads the ledger back from git: the ledger events of every run in the workspace's
// repository, from every note under LedgerRef whichever worktree or branch wrote it, oldest first.
func Ledger(workspaceRoot string) ([]Event, error) {
	notes, err := gitx.AllNotes(workspaceRoot, LedgerRef)
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, n := range notes {
		for _, line := range strings.Split(n.Text, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			var ev Event
			if err := json.Unmarshal([]byte(line), &ev); err != nil {
				return out, fmt.Errorf("ledger parse (note on %s): %w", n.Commit, err)
			}
			out = append(out, ev)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Time.Before(out[j].Time) })
	return out, nil
}

// Log is a convenience wrapper that ignores empty workspace errors softly in CLI.
func Log(workspaceRoot, runID, typ string, ph phase.ID, detail string) {
	_ = Append(workspaceRoot, Event{
		RunID:  runID,
		Type:   typ,
		Phase:  ph,
		Detail: detail,
	})
}

// ReadAll loads events (best-effort; for tests/status).
func ReadAll(workspaceRoot string) ([]Event, error) {
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Event
	for _, line := range splitLines(string(data)) {
		if line == "" {
			continue
		}
		var ev Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			return out, fmt.Errorf("events parse: %w", err)
		}
		out = append(out, ev)
	}
	return out, nil
}

// Tail returns the last n events (n<=0 → 20).
func Tail(workspaceRoot string, n int) ([]Event, error) {
	all, err := ReadAll(workspaceRoot)
	if err != nil {
		return nil, err
	}
	return Last(all, n), nil
}

// Last returns the last n of evs (n<=0 → 20).
func Last(evs []Event, n int) []Event {
	if n <= 0 {
		n = 20
	}
	if len(evs) <= n {
		return evs
	}
	return evs[len(evs)-n:]
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

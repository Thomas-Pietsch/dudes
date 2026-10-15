// Package events appends a durable run event log (F29-lite).
package events

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const relPath = ".dudes/events.jsonl"

// Event is one auditable engine action (not LLM chat).
type Event struct {
	Time   time.Time `json:"time"`
	RunID  string    `json:"run_id,omitempty"`
	Type   string    `json:"type"`
	Phase  phase.ID  `json:"phase,omitempty"`
	Detail string    `json:"detail,omitempty"`
}

// Path returns the log relative path.
func Path() string { return relPath }

// Append writes one JSON line under the workspace.
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
	return nil
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
	if n <= 0 {
		n = 20
	}
	all, err := ReadAll(workspaceRoot)
	if err != nil {
		return nil, err
	}
	if len(all) <= n {
		return all, nil
	}
	return all[len(all)-n:], nil
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

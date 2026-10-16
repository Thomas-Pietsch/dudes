package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/ollama"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/redact"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// What a worker call cost and what it said (F88). Each call is a `worker` event with its local tokens
// and its duration, so a slice's cost is read from the run instead of reconstructed; the reply is kept
// under .dudes/replies/ for the task's review packet.

// metered wraps a code worker's call so that recordCall records it: the tokens Aider reports
// (aider.Usage; Pi reports none), the duration, the reply.
func metered(cfg config.Config, st *runstate.State, worker, model string, invoke func(string) (string, error)) func(string) (string, error) {
	return func(msg string) (string, error) {
		start := time.Now()
		out, err := invoke(msg)
		u := events.Usage{Worker: worker, Model: model}
		if worker == "aider" {
			if sent, received, ok := aider.Usage(out); ok {
				u.PromptTokens, u.CompletionTokens, u.Counted = sent, received, "aider"
			}
		}
		recordCall(cfg, st, u, time.Since(start), out, err)
		return out, err
	}
}

// ollamaUsage is a text worker call's usage: Ollama's own counts, when it answered.
func ollamaUsage(model string, u ollama.Usage, err error) events.Usage {
	ev := events.Usage{Worker: "ollama", Model: model, PromptTokens: u.PromptTokens, CompletionTokens: u.CompletionTokens}
	if err == nil || u.PromptTokens+u.CompletionTokens > 0 {
		ev.Counted = "ollama"
	}
	return ev
}

// recordCall records one worker call: a worker event with u, the task and the duration; and, when
// the call succeeded, its reply as the task's last (replyRel). Without a run (st nil) the event has
// no run, phase or task and no reply is kept: the call is still measured.
func recordCall(cfg config.Config, st *runstate.State, u events.Usage, took time.Duration, reply string, callErr error) {
	var runID string
	var ph phase.ID
	if st != nil {
		u.Task, runID, ph = st.TaskID, st.RunID, st.Phase
	}
	u.DurationMs = took.Milliseconds()
	tokens := "no token count"
	if u.Counted != "" {
		tokens = fmt.Sprintf("%d prompt + %d completion tokens (%s)", u.PromptTokens, u.CompletionTokens, u.Counted)
	}
	detail := strings.TrimSpace(u.Task+" "+u.Worker) + ": " + tokens + ", " + took.Round(time.Second).String()
	if callErr != nil {
		detail += ", failed"
	}
	if err := events.Append(cfg.Workspace, events.Event{RunID: runID, Type: "worker", Phase: ph, Detail: detail, Usage: &u}); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: the worker call is not recorded in full: %v\n", err)
	}
	if st == nil || callErr != nil || strings.TrimSpace(reply) == "" {
		return
	}
	p := filepath.Join(cfg.Workspace, filepath.FromSlash(replyRel(nonzero(st.TaskID, string(st.Phase)))))
	err := os.MkdirAll(filepath.Dir(p), 0o755)
	if err == nil {
		err = os.WriteFile(p, []byte(redact.String(reply)), 0o644)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: the worker's reply was not kept for the review: %v\n", err)
	}
}

// replyRel is where the worker's last reply on task is kept, relative to the workspace: engine
// state, never committed.
func replyRel(task string) string {
	var b strings.Builder
	for _, r := range task {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return ".dudes/replies/" + b.String() + ".md"
}

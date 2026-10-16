package pi

import (
	"encoding/json"
	"strings"
)

// Stats is what one Pi call did, read from its --mode json event stream (docs/json.md): the
// authoritative counts are the assistant messages' usage and the tool executions. The stream is
// the only record — Pi's exit code is 0 even when the last reply is an error or hit the token limit.
type Stats struct {
	ToolCalls  int    // tool_execution_start events: a run with none did no work
	ToolErrors int    // tool_execution_end events with isError
	Sent       int    // prompt tokens, summed over the assistant messages
	Received   int    // completion tokens, summed (reasoning is already in them)
	Counted    bool   // Pi reported usage at all (Ollama's /v1 may not stream it)
	Stop       string // the last assistant message's stopReason: stop, toolUse, length, error, aborted
}

type event struct {
	Type    string `json:"type"`
	IsError bool   `json:"isError"`
	Message struct {
		Role       string `json:"role"`
		StopReason string `json:"stopReason"`
		Usage      struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		} `json:"usage"`
	} `json:"message"`
}

// Parse reads Pi's JSONL output; a line that is no JSON event (a diagnostic) is skipped.
func Parse(out string) Stats {
	var s Stats
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var e event
		if json.Unmarshal([]byte(line), &e) != nil {
			continue
		}
		switch e.Type {
		case "tool_execution_start":
			s.ToolCalls++
		case "tool_execution_end":
			if e.IsError {
				s.ToolErrors++
			}
		case "message_end":
			if e.Message.Role != "assistant" {
				continue
			}
			s.Sent += e.Message.Usage.Input
			s.Received += e.Message.Usage.Output
			if e.Message.Usage.Input+e.Message.Usage.Output > 0 {
				s.Counted = true
			}
			s.Stop = e.Message.StopReason
		}
	}
	return s
}

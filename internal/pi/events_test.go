package pi

import "testing"

const stream = `{"type":"session","version":3,"id":"x","cwd":"/w"}
{"type":"agent_start"}
{"type":"message_end","message":{"role":"user","content":"task"}}
{"type":"message_update","usage":{"input":900,"output":1},"assistantMessageEvent":{"type":"text_delta","delta":"x"}}
{"type":"message_end","message":{"role":"assistant","stopReason":"toolUse","usage":{"input":1000,"output":40,"totalTokens":1040}}}
{"type":"tool_execution_start","toolCallId":"a","toolName":"read","args":{"path":"go.mod"}}
{"type":"tool_execution_end","toolCallId":"a","toolName":"read","isError":false,"durationMs":3}
{"type":"tool_execution_start","toolCallId":"b","toolName":"edit","args":{}}
{"type":"tool_execution_end","toolCallId":"b","toolName":"edit","isError":true}
not json, a diagnostic
{"type":"message_end","message":{"role":"assistant","stopReason":"stop","usage":{"input":1200,"output":60}}}
{"type":"agent_settled","aborted":false}
`

func TestParseCountsToolCallsAndAssistantUsage(t *testing.T) {
	s := Parse(stream)
	want := Stats{ToolCalls: 2, ToolErrors: 1, Sent: 2200, Received: 100, Counted: true, Stop: "stop"}
	if s != want {
		t.Fatalf("got %+v, want %+v", s, want)
	}
}

// An error reply exits 0 in JSON mode: the stream is the only record, and a run with no tool call
// shows as zero.
func TestParseErrorReplyWithoutToolCalls(t *testing.T) {
	s := Parse(`{"type":"message_end","message":{"role":"assistant","stopReason":"error","usage":{"input":0,"output":0}}}` + "\n")
	if s.ToolCalls != 0 || s.Counted || s.Stop != "error" {
		t.Fatalf("got %+v", s)
	}
}

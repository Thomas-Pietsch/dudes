package events_test

import (
	"testing"

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

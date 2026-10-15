package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/bench"
)

// The hermetic stub lifecycle (F53) runs in `make ci`: every phase from idea to the human gate
// with stub workers, its gates and git notes, no model and no network. It checks the engine's
// plumbing, not a model (TREND D1; models are measured by the bench ladder, F82).
func TestStubLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("stub lifecycle runs the whole phase machine")
	}
	dir := t.TempDir()
	if err := bench.Prepare(dir, "hello"); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	code := cmdRun(filepath.Join(dir, "dudes.json"), "hello CLI that greets", true, false, true, "")
	rep, err := bench.Summarize(dir, "hello", started, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 || !rep.OK {
		t.Fatalf("stub lifecycle: exit %d, report %+v", code, rep)
	}
}

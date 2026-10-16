package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
)

// Both code workers get the same files: the ones to edit and the ones to read only. A name in the
// task's tools list that is no existing file (a Pi tool name) is not one.
func TestHandoverSplitsEditAndReadFiles(t *testing.T) {
	ws := t.TempDir()
	for _, f := range []string{"go.mod", "cfg.go"} {
		if err := os.WriteFile(filepath.Join(ws, f), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	msg := aider.Message{Tests: []string{"a_test.go"}, ReadOnlyTests: []string{"ac8_test.go"}}
	files, reads := handover(ws, []string{"cfg.go", "go.mod", "read", "gone.go"}, msg, []string{"b.go", "a_test.go"}, []string{"ref.go"}, nil)
	if want := []string{"a_test.go", "b.go"}; !slices.Equal(files, want) {
		t.Fatalf("files=%v want %v", files, want)
	}
	if want := []string{"cfg.go", "go.mod", "ac8_test.go", "ref.go"}; !slices.Equal(reads, want) {
		t.Fatalf("reads=%v want %v", reads, want)
	}
}

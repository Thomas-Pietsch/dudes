package smoke_test

import (
	"strings"
	"testing"

	"bench/notes"
)

// AC-2 GIVEN an input named `notes.txt` with the lines `v1.0.0 First` and `1.x.0 Broken`, WHEN
// `notes.Parse` runs, THEN it returns an error starting with `ERR_NOTES: notes.txt:2:`.
func TestBL004AC2(t *testing.T) {
	_, err := notes.Parse("notes.txt", strings.NewReader("v1.0.0 First\n1.x.0 Broken\n"))
	if err == nil {
		t.Fatal("Parse accepted the broken line, want ERR_NOTES: notes.txt:2:")
	}
	if !strings.HasPrefix(err.Error(), "ERR_NOTES: notes.txt:2:") {
		t.Fatalf("Parse error %q, want the prefix ERR_NOTES: notes.txt:2:", err.Error())
	}
}

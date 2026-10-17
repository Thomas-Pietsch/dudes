package smoke_test

import (
	"os"
	"testing"

	"bench/notes"
)

// AC-1 GIVEN `testdata/notes.txt`, WHEN `notes.Parse` reads it and `notes.Render` renders the result,
// THEN the output equals `testdata/release-notes.md` byte for byte.
func TestBL004AC1(t *testing.T) {
	in, err := os.Open("../testdata/notes.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	ns, err := notes.Parse("notes.txt", in)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := notes.Render(ns)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	want, err := os.ReadFile("../testdata/release-notes.md")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Fatalf("Render differs from testdata/release-notes.md\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

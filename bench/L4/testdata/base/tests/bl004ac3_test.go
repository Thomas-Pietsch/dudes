package smoke_test

import (
	"strings"
	"testing"

	"bench/notes"
)

// AC-3 GIVEN the notes `1.2.0 Patch`, `v1.10.0 Minor` and `v1.9.3 Fix` in that order, WHEN
// `notes.Render` runs, THEN the headings follow as `## v1.10.0`, `## v1.9.3`, `## 1.2.0` and the
// file holds the line `Latest release: v1.10.0`.
func TestBL004AC3(t *testing.T) {
	got, err := notes.Render([]notes.Note{{Tag: "1.2.0", Text: "Patch"}, {Tag: "v1.10.0", Text: "Minor"}, {Tag: "v1.9.3", Text: "Fix"}})
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	last := -1
	for _, h := range []string{"## v1.10.0\n", "## v1.9.3\n", "## 1.2.0\n"} {
		i := strings.Index(got, h)
		if i < 0 || i < last {
			t.Fatalf("heading %q missing or out of order in:\n%s", strings.TrimSpace(h), got)
		}
		last = i
	}
	if !strings.Contains(got, "\nLatest release: v1.10.0\n") {
		t.Fatalf("no line `Latest release: v1.10.0` in:\n%s", got)
	}
}

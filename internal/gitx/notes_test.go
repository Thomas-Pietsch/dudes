package gitx_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// A note appended to grows, and every note of a ref reads back — a note on a commit of a branch that
// is gone as well (F88: a slice's ledger after its worktree and branch).
func TestAppendNoteAllNotes(t *testing.T) {
	dir, first := repoWithCommit(t)
	if notes, err := gitx.AllNotes(dir, "dudes/x"); err != nil || len(notes) != 0 {
		t.Fatalf("no ref yet: %v %v", notes, err)
	}
	for _, line := range []string{"one\n", "two\n"} {
		if err := gitx.AppendNote(dir, "dudes/x", line); err != nil {
			t.Fatal(err)
		}
	}
	gitOut(t, dir, "checkout", "-q", "-b", "slice")
	gitOut(t, dir, "commit", "-q", "--allow-empty", "-m", "on the slice")
	slice := gitOut(t, dir, "rev-parse", "HEAD")
	if err := gitx.AppendNote(dir, "dudes/x", "three\n"); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "checkout", "-q", "main")
	gitOut(t, dir, "branch", "-D", "slice")

	notes, err := gitx.AllNotes(dir, "dudes/x")
	if err != nil || len(notes) != 2 {
		t.Fatalf("notes %+v %v", notes, err)
	}
	got := map[string]string{}
	for _, n := range notes {
		got[n.Commit] = n.Text
	}
	if strings.Join(strings.Fields(got[first]), " ") != "one two" || strings.TrimSpace(got[slice]) != "three" {
		t.Errorf("notes %q", got)
	}
}

package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// A run relies on the committed spec, not on a working copy someone is still editing.
func TestReadCommitted(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "dudes@test"}, {"config", "user.name", "dudes-test"}, {"config", "core.autocrlf", "false"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	spec := filepath.Join(dir, "docs", "PRD.md")
	if err := os.MkdirAll(filepath.Dir(spec), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(spec, []byte("frozen\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.AddCommit(dir, []string{"docs/PRD.md"}, "spec"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(spec, []byte("work in progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := gitx.ReadCommitted(spec); err != nil || string(got) != "frozen\n" {
		t.Fatalf("committed: %q %v", got, err)
	}
	draft := filepath.Join(dir, "docs", "DRAFT.md")
	if err := os.WriteFile(draft, []byte("never committed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := gitx.ReadCommitted(draft); !os.IsNotExist(err) {
		t.Fatalf("an uncommitted file must not exist for a run: %v", err)
	}
	plain := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(plain, []byte("outside git\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := gitx.ReadCommitted(plain); err != nil || !strings.Contains(string(got), "outside git") {
		t.Fatalf("outside git: %q %v", got, err)
	}
}

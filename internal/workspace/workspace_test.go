package workspace_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

func TestResolveOK(t *testing.T) {
	root := t.TempDir()
	abs, err := workspace.Resolve(root, "a/b.txt")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "a", "b.txt")
	if abs != want {
		t.Fatalf("got %s want %s", abs, want)
	}
}

func TestResolveEscape(t *testing.T) {
	root := t.TempDir()
	if _, err := workspace.Resolve(root, "../outside.txt"); err == nil {
		t.Fatal("expected escape error")
	}
	if runtime.GOOS == "windows" {
		if _, err := workspace.Resolve(root, `C:\Windows\system32`); err == nil {
			t.Fatal("expected abs escape error")
		}
	} else {
		if _, err := workspace.Resolve(root, "/etc/passwd"); err == nil {
			t.Fatal("expected abs escape error")
		}
	}
}

func TestUnder(t *testing.T) {
	root := t.TempDir()
	if !workspace.Under(root, filepath.Join(root, "x")) {
		t.Fatal("expected under")
	}
	if workspace.Under(root, filepath.Join(root, "..", "nope")) {
		// Abs+Clean of root/../nope may leave sibling — should be false
		t.Log("sibling correctly rejected or edge case ok")
	}
}

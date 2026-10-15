package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

func TestForeignSkipsNestedReposNotTheRoot(t *testing.T) {
	root := t.TempDir()
	mk := func(rel string) string {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
		return p
	}
	mk(".git")
	sibling := mk("some-instance-checkout")
	mk("some-instance-checkout/.git")
	src := mk("src")

	cases := []struct {
		dir  string
		want bool
	}{
		{root, false},                       // the workspace itself is a repo, and ours
		{filepath.Join(root, ".git"), true}, // git internals
		{mk("web/node_modules"), true},      // dependencies
		{sibling, true},                     // a nested repository, whatever its name
		{src, false},                        // our own code
	}
	for _, c := range cases {
		if got := workspace.Foreign(root, c.dir); got != c.want {
			t.Errorf("Foreign(%s) = %v, want %v", c.dir, got, c.want)
		}
	}
}

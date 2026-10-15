package sandbox_test

import (
	"os/exec"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/sandbox"
)

func TestHonour(t *testing.T) {
	if !sandbox.Honour(nil, "edit") {
		t.Fatal("empty allow")
	}
	if sandbox.Honour([]string{"edit"}, "bash") {
		t.Fatal("bash not allowed")
	}
	if !sandbox.Honour([]string{"none", "edit"}, "edit") {
		t.Fatal("edit")
	}
	if !sandbox.Honour([]string{"worktree"}, "edit") {
		t.Fatal("worktree-only list is not a worker allow-list")
	}
}

func TestApply(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("true")
	if err := sandbox.Apply(cmd, dir); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range cmd.Env {
		if e == "DUDES_SANDBOX=1" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing sandbox marker")
	}
}

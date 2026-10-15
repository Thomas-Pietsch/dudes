// Package sandbox is a host-lite execution envelope (F46).
// It locks cwd, marks DUDES_SANDBOX, and does not start a container.
package sandbox

import (
	"os"
	"os/exec"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Environ copies the process env and sets DUDES_SANDBOX=1.
// API keys are left intact so Aider/Pi can still call a configured backend;
// log redaction remains the leak barrier (F44).
func Environ() []string {
	out := append([]string{}, os.Environ()...)
	out = append(out, "DUDES_SANDBOX=1")
	return out
}

// Honour reports whether name is in the recipe/task tool allow-list.
// Empty allow-list means the engine default (no extra restriction).
func Honour(allow []string, name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return true
	}
	filtered := Filter(allow)
	if len(filtered) == 0 {
		return true
	}
	for _, a := range filtered {
		if a == name || a == "all" {
			return true
		}
	}
	return false
}

// Filter drops placeholders like "none" and "worktree" (orchestrator flag, not a worker tool).
func Filter(tools []string) []string {
	var out []string
	for _, t := range tools {
		t = strings.ToLower(strings.TrimSpace(t))
		switch t {
		case "", "none", "worktree":
			continue
		default:
			out = append(out, t)
		}
	}
	return out
}

// Apply sets cwd under workspace and the sandbox env on cmd.
func Apply(cmd *exec.Cmd, workspaceRoot string) error {
	if cmd == nil {
		return nil
	}
	abs, err := workspace.Resolve(workspaceRoot, ".")
	if err != nil {
		return err
	}
	cmd.Dir = abs
	cmd.Env = Environ()
	return nil
}

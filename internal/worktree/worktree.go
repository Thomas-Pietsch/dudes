// Package worktree isolates an implementation cycle on a git worktree (F54).
package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const relDir = ".dudes/worktrees"

// RelDir is the workspace-relative parent of per-run worktrees.
func RelDir() string { return relDir }

// Branch is dudes/<sanitized runID>.
func Branch(runID string) string {
	return "dudes/" + safeRun(runID)
}

func safeRun(runID string) string {
	s := strings.TrimSpace(runID)
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := b.String()
	if out == "" {
		return "run"
	}
	return out
}

// Abs is .dudes/worktrees/<runID> under workspace.
func Abs(workspaceRoot, runID string) (string, error) {
	rel := filepath.ToSlash(filepath.Join(relDir, safeRun(runID)))
	return workspace.Resolve(workspaceRoot, rel)
}

// Want reports whether isolation was requested (config or tools list).
func Want(enabled bool, tools []string) bool {
	if enabled {
		return true
	}
	for _, t := range tools {
		if strings.EqualFold(strings.TrimSpace(t), "worktree") {
			return true
		}
	}
	return false
}

// Add creates a linked worktree on branch dudes/<runID>. Existing path is reused.
func Add(workspaceRoot, runID string) (abs, branch string, err error) {
	ok, err := gitx.IsRepo(workspaceRoot)
	if err != nil {
		return "", "", err
	}
	if !ok {
		return "", "", fmt.Errorf("worktree: not a git repository")
	}
	abs, err = Abs(workspaceRoot, runID)
	if err != nil {
		return "", "", err
	}
	branch = "dudes/" + safeRun(runID)
	if fi, err := os.Stat(abs); err == nil && fi.IsDir() {
		return abs, branch, nil
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", "", err
	}
	cmd := exec.Command("git", "worktree", "add", "-B", branch, abs, "HEAD")
	cmd.Dir = workspaceRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("git worktree add: %s", strings.TrimSpace(string(out)))
	}
	return abs, branch, nil
}

// Remove deletes the worktree (force). Missing is a no-op.
func Remove(workspaceRoot, runID string) error {
	abs, err := Abs(workspaceRoot, runID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	cmd := exec.Command("git", "worktree", "remove", "--force", abs)
	cmd.Dir = workspaceRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		_ = os.RemoveAll(abs)
		return fmt.Errorf("git worktree remove: %s", strings.TrimSpace(string(out)))
	}
	return nil
}

// MergeIntoHEAD fast-forwards/merges the worktree branch into the current HEAD (seal).
func MergeIntoHEAD(workspaceRoot, runID string) error {
	branch := "dudes/" + safeRun(runID)
	cmd := exec.Command("git", "merge", "--no-edit", branch)
	cmd.Dir = workspaceRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge %s: %s", branch, strings.TrimSpace(string(out)))
	}
	return nil
}

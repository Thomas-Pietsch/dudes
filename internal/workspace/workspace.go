// Package workspace enforces path containment under the engine workspace (F43).
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Resolve joins root and a relative path and ensures the result stays under root.
// Absolute rel that is already under root is accepted; escapes and ".." breakouts fail.
func Resolve(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("workspace root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)

	rel = filepath.FromSlash(strings.TrimSpace(rel))
	if rel == "" {
		return rootAbs, nil
	}

	var candidate string
	if filepath.IsAbs(rel) {
		candidate = filepath.Clean(rel)
	} else {
		candidate = filepath.Clean(filepath.Join(rootAbs, rel))
	}

	if !under(rootAbs, candidate) {
		return "", fmt.Errorf("path escapes workspace: %s", rel)
	}
	return candidate, nil
}

// Rel returns a slash-separated path relative to root, or error if outside.
func Rel(root, abs string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	abs, err = filepath.Abs(abs)
	if err != nil {
		return "", err
	}
	rootAbs = filepath.Clean(rootAbs)
	abs = filepath.Clean(abs)
	if !under(rootAbs, abs) {
		return "", fmt.Errorf("path outside workspace: %s", abs)
	}
	rel, err := filepath.Rel(rootAbs, abs)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

// Under reports whether abs is root or a descendant of root.
func Under(root, abs string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	abs, err = filepath.Abs(abs)
	if err != nil {
		return false
	}
	return under(filepath.Clean(rootAbs), filepath.Clean(abs))
}

func under(root, candidate string) bool {
	if candidate == root {
		return true
	}
	sep := string(filepath.Separator)
	return strings.HasPrefix(candidate, root+sep)
}

// Foreign reports whether a directory met while walking root is not this workspace's own content:
// git internals, node_modules, or a nested repository (a sibling instance checkout, a worktree).
func Foreign(root, dir string) bool {
	switch filepath.Base(dir) {
	case ".git", "node_modules":
		return true
	}
	if filepath.Clean(dir) == filepath.Clean(root) {
		return false
	}
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

package gitx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The git operations of item handoff (F88): a worktree per item on its own branch, the spec roots
// pinned as detached worktrees, nothing written into a repository's tracked files.

// Toplevel is the root of the work tree that contains dir.
func Toplevel(dir string) (string, error) {
	out, _, err := run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a git work tree: %w", dir, err)
	}
	return filepath.FromSlash(strings.TrimSpace(out)), nil
}

// CheckTopLevel refuses a directory that lies inside a repository without being its top level: git
// works on the repository around it, so the engine's commits and notes would land there. A directory
// outside any repository, or one that is its own top level, passes — the paths are compared as the
// file system sees them (case, links, short names), not as they are spelt.
func CheckTopLevel(dir string) error {
	if ok, err := IsRepo(dir); err != nil || !ok {
		return nil // no repository: other checks say so where it matters
	}
	top, err := Toplevel(dir)
	if err != nil {
		return nil // inside a .git folder or a bare repository: no work tree to speak of
	}
	a, errA := os.Stat(dir)
	b, errB := os.Stat(top)
	if errA != nil || errB != nil || os.SameFile(a, b) {
		return nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	return fmt.Errorf("workspace %s is inside the repository %s, not its top level: commits and notes would land in %s, the outer repository. Run from the top level of a repository (workspace \".\"), or make this folder a repository of its own (git init)", abs, top, top)
}

// Upstream is the remote-tracking branch the current branch follows ("origin/development");
// ok is false without one (a detached HEAD, a branch that tracks nothing).
func Upstream(dir string) (name string, ok bool) {
	out, _, err := run(dir, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	name = strings.TrimSpace(out)
	return name, err == nil && name != ""
}

// Commit resolves ref — a branch, a tag or a commit — to a commit SHA.
func Commit(dir, ref string) (string, error) {
	if ref == "" || strings.HasPrefix(ref, "-") {
		return "", fmt.Errorf("%q is not a git ref", ref)
	}
	out, _, err := run(dir, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if sha := strings.TrimSpace(out); err == nil && sha != "" {
		return sha, nil
	}
	return "", fmt.Errorf("%q is not a commit in %s", ref, dir)
}

// ShowFile returns the file rel (slash-separated, relative to the repository root) as commit rev
// holds it. A file the commit lacks is os.ErrNotExist.
func ShowFile(dir, rev, rel string) ([]byte, error) {
	if rev == "" || strings.HasPrefix(rev, "-") {
		return nil, fmt.Errorf("%q is not a git revision", rev)
	}
	out, _, err := run(dir, "show", rev+":"+filepath.ToSlash(rel))
	if err != nil {
		return nil, &os.PathError{Op: "git show", Path: rev + ":" + rel, Err: os.ErrNotExist}
	}
	return []byte(out), nil
}

// BranchExists reports whether the local branch name exists.
func BranchExists(dir, name string) bool {
	_, code, err := run(dir, "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil && code == 0
}

// AddWorktree creates a linked worktree at path: on a new branch started at commit when branch is
// set, else detached at commit. The branch tracks nothing: it is the item's own.
func AddWorktree(dir, path, branch, commit string) error {
	args := []string{"worktree", "add"}
	if branch != "" {
		args = append(args, "-b", branch, path, commit)
	} else {
		args = append(args, "--detach", path, commit)
	}
	if _, _, err := run(dir, args...); err != nil {
		return fmt.Errorf("git worktree add %s: %w", path, err)
	}
	return nil
}

// RemoveWorktree removes the linked worktree at path with its files and forgets stale registrations.
// A worktree that is already gone is fine.
func RemoveWorktree(dir, path string) error {
	_, _, err := run(dir, "worktree", "remove", "--force", path)
	if err != nil {
		_ = os.RemoveAll(path)
	}
	_, _, _ = run(dir, "worktree", "prune")
	if _, statErr := os.Stat(path); statErr == nil {
		return fmt.Errorf("git worktree remove %s: the directory is still there (%v)", path, err)
	}
	return nil
}

// DeleteBranch force-deletes the local branch name.
func DeleteBranch(dir, name string) error {
	if _, _, err := run(dir, "branch", "-D", name); err != nil {
		return fmt.Errorf("git branch -D %s: %w", name, err)
	}
	return nil
}

// Ignored reports whether git ignores the path rel (relative to dir), whether or not it exists.
func Ignored(dir, rel string) bool {
	_, code, err := run(dir, "check-ignore", "-q", "--", rel)
	return err == nil && code == 0
}

// Exclude adds pattern to the repository's local exclude file (info/exclude, shared by its
// worktrees, never committed) unless git already ignores probe, a path the pattern is meant to cover.
// The engine's own folders (.dudes/, .worktrees/) then stay out of `git status` and `git add -A`
// without a change to the repository's tracked .gitignore.
func Exclude(dir, pattern, probe string) error {
	if Ignored(dir, probe) {
		return nil
	}
	out, _, err := run(dir, "rev-parse", "--path-format=absolute", "--git-path", "info/exclude")
	if err != nil {
		return fmt.Errorf("git info/exclude: %w", err)
	}
	file := filepath.FromSlash(strings.TrimSpace(out))
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	have, err := os.ReadFile(file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	text := string(have)
	if text != "" && !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	return os.WriteFile(file, []byte(text+pattern+"\n"), 0o644)
}

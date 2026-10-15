// Package gitx runs orchestrator-owned git operations (no go-git dependency).
package gitx

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// IsRepo reports whether dir is inside a git work tree.
func IsRepo(dir string) (bool, error) {
	_, code, err := run(dir, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		if code == 128 {
			return false, nil
		}
		if strings.Contains(err.Error(), "not a git repository") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ChangedPaths returns worktree paths that are modified or untracked (porcelain).
// Paths under .dudes/ are omitted (engine state is not a worker artifact).
func ChangedPaths(dir string) ([]string, error) {
	ok, err := IsRepo(dir)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("not a git repository: %s", dir)
	}
	// -z: raw paths, NUL-separated — the text form quotes and octal-escapes a name with spaces or
	// non-ASCII (a "\302\247" in the name), which then matches no file.
	out, _, err := run(dir, "status", "--porcelain", "-z", "-u")
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}
	var paths []string
	seen := map[string]bool{}
	entries := strings.Split(out, "\x00")
	for i := 0; i < len(entries); i++ {
		e := entries[i]
		if len(e) < 4 {
			continue
		}
		path := e[3:]
		if e[0] == 'R' || e[0] == 'C' {
			i++ // a rename or copy is followed by its source path
		}
		// Engine state and the code worker's own files (Aider's chat history, tags cache) are
		// never worker output.
		if strings.HasPrefix(path, ".dudes/") || path == ".dudes" || strings.HasPrefix(path, ".aider") {
			continue
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	return paths, nil
}

// Head returns the current commit SHA (short), or empty if not a repo / no commits.
func Head(dir string) (string, error) {
	return revParse(dir, "--short", "HEAD")
}

// HeadFull returns the full HEAD SHA, or empty if unavailable.
func HeadFull(dir string) (string, error) {
	return revParse(dir, "HEAD")
}

func revParse(dir string, args ...string) (string, error) {
	ok, err := IsRepo(dir)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	out, code, err := run(dir, append([]string{"rev-parse"}, args...)...)
	if err != nil {
		if code == 128 {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// RemoteURL returns origin URL, or the directory base name if origin is unset.
func RemoteURL(dir string) string {
	out, _, err := run(dir, "remote", "get-url", "origin")
	if err != nil {
		base := filepath.Base(dir)
		if base == "" || base == "." {
			return dir
		}
		return base
	}
	s := strings.TrimSpace(out)
	if s == "" {
		return filepath.Base(dir)
	}
	return s
}

// Provenance is repo identity + revision for freeze records (F52).
func Provenance(dir string) (repo, rev string) {
	repo = RemoteURL(dir)
	rev, _ = HeadFull(dir)
	if rev == "" {
		rev, _ = Head(dir)
	}
	return repo, rev
}

// Discard drops the uncommitted changes of paths: a path HEAD knows is restored, a new one removed.
func Discard(dir string, paths []string) error {
	for _, p := range paths {
		if _, _, err := run(dir, "cat-file", "-e", "HEAD:"+p); err == nil {
			if _, _, err := run(dir, "checkout", "HEAD", "--", p); err != nil {
				return fmt.Errorf("git checkout %s: %w", p, err)
			}
			continue
		}
		// -f: Aider stages a file it creates, then may change it again before it exits.
		if _, _, err := run(dir, "rm", "-q", "--cached", "-f", "--ignore-unmatch", "--", p); err != nil {
			return fmt.Errorf("git rm --cached %s: %w", p, err)
		}
		if err := os.Remove(exactPath(dir, p)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// AddCommit stages paths and creates a commit. No-op if nothing staged.
func AddCommit(dir string, paths []string, message string) error {
	ok, err := IsRepo(dir)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not a git repository: %s", dir)
	}
	if message == "" {
		return fmt.Errorf("empty commit message")
	}
	if len(paths) == 0 {
		return fmt.Errorf("no paths to commit")
	}
	for _, p := range paths {
		if filepath.IsAbs(p) || strings.Contains(p, "..") {
			if _, err := workspace.Resolve(dir, p); err != nil {
				return fmt.Errorf("git add path: %w", err)
			}
		} else if _, err := workspace.Resolve(dir, p); err != nil {
			return fmt.Errorf("git add path: %w", err)
		}
	}
	args := append([]string{"add", "--"}, paths...)
	if _, _, err := run(dir, args...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	_, code, err := run(dir, "diff", "--cached", "--quiet")
	if err != nil && code != 1 {
		return fmt.Errorf("git diff --cached: %w", err)
	}
	if code == 0 {
		return nil // nothing staged
	}
	if _, _, err := run(dir, "--no-pager", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// AnnotatedTag creates an annotated tag at HEAD. name must be a valid ref leaf
// (e.g. dudes/seal/demo/run-1). Fails if the tag already exists.
func AnnotatedTag(dir, name, message string) error {
	return tagAtHead(dir, name, message, false)
}

// tagAtHead makes an annotated tag at HEAD, signed (git tag -s) when sign is set.
func tagAtHead(dir, name, message string, sign bool) error {
	ok, err := IsRepo(dir)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("not a git repository: %s", dir)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty tag name")
	}
	if message == "" {
		message = name
	}
	kind := "-a"
	if sign {
		kind = "-s"
	}
	if _, _, err := run(dir, "tag", kind, name, "-m", message); err != nil {
		return fmt.Errorf("git tag %s: %w", kind, err)
	}
	return nil
}

// EmptyCommit records an event that changes no file (e.g. the start of a run in a repo without
// commits yet) as a commit of its own.
func EmptyCommit(dir, message string) error {
	if _, _, err := run(dir, "--no-pager", "commit", "--allow-empty", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// EnsureHead gives a repository without commits its first one (empty, with message), so notes have a
// commit to attach to. A repository with commits is left alone; no repository is an error.
func EnsureHead(dir, message string) error {
	if head, _ := Head(dir); head != "" {
		return nil
	}
	if ok, err := IsRepo(dir); err != nil || !ok {
		return fmt.Errorf("%s is not a git repository (git init) — DUDES keeps its records in git", dir)
	}
	return EmptyCommit(dir, message)
}

// AddNote attaches text to HEAD under refs/notes/<ref> (CORE §10 annotation: facts about a commit,
// added later without changing it). An existing note on HEAD is replaced. Notes stay local unless
// the namespace is pushed on purpose.
func AddNote(dir, ref, text string) error {
	f, err := os.CreateTemp("", "dudes-note-*")
	if err != nil {
		return fmt.Errorf("git notes: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(text); err != nil {
		f.Close()
		return fmt.Errorf("git notes: %w", err)
	}
	f.Close()
	if _, _, err := run(dir, "notes", "--ref", ref, "add", "-f", "-F", f.Name(), "HEAD"); err != nil {
		return fmt.Errorf("git notes add: %w", err)
	}
	return nil
}

// LatestNote returns the newest note under refs/notes/<ref> in HEAD's history (at most depth
// commits back) and the commit it is attached to. ok=false: no note, or no commit yet.
func LatestNote(dir, ref string, depth int) (commit, text string, ok bool, err error) {
	notes, err := Notes(dir, ref, depth)
	if err != nil || len(notes) == 0 {
		return "", "", false, err
	}
	return notes[0].Commit, notes[0].Text, true, nil
}

// Note is one note and the commit it is attached to.
type Note struct{ Commit, Text string }

// Notes returns the notes under refs/notes/<ref> in HEAD's history (at most depth commits back),
// newest first. None: no note, or no commit yet.
func Notes(dir, ref string, depth int) ([]Note, error) {
	if head, _ := revParse(dir, "--verify", "-q", "HEAD"); head == "" {
		return nil, nil // not a repo, or no commit yet
	}
	out, _, err := run(dir, "log", "-n", fmt.Sprint(depth), "--notes="+ref, "--format=%H%x1f%N%x1e", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git log notes: %w", err)
	}
	var notes []Note
	for _, rec := range strings.Split(out, "\x1e") {
		c, note, found := strings.Cut(strings.TrimLeft(rec, "\r\n"), "\x1f")
		if found && strings.TrimSpace(note) != "" {
			notes = append(notes, Note{Commit: c, Text: strings.TrimRight(note, "\r\n") + "\n"})
		}
	}
	return notes, nil
}

// RevertSince undoes the commits after base in one new commit: the tree is base's again, the
// undone commits stay in the history.
func RevertSince(dir, base, message string) error {
	if _, _, err := run(dir, "revert", "--no-commit", base+"..HEAD"); err != nil {
		_, _, _ = run(dir, "revert", "--abort")
		return fmt.Errorf("git revert %s..HEAD: %w", base, err)
	}
	if _, _, err := run(dir, "commit", "-q", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Message returns the full message of commit rev.
func Message(dir, rev string) (string, error) {
	out, _, err := run(dir, "log", "-1", "--format=%B", rev)
	return out, err
}

// IsAncestor reports whether commit a is an ancestor of commit b (a == b counts).
func IsAncestor(dir, a, b string) bool {
	_, code, _ := run(dir, "merge-base", "--is-ancestor", a, b)
	return code == 0
}

var semver = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

// ReleaseTagName is the seal tag of the git contract, <product>-vX.Y.Z (CORE §10): the same name
// in the spec repo and the dev repos traces spec ↔ code.
func ReleaseTagName(product, version string) (string, error) {
	if !semver.MatchString(version) {
		return "", fmt.Errorf("release version %q is not X.Y.Z", version)
	}
	return product + "-v" + version, nil
}

// TagExists reports whether tag name exists in the repo at dir.
func TagExists(dir, name string) bool {
	_, code, err := run(dir, "rev-parse", "-q", "--verify", "refs/tags/"+name)
	return err == nil && code == 0
}

// SealTagName builds dudes/seal/<product>/<runID> with unsafe chars replaced — the seal of a run
// that names no release version.
func SealTagName(product, runID string) string {
	safe := func(s string) string {
		s = strings.TrimSpace(s)
		var b strings.Builder
		for _, r := range s {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
				b.WriteRune(r)
			default:
				b.WriteByte('-')
			}
		}
		out := b.String()
		if out == "" {
			return "unknown"
		}
		return out
	}
	return "dudes/seal/" + safe(product) + "/" + safe(runID)
}

func run(dir string, args ...string) (stdout string, exitCode int, err error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	stdout = outBuf.String()
	if runErr == nil {
		return stdout, 0, nil
	}
	if ee, ok := runErr.(*exec.ExitError); ok {
		exitCode = ee.ExitCode()
	} else {
		exitCode = -1
	}
	msg := strings.TrimSpace(errBuf.String())
	if msg == "" {
		msg = runErr.Error()
	}
	return stdout, exitCode, fmt.Errorf("%s", msg)
}

// Tracked lists the repository's tracked files, raw (ls-files -z: no quoting).
func Tracked(dir string) ([]string, error) {
	out, _, err := run(dir, "ls-files", "-z")
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w", err)
	}
	var files []string
	for _, f := range strings.Split(out, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

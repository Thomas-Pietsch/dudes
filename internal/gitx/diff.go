package gitx

import (
	"fmt"
	"strconv"
	"strings"
)

// What changed between two commits, for a reviewer (F88): the files with their line counts, the patch
// and the commits, each limited to the paths given. A rename counts as a deletion and an addition.

// FileChange is one file that differs between two commits: lines added and deleted, or Binary when
// git counts none.
type FileChange struct {
	Path           string
	Added, Deleted int
	Binary         bool
}

// Changes lists the files that differ between commits from and to.
func Changes(dir, from, to string) ([]FileChange, error) {
	out, _, err := run(dir, "diff", "--numstat", "-z", "--no-renames", from, to, "--")
	if err != nil {
		return nil, fmt.Errorf("git diff --numstat: %w", err)
	}
	var files []FileChange
	for _, rec := range strings.Split(out, "\x00") {
		added, rest, ok := strings.Cut(rec, "\t")
		if !ok {
			continue
		}
		deleted, path, ok := strings.Cut(rest, "\t")
		if !ok || path == "" {
			continue
		}
		f := FileChange{Path: path}
		if added == "-" && deleted == "-" {
			f.Binary = true
		} else {
			f.Added, _ = strconv.Atoi(added)
			f.Deleted, _ = strconv.Atoi(deleted)
		}
		files = append(files, f)
	}
	return files, nil
}

// Diff is the patch from commit from to commit to, for paths only (taken literally); no paths → every
// file. Plain git: no colour, no external diff or text conversion the user may have configured.
func Diff(dir, from, to string, paths []string) (string, error) {
	args := append([]string{"diff", "--no-color", "--no-ext-diff", "--no-textconv", "--no-renames", from, to, "--"}, literal(paths)...)
	out, _, err := run(dir, args...)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}

// Subjects lists the commits after from up to to that touch paths (taken literally; none → any), newest
// first: the short hash and the subject.
func Subjects(dir, from, to string, paths []string) ([]string, error) {
	args := append([]string{"log", "--format=%h %s", from + ".." + to, "--"}, literal(paths)...)
	out, _, err := run(dir, args...)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	var subjects []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			subjects = append(subjects, line)
		}
	}
	return subjects, nil
}

// literal turns paths into pathspecs that match those paths and nothing else.
func literal(paths []string) []string {
	specs := make([]string, 0, len(paths))
	for _, p := range paths {
		specs = append(specs, ":(literal)"+p)
	}
	return specs
}

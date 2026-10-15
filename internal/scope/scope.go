// Package scope writes the work item's scope file that the git hooks enforce (F78, F77):
// `.dudes/writable` in the work tree — `# item: <ID>` (commit-msg then requires the trailer
// `Item: <ID>`) and one path pattern per line (pre-commit refuses staged paths outside them).
// A pattern with a leading `!` excludes: what it matches is never writable, even inside another
// pattern — how a task keeps the files an earlier task wrote (tests, hooks) read-only (F88).
package scope

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const rel = ".dudes/writable"

// Path returns the scope file in work tree dir.
func Path(dir string) string { return filepath.Join(dir, filepath.FromSlash(rel)) }

// Write sets the scope of work tree dir. Nothing to scope (no item, no patterns) → no file.
func Write(dir, item string, patterns []string) error {
	item = strings.TrimSpace(item)
	var b strings.Builder
	b.WriteString("# Scope of the bound work item — written by dudes, enforced by hooks/ (F78)\n")
	if item != "" {
		fmt.Fprintf(&b, "# item: %s\n", item)
	}
	n := 0
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p != "" {
			b.WriteString(p + "\n")
			n++
		}
	}
	if item == "" && n == 0 {
		return Remove(dir)
	}
	if err := os.MkdirAll(filepath.Dir(Path(dir)), 0o755); err != nil {
		return fmt.Errorf("scope: %w", err)
	}
	return os.WriteFile(Path(dir), []byte(b.String()), 0o644)
}

// Remove clears the scope of work tree dir; a missing file is fine.
func Remove(dir string) error {
	if err := os.Remove(Path(dir)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("scope: %w", err)
	}
	return nil
}

// Inside reports whether path (slash-separated, repository-relative) matches one of the patterns
// the way the pre-commit hook's `case` does: `*` matches any run of characters, `/` included,
// `?` one character. A pattern with a leading `!` excludes: a path it matches is never inside,
// whatever else matches. No patterns, or only exclusions, → everything not excluded is inside.
func Inside(path string, patterns []string) bool {
	positive, matched := 0, false
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p == "" || strings.HasPrefix(p, "#") {
			continue
		}
		if ex, ok := strings.CutPrefix(p, "!"); ok {
			if glob(ex).MatchString(path) {
				return false
			}
			continue
		}
		positive++
		if glob(p).MatchString(path) {
			matched = true
		}
	}
	return matched || positive == 0
}

// Included are the patterns that admit paths: no exclusions, blanks or comments.
func Included(patterns []string) []string {
	var out []string
	for _, p := range patterns {
		if p = strings.TrimSpace(p); p != "" && !strings.HasPrefix(p, "#") && !strings.HasPrefix(p, "!") {
			out = append(out, p)
		}
	}
	return out
}

func glob(p string) *regexp.Regexp {
	var b strings.Builder
	b.WriteString("^")
	for _, r := range p {
		switch r {
		case '*':
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

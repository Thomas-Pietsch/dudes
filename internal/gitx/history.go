package gitx

import (
	"fmt"
	"strings"
)

// The history of an item's branch (F88): which commits a task made, read from their trailers, and
// the squash that lands an approved task as one commit.

// Entry is one commit of a history walk: its hash, its subject and its trailers.
type Entry struct {
	SHA, Subject string
	Trailers     map[string][]string // by key as written, values in order
}

// Trailer is the first value of the trailer key, matched without case; "" when the commit has none.
func (e Entry) Trailer(key string) string {
	for k, vs := range e.Trailers {
		if strings.EqualFold(k, key) && len(vs) > 0 {
			return vs[0]
		}
	}
	return ""
}

// FirstParents is HEAD's first-parent history, newest first, at most n commits; none without a commit.
func FirstParents(dir string, n int) ([]Entry, error) {
	if head, _ := revParse(dir, "--verify", "-q", "HEAD"); head == "" {
		return nil, nil
	}
	out, _, err := run(dir, "log", "--first-parent", "-n", fmt.Sprint(n), "--format=%H%x1f%s%x1f%(trailers:only,unfold)%x1e", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	var entries []Entry
	for _, rec := range strings.Split(out, "\x1e") {
		f := strings.Split(strings.TrimLeft(rec, "\r\n"), "\x1f")
		if len(f) != 3 || f[0] == "" {
			continue
		}
		e := Entry{SHA: f[0], Subject: f[1], Trailers: map[string][]string{}}
		for _, line := range strings.Split(f[2], "\n") {
			if k, v, ok := strings.Cut(line, ":"); ok && strings.TrimSpace(k) != "" {
				k = strings.TrimSpace(k)
				e.Trailers[k] = append(e.Trailers[k], strings.TrimSpace(v))
			}
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// Files lists the files commit rev holds under the folder rel (slash-separated, from the repository's
// root), by their paths from the root; none when it holds no such folder.
func Files(dir, rev, rel string) ([]string, error) {
	if rev == "" || strings.HasPrefix(rev, "-") {
		return nil, fmt.Errorf("%q is not a git revision", rev)
	}
	out, _, err := run(dir, "ls-tree", "-r", "-z", "--full-tree", "--name-only", rev, "--", strings.TrimSuffix(rel, "/")+"/")
	if err != nil {
		return nil, fmt.Errorf("git ls-tree %s: %w", rev, err)
	}
	var files []string
	for _, f := range strings.Split(out, "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

// Branch is the full name of the branch checked out in dir ("refs/heads/feature/bk-001a"); "" when
// HEAD is detached.
func Branch(dir string) string {
	out, _, err := run(dir, "symbolic-ref", "-q", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// Containing lists the branches and remote-tracking branches whose history holds commit, by full ref
// name. A commit another ref holds is shared: rewriting it forks that ref's history.
func Containing(dir, commit string) ([]string, error) {
	if commit == "" || strings.HasPrefix(commit, "-") {
		return nil, fmt.Errorf("%q is not a commit", commit)
	}
	out, _, err := run(dir, "for-each-ref", "--contains", commit, "--format=%(refname)", "refs/heads", "refs/remotes")
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	return strings.Fields(out), nil
}

// Refs lists the refs under the folder prefix ("refs/dudes/passes/x/"), by full name, sorted.
func Refs(dir, prefix string) ([]string, error) {
	out, _, err := run(dir, "for-each-ref", "--format=%(refname)", strings.TrimSuffix(prefix, "/")+"/")
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref: %w", err)
	}
	return strings.Fields(out), nil
}

// NotesRefs lists the notes refs under refs/notes/<prefix>, by the name git notes --ref takes
// ("dudes/test").
func NotesRefs(dir, prefix string) ([]string, error) {
	full, err := Refs(dir, "refs/notes/"+prefix)
	if err != nil {
		return nil, err
	}
	var refs []string
	for _, r := range full {
		refs = append(refs, strings.TrimPrefix(r, "refs/notes/"))
	}
	return refs, nil
}

// CreateRef makes the ref name (a full name, "refs/dudes/…") point at commit; it refuses a ref that
// exists. A ref outside refs/heads and refs/tags is kept local: neither git push nor a clone takes it
// by default.
func CreateRef(dir, name, commit string) error {
	if _, _, err := run(dir, "update-ref", name, commit, ""); err != nil {
		return fmt.Errorf("git update-ref %s: %w", name, err)
	}
	return nil
}

// MergeNoFF merges branch into the branch checked out in dir with a merge commit of its own, message
// as its message, through git merge and its hooks. A merge git stops — a conflict, a hook's refusal —
// is taken back (git merge --abort), so dir is as before.
func MergeNoFF(dir, branch, message string) error {
	if branch == "" || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("%q is not a branch", branch)
	}
	out, _, err := run(dir, "merge", "--no-ff", "-m", message, branch)
	if err == nil {
		return nil
	}
	if _, _, verr := run(dir, "rev-parse", "-q", "--verify", "MERGE_HEAD"); verr == nil {
		_, _, _ = run(dir, "merge", "--abort")
	}
	return fmt.Errorf("git merge %s: %v %s", branch, err, strings.TrimSpace(out))
}

// DeleteMergedBranch deletes the local branch name, which must be merged into HEAD (git branch -d).
func DeleteMergedBranch(dir, name string) error {
	if _, _, err := run(dir, "branch", "-d", name); err != nil {
		return fmt.Errorf("git branch -d %s: %w", name, err)
	}
	return nil
}

// DeleteRef removes the ref name; a ref that is not there is fine.
func DeleteRef(dir, name string) error {
	if _, _, err := run(dir, "update-ref", "-d", name); err != nil {
		return fmt.Errorf("git update-ref -d %s: %w", name, err)
	}
	return nil
}

// Squash makes the commits after base one commit of HEAD's tree with message, committed by git commit,
// so the hooks run on it. The notes HEAD carries under notesRefs go to the new commit: a test run is a
// fact about the tree, and the tree is the same. It refuses, and changes nothing, when the index holds
// a staged change (it would join the commit) or the commits change nothing against base; a commit the
// hooks refuse leaves HEAD where it was.
func Squash(dir, base, message string, notesRefs []string) (string, error) {
	head, err := HeadFull(dir)
	if err != nil || head == "" {
		return "", fmt.Errorf("no HEAD to squash in %s (%v)", dir, err)
	}
	if _, code, err := run(dir, "diff", "--cached", "--quiet"); code == 1 {
		return "", fmt.Errorf("the index holds staged changes, and they would join the commit")
	} else if err != nil {
		return "", fmt.Errorf("git diff --cached: %w", err)
	}
	if _, code, err := run(dir, "diff", "--quiet", base, head, "--"); code == 0 {
		return "", fmt.Errorf("the commits after %s change nothing", shortRev(base))
	} else if code != 1 {
		return "", fmt.Errorf("git diff %s: %w", shortRev(base), err)
	}
	if err := ResetSoft(dir, base); err != nil {
		return "", err
	}
	if _, _, err := run(dir, "--no-pager", "commit", "-q", "-m", message); err != nil {
		if rerr := ResetSoft(dir, head); rerr != nil {
			return "", fmt.Errorf("git commit: %v — and HEAD could not go back to %s: %v", err, shortRev(head), rerr)
		}
		return "", fmt.Errorf("git commit: %w", err)
	}
	sha, err := HeadFull(dir)
	if err != nil {
		return "", err
	}
	for _, ref := range notesRefs {
		if _, _, err := run(dir, "notes", "--ref", ref, "list", head); err != nil {
			continue // no note there
		}
		if _, _, err := run(dir, "notes", "--ref", ref, "copy", "-f", head, sha); err != nil {
			return sha, fmt.Errorf("git notes --ref %s copy: %w", ref, err)
		}
	}
	return sha, nil
}

func shortRev(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

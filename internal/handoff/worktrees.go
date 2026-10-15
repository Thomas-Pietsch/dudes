package handoff

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// Plan is where one slice's worktrees go and what they start from. NewPlan checks that nothing is
// in the way; PlanPins resolves the spec roots; Create makes the dev worktree and the pins;
// Rollback takes back what Create made.
type Plan struct {
	Root     string // the dev repository's root, absolute
	Slicing  *Slicing
	Base     string // the base as named, or the default: the upstream of the current branch, else HEAD
	BaseSHA  string
	Worktree string // absolute
	Branch   string
	Pins     []Pin // set by PlanPins
	undo     []func() error
}

// Pin is a spec root checked out, detached, at its pin: the spec as frozen, which stays fixed while
// people keep committing in their checkout of the spec repository.
type Pin struct {
	Root   string // the root's name
	Repo   string // the spec repository's root, absolute
	Ref    string // the pin: a branch or a tag of that repository
	Commit string
	Path   string // the detached worktree, absolute
}

// NewPlan resolves the base and refuses when the slice's worktree or branch exists: preparing a
// slice twice would overwrite its work. base "" is the upstream of the current branch, else HEAD.
func NewPlan(root string, s *Slicing, base string) (*Plan, error) {
	if base == "" {
		if up, ok := gitx.Upstream(root); ok {
			base = up
		} else {
			base = "HEAD"
		}
	}
	sha, err := gitx.Commit(root, base)
	if err != nil {
		return nil, fmt.Errorf("base: %w", err)
	}
	p := &Plan{
		Root: root, Slicing: s, Base: base, BaseSHA: sha,
		Worktree: filepath.Join(root, filepath.FromSlash(s.WorktreeRel())),
		Branch:   s.Branch(),
	}
	if _, err := os.Stat(p.Worktree); err == nil {
		return nil, fmt.Errorf("the worktree %s already exists — the slice %s is prepared; remove it (git worktree remove) or name another slice", s.WorktreeRel(), s.ID)
	}
	if gitx.BranchExists(root, p.Branch) {
		return nil, fmt.Errorf("the branch %s already exists — the slice %s is prepared; delete it or name another slice", p.Branch, s.ID)
	}
	return p, nil
}

// PlanPins resolves every root's repository and pin, in the order of names, and checks that no
// pinned worktree is in the way; it creates nothing. A slice pins one worktree per repository, so two
// roots in one repository are refused.
func (p *Plan) PlanPins(roots map[string]Root, names []string) error {
	byRepo := map[string]string{} // repository → the root that names it
	p.Pins = nil
	for _, name := range names {
		r := roots[name]
		repo := r.Repo
		if !filepath.IsAbs(repo) {
			repo = filepath.Join(p.Root, filepath.FromSlash(repo))
		}
		repo = filepath.Clean(repo)
		top, err := gitx.Toplevel(repo)
		if err != nil || !SameDir(top, repo) {
			return fmt.Errorf("root %s: %s is not the root of a git repository", name, r.Repo)
		}
		if first, ok := byRepo[top]; ok {
			return fmt.Errorf("roots %s and %s are one repository (%s): a slice pins one worktree per repository", first, name, r.Repo)
		}
		byRepo[top] = name
		sha, err := gitx.Commit(top, r.Pin)
		if err != nil {
			return fmt.Errorf("root %s: pin %s is not a branch or tag of %s", name, r.Pin, r.Repo)
		}
		path := filepath.Join(top, filepath.FromSlash(p.Slicing.PinRel()))
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("root %s: the pinned worktree %s already exists — remove it (git -C %s worktree remove) or name another slice", name, path, top)
		}
		p.Pins = append(p.Pins, Pin{Root: name, Repo: top, Ref: r.Pin, Commit: sha, Path: path})
	}
	return nil
}

// Create makes the dev worktree on its feature branch and every pinned worktree, and keeps the
// engine's .worktrees/ out of git status in each repository it adds to (the repository's local
// exclude file; its tracked .gitignore stays as it is). A failure takes back what was made.
func (p *Plan) Create() error {
	if err := gitx.Exclude(p.Root, "/.worktrees/", p.Slicing.WorktreeRel()+"/x"); err != nil {
		return err
	}
	p.undo = append(p.undo, func() error {
		err := gitx.RemoveWorktree(p.Root, p.Worktree)
		if gitx.BranchExists(p.Root, p.Branch) {
			if berr := gitx.DeleteBranch(p.Root, p.Branch); err == nil {
				err = berr
			}
		}
		return err
	})
	if err := gitx.AddWorktree(p.Root, p.Worktree, p.Branch, p.BaseSHA); err != nil {
		p.Rollback()
		return err
	}
	for _, pin := range p.Pins {
		pin := pin
		if err := gitx.Exclude(pin.Repo, "/.worktrees/", p.Slicing.PinRel()+"/x"); err != nil {
			p.Rollback()
			return err
		}
		p.undo = append(p.undo, func() error { return gitx.RemoveWorktree(pin.Repo, pin.Path) })
		if err := gitx.AddWorktree(pin.Repo, pin.Path, "", pin.Commit); err != nil {
			p.Rollback()
			return err
		}
	}
	return nil
}

// Rollback takes back what Create made, newest first: the slice is then as before the command. What
// it could not remove is returned, to tell the human.
func (p *Plan) Rollback() []error {
	var errs []error
	for i := len(p.undo) - 1; i >= 0; i-- {
		if err := p.undo[i](); err != nil {
			errs = append(errs, err)
		}
	}
	p.undo = nil
	return errs
}

// PinPaths are the pinned worktrees by root name.
func (p *Plan) PinPaths() map[string]string {
	out := map[string]string{}
	for _, pin := range p.Pins {
		out[pin.Root] = pin.Path
	}
	return out
}

// SameDir reports whether a and b are one directory, however their paths are spelt.
func SameDir(a, b string) bool {
	fa, err := os.Stat(a)
	if err != nil {
		return false
	}
	fb, err := os.Stat(b)
	if err != nil {
		return false
	}
	return os.SameFile(fa, fb)
}

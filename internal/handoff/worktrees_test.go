package handoff_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/handoff"
)

// repos is a dev repository beside a spec repository with a release branch, a tag and a later commit.
type repos struct {
	dev, spec       string
	release, tagged string // the commits of the release branch and of the tag
	devHead         string
}

func newRepos(t *testing.T) *repos {
	t.Helper()
	parent := t.TempDir()
	r := &repos{dev: filepath.Join(parent, "dev"), spec: filepath.Join(parent, "spec")}
	newRepo(t, r.spec)
	writeFile(t, r.spec, "P_PRD.md", "1.0\n")
	gitT(t, r.spec, "add", "-A")
	gitT(t, r.spec, "commit", "-qm", "1.0")
	r.release = gitT(t, r.spec, "rev-parse", "HEAD")
	gitT(t, r.spec, "branch", "release/p-v1.0")
	writeFile(t, r.spec, "P_PRD.md", "1.1\n")
	gitT(t, r.spec, "commit", "-qam", "1.1")
	r.tagged = gitT(t, r.spec, "rev-parse", "HEAD")
	gitT(t, r.spec, "tag", "p-v1.1.0")
	writeFile(t, r.spec, "P_PRD.md", "later\n")
	gitT(t, r.spec, "commit", "-qam", "later")

	newRepo(t, r.dev)
	writeFile(t, r.dev, "main.go", "package main\n")
	gitT(t, r.dev, "add", "-A")
	gitT(t, r.dev, "commit", "-qm", "dev")
	r.devHead = gitT(t, r.dev, "rev-parse", "HEAD")
	return r
}

func slice(id string) *handoff.Slicing {
	return &handoff.Slicing{ID: id, Acceptance: []string{"AC-1"}, Scope: []string{"pkg/*"}, Tests: []string{"pkg/x_test.go"}}
}

func TestPlanBase(t *testing.T) {
	r := newRepos(t)
	// no upstream: HEAD
	p, err := handoff.NewPlan(r.dev, slice("a"), "")
	if err != nil || p.Base != "HEAD" || p.BaseSHA != r.devHead {
		t.Fatalf("default base: %+v %v", p, err)
	}
	if p.Branch != "feature/a" || p.Worktree != filepath.Join(r.dev, ".worktrees", "a") {
		t.Errorf("plan %s %s", p.Branch, p.Worktree)
	}
	// the upstream of the current branch: a remote that is behind the local branch
	remote := filepath.Join(t.TempDir(), "origin.git")
	gitT(t, r.dev, "init", "-q", "--bare", remote)
	gitT(t, r.dev, "remote", "add", "origin", remote)
	gitT(t, r.dev, "push", "-q", "-u", "origin", "main")
	writeFile(t, r.dev, "local.go", "package main\n")
	gitT(t, r.dev, "add", "-A")
	gitT(t, r.dev, "commit", "-qm", "local, not pushed")
	if p, err = handoff.NewPlan(r.dev, slice("a"), ""); err != nil || p.Base != "origin/main" || p.BaseSHA != r.devHead {
		t.Fatalf("base from the upstream: %+v %v", p, err)
	}
	// -base names another ref, or fails
	if p, err = handoff.NewPlan(r.dev, slice("a"), "HEAD"); err != nil || p.BaseSHA == r.devHead {
		t.Fatalf("-base HEAD: %+v %v", p, err)
	}
	if _, err = handoff.NewPlan(r.dev, slice("a"), "no-such-ref"); err == nil || !strings.Contains(err.Error(), "base") {
		t.Errorf("unknown base: %v", err)
	}
}

// The pin is made at the release branch, detached, in the spec repository's .worktrees/; nothing is
// made until Create; Rollback takes back the worktree, the branch and the pin.
func TestPlanPinsAndCreate(t *testing.T) {
	r := newRepos(t)
	roots := map[string]handoff.Root{"pm": {Repo: "../spec", Pin: "release/p-v1.0"}}
	p, err := handoff.NewPlan(r.dev, slice("a"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.PlanPins(roots, []string{"pm"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(r.dev, ".worktrees")); !os.IsNotExist(err) {
		t.Fatal("planning made a directory")
	}
	if err := p.Create(); err != nil {
		t.Fatal(err)
	}
	pin := filepath.Join(r.spec, ".worktrees", "spec-a")
	if got := gitT(t, pin, "rev-parse", "HEAD"); got != r.release {
		t.Errorf("pin at %s, want the release %s", got, r.release)
	}
	if got := gitT(t, pin, "rev-parse", "--abbrev-ref", "HEAD"); got != "HEAD" {
		t.Errorf("the pin is on branch %q", got)
	}
	if paths := p.PinPaths(); len(paths) != 1 || paths["pm"] != pin {
		t.Errorf("pin paths %v", paths)
	}
	if got := gitT(t, p.Worktree, "branch", "--show-current"); got != "feature/a" {
		t.Errorf("branch %q", got)
	}
	if got := gitT(t, p.Worktree, "rev-parse", "HEAD"); got != r.devHead {
		t.Errorf("the worktree starts at %s, want the base %s", got, r.devHead)
	}
	// .worktrees/ stays out of git status in both repositories, without a tracked change
	for _, repo := range []string{r.dev, r.spec} {
		if got := gitT(t, repo, "status", "--porcelain"); got != "" {
			t.Errorf("%s: git status shows %q", repo, got)
		}
	}
	if _, err := handoff.NewPlan(r.dev, slice("a"), ""); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("a second plan for the slice: %v", err)
	}

	if errs := p.Rollback(); len(errs) != 0 {
		t.Fatalf("rollback: %v", errs)
	}
	for _, path := range []string{p.Worktree, pin} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s remains (%v)", path, err)
		}
	}
	if gitT(t, r.dev, "branch", "--list", "feature/a") != "" {
		t.Error("the branch remains")
	}
}

// Two spec repositories, each pinned at its own ref — the first instance has a lead and a pm repo —
// are both checked out; a pin that does not exist makes nothing.
func TestPlanTwoSpecRepos(t *testing.T) {
	r := newRepos(t)
	lead := filepath.Join(filepath.Dir(r.dev), "lead")
	newRepo(t, lead)
	writeFile(t, lead, "STRATEGY.md", "s\n")
	gitT(t, lead, "add", "-A")
	gitT(t, lead, "commit", "-qm", "s")
	leadPin := gitT(t, lead, "rev-parse", "HEAD")
	gitT(t, lead, "tag", "p-v1.0.0")
	writeFile(t, lead, "STRATEGY.md", "s2\n")
	gitT(t, lead, "commit", "-qam", "s2")

	roots := map[string]handoff.Root{
		"lead": {Repo: "../lead", Pin: "p-v1.0.0"},
		"pm":   {Repo: "../spec", Pin: "release/p-v1.0"},
	}
	p, err := handoff.NewPlan(r.dev, slice("a"), "")
	if err != nil {
		t.Fatal(err)
	}
	bad := map[string]handoff.Root{"lead": roots["lead"], "pm": {Repo: "../spec", Pin: "release/none"}}
	if err := p.PlanPins(bad, []string{"lead", "pm"}); err == nil {
		t.Fatal("a pin that does not exist was planned")
	}
	if _, err := os.Stat(filepath.Join(lead, ".worktrees")); !os.IsNotExist(err) {
		t.Errorf("planning made a worktree (%v)", err)
	}

	if err := p.PlanPins(roots, []string{"lead", "pm"}); err != nil {
		t.Fatal(err)
	}
	if err := p.Create(); err != nil {
		t.Fatal(err)
	}
	paths := p.PinPaths()
	if got := gitT(t, paths["lead"], "rev-parse", "HEAD"); got != leadPin {
		t.Errorf("lead pinned at %s, want %s", got, leadPin)
	}
	if got := gitT(t, paths["pm"], "rev-parse", "HEAD"); got != r.release {
		t.Errorf("pm pinned at %s, want %s", got, r.release)
	}
	if errs := p.Rollback(); len(errs) != 0 {
		t.Fatalf("rollback: %v", errs)
	}
	for _, path := range []string{paths["lead"], paths["pm"], p.Worktree} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s remains (%v)", path, err)
		}
	}
}

func TestPlanPinsRefusals(t *testing.T) {
	r := newRepos(t)
	p, err := handoff.NewPlan(r.dev, slice("a"), "")
	if err != nil {
		t.Fatal(err)
	}
	// a tag works as a pin
	if err := p.PlanPins(map[string]handoff.Root{"pm": {Repo: "../spec", Pin: "p-v1.1.0"}}, []string{"pm"}); err != nil || p.Pins[0].Commit != r.tagged {
		t.Fatalf("tag pin: %+v %v", p.Pins, err)
	}
	if err := os.MkdirAll(filepath.Join(r.spec, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		roots map[string]handoff.Root
		want  string
	}{
		"no such pin":      {map[string]handoff.Root{"pm": {Repo: "../spec", Pin: "release/none"}}, "not a branch or tag"},
		"not a repository": {map[string]handoff.Root{"pm": {Repo: "../nowhere", Pin: "main"}}, "not the root of a git repository"},
		"a subdirectory":   {map[string]handoff.Root{"pm": {Repo: "../spec/sub", Pin: "main"}}, "not the root of a git repository"},
		"one repository":   {map[string]handoff.Root{"a": {Repo: "../spec", Pin: "main"}, "b": {Repo: "../spec", Pin: "p-v1.1.0"}}, "one repository"},
	} {
		names := make([]string, 0, len(c.roots))
		for n := range c.roots {
			names = append(names, n)
		}
		if len(names) == 2 && names[0] > names[1] {
			names[0], names[1] = names[1], names[0]
		}
		if err := p.PlanPins(c.roots, names); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", name, err, c.want)
		}
	}
	// a pinned worktree in the way
	if err := os.MkdirAll(filepath.Join(r.spec, ".worktrees", "spec-a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := p.PlanPins(map[string]handoff.Root{"pm": {Repo: "../spec", Pin: "main"}}, []string{"pm"}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("pin in the way: %v", err)
	}
}

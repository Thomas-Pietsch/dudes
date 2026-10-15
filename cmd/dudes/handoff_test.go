package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// handoffFixture is an engine-ready dev repository (a tracked dudes/ folder) beside the spec
// repository it pins: the bench's spec documents on a release branch, and a later commit on main
// that the pin must not see.
type handoffFixture struct {
	parent, dev, spec string
	release           string // the commit of the spec's release branch
	cfg               string // the user's own dudes.json
	notReady, ready   string // the dev repo's commits before and with its dudes/ folder; HEAD is one later
}

const handoffPin = "release/bench-v1.0"

func newHandoffFixture(t *testing.T) *handoffFixture {
	t.Helper()
	// the machine's model host comes from the environment (set per test), never from a file here
	t.Setenv(config.OllamaURLEnv, "http://model-host.test:11434")
	t.Setenv("OLLAMA_API_BASE", "")

	parent := t.TempDir()
	f := &handoffFixture{parent: parent, dev: filepath.Join(parent, "dev"), spec: filepath.Join(parent, "spec"), cfg: filepath.Join(parent, "user.json")}
	write := func(dir, rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	initRepo := func(dir string) {
		t.Helper()
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		gitIn(t, dir, "init", "-q", "-b", "main")
		gitIn(t, dir, "config", "user.email", "dudes@test")
		gitIn(t, dir, "config", "user.name", "dudes-test")
		gitIn(t, dir, "config", "core.autocrlf", "false")
	}
	bench := filepath.Join("..", "..", "bench", "workspace")
	read := func(rel string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(bench, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}

	initRepo(f.spec)
	for _, name := range []string{"BENCH_PRD.md", "BENCH_TA.md", "BENCH_QR.md", "BENCH_BACKLOG.md", "BENCH_ROADMAP.md"} {
		write(f.spec, name, read(name))
	}
	gitIn(t, f.spec, "add", "-A")
	gitIn(t, f.spec, "commit", "-qm", "spec 1.0")
	f.release = gitIn(t, f.spec, "rev-parse", "HEAD")
	gitIn(t, f.spec, "branch", handoffPin)
	write(f.spec, "BENCH_PRD.md", read("BENCH_PRD.md")+"\nA later spec change.\n")
	gitIn(t, f.spec, "commit", "-qam", "a change after the release")

	initRepo(f.dev)
	write(f.dev, "go.mod", "module bench\n\ngo 1.22\n")
	// today's package: it compiles, and the behaviour is not there
	write(f.dev, "semver/semver.go", "package semver\n\ntype Version struct{}\n\nfunc Parse(s string) (Version, error) { return Version{}, nil }\n\nfunc Compare(a, b string) (int, error) { return 0, nil }\n")
	// what the slice's tests must not break: a package that calls the one being built
	write(f.dev, "api/api.go", "package api\n\nimport \"bench/semver\"\n\nfunc Newer(a, b string) bool { c, _ := semver.Compare(a, b); return c > 0 }\n")
	write(f.dev, "api/api_test.go", "package api\n\nimport \"testing\"\n\nfunc TestEqualIsNotNewer(t *testing.T) {\n\tif Newer(\"1.0.0\", \"1.0.0\") {\n\t\tt.Fatal(\"equal versions\")\n\t}\n}\n")
	gitIn(t, f.dev, "add", "-A")
	gitIn(t, f.dev, "commit", "-qm", "code, not yet engine-ready")
	f.notReady = gitIn(t, f.dev, "rev-parse", "HEAD")
	write(f.dev, "dudes/profile.json", read("dudes/profile.json"))
	write(f.dev, "dudes/specset.json", `{
  "product": "bench",
  "roots": {"spec": {"repo": "../spec", "pin": "`+handoffPin+`"}},
  "entries": [
    {"id": "PRD", "kind": "prd", "root": "spec", "path": "BENCH_PRD.md"},
    {"id": "TA", "kind": "ta", "root": "spec", "path": "BENCH_TA.md"},
    {"id": "QR", "kind": "qr", "root": "spec", "path": "BENCH_QR.md"},
    {"id": "BACKLOG", "kind": "backlog", "root": "spec", "path": "BENCH_BACKLOG.md"},
    {"id": "ROADMAP", "kind": "roadmap", "root": "spec", "path": "BENCH_ROADMAP.md"}
  ]
}
`)
	// a repository's config names the worker and the gate; the model host it names is the machine's
	write(f.dev, "dudes/config.json", `{
  "product": "bench",
  "ollama": {"model": "m", "base_url": "http://repo-host.invalid:11434"},
  "aider": {"model": "ollama/m"},
  "test_command": ["go", "test", "./..."]
}
`)
	gitIn(t, f.dev, "add", "-A")
	gitIn(t, f.dev, "commit", "-qm", "engine-ready")
	f.ready = gitIn(t, f.dev, "rev-parse", "HEAD")
	write(f.dev, "later.txt", "work that came after\n")
	gitIn(t, f.dev, "add", "-A")
	gitIn(t, f.dev, "commit", "-qm", "later work")

	// the user's own dudes.json: machine values, and a product that is not this repository's
	write(f.parent, "user.json", `{"workspace": ".", "product": "demo", "aider": {"launch": ["sh", "../../../scripts/worker.sh"]}}`)
	return f
}

// slicing writes a slicing file for the fixture's item BL-002 and returns its path; mutate edits it first.
func (f *handoffFixture) slicing(t *testing.T, id string, mutate func(*handoff.Slicing)) string {
	t.Helper()
	s := handoff.Slicing{
		ID: id, Item: "BL-002", Acceptance: []string{"AC-1", "AC-3"},
		Scope: []string{"semver/*"}, Tests: []string{"semver/bl002_test.go"}, Hooks: []string{"semver/hook.go"},
		TestCommand: []string{"go", "test", "-count=1", "./semver/..."},
		GateCommand: []string{"go", "test", "-count=1", "./semver/...", "./api/..."},
		Tools:       []string{"go.mod"}, Refs: []string{"TA:§3"},
		Constraints: []string{"`semver` exports `Compare`: its signature stays."}, Later: "AC-2",
	}
	if mutate != nil {
		mutate(&s)
	}
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(f.parent, id+".slicing.json")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func (f *handoffFixture) worktree(id string) string { return filepath.Join(f.dev, ".worktrees", id) }

// worktrees counts the worktrees git lists for the fixture's two repositories, the main ones included.
func (f *handoffFixture) worktrees(t *testing.T) [2]int {
	t.Helper()
	var n [2]int
	for i, repo := range []string{f.dev, f.spec} {
		n[i] = strings.Count(gitIn(t, repo, "worktree", "list", "--porcelain"), "worktree ")
	}
	return n
}

// nothingMade asserts that a refused or failed preparation of slice id made no worktree, pin or
// branch: git lists as many worktrees as before and the slice's directories and branch are absent.
func (f *handoffFixture) nothingMade(t *testing.T, id string, before [2]int) {
	t.Helper()
	for _, p := range []string{f.worktree(id), filepath.Join(f.spec, ".worktrees", "spec-"+id)} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s was made (%v)", p, err)
		}
	}
	if got := f.worktrees(t); got != before {
		t.Errorf("worktrees listed %v, were %v", got, before)
	}
	if out := gitIn(t, f.dev, "branch", "--list", "feature/"+id); out != "" {
		t.Errorf("branch left: %s", out)
	}
}

// One command makes a slice ready: its worktree on a feature branch, the spec pinned at its release,
// the contract with a red-test task and a code task, the freeze — and no placeholder test.
func TestPrepareSlice(t *testing.T) {
	f := newHandoffFixture(t)
	base := gitIn(t, f.dev, "rev-parse", "HEAD")

	res, err := prepareSlice(f.cfg, f.dev, sliceArgs{item: "bl-002", slicing: f.slicing(t, "bl-002a", nil)})
	if err != nil {
		t.Fatal(err)
	}
	wt := f.worktree("bl-002a")

	// the worktree on its branch, started from the base; the run's start is a commit of its own
	if got := gitIn(t, wt, "branch", "--show-current"); got != "feature/bl-002a" {
		t.Errorf("branch %q", got)
	}
	if got := gitIn(t, wt, "rev-parse", "HEAD~1"); got != base {
		t.Errorf("the branch starts at %s, want the base %s", got, base)
	}
	head := gitIn(t, wt, "log", "-1", "--format=%B")
	if !strings.HasPrefix(head, "chore(run): start bl-002a [🔹engine]") || !strings.Contains(head, "\nItem: BL-002\n") {
		t.Errorf("start commit:\n%s", head)
	}
	if gitIn(t, wt, "status", "--porcelain") != "" || gitIn(t, f.dev, "status", "--porcelain") != "" || gitIn(t, f.spec, "status", "--porcelain") != "" {
		t.Error("the engine's folders show up in git status")
	}

	// the spec root pinned at the release: detached, at the release commit — not the later one on main
	pin := filepath.Join(f.spec, ".worktrees", "spec-bl-002a")
	if got := gitIn(t, pin, "rev-parse", "HEAD"); got != f.release {
		t.Errorf("pin at %s, want the release %s", got, f.release)
	}
	if got := gitIn(t, pin, "rev-parse", "--abbrev-ref", "HEAD"); got != "HEAD" {
		t.Errorf("the pin is on branch %q, want detached", got)
	}
	if data, _ := os.ReadFile(filepath.Join(pin, "BENCH_PRD.md")); strings.Contains(string(data), "A later spec change.") {
		t.Error("the pinned spec holds a change made after the release")
	}
	set, ok, err := specset.Load(wt, "bench")
	if err != nil || !ok {
		t.Fatalf("specset: %v %v", ok, err)
	}
	if root, err := set.RootAbs(wt, "spec"); err != nil || !handoff.SameDir(root, pin) {
		t.Errorf("the run's spec-set root is %q (%v), want the pin %s", root, err, pin)
	}

	// the run config: the user's machine values plus the repository's settings plus the workspace
	cfg, ok, err := config.Load(filepath.Join(wt, ".dudes", "dudes.json"))
	if err != nil || !ok {
		t.Fatalf("run config: %v %v", ok, err)
	}
	if cfg.Workspace != filepath.ToSlash(wt) || cfg.Product != "bench" || cfg.Worktree {
		t.Errorf("workspace %q product %q worktree %v", cfg.Workspace, cfg.Product, cfg.Worktree)
	}
	if cfg.Ollama.BaseURL != "http://model-host.test:11434" || cfg.Ollama.Model != "m" || cfg.Aider.Model != "ollama/m" {
		t.Errorf("ollama %+v aider %+v", cfg.Ollama, cfg.Aider)
	}
	if len(cfg.Aider.Launch) != 2 || cfg.Aider.Launch[0] != "sh" {
		t.Errorf("the user's launcher was lost: %v", cfg.Aider.Launch)
	}

	// the contract: the item narrowed to the slice's criteria, two tasks, facts only
	con, ok, err := contract.Load(wt)
	if err != nil || !ok {
		t.Fatalf("contract: %v %v", ok, err)
	}
	if con.Item != "BL-002" || con.ID != "bl-002a" || len(con.Acceptance) != 2 || con.Acceptance[0].ID != "bl-002-ac-1" || con.Acceptance[1].ID != "bl-002-ac-3" {
		t.Errorf("contract %s item %s acceptance %+v", con.ID, con.Item, con.Acceptance)
	}
	if len(con.Tasks) != 2 || len(con.TasksFor(phase.TestAuthoring)) != 0 {
		t.Fatalf("tasks %+v: want the red-test and the code task, no test_authoring stub task", con.Tasks)
	}
	red, code := con.Tasks[0], con.Tasks[1]
	if red.ID != "bl-002a-tests" || red.Phase != phase.Implementation || red.Expect != "fail" || red.Approval != "after" {
		t.Errorf("red task %+v", red)
	}
	if code.ID != "bl-002a-code" || code.Phase != phase.Implementation || code.Expect != "" || code.Approval != "after" {
		t.Errorf("code task %+v", code)
	}
	wantRed := "BL-002a, red first — the acceptance tests of BL-002 AC-1 and AC-3, not the code: `semver/bl002_test.go` with the test functions `TestBL002AC1` and `TestBL002AC3`. " +
		"It also writes the hook file `semver/hook.go`. The tests compile against today's package and fail until the code is written; the next task cannot change them."
	wantCode := "BL-002a — the code: make `semver/bl002_test.go` pass. It and `semver/hook.go` are written and read-only."
	if red.Title != wantRed {
		t.Errorf("red title:\n%s\nwant:\n%s", red.Title, wantRed)
	}
	if code.Title != wantCode {
		t.Errorf("code title:\n%s\nwant:\n%s", code.Title, wantCode)
	}
	if strings.Join(red.Writable, " ") != "semver/bl002_test.go semver/hook.go" {
		t.Errorf("red writable %v: the slice's tests plus hooks", red.Writable)
	}
	if strings.Join(code.Writable, " ") != "semver/* !semver/bl002_test.go !semver/hook.go" {
		t.Errorf("code writable %v: the scope minus tests and hooks", code.Writable)
	}
	if strings.Join(red.Test, " ") != "go test -count=1 ./semver/..." || strings.Join(code.Test, " ") != "go test -count=1 ./semver/... ./api/..." {
		t.Errorf("gates: red %v, code %v", red.Test, code.Test)
	}
	wantCons := []string{"In `semver/*` only; later slices: AC-2.", "`semver` exports `Compare`: its signature stays.", "Do not rename instance documents."}
	if strings.Join(con.Constraints, "\n") != strings.Join(wantCons, "\n") {
		t.Errorf("constraints %q", con.Constraints)
	}
	if strings.Join(con.Tests, " ") != "semver/bl002_test.go" {
		t.Errorf("tests %v", con.Tests)
	}
	// the model host is the environment's; the repository's own, in its config, is read nowhere
	_ = filepath.WalkDir(filepath.Join(wt, ".dudes"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if data, _ := os.ReadFile(p); strings.Contains(string(data), "repo-host.invalid") {
				t.Errorf("%s carries the repository's model host", p)
			}
		}
		return nil
	})

	// no placeholder under tests/: the run starts where its first task is
	if _, err := os.Stat(filepath.Join(wt, "tests")); !os.IsNotExist(err) {
		t.Errorf("tests/ exists (%v)", err)
	}
	st, ok, err := runstate.Load(wt)
	if err != nil || !ok || st.Phase != phase.Implementation || st.Prompt != "BL-002a" {
		t.Fatalf("run %+v %v %v", st, ok, err)
	}
	if _, ok, _ := runstate.Load(f.dev); ok {
		t.Error("the run state is visible from the main checkout")
	}

	// the freeze records the pinned spec, and implementation can be entered
	fr, ok, err := freeze.Load(wt)
	if err != nil || !ok || len(fr.Inputs) == 0 {
		t.Fatalf("freeze %+v %v %v", fr, ok, err)
	}
	for _, in := range fr.Inputs {
		if in.ID == "prd" && in.Revision != f.release {
			t.Errorf("the PRD froze at %s, want the release %s", in.Revision, f.release)
		}
	}
	if err := gate.CheckEnter(wt, "bench", phase.Implementation); err != nil {
		t.Errorf("implementation cannot be entered: %v", err)
	}

	var out bytes.Buffer
	res.print(&out)
	for _, want := range []string{"prepared BL-002a", ".worktrees/bl-002a on feature/bl-002a", "spec pinned at " + handoffPin, "bl-002a-tests (red first)",
		"next: dudes -config .worktrees/bl-002a/.dudes/dudes.json step\n"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "model-host.test") {
		t.Error("the model host was printed")
	}
}

// -base names where the slice starts — a commit with the repository's dudes/ folder, not one without.
// A slice is prepared once: its worktree, its branch or a pin in the way refuses before anything is
// made; so does a command run anywhere but the repository's root, or naming another item.
func TestPrepareSliceBaseAndRefusals(t *testing.T) {
	f := newHandoffFixture(t)
	before := f.worktrees(t)
	if _, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, "old", nil), base: f.notReady}); err == nil || !strings.Contains(err.Error(), "not engine-ready") {
		t.Fatalf("a base without dudes/: %v", err)
	}
	f.nothingMade(t, "old", before)

	args := sliceArgs{slicing: f.slicing(t, "bl-002a", nil), base: f.ready}
	if _, err := prepareSlice(f.cfg, f.dev, args); err != nil {
		t.Fatal(err)
	}
	if got := gitIn(t, f.worktree("bl-002a"), "rev-parse", "HEAD~1"); got != f.ready {
		t.Errorf("the slice starts at %s, want -base %s", got, f.ready)
	}
	if _, err := os.Stat(filepath.Join(f.worktree("bl-002a"), "later.txt")); !os.IsNotExist(err) {
		t.Errorf("work after the base is in the slice (%v)", err)
	}
	st, _, _ := runstate.Load(f.worktree("bl-002a"))
	_, err := prepareSlice(f.cfg, f.dev, args)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second preparation: %v", err)
	}
	if again, ok, _ := runstate.Load(f.worktree("bl-002a")); !ok || again.RunID != st.RunID {
		t.Error("the refused preparation touched the first slice's run")
	}

	before = f.worktrees(t)
	gitIn(t, f.dev, "branch", "feature/taken")
	if _, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, "taken", nil)}); err == nil || !strings.Contains(err.Error(), "branch feature/taken already exists") {
		t.Fatalf("branch in the way: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(f.spec, ".worktrees", "spec-blocked"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, "blocked", nil)}); err == nil || !strings.Contains(err.Error(), "pinned worktree") {
		t.Fatalf("pin in the way: %v", err)
	}
	if _, err := prepareSlice(f.cfg, filepath.Join(f.dev, "semver"), sliceArgs{slicing: f.slicing(t, "sub", nil)}); err == nil || !strings.Contains(err.Error(), "dev repository's root") {
		t.Fatalf("from a subdirectory: %v", err)
	}
	if _, err := prepareSlice(f.cfg, f.dev, sliceArgs{item: "BL-003", slicing: f.slicing(t, "other", nil)}); err == nil || !strings.Contains(err.Error(), "BL-003") {
		t.Fatalf("another item than the slicing's: %v", err)
	}
	for _, id := range []string{"taken", "blocked", "sub", "other"} {
		if _, err := os.Stat(f.worktree(id)); !os.IsNotExist(err) {
			t.Errorf("a worktree was made for %s (%v)", id, err)
		}
	}
	if got := f.worktrees(t); got != before {
		t.Errorf("worktrees listed %v, were %v", got, before)
	}
}

// No model host from the environment or the user's own dudes.json: the repository names one in its
// config, and it is not a source — the command fails clearly, having made nothing. (With a host set,
// TestPrepareSlice shows the run config carries that one and the repository's nowhere.)
func TestPrepareSliceNeverTakesTheModelHostFromTheRepo(t *testing.T) {
	f := newHandoffFixture(t) // the repository's config names http://repo-host.invalid:11434
	t.Setenv(config.OllamaURLEnv, "")
	before := f.worktrees(t)
	_, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, "host-none", nil)})
	if err == nil || !strings.Contains(err.Error(), config.OllamaURLEnv) || strings.Contains(err.Error(), "repo-host.invalid") {
		t.Fatalf("no model host: %v", err)
	}
	f.nothingMade(t, "host-none", before)
}

// A preparation that fails after it made the worktree takes everything back: the slice can be
// prepared again once the slicing is right.
func TestPrepareSliceTakesBackWhatItMade(t *testing.T) {
	f := newHandoffFixture(t)
	for _, c := range []struct {
		id, want string
		mutate   func(*handoff.Slicing)
	}{
		{"bad-ac", "has no criterion AC-9 (it has AC-1, AC-2, AC-3)", func(s *handoff.Slicing) { s.Acceptance = []string{"AC-1", "AC-9"} }},
		{"bad-ref", "TA:§99 is not a reference of BL-002", func(s *handoff.Slicing) { s.Refs = []string{"TA:§99"} }},
		{"bad-item", "work item BL-777 not found", func(s *handoff.Slicing) { s.Item = "BL-777" }},
	} {
		before := f.worktrees(t)
		_, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, c.id, c.mutate)})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", c.id, err, c.want)
		}
		f.nothingMade(t, c.id, before)
	}
}

// -slice switches intake to item handoff; without it every word is the prompt, -base included
// (`dudes do` without -slice stays F84: TestIntake runs through the same dispatch).
func TestParseSliceArgs(t *testing.T) {
	for _, c := range []struct {
		words []string
		want  sliceArgs
		slice bool
		fail  bool
	}{
		{[]string{"X-001", "-slice", "s.json"}, sliceArgs{item: "X-001", slicing: "s.json"}, true, false},
		{[]string{"-slice=s.json", "-base", "origin/development", "X-001"}, sliceArgs{item: "X-001", slicing: "s.json", base: "origin/development"}, true, false},
		{[]string{"--slice", "s.json"}, sliceArgs{slicing: "s.json"}, true, false},
		{[]string{"X-003"}, sliceArgs{}, false, false},
		{[]string{"-pick", "sort", "versions"}, sliceArgs{}, false, false},
		{[]string{"-base", "x", "something"}, sliceArgs{}, false, false},
		{[]string{"X-001", "-slice"}, sliceArgs{}, true, true},
		{[]string{"X-001", "X-002", "-slice", "s.json"}, sliceArgs{}, true, true},
	} {
		got, slice, err := parseSliceArgs(c.words)
		if (err != nil) != c.fail || slice != c.slice || (!c.fail && got != c.want) {
			t.Errorf("%v: got %+v slice=%v err=%v, want %+v slice=%v fail=%v", c.words, got, slice, err, c.want, c.slice, c.fail)
		}
	}
}

// fakeWorker stands in for Aider: it reads the task from the message file and writes what the task
// asks — and, like a worker that cannot be trusted, files it may not touch.
const fakeWorker = `#!/bin/sh
while [ $# -gt 0 ]; do
	if [ "$1" = "--message-file" ]; then msg=$2; shift; fi
	shift
done
mkdir -p semver
if grep -q 'red first' "$msg"; then
cat > semver/bl002_test.go <<'EOF'
package semver

import "testing"

func TestBL002AC1(t *testing.T) {
	if got, _ := Compare("v1.10.0", "1.9.3"); got != 1 {
		t.Fatalf("Compare = %d, want 1", got)
	}
}

func TestBL002AC3(t *testing.T) {
	if _, err := Parse("1.x.0"); err == nil {
		t.Fatal("no error for 1.x.0")
	}
}
EOF
printf 'package semver\n\nvar hook = func() {}\n' > semver/hook.go
printf 'package semver\n' > semver/stray.go
else
cat > semver/semver.go <<'EOF'
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct{ Major, Minor, Patch int }

func Parse(s string) (Version, error) {
	p := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(p) != 3 {
		return Version{}, fmt.Errorf("ERR_VERSION: %q", s)
	}
	var n [3]int
	for i, x := range p {
		v, err := strconv.Atoi(x)
		if err != nil {
			return Version{}, fmt.Errorf("ERR_VERSION: %q", s)
		}
		n[i] = v
	}
	return Version{n[0], n[1], n[2]}, nil
}

func Compare(a, b string) (int, error) {
	x, err := Parse(a)
	if err != nil {
		return 0, err
	}
	y, err := Parse(b)
	if err != nil {
		return 0, err
	}
	for _, d := range []int{x.Major - y.Major, x.Minor - y.Minor, x.Patch - y.Patch} {
		if d != 0 {
			return map[bool]int{true: 1, false: -1}[d > 0], nil
		}
	}
	return 0, nil
}
EOF
printf 'tampered\n' >> semver/bl002_test.go
printf 'tampered\n' >> semver/hook.go
fi
echo done
`

// The prepared slice runs through the engine's own steps: the red-first task writes the tests and a
// hook and pauses for a human, the code task cannot change them, and the acceptance runs are recorded
// for the slice's criteria — with a worker that stands in for Aider.
func TestSliceRunsThroughTheEngine(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	f := newHandoffFixture(t)
	script := filepath.Join(t.TempDir(), "worker.sh")
	if err := os.WriteFile(script, []byte(fakeWorker), 0o755); err != nil {
		t.Fatal(err)
	}
	user := filepath.Join(f.parent, "worker.json")
	if err := os.WriteFile(user, []byte(`{"aider": {"launch": ["sh", "`+filepath.ToSlash(script)+`"], "timeout_sec": 300}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareSlice(user, f.dev, sliceArgs{slicing: f.slicing(t, "bl-002a", nil)}); err != nil {
		t.Fatal(err)
	}
	wt := f.worktree("bl-002a")
	cfgPath := filepath.Join(wt, ".dudes", "dudes.json")
	changed := func() string {
		return strings.Join(strings.Fields(gitIn(t, wt, "show", "--name-only", "--format=", "HEAD")), " ")
	}

	// the red-first task: tests and hook written, the stray file discarded, the gate red as expected, a pause
	if code := cmdStep(cfgPath, false, false, false); code != 0 {
		t.Fatalf("step 1: exit %d", code)
	}
	st, _, _ := runstate.Load(wt)
	if st.Status != runstate.StatusWaitingDecision || st.TaskID != "bl-002a-tests" || !st.HasCompleted("bl-002a-tests") {
		t.Fatalf("after the red task: %+v", st)
	}
	if got := changed(); got != "semver/bl002_test.go semver/hook.go" {
		t.Errorf("the red task committed %q", got)
	}
	if _, err := os.Stat(filepath.Join(wt, "semver", "stray.go")); !os.IsNotExist(err) {
		t.Errorf("a file outside the red task's scope was kept (%v)", err)
	}
	if !strings.Contains(gitIn(t, wt, "log", "-1", "--format=%B"), "Item: BL-002") {
		t.Error("the worker's commit does not name the item")
	}
	if code := cmdDecide(cfgPath, []string{"bl-002a-tests"}); code != 0 {
		t.Fatalf("decide: exit %d", code)
	}

	// the code task: the tests and the hook stay as written, the gate is the slice's and green
	tests, _ := os.ReadFile(filepath.Join(wt, "semver", "bl002_test.go"))
	if code := cmdStep(cfgPath, false, false, false); code != 0 {
		t.Fatalf("step 2: exit %d", code)
	}
	st, _, _ = runstate.Load(wt)
	if st.Status != runstate.StatusWaitingDecision || st.TaskID != "bl-002a-code" || !st.HasCompleted("bl-002a-code") {
		t.Fatalf("after the code task: %+v", st)
	}
	if got := changed(); got != "semver/semver.go" {
		t.Errorf("the code task committed %q, want the code alone", got)
	}
	if now, _ := os.ReadFile(filepath.Join(wt, "semver", "bl002_test.go")); string(now) != string(tests) {
		t.Errorf("the code task changed the tests:\n%s", now)
	}
	if hook, _ := os.ReadFile(filepath.Join(wt, "semver", "hook.go")); strings.Contains(string(hook), "tampered") {
		t.Error("the code task changed the hook")
	}
	acc, err := testrun.Acceptance(wt)
	if err != nil || acc["TestBL002AC1"].Result != testrun.Pass || acc["TestBL002AC3"].Result != testrun.Pass || len(acc) != 2 {
		t.Errorf("acceptance runs %v (%v): want the slice's two criteria, passing", acc, err)
	}
	// the code worker was told the tests are read-only: they are named, though they live beside the code
	msg, _ := os.ReadFile(filepath.Join(wt, ".dudes", "worker_message.md"))
	if want := "The acceptance tests are written and read-only: `semver/bl002_test.go`"; !strings.Contains(string(msg), want) {
		t.Errorf("the code task's message lacks %q", want)
	}
}

// A contract that names its tests (F88) keeps the engine from waiting for them before its first task
// and makes the worker's test files count: red-first tasks edit them, code tasks read them. The run
// starts at implementation.
func TestContractNamingTests(t *testing.T) {
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	con := &contract.Contract{ID: "s", Product: "p", Item: "X1-001", Tests: []string{"pkg/x_test.go"},
		Acceptance: []contract.Acceptance{{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1 it works"}},
		Tasks: []contract.Task{
			{ID: "s-tests", Phase: phase.Implementation, Expect: "fail", Writable: []string{"pkg/x_test.go", "pkg/hook.go"}},
			{ID: "s-code", Phase: phase.Implementation, Writable: []string{"pkg/*", "!pkg/x_test.go", "!pkg/hook.go"}},
		}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws, Product: "p"}

	// before the red-first task ran: no test exists, and the gate does not ask for one
	if err := gate.RequireTests(ws, "p"); err != nil {
		t.Errorf("RequireTests before the red task: %v", err)
	}
	if edit, ro := taskTests(cfg, []string{"pkg/x_test.go", "pkg/hook.go"}, nil); len(edit)+len(ro) != 0 {
		t.Errorf("tests that do not exist yet: %v %v", edit, ro)
	}

	// the red-first task wrote the test, beside the code
	if err := os.MkdirAll(filepath.Join(ws, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "pkg", "x_test.go"), []byte("package pkg\nimport \"testing\"\nfunc TestX1001AC1(t *testing.T) { t.Fatal(\"red\") }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edit, ro := taskTests(cfg, con.Tasks[0].Writable, nil) // the red-first task edits it
	if len(edit) != 1 || edit[0] != "pkg/x_test.go" || len(ro) != 0 {
		t.Errorf("red task: edit %v read-only %v", edit, ro)
	}
	edit, ro = taskTests(cfg, con.Tasks[1].Writable, nil) // the code task reads it
	if len(edit) != 0 || len(ro) != 1 || ro[0] != "pkg/x_test.go" {
		t.Errorf("code task: edit %v read-only %v", edit, ro)
	}
	if scope.Inside("pkg/x_test.go", con.Tasks[1].Writable) || !scope.Inside("pkg/code.go", con.Tasks[1].Writable) {
		t.Errorf("code task scope %v", con.Tasks[1].Writable)
	}
	// a gate log that names the test by its bare file name hands a worker the files it may edit:
	// the red task its test, the code task none — the test is outside its scope
	if got := inScope(ws, "x_test.go", con.Tasks[0].Writable); len(got) != 1 || got[0] != "pkg/x_test.go" {
		t.Errorf("red task: files to edit %v", got)
	}
	if got := inScope(ws, "x_test.go", con.Tasks[1].Writable); len(got) != 0 {
		t.Errorf("code task: files to edit %v, want none", got)
	}

	// init starts where the first task is — no test_authoring, no placeholder under tests/
	cfgPath := filepath.Join(ws, ".dudes", "dudes.json")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cfgPath, []byte(`{"workspace":"`+filepath.ToSlash(ws)+`","product":"p"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code := cmdInit(cfgPath, "s"); code != 0 {
		t.Fatalf("init: %d", code)
	}
	if st, ok, err := runstate.Load(ws); err != nil || !ok || st.Phase != phase.Implementation {
		t.Fatalf("run %+v %v %v", st, ok, err)
	}
	if _, err := os.Stat(filepath.Join(ws, "tests")); !os.IsNotExist(err) {
		t.Errorf("tests/ exists (%v)", err)
	}

	// without named tests the same gate asks for tests, as before
	con.Tests = nil
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(ws, "pkg", "x_test.go")); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireTests(ws, "p"); err == nil {
		t.Error("RequireTests passes with no test at all and none named")
	}
}

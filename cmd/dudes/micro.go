package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/goals"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/recipe"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/testids"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// Micro-agent context (CORE §2.4): a worker gets its task, its goals and the output contract.
// The mechanisms here replace rules a worker would otherwise have to read.

// applyGoals puts the recipe's goals — their `## ` lines from the makefile, nothing else — into
// the phase instructions (F75). A goal the makefile does not document stops the gate.
func applyGoals(cfg config.Config, rec *recipe.Recipe) error {
	if rec == nil || len(rec.Goals) == 0 {
		return nil
	}
	all, err := goals.List(cfg.Workspace)
	if err != nil {
		return err
	}
	sel, err := goals.Select(all, rec.Goals)
	if err != nil {
		return fmt.Errorf("recipe %s: %w", recipe.Path(rec.Phase), err)
	}
	rec.Instructions = strings.TrimSpace(strings.TrimSpace(rec.Instructions) + "\n\n" + goals.Block(sel))
	return nil
}

// bindScope writes the bound work item into work tree ws for the git hooks: its ID (commit-msg
// requires the `Item:` trailer, F77) and its writable paths (pre-commit, F78). The returned
// func clears the scope again once the worker's commits are done.
func bindScope(cfg config.Config, ws string, st *runstate.State) (func(), error) {
	none := func() {}
	con, ok, err := contract.Load(cfg.Workspace)
	pats := writable(con, st)
	if err != nil || !ok || (con.Item == "" && len(pats) == 0) {
		return none, err
	}
	if err := scope.Write(ws, con.Item, pats); err != nil {
		return none, err
	}
	fmt.Printf("scope: item %s, %d writable pattern(s) → %s\n", nonzero(con.Item, "-"), len(pats), scope.Path(ws))
	return func() { _ = scope.Remove(ws) }, nil
}

// writable is the current task's scope: its own patterns narrow the contract's (F78).
func writable(con *contract.Contract, st *runstate.State) []string {
	if con == nil {
		return nil
	}
	if st != nil {
		if tk, ok := con.TaskByID(st.TaskID); ok && len(tk.Writable) > 0 {
			return tk.Writable
		}
	}
	return con.Writable
}

// taskTests are the authored test files of the task: to edit — the bound item's acceptance test
// files inside the task's scope (one per criterion), else the test phase's primary artifact — and
// read-only — the item's test files outside the scope that the task's gate runs (a red-first test
// the implementation task makes pass and cannot change). A contract that names its test files
// (F88) gives those: the ones that exist, to edit inside the scope, read-only outside it.
func taskTests(cfg config.Config, pats, gate []string) (edit, readOnly []string) {
	if con, ok, _ := contract.Load(cfg.Workspace); ok && con.NamesTests() {
		for _, f := range con.ExistingTests(cfg.Workspace) {
			if scope.Inside(f, pats) {
				edit = append(edit, f)
			} else {
				readOnly = append(readOnly, f)
			}
		}
		return edit, readOnly
	}
	if _, acs, err := itemAcceptance(cfg); err == nil {
		for _, ac := range acs {
			f := spec.ItemTestFile(ac.ID)
			if !artifact.Exists(cfg.Workspace, f) {
				continue
			}
			if scope.Inside(f, pats) {
				edit = append(edit, f)
			} else if gateSelects(gate, spec.AcceptanceTestName(ac.ID)) {
				readOnly = append(readOnly, f)
			}
		}
	}
	if len(edit) > 0 || len(readOnly) > 0 {
		return edit, readOnly
	}
	if rel, ok := artifact.Path(cfg.Product, phase.TestAuthoring); ok && artifact.Exists(cfg.Workspace, rel) {
		return []string{rel}, nil
	}
	return nil, nil
}

// gateSelects reports whether the gate command runs test name: its -run pattern matches (if any)
// and its -skip pattern does not. A command without either runs everything.
func gateSelects(gate []string, name string) bool {
	flag := func(f string) (*regexp.Regexp, bool) {
		for i, a := range gate {
			v, ok := "", false
			switch {
			case a == f && i+1 < len(gate):
				v, ok = gate[i+1], true
			case strings.HasPrefix(a, f+"="):
				v, ok = strings.TrimPrefix(a, f+"="), true
			}
			if ok {
				re, err := regexp.Compile(v)
				return re, err == nil
			}
		}
		return nil, false
	}
	if re, ok := flag("-run"); ok && !re.MatchString(name) {
		return false
	}
	if re, ok := flag("-skip"); ok && re.MatchString(name) {
		return false
	}
	return true
}

// acceptanceGate runs after the unit tests passed, when the instance profile declares its
// acceptance tests (`acceptance_tests`): first their goal (`make <goal>`, e.g. e2e) as a hard gate
// (F35), recorded under refs/notes/dudes/e2e; then the ID check (F77) — report only unless `gate`.
// No profile or no `acceptance_tests` → nothing to run.
func acceptanceGate(cfg config.Config, st *runstate.State) int {
	data, err := os.ReadFile(filepath.Join(cfg.Workspace, filepath.FromSlash(itempack.ProfilePath())))
	if err != nil {
		return 0
	}
	prof, err := itempack.ParseProfile(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: acceptance gate skipped: %v\n", err)
		return 0
	}
	at := prof.AcceptanceTests
	if code := runAcceptanceGoal(cfg, st, at.Goal); code != 0 {
		return code
	}
	return checkTestIDs(cfg, st, at)
}

func runAcceptanceGoal(cfg config.Config, st *runstate.State, goal string) int {
	if goal == "" {
		return 0
	}
	if !goals.Has(cfg.Workspace, goal) {
		fmt.Fprintf(os.Stderr, "dudes: gate: acceptance goal %q is not documented in the makefile (target with `## description`)\n", goal)
		return 1
	}
	fmt.Printf("gate: running acceptance tests (make %s)…\n", goal)
	res, err := testrun.RunRecorded(cfg.Workspace, []string{"make", goal}, testrun.E2ENotesRef)
	if err != nil && res == nil {
		fmt.Fprintf(os.Stderr, "dudes: acceptance gate: %v\n", err)
		return 1
	}
	warnUnrecorded(err)
	if st != nil {
		events.Log(cfg.Workspace, st.RunID, "acceptance", st.Phase, fmt.Sprintf("goal=%s ok=%v", goal, res.OK))
	}
	if res.OK {
		fmt.Printf("acceptance: PASS (make %s, %s)\n", goal, res.Duration)
		return 0
	}
	fmt.Print(res.Output)
	fmt.Fprintf(os.Stderr, "dudes: acceptance gate FAILED (make %s) — phase not advanced\n", goal)
	return 1
}

// checkTestIDs: every acceptance test names the requirement it proves (F77).
func checkTestIDs(cfg config.Config, st *runstate.State, at itempack.AcceptanceCfg) int {
	if at.ID == "" || len(at.Paths) == 0 {
		return 0
	}
	res, err := testids.Run(cfg.Workspace, at.ID, at.Paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: test-ID check skipped: %v\n", err)
		return 0
	}
	if st != nil {
		events.Log(cfg.Workspace, st.RunID, "test_ids", st.Phase, fmt.Sprintf("missing=%d gate=%v", res.Missing, at.Gate))
	}
	if res.OK {
		fmt.Println("acceptance tests: every one names a requirement ID")
		return 0
	}
	fmt.Print(res.Output)
	if at.Gate {
		return 1
	}
	fmt.Println("(report only — set acceptance_tests.gate in the profile to make this a gate)")
	return 0
}

// logPromptSize records how large a worker prompt was — the measure for F76.
func logPromptSize(cfg config.Config, st *runstate.State, ph phase.ID, prompt string) {
	events.Log(cfg.Workspace, st.RunID, "prompt", ph, fmt.Sprintf("chars=%d", len(prompt)))
}

// scopeCodeBudget caps the existing code handed to the worker with its task, in bytes.
const scopeCodeBudget = 48 << 10

// scopeCode is the tracked code inside the task's writable scope, smallest first, up to
// scopeCodeBudget: the worker sees what it extends instead of writing it anew from the spec (L3: the
// worker asked "provide the full content of semver/semver.go"; an instance trial: config.go rewritten blind).
// No scope (the whole repository) → nothing extra.
func scopeCode(ws string, pats, have []string) (files []string, skipped int) {
	if len(pats) == 0 {
		return nil, 0
	}
	tracked, err := gitx.Tracked(ws)
	if err != nil {
		return nil, 0
	}
	type sized struct {
		path string
		size int
	}
	var in []sized
	for _, f := range tracked {
		if !scope.Inside(f, pats) || slices.Contains(have, f) || strings.HasPrefix(f, ".dudes/") {
			continue
		}
		if info, err := os.Stat(filepath.Join(ws, filepath.FromSlash(f))); err == nil {
			in = append(in, sized{f, int(info.Size())})
		}
	}
	slices.SortStableFunc(in, func(a, b sized) int { return a.size - b.size }) // the most files that fit
	used := 0
	for _, f := range in {
		if used+f.size > scopeCodeBudget {
			skipped++
			continue
		}
		used += f.size
		files = append(files, f.path)
	}
	return files, skipped
}

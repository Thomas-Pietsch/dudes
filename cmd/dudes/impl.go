package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/converge"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/extract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// paused is returned by phase work that parked the run in waiting_decision.
const paused = -1

// gateLogMax caps the failed gate output handed back to the worker: compilers and test runners
// print the first errors first.
const gateLogMax = 6000

// logPath is a file:line in tool output — relative, absolute when the toolchain names it so
// (`go test` reports a test package's setup failure with the full path, on Windows with the drive),
// or a bare file name (a failing Go test, relative to its package — inScope finds it).
var logPath = regexp.MustCompile(`((?:[A-Za-z]:)?[/\\]?[\w.~\-]+(?:[/\\][\w.~\-]+)*\.\w+):\d+`)

// gateFailure is the failed run of this task's gate (cmd) recorded on HEAD — the worker's last
// change — with the files it names: inside the task's scope (to edit) and outside it (to read —
// another task's file holding the other declaration); nothing when HEAD's gate passed, never ran,
// or was another command (a full-suite run before the item was sliced).
func gateFailure(ws string, cmd, pats []string) (log string, files, refs []string) {
	log, commit, ok, err := testrun.LastRun(ws)
	head, _ := gitx.HeadFull(ws)
	pass, why := revertedPass(ws, head) // HEAD undid a pass that broke green tests: its failure goes back
	if pass != "" {
		head = pass
	}
	if err != nil || !ok || commit != head || !strings.Contains(log, fmt.Sprintf("\n# args: %v\n", cmd)) {
		return "", nil, nil
	}
	defer func() {
		if pass != "" && log != "" {
			log = "Your last pass was reverted (" + why + "): you start again from the last green state. Do not repeat what broke them. Its gate output:\n\n" + log
		}
	}()
	if !strings.Contains(log, "\n# ok: false\n") {
		// a green run the gate still refused: its acceptance rerun skipped a criterion — a plain
		// `go test` log says nothing of it (in a replay the worker skipped three times, blind)
		if log = skippedLog(ws, head); log == "" {
			return "", nil, nil
		}
	}
	seen := map[string]bool{}
	abs, _ := filepath.Abs(ws)
	for _, m := range logPath.FindAllStringSubmatch(log, -1) {
		p := m[1]
		if filepath.IsAbs(p) {
			r, err := filepath.Rel(abs, p)
			if err != nil || strings.HasPrefix(r, "..") {
				continue
			}
			p = r
		}
		p = strings.ReplaceAll(p, `\`, "/") // a Windows toolchain prints backslashes
		for _, rel := range inScope(ws, p, pats) {
			if !seen[rel] {
				seen[rel] = true
				files = append(files, rel)
			}
		}
		if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(p))); err == nil && strings.Contains(p, "/") &&
			!scope.Inside(p, pats) && !seen[p] {
			seen[p] = true
			refs = append(refs, p)
		}
	}
	if len(log) > gateLogMax {
		log = log[:gateLogMax] + "\n…"
	}
	return log, files, refs
}

// skipRule is why a skipped acceptance test fails the gate: the gate's rule, a fact. What to write
// instead is the worker's judgement, not the engine's advice (CORE §2.4).
const skipRule = "The gate counts a skipped acceptance test as not green."

// skippedLog is the acceptance rerun recorded on head, when it skipped a test: each skipped test's
// run (its reason and file:line) and the rule; "" when nothing was skipped.
func skippedLog(ws, head string) string {
	commit, note, ok, err := gitx.LatestNote(ws, testrun.AcceptanceNotesRef, 200)
	if err != nil || !ok || commit != head {
		return ""
	}
	lines := strings.Split(note, "\n")
	var b strings.Builder
	for i, l := range lines {
		name, found := strings.CutPrefix(strings.TrimLeft(l, " \t"), "--- SKIP: ")
		if !found {
			continue
		}
		name, _, _ = strings.Cut(name, " ")
		start := i
		for j := i - 1; j >= 0; j-- { // the test's own run: from its === RUN line
			if strings.TrimLeft(lines[j], " \t") == "=== RUN   "+name {
				start = j
				break
			}
		}
		b.WriteString(strings.Join(lines[start:i+1], "\n") + "\n")
	}
	if b.Len() == 0 {
		return ""
	}
	return "acceptance test(s) skipped — the gate refused this run:\n" + b.String() + skipRule + "\n"
}

// inScope is the existing in-scope file a log names by path, or — a bare file name — the in-scope
// files of that name the task's patterns reach.
func inScope(ws, rel string, pats []string) []string {
	if strings.Contains(rel, "/") {
		if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(rel))); err == nil && scope.Inside(rel, pats) {
			return []string{rel}
		}
		return nil
	}
	var out []string
	for _, p := range pats {
		matches, _ := filepath.Glob(filepath.Join(ws, filepath.FromSlash(p)))
		for _, m := range matches {
			if r, err := filepath.Rel(ws, m); err == nil && filepath.Base(m) == rel && scope.Inside(filepath.ToSlash(r), pats) {
				out = append(out, filepath.ToSlash(r))
			}
		}
	}
	return out
}

// implementConverged runs pass 1 (the work) and, with convergence on, fresh
// re-analysis passes until one changes nothing (F72). Every pass that changes
// files is its own commit — git records how the task converged. subject names
// the step's commit-th commit, made in pass pass (passSubjects).
func implementConverged(cfg config.Config, st *runstate.State, ws string, pats []string, invoke func(string) (string, error), message string, subject func(commit, pass int) string, trailers string) int {
	format := formatter(cfg, ws)
	made := 0
	changed, esc, err := implPass(ws, pats, invoke, format, message, subject(made+1, 1), trailers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if changed {
		made++
	}
	if esc != nil {
		return escalate(cfg, st, *esc)
	}
	if !cfg.Convergence.Enabled {
		if !changed {
			fmt.Println("code worker produced no file changes — advance only")
		}
		return 0
	}
	max := cfg.Convergence.MaxPasses
	if max == 0 {
		max = converge.DefaultMaxPasses
	}
	var stop *extract.Escalation
	n, err := converge.Run(max, func(n int) (bool, error) {
		ch, e, err := implPass(ws, pats, invoke, format, message+converge.Instruction(n), subject(made+1, n), trailers)
		if err != nil {
			return false, err
		}
		if ch {
			made++
		}
		if e != nil {
			stop = e
			return false, converge.ErrStop
		}
		return ch, nil
	})
	events.Log(cfg.Workspace, st.RunID, "converge", phase.Implementation, fmt.Sprintf("passes=%d err=%v", n, err))
	switch {
	case stop != nil:
		return escalate(cfg, st, *stop)
	case errors.Is(err, converge.ErrNotConverged):
		fmt.Fprintf(os.Stderr, "dudes: not converged — pass %d still changed files. Phase not advanced; step again or raise convergence.max_passes\n", n)
		return 1
	case err != nil:
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	fmt.Printf("converged after %d passes — pass %d changed nothing (F72)\n", n, n)
	return 0
}

// implPass runs one fresh worker invoke and commits its changes (F42). It
// reports whether files changed and any BLOCKED:/CONTRACT: report (F74). Changes outside the
// task's writable patterns are discarded and named (F78); the pre-commit hook stays the backstop.
func implPass(ws string, pats []string, invoke func(string) (string, error), format func([]string), message, subject, trailers string) (bool, *extract.Escalation, error) {
	headBefore, _ := gitx.Head(ws)
	out, err := invoke(message)
	if err != nil {
		return false, nil, err
	}
	headAfter, _ := gitx.Head(ws)
	if headBefore != "" && headAfter != "" && headBefore != headAfter {
		fmt.Printf("warning: worker moved HEAD %s → %s (F42: release commits stay with orchestrator)\n", headBefore, headAfter)
	}
	var esc *extract.Escalation
	if e, ok := extract.FindEscalation(out); ok {
		esc = &e
		trailers = joinTrailers(trailers, "Escalation: "+e.Text)
	}
	paths, err := gitx.ChangedPaths(ws)
	if err != nil {
		return false, esc, fmt.Errorf("git: %w", err)
	}
	var inside, outside []string
	for _, p := range paths {
		if scope.Inside(p, pats) {
			inside = append(inside, p)
		} else {
			outside = append(outside, p)
		}
	}
	if len(outside) > 0 {
		if err := gitx.Discard(ws, outside); err != nil {
			return false, esc, fmt.Errorf("git: %w", err)
		}
		fmt.Printf("out of scope, discarded (F78): %s\n", strings.Join(outside, ", "))
	}
	paths = inside
	if len(paths) == 0 {
		return false, esc, nil
	}
	if format != nil {
		format(paths)
	}
	if err := gitx.AddCommit(ws, paths, withTrailers(subject, trailers)); err != nil {
		return false, esc, fmt.Errorf("git: %w", err)
	}
	fmt.Printf("committed %d path(s)\n", len(paths))
	return true, esc, nil
}

// escalationError carries a worker's stop report out of a text phase.
type escalationError struct{ e extract.Escalation }

func (x *escalationError) Error() string { return "worker escalated: " + x.e.Text }

// escalate parks the run on a worker's BLOCKED:/CONTRACT: report (F74).
// The worker does not invent; the human answers with dudes decide.
func escalate(cfg config.Config, st *runstate.State, e extract.Escalation) int {
	if st.TaskID == "" {
		st.TaskID = "escalation"
	}
	st.Status = runstate.StatusWaitingDecision
	st.Escalation = e.Text
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "escalation", st.Phase, e.Text)
	fmt.Printf("worker escalated (%s): %s\nanswer with: dudes decide %s <decision>\n", e.Kind, e.Text, st.TaskID)
	return paused
}

// escalationAnswer hands the human's decision back to the re-invoked worker.
func escalationAnswer(st *runstate.State) string {
	if st.Escalation == "" {
		return ""
	}
	return "\n\n## Human decision on your escalation\n\n" + st.Escalation +
		"\n\nDecision: " + nonzero(st.Decision, "(no note — proceed with the documented defaults)") + "\n"
}

// clearEscalation closes an answered escalation once the work succeeded.
func clearEscalation(st *runstate.State) {
	if st.Escalation == "" {
		return
	}
	st.Escalation = ""
	if st.TaskID == "escalation" {
		st.TaskID = ""
	}
}

func withTrailers(subject, trailers string) string {
	if trailers == "" {
		return subject
	}
	return subject + "\n\n" + trailers
}

func joinTrailers(a, b string) string {
	if a == "" {
		return b
	}
	return a + "\n" + b
}

// formatter is the configured format command for a worker's changed files (config format), run
// through the gate's launcher in the workspace; nil when none is configured. A failing formatter
// does not stop the step — the gate judges the files either way.
func formatter(cfg config.Config, ws string) func([]string) {
	if len(cfg.Format.Command) == 0 {
		return nil
	}
	return func(paths []string) {
		var files []string
		for _, p := range paths {
			if _, err := os.Stat(filepath.Join(ws, filepath.FromSlash(p))); err != nil {
				continue // deleted
			}
			if len(cfg.Format.Ext) == 0 || slices.Contains(cfg.Format.Ext, filepath.Ext(p)) {
				files = append(files, p)
			}
		}
		if len(files) == 0 {
			return
		}
		args := cfg.Launched(append(append([]string{}, cfg.Format.Command...), files...))
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = ws
		if out, err := cmd.CombinedOutput(); err != nil {
			fmt.Printf("format: %v — left as the worker wrote it\n%s", err, out)
			return
		}
		fmt.Printf("format: %d file(s)\n", len(files))
	}
}

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// The item branch (F88) merges as it is: every engine commit on it names its task, an approved task
// lands as one commit, and the slice ends with its last task — the QR is the spec repository's, so no
// later phase writes into the dev repository's branch.

// taskLog is how many commits back the engine reads a branch for a task's commits.
const taskLog = 200

// passSubjects names the commits of this step's passes: subject(commit, pass) is the subject of the
// step's commit-th commit, made in convergence pass pass (F72). A task of an item-bound contract
// numbers its passes on from the commits it already made and gains the Task trailer; anything else
// keeps the phase subjects.
func passSubjects(cfg config.Config, st *runstate.State, ws, trailers string) (func(commit, pass int) string, string) {
	con, _, _ := contract.Load(cfg.Workspace)
	names, ok := handoff.NamesFor(con, st.TaskID)
	if !ok {
		return func(_, pass int) string {
			if pass <= 1 {
				return fmt.Sprintf("chore(phase): %s [🔹engine]", phase.Implementation)
			}
			return fmt.Sprintf("chore(phase): %s pass %d [🔹engine]", phase.Implementation, pass)
		}, trailers
	}
	log, _ := gitx.FirstParents(ws, taskLog)
	done := handoff.TaskCommits(log, st.TaskID).Passes
	return func(commit, _ int) string { return names.Pass(done + commit) }, names.Trailers(trailers, 0)
}

// landTask runs when the role approves taskID: on an item's own branch — a slice's contract names its
// tests — the task lands as one commit (land), and the approval that leaves no task of the slice open
// ends the run (StatusEnded).
func landTask(cfg config.Config, st *runstate.State, taskID string) {
	land(cfg, st, taskID)
	endSlice(cfg, st)
}

// land makes the approved task taskID one commit named for it: its passes — a failed one, the guard's
// revert of it, a redo after a review — are what a reviewer would squash before a merge, so the engine
// does, through git commit and its hooks. The new commit holds the tree of the task's last commit and
// that commit's notes: its gate run, its acceptance run, the run state. The passes stay the record: the
// task's old head is kept first under refs/dudes/passes/<slice>/<task>/<n> (n counts the task's
// landings), a local ref that neither git push nor a clone takes by default, so the passes and the
// notes on them stay in the repository. Any other branch is left as it is. A squash that cannot be
// made — a commit another branch or a remote holds, a staged change — leaves the commits as they are
// and says why; the approval stands either way.
func land(cfg config.Config, st *runstate.State, taskID string) {
	con, ok, _ := contract.Load(cfg.Workspace)
	if !ok || !con.NamesTests() {
		return
	}
	tk, ok := con.TaskByID(taskID)
	if !ok || !tk.NeedsAfterDecision() || !st.HasCompleted(taskID) {
		return
	}
	names, ok := handoff.NamesFor(con, taskID)
	if !ok {
		return
	}
	ws := cfg.Workspace
	log, err := gitx.FirstParents(ws, taskLog)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: %v — the commits of %s stay as they are\n", err, taskID)
		return
	}
	run := handoff.TaskCommits(log, taskID)
	if len(run.Commits) == 0 || run.Base == "" || (len(run.Commits) == 1 && run.Commits[0].Subject == names.Landed()) {
		return // no commit, or landed already
	}
	stay := func(why string) {
		fmt.Fprintf(os.Stderr, "dudes: warning: the %d commit(s) of %s stay as they are: %s\n", len(run.Commits), taskID, why)
	}
	own := gitx.Branch(ws)
	refs, err := gitx.Containing(ws, run.Commits[len(run.Commits)-1].SHA)
	if err != nil {
		stay(err.Error())
		return
	}
	var shared []string
	for _, r := range refs {
		if r != own {
			shared = append(shared, strings.TrimPrefix(r, "refs/"))
		}
	}
	if len(shared) > 0 {
		stay("they are in " + strings.Join(shared, ", ") + " as well, and a rewrite would fork that history")
		return
	}
	notes, err := gitx.NotesRefs(ws, "dudes")
	if err != nil {
		stay(err.Error())
		return
	}
	// the ledger records each event on the commit it happened on and is read whole (events.Ledger):
	// it stays on the kept passes, where a copy would count them twice
	notes = slices.DeleteFunc(notes, func(r string) bool { return r == events.LedgerRef })
	kept, err := keepPasses(ws, con.ID, taskID, run.Commits[0].SHA)
	if err != nil {
		stay(err.Error())
		return
	}
	msg := names.Landed() + "\n\n" + names.Trailers(itemTrailers(run.Commits), run.Passes)
	sha, err := gitx.Squash(ws, run.Base, msg, notes)
	if err != nil {
		if derr := gitx.DeleteRef(ws, kept); derr != nil {
			err = fmt.Errorf("%v (and %v)", err, derr)
		}
		stay(err.Error())
		return
	}
	events.Log(cfg.Workspace, st.RunID, "land", st.Phase, fmt.Sprintf("%s %s passes=%d commits=%d kept=%s", taskID, shortSHA(sha), run.Passes, len(run.Commits), kept))
	fmt.Printf("landed %s as one commit (%s, from %d): %s\n", taskID, shortSHA(sha), len(run.Commits), names.Landed())
	fmt.Printf("the passes stay at %s\n", kept)
}

// passesRefs is where the passes of landed tasks are kept: refs/dudes/passes/<slice>/<task>/<n>.
const passesRefs = "refs/dudes/passes/"

// keepPasses keeps the task's old head head under the task's next passes ref and returns that ref.
func keepPasses(ws, slice, task, head string) (string, error) {
	dir := passesRefs + refPart(slice) + "/" + refPart(task) + "/"
	have, err := gitx.Refs(ws, dir)
	if err != nil {
		return "", err
	}
	n := 1
	for _, r := range have {
		if k, err := strconv.Atoi(strings.TrimPrefix(r, dir)); err == nil && k >= n {
			n = k + 1
		}
	}
	ref := dir + strconv.Itoa(n)
	return ref, gitx.CreateRef(ws, ref, head)
}

// refPart is s as one part of a ref name: what git refuses there becomes "-".
func refPart(s string) string {
	var b strings.Builder
	for _, r := range strings.Trim(s, ".") {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := strings.ReplaceAll(b.String(), "..", "-")
	if out == "" || strings.HasSuffix(out, ".lock") {
		out += "-"
	}
	return out
}

// itemTrailers are the item's trailers of a task's newest commit that has them: the item and the hash
// of every source document (the item pack's, F71).
func itemTrailers(commits []gitx.Entry) string {
	for _, c := range commits {
		item, sources := c.Trailers["Item"], c.Trailers["Spec-Source"]
		if len(sources) == 0 {
			continue // the guard's revert names the item, not the sources
		}
		var lines []string
		for _, v := range item {
			lines = append(lines, "Item: "+v)
		}
		for _, v := range sources {
			lines = append(lines, "Spec-Source: "+v)
		}
		return strings.Join(lines, "\n")
	}
	if len(commits) > 0 {
		if item := commits[0].Trailer("Item"); item != "" {
			return "Item: " + item
		}
	}
	return ""
}

// sliceDone is what is left when the run is an item slice (its contract names its tests, F88) and
// every task of it is done and decided: the role's merge of the branch. ok is false otherwise.
func sliceDone(cfg config.Config, st *runstate.State) (next string, ok bool) {
	con, found, _ := contract.Load(cfg.Workspace)
	if !found || st == nil || !con.NamesTests() || len(con.Tasks) == 0 {
		return "", false
	}
	for _, t := range con.Tasks {
		if !st.HasCompleted(t.ID) || (t.NeedsAfterDecision() && !st.HasApproved(t.ID)) {
			return "", false
		}
	}
	branch := strings.TrimPrefix(gitx.Branch(cfg.Workspace), "refs/heads/")
	if branch == "" {
		branch = "<the slice's branch>"
	}
	return fmt.Sprintf("slice %s done: every task is decided, and %s merges as it is — next, from the repo's root on the branch it came from: make land SLICE=%s ROLE=<your role>, or dudes land %s -by <your role>",
		handoff.SliceName(con), branch, con.ID, con.ID), true
}

// endSlice ends the run of an item slice whose every task is done and decided: its state says so
// (StatusEnded), nothing steps it on, and once the branch is merged its newest run note reads as no
// run under way in every checkout that has it. ok is false, and st unchanged, for anything else.
func endSlice(cfg config.Config, st *runstate.State) (next string, ok bool) {
	if next, ok = sliceDone(cfg, st); ok {
		st.Status = runstate.StatusEnded
	}
	return next, ok
}

// endSliceStep ends a slice whose last task a step finished without a decision to wait for: the state
// is saved as ended and the merge is named instead of the next phase. ended is false for anything else.
func endSliceStep(cfg config.Config, st *runstate.State) (code int, ended bool) {
	next, ok := endSlice(cfg, st)
	if !ok {
		return 0, false
	}
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1, true
	}
	fmt.Println(next)
	return 0, true
}

const landUsage = "usage: dudes land <slice> -by <role>, from the dev repository's root"

// cmdLand lands an ended slice (F88), run from the dev repository's root: git merge --no-ff of its
// branch into the branch checked out there, the role's tag and the item on the merge commit, through
// git's hooks. A merge git stops is taken back, and nothing else happens. Merged, the slice's spec
// pins, its worktree and its branch go; its run notes and its kept passes stay. Exit 1: refused, or
// merged with something left to remove; 2: a usage error.
func cmdLand(args []string) int {
	id, extra, role, err := answerArgs(args)
	if err != nil || id == "" || extra != "" || role == "" {
		fmt.Fprintf(os.Stderr, "dudes: %s%s\n", errPrefix(err), landUsage)
		return 2
	}
	if err := handoff.CheckID(id); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 2
	}
	fail := func(format string, a ...any) int {
		fmt.Fprintf(os.Stderr, "dudes: "+format+"\n", a...)
		return 1
	}
	dir, err := os.Getwd()
	if err != nil {
		return fail("%v", err)
	}
	root, err := gitx.Toplevel(dir)
	if err != nil {
		return fail("%v", err)
	}
	if !handoff.SameDir(root, dir) {
		return fail("run it from the dev repository's root (%s), not from %s", root, dir)
	}
	wt := filepath.Join(root, ".worktrees", id)
	cfg, _, err := config.Load(filepath.Join(wt, ".dudes", "dudes.json"))
	if err != nil {
		return fail("no slice %s here: %v", id, err)
	}
	st, ok, err := runstate.Load(wt)
	if err != nil || !ok {
		return fail("slice %s has no run (%v)", id, err)
	}
	if st.Status != runstate.StatusEnded {
		return fail("slice %s is not done (run status %s): review and decide its tasks first (dudes status)", id, st.Status)
	}
	con, ok, err := contract.Load(wt)
	if err != nil || !ok {
		return fail("slice %s has no contract (%v)", id, err)
	}
	branch := strings.TrimPrefix(gitx.Branch(wt), "refs/heads/")
	into := strings.TrimPrefix(gitx.Branch(root), "refs/heads/")
	if branch == "" || into == "" || branch == into {
		return fail("slice %s: its branch %q cannot be merged into %q — check out the branch it came from", id, branch, into)
	}
	if err := gitx.MergeNoFF(root, branch, handoff.MergeMessage(con, branch, role)); err != nil {
		return fail("%v\nnothing landed: the merge is taken back", err)
	}
	head, _ := gitx.Head(root)
	fmt.Printf("landed %s: %s merged into %s (%s), by %s\n", handoff.SliceName(con), branch, into, head, role)

	// the merge is in: what only the slice's run needed goes, its pins first (the worktree names them)
	var left []string
	if set, ok, err := specset.Load(wt, cfg.Product); err == nil && ok {
		for name := range set.Roots {
			pin, err := set.RootAbs(wt, name)
			if err != nil || filepath.Base(pin) != "spec-"+id || filepath.Base(filepath.Dir(pin)) != ".worktrees" {
				continue // not a pin this slice made
			}
			if err := gitx.RemoveWorktree(filepath.Dir(filepath.Dir(pin)), pin); err != nil {
				left = append(left, err.Error())
			}
		}
	}
	if err := gitx.RemoveWorktree(root, wt); err != nil {
		left = append(left, err.Error())
	} else if err := gitx.DeleteMergedBranch(root, branch); err != nil {
		left = append(left, err.Error())
	}
	if len(left) > 0 {
		return fail("landed, and left: %s", strings.Join(left, "; "))
	}
	fmt.Printf("removed .worktrees/%s, its spec pins and %s; its run notes and kept passes stay\n", id, branch)
	return 0
}

// endedRun is what a command says about a run that ended (StatusEnded): in the slice's worktree the
// merge that is left, elsewhere — the branch it came from, after the merge — that nothing runs.
func endedRun(cfg config.Config, st *runstate.State) string {
	if next, ok := sliceDone(cfg, st); ok {
		return next
	}
	return fmt.Sprintf("run %s (%s) ended: every task of its slice was decided — no run is under way", st.RunID, nonzero(st.Prompt, "-"))
}

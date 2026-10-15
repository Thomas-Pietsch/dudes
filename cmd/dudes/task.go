package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/recipe"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
)

func pauseDecision(cfg config.Config, st *runstate.State, tk contract.Task) int {
	st.Status = runstate.StatusWaitingDecision
	st.TaskID = tk.ID
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "waiting_decision", st.Phase, tk.ID)
	fmt.Printf("waiting for decision on task %s — dudes decide %s\n", tk.ID, tk.ID)
	return 0
}

// nextOpenTask is the phase's first task that is not done: one still to run, or one that ran and
// still waits for its review (approval: after, not approved) — a review comes before the next task.
func nextOpenTask(con *contract.Contract, st *runstate.State) (contract.Task, bool) {
	if con == nil || st == nil {
		return contract.Task{}, false
	}
	for _, t := range con.TasksFor(st.Phase) {
		if !st.HasCompleted(t.ID) || (t.NeedsAfterDecision() && !st.HasApproved(t.ID)) {
			return t, true
		}
	}
	return contract.Task{}, false
}

func firstTaskInPhase(con *contract.Contract, st *runstate.State) bool {
	if con == nil || st == nil {
		return true
	}
	tasks := con.TasksFor(st.Phase)
	if len(tasks) == 0 {
		return true
	}
	for _, t := range tasks {
		if st.HasCompleted(t.ID) {
			return false
		}
	}
	return true
}

func stepContractTask(cfg config.Config, st *runstate.State, con *contract.Contract, rec *recipe.Recipe, hasRecipe, stub, skipTest bool, discipline string) int {
	tk, ok := nextOpenTask(con, st)
	if !ok {
		if code, ended := endSliceStep(cfg, st); ended {
			return code
		}
		if err := gate.CheckLeave(cfg.Workspace, cfg.Product, st.Phase); err != nil {
			events.Log(cfg.Workspace, st.RunID, "gate_fail", st.Phase, err.Error())
			fmt.Fprintf(os.Stderr, "dudes: gate: %v\n", err)
			return 1
		}
		events.Log(cfg.Workspace, st.RunID, "phase_leave", st.Phase, "")
		st.TaskID = ""
		return advanceLocked(cfg, st)
	}

	st.TaskID = tk.ID
	if st.HasCompleted(tk.ID) { // it ran and its review is still due: it pauses again
		return pauseDecision(cfg, st, tk)
	}
	if tk.NeedsDecision() && !st.HasApproved(tk.ID) {
		return pauseDecision(cfg, st, tk)
	}

	taskOverride := strings.TrimSpace(tk.Title)
	if hasRecipe {
		if extra := rec.TaskOverride(); extra != "" {
			taskOverride = extra + "\n" + taskOverride
		}
	}
	if len(tk.Skills) > 0 || len(tk.Rules) > 0 {
		extra, err := loadDiscipline(cfg, &recipe.Recipe{Skills: tk.Skills, Rules: tk.Rules})
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		if extra != "" {
			if discipline != "" {
				discipline += "\n\n"
			}
			discipline += extra
		}
	}
	tools := tk.Tools
	if len(tools) == 0 && hasRecipe {
		tools = rec.Tools
	}
	if tk.Worker == "stub" {
		stub = true
	}

	packCtx := loadPackContext(cfg)
	fmt.Printf("task: %s (%s)\n", tk.ID, tk.Title)
	events.Log(cfg.Workspace, st.RunID, "task_enter", st.Phase, tk.ID)
	// Where the task's review diff starts (F88), saved now: a step whose gate fails saves no state, and
	// the next one would start the diff after this step's commits.
	if head, _ := gitx.HeadFull(cfg.Workspace); st.StartTask(tk.ID, head) {
		if err := runstate.Save(cfg.Workspace, st); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
	}
	if code := executePhaseWork(cfg, st, stub, skipTest, packCtx, taskOverride, discipline, tools, tk.Worker); code == paused {
		return 0
	} else if code != 0 {
		return code
	}
	st.MarkCompleted(tk.ID)
	if tk.NeedsAfterDecision() && !st.HasApproved(tk.ID) {
		return pauseDecision(cfg, st, tk)
	}
	st.Status = runstate.StatusRunning
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "task_leave", st.Phase, tk.ID)
	if _, more := nextOpenTask(con, st); !more {
		if code, ended := endSliceStep(cfg, st); ended {
			return code
		}
		if err := gate.CheckLeave(cfg.Workspace, cfg.Product, st.Phase); err != nil {
			events.Log(cfg.Workspace, st.RunID, "gate_fail", st.Phase, err.Error())
			fmt.Fprintf(os.Stderr, "dudes: gate: %v\n", err)
			return 1
		}
		events.Log(cfg.Workspace, st.RunID, "phase_leave", st.Phase, "")
		st.TaskID = ""
		return advanceLocked(cfg, st)
	}
	fmt.Printf("task %s done — phase %s has more tasks\n", tk.ID, st.Phase)
	return 0
}

func executePhaseWork(cfg config.Config, st *runstate.State, stub, skipTest bool, packCtx, taskOverride, discipline string, tools []string, worker string) int {
	cur := st.Phase
	commitPaths := []string{}
	if cur == phase.Implementation {
		if worker == "ollama" {
			fmt.Fprintln(os.Stderr, "dudes: worker ollama is not valid for implementation (use aider, pi, or stub)")
			return 1
		}
		pre, _ := gitx.HeadFull(cfg.Workspace) // where this pass starts: the guard returns here (F86)
		if code := runImplementation(cfg, st, stub, packCtx, taskOverride, discipline, tools, worker); code == paused {
			return paused
		} else if code != 0 {
			events.Log(cfg.Workspace, st.RunID, "worker_fail", cur, "implementation")
			return code
		}
		clearEscalation(st)
		if !skipTest {
			if code := runTestsGate(cfg, st); code != 0 {
				events.Log(cfg.Workspace, st.RunID, "test_fail", cur, "post-impl")
				guardRegression(cfg, st, pre)
				return code
			}
			events.Log(cfg.Workspace, st.RunID, "test_pass", cur, "post-impl")
		}
		return 0
	}
	rel, ok := artifact.Path(cfg.Product, cur)
	if !ok {
		fmt.Printf("phase %s has no primary artifact — task records only\n", cur)
		return 0
	}
	body, src, err := phaseBody(cfg, st, cur, stub, packCtx, taskOverride, discipline)
	var esc *escalationError
	if errors.As(err, &esc) {
		return escalate(cfg, st, esc.e)
	}
	if err != nil {
		events.Log(cfg.Workspace, st.RunID, "worker_fail", cur, err.Error())
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if err := artifact.Write(cfg.Workspace, rel, []byte(body)); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	commitPaths = append(commitPaths, rel)
	if cur == phase.TestAuthoring {
		item, acs, _ := itemAcceptance(cfg)
		for _, ac := range acs {
			f := spec.ItemTestFile(ac.ID)
			if err := artifact.Write(cfg.Workspace, f, []byte(spec.ItemGoTest(item, ac))); err != nil {
				fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
				return 1
			}
			commitPaths = append(commitPaths, f)
		}
	}
	if cur == phase.Spec {
		taRel := artifact.TAPath(cfg.Product)
		if err := artifact.Write(cfg.Workspace, taRel, []byte(artifact.TAContent(cfg.Product, st.Prompt))); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: ta: %v\n", err)
			return 1
		}
		commitPaths = append(commitPaths, taRel)
		fmt.Printf("wrote %s (F13 companion TA)\n", taRel)
	}
	msg := fmt.Sprintf("chore(phase): %s [🔹engine]", cur)
	if cur == phase.TestAuthoring || cur == phase.QA || cur == phase.Documentation {
		if item, err := itemSpec(cfg); err == nil && item != nil {
			msg = withTrailers(msg, item.trailers)
		}
	}
	clearEscalation(st)
	if err := gitx.AddCommit(cfg.Workspace, commitPaths, msg); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: git: %v\n", err)
		return 1
	}
	fmt.Printf("wrote %s via %s and committed\n", rel, src)
	if cur == phase.Spec {
		if fr, err := freeze.WriteAll(cfg.Workspace, cfg.Product, st.RunID); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: freeze: %v\n", err)
			return 1
		} else {
			events.Log(cfg.Workspace, st.RunID, "spec_freeze", cur, fr.SHA256)
			fmt.Printf("froze spec %s (sha256 %s…, %d inputs)\n", fr.PRD, trimSHA(fr.SHA256), len(fr.Inputs))
		}
	}
	return 0
}

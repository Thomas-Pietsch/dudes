package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/evidence"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

func cmdSpecset(configPath string, args []string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	init := false
	for _, a := range args {
		if a == "init" {
			init = true
		}
	}
	if init {
		s := specset.EngineSet(cfg.Product)
		if err := specset.Save(cfg.Workspace, s); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		fmt.Printf("wrote %s (%d entries, engine names)\n", specset.Path(), len(s.Entries))
		fmt.Println("instance filenames: edit the file — do not rename source docs")
		return 0
	}
	s, ok, err := specset.Load(cfg.Workspace, cfg.Product)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Printf("specset: (default engine set — dudes specset init to write %s)\n", specset.Path())
	} else {
		fmt.Printf("specset: %s product=%s\n", specset.Path(), s.Product)
	}
	for k, v := range s.Roots {
		fmt.Printf("  root %s = %s\n", k, v)
	}
	for _, e := range s.Entries {
		fmt.Printf("  %-12s %-8s %s:%s\n", e.ID, e.Kind, e.Root, e.Path)
	}
	return 0
}

func cmdContract(configPath string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	c, ok, err := contract.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Printf("contract: (none — %s)\n", contract.Path())
		return 0
	}
	if err := c.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	fmt.Printf("contract: %s id=%s product=%s variant=%s\n", contract.Path(), c.ID, c.Product, nonzero(c.Variant, "standard"))
	fmt.Printf("  sources: %d  acceptance: %d  tasks: %d\n", len(c.Sources), len(c.Acceptance), len(c.Tasks))
	for _, t := range c.Tasks {
		fmt.Printf("  task %-18s phase=%-16s worker=%-6s approval=%s\n", t.ID, t.Phase, nonzero(t.Worker, "default"), nonzero(t.Approval, "none"))
	}
	return 0
}

func cmdEvidenceShow(configPath string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	prd, _ := artifact.Path(cfg.Product, phase.Spec)
	tests, _ := artifact.Path(cfg.Product, phase.TestAuthoring)
	f, err := evidence.Derive(cfg.Workspace, cfg.Product, prd, tests)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if len(f.Rows) == 0 {
		fmt.Println("evidence: (none — rows appear once tests are authored from the PRD)")
		return 0
	}
	fmt.Printf("evidence: derived from %s + %s, product=%s\n", prd, evidence.Source(), f.Product)
	for _, r := range f.Rows {
		fmt.Printf("  %-8s %-8s req=%s test=%s commit=%s\n", r.Class, r.Status, r.Requirement, r.Test, r.Commit)
	}
	return 0
}

func cmdDecide(configPath string, args []string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes init")
		return 1
	}
	taskID, note, role, err := answerArgs(args)
	if err != nil || taskID == "" {
		fmt.Fprintf(os.Stderr, "dudes: %susage: dudes decide <task-id> [-by <role>] [note]\n", errPrefix(err))
		return 2
	}
	answersEscalation := st.Escalation != "" && taskID == st.TaskID // F74
	if con, cok, _ := contract.Load(cfg.Workspace); cok && con != nil && !answersEscalation {
		if _, ok := con.TaskByID(taskID); !ok {
			fmt.Fprintf(os.Stderr, "dudes: unknown task %q\n", taskID)
			return 1
		}
	}
	by, err := decider(cfg.Workspace, role)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if answersEscalation {
		// the answer goes to the task's next run; it is no approval: a task that pauses after its run
		// still pauses for its review then
		st.Answer(taskID, note, by)
	} else {
		st.Approve(taskID, note, by)
		landTask(cfg, st, taskID) // F88: on a slice's branch the task lands as one commit; the last ends the run
	}
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	logAnswer(cfg.Workspace, st, "decide", taskID+" "+note, by)
	what := "decided"
	if answersEscalation {
		what = "answered the escalation of"
	}
	if next, done := sliceDone(cfg, st); done {
		fmt.Printf("%s %s, by %s\n%s\n", what, taskID, by, next)
		return 0
	}
	fmt.Printf("%s %s, by %s — status %s (continue with dudes step / dudes run)\n", what, taskID, by, st.Status)
	return 0
}

// cmdRedo sends a task back with the human's feedback — the other answer at an
// `approval: after` pause, e.g. a red test that tests the wrong thing. The next step runs the task
// again; only its worker gets the feedback.
func cmdRedo(configPath string, args []string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	taskID, note, role, err := answerArgs(args)
	if err != nil || taskID == "" || note == "" {
		fmt.Fprintf(os.Stderr, "dudes: %susage: dudes redo <task-id> [-by <role>] <feedback>\n", errPrefix(err))
		return 2
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil || !ok {
		fmt.Fprintf(os.Stderr, "dudes: no run state (%v)\n", err)
		return 1
	}
	if con, cok, _ := contract.Load(cfg.Workspace); !cok || con == nil {
		fmt.Fprintln(os.Stderr, "dudes: redo needs a contract task")
		return 1
	} else if tk, ok := con.TaskByID(taskID); !ok || tk.Phase != st.Phase {
		fmt.Fprintf(os.Stderr, "dudes: %s is not a task of the current phase %s\n", taskID, st.Phase)
		return 1
	}
	by, err := decider(cfg.Workspace, role)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	st.Reopen(taskID, note, by) // a finished task opens again; an open one gets the note for its next run
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	logAnswer(cfg.Workspace, st, "redo", taskID+" "+note, by)
	fmt.Printf("sent %s back, by %s — the next step runs it again with your feedback\n", taskID, by)
	return 0
}

// answerArgs reads the words after dudes decide or dudes redo: the task, then the note, and anywhere
// among them -by <role> (or -by=<role>), the role of whoever answers (F88).
func answerArgs(words []string) (task, note, role string, err error) {
	var rest []string
	for i := 0; i < len(words); i++ {
		name, val, hasVal := strings.Cut(words[i], "=")
		if name != "-by" && name != "--by" {
			rest = append(rest, words[i])
			continue
		}
		if !hasVal && i+1 < len(words) {
			i++
			val = words[i]
		}
		if role = strings.TrimSpace(val); role == "" {
			return "", "", "", fmt.Errorf("-by needs a role, e.g. -by Developer")
		}
	}
	if len(rest) > 0 {
		task, note = rest[0], strings.TrimSpace(strings.Join(rest[1:], " "))
	}
	return task, note, role, nil
}

func errPrefix(err error) string {
	if err == nil {
		return ""
	}
	return err.Error() + " — "
}

// decider is who answers a pause from ws: the git identity a commit made now would record, and the
// role -by named. Without an identity nothing is recorded: git could not write the run state either.
func decider(ws, role string) (runstate.Decider, error) {
	id, err := gitx.Identity(ws)
	if err != nil {
		return runstate.Decider{}, err
	}
	return runstate.Decider{Identity: id, Role: role}, nil
}

// logAnswer appends a decide or redo event that names who answered (F88): to the log and the ledger.
func logAnswer(ws string, st *runstate.State, typ, detail string, by runstate.Decider) {
	if err := events.Append(ws, events.Event{RunID: st.RunID, Type: typ, Phase: st.Phase, Detail: detail, By: by.Identity, Role: by.Role}); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: the answer is not recorded in full: %v\n", err)
	}
}

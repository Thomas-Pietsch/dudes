package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// revertTrailer names, in the guard's revert commit, the last commit of the pass it undid.
const revertTrailer = "Dudes-Reverted:"

// guardRegression undoes the worker's pass — the commits after pre — when its gate failed and it
// broke what was green before: an acceptance test an earlier gate recorded passing that now fails
// or no longer builds (F86). A failed pass that breaks nothing stays: the next step builds on it.
// In a replay of real product slices, in 3 of 6 runs a failed pass broke the build and earlier
// slices stayed red to the end.
func guardRegression(cfg config.Config, st *runstate.State, pre string) {
	if pre == "" || expectRed(cfg, st) {
		return
	}
	head, _ := gitx.HeadFull(cfg.Workspace)
	cmd, _ := gateCommand(cfg, st)
	if head == "" || head == pre || len(cmd) < 2 || cmd[0] != "go" || cmd[1] != "test" {
		return
	}
	acc, err := testrun.Acceptance(cfg.Workspace) // the pass's commits carry none: what was green before it
	if err != nil {
		return
	}
	var green []string
	for name, o := range acc {
		if o.Result == testrun.Pass && !strings.Contains(name, "/") {
			green = append(green, name)
		}
	}
	if len(green) == 0 {
		return
	}
	sort.Strings(green)
	verbose := append([]string{"go", "test", "-v", "-run", "^(" + strings.Join(green, "|") + ")$"}, cmd[2:]...)
	res, _ := testrun.RunRecorded(cfg.Workspace, cfg.Launched(verbose), "") // a check, not evidence
	if res == nil {
		return
	}
	results := testrun.Results(res.Output)
	var broken []string
	for _, n := range green {
		if results[n] != testrun.Pass { // failed, skipped, or not built
			broken = append(broken, n)
		}
	}
	if len(broken) == 0 {
		return
	}
	msg := fmt.Sprintf("chore(phase): revert a pass that broke %s, green before [🔹engine]\n\n%s %s\n",
		strings.Join(broken, ", "), revertTrailer, head)
	if err := gitx.RevertSince(cfg.Workspace, pre, msg); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: guard: %v — the pass stays\n", err)
		return
	}
	events.Log(cfg.Workspace, st.RunID, "revert", st.Phase, strings.Join(broken, " "))
	fmt.Printf("guard: the pass broke %s, green before — reverted; the next step starts from the last green state, with this gate's output\n",
		strings.Join(broken, ", "))
}

// revertedPass is the pass HEAD undid (the guard's revert commit) and why: its last commit and the
// commit's subject; "" when HEAD is no such revert.
func revertedPass(ws, head string) (pass, why string) {
	msg, err := gitx.Message(ws, head)
	if err != nil {
		return "", ""
	}
	for _, line := range strings.Split(msg, "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), revertTrailer); ok {
			subject, _, _ := strings.Cut(msg, "\n")
			return strings.TrimSpace(rest), strings.TrimSpace(subject)
		}
	}
	return "", ""
}

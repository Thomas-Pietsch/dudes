// Package testrun executes framework-owned test commands (not LLM-decided).
package testrun

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/awake"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/goals"
	"github.com/Thomas-Pietsch/dudes/internal/redact"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Result of a framework-owned test run.
type Result struct {
	Args     []string
	Dir      string
	ExitCode int
	Output   string
	OK       bool
	Duration time.Duration
}

// DefaultArgs is used when config omits test_command and the repo has no `test` goal. Verbose,
// because evidence (F67) turns green only for tests whose names appear in the log.
var DefaultArgs = []string{"go", "test", "-v", "./..."}

// DefaultFor is the test command for dir when config omits test_command: the repo's own
// `make test` goal if its makefile documents one (goal interface, F75), else DefaultArgs.
func DefaultFor(dir string) []string {
	if goals.Has(dir, "test") {
		return []string{"make", "test"}
	}
	return append([]string{}, DefaultArgs...)
}

// Run executes args in dir and records the run as a git note on the tested commit (CORE §10:
// facts about a commit, added without changing it). Evidence (F67) and the gates read it back.
func Run(dir string, args []string) (*Result, error) {
	return RunRecorded(dir, args, NotesRef)
}

// E2ENotesRef holds acceptance (E2E) runs, apart from the unit runs evidence is derived from.
const E2ENotesRef = "dudes/e2e"

// RunRecorded is Run with the notes ref the run is recorded under.
func RunRecorded(dir string, args []string, ref string) (*Result, error) {
	dirAbs, err := workspace.Resolve(dir, ".")
	if err != nil {
		return nil, fmt.Errorf("test workspace: %w", err)
	}
	if len(args) == 0 {
		args = append([]string{}, DefaultArgs...)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dirAbs
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	start := time.Now()
	// Standby would freeze a gate that runs in WSL or a container (the VM sleeps with the host).
	release := awake.Hold()
	err = cmd.Run()
	release()
	res := &Result{
		Args:     append([]string{}, args...),
		Dir:      dirAbs,
		Output:   redact.String(buf.String()),
		Duration: time.Since(start),
		OK:       err == nil,
	}
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
		} else {
			res.ExitCode = -1
			return res, fmt.Errorf("run tests: %w", err)
		}
	}
	if ref == "" { // a check, not evidence: nothing recorded
		return res, nil
	}
	if noteErr := record(dirAbs, ref, res); noteErr != nil {
		return res, noteErr
	}
	return res, nil
}

// NotesRef holds test runs: refs/notes/dudes/test, one note per tested commit (the last run wins).
const NotesRef = "dudes/test"

// LastRun returns the newest recorded test run in HEAD's history and the commit it tested.
func LastRun(dir string) (log, commit string, ok bool, err error) {
	commit, log, ok, err = gitx.LatestNote(dir, NotesRef, 200)
	return log, commit, ok, err
}

// AcceptanceNotesRef holds the verbose rerun of a bound item's acceptance tests after a green gate.
// A plain `go test` names no passing test; this run does, so item status and evidence read it.
const AcceptanceNotesRef = "dudes/acceptance"

// Results of one test in a verbose log.
const (
	Pass = "pass"
	Fail = "fail"
	Skip = "skip"
)

// Outcome is a test's result and the commit of the run that reported it.
type Outcome struct{ Result, Commit string }

// Acceptance merges the recorded acceptance runs in HEAD's history: per test, the newest run that
// reports it wins. Slice gates accumulate, and one item's gate leaves another item's tests as they
// were. Empty: no acceptance run recorded.
func Acceptance(dir string) (map[string]Outcome, error) {
	notes, err := gitx.Notes(dir, AcceptanceNotesRef, 200)
	if err != nil {
		return nil, err
	}
	out := map[string]Outcome{}
	for _, n := range notes { // newest first
		for name, res := range Results(n.Text) {
			if _, seen := out[name]; !seen {
				out[name] = Outcome{Result: res, Commit: n.Commit}
			}
		}
	}
	return out, nil
}

// Results parses a verbose `go test` log into pass / fail / skip per test, from its
// `--- PASS: <name> `, `--- FAIL: <name> ` and `--- SKIP: <name> ` lines. A test reported more than
// once (-count) keeps its worst result: fail over skip over pass.
func Results(log string) map[string]string {
	rank := map[string]int{Pass: 1, Skip: 2, Fail: 3}
	out := map[string]string{}
	for _, line := range strings.Split(log, "\n") {
		line = strings.TrimLeft(line, " \t")
		for prefix, res := range map[string]string{"--- PASS: ": Pass, "--- FAIL: ": Fail, "--- SKIP: ": Skip} {
			rest, found := strings.CutPrefix(line, prefix)
			if !found {
				continue
			}
			if name, _, sp := strings.Cut(rest, " "); sp && name != "" && rank[res] > rank[out[name]] {
				out[name] = res
			}
		}
	}
	return out
}

// ReadLastLog returns the newest recorded test run's log.
func ReadLastLog(dir string) (string, bool, error) {
	log, _, ok, err := LastRun(dir)
	return log, ok, err
}

// Record is a run read back from its note: the header record writes, then the output.
type Record struct {
	Args     string // the command as recorded: "[go test ./...]"
	Exit     int
	OK       bool
	Duration string
	Dirty    bool // the run tested uncommitted changes too
	Output   string
}

// ParseRecord reads a note record wrote. A note without its first line, "# dudes test", is all output.
func ParseRecord(note string) Record {
	first, rest, found := strings.Cut(note, "\n")
	if first != "# dudes test" || !found {
		return Record{Output: note}
	}
	var r Record
	for {
		line, after, found := strings.Cut(rest, "\n")
		if !strings.HasPrefix(line, "# ") {
			break
		}
		key, val, _ := strings.Cut(strings.TrimPrefix(line, "# "), ": ")
		switch key {
		case "args":
			r.Args = val
		case "exit":
			fmt.Sscanf(val, "%d", &r.Exit)
		case "ok":
			r.OK = val == "true"
		case "duration":
			r.Duration = val
		case "tree":
			r.Dirty = strings.HasPrefix(val, "dirty")
		}
		if !found {
			return r
		}
		rest = after
	}
	r.Output = strings.TrimPrefix(rest, "\n")
	return r
}

// record writes the run as a note on HEAD. No commit yet → nothing to attach it to.
func record(dir, ref string, res *Result) error {
	if head, _ := gitx.Head(dir); head == "" {
		return nil
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# dudes test\n# args: %v\n# exit: %d\n# ok: %v\n# duration: %s\n", res.Args, res.ExitCode, res.OK, res.Duration)
	if changed, err := gitx.ChangedPaths(dir); err == nil && len(changed) > 0 {
		fmt.Fprintf(&b, "# tree: dirty (%d uncommitted path(s) — the run tested more than this commit)\n", len(changed))
	}
	b.WriteString("\n")
	b.WriteString(res.Output)
	return gitx.AddNote(dir, ref, b.String())
}

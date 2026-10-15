// Package bench prepares an isolated stub workspace for `dudes bench` (F53).
package bench

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

// Report is one hermetic stub-run measurement.
type Report struct {
	Product     string      `json:"product"`
	Stub        bool        `json:"stub"`
	SkipTest    bool        `json:"skip_test"`
	StartedAt   time.Time   `json:"started_at"`
	FinishedAt  time.Time   `json:"finished_at"`
	DurationMs  int64       `json:"duration_ms"`
	FinalPhase  phase.ID    `json:"final_phase"`
	GateFails   int         `json:"gate_fails"`
	PhaseLeaves int         `json:"phase_leaves"`
	CheckOK     bool        `json:"check_ok"`
	Check       gate.Report `json:"check"`
	OK          bool        `json:"ok"`
	Workspace   string      `json:"workspace"`
}

// Prepare creates a git workspace with dudes.json, go.mod, and git identity.
func Prepare(dir, product string) error {
	if product == "" {
		product = "hello"
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(abs, "go.mod"), []byte("module bench.local\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	cfg := fmt.Sprintf("{\n  \"workspace\": %s,\n  \"product\": %s,\n  \"test_command\": [\"go\", \"test\", \"./...\"]\n}\n",
		jsonString(abs), jsonString(product))
	if err := os.WriteFile(filepath.Join(abs, "dudes.json"), []byte(cfg), 0o644); err != nil {
		return err
	}
	if err := git(abs, "init", "-q"); err != nil {
		return err
	}
	_ = git(abs, "config", "user.email", "bench@example.com")
	_ = git(abs, "config", "user.name", "dudes-bench")
	if err := git(abs, "add", "go.mod", "dudes.json"); err != nil {
		return err
	}
	if err := git(abs, "commit", "-q", "-m", "chore: bench fixture"); err != nil {
		return err
	}
	return nil
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %v: %w (%s)", args, err, out)
	}
	return nil
}

// Summarize reads events + two-axis check after a run.
func Summarize(workspace, product string, started time.Time, stub, skipTest bool) (Report, error) {
	rep := Report{
		Product:    product,
		Stub:       stub,
		SkipTest:   skipTest,
		StartedAt:  started.UTC(),
		FinishedAt: time.Now().UTC(),
		Workspace:  workspace,
	}
	rep.DurationMs = rep.FinishedAt.Sub(rep.StartedAt).Milliseconds()
	evs, err := events.ReadAll(workspace)
	if err != nil {
		return rep, err
	}
	for _, ev := range evs {
		switch ev.Type {
		case "gate_fail":
			rep.GateFails++
		case "phase_leave":
			rep.PhaseLeaves++
		}
		if ev.Phase != "" {
			rep.FinalPhase = ev.Phase
		}
	}
	rep.Check = gate.Assess(workspace, product)
	rep.CheckOK = rep.Check.OK()
	rep.OK = rep.GateFails == 0 && rep.CheckOK && rep.FinalPhase == phase.HumanDecision
	return rep, nil
}

// WriteLast stores the report under .dudes/bench/last.json.
func WriteLast(workspace string, rep Report) error {
	dir := filepath.Join(workspace, ".dudes", "bench")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "last.json"), append(data, '\n'), 0o644)
}

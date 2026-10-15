package humangate_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/humangate"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

func TestCanDecide(t *testing.T) {
	st := &runstate.State{Phase: phase.QA, Status: runstate.StatusRunning}
	if err := humangate.CanDecide(st); err == nil {
		t.Fatal("expected error")
	}
	st.Phase = phase.HumanDecision
	st.Status = runstate.StatusWaiting
	if err := humangate.CanDecide(st); err != nil {
		t.Fatal(err)
	}
}

func TestApplySeal(t *testing.T) {
	st := &runstate.State{Phase: phase.HumanDecision, Status: runstate.StatusWaiting, RunID: "r1"}
	if err := humangate.ApplySeal(st); err != nil {
		t.Fatal(err)
	}
	if st.Phase != phase.Sealed || st.Status != runstate.StatusSealed {
		t.Fatalf("got phase=%s status=%s", st.Phase, st.Status)
	}
	if err := humangate.ApplySeal(st); err == nil {
		t.Fatal("expected already sealed")
	}
}

func TestApplyReject(t *testing.T) {
	st := &runstate.State{Phase: phase.HumanDecision, Status: runstate.StatusWaiting, Cycle: 0}
	if err := humangate.ApplyReject(st, "fix the CLI"); err != nil {
		t.Fatal(err)
	}
	if st.Phase != humangate.ResumeAfterReject || st.Status != runstate.StatusRunning {
		t.Fatalf("got phase=%s status=%s", st.Phase, st.Status)
	}
	if st.Cycle != 1 || st.Feedback != "fix the CLI" {
		t.Fatalf("cycle=%d feedback=%q", st.Cycle, st.Feedback)
	}
	fb := humangate.FeedbackMarkdown("demo", st)
	if !strings.Contains(fb, "fix the CLI") || !strings.Contains(fb, string(humangate.ResumeAfterReject)) {
		t.Fatalf("fb=%s", fb)
	}
}

func TestArtifacts(t *testing.T) {
	st := &runstate.State{RunID: "r1", Prompt: "demo", Cycle: 0}
	seal := humangate.SealMarkdown("demo", st)
	if !strings.Contains(seal, "r1") || !strings.Contains(seal, "Human sealed") {
		t.Fatalf("seal=%s", seal)
	}
}

// Package humangate implements the human seal / reject boundary (M6).
package humangate

import (
	"fmt"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// ResumeAfterReject is where the engine restarts the cycle after "not good enough".
const ResumeAfterReject = phase.Implementation

// ErrNotWaiting is returned when seal/reject is attempted outside human_decision.
var ErrNotWaiting = fmt.Errorf("run is not waiting for a human decision (need phase human_decision)")

// ErrAlreadySealed is returned when the run is already sealed.
var ErrAlreadySealed = fmt.Errorf("run is already sealed")

// CanDecide reports whether seal/reject is allowed.
func CanDecide(st *runstate.State) error {
	if st == nil {
		return fmt.Errorf("nil state")
	}
	if st.Phase == phase.Sealed || st.Status == runstate.StatusSealed {
		return ErrAlreadySealed
	}
	if st.Phase != phase.HumanDecision {
		return ErrNotWaiting
	}
	return nil
}

// ApplySeal updates state to sealed release.
func ApplySeal(st *runstate.State) error {
	if err := CanDecide(st); err != nil {
		return err
	}
	st.Phase = phase.Sealed
	st.Status = runstate.StatusSealed
	return nil
}

// ApplyReject records feedback and resumes the cycle at ResumeAfterReject.
func ApplyReject(st *runstate.State, feedback string) error {
	if err := CanDecide(st); err != nil {
		return err
	}
	feedback = strings.TrimSpace(feedback)
	if feedback == "" {
		feedback = "(no feedback text)"
	}
	st.Feedback = feedback
	st.Cycle++
	st.Phase = ResumeAfterReject
	st.Status = runstate.StatusRunning
	return nil
}

// SealSubject is the subject of the seal commit — the human's commit that ends a run.
const SealSubject = "chore(phase): seal [🔹human]"

// SealedByKey is the trailer of the seal commit that names the sealer (F89).
const SealedByKey = "Sealed-by"

// SealMessage is the seal commit's message: its subject, and the Sealed-by trailer naming the sealer
// — the git identity of whoever ran dudes seal, as `Name <email>`.
func SealMessage(sealer string) string {
	return SealSubject + "\n\n" + SealedByKey + ": " + sealer
}

// SealArtifactRel is the durable seal evidence path.
func SealArtifactRel(product string) string {
	if product == "" {
		product = "product"
	}
	return product + "_SEAL.md"
}

// FeedbackArtifactRel is the durable reject/feedback path.
func FeedbackArtifactRel(product string) string {
	if product == "" {
		product = "product"
	}
	return product + "_FEEDBACK.md"
}

// SealMarkdown builds the seal evidence body.
func SealMarkdown(product string, st *runstate.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Seal — %s\n\n", product)
	fmt.Fprintf(&b, "- **run_id:** %s\n", st.RunID)
	fmt.Fprintf(&b, "- **sealed_at:** %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- **cycle:** %d\n", st.Cycle)
	if st.Prompt != "" {
		fmt.Fprintf(&b, "- **prompt:** %s\n", st.Prompt)
	}
	b.WriteString("\nHuman sealed this release. Workers did not self-seal.\n")
	return b.String()
}

// FeedbackMarkdown builds the reject evidence body (call after ApplyReject).
func FeedbackMarkdown(product string, st *runstate.State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Feedback — %s\n\n", product)
	fmt.Fprintf(&b, "- **run_id:** %s\n", st.RunID)
	fmt.Fprintf(&b, "- **rejected_at:** %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "- **resume_phase:** %s\n", st.Phase)
	fmt.Fprintf(&b, "- **cycle:** %d\n\n", st.Cycle)
	b.WriteString("## Not good enough\n\n")
	b.WriteString(st.Feedback)
	if !strings.HasSuffix(st.Feedback, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

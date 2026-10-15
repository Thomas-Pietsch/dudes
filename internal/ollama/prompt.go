package ollama

import (
	"fmt"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

// PhasePrompt builds a single-shot worker prompt (no prior chat). Prior markdown is durable handoff only.
// specRel is the authoritative PRD path when relevant (F12).
// taskOverride, when non-empty, replaces the built-in Task section (F27 recipe instructions).
// feedback is human reject text from a prior cycle (F51); empty when none.
func PhasePrompt(id phase.ID, product, userPrompt, priorMarkdown, contextPack, specRel, taskOverride, feedback, discipline string) string {
	if userPrompt == "" {
		userPrompt = "(no user prompt)"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "You are a documentation worker for product %q in the DUDES engine.\n", product)
	b.WriteString("Reply with Markdown ONLY. No preamble about being an AI.\n")
	b.WriteString("Wrap the entire artifact between these exact lines:\n")
	b.WriteString("===BEGIN ARTIFACT===\n…markdown…\n===END ARTIFACT===\n\n")
	if specRel != "" && (id == phase.QA || id == phase.Documentation || id == phase.Spec) {
		fmt.Fprintf(&b, "## Spec path (authoritative)\n\n`%s`\n\n", specRel)
	}
	if priorMarkdown != "" {
		b.WriteString("## Prior durable artifact (read-only context)\n\n")
		b.WriteString(priorMarkdown)
		b.WriteString("\n\n")
	}
	if contextPack != "" {
		b.WriteString("## Phase-scoped repository pack (read-only)\n\n")
		b.WriteString(contextPack)
		b.WriteString("\n\n")
	}
	b.WriteString("## User intent\n\n")
	b.WriteString(userPrompt)
	if strings.TrimSpace(feedback) != "" {
		b.WriteString("\n\n## Human rejection feedback (must address)\n\n")
		b.WriteString(strings.TrimSpace(feedback))
	}
	b.WriteString("\n\n## Task\n\n")
	if strings.TrimSpace(taskOverride) != "" {
		b.WriteString(strings.TrimSpace(taskOverride))
	} else {
		switch id {
		case phase.Idea:
			b.WriteString("Write a short idea capture: title, problem, proposed product direction.")
		case phase.Strategy:
			b.WriteString("Write a Product Strategy (*_PS.md): market/intent, goals, non-goals, risks. Keep it concise.")
		case phase.Spec:
			b.WriteString("Write a PRD (*_PRD.md) as a machine-checkable contract. Use this exact shape:\n")
			b.WriteString("### Requirement: <short-id>\nThe system SHALL …\n")
			b.WriteString("#### Scenario: <name>\n- GIVEN …\n- WHEN …\n- THEN …\n")
			b.WriteString("At least one requirement with one complete scenario. Short ids (e.g. greet), not whole sentences.")
		case phase.QA:
			b.WriteString("Write a QA report stub (*_QR.md): what was checked, results unknown/stub, open risks. Ground claims in the PRD path above.")
		case phase.Documentation:
			b.WriteString("Write short user-facing docs: what the product does and how to run it (placeholder commands OK). Align with the PRD path above.")
		default:
			fmt.Fprintf(&b, "Produce the durable Markdown artifact for phase %s.", id)
		}
	}
	if strings.TrimSpace(discipline) != "" {
		b.WriteString("\n\n## Discipline (this phase only — do not change the lifecycle)\n\n")
		b.WriteString(strings.TrimSpace(discipline))
	}
	return b.String()
}

// FixFromTestPrompt asks for a fix using ONLY the test output (M4).
func FixFromTestPrompt(testOutput string) string {
	var b strings.Builder
	b.WriteString("You are a coding fix worker. The framework ran tests and they FAILED.\n")
	b.WriteString("Propose a minimal fix. Reply with Markdown explaining the fix, then a ```go code block if code is needed.\n")
	b.WriteString("Do NOT invent a new product. Use ONLY the test output below as evidence.\n\n")
	b.WriteString("## Test output\n\n```\n")
	b.WriteString(truncate(testOutput, 12000))
	b.WriteString("\n```\n")
	return b.String()
}

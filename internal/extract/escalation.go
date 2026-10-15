package extract

import "strings"

// Escalation is a worker's stop report (F74): BLOCKED (missing or contradictory
// input) or CONTRACT (a referenced contract looks wrong). The worker must not
// invent; the engine turns the report into a durable waiting-for-decision.
type Escalation struct {
	Kind string // BLOCKED | CONTRACT
	Text string // the whole report line
}

// FindEscalation returns the first line that starts with "BLOCKED:" or
// "CONTRACT:". Text inside a delimited artifact is the artifact's content
// (a QA report may list blocked hypotheses) and is ignored.
func FindEscalation(raw string) (Escalation, bool) {
	text := raw
	if inner, ok := between(raw, BeginMark, EndMark); ok {
		text = strings.Replace(raw, inner, "", 1)
	}
	for _, line := range strings.Split(text, "\n") {
		l := strings.TrimSpace(line)
		for _, kind := range []string{"BLOCKED", "CONTRACT"} {
			if strings.HasPrefix(l, kind+":") {
				return Escalation{Kind: kind, Text: l}, true
			}
		}
	}
	return Escalation{}, false
}

package phase_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

// One rule source for the flow: the role-play flow in rules/DUDES.md §4 and the engine's phase
// registry describe the same cycle. This test keeps them from drifting apart — a role added,
// renamed or moved in the rules, or a phase added to the engine, fails here until both agree.

// roleToPhase maps each role of the unified flow to the engine phase that carries its work.
var roleToPhase = map[string]phase.ID{
	"CEO":               phase.Idea,
	"Product Visionary": phase.Strategy,
	"PM":                phase.Spec,
	"UI/UX":             phase.Spec,
	"Security":          phase.Spec,
	"Writer":            phase.Documentation,
	"QA (author)":       phase.TestAuthoring,
	"Dev(s)":            phase.Implementation,
	"QA (verify)":       phase.QA,
	"✅ HUMAN GATE":      phase.HumanDecision,
}

func unifiedFlow(t *testing.T) []string {
	t.Helper()
	_, this, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(this), "..", "..", "rules", "DUDES.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	for i, l := range lines {
		if strings.HasPrefix(l, "**The unified flow:**") && i+2 < len(lines) {
			var roles []string
			for _, r := range strings.Split(lines[i+2], "→") {
				roles = append(roles, strings.TrimSpace(r))
			}
			return roles
		}
	}
	t.Fatal("rules/DUDES.md: no **The unified flow:** block")
	return nil
}

func TestRolePlayFlowMatchesTheEngine(t *testing.T) {
	pos := map[phase.ID]int{}
	for i, d := range phase.Default().Order() {
		pos[d.ID] = i
	}
	roles := unifiedFlow(t)
	covered := map[phase.ID]bool{}
	last := -1
	for _, r := range roles {
		ph, ok := roleToPhase[r]
		if !ok {
			t.Errorf("role %q in rules/DUDES.md has no engine phase — map it here or fix the rule", r)
			continue
		}
		covered[ph] = true
		i, known := pos[ph]
		if !known {
			t.Errorf("role %q maps to %q, which is not in the engine registry", r, ph)
			continue
		}
		if i < last {
			t.Errorf("role %q (%s) comes after a later engine phase — the orders disagree", r, ph)
		}
		last = i
	}
	for _, d := range phase.DefaultOrder {
		if d.ID != phase.Sealed && !covered[d.ID] {
			t.Errorf("engine phase %q has no role in the rules/DUDES.md flow", d.ID)
		}
	}
}

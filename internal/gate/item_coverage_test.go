package gate_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
)

// Item-bound: every behavior criterion of the item needs its test (spec.AcceptanceTestName).
func TestItemCoverageNeedsATestPerCriterion(t *testing.T) {
	dir := t.TempDir()
	con := &contract.Contract{
		ID: "x-012", Product: "demo", Item: "X-012",
		Sources: []contract.Source{{ID: "backlog", Root: "workspace", Path: "x.md"}},
		Acceptance: []contract.Acceptance{
			{ID: "x-012-ac-1", Kind: "behavior", Text: "AC-1 saved"},
			{ID: "x-012-ac-2", Kind: "behavior", Text: "AC-2 opens"},
			{ID: "field-1", Kind: "field", Text: "interviews"},
		},
	}
	if err := contract.Save(dir, con); err != nil {
		t.Fatal(err)
	}
	tests := filepath.Join(dir, "tests", "demo_smoke_test.go")
	if err := os.MkdirAll(filepath.Dir(tests), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tests, []byte("package smoke_test\nfunc TestX012AC1(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := gate.RequireRequirementCoverage(dir, "demo")
	if err == nil || !strings.Contains(err.Error(), "TestX012AC2") || strings.Contains(err.Error(), "field-1") {
		t.Fatalf("want AC-2 reported (field criteria need no test), got %v", err)
	}
	if err := os.WriteFile(tests, []byte("package smoke_test\nfunc TestX012AC1(t *testing.T) {}\nfunc TestX012AC2(t *testing.T) {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireRequirementCoverage(dir, "demo"); err != nil {
		t.Fatalf("all criteria tested: %v", err)
	}
}

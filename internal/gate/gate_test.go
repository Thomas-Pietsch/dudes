package gate_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

const validPRD = `# PRD

### Requirement: greet
The CLI SHALL greet.

#### Scenario: default
- GIVEN a binary
- WHEN it runs
- THEN it greets
`

func TestRequireSpec(t *testing.T) {
	dir := gitRepo(t)
	if err := gate.RequireSpec(dir, "demo"); err == nil {
		t.Fatal("expected missing spec error")
	}
	if err := artifact.Write(dir, "demo_PRD.md", []byte("# PRD\n")); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireSpec(dir, "demo"); err != nil {
		t.Fatal(err)
	}
}

// A bound item is the spec in a dev repo: no workspace PRD needed (the spec repo owns it).
func TestRequireSpecItemBound(t *testing.T) {
	dir := gitRepo(t)
	if gate.ItemBound(dir) {
		t.Fatal("no contract: not item-bound")
	}
	if err := contract.Save(dir, &contract.Contract{ID: "x-001", Product: "demo", Item: "X-001"}); err != nil {
		t.Fatal(err)
	}
	if !gate.ItemBound(dir) {
		t.Fatal("contract with item: item-bound")
	}
	if err := gate.RequireSpec(dir, "demo"); err != nil {
		t.Fatal(err)
	}
}

func TestRequireTestsAndEnterImpl(t *testing.T) {
	dir := gitRepo(t)
	_ = artifact.Write(dir, "demo_PRD.md", []byte(validPRD))
	commit(t, dir, "spec")
	if _, err := freeze.Write(dir, "demo", "r"); err != nil {
		t.Fatal(err)
	}
	if err := gate.CheckEnter(dir, "demo", phase.Implementation); err == nil {
		t.Fatal("expected test-before-impl failure")
	}
	rel := filepath.ToSlash(filepath.Join("tests", "demo_smoke_test.go"))
	body := "package smoke_test\n\nimport \"testing\"\n\nfunc TestSmoke(t *testing.T) { t.Log(\"greet\") }\n"
	if err := artifact.Write(dir, rel, []byte(body)); err != nil {
		t.Fatal(err)
	}
	if err := gate.CheckEnter(dir, "demo", phase.Implementation); err != nil {
		t.Fatal(err)
	}
}

// A contract that names its tests (F88) writes them in its red-first task: before it ran, the gate
// does not look for tests; once the files exist they — wherever they are, whatever they are called —
// are what coverage counts.
func TestContractNamingTests(t *testing.T) {
	dir := t.TempDir()
	con := &contract.Contract{ID: "s", Product: "demo", Item: "X1-001", Tests: []string{"pkg/x_test.go", "e2e/x.spec.js"},
		Acceptance: []contract.Acceptance{
			{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1 one"},
			{ID: "x1-001-ac-2", Kind: "behavior", Text: "AC-2 two"},
		}}
	if gate.NamesTests(dir) {
		t.Fatal("no contract names no tests")
	}
	if err := contract.Save(dir, con); err != nil {
		t.Fatal(err)
	}
	if !gate.NamesTests(dir) {
		t.Fatal("NamesTests")
	}
	if err := gate.RequireTests(dir, "demo"); err != nil {
		t.Errorf("RequireTests before the red task: %v", err)
	}
	err := gate.RequireRequirementCoverage(dir, "demo")
	if err == nil || !strings.Contains(err.Error(), "TestX1001AC1") || !strings.Contains(err.Error(), "TestX1001AC2") {
		t.Fatalf("coverage with no test: %v", err)
	}
	if err := artifact.Write(dir, "pkg/x_test.go", []byte("package pkg\nfunc TestX1001AC1(t *testing.T) {}\n")); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireRequirementCoverage(dir, "demo"); err == nil || strings.Contains(err.Error(), "TestX1001AC1") || !strings.Contains(err.Error(), "TestX1001AC2") {
		t.Fatalf("coverage with one test: %v", err)
	}
	// a named file counts whatever its name — a JavaScript spec is no *_test.go
	if err := artifact.Write(dir, "e2e/x.spec.js", []byte("test('TestX1001AC2', () => {})\n")); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireRequirementCoverage(dir, "demo"); err != nil {
		t.Errorf("coverage with both named files: %v", err)
	}
	// without named tests the same gates ask for tests, as before
	con.Tests = nil
	if err := contract.Save(dir, con); err != nil {
		t.Fatal(err)
	}
	if err := gate.RequireRequirementCoverage(dir, "demo"); err == nil || !strings.Contains(err.Error(), "TestX1001AC2") {
		t.Errorf("a file nobody names is not counted: %v", err)
	}
}

func TestCheckLeave(t *testing.T) {
	dir := gitRepo(t)
	if err := gate.CheckLeave(dir, "demo", phase.Idea); err == nil {
		t.Fatal("expected missing idea artifact")
	}
	_ = artifact.Write(dir, "demo_IDEA.md", []byte("idea\n"))
	if err := gate.CheckLeave(dir, "demo", phase.Idea); err != nil {
		t.Fatal(err)
	}
}

func TestConverge(t *testing.T) {
	dir := gitRepo(t)
	findings := gate.Converge(dir, "demo")
	if len(findings) == 0 {
		t.Fatal("expected findings")
	}
	_ = artifact.Write(dir, "demo_IDEA.md", []byte("i\n"))
	_ = artifact.Write(dir, "demo_PS.md", []byte("p\n"))
	_ = artifact.Write(dir, "demo_PRD.md", []byte(validPRD))
	_ = artifact.Write(dir, "tests/demo_smoke_test.go", []byte("package smoke_test\nimport \"testing\"\nfunc TestSmoke(t *testing.T){ t.Log(\"greet\") }\n"))
	_ = artifact.Write(dir, "demo_QR.md", []byte("q\n"))
	_ = artifact.Write(dir, "docs/USER.md", []byte("d\n"))
	_ = artifact.Write(dir, "demo_TA.md", []byte("# TA\n"))
	commit(t, dir, "spec")
	if _, err := freeze.Write(dir, "demo", "r1"); err != nil {
		t.Fatal(err)
	}
	if !gate.ReadyForHuman(dir, "demo") {
		t.Fatalf("findings=%v", gate.Converge(dir, "demo"))
	}
	rep := gate.Assess(dir, "demo")
	if !rep.OK() || len(rep.Spec) != 0 || len(rep.Standards) != 0 {
		t.Fatalf("report=%+v", rep)
	}
	if !artifact.Exists(dir, artifact.TasksPath("demo")) {
		t.Fatal("expected TASKS.md")
	}
}

func TestCheckLeaveSpecShape(t *testing.T) {
	dir := gitRepo(t)
	_ = artifact.Write(dir, "demo_PRD.md", []byte("# PRD\n- bullet\n"))
	if err := gate.CheckLeave(dir, "demo", phase.Spec); err == nil {
		t.Fatal("expected spec-shape failure")
	}
	_ = artifact.Write(dir, "demo_PRD.md", []byte(validPRD))
	if err := gate.CheckLeave(dir, "demo", phase.Spec); err != nil {
		t.Fatal(err)
	}
}

func TestTwoAxisSeparatesFindings(t *testing.T) {
	dir := gitRepo(t)
	rep := gate.Assess(dir, "demo")
	if rep.OK() {
		t.Fatal("empty workspace should fail")
	}
	if len(rep.Standards) == 0 {
		t.Fatal("expected standards findings for missing artifacts")
	}
}

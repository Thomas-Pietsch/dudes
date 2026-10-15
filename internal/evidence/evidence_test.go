package evidence_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/evidence"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

const prd = "### Requirement: greet\nThe CLI SHALL greet.\n#### Scenario: s\nGIVEN a\nWHEN b\nTHEN c\n"

func TestPlanAndApplyRun(t *testing.T) {
	f := evidence.PlanSoftwareRows("demo", "standard", spec.Parse(prd))
	if len(f.Rows) != 1 || f.Rows[0].Status != evidence.StatusPlanned {
		t.Fatalf("%+v", f.Rows)
	}
	f.EnsureFieldRow("pmf", "field interviews — not CI")
	f.ApplyRun("# ok: true\nPASS TestReq_greet\n", "abc123", evidence.Source())
	if f.Rows[0].Status != evidence.StatusGreen || f.Rows[0].Commit != "abc123" {
		t.Fatalf("%+v", f.Rows[0])
	}
	if f.Rows[1].Class != evidence.ClassField || f.Rows[1].Status != evidence.StatusBlocked {
		t.Fatal("field row must stay blocked")
	}
	if !strings.Contains(f.Rows[0].Test, "greet") {
		t.Fatal(f.Rows[0].Test)
	}
}

// Derive: rows from the PRD once tests are authored, results from the test-run note in history.
func TestDeriveFromSpecAndRecordedRun(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	write("demo_PRD.md", prd)
	if err := gitx.AddCommit(dir, []string{"demo_PRD.md"}, "spec"); err != nil {
		t.Fatal(err)
	}

	f, err := evidence.Derive(dir, "demo", "demo_PRD.md", "tests/demo_smoke_test.go")
	if err != nil || len(f.Rows) != 0 {
		t.Fatalf("before test authoring: want no rows, got %+v err=%v", f, err)
	}

	write("tests/demo_smoke_test.go", "package tests\n")
	if err := gitx.AddCommit(dir, []string{"tests/demo_smoke_test.go"}, "tests"); err != nil {
		t.Fatal(err)
	}
	f, _ = evidence.Derive(dir, "demo", "demo_PRD.md", "tests/demo_smoke_test.go")
	if got := f.UnmetSoftware(); len(got) != 1 || got[0] != "greet" {
		t.Fatalf("authored, not run: want greet unmet, got %v", got)
	}

	if err := gitx.AddNote(dir, testrun.NotesRef, "# ok: true\n=== RUN TestReq_greet\n--- PASS: TestReq_greet\nPASS\n"); err != nil {
		t.Fatal(err)
	}
	head, _ := gitx.Head(dir)
	f, _ = evidence.Derive(dir, "demo", "demo_PRD.md", "tests/demo_smoke_test.go")
	if len(f.UnmetSoftware()) != 0 || !strings.HasPrefix(f.Rows[0].Commit, head) {
		t.Fatalf("after a green run: %+v", f.Rows)
	}
}

// A contract that names its tests (F88) has authored tests once one of those files exists: there is
// no tests/ artifact of the test phase, and the rows turn green from the acceptance run all the same.
func TestDeriveNamedTests(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	con := &contract.Contract{ID: "s", Product: "p", Item: "X1-001", Tests: []string{"pkg/x_test.go"}, Acceptance: []contract.Acceptance{
		{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1 one"},
	}}
	if err := contract.Save(dir, con); err != nil {
		t.Fatal(err)
	}
	derive := func() *evidence.File {
		t.Helper()
		f, err := evidence.Derive(dir, "p", "", "tests/p_smoke_test.go")
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	if f := derive(); len(f.Rows) != 0 {
		t.Fatalf("before the red task: %+v", f.Rows)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "x_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "red")
	if f := derive(); len(f.Rows) != 1 || f.Rows[0].Test != "TestX1001AC1" || f.Rows[0].Status != evidence.StatusPlanned {
		t.Fatalf("with the named test: %+v", f.Rows)
	}
	if err := gitx.AddNote(dir, testrun.AcceptanceNotesRef, "--- PASS: TestX1001AC1 (0.00s)\n"); err != nil {
		t.Fatal(err)
	}
	if f := derive(); len(f.Rows) != 1 || f.Rows[0].Status != evidence.StatusGreen {
		t.Fatalf("after a green acceptance run: %+v", f.Rows)
	}
}

// An item-bound workspace: a plain go test names no passing test, the recorded acceptance rerun
// does — its rows turn green from it; a skip is red; a later unit run that fails the test wins.
func TestDeriveFromAcceptanceRun(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	if err := os.MkdirAll(filepath.Join(dir, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tests", "x1001ac1_test.go"), []byte("package tests\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", "-A")
	git("commit", "-q", "-m", "tests")
	con := &contract.Contract{ID: "x", Product: "p", Item: "X1-001", Acceptance: []contract.Acceptance{
		{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1 one"},
		{ID: "x1-001-ac-2", Kind: "behavior", Text: "AC-2 two"},
	}}
	if err := contract.Save(dir, con); err != nil {
		t.Fatal(err)
	}
	note := func(ref, text string) {
		t.Helper()
		if err := gitx.AddNote(dir, ref, text); err != nil {
			t.Fatal(err)
		}
	}
	status := func() map[string]string {
		t.Helper()
		f, err := evidence.Derive(dir, "p", "", "tests")
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]string{}
		for _, r := range f.Rows {
			out[r.Requirement] = r.Status
		}
		return out
	}
	note(testrun.NotesRef, "# ok: true\nok  \tx/tests\t0.1s\n")
	note(testrun.AcceptanceNotesRef, "# ok: true\n--- PASS: TestX1001AC1 (0.00s)\n--- SKIP: TestX1001AC2 (0.00s)\n")
	if got := status(); got["x1-001-ac-1"] != evidence.StatusGreen || got["x1-001-ac-2"] != evidence.StatusRed {
		t.Fatalf("from the acceptance run: %v", got)
	}
	git("commit", "-q", "--allow-empty", "-m", "later")
	note(testrun.NotesRef, "# ok: false\n--- FAIL: TestX1001AC1 (0.00s)\nFAIL\n")
	if got := status(); got["x1-001-ac-1"] != evidence.StatusRed {
		t.Fatalf("a later failing unit run must win: %v", got)
	}
}

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// A task with its own test is a slice of the item: its gate runs that command; the task
// without one runs the engine's test_command, the full suite.
func TestGateCommandSlice(t *testing.T) {
	ws := t.TempDir()
	con := &contract.Contract{ID: "x", Product: "p", Tasks: []contract.Task{
		{ID: "slice", Phase: phase.Implementation, Test: []string{"go", "test", "-run", "AC8$", "./tests/..."}},
		{ID: "rest", Phase: phase.Implementation},
	}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws, TestCommand: []string{"go", "test", "./..."}}

	cmd, slice := gateCommand(cfg, &runstate.State{TaskID: "slice"})
	if !slice || strings.Join(cmd, " ") != "go test -run AC8$ ./tests/..." {
		t.Fatalf("slice task: cmd=%v slice=%v", cmd, slice)
	}
	for _, id := range []string{"rest", ""} {
		cmd, slice = gateCommand(cfg, &runstate.State{TaskID: id})
		if slice || strings.Join(cmd, " ") != "go test ./..." {
			t.Fatalf("task %q: cmd=%v slice=%v", id, cmd, slice)
		}
	}
	if cmd, slice = gateCommand(cfg, nil); slice || len(cmd) != 3 {
		t.Fatalf("no run: cmd=%v slice=%v", cmd, slice)
	}
}

// A task's writable patterns narrow the contract's; out-of-scope worker output is discarded, the
// rest committed.
func TestTaskScopeDiscardsOutside(t *testing.T) {
	con := &contract.Contract{Writable: []string{"app/*", "go.mod"}, Tasks: []contract.Task{
		{ID: "slice", Writable: []string{"app/internal/config/*"}},
		{ID: "rest"},
	}}
	if got := writable(con, &runstate.State{TaskID: "slice"}); len(got) != 1 {
		t.Fatalf("slice scope %v", got)
	}
	if got := writable(con, &runstate.State{TaskID: "rest"}); len(got) != 2 {
		t.Fatalf("contract scope %v", got)
	}

	ws := t.TempDir()
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "dudes@test"}, {"config", "user.name", "dudes-test"}, {"commit", "-q", "--allow-empty", "-m", "init"}} {
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	worker := func(string) (string, error) {
		for _, f := range []string{"app/internal/config/config.go", "app/internal/db/db.go"} {
			p := filepath.Join(ws, filepath.FromSlash(f))
			if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(p, []byte("package x\n"), 0o644); err != nil {
				return "", err
			}
		}
		return "done", nil
	}
	changed, _, err := implPass(ws, writable(con, &runstate.State{TaskID: "slice"}), worker, nil, "go", "chore(phase): implementation", "")
	if err != nil || !changed {
		t.Fatalf("implPass changed=%v err=%v", changed, err)
	}
	if _, err := os.Stat(filepath.Join(ws, "app", "internal", "db", "db.go")); !os.IsNotExist(err) {
		t.Fatalf("out-of-scope file kept: %v", err)
	}
	out, err := exec.Command("git", "-C", ws, "show", "--name-only", "--format=", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(out)) != "app/internal/config/config.go" {
		t.Fatalf("committed %q %v", out, err)
	}
}

// A failed gate on HEAD goes back to the worker with the in-scope files it names; once a newer
// commit exists the old failure is history.
func TestGateFailureOnHead(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	for _, f := range []string{"pkg/a.go", "pkg/c_test.go", "pkg/d_test.go", "other/b.go"} {
		if err := os.MkdirAll(filepath.Join(ws, filepath.Dir(f)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ws, f), []byte("package x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("add", ".")
	git("commit", "-q", "-m", "work")
	// go test names a test package's setup failure by its absolute path, a failing test by its bare
	// file name
	abs := filepath.Join(ws, "pkg", "c_test.go")
	note := "# dudes test\n# args: [go test]\n# exit: 1\n# ok: false\n\npkg\\a.go:3:1: undefined: yaml.SyntaxError\nother/b.go:1:1: x\n" +
		abs + ":187:1: wrong signature for TestX\n--- FAIL: TestD (0.05s)\n    d_test.go:23: build failed\n"
	if err := gitx.AddNote(ws, testrun.NotesRef, note); err != nil {
		t.Fatal(err)
	}
	gate := []string{"go", "test"}
	log, files, refs := gateFailure(ws, gate, []string{"pkg/*"})
	if !strings.Contains(log, "yaml.SyntaxError") || len(files) != 3 || files[0] != "pkg/a.go" || files[1] != "pkg/c_test.go" || files[2] != "pkg/d_test.go" {
		t.Fatalf("log=%q files=%v", log, files)
	}
	if len(refs) != 1 || refs[0] != "other/b.go" { // outside the scope: handed over to read
		t.Fatalf("refs=%v", refs)
	}
	if log, _, _ = gateFailure(ws, []string{"go", "test", "-skip", "AC1$"}, []string{"pkg/*"}); log != "" {
		t.Fatalf("another command's failure handed over: %q", log)
	}
	git("commit", "-q", "--allow-empty", "-m", "next")
	if log, files, _ = gateFailure(ws, gate, []string{"pkg/*"}); log != "" || files != nil {
		t.Fatalf("stale failure handed over: %q %v", log, files)
	}
}

// In a replay of real product slices the gate refused a green run whose acceptance rerun skipped
// AC-9, but the plain `go test` log on HEAD said ok, so the worker got nothing and skipped again,
// three times. The skip, its reason and the file it names must go back.
func TestGateFailureSkippedAcceptance(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	if err := os.MkdirAll(filepath.Join(ws, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "tests", "x001ac9_test.go"), []byte("package x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git("add", ".")
	git("commit", "-q", "-m", "work")
	gate := []string{"go", "test"}
	if err := gitx.AddNote(ws, testrun.NotesRef, "# dudes test\n# args: [go test]\n# exit: 0\n# ok: true\n\nok  \tx/tests\t0.1s\n"); err != nil {
		t.Fatal(err)
	}
	if log, _, _ := gateFailure(ws, gate, []string{"tests/*"}); log != "" {
		t.Fatalf("a green run without a skip handed over: %q", log)
	}
	acc := "=== RUN   TestX001AC7\n--- PASS: TestX001AC7 (0.00s)\n=== RUN   TestX001AC9\n" +
		"    x001ac9_test.go:76: fixture directory does not exist, skipping\n--- SKIP: TestX001AC9 (0.00s)\nPASS\n"
	if err := gitx.AddNote(ws, testrun.AcceptanceNotesRef, acc); err != nil {
		t.Fatal(err)
	}
	log, files, _ := gateFailure(ws, gate, []string{"tests/*"})
	if !strings.Contains(log, "x001ac9_test.go:76: fixture directory does not exist") || !strings.Contains(log, skipRule) ||
		strings.Contains(log, "TestX001AC7") {
		t.Fatalf("log=%q", log)
	}
	if len(files) != 1 || files[0] != "tests/x001ac9_test.go" {
		t.Fatalf("files=%v", files)
	}
}

// An item's test phase writes one test file per behavior criterion beside the index; a task's
// worker gets the test files inside its scope.
func TestItemTestsPerCriterion(t *testing.T) {
	ws := t.TempDir()
	con := &contract.Contract{ID: "x", Product: "demo", Item: "X-001", Acceptance: []contract.Acceptance{
		{ID: "X-001-AC-5", Kind: "behavior", Text: "routes"},
		{ID: "X-001-AC-8", Kind: "behavior", Text: "config"},
		{ID: "T2", Kind: "field", Text: "field hypothesis"},
	}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws, Product: "demo"}
	idx, _, err := authoredTests(cfg, &runstate.State{})
	if err != nil || !strings.Contains(idx, "tests/x001ac8_test.go") || strings.Contains(idx, "T2") {
		t.Fatalf("index %q: %v", idx, err)
	}
	if err := os.MkdirAll(filepath.Join(ws, "tests"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "tests", "demo_smoke_test.go"), []byte(idx), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, _ := taskTests(cfg, []string{"tests/x001ac8_test.go"}, nil); len(got) != 1 || got[0] != "tests/demo_smoke_test.go" {
		t.Fatalf("before the criterion files exist, the index: %v", got)
	}
	_, acs, _ := itemAcceptance(cfg)
	for _, ac := range acs {
		if err := os.WriteFile(filepath.Join(ws, filepath.FromSlash(spec.ItemTestFile(ac.ID))), []byte(spec.ItemGoTest("X-001", ac)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if got, ro := taskTests(cfg, []string{"app/*", "tests/x001ac8_test.go"}, []string{"go", "test", "-skip", "^TestX001AC(1|2)$"}); len(got) != 1 || got[0] != "tests/x001ac8_test.go" || len(ro) != 1 || ro[0] != "tests/x001ac5_test.go" {
		t.Fatalf("slice: %v, read-only %v", got, ro)
	}
	// The implementation after a red-first task: its tests are read-only, and only those its gate runs.
	if got, ro := taskTests(cfg, []string{"app/*"}, []string{"go", "test", "-skip", "^TestX001AC5$"}); len(got) != 0 || len(ro) != 1 || ro[0] != "tests/x001ac8_test.go" {
		t.Fatalf("implementation: edit %v, read-only %v", got, ro)
	}
	if got, _ := taskTests(cfg, nil, nil); len(got) != 2 {
		t.Fatalf("whole item: %v", got)
	}
}

// A red-first task (expect: fail) is done when its tests run and fail; tests that already pass, or
// do not build, are not a red test of the criterion.
func TestRedTaskGate(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(ws, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module red\n\ngo 1.22\n")
	con := &contract.Contract{ID: "x", Product: "p", Tasks: []contract.Task{
		{ID: "test", Phase: phase.Implementation, Test: []string{"go", "test", "./..."}, Expect: "fail"},
	}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws}
	st := &runstate.State{TaskID: "test"}
	for _, c := range []struct {
		name, test string
		want       int
	}{
		{"red", "package red\nimport \"testing\"\nfunc TestAC(t *testing.T) { t.Fatal(\"not built yet\") }\n", 0},
		{"green already", "package red\nimport \"testing\"\nfunc TestAC(t *testing.T) {}\n", 1},
		{"does not build", "package red\nimport \"testing\"\nfunc TestAC(t *testing.T) { undefinedCall() }\n", 1},
	} {
		write("red_test.go", c.test)
		git("add", "-A")
		git("commit", "-q", "--allow-empty", "-m", c.name)
		if got := runTestsGate(cfg, st); got != c.want {
			t.Errorf("%s: gate = %d, want %d", c.name, got, c.want)
		}
	}
	// a red test sent back after its implementation ran may pass — the review decides
	write("red_test.go", "package red\nimport \"testing\"\nfunc TestAC(t *testing.T) {}\n")
	git("add", "-A")
	git("commit", "-q", "--allow-empty", "-m", "redo")
	if got := runTestsGate(cfg, &runstate.State{TaskID: "test", Redo: map[string]string{"test": "fix the path"}}); got != 0 {
		t.Errorf("redone red test that passes: gate = %d, want 0", got)
	}
	if err := (&contract.Contract{ID: "x", Product: "p", Tasks: []contract.Task{{ID: "t", Expect: "maybe"}}}).Validate(); err == nil {
		t.Error("expect: maybe validated")
	}
}

// A green go test gate whose acceptance test skipped is not green: the skip hid the criterion.
func TestSkippedAcceptanceFailsGate(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(ws, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module skip\n\ngo 1.22\n")
	con := &contract.Contract{ID: "x", Product: "p", Item: "X1-001", Acceptance: []contract.Acceptance{
		{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1 it works"},
	}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws, TestCommand: []string{"go", "test", "./..."}}
	for _, c := range []struct {
		name, body string
		want       int
	}{
		{"skips", `t.Skip("fixture not found")`, 1},
		{"tests", `if 1+1 != 2 { t.Fatal("math") }`, 0},
	} {
		write("ac_test.go", "package skip\nimport \"testing\"\nfunc TestX1001AC1(t *testing.T) { "+c.body+" }\n")
		git("add", "-A")
		git("commit", "-q", "-m", c.name)
		if got := runTestsGate(cfg, &runstate.State{}); got != c.want {
			t.Errorf("%s: gate = %d, want %d", c.name, got, c.want)
		}
		// the verbose rerun is recorded as evidence on the tested commit
		acc, err := testrun.Acceptance(ws)
		head, _ := gitx.HeadFull(ws)
		want := testrun.Outcome{Result: map[int]string{0: testrun.Pass, 1: testrun.Skip}[c.want], Commit: head}
		if err != nil || acc["TestX1001AC1"] != want {
			t.Errorf("%s: acceptance note %v, want %v (err=%v)", c.name, acc, want, err)
		}
	}
}

// Red for a bound item: its acceptance test fails, while a package of the code that does not exist
// yet fails to set up — a gate names ./app/... before the implementation.
func TestRedTaskGateMissingCode(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	for name, body := range map[string]string{
		"go.mod":            "module red\n\ngo 1.22\n",
		"tests/ac1_test.go": "package tests\nimport \"testing\"\nfunc TestX1001AC1(t *testing.T) { t.Fatal(\"service not built yet\") }\n",
	} {
		_ = os.MkdirAll(filepath.Join(ws, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(ws, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "red")
	con := &contract.Contract{ID: "x", Product: "p", Item: "X1-001",
		Acceptance: []contract.Acceptance{{ID: "x1-001-ac-1", Kind: "behavior", Text: "AC-1"}},
		Tasks:      []contract.Task{{ID: "test", Phase: phase.Implementation, Test: []string{"go", "test", "./tests/...", "./code/..."}, Expect: "fail"}}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	if got := runTestsGate(config.Config{Workspace: ws}, &runstate.State{TaskID: "test"}); got != 0 {
		t.Fatalf("gate = %d: the acceptance test is red, the missing code package does not make it a build error", got)
	}
}

// A scoped task's worker gets the tracked code inside its scope — what it extends — within a budget.
func TestScopeCode(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	for name, size := range map[string]int{"semver/semver.go": 100, "release/doc.go": 10, "slug/slug.go": 10, "semver/big.go": scopeCodeBudget} {
		_ = os.MkdirAll(filepath.Join(ws, filepath.Dir(name)), 0o755)
		if err := os.WriteFile(filepath.Join(ws, name), []byte(strings.Repeat("x", size)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("add", "-A")
	git("commit", "-q", "-m", "code")
	files, skipped := scopeCode(ws, []string{"semver/*", "release/*"}, []string{"release/doc.go"})
	if len(files) != 1 || files[0] != "semver/semver.go" || skipped != 1 {
		t.Fatalf("files %v skipped %d: want semver.go, big.go over budget, doc.go already there, slug outside", files, skipped)
	}
	if files, _ := scopeCode(ws, nil, nil); files != nil {
		t.Fatalf("no scope, no extra files: %v", files)
	}
}

// The configured formatter runs on the worker's changed files of the configured extensions.
func TestFormatter(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	ws := t.TempDir()
	for _, f := range []string{"a.go", "b.yml"} {
		if err := os.WriteFile(filepath.Join(ws, f), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Config{Format: config.FormatConfig{
		Command: []string{"sh", "-c", `for f; do echo formatted >> "$f"; done`, "sh"},
		Ext:     []string{".go"},
	}}
	if formatter(config.Config{}, ws) != nil {
		t.Fatal("no format command, no formatter")
	}
	formatter(cfg, ws)([]string{"a.go", "b.yml", "gone.go"})
	got, _ := os.ReadFile(filepath.Join(ws, "a.go"))
	other, _ := os.ReadFile(filepath.Join(ws, "b.yml"))
	if string(got) != "x\nformatted\n" || string(other) != "x\n" {
		t.Fatalf("a.go %q, b.yml %q: want only the .go file formatted", got, other)
	}
}

// Status counts the contract's completed tasks over all phases: the bench steps a run to its end
// on it (F82).
func TestContractProgress(t *testing.T) {
	ws := t.TempDir()
	cfg := config.Config{Workspace: ws}
	if d, n := contractProgress(cfg, &runstate.State{}); d != 0 || n != 0 {
		t.Fatalf("no contract: %d/%d", d, n)
	}
	con := &contract.Contract{ID: "x", Product: "p", Tasks: []contract.Task{
		{ID: "author-tests", Phase: phase.TestAuthoring},
		{ID: "red", Phase: phase.Implementation, Expect: "fail"},
		{ID: "impl", Phase: phase.Implementation},
	}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	if d, n := contractProgress(cfg, &runstate.State{Completed: []string{"author-tests", "red", "other"}}); d != 2 || n != 3 {
		t.Fatalf("progress %d/%d, want 2/3", d, n)
	}
}

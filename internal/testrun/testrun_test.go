package testrun_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// A run is recorded as a git note on the tested commit; ReadLastLog reads it back.
func TestRunRecordsANoteOnTheTestedCommit(t *testing.T) {
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"}, {"config", "user.email", "dudes@test"}, {"config", "user.name", "dudes-test"},
		{"commit", "--allow-empty", "-m", "base"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// no go module here — go test fails; we only check capture and recording
	res, err := testrun.Run(dir, []string{"go", "test", "./..."})
	if err != nil || res == nil {
		t.Fatalf("res=%v err=%v", res, err)
	}
	if res.OK {
		t.Fatal("expected failure without a module")
	}
	log, commit, ok, err := testrun.LastRun(dir)
	head, _ := gitx.HeadFull(dir)
	if err != nil || !ok || commit != head || !strings.Contains(log, "# ok: false") {
		t.Fatalf("ok=%v err=%v commit=%s head=%s log=%q", ok, err, commit, head, log)
	}
	// the note reads back as the run: its command, outcome and output (the review packet, F88)
	rec := testrun.ParseRecord(log)
	if rec.Args != "[go test ./...]" || rec.OK || rec.Exit != res.ExitCode || rec.Duration == "" || rec.Dirty {
		t.Errorf("record %+v", rec)
	}
	if strings.TrimSpace(rec.Output) != strings.TrimSpace(res.Output) {
		t.Errorf("output %q, want %q", rec.Output, res.Output)
	}
}

// A note that is not a recorded run is all output; a build failure's "# pkg" lines stay output.
func TestParseRecordKeepsTheOutput(t *testing.T) {
	note := "# dudes test\n# args: [go test]\n# exit: 1\n# ok: false\n# duration: 1.5s\n# tree: dirty (1 uncommitted path(s) — the run tested more than this commit)\n\n# bench/semver\nsemver.go:3:1: syntax error\n"
	rec := testrun.ParseRecord(note)
	if rec.Args != "[go test]" || rec.Exit != 1 || rec.OK || rec.Duration != "1.5s" || !rec.Dirty {
		t.Errorf("header %+v", rec)
	}
	if rec.Output != "# bench/semver\nsemver.go:3:1: syntax error\n" {
		t.Errorf("output %q", rec.Output)
	}
	if other := testrun.ParseRecord("# bench/semver\nx\n"); other.Output != "# bench/semver\nx\n" || other.Args != "" {
		t.Errorf("a note without the header: %+v", other)
	}
}

// Without a commit there is nothing to attach the record to — the run still reports.
func TestRunWithoutCommitIsNotRecorded(t *testing.T) {
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init").CombinedOutput(); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	res, err := testrun.Run(dir, []string{"go", "version"})
	if err != nil || res == nil || !res.OK {
		t.Fatalf("res=%v err=%v", res, err)
	}
	if _, ok, err := testrun.ReadLastLog(dir); ok || err != nil {
		t.Fatalf("want no record, ok=%v err=%v", ok, err)
	}
}

// A verbose log names each test's result; a test reported twice keeps its worst.
func TestResultsFromAVerboseLog(t *testing.T) {
	log := "=== RUN   TestA\n--- PASS: TestA (0.00s)\n--- FAIL: TestB (0.01s)\n    --- SKIP: TestC/sub (0.00s)\n" +
		"--- PASS: TestD (0.00s)\r\n--- FAIL: TestD (0.00s)\n--- PASS: TestAB (0.00s)\nPASS: TestE\n--- PASS: TestF\n"
	got := testrun.Results(log)
	want := map[string]string{"TestA": testrun.Pass, "TestB": testrun.Fail, "TestC/sub": testrun.Skip, "TestD": testrun.Fail, "TestAB": testrun.Pass}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

// Acceptance runs accumulate over HEAD's history: the newest run per test wins, tests only an older
// run reported keep its result (slice 1 stays green after slice 2), unit-run notes do not count.
func TestAcceptanceAccumulatesNewestPerTest(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	git("commit", "-q", "--allow-empty", "-m", "base")
	if acc, err := testrun.Acceptance(dir); len(acc) != 0 || err != nil {
		t.Fatalf("no note yet: %v err=%v", acc, err)
	}
	note := func(ref, text string) {
		t.Helper()
		if err := gitx.AddNote(dir, ref, text); err != nil {
			t.Fatal(err)
		}
	}
	note(testrun.AcceptanceNotesRef, "--- PASS: TestX1001AC1 (0.00s)\n--- FAIL: TestX1001AC2 (0.00s)\n")
	slice1 := git("rev-parse", "HEAD")
	git("commit", "-q", "--allow-empty", "-m", "slice 2")
	note(testrun.AcceptanceNotesRef, "--- PASS: TestX1001AC2 (0.00s)\n--- PASS: TestX1001AC3 (0.00s)\n")
	note(testrun.NotesRef, "--- FAIL: TestX1001AC1 (0.00s)\n")
	slice2 := git("rev-parse", "HEAD")
	git("commit", "-q", "--allow-empty", "-m", "untested")
	acc, err := testrun.Acceptance(dir)
	want := map[string]testrun.Outcome{
		"TestX1001AC1": {Result: testrun.Pass, Commit: slice1}, // only the older run: kept
		"TestX1001AC2": {Result: testrun.Pass, Commit: slice2}, // the newer run overrides the fail
		"TestX1001AC3": {Result: testrun.Pass, Commit: slice2},
	}
	if err != nil || len(acc) != len(want) {
		t.Fatalf("acc=%v err=%v", acc, err)
	}
	for name, w := range want {
		if acc[name] != w {
			t.Errorf("%s = %v, want %v", name, acc[name], w)
		}
	}
}

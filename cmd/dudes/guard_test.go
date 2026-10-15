package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// In a replay of real product slices, one slice's tests were green, then the next slice's first
// pass broke the build and the following steps never repaired it — eight green acceptance tests
// ended red. A failed pass that breaks what was green is undone; the next worker starts from the
// green state with that pass's gate output. A failed pass that breaks nothing stays.
func TestGuardRevertsAPassThatBreaksGreen(t *testing.T) {
	ws := t.TempDir()
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	write := func(rel, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(ws, filepath.Dir(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ws, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git("init", "-q")
	git("config", "user.email", "dudes@test")
	git("config", "user.name", "dudes-test")
	write(".git/info/exclude", ".dudes/\n")
	write("go.mod", "module g\n\ngo 1.21\n")
	write("svc/svc.go", "package svc\n\nfunc Health() string { return \"ok\" }\n")
	write("svc/svc_test.go", "package svc\n\nimport \"testing\"\n\nfunc TestX001AC5(t *testing.T) {\n\tif Health() != \"ok\" {\n\t\tt.Fatal(\"health\")\n\t}\n}\n")
	git("add", "-A")
	git("commit", "-q", "-m", "slice green")
	green := git("rev-parse", "HEAD")
	if err := gitx.AddNote(ws, testrun.AcceptanceNotesRef, "=== RUN   TestX001AC5\n--- PASS: TestX001AC5 (0.00s)\nPASS\n"); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{Workspace: ws, TestCommand: []string{"go", "test", "./..."}}
	st := &runstate.State{RunID: "r"}

	// a failed pass that breaks nothing green stays
	write("svc/new.go", "package svc\n\nfunc Later() int { return 1 }\n")
	git("add", "-A")
	git("commit", "-q", "-m", "pass: harmless")
	harmless := git("rev-parse", "HEAD")
	guardRegression(cfg, st, green)
	if git("rev-parse", "HEAD") != harmless {
		t.Fatal("a pass that broke nothing green was reverted")
	}

	// a failed pass that stops the build is undone
	write("svc/svc.go", "package svc\n\nfunc Health() string { return undefined }\n")
	git("add", "-A")
	git("commit", "-q", "-m", "pass: breaks the build")
	broken := git("rev-parse", "HEAD")
	note := "# dudes test\n# args: [go test ./...]\n# exit: 1\n# ok: false\n\nsvc/svc.go:3:38: undefined: undefined\n"
	if err := gitx.AddNote(ws, testrun.NotesRef, note); err != nil {
		t.Fatal(err)
	}
	guardRegression(cfg, st, harmless)
	head := git("rev-parse", "HEAD")
	if head == broken || git("diff", "--stat", harmless, head) != "" {
		t.Fatalf("not back at the last green state: HEAD %s", head)
	}
	if !strings.Contains(git("log", "--format=%H", "-n", "3"), broken) {
		t.Fatal("the undone pass left the history")
	}
	if pass, why := revertedPass(ws, head); pass != broken || !strings.Contains(why, "TestX001AC5") {
		t.Fatalf("revertedPass = %q, %q", pass, why)
	}
	log, files, _ := gateFailure(ws, cfg.TestCommand, []string{"svc/*"})
	if !strings.HasPrefix(log, "Your last pass was reverted") || !strings.Contains(log, "undefined: undefined") {
		t.Fatalf("log=%q", log)
	}
	if len(files) != 1 || files[0] != "svc/svc.go" {
		t.Fatalf("files=%v", files)
	}
}

package aider_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
)

// The task comes first, the spec last: a task under the spec reads as documentation.
func TestMessageTaskFirst(t *testing.T) {
	msg := aider.Message{
		Product: "demo", Intent: "build cli", Task: "Implement DEMO-1", Feedback: "needs --help",
		Tests: []string{"tests/demo_smoke_test.go"}, Writable: []string{"cmd/*"},
		SpecRel: "demo_PRD.md", Spec: "# PRD\n\nreq\n", Pack: "pack bits",
	}.String()
	for _, want := range []string{"Implement DEMO-1", "build cli", "needs --help", "`tests/demo_smoke_test.go`", "`cmd/*`", "BLOCKED:", "demo_PRD.md", "req", "pack bits"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message misses %q:\n%s", want, msg)
		}
	}
	if !(strings.Index(msg, "Implement DEMO-1") < strings.Index(msg, "Done when") &&
		strings.Index(msg, "Done when") < strings.Index(msg, "## Spec")) {
		t.Fatalf("order: task, done when, spec:\n%s", msg)
	}
}

// A task slice names its own command as done and leaves the other tests to later tasks.
func TestMessageSliceDoneCmd(t *testing.T) {
	msg := aider.Message{
		Product: "demo", Task: "Config loader (AC-8)", Tests: []string{"tests/demo_smoke_test.go"},
		DoneCmd:     []string{"go", "test", "-run", "TestDEMO1AC8$", "./tests/..."},
		Constraints: []string{"Standard library only."},
	}.String()
	for _, want := range []string{"`go test -run TestDEMO1AC8$ ./tests/...` passes", "`tests/demo_smoke_test.go`", "Leave every other test", "- Standard library only."} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message misses %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "These failing tests pass") {
		t.Fatalf("a slice must not ask for every test:\n%s", msg)
	}
}

// A failed gate comes back before the done-when: the fix starts from the compiler's words.
func TestMessageGateLogFirst(t *testing.T) {
	msg := aider.Message{Task: "Config loader", GateLog: "config.go:49:23: undefined: yaml.SyntaxError"}.String()
	if i, j := strings.Index(msg, "yaml.SyntaxError"), strings.Index(msg, "## Done when"); i < 0 || i > j {
		t.Fatalf("gate log missing or after done-when:\n%s", msg)
	}
}

func TestReadFlags(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := aider.ReadFlags(dir, []string{"none", "notes.md", "missing.md", "worktree"})
	if len(got) != 2 || got[0] != "--read" || got[1] != "notes.md" {
		t.Fatalf("got=%v", got)
	}
}

func TestWriteStubImpl(t *testing.T) {
	dir := t.TempDir()
	rel, err := aider.WriteStubImpl(dir, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if rel != "cmd/demo/main.go" {
		t.Fatalf("rel=%s", rel)
	}
	data, err := os.ReadFile(filepath.Join(dir, "cmd", "demo", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "demo stub") {
		t.Fatalf("body=%s", data)
	}
}

func TestRunWithFakeBinary(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeAider(t, dir)
	_, err := aider.Run(context.Background(), dir, "implement demo", aider.Config{Bin: bin, Timeout: 0})
	if err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(dir, "aider-ran.txt")
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("fake aider did not write marker: %v", err)
	}
	// The prompt travels in a file: packs exceed the Windows command-line limit.
	msg, err := os.ReadFile(filepath.Join(dir, ".dudes", "worker_message.md"))
	if err != nil || string(msg) != "implement demo" {
		t.Fatalf("message file: %q %v", msg, err)
	}
}

// A relative workspace still hands the worker a message file it can open from inside it.
func TestRunRelativeWorkspace(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Dir(dir)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })
	bin := writeFakeAider(t, dir)
	if _, err := aider.Run(context.Background(), filepath.Base(dir), "implement demo", aider.Config{Bin: bin}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "msg-found.txt")); err != nil {
		t.Fatalf("worker could not open the message file: %v", err)
	}
}

// A launcher starts the worker: its own arguments, then the worker binary, then the worker's flags.
func TestRunThroughLauncher(t *testing.T) {
	dir := t.TempDir()
	launcher := writeFakeAider(t, dir)
	cfg := aider.Config{Bin: "aider-in-image", Launch: []string{launcher}}
	if cfg.Exe() != launcher {
		t.Fatalf("exe=%s", cfg.Exe())
	}
	if _, err := aider.Run(context.Background(), dir, "implement demo", cfg); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "first-arg.txt"))
	if err != nil || strings.TrimSpace(string(got)) != "aider-in-image" {
		t.Fatalf("launcher got first arg %q (%v), want the worker binary", got, err)
	}
}

// Aider sees only the item's writable scope: everything top-level is ignored, the scope re-included.
func TestIgnoreFlags(t *testing.T) {
	dir := t.TempDir()
	if flags, err := aider.IgnoreFlags(dir, nil); err != nil || flags != nil {
		t.Fatalf("no scope: %v %v", flags, err)
	}
	flags, err := aider.IgnoreFlags(dir, []string{"app/*", "go.mod", "web/src/*"})
	if err != nil || len(flags) != 2 || flags[0] != "--aiderignore" {
		t.Fatalf("flags %v %v", flags, err)
	}
	data, err := os.ReadFile(flags[1])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\n/*\n", "\n!/app/\n", "\n!/go.mod\n", "\n!/web/\n"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("aiderignore misses %q:\n%s", want, data)
		}
	}
}

func TestRunRejectsEmptyMessage(t *testing.T) {
	_, err := aider.Run(context.Background(), t.TempDir(), "  ", aider.Config{Bin: "true"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeFakeAider(t *testing.T, dir string) string {
	t.Helper()
	var path string
	var body string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "fake-aider.bat")
		// Write a marker file in cwd (workspace) so the test can assert Run sets Dir.
		body = "@echo off\r\necho ran>aider-ran.txt\r\necho %1>first-arg.txt\r\nfor %%a in (%*) do if exist %%a echo found>msg-found.txt\r\n"
	} else {
		path = filepath.Join(dir, "fake-aider")
		body = "#!/bin/sh\necho ran > aider-ran.txt\necho \"$1\" > first-arg.txt\nfor a; do [ -f \"$a\" ] && echo found > msg-found.txt; done\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(path, 0o755)
	}
	// Ensure executable is resolvable as given path
	if _, err := exec.LookPath(path); err != nil {
		// LookPath may fail for relative paths; absolute path to Command is enough.
	}
	return path
}

// A timeout ends the run even when the launcher's child still holds the output pipe (a container
// client): the engine does not wait on it, and the launcher learns the limit (WORKER_TIMEOUT).
func TestRunTimeoutThroughLauncher(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	cfg := aider.Config{Bin: "aider", Launch: []string{"sh", "-c", `echo "$WORKER_TIMEOUT" > limit.txt; sleep 60 & sleep 60`}, Timeout: time.Second}
	start := time.Now()
	if _, err := aider.Run(context.Background(), dir, "implement demo", cfg); err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("want a timeout error, got %v", err)
	}
	if d := time.Since(start); d > 40*time.Second {
		t.Fatalf("run took %s — waited on the launcher's child", d)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "limit.txt")); strings.TrimSpace(string(got)) != "1" {
		t.Fatalf("WORKER_TIMEOUT=%q, want 1", got)
	}
}

// A red-first task writes the test and must not implement: its command is expected to fail.
func TestMessageRedTask(t *testing.T) {
	msg := aider.Message{
		Product: "demo", Task: "Acceptance test (AC-5)", Tests: []string{"tests/demo1ac5_test.go"},
		DoneCmd: []string{"go", "test", "-run", "TestDEMO1AC5$", "./tests/..."}, Red: true,
	}.String()
	for _, want := range []string{"runs and **fails**", "`tests/demo1ac5_test.go`", "Do not write the implementation"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message misses %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "write the code that makes") {
		t.Fatalf("a red task must not ask for the implementation:\n%s", msg)
	}
}

// The implementation after a red-first task gets its tests read-only, and a way out if one is wrong.
func TestMessageReadOnlyTests(t *testing.T) {
	msg := aider.Message{
		Product: "demo", Task: "Config code (AC-8)", ReadOnlyTests: []string{"tests/demo1ac8_test.go"},
		DoneCmd: []string{"go", "test", "./tests/..."},
	}.String()
	for _, want := range []string{"written and read-only: `tests/demo1ac8_test.go`", "BLOCKED: <test>", "Then write the code"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("message misses %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "replace the body") {
		t.Fatalf("read-only tests must not be rewritten:\n%s", msg)
	}
}

// A worker cannot seal (F89): Aider is started with its shell commands off — what the model's reply
// suggests is never run — and with no test or lint command of its own, and the engine commits for it.
func TestRunKeepsTheWorkersShellOff(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	// the launcher is a shell that writes the arguments the worker would have been started with
	cfg := aider.Config{Bin: "aider", Launch: []string{"sh", "-c", `printf '%s\n' "$@" > args.txt`}, Timeout: 30 * time.Second}
	if _, err := aider.Run(context.Background(), dir, "implement demo", cfg); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "args.txt"))
	if err != nil {
		t.Fatal(err)
	}
	args := strings.Fields(string(data))
	has := func(flag string) bool {
		for _, a := range args {
			if a == flag {
				return true
			}
		}
		return false
	}
	for _, want := range []string{"--no-suggest-shell-commands", "--no-auto-commits", "--no-dirty-commits", "--no-auto-lint"} {
		if !has(want) {
			t.Errorf("aider is not started with %s: %v", want, args)
		}
	}
	for _, bad := range []string{"--suggest-shell-commands", "--auto-test", "--test-cmd", "--lint-cmd", "--auto-commits"} {
		if has(bad) {
			t.Errorf("aider is started with %s: %v", bad, args)
		}
	}
}

// The code worker's image holds Aider and git, never the engine: nothing of the repository is copied
// into it, so there is no dudes binary in the container to seal with (F89). The engine runs on the host.
func TestWorkerImageHasNoEngine(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "docker", "worker", "Dockerfile"))
	if err != nil {
		t.Fatal(err)
	}
	instructions := 0
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		instructions++
		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "copy ") || strings.HasPrefix(lower, "add ") || strings.Contains(lower, "dudes") || strings.Contains(lower, "go install") || strings.Contains(lower, "go build") {
			t.Errorf("the worker image takes in something of the engine or the repository: %q", line)
		}
	}
	if instructions == 0 {
		t.Fatal("the Dockerfile holds no instructions")
	}
}

// A reply cut off at the model's context fails the step instead of reading as "no changes".
func TestRunTokenLimit(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	cfg := aider.Config{Bin: "aider", Launch: []string{"sh", "-c", `echo "> Model ollama_chat/m has hit a token limit!"; echo "> Input tokens: ~27,375 of 32,768"`}, Timeout: 30 * time.Second}
	if _, err := aider.Run(context.Background(), dir, "implement demo", cfg); err == nil || !strings.Contains(err.Error(), "reply was cut off") {
		t.Fatalf("want a context-full error, got %v", err)
	}
}

// Aider reports each model call's tokens as it prints them, rounded; a run's cost is their sum (F88).
func TestUsage(t *testing.T) {
	for _, c := range []struct {
		name, out      string
		sent, received int
		ok             bool
	}{
		{"one call", "Applied edit to semver/semver.go\nTokens: 2.1k sent, 302 received.\n", 2100, 302, true},
		{"calls summed", "Tokens: 9.8k sent, 1.2k received.\nreflection\nTokens: 12k sent, 85 received.\n", 21800, 1285, true},
		{"cache figures and a cost", "Tokens: 12k sent, 1.2k cache write, 3.4k cache hit, 1.3k received. Cost: $0.01 message, $0.02 session.\n", 12000, 1300, true},
		{"wrapped by the console", "Tokens: 12k sent, 1.2k cache write, 3.4k cache hit,\n1.3k received.\n", 12000, 1300, true},
		{"the chat history's quote", "> Tokens: 950 sent, 12 received.\n", 950, 12, true},
		{"a reply cut at a token limit", "> Model ollama_chat/m has hit a token limit!  \n> Token counts below are approximate.  \n>  \n> Input tokens: ~24,638 of 262,144  \n> Output tokens: ~11,183 of 262,144  \n", 24638, 11183, true},
		{"a finished call, then a cut one", "Tokens: 9.8k sent, 1.2k received.\nModel m has hit a token limit!\nInput tokens: ~24,523 of 32,768\nOutput tokens: ~11,068 of 32,768\n", 34323, 12268, true},
		{"a cut call with its own Tokens line", "Model m has hit a token limit!\nInput tokens: ~27,375 of 32,768\nOutput tokens: ~29 of 32,768\nTokens: 27k sent, 29 received.\n", 27000, 29, true},
		{"no report", "Applied edit to semver/semver.go\n", 0, 0, false},
		{"the input's tokens, not a report", "> Input tokens: ~27,375 of 32,768\n", 0, 0, false},
	} {
		sent, received, ok := aider.Usage(c.out)
		if sent != c.sent || received != c.received || ok != c.ok {
			t.Errorf("%s: got %d/%d %v, want %d/%d %v", c.name, sent, received, ok, c.sent, c.received, c.ok)
		}
	}
}

// A run that failed after the model answered still returns what the worker wrote: its tokens count
// (a reply cut off at the context had filled it).
func TestRunReturnsTheOutputOfAFailedCall(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	dir := t.TempDir()
	cfg := aider.Config{Bin: "aider", Launch: []string{"sh", "-c", `echo "Model ollama_chat/m has hit a token limit!"; echo "Tokens: 27k sent, 29 received."`}, Timeout: 30 * time.Second}
	out, err := aider.Run(context.Background(), dir, "implement demo", cfg)
	if err == nil || !strings.Contains(err.Error(), "reply was cut off") {
		t.Fatalf("want the context-full error, got %v", err)
	}
	if sent, received, ok := aider.Usage(out); !ok || sent != 27000 || received != 29 {
		t.Errorf("the failed call's tokens: %d/%d %v in %q", sent, received, ok, out)
	}
}

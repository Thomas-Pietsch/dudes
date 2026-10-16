package pi

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelperProcess stands in for the launcher (scripts/worker.sh) and Pi: it writes the arguments
// it got after "--" and its stdin to ./helper.out in the working directory, and prints a Pi event.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("DUDES_PI_HELPER") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	in, _ := io.ReadAll(os.Stdin)
	rec := strings.Join(args, "\x00") + "\x01" + string(in) + "\x01" + os.Getenv("WORKER_TIMEOUT")
	if err := os.WriteFile("helper.out", []byte(rec), 0o644); err != nil {
		os.Exit(3)
	}
	os.Stdout.WriteString(`{"type":"tool_execution_start","toolName":"read"}` + "\n")
	os.Exit(0)
}

// The files travel inside the prompt on stdin, never as @arguments: Git's sh on Windows expands an
// @argument as a response file and crashes (found 2026-10-16).
func TestRunThroughALauncherHandsOverFilesToolsAndStdin(t *testing.T) {
	t.Setenv("DUDES_PI_HELPER", "1")
	ws := t.TempDir()
	for name, body := range map[string]string{"tests/a_test.go": "package a", "main.go": "package main"} {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(ws, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ws, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := Config{
		Bin:    "pi",
		Launch: []string{os.Args[0], "-test.run=TestHelperProcess", "--", "--image", "pi-image"},
		Tools:  []string{"read", "edit"},
		Extra:  []string{"--provider", "ollama"},
		Read:   []string{"tests/a_test.go"},
		Files:  []string{"main.go", "gone.go"},
	}
	msg := "do the \"task\"\nline two"
	out, err := Run(context.Background(), ws, msg, []string{"go.mod"}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if Parse(out).ToolCalls != 1 {
		t.Fatalf("out=%q", out)
	}
	raw, err := os.ReadFile(filepath.Join(ws, "helper.out"))
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(string(raw), "\x01")
	got := strings.Split(parts[0], "\x00")
	want := []string{"--image", "pi-image", "pi", "-p", stdinCue, "--mode", "json", "--tools", "read,edit", "--provider", "ollama"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("args:\n got %q\nwant %q", got, want)
	}
	wantIn := "<file name=\"tests/a_test.go\" access=\"read-only\">\npackage a\n</file>\n<file name=\"main.go\">\npackage main\n</file>\n\n" + msg
	if parts[1] != wantIn {
		t.Fatalf("stdin=%q\nwant %q", parts[1], wantIn)
	}
	if parts[2] != "1800" {
		t.Fatalf("WORKER_TIMEOUT=%q", parts[2])
	}
}

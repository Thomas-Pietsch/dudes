// Package pi invokes the Pi CLI as an optional second code worker (F55).
package pi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/awake"
	"github.com/Thomas-Pietsch/dudes/internal/redact"
	"github.com/Thomas-Pietsch/dudes/internal/sandbox"
	wsbound "github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Config for the Pi subprocess.
type Config struct {
	Bin     string
	Timeout time.Duration
	Extra   []string
	Launch  []string // command Pi is started through, e.g. sh scripts/worker.sh --image <pi image> (the container)
	Tools   []string // Pi's own tool names (--tools); empty → the task's tools list
	Files   []string // workspace-relative files Pi may edit, put into the prompt (fileBlocks)
	Read    []string // workspace-relative files Pi reads and must not change, put into the prompt the same way
}

// Exe is the program the engine starts: the launcher's, else Pi's own.
func (c Config) Exe() string {
	if len(c.Launch) > 0 {
		return c.Launch[0]
	}
	if c.Bin == "" {
		return "pi"
	}
	return c.Bin
}

// DefaultConfig returns orchestrator-safe defaults.
func DefaultConfig() Config {
	return Config{
		Bin:     "pi",
		Timeout: 30 * time.Minute,
	}
}

// LookPath reports whether the program that starts Pi (the launcher, else bin) is available.
func LookPath(bin string) error {
	if bin == "" {
		bin = "pi"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%s not found on PATH (install Pi or set pi.bin / pi.launch / use aider / -stub)", bin)
	}
	return nil
}

// Run executes `pi -p <message>` in workspace (fresh process = context boundary).
// Run invokes the worker once and returns its (redacted) stdout — with an error as well, as far as
// Pi wrote any: a call that failed after the model answered still cost its tokens and tool calls
// (Parse). Pi's JSON mode exits 0 even when the last reply is an error, so the caller reads the
// events, not the exit code.
func Run(ctx context.Context, workspaceRoot, message string, tools []string, cfg Config) (string, error) {
	if cfg.Bin == "" {
		cfg.Bin = "pi"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Minute
	}
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("empty pi message")
	}

	// The files go into the prompt as Pi's own @file does, but from here: an @file argument cannot pass the
	// launcher on Windows (Git's sh expands it as a response file), and a prompt with files always
	// travels on stdin.
	files := fileBlocks(workspaceRoot, cfg.Read, cfg.Files)
	args, stdin := invocation(files+message, len(cfg.Launch) > 0 || files != "")
	if len(cfg.Tools) > 0 {
		tools = cfg.Tools
	} else {
		tools = sandbox.Filter(tools)
	}
	if len(tools) > 0 {
		args = append(args, "--tools", strings.Join(tools, ","))
	}
	args = append(args, cfg.Extra...)

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if len(cfg.Launch) > 0 {
		args = append(append(append([]string{}, cfg.Launch[1:]...), cfg.Bin), args...)
	}
	cmd := exec.CommandContext(ctx, cfg.Exe(), args...)
	if err := sandbox.Apply(cmd, workspaceRoot); err != nil {
		return "", fmt.Errorf("pi workspace: %w", err)
	}
	if _, err := wsbound.Resolve(workspaceRoot, "."); err != nil {
		return "", fmt.Errorf("pi workspace: %w", err)
	}
	if len(cfg.Launch) > 0 {
		// The timeout kills the launcher, not what it started: the container ends itself
		// (scripts/worker.sh runs the command under `timeout`), and Wait stops waiting for output
		// pipes that the launcher's children still hold.
		cmd.Env = append(cmd.Env, fmt.Sprintf("WORKER_TIMEOUT=%d", int(cfg.Timeout.Seconds())))
		cmd.WaitDelay = 15 * time.Second
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Standby would freeze the worker's container mid-step (the Docker VM sleeps with the host).
	release := awake.Hold()
	err := cmd.Run()
	release()
	out := redact.String(stdout.String())
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out, fmt.Errorf("pi: timed out after %s (pi.timeout_sec)", cfg.Timeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("pi: %s", truncate(redact.String(msg), 800))
	}
	return out, nil
}

// fileBlocks is the files' contents as the first part of the prompt, one block per file in the form
// Pi's @file gives (<file name="…">), the path relative to the workspace; a read-only file says so.
// A file that does not exist (yet) is left out.
func fileBlocks(workspaceRoot string, read, edit []string) string {
	var b strings.Builder
	add := func(rel, access string) {
		abs, err := wsbound.Resolve(workspaceRoot, rel)
		if err != nil {
			return
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			return
		}
		fmt.Fprintf(&b, "<file name=%q%s>\n%s\n</file>\n", filepath.ToSlash(rel), access, data)
	}
	for _, f := range read {
		add(f, ` access="read-only"`)
	}
	for _, f := range edit {
		add(f, "")
	}
	if b.Len() > 0 {
		b.WriteString("\n")
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// maxArgMessage keeps the -p prompt under the Windows command-line limit
// (32767 chars, minus room for the other arguments).
const maxArgMessage = 30000

// stdinCue is the -p prompt when the task travels on stdin, which pi merges into the prompt in
// print mode.
const stdinCue = "Carry out the task given on stdin."

// invocation returns pi's base arguments and stdin for message: short prompts as the -p argument,
// longer ones on stdin, so the Windows command-line limit no longer caps a worker's prompt. Through a
// launcher (stdin true) the prompt always travels on stdin: an argument would pass through the
// shell, wsl.exe and docker, and their quoting of newlines, quotes and backslashes is not ours.
func invocation(message string, stdin bool) ([]string, io.Reader) {
	if len(message) <= maxArgMessage && !stdin {
		return []string{"-p", message, "--mode", "json"}, nil
	}
	return []string{"-p", stdinCue, "--mode", "json"}, strings.NewReader(message)
}

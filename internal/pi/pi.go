// Package pi invokes the Pi CLI as an optional second code worker (F55).
package pi

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/redact"
	"github.com/Thomas-Pietsch/dudes/internal/sandbox"
	wsbound "github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Config for the Pi subprocess.
type Config struct {
	Bin     string
	Timeout time.Duration
	Extra   []string
}

// DefaultConfig returns orchestrator-safe defaults.
func DefaultConfig() Config {
	return Config{
		Bin:     "pi",
		Timeout: 30 * time.Minute,
	}
}

// LookPath reports whether the pi binary is available.
func LookPath(bin string) error {
	if bin == "" {
		bin = "pi"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return fmt.Errorf("%s not found on PATH (install Pi or set pi.bin / use aider / -stub)", bin)
	}
	return nil
}

// Run executes `pi -p <message>` in workspace (fresh process = context boundary).
// Run invokes the worker once and returns its (redacted) stdout.
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

	args, stdin := invocation(message)
	if filtered := sandbox.Filter(tools); len(filtered) > 0 {
		args = append(args, "--tools", strings.Join(filtered, ","))
	}
	args = append(args, cfg.Extra...)

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cfg.Bin, args...)
	if err := sandbox.Apply(cmd, workspaceRoot); err != nil {
		return "", fmt.Errorf("pi workspace: %w", err)
	}
	if _, err := wsbound.Resolve(workspaceRoot, "."); err != nil {
		return "", fmt.Errorf("pi workspace: %w", err)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdin = stdin
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("pi: %s", truncate(redact.String(msg), 800))
	}
	return redact.String(stdout.String()), nil
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
// longer ones on stdin, so the Windows command-line limit no longer caps a worker's prompt.
func invocation(message string) ([]string, io.Reader) {
	if len(message) <= maxArgMessage {
		return []string{"-p", message, "--mode", "json"}, nil
	}
	return []string{"-p", stdinCue, "--mode", "json"}, strings.NewReader(message)
}

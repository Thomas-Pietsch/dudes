// Package aider invokes the Aider CLI as a bounded code worker (M5).
package aider

import (
	"bytes"
	"context"
	"fmt"
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

// Config for the Aider subprocess.
type Config struct {
	Bin     string        // default "aider"
	Model   string        // e.g. ollama/qwen2.5-coder:7b
	Timeout time.Duration // 0 = 30m
	Extra   []string      // additional CLI flags
	Files   []string      // workspace-relative files the worker may edit (added to its chat)
	Env     []string      // KEY=value added to the worker's environment (e.g. OLLAMA_API_BASE)
	Launch  []string      // command the worker is started through, e.g. sh scripts/worker.sh (its image)
}

// Exe is the program the engine starts: the launcher's, else the worker's own.
func (c Config) Exe() string {
	if len(c.Launch) > 0 {
		return c.Launch[0]
	}
	if c.Bin == "" {
		return "aider"
	}
	return c.Bin
}

// DefaultConfig returns orchestrator-safe defaults.
func DefaultConfig() Config {
	return Config{
		Bin:     "aider",
		Timeout: 30 * time.Minute,
	}
}

// Run executes Aider in workspace with a single --message (fresh process = context boundary).
// Always passes --yes and --no-auto-commits so the orchestrator owns Git.
// Run invokes the worker once and returns its (redacted) stdout.
func Run(ctx context.Context, workspace, message string, cfg Config) (string, error) {
	if cfg.Bin == "" {
		cfg.Bin = "aider"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Minute
	}
	if strings.TrimSpace(message) == "" {
		return "", fmt.Errorf("empty aider message")
	}

	// --yes answers every Aider question, so nothing may ask one that reaches outside the
	// workspace — and git is the engine's: no commits at all, not even of dirty files: no .gitignore edit (outside the item's scope), no fetching URLs found in the
	// spec (which also offers to install a browser), no shell commands, no update check, no analytics.
	args := []string{"--yes", "--no-auto-commits", "--no-dirty-commits", "--no-gitignore", "--no-detect-urls",
		"--no-suggest-shell-commands", "--no-check-update", "--no-show-release-notes", "--analytics-disable",
		// Nobody watches the terminal: rich re-rendering of a long reply kept Aider busy for minutes
		// after the model finished, and crashed on the Windows console.
		"--no-pretty",
		// The engine's gate judges the work. Aider's own lint misread valid Postgres SQL and, with
		// --yes, sent the worker into fix rounds (an instance trial).
		"--no-auto-lint"}
	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}
	// The prompt (item pack or PRD) is often larger than the OS command-line
	// limit (~32k chars on Windows), so it goes through a file under .dudes/,
	// which is engine state and never committed as worker output. Absolute: Aider runs in the
	// workspace, so a path relative to the engine's directory would not resolve there.
	msgFile, err := filepath.Abs(filepath.Join(workspace, ".dudes", "worker_message.md"))
	if err != nil {
		return "", fmt.Errorf("aider message: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(msgFile), 0o755); err != nil {
		return "", fmt.Errorf("aider message: %w", err)
	}
	if err := os.WriteFile(msgFile, []byte(message), 0o644); err != nil {
		return "", fmt.Errorf("aider message: %w", err)
	}
	args = append(args, "--message-file", msgFile)
	args = append(args, cfg.Extra...)
	args = append(args, cfg.Files...)

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	if len(cfg.Launch) > 0 {
		args = append(append(append([]string{}, cfg.Launch[1:]...), cfg.Bin), args...)
	}
	cmd := exec.CommandContext(ctx, cfg.Exe(), args...)
	if err := sandbox.Apply(cmd, workspace); err != nil {
		return "", fmt.Errorf("aider workspace: %w", err)
	}
	cmd.Env = append(cmd.Env, cfg.Env...)
	if len(cfg.Launch) > 0 {
		// The timeout kills the launcher, not what it started: the container ends itself
		// (scripts/worker.sh runs the command under `timeout`), and Wait stops waiting for output
		// pipes that the launcher's children still hold.
		cmd.Env = append(cmd.Env, fmt.Sprintf("WORKER_TIMEOUT=%d", int(cfg.Timeout.Seconds())))
		cmd.WaitDelay = 15 * time.Second
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Standby would freeze the worker's container mid-step (the Docker VM sleeps with the host).
	release := awake.Hold()
	err = cmd.Run()
	release()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("aider: timed out after %s (aider.timeout_sec)", cfg.Timeout)
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("aider: %s", truncate(redact.String(msg), 800))
	}
	out := redact.String(stdout.String())
	// A reply cut off at the model's context is no answer — it read as "no file changes" before
	// (an instance trial: 29 output tokens, five steps lost). Either limit cuts it: the context, or
	// the reply cap — another slice stopped at ~10k of num_predict 12288 rewriting whole files.
	if i := strings.Index(out, "has hit a token limit"); i >= 0 {
		return "", fmt.Errorf("aider: the model's reply was cut off — raise its num_ctx (context) or num_predict (reply cap), or narrow the task's refs or scope:\n%s", truncate(out[i:], 400))
	}
	return out, nil
}

// Message is the code worker's input in the micro-agent shape (CORE §2.4): the task first,
// then when it is done (the failing tests) and the rules of the output, then the spec as
// reference. A task buried under the spec reads as documentation: an instance trial's worker
// answered "no changes required" four times.
type Message struct {
	Product  string
	Intent   string   // the run's prompt
	Task     string   // recipe/contract task title; empty → implement the product
	Feedback string   // human reject text from a prior cycle (F51)
	GateLog  string   // the failed test gate on the current commit — fixed first
	Tests    []string // failing tests the work turns green — also handed to the worker as files to edit
	DoneCmd  []string // the task's own test command — a slice of the item; empty → all of Tests
	Red      bool     // the task writes the tests and its command must fail (contract expect: fail)
	// ReadOnlyTests are acceptance tests a red-first task wrote: this task makes them pass and
	// cannot change them (outside its writable).
	ReadOnlyTests []string
	Writable      []string // the item's writable paths (F78); empty → the repository
	Constraints   []string // the contract's constraints — rules for every task of the item
	SpecRel       string   // workspace-relative spec (PRD or item pack, F12/F71)
	Spec          string
	Pack          string // phase-scoped repository pack
	Discipline    string
}

// String renders the message.
func (m Message) String() string {
	var b strings.Builder
	b.WriteString("# Task\n\n")
	if t := strings.TrimSpace(m.Task); t != "" {
		b.WriteString(t + "\n")
	} else {
		fmt.Fprintf(&b, "Implement product %q per the spec below.\n", m.Product)
	}
	if t := strings.TrimSpace(m.Intent); t != "" {
		fmt.Fprintf(&b, "Intent: %s\n", t)
	}
	if t := strings.TrimSpace(m.Feedback); t != "" {
		b.WriteString("\n## Human rejection feedback (must address)\n\n" + t + "\n")
	}
	if t := strings.TrimSpace(m.GateLog); t != "" && m.Red {
		b.WriteString("\n## The last run of your tests\n\nFailing is the goal of this task, but only because the behaviour is missing — not a build error, not a test that can never pass (a wrong path, a wrong command), and not a hang: every wait (a server starting, a container) is bounded and fails with a message. Output:\n\n```text\n" + t + "\n```\n")
	} else if t != "" {
		b.WriteString("\n## The test gate failed on your last change — fix this first\n\nThe files it names are in the chat. Gate output:\n\n```text\n" + t + "\n```\n")
	}
	b.WriteString("\n## Done when\n\n")
	if len(m.DoneCmd) > 0 && m.Red {
		fmt.Fprintf(&b, "`%s` runs and **fails**, because the behaviour is not built yet — this task writes the test, nothing else.", strings.Join(m.DoneCmd, " "))
		if len(m.Tests) > 0 {
			fmt.Fprintf(&b, " Replace the body of each test in %s with the real test of its acceptance criterion, in place, keeping its name: through the public interface, every case the criterion names, expected values taken from the spec.", strings.Join(quoted(m.Tests), ", "))
		}
		b.WriteString(" Do not write the implementation: the next task does, and it cannot change this test.\n")
	} else if len(m.DoneCmd) > 0 {
		fmt.Fprintf(&b, "`%s` passes — this task is that slice of the work item, nothing more.", strings.Join(m.DoneCmd, " "))
		if len(m.Tests) > 0 {
			fmt.Fprintf(&b, " The tests it selects are in %s: replace the body of each with the real test of its acceptance criterion, in place — no second test file — and keep its name (it carries the ID). Leave every other test as it is — later tasks own them.", strings.Join(quoted(m.Tests), ", "))
		}
		if len(m.ReadOnlyTests) > 0 {
			fmt.Fprintf(&b, " The acceptance tests are written and read-only: %s — make them pass; you cannot change them. If one cannot pass whatever the code does (a wrong path, a wrong command), reply `BLOCKED: <test> — <why>` instead.", strings.Join(quoted(m.ReadOnlyTests), ", "))
		}
		b.WriteString(" Then write the code that makes them pass.\n")
	} else if len(m.Tests) > 0 {
		fmt.Fprintf(&b, "These failing tests pass: %s. Each test names one acceptance criterion; replace its body with the real test of that criterion and keep the test name (it carries the ID). Then write the code that makes it pass.\n", strings.Join(quoted(m.Tests), ", "))
	} else {
		b.WriteString("The repository's tests pass and the spec's requirements are implemented.\n")
	}
	b.WriteString("\n## Rules\n\n")
	b.WriteString("- The code does not exist yet: create the files the spec names, at the paths it names.\n")
	if len(m.Writable) > 0 {
		fmt.Fprintf(&b, "- Write only inside: %s.\n", strings.Join(quoted(m.Writable), ", "))
	}
	for _, c := range m.Constraints {
		if c = strings.TrimSpace(c); c != "" {
			b.WriteString("- " + c + "\n")
		}
	}
	b.WriteString("- Return the complete content of every file you create or change.\n")
	b.WriteString("- If the spec is missing or contradicts itself for what you need, reply with one line `BLOCKED: <what is missing>` instead of guessing.\n")
	if t := strings.TrimSpace(m.Discipline); t != "" {
		b.WriteString("\n## Discipline (this phase only — do not change the lifecycle)\n\n" + t + "\n")
	}
	if strings.TrimSpace(m.Spec) != "" {
		b.WriteString("\n## Spec (reference — authoritative, read-only")
		if m.SpecRel != "" {
			fmt.Fprintf(&b, ", `%s`", m.SpecRel)
		}
		b.WriteString(")\n\n" + m.Spec + "\n")
	}
	if strings.TrimSpace(m.Pack) != "" {
		b.WriteString("\n## Repository pack (read-only)\n\n" + m.Pack + "\n")
	}
	return b.String()
}

func quoted(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = "`" + x + "`"
	}
	return out
}

// LookPath reports whether the aider binary is available.
func LookPath(bin string) error {
	if bin == "" {
		bin = "aider"
	}
	_, err := exec.LookPath(bin)
	if err != nil {
		return fmt.Errorf("%s not found on PATH (install Aider or set aider.bin / use -stub)", bin)
	}
	return nil
}

// StubImplRel is the file WriteStubImpl writes, relative to the workspace: cmd/<product>/main.go.
func StubImplRel(product string) string {
	if product == "" {
		product = "product"
	}
	return "cmd/" + product + "/main.go"
}

// WriteStubImpl drops a minimal placeholder when Aider is unavailable (-stub).
func WriteStubImpl(workspaceRoot, product string) (rel string, err error) {
	if product == "" {
		product = "product"
	}
	rel = StubImplRel(product)
	dir, err := wsbound.Resolve(workspaceRoot, filepath.Join("cmd", product))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	body := fmt.Sprintf("// Code generated by dudes -stub (M5). DO NOT use in production.\npackage main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(%q)\n}\n", product+" stub")
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(body), 0o644); err != nil {
		return "", err
	}
	return rel, nil
}

// ReadFlags maps existing workspace files in the tool list to Aider --read (F56).
func ReadFlags(workspaceRoot string, tools []string) []string {
	var extra []string
	for _, t := range sandbox.Filter(tools) {
		abs, err := wsbound.Resolve(workspaceRoot, t)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		extra = append(extra, "--read", t)
	}
	return extra
}

// IgnoreFlags hides every top-level path outside the work item's writable patterns from Aider
// (repo map, files it pulls into the chat): the worker sees what it may change (F78). The ignore
// file is engine state under .dudes/. No patterns → no flags.
func IgnoreFlags(workspaceRoot string, writable []string) ([]string, error) {
	var b strings.Builder
	b.WriteString("# Written by dudes from .dudes/writable — Aider sees only the item's scope\n/*\n")
	n := 0
	for _, p := range writable {
		p = strings.Trim(strings.TrimSpace(filepath.ToSlash(p)), "/")
		if p == "" {
			continue
		}
		if top, _, nested := strings.Cut(p, "/"); nested {
			fmt.Fprintf(&b, "!/%s/\n", top)
		} else {
			fmt.Fprintf(&b, "!/%s\n", p)
		}
		n++
	}
	if n == 0 {
		return nil, nil
	}
	abs, err := filepath.Abs(filepath.Join(workspaceRoot, ".dudes", "aiderignore"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return nil, fmt.Errorf("aider ignore: %w", err)
	}
	if err := os.WriteFile(abs, []byte(b.String()), 0o644); err != nil {
		return nil, fmt.Errorf("aider ignore: %w", err)
	}
	return []string{"--aiderignore", abs}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

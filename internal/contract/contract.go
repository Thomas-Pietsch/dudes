// Package contract is the versioned execution contract (F58, F65).
package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const relPath = ".dudes/contract.json"

// Source points at an original document section (never a summary).
type Source struct {
	ID      string `json:"id"`
	Root    string `json:"root"`
	Path    string `json:"path"`
	Section string `json:"section,omitempty"`
	Rev     string `json:"rev,omitempty"`
}

// Acceptance is one checkable claim.
type Acceptance struct {
	ID     string `json:"id"`
	Source string `json:"source,omitempty"`
	Kind   string `json:"kind"` // behavior | constraint | slo | human | field
	Text   string `json:"text,omitempty"`
}

// Task is one bounded unit inside a phase (F65).
type Task struct {
	ID       string   `json:"id"`
	Phase    phase.ID `json:"phase"`
	Title    string   `json:"title"`
	Inputs   []string `json:"inputs,omitempty"`
	Output   string   `json:"output,omitempty"`
	Skills   []string `json:"skills,omitempty"`
	Rules    []string `json:"rules,omitempty"`
	Worker   string   `json:"worker,omitempty"`   // ollama | aider | stub | pi | human
	Approval string   `json:"approval,omitempty"` // none | before | after
	Tools    []string `json:"tools,omitempty"`
	Variant  string   `json:"variant,omitempty"`
	// Test is the command whose pass ends this task — a slice of the item. Empty → the engine's
	// test_command, the full suite: the phase's last task leaves it empty.
	Test []string `json:"test,omitempty"`
	// Expect "fail" makes the task red first: its gate must fail — the tests compile and run but
	// the behaviour is not there yet. A test task writes a criterion's test this way; the next task
	// implements with that test outside its writable, so it cannot weaken it. Empty → pass.
	Expect string `json:"expect,omitempty"`
	// Writable narrows the contract's writable paths for this task (F78): what a slice may change.
	Writable []string `json:"writable,omitempty"`
	// Refs narrows the item pack to these references ("TA:§20.4") for this task; the rules and
	// the task text stay. Empty → the whole pack.
	Refs []string `json:"refs,omitempty"`
}

// Contract is a projection of approved sources for one increment.
type Contract struct {
	ID          string       `json:"id"`
	Product     string       `json:"product"`
	Variant     string       `json:"variant,omitempty"` // standard | pilot | demo
	Item        string       `json:"item,omitempty"`    // backlog work item; its pack is the worker spec input (F71)
	Sources     []Source     `json:"sources"`
	Constraints []string     `json:"constraints,omitempty"`
	Writable    []string     `json:"writable,omitempty"` // path patterns the item may change; enforced by hooks/pre-commit (F78)
	Acceptance  []Acceptance `json:"acceptance,omitempty"`
	Tasks       []Task       `json:"tasks,omitempty"`
	// Tests names the files the acceptance tests live in (workspace-relative) when they sit where the
	// language needs them — beside the code, as a Go internal package demands — and not under tests/
	// (F88). The contract's red-first task writes them: the engine then writes no placeholder test
	// per criterion, starts at implementation, and its coverage and acceptance checks count these
	// files. Empty → test_authoring writes one file per criterion (spec.ItemTestFile).
	Tests []string `json:"tests,omitempty"`
}

// NamesTests reports whether the contract names the files of its acceptance tests (Tests).
func (c *Contract) NamesTests() bool { return c != nil && len(c.Tests) > 0 }

// ExistingTests are the named test files that exist under workspaceRoot, in the contract's order.
func (c *Contract) ExistingTests(workspaceRoot string) []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, rel := range c.Tests {
		abs, err := workspace.Resolve(workspaceRoot, rel)
		if err != nil {
			continue
		}
		if fi, err := os.Stat(abs); err == nil && !fi.IsDir() {
			out = append(out, rel)
		}
	}
	return out
}

// Path relative to workspace.
func Path() string { return relPath }

// Load reads the contract. ok=false if missing.
func Load(workspaceRoot string) (*Contract, bool, error) {
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("contract: %w", err)
	}
	var c Contract
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false, fmt.Errorf("contract: parse: %w", err)
	}
	return &c, true, nil
}

// Save writes the contract.
func Save(workspaceRoot string, c *Contract) error {
	if c == nil {
		return fmt.Errorf("contract: nil")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(abs, append(data, '\n'), 0o644)
}

// Validate mechanical checks (not LLM).
func (c *Contract) Validate() error {
	if c == nil {
		return fmt.Errorf("contract: nil")
	}
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("contract: empty id")
	}
	if strings.TrimSpace(c.Product) == "" {
		return fmt.Errorf("contract: empty product")
	}
	seen := map[string]bool{}
	for _, t := range c.Tasks {
		if t.ID == "" {
			return fmt.Errorf("contract: task missing id")
		}
		if seen[t.ID] {
			return fmt.Errorf("contract: duplicate task %s", t.ID)
		}
		seen[t.ID] = true
		if t.Phase != "" && !phase.Default().Contains(t.Phase) {
			return fmt.Errorf("contract: task %s unknown phase %s", t.ID, t.Phase)
		}
		switch strings.ToLower(t.Approval) {
		case "", "none", "before", "after":
		default:
			return fmt.Errorf("contract: task %s bad approval %q", t.ID, t.Approval)
		}
		switch strings.ToLower(t.Worker) {
		case "", "ollama", "aider", "stub", "pi", "human":
		default:
			return fmt.Errorf("contract: task %s bad worker %q", t.ID, t.Worker)
		}
		switch t.Expect {
		case "", "pass", "fail":
		default:
			return fmt.Errorf("contract: task %s bad expect %q (pass | fail)", t.ID, t.Expect)
		}
	}
	return nil
}

// TasksFor returns tasks for a phase in order.
func (c *Contract) TasksFor(id phase.ID) []Task {
	if c == nil {
		return nil
	}
	var out []Task
	for _, t := range c.Tasks {
		if t.Phase == id {
			out = append(out, t)
		}
	}
	return out
}

// TaskByID finds a task.
func (c *Contract) TaskByID(id string) (Task, bool) {
	if c == nil {
		return Task{}, false
	}
	for _, t := range c.Tasks {
		if t.ID == id {
			return t, true
		}
	}
	return Task{}, false
}

// NextTask returns the first task in phase after afterID (empty = first).
func (c *Contract) NextTask(id phase.ID, afterID string) (Task, bool) {
	list := c.TasksFor(id)
	if len(list) == 0 {
		return Task{}, false
	}
	if afterID == "" {
		return list[0], true
	}
	for i, t := range list {
		if t.ID == afterID && i+1 < len(list) {
			return list[i+1], true
		}
	}
	return Task{}, false
}

// NeedsDecision reports whether the task must pause for a human before work (F66).
func (t Task) NeedsDecision() bool {
	a := strings.ToLower(strings.TrimSpace(t.Approval))
	return a == "before" || t.Worker == "human"
}

// NeedsAfterDecision reports a pause after the invoke (F66).
func (t Task) NeedsAfterDecision() bool {
	return strings.ToLower(strings.TrimSpace(t.Approval)) == "after"
}

// Project builds the item-bound contract (F71). Sources and acceptance are
// projected from the work item; constraints, tasks and variant are kept from prev.
func Project(product, item string, sources []Source, acceptance []Acceptance, prev *Contract) *Contract {
	c := &Contract{
		ID:         strings.ToLower(item),
		Product:    product,
		Item:       item,
		Sources:    sources,
		Acceptance: acceptance,
	}
	if prev != nil {
		c.Variant = prev.Variant
		c.Constraints = prev.Constraints
		c.Tasks = prev.Tasks
	}
	return c
}

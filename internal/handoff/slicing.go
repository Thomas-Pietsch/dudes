// Package handoff prepares one slice of a work item for the engine (F88): the role's slicing, the
// repository's tracked dudes/ folder, a worktree on a feature branch with the spec roots pinned
// beside it, and the contract's two tasks per slice — a red test, then the code — whose titles
// carry facts only (CORE §2.4, §13).
package handoff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/scope"
)

// Slicing is the role's cut of a work item into one slice the engine runs: which criteria, where
// the code goes, where the tests live, how it is checked. The role writes it short; the engine
// derives the rest — names, task titles, writable sets (Apply).
type Slicing struct {
	ID         string   `json:"id"`             // the slice: "bk-001a" — its worktree, its branch (feature/<id>), its pins
	Item       string   `json:"item,omitempty"` // the backlog work item, "BK-001"; the command's argument may name it instead
	Acceptance []string `json:"acceptance"`     // the item's criteria this slice proves: "AC-2", "AC-5"
	Scope      []string `json:"scope"`          // path patterns the code task may change: "app/internal/store/*"
	Tests      []string `json:"tests"`          // the files the acceptance tests live in, where the language needs them
	Hooks      []string `json:"hooks,omitempty"`
	// TestCommand is the red task's gate: it must run and fail. GateCommand is the code task's: it
	// must pass — the slice's tests and what they must not break; the default is TestCommand.
	TestCommand []string `json:"test_command,omitempty"`
	GateCommand []string `json:"gate_command,omitempty"`
	Tools       []string `json:"tools,omitempty"`       // files the worker reads (contract task tools, F56)
	Refs        []string `json:"refs,omitempty"`        // item pack references the slice works on: "TA:§14"
	Constraints []string `json:"constraints,omitempty"` // facts every task of the slice keeps
	Later       string   `json:"later,omitempty"`       // what later slices bring, so this one leaves it out
}

var idRe = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`)

// ReadSlicing reads and checks a slicing file. An unknown key is an error: a misspelt field would
// silently drop part of the slice.
func ReadSlicing(file string) (*Slicing, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("slicing: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var s Slicing
	if err := dec.Decode(&s); err != nil {
		return nil, fmt.Errorf("slicing %s: %w", file, err)
	}
	if dec.More() {
		return nil, fmt.Errorf("slicing %s: more than one JSON value", file)
	}
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("slicing %s: %w", file, err)
	}
	return &s, nil
}

// Validate checks what the engine relies on: a name that makes a safe branch and directory name, the
// criteria, a scope and tests, and paths that stay inside the repository.
func (s *Slicing) Validate() error {
	if !idRe.MatchString(s.ID) {
		return fmt.Errorf("id %q: lower-case letters, digits and - . _ only, like bk-001a", s.ID)
	}
	if len(s.Acceptance) == 0 {
		return fmt.Errorf("acceptance names no criterion of the item")
	}
	if len(scope.Included(s.Scope)) == 0 {
		return fmt.Errorf("scope names no path the code task may change")
	}
	if len(s.Tests) == 0 {
		return fmt.Errorf("tests names no file — the slice's acceptance tests are the red task's output")
	}
	for _, list := range []struct {
		name  string
		paths []string
		files bool // a file the engine looks for, not a pattern
	}{{"scope", s.Scope, false}, {"tests", s.Tests, true}, {"hooks", s.Hooks, true}} {
		for _, p := range list.paths {
			p = strings.TrimPrefix(p, "!")
			switch {
			case strings.HasPrefix(p, "/") || strings.Contains(p, ":") || slices.Contains(strings.Split(p, "/"), ".."):
				return fmt.Errorf("%s %q: a path inside the repository, relative to its root", list.name, p)
			case list.files && strings.ContainsAny(p, "*?"):
				return fmt.Errorf("%s %q: a file, not a pattern", list.name, p)
			}
		}
	}
	return nil
}

// Name is the slice as people write it: the item's ID and the slice's suffix, "BK-001a" for the
// slice bk-001a of BK-001.
func (s *Slicing) Name(item string) string {
	if low := strings.ToLower(item); item != "" && strings.HasPrefix(s.ID, low) {
		return item + s.ID[len(low):]
	}
	return s.ID
}

// Branch is the slice's feature branch.
func (s *Slicing) Branch() string { return "feature/" + s.ID }

// WorktreeRel is the slice's worktree, relative to the dev repository.
func (s *Slicing) WorktreeRel() string { return ".worktrees/" + s.ID }

// PinRel is a spec root's pinned worktree, relative to that root's repository.
func (s *Slicing) PinRel() string { return ".worktrees/spec-" + s.ID }

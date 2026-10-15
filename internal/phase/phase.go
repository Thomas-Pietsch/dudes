// Package phase defines the fixed DUDES lifecycle registry (engine-owned graph).
package phase

import "fmt"

// ID is a stable phase identifier.
type ID string

const (
	Idea           ID = "idea"
	Strategy       ID = "strategy"
	Spec           ID = "spec"
	TestAuthoring  ID = "test_authoring"
	Implementation ID = "implementation"
	QA             ID = "qa"
	Documentation  ID = "documentation"
	HumanDecision  ID = "human_decision"
	Sealed         ID = "sealed"
)

// Def describes one phase in the registry (MVP: id + capability label).
type Def struct {
	ID         ID
	Capability string // human label; not an LLM persona
}

// DefaultOrder is the canonical forward lifecycle (CORE §4 / ARCHITECTURE §4).
var DefaultOrder = []Def{
	{Idea, "intake"},
	{Strategy, "strategy"},
	{Spec, "specification"},
	{TestAuthoring, "test_authoring"},
	{Implementation, "implementation"},
	{QA, "qa"},
	{Documentation, "documentation"},
	{HumanDecision, "human_decision"},
	{Sealed, "sealed"},
}

// Registry is an immutable ordered phase graph.
type Registry struct {
	order []Def
	index map[ID]int
}

// Default returns the built-in lifecycle registry.
func Default() *Registry {
	return New(DefaultOrder)
}

// New builds a registry from an ordered definition list.
func New(order []Def) *Registry {
	idx := make(map[ID]int, len(order))
	for i, d := range order {
		idx[d.ID] = i
	}
	cp := make([]Def, len(order))
	copy(cp, order)
	return &Registry{order: cp, index: idx}
}

// First returns the starting phase.
func (r *Registry) First() ID {
	return r.order[0].ID
}

// Contains reports whether id is in the registry.
func (r *Registry) Contains(id ID) bool {
	_, ok := r.index[id]
	return ok
}

// Next returns the successor of id, or false at the end / unknown id.
func (r *Registry) Next(id ID) (ID, bool) {
	i, ok := r.index[id]
	if !ok || i+1 >= len(r.order) {
		return "", false
	}
	return r.order[i+1].ID, true
}

// Order returns a copy of the phase list.
func (r *Registry) Order() []Def {
	out := make([]Def, len(r.order))
	copy(out, r.order)
	return out
}

// Parse validates a phase id string.
func (r *Registry) Parse(s string) (ID, error) {
	id := ID(s)
	if !r.Contains(id) {
		return "", fmt.Errorf("unknown phase %q", s)
	}
	return id, nil
}

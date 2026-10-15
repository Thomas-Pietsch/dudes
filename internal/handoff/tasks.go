package handoff

import (
	"fmt"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
)

// Apply writes the slice into c, the item's projection already narrowed to the slice's criteria
// (criteria: their IDs as the item writes them, "AC-2"): the contract's ID, its writable set, the
// files its acceptance tests live in, the constraints and two tasks —
//
//   - <id>-tests: red first. It writes the slice's tests and hooks and is done when its gate runs
//     and fails; a human reads the result (approval: after) before any code is written.
//   - <id>-code: the code. Its writable set is the scope minus those tests and hooks, so it cannot
//     weaken what it must make pass; its gate is the slice's.
//
// Titles carry facts only — which criteria, which files, which test functions, what stays read-only.
// How to solve is the worker's judgement (CORE §2.4). defaultTest is the engine's test_command, for a
// slicing that names none.
func (s *Slicing) Apply(c *contract.Contract, criteria, defaultTest []string) error {
	test := firstNonEmpty(s.TestCommand, defaultTest)
	if len(test) == 0 {
		return fmt.Errorf("slicing %s: no test_command — name one in the slicing or in dudes/config.json", s.ID)
	}
	gate := firstNonEmpty(s.GateCommand, test)
	if len(criteria) == 0 {
		return fmt.Errorf("slicing %s: no acceptance criterion", s.ID)
	}
	name := s.Name(c.Item)

	// The red task writes tests and hooks; the code task gets the scope minus them.
	written := union(s.Tests, s.Hooks)
	inScope := scope.Included(s.Scope)
	codeWritable := append([]string{}, s.Scope...)
	extra := []string{}
	for _, f := range written {
		if scope.Inside(f, inScope) {
			codeWritable = append(codeWritable, "!"+f)
		} else {
			extra = append(extra, f)
		}
	}

	var fns []string
	for _, ac := range criteria {
		fns = append(fns, spec.AcceptanceTestName(spec.AcceptanceID(c.Item, ac)))
	}

	c.ID = s.ID
	c.Tests = append([]string{}, s.Tests...)
	c.Writable = union(inScope, extra)
	c.Constraints = s.constraints()
	c.Tasks = []contract.Task{
		{
			ID: s.ID + "-tests", Phase: phase.Implementation,
			Title:    redTitle(name, c.Item, criteria, s.Tests, fns, s.Hooks),
			Approval: "after", Expect: "fail",
			Test: test, Writable: written, Tools: s.Tools, Refs: s.Refs,
		},
		{
			ID: s.ID + "-code", Phase: phase.Implementation,
			Title:    codeTitle(name, s.Tests, s.Hooks),
			Approval: "after",
			Test:     gate, Writable: codeWritable, Tools: s.Tools, Refs: s.Refs,
		},
	}
	return c.Validate()
}

// constraints are the facts every task of the slice keeps: where the slice stops, the role's own,
// and the standing one.
func (s *Slicing) constraints() []string {
	where := "In " + listAnd(code(scope.Included(s.Scope))) + " only"
	if later := strings.TrimRight(strings.TrimSpace(s.Later), ". "); later != "" {
		where += "; later slices: " + later
	}
	out := []string{where + "."}
	out = append(out, s.Constraints...)
	return append(out, "Do not rename instance documents.")
}

func redTitle(name, item string, criteria, tests, fns, hooks []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s, red first — the acceptance %s of %s %s, not the code: %s with the test %s %s.",
		name, plural(len(criteria), "test", "tests"), item, listAnd(criteria),
		listAnd(code(tests)), plural(len(fns), "function", "functions"), listAnd(code(fns)))
	if len(hooks) > 0 {
		fmt.Fprintf(&b, " It also writes the hook %s %s.", plural(len(hooks), "file", "files"), listAnd(code(hooks)))
	}
	b.WriteString(" The tests compile against today's package and fail until the code is written; the next task cannot change them.")
	return b.String()
}

func codeTitle(name string, tests, hooks []string) string {
	it := "It"
	if len(tests) > 1 {
		it = "They"
	}
	readOnly := append([]string{it}, code(hooks)...)
	verb := "are"
	if len(readOnly) == 1 {
		verb = "is"
	}
	return fmt.Sprintf("%s — the code: make %s pass. %s %s written and read-only.", name, listAnd(code(tests)), listAnd(readOnly), verb)
}

func firstNonEmpty(lists ...[]string) []string {
	for _, l := range lists {
		if len(l) > 0 {
			return append([]string{}, l...)
		}
	}
	return nil
}

// union is the lists joined in order, a repeated entry kept once.
func union(lists ...[]string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, l := range lists {
		for _, x := range l {
			if !seen[x] {
				seen[x] = true
				out = append(out, x)
			}
		}
	}
	return out
}

func code(xs []string) []string {
	out := make([]string, len(xs))
	for i, x := range xs {
		out[i] = "`" + x + "`"
	}
	return out
}

// listAnd joins "a", "a and b", "a, b and c".
func listAnd(xs []string) string {
	switch len(xs) {
	case 0:
		return ""
	case 1:
		return xs[0]
	}
	return strings.Join(xs[:len(xs)-1], ", ") + " and " + xs[len(xs)-1]
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

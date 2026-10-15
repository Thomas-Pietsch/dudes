package handoff_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
)

// x001a is the slicing a role writes for a real item (a small store package): short, the rest derived.
const x001a = `{
  "id": "x-001a",
  "item": "X-001",
  "acceptance": ["AC-1", "AC-2", "AC-3"],
  "scope": ["app/internal/store/*"],
  "tests": ["app/internal/store/x001_test.go"],
  "hooks": ["app/internal/store/hook.go"],
  "test_command": ["go", "test", "-count=1", "./app/internal/store/..."],
  "gate_command": ["go", "test", "-count=1", "./app/internal/store/...", "./app/api/..."],
  "tools": ["go.mod"],
  "refs": ["TA:§3"],
  "constraints": ["` + "`app/api`" + ` calls ` + "`Run(...)`" + `: these signatures stay."],
  "later": "API-02 and AC-4"
}`

func writeSlicing(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "slicing.json")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func mustSlicing(t *testing.T, body string) *handoff.Slicing {
	t.Helper()
	s, err := handoff.ReadSlicing(writeSlicing(t, body))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestReadSlicing(t *testing.T) {
	s := mustSlicing(t, x001a)
	if s.ID != "x-001a" || s.Item != "X-001" || len(s.Acceptance) != 3 || s.Branch() != "feature/x-001a" ||
		s.WorktreeRel() != ".worktrees/x-001a" || s.PinRel() != ".worktrees/spec-x-001a" {
		t.Fatalf("slicing %+v", s)
	}
	if got := s.Name("X-001"); got != "X-001a" {
		t.Errorf("name %q", got)
	}
	if got := s.Name("X-002"); got != "x-001a" {
		t.Errorf("another item's name %q", got)
	}
	const ok = `"acceptance":["AC-1"],"scope":["pkg/*"],"tests":["pkg/x_test.go"]`
	for name, c := range map[string]struct{ body, want string }{
		"unknown field":   {`{"id":"x",` + ok + `,"test_cmd":["go"]}`, "unknown field"},
		"two values":      {`{"id":"x",` + ok + `} {}`, "more than one"},
		"id upper case":   {`{"id":"X-001a",` + ok + `}`, "lower-case"},
		"id with a slash": {`{"id":"a/b",` + ok + `}`, "lower-case"},
		"no criterion":    {`{"id":"x","scope":["pkg/*"],"tests":["pkg/x_test.go"]}`, "acceptance"},
		"no scope":        {`{"id":"x","acceptance":["AC-1"],"tests":["pkg/x_test.go"]}`, "scope"},
		"only exclusions": {`{"id":"x","acceptance":["AC-1"],"scope":["!pkg/a.go"],"tests":["pkg/x_test.go"]}`, "scope"},
		"scope escapes":   {`{"id":"x","acceptance":["AC-1"],"scope":["../other/*"],"tests":["pkg/x_test.go"]}`, "inside the repository"},
		"no tests":        {`{"id":"x","acceptance":["AC-1"],"scope":["pkg/*"]}`, "tests"},
		"tests a pattern": {`{"id":"x","acceptance":["AC-1"],"scope":["pkg/*"],"tests":["pkg/*_test.go"]}`, "a file, not a pattern"},
		"tests absolute":  {`{"id":"x","acceptance":["AC-1"],"scope":["pkg/*"],"tests":["/pkg/x_test.go"]}`, "inside the repository"},
	} {
		_, err := handoff.ReadSlicing(writeSlicing(t, c.body))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want an error naming %q", name, err, c.want)
		}
	}
	if _, err := handoff.ReadSlicing(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("a missing file")
	}
}

// The two tasks of the first real item, as its role would have trimmed them by hand: facts only —
// criteria, files, test functions, what stays read-only — never how to solve.
func TestApplyTasks(t *testing.T) {
	s := mustSlicing(t, x001a)
	c := &contract.Contract{ID: "x-001", Product: "demo", Item: "X-001"}
	if err := s.Apply(c, []string{"AC-1", "AC-2", "AC-3"}, []string{"make", "test"}); err != nil {
		t.Fatal(err)
	}
	if c.ID != "x-001a" || len(c.Tasks) != 2 {
		t.Fatalf("contract %s with %d tasks", c.ID, len(c.Tasks))
	}
	red, code := c.Tasks[0], c.Tasks[1]
	wantRed := "X-001a, red first — the acceptance tests of X-001 AC-1, AC-2 and AC-3, not the code: `app/internal/store/x001_test.go` " +
		"with the test functions `TestX001AC1`, `TestX001AC2` and `TestX001AC3`. It also writes the hook file `app/internal/store/hook.go`. " +
		"The tests compile against today's package and fail until the code is written; the next task cannot change them."
	wantCode := "X-001a — the code: make `app/internal/store/x001_test.go` pass. It and `app/internal/store/hook.go` are written and read-only."
	if red.Title != wantRed {
		t.Errorf("red title:\n%s\nwant:\n%s", red.Title, wantRed)
	}
	if code.Title != wantCode {
		t.Errorf("code title:\n%s\nwant:\n%s", code.Title, wantCode)
	}
	if red.ID != "x-001a-tests" || red.Phase != phase.Implementation || red.Expect != "fail" || red.Approval != "after" ||
		strings.Join(red.Test, " ") != "go test -count=1 ./app/internal/store/..." {
		t.Errorf("red task %+v", red)
	}
	if code.ID != "x-001a-code" || code.Phase != phase.Implementation || code.Expect != "" || code.Approval != "after" ||
		strings.Join(code.Test, " ") != "go test -count=1 ./app/internal/store/... ./app/api/..." {
		t.Errorf("code task %+v", code)
	}
	if strings.Join(red.Writable, " ") != "app/internal/store/x001_test.go app/internal/store/hook.go" {
		t.Errorf("red writable %v", red.Writable)
	}
	if strings.Join(code.Writable, " ") != "app/internal/store/* !app/internal/store/x001_test.go !app/internal/store/hook.go" {
		t.Errorf("code writable %v", code.Writable)
	}
	for path, want := range map[string]bool{
		"app/internal/store/backup.go":    true,
		"app/internal/store/verify.go":    true,
		"app/internal/store/x001_test.go": false,
		"app/internal/store/hook.go":      false,
		"app/api/api.go":                  false,
	} {
		if scope.Inside(path, code.Writable) != want {
			t.Errorf("code task may change %s: %v, want %v", path, !want, want)
		}
	}
	for _, tk := range c.Tasks {
		if strings.Join(tk.Refs, " ") != "TA:§3" || strings.Join(tk.Tools, " ") != "go.mod" || tk.Worker != "" || len(tk.Skills) != 0 {
			t.Errorf("task %s: refs %v tools %v worker %q skills %v", tk.ID, tk.Refs, tk.Tools, tk.Worker, tk.Skills)
		}
	}
	wantCons := []string{
		"In `app/internal/store/*` only; later slices: API-02 and AC-4.",
		"`app/api` calls `Run(...)`: these signatures stay.",
		"Do not rename instance documents.",
	}
	if strings.Join(c.Constraints, "\n") != strings.Join(wantCons, "\n") {
		t.Errorf("constraints %q", c.Constraints)
	}
	if strings.Join(c.Tests, " ") != "app/internal/store/x001_test.go" || strings.Join(c.Writable, " ") != "app/internal/store/*" {
		t.Errorf("tests %v writable %v", c.Tests, c.Writable)
	}
}

func TestApplyVariants(t *testing.T) {
	apply := func(body string, criteria []string, defaultTest []string) (*contract.Contract, error) {
		s := mustSlicing(t, body)
		c := &contract.Contract{ID: "p-001", Product: "p", Item: "P-001"}
		return c, s.Apply(c, criteria, defaultTest)
	}

	// one criterion, no hook, the gate defaults to the test command, which defaults to the repository's
	c, err := apply(`{"id":"p-001","acceptance":["AC-1"],"scope":["pkg/*"],"tests":["pkg/x_test.go"]}`, []string{"AC-1"}, []string{"go", "test", "./..."})
	if err != nil {
		t.Fatal(err)
	}
	wantRed := "P-001, red first — the acceptance test of P-001 AC-1, not the code: `pkg/x_test.go` with the test function `TestP001AC1`. " +
		"The tests compile against today's package and fail until the code is written; the next task cannot change them."
	if c.Tasks[0].Title != wantRed {
		t.Errorf("red title:\n%s", c.Tasks[0].Title)
	}
	if want := "P-001 — the code: make `pkg/x_test.go` pass. It is written and read-only."; c.Tasks[1].Title != want {
		t.Errorf("code title:\n%s", c.Tasks[1].Title)
	}
	if strings.Join(c.Tasks[0].Test, " ") != "go test ./..." || strings.Join(c.Tasks[1].Test, " ") != "go test ./..." {
		t.Errorf("gates %v %v", c.Tasks[0].Test, c.Tasks[1].Test)
	}
	if want := "In `pkg/*` only."; c.Constraints[0] != want {
		t.Errorf("constraint %q", c.Constraints[0])
	}

	// two test files and two hooks
	c, err = apply(`{"id":"p-001","acceptance":["AC-1","AC-2"],"scope":["pkg/*"],"tests":["pkg/a_test.go","pkg/b_test.go"],"hooks":["pkg/h1.go","pkg/h2.go"],"test_command":["t"]}`, []string{"AC-1", "AC-2"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := "P-001 — the code: make `pkg/a_test.go` and `pkg/b_test.go` pass. They, `pkg/h1.go` and `pkg/h2.go` are written and read-only."; c.Tasks[1].Title != want {
		t.Errorf("code title:\n%s", c.Tasks[1].Title)
	}
	if want := "It also writes the hook files `pkg/h1.go` and `pkg/h2.go`."; !strings.Contains(c.Tasks[0].Title, want) {
		t.Errorf("red title:\n%s", c.Tasks[0].Title)
	}

	// a test outside the scope is not excluded from it — it is outside already — but the item may change it
	c, err = apply(`{"id":"p-001","acceptance":["AC-1"],"scope":["pkg/*"],"tests":["e2e/x_test.go"],"test_command":["t"]}`, []string{"AC-1"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(c.Tasks[1].Writable, " ") != "pkg/*" || strings.Join(c.Writable, " ") != "pkg/* e2e/x_test.go" || strings.Join(c.Tasks[0].Writable, " ") != "e2e/x_test.go" {
		t.Errorf("code writable %v, item writable %v, red writable %v", c.Tasks[1].Writable, c.Writable, c.Tasks[0].Writable)
	}

	// no gate named anywhere
	if _, err = apply(`{"id":"p-001","acceptance":["AC-1"],"scope":["pkg/*"],"tests":["pkg/x_test.go"]}`, []string{"AC-1"}, nil); err == nil || !strings.Contains(err.Error(), "test_command") {
		t.Errorf("no test command: %v", err)
	}
}

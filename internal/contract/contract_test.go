package contract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestSaveLoadValidate(t *testing.T) {
	dir := t.TempDir()
	c := &contract.Contract{
		ID:      "sample-r01",
		Product: "bookshelf",
		Variant: "standard",
		Sources: []contract.Source{{ID: "prd", Root: "pm", Path: "BOOKSHELF_PRD.md", Section: "§2"}},
		Acceptance: []contract.Acceptance{
			{ID: "greet", Kind: "behavior", Source: "prd", Text: "SHALL greet"},
			{ID: "field-1", Kind: "field", Text: "readers add a book in under ten seconds — not a CI gate"},
		},
		Tasks: []contract.Task{
			{ID: "ta-1", Phase: phase.TestAuthoring, Title: "author tests", Worker: "stub", Approval: "none"},
			{ID: "impl-1", Phase: phase.Implementation, Title: "implement", Worker: "aider", Approval: "none", Tools: []string{"edit"}},
			{ID: "go-nogo", Phase: phase.Strategy, Title: "scope call", Worker: "human", Approval: "before"},
		},
	}
	if err := contract.Save(dir, c); err != nil {
		t.Fatal(err)
	}
	got, ok, err := contract.Load(dir)
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if len(got.TasksFor(phase.Implementation)) != 1 {
		t.Fatal("impl tasks")
	}
	n, ok := got.NextTask(phase.TestAuthoring, "")
	if !ok || n.ID != "ta-1" {
		t.Fatalf("%+v", n)
	}
	_, ok = got.NextTask(phase.TestAuthoring, "ta-1")
	if ok {
		t.Fatal("expected no next")
	}
	tk, _ := got.TaskByID("go-nogo")
	if !tk.NeedsDecision() {
		t.Fatal("human task needs decision")
	}
}

// A contract can name the files its acceptance tests live in (F88); the existing ones count.
func TestNamedTests(t *testing.T) {
	dir := t.TempDir()
	c := &contract.Contract{ID: "s", Product: "p", Item: "X-001", Tests: []string{"pkg/x_test.go", "pkg/later_test.go", "pkg", "../outside_test.go"}}
	if !c.NamesTests() || (&contract.Contract{}).NamesTests() || (*contract.Contract)(nil).NamesTests() {
		t.Fatal("NamesTests")
	}
	if got := c.ExistingTests(dir); len(got) != 0 {
		t.Fatalf("nothing written yet: %v", got)
	}
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "x_test.go"), []byte("package pkg\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := c.ExistingTests(dir); len(got) != 1 || got[0] != "pkg/x_test.go" { // not the missing one, not the directory, not an escape
		t.Fatalf("existing tests %v", got)
	}
	if err := contract.Save(dir, c); err != nil {
		t.Fatal(err)
	}
	if back, ok, err := contract.Load(dir); err != nil || !ok || len(back.Tests) != 4 {
		t.Fatalf("the tests do not survive a round trip: %+v %v %v", back, ok, err)
	}
}

func TestRejectUnknownPhase(t *testing.T) {
	c := &contract.Contract{ID: "x", Product: "p", Tasks: []contract.Task{{ID: "t", Phase: "nope"}}}
	if err := c.Validate(); err == nil {
		t.Fatal("expected error")
	}
}

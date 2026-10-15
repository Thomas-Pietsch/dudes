package contract_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestProjectFromItemKeepsConstraintsAndTasks(t *testing.T) {
	prev := &contract.Contract{
		ID: "sample-r01", Product: "bookshelf", Variant: "demo",
		Sources:     []contract.Source{{ID: "prd", Path: "BOOKSHELF_PRD.md"}},
		Constraints: []string{"Do not rename instance documents."},
		Tasks:       []contract.Task{{ID: "implement", Phase: phase.Implementation, Worker: "aider"}},
	}
	src := []contract.Source{{ID: "backlog", Root: "pm", Path: "BOOKSHELF_BACKLOG.md", Section: "BK-002"}}
	acc := []contract.Acceptance{{ID: "bk-002-ac-1", Kind: "behavior", Source: "backlog", Text: "AC-1 GIVEN …"}}
	c := contract.Project("bookshelf", "BK-002", src, acc, prev)
	if c.ID != "bk-002" || c.Item != "BK-002" || c.Variant != "demo" {
		t.Fatalf("%+v", c)
	}
	if len(c.Sources) != 1 || c.Sources[0].Section != "BK-002" || len(c.Acceptance) != 1 {
		t.Fatal("sources/acceptance must come from the item, not the previous contract")
	}
	if len(c.Constraints) != 1 || len(c.Tasks) != 1 {
		t.Fatal("constraints and tasks are kept")
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if fresh := contract.Project("bookshelf", "BK-001", src, acc, nil); fresh.Tasks != nil || fresh.Variant != "" {
		t.Fatalf("%+v", fresh)
	}
}

package spec_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/spec"
)

const sample = `# PRD

### Requirement: greet
The CLI SHALL print a greeting.

#### Scenario: default run
- GIVEN the binary is built
- WHEN the user runs it
- THEN it prints the product name
`

func TestParseValidShape(t *testing.T) {
	d := spec.Parse(sample)
	if !d.ValidShape() {
		t.Fatalf("want valid, got %+v", d.Requirements)
	}
	if got := d.IDs(); len(got) != 1 || got[0] != "greet" {
		t.Fatalf("ids=%v", got)
	}
}

func TestParseREQID(t *testing.T) {
	d := spec.Parse("### Requirement: REQ-login User Authentication\nThe system SHALL authenticate.\n#### Scenario: ok\nGIVEN x\nWHEN y\nTHEN z\n")
	if len(d.Requirements) != 1 || d.Requirements[0].ID != "REQ-LOGIN" {
		t.Fatalf("got %+v", d.Requirements)
	}
	if !d.ValidShape() {
		t.Fatal("expected valid")
	}
}

func TestParseRejectsBulletList(t *testing.T) {
	d := spec.Parse("# PRD\n\n## Requirements\n\n- TBD (engine stub)\n")
	if d.ValidShape() {
		t.Fatal("bullets must not pass")
	}
	if err := spec.ShapeError(d); err == nil {
		t.Fatal("expected shape error")
	}
}

func TestParseIncompleteScenario(t *testing.T) {
	d := spec.Parse("### Requirement: x\nIt SHALL work.\n#### Scenario: a\n- GIVEN only\n")
	if d.ValidShape() {
		t.Fatal("incomplete scenario")
	}
}

func TestUncovered(t *testing.T) {
	d := spec.Parse(sample)
	if miss := spec.Uncovered(d, "func TestSmoke"); len(miss) != 1 || miss[0] != "greet" {
		t.Fatalf("miss=%v", miss)
	}
	if miss := spec.Uncovered(d, "// covers requirement greet\n"); len(miss) != 0 {
		t.Fatalf("miss=%v", miss)
	}
}

func TestGoTestsMentionsIDs(t *testing.T) {
	d := spec.Parse(sample)
	body := spec.GoTests("demo", d)
	if !strings.Contains(body, "func TestReq_greet") || !strings.Contains(body, "requirement greet") {
		t.Fatalf("body=%s", body)
	}
	if miss := spec.Uncovered(d, body); len(miss) != 0 {
		t.Fatalf("miss=%v", miss)
	}
}

func TestSlugFromLongHeading(t *testing.T) {
	d := spec.Parse("### Requirement: User Can Export CSV\nThe system SHALL export.\n#### Scenario: s\nGIVEN a\nWHEN b\nTHEN c\n")
	if !strings.HasPrefix(d.Requirements[0].ID, "user-can-export") {
		t.Fatalf("id=%q", d.Requirements[0].ID)
	}
}

func TestItemGoTestsFailUntilImplemented(t *testing.T) {
	if got := spec.AcceptanceTestName("bk-001-ac-1"); got != "TestBK001AC1" {
		t.Fatalf("name: %q", got)
	}
	acs := []spec.Criterion{
		{ID: "bk-001-ac-1", Text: "AC-1 a book is saved on the shelf"},
		{ID: "bk-001-ac-2", Text: "AC-2 the shelf opens within a second"},
	}
	if got := spec.ItemTestFile(acs[1].ID); got != "tests/bk001ac2_test.go" {
		t.Fatalf("file: %q", got)
	}
	src := spec.ItemGoTest("BK-001", acs[0])
	for _, want := range []string{"func TestBK001AC1(t *testing.T)", "t.Fatal(", "work item BK-001"} {
		if !strings.Contains(src, want) {
			t.Fatalf("missing %q in:\n%s", want, src)
		}
	}
	if strings.Contains(src, "TestBK001AC2") {
		t.Fatalf("one criterion per file:\n%s", src)
	}
	idx := spec.ItemTestIndex("BK-001", acs)
	for _, want := range []string{"package smoke_test", "tests/bk001ac1_test.go", "tests/bk001ac2_test.go"} {
		if !strings.Contains(idx, want) {
			t.Fatalf("missing %q in index:\n%s", want, idx)
		}
	}
}

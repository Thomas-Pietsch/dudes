// Package spec parses machine-checkable PRDs (F37).
//
// Shape (OpenSpec-inspired, CORE §6 "specification is the contract"):
//
//	### Requirement: <id>
//	The system SHALL …
//	#### Scenario: <name>
//	- GIVEN …
//	- WHEN …
//	- THEN …
package spec

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Requirement is one SHALL with scenario evidence.
type Requirement struct {
	ID        string
	Title     string
	HasSHALL  bool
	Scenarios []Scenario
}

// Scenario is a GIVEN/WHEN/THEN block under a requirement.
type Scenario struct {
	Name  string
	Given bool
	When  bool
	Then  bool
}

// Valid reports a complete scenario.
func (s Scenario) Valid() bool {
	return s.Given && s.When && s.Then
}

// Document is a parsed PRD.
type Document struct {
	Requirements []Requirement
}

// ValidShape is true when at least one requirement has SHALL and one valid scenario.
func (d *Document) ValidShape() bool {
	if d == nil {
		return false
	}
	for _, r := range d.Requirements {
		if r.HasSHALL && len(r.Scenarios) > 0 {
			for _, s := range r.Scenarios {
				if s.Valid() {
					return true
				}
			}
		}
	}
	return false
}

// IDs returns requirement identifiers in document order.
func (d *Document) IDs() []string {
	if d == nil {
		return nil
	}
	out := make([]string, 0, len(d.Requirements))
	for _, r := range d.Requirements {
		if r.ID != "" {
			out = append(out, r.ID)
		}
	}
	return out
}

var (
	reReqHead  = regexp.MustCompile(`(?i)^#{3}\s+Requirement\s*:?\s*(.*)$`)
	reScenHead = regexp.MustCompile(`(?i)^#{2,4}\s+Scenario\s*:?\s*(.*)$`)
	reReqID    = regexp.MustCompile(`(?i)^(REQ[-_][A-Za-z0-9]+)(?:\b[:\s]+(.*))?$`)
	reSHALL    = regexp.MustCompile(`(?i)\bSHALL\b`)
	reGiven    = regexp.MustCompile(`(?i)\bGIVEN\b`)
	reWhen     = regexp.MustCompile(`(?i)\bWHEN\b`)
	reThen     = regexp.MustCompile(`(?i)\bTHEN\b`)
	reHeading  = regexp.MustCompile(`^#{1,6}\s+`)
)

// Parse extracts requirements and scenarios from Markdown.
func Parse(md string) *Document {
	doc := &Document{}
	var cur *Requirement
	var scen *Scenario
	unnamed := 0

	flushScen := func() {
		if cur != nil && scen != nil {
			cur.Scenarios = append(cur.Scenarios, *scen)
			scen = nil
		}
	}
	flushReq := func() {
		flushScen()
		if cur != nil {
			doc.Requirements = append(doc.Requirements, *cur)
			cur = nil
		}
	}

	for _, raw := range strings.Split(md, "\n") {
		line := strings.TrimRight(raw, "\r")
		if m := reReqHead.FindStringSubmatch(line); m != nil {
			flushReq()
			id, title := parseReqHeading(m[1])
			if id == "" {
				unnamed++
				id = fmt.Sprintf("unnamed-%d", unnamed)
			}
			cur = &Requirement{ID: id, Title: title}
			continue
		}
		if cur == nil {
			continue
		}
		if m := reScenHead.FindStringSubmatch(line); m != nil {
			flushScen()
			name := strings.TrimSpace(m[1])
			if name == "" {
				name = "unnamed"
			}
			scen = &Scenario{Name: name}
			continue
		}
		if reHeading.MatchString(line) {
			// Other heading ends the current scenario, not the requirement
			// (unless it is another requirement, handled above).
			flushScen()
			continue
		}
		if scen != nil {
			if reGiven.MatchString(line) {
				scen.Given = true
			}
			if reWhen.MatchString(line) {
				scen.When = true
			}
			if reThen.MatchString(line) {
				scen.Then = true
			}
		} else if reSHALL.MatchString(line) {
			cur.HasSHALL = true
		}
	}
	flushReq()
	return doc
}

func parseReqHeading(rest string) (id, title string) {
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return "", ""
	}
	if m := reReqID.FindStringSubmatch(rest); m != nil {
		id = strings.ToUpper(strings.ReplaceAll(m[1], "_", "-"))
		title = strings.TrimSpace(m[2])
		if title == "" {
			title = id
		}
		return id, title
	}
	return slug(rest), rest
}

func slug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	prevHyphen := false
	n := 0
	for _, r := range s {
		if n >= 48 {
			break
		}
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevHyphen = false
			n++
		default:
			if !prevHyphen && b.Len() > 0 {
				b.WriteByte('-')
				prevHyphen = true
				n++
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// Uncovered returns requirement IDs that do not appear in haystack (tests, QR, …).
func Uncovered(d *Document, haystack string) []string {
	if d == nil {
		return nil
	}
	lower := strings.ToLower(haystack)
	var missing []string
	for _, id := range d.IDs() {
		if id == "" {
			continue
		}
		if !strings.Contains(lower, strings.ToLower(id)) {
			missing = append(missing, id)
		}
	}
	return missing
}

// Criterion is one acceptance criterion of a work item (contract acceptance, kind behavior).
type Criterion struct {
	ID   string // e.g. "bk-001-ac-1"
	Text string
}

// AcceptanceID is the contract ID of an item's criterion ("BK-001", "AC-2" → "bk-001-ac-2"); its
// test is AcceptanceTestName of it.
func AcceptanceID(item, ac string) string { return strings.ToLower(item + "-" + ac) }

// AcceptanceTestName names the test that proves criterion id: item and criterion in the name, as
// F77 asks of acceptance tests ("bk-001-ac-1" → "TestBK001AC1").
func AcceptanceTestName(id string) string {
	return "Test" + strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r - 'a' + 'A'
		case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			return r
		}
		return -1
	}, id)
}

// ItemTestFile is the file holding the acceptance test of criterion id — one file per criterion,
// so a task's scope names its own tests and cannot touch another task's.
func ItemTestFile(id string) string {
	return "tests/" + strings.ToLower(strings.TrimPrefix(AcceptanceTestName(id), "Test")) + "_test.go"
}

// ItemGoTest emits the failing skeleton of one acceptance criterion of a work item — tests first
// (TDD): it names its obligation and fails until implementation replaces it with the real
// acceptance test and makes it pass.
func ItemGoTest(item string, ac Criterion) string {
	var b strings.Builder
	b.WriteString("package smoke_test\n\nimport \"testing\"\n\n")
	fmt.Fprintf(&b, "// Generated by dudes test_authoring from work item %s. Replace the body with the real\n", item)
	b.WriteString("// acceptance test; keep the name (it carries the ID).\n\n")
	fmt.Fprintf(&b, "// %s\n", strings.TrimSpace(ac.Text))
	fmt.Fprintf(&b, "func %s(t *testing.T) {\n", AcceptanceTestName(ac.ID))
	fmt.Fprintf(&b, "\tt.Fatal(%q)\n", "not implemented yet: "+strings.TrimSpace(ac.Text))
	b.WriteString("}\n")
	return b.String()
}

// ItemTestIndex is the test phase's primary artifact for a work item: the package and the list of
// its acceptance test files (ItemTestFile).
func ItemTestIndex(item string, acs []Criterion) string {
	var b strings.Builder
	b.WriteString("package smoke_test\n\n")
	fmt.Fprintf(&b, "// Acceptance tests of work item %s, generated by dudes test_authoring — one file per criterion:\n", item)
	for _, ac := range acs {
		fmt.Fprintf(&b, "//   %s  %s\n", ItemTestFile(ac.ID), AcceptanceTestName(ac.ID))
	}
	return b.String()
}

// GoTests emits compilable smoke tests from parsed requirements (F32 real authoring, F67).
// Each TestReq_* mentions the requirement id so coverage is mechanical, not LLM prose.
func GoTests(product string, d *Document) string {
	if product == "" {
		product = "product"
	}
	var b strings.Builder
	b.WriteString("package smoke_test\n\nimport \"testing\"\n\n")
	b.WriteString("// Generated by dudes test_authoring from the frozen PRD.\n")
	fmt.Fprintf(&b, "// Product: %s\n\n", product)
	if d == nil || len(d.Requirements) == 0 {
		b.WriteString("func TestSmoke(t *testing.T) {\n\tt.Log(\"dudes: no ### Requirement: ids in PRD\")\n}\n")
		return b.String()
	}
	for _, r := range d.Requirements {
		ident := exportIdent(r.ID)
		fmt.Fprintf(&b, "// Requirement %s", r.ID)
		if r.Title != "" && r.Title != r.ID {
			fmt.Fprintf(&b, ": %s", r.Title)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "func TestReq_%s(t *testing.T) {\n", ident)
		fmt.Fprintf(&b, "\tt.Log(%q)\n", "requirement "+r.ID)
		b.WriteString("}\n\n")
	}
	return b.String()
}

func exportIdent(id string) string {
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	if s == "" {
		return "X"
	}
	if s[0] >= '0' && s[0] <= '9' {
		return "R_" + s
	}
	return s
}

// ShapeError explains why ValidShape is false.
func ShapeError(d *Document) error {
	if d == nil || len(d.Requirements) == 0 {
		return fmt.Errorf("spec-shape: no ### Requirement: headings (need SHALL + GIVEN/WHEN/THEN scenario)")
	}
	for _, r := range d.Requirements {
		if !r.HasSHALL {
			return fmt.Errorf("spec-shape: requirement %q has no SHALL", r.ID)
		}
		ok := false
		for _, s := range r.Scenarios {
			if s.Valid() {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("spec-shape: requirement %q needs a #### Scenario: with GIVEN, WHEN and THEN", r.ID)
		}
	}
	if d.ValidShape() {
		return nil
	}
	return fmt.Errorf("spec-shape: no valid requirement/scenario pair")
}

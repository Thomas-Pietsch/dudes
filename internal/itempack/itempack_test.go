package itempack

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// Conformance suite of context-pack/1 (CP-AC-01..17), ported from the kit.
// Fixtures: testdata/demo — the kit's DEMO spec set (BM listed, deliberately absent).

var demoKeys = map[string]string{
	"STRATEGY": "STRATEGY.md", "PS": "DEMO_PS.md", "BM": "DEMO_BM.md", "PRD": "DEMO_PRD.md",
	"TA": "DEMO_TA.md", "UX": "DEMO_UX.md", "ROADMAP": "DEMO_ROADMAP.md", "BACKLOG": "DEMO_BACKLOG.md",
	"SC": "DEMO_SC.md", "QR": "DEMO_QR.md",
}

var demoOrder = []string{"STRATEGY", "PS", "BM", "PRD", "TA", "UX", "ROADMAP", "BACKLOG", "SC", "QR"}

func loadDir(t *testing.T, dir string) *Set {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "demo", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := ParseProfile(data)
	if err != nil {
		t.Fatal(err)
	}
	var docs []Doc
	for _, k := range demoOrder {
		docs = append(docs, Doc{Key: k, Path: filepath.Join(dir, demoKeys[k])})
	}
	s, err := New(p, docs)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func demo(t *testing.T) *Set { return loadDir(t, filepath.Join("testdata", "demo")) }

// mutated copies the fixtures and applies one replacement per document key.
func mutated(t *testing.T, edits map[string][2]string) *Set {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join("testdata", "demo")
	entries, _ := os.ReadDir(src)
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(src, e.Name()))
		_ = os.WriteFile(filepath.Join(dir, e.Name()), data, 0o644)
	}
	for key, ed := range edits {
		p := filepath.Join(dir, demoKeys[key])
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), ed[0]) {
			t.Fatalf("%s: %q not in fixture", key, ed[0])
		}
		_ = os.WriteFile(p, []byte(strings.ReplaceAll(string(data), ed[0], ed[1])), 0o644) // Python str.replace: all
	}
	return loadDir(t, dir)
}

func resolveOne(t *testing.T, s *Set, doc, target string) []string {
	t.Helper()
	parts, err := s.Resolve(doc, target)
	if err != nil {
		t.Fatal(err)
	}
	return parts
}

func mustContain(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("missing %q in:\n%s", sub, s)
	}
}

func codes(s *Set) map[string]bool {
	out := map[string]bool{}
	for _, f := range s.Check() {
		out[f.Code] = true
	}
	return out
}

// --- Resolution (CP-FR-02/03) ---

func TestTableRowWithHeader(t *testing.T) { // CP-AC-01
	text := resolveOne(t, demo(t), "PRD", "AC-NOTE-01")[0]
	if !strings.HasPrefix(text, "| ID | Criterion |") {
		t.Fatal(text)
	}
	mustContain(t, text, "AC-NOTE-01 | GIVEN a note")
	if strings.Contains(text, "AC-NOTE-02") {
		t.Fatal("neighbouring row leaked")
	}
}

func TestRangeMergesRowsUnderOneHeader(t *testing.T) { // CP-AC-02
	parts := resolveOne(t, demo(t), "PRD", "AC-NOTE-01..02")
	if len(parts) != 1 || strings.Count(parts[0], "| ID | Criterion |") != 1 {
		t.Fatalf("%d parts: %q", len(parts), parts)
	}
	mustContain(t, parts[0], "AC-NOTE-02")
}

func TestHeadingSectionUntilNextEqualHeading(t *testing.T) { // CP-AC-03
	mustContain(t, resolveOne(t, demo(t), "PRD", "TC-P0-01")[0], "THEN the author is shown")
}

func TestSectionReference(t *testing.T) { // CP-AC-03
	text := resolveOne(t, demo(t), "PRD", "§6")[0]
	if !strings.HasPrefix(text, "## 6. Tests") || strings.Contains(text, "## 2.") {
		t.Fatal(text)
	}
}

func TestBoldListItemAndSubruleRange(t *testing.T) { // CP-AC-04
	parts := resolveOne(t, demo(t), "TA", "ALG-01.1..2")
	if len(parts) != 2 {
		t.Fatalf("%q", parts)
	}
	mustContain(t, parts[1], "ALG-01.2")
}

func TestBoldParagraphBlock(t *testing.T) { // CP-AC-04
	mustContain(t, resolveOne(t, demo(t), "PRD", "AC-PRIV-01")[0], "contains no count per person")
}

func TestTableIncludesDDLAndTriggers(t *testing.T) { // CP-AC-05
	text := strings.Join(resolveOne(t, demo(t), "TA", "TBL-01"), "\n")
	mustContain(t, text, "CREATE TABLE notes")
	mustContain(t, text, "trg_notes_author")
}

func TestAPIIncludesIndexRowAndSchema(t *testing.T) { // CP-AC-06
	text := strings.Join(resolveOne(t, demo(t), "TA", "API-01"), "\n")
	mustContain(t, text, "| API-01 | `POST`")
	mustContain(t, text, "ERR_EMPTY")
}

func TestCodeFenceIsNotStructure(t *testing.T) { // CP-AC-07
	if _, err := demo(t).Resolve("TA", "§not"); err == nil {
		t.Fatal("heading inside a code fence resolved")
	}
}

func TestUnresolvedReferenceFailsLoudly(t *testing.T) { // CP-AC-08
	s := demo(t)
	var re *ResolveError
	if _, err := s.Resolve("TA", "API-99"); !errors.As(err, &re) {
		t.Fatalf("got %v", err)
	}
	if _, err := s.Resolve("XX", "API-01"); !errors.As(err, &re) {
		t.Fatalf("got %v", err)
	}
}

// --- Ownership (CP-AC-16/17) ---

func TestSCThreatAndSecurityTestResolveInSC(t *testing.T) {
	s := demo(t)
	mustContain(t, resolveOne(t, s, "SC", "T-DM-01")[0], "Author of a note rewritten later")
	mustContain(t, resolveOne(t, s, "SC", "TC-SEC-ATTR-01")[0], "aborts")
}

func TestPSSectionResolves(t *testing.T) {
	mustContain(t, resolveOne(t, demo(t), "PS", "§14.7")[0], "zertifiziert")
}

func TestNarrowOwnerExceptionWins(t *testing.T) {
	s := demo(t)
	for id, want := range map[string]string{"TC-SEC-01": "PRD", "TC-SEC-13": "SC", "TC-SEC-ATTR-01": "SC"} {
		if got := s.OwnerOf(id); got != want {
			t.Fatalf("%s owner=%s want %s", id, got, want)
		}
	}
}

func TestListedButAbsentDocumentIsToleratedUntilReferenced(t *testing.T) {
	s := demo(t)
	if s.Has("BM") {
		t.Fatal("BM fixture must be absent")
	}
	if f := s.Check(); len(f) != 0 {
		t.Fatalf("clean fixture: %v", f)
	}
	m := mutated(t, map[string][2]string{"BACKLOG": {"[[PS:§14.7]] [[PRD:AC-NOTE-01..02]]", "[[BM:§2]] [[PRD:AC-NOTE-01..02]]"}})
	hit := false
	for _, f := range m.Check() {
		if f.Code == "C1" && strings.Contains(f.Msg, "BM") {
			hit = true
		}
	}
	if !hit {
		t.Fatal("reference into absent BM not reported as C1")
	}
}

// --- Packs (CP-FR-04/05) ---

func TestPackHasFiveFilesAndManifestContract(t *testing.T) { // CP-AC-09
	s := demo(t)
	p, err := s.Build("DM-001", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Files) != 4 {
		t.Fatalf("files %v", p.Files)
	}
	raw, _ := json.Marshal(p.Manifest)
	var got map[string]any
	_ = json.Unmarshal(raw, &got)
	var schema struct {
		Required []string `json:"required"`
	}
	data, _ := os.ReadFile(filepath.Join("testdata", "manifest.schema.json"))
	_ = json.Unmarshal(data, &schema)
	for _, k := range schema.Required {
		if _, ok := got[k]; !ok {
			t.Fatalf("manifest missing %s", k)
		}
	}
	m := p.Manifest
	if m.Release == nil || *m.Release != "R0.1" {
		t.Fatalf("release %v", m.Release)
	}
	want := []string{"TC-P0-01", "AC-NOTE-01", "AC-NOTE-02", "TC-SEC-01", "TC-SEC-ATTR-01", "TC-SEC-13"}
	if strings.Join(m.DoD, ",") != strings.Join(want, ",") {
		t.Fatalf("dod %v", m.DoD)
	}
	if len(m.Sources) != len(s.keys) {
		t.Fatalf("sources %d, loaded %d", len(m.Sources), len(s.keys))
	}
}

func TestTaskIsVerbatimAndContextIsGrouped(t *testing.T) { // CP-AC-10
	p, err := demo(t).Build("DM-001", 0)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, p.Files["TASK.md"], "#### DM-001 · Save a note with author")
	ctx := p.Files["CONTEXT.md"]
	var pos []int
	for _, d := range []string{"PS", "PRD", "TA", "UX", "SC"} {
		pos = append(pos, strings.Index(ctx, "## "+d+"\n"))
	}
	if !sort.IntsAreSorted(pos) || pos[0] < 0 {
		t.Fatalf("group order %v", pos)
	}
}

func TestRulesComeFromConfiguredSections(t *testing.T) { // CP-AC-11
	p, _ := demo(t).Build("DM-001", 0)
	for _, h := range []string{"#### H.6 Backlog method", "## 0. How to work this backlog", "## 6. Global Definition of Done"} {
		mustContain(t, p.Files["RULES.md"], h)
	}
}

func TestBudgetTrimsLargestFirstAndFlagsOverrun(t *testing.T) { // CP-AC-12
	p, err := demo(t).Build("DM-001", 500)
	if err != nil {
		t.Fatal(err)
	}
	trimmed := false
	for _, r := range p.Manifest.References {
		trimmed = trimmed || r.Trimmed
	}
	if !trimmed || !p.Manifest.OverBudget {
		t.Fatalf("trimmed=%v over=%v", trimmed, p.Manifest.OverBudget)
	}
	mustContain(t, p.Files["CONTEXT.md"], "Shortened to fit the budget")
	mustContain(t, p.Files["TASK.md"], "**DoD.**")
}

func TestOutlineItemsAreNotPackable(t *testing.T) { // CP-AC-13
	if _, err := demo(t).Build("DM-002", 0); err == nil {
		t.Fatal("outline row packed")
	}
}

func TestWritePackCreatesFolder(t *testing.T) { // CP-AC-09
	p, _ := demo(t).Build("DM-001", 0)
	ws := t.TempDir()
	rel, err := Write(ws, p)
	if err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(ws, filepath.FromSlash(rel)))
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if strings.Join(names, ",") != "CONTEXT.md,PACK.md,RULES.md,TASK.md,manifest.json" {
		t.Fatalf("files %v", names)
	}
}

// --- Status (CP-AC-14) ---

func statusOf(t *testing.T, s *Set) (string, []string) {
	t.Helper()
	it, ok := s.FindItem("DM-001")
	if !ok {
		t.Fatal("DM-001 missing")
	}
	st, _, missing := s.DeriveStatus(it, nil)
	return st, missing
}

func TestStatusOpen(t *testing.T) {
	if st, _ := statusOf(t, demo(t)); st != StatusOpen {
		t.Fatal(st)
	}
}

func TestStatusDone(t *testing.T) {
	s := mutated(t, map[string][2]string{"QR": {"| W1 | ⬜ |", "| W1 | 🟢 |"}})
	if st, _ := statusOf(t, s); st != StatusDone {
		t.Fatal(st)
	}
}

func TestStatusBlocked(t *testing.T) {
	s := mutated(t, map[string][2]string{"QR": {"| AC-NOTE-02 | Previous text readable | W1 | ⬜ |", "| AC-NOTE-02 | Previous text readable | W1 | 🔴 |"}})
	if st, _ := statusOf(t, s); st != StatusBlocked {
		t.Fatal(st)
	}
}

func TestStatusUntracked(t *testing.T) {
	s := mutated(t, map[string][2]string{"QR": {"| AC-NOTE-02 | Previous text readable | W1 | ⬜ |\n", ""}})
	st, missing := statusOf(t, s)
	if st != StatusUntracked || strings.Join(missing, ",") != "AC-NOTE-02" {
		t.Fatal(st, missing)
	}
}

// --- Not applicable ("—", only for the first cell's IDs, never over a real mark) ---

const acNote02Row = "| AC-NOTE-02 | Previous text readable | W1 | ⬜ |"

func withNA(s *Set) *Set {
	s.Profile.Status.NotApplicable = []string{"—"}
	return s
}

func TestNotApplicableInFirstCellIsTracked(t *testing.T) {
	s := withNA(mutated(t, map[string][2]string{"QR": {acNote02Row, "| AC-NOTE-02 | Not this item | — |"}}))
	if st, missing := statusOf(t, s); st == StatusUntracked || len(missing) != 0 {
		t.Fatal(st, missing)
	}
	if f := s.Check(); len(f) != 0 {
		t.Fatalf("C8 must accept a not-applicable row: %v", f)
	}
}

func TestNotApplicableNeedsTheProfileMark(t *testing.T) {
	s := mutated(t, map[string][2]string{"QR": {acNote02Row, "| AC-NOTE-02 | Not this item | — |"}})
	if st, _ := statusOf(t, s); st != StatusUntracked {
		t.Fatal(st)
	}
}

func TestNotApplicableOnlyForTheFirstCell(t *testing.T) {
	s := withNA(mutated(t, map[string][2]string{"QR": {acNote02Row, "| Weights | covers AC-NOTE-02 | — |"}}))
	if st, missing := statusOf(t, s); st != StatusUntracked || strings.Join(missing, ",") != "AC-NOTE-02" {
		t.Fatal("a mention outside the first cell must not hide a missing row:", st, missing)
	}
}

func TestNotApplicableNeverOverridesARealMark(t *testing.T) {
	s := withNA(mutated(t, map[string][2]string{"QR": {acNote02Row,
		"| AC-NOTE-02 | Previous text readable | W1 | 🔴 |\n| AC-NOTE-02 | Not this item | — |"}}))
	if st, _ := statusOf(t, s); st != StatusBlocked {
		t.Fatal(st)
	}
}

// --- Consistency (CP-AC-15) ---

func TestCleanFixtureHasNoFindings(t *testing.T) {
	if f := demo(t).Check(); len(f) != 0 {
		t.Fatalf("%v", f)
	}
}

func TestC1Unresolved(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"BACKLOG": {"[[TA:SLO-01]]", "[[TA:SLO-09]]"}}))["C1"] {
		t.Fatal("C1 not reported")
	}
}

func TestC2DoubleDefinition(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"UX": {"| W-MC-03 | Autorzeile |", "| W-MC-02 | Doppelt |"}}))["C2"] {
		t.Fatal("C2 not reported")
	}
}

func TestC3RouteShapeAndUIStringInBacklog(t *testing.T) {
	s := mutated(t, map[string][2]string{"BACKLOG": {
		"WHEN [[TA:API-01]] is called, THEN the card shows „Notiz von {Name}“ ([[UX:W-MC-03]])",
		"WHEN `POST /notes` is called, THEN the card shows „Notiz von {Name}“"}})
	n := 0
	for _, f := range s.Check() {
		if f.Code == "C3" {
			n++
		}
	}
	if n != 2 {
		t.Fatalf("C3 findings %d, want 2", n)
	}
}

func TestC4RouteTableInPRD(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"PRD": {"## 6. Tests", "| POST | `/api/v1/demo/notes` | x |\n\n## 6. Tests"}}))["C4"] {
		t.Fatal("C4 not reported")
	}
}

func TestC5RouteNotOwned(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"PRD": {"## 6. Tests", "See /api/v1/demo/unknown for details.\n\n## 6. Tests"}}))["C5"] {
		t.Fatal("C5 not reported")
	}
}

func TestC6ReleaseNotInRoadmap(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"BACKLOG": {"**Release.** R0.1 ·", "**Release.** R0.9 ·"}}))["C6"] {
		t.Fatal("C6 not reported")
	}
}

func TestC7MilestoneWithoutItems(t *testing.T) {
	row := "| 2026-11-01 | R0.2 \"Search\" | exit set R0.2 |"
	if !codes(mutated(t, map[string][2]string{"ROADMAP": {row, row + "\n| 2027-01-01 | R0.3 \"Empty\" | x |"}}))["C7"] {
		t.Fatal("C7 not reported")
	}
}

func TestC2SCIDDefinedTwice(t *testing.T) {
	s := mutated(t, map[string][2]string{"SC": {"| TC-SEC-13 |", "| TC-SEC-ATTR-01 |"}})
	for _, f := range s.Check() {
		if f.Code == "C2" && strings.HasPrefix(f.Msg, "TC-SEC-ATTR-01") {
			return
		}
	}
	t.Fatal("double SC definition not reported")
}

func TestC8DoDWithoutQRRow(t *testing.T) {
	if !codes(mutated(t, map[string][2]string{"QR": {"| TC-P0-01 | Note saved with author | W1 | ⬜ |\n", ""}}))["C8"] {
		t.Fatal("C8 not reported")
	}
}

// --- Byte parity with the Python reference (testdata/golden, generated by the kit) ---

func TestGoldenParityWithReference(t *testing.T) {
	now = func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
	defer func() { now = time.Now }()
	for name, budget := range map[string]int{"DM-001": 0, "DM-001-b500": 500} {
		p, err := demo(t).Build("DM-001", budget)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range FileNames {
			want, err := os.ReadFile(filepath.Join("testdata", "golden", name, f))
			if err != nil {
				t.Fatal(err)
			}
			got := p.Files[f]
			if !strings.HasSuffix(got, "\n") {
				got += "\n"
			}
			// The kit writes CRLF on Windows (Python text mode); the pack text is what counts.
			if a, b := dropStamp(got), dropStamp(strings.ReplaceAll(string(want), "\r\n", "\n")); a != b {
				t.Fatalf("%s/%s differs from reference\n--- go\n%s\n--- python\n%s", name, f, a, b)
			}
		}
		var ref Manifest
		data, _ := os.ReadFile(filepath.Join("testdata", "golden", name, "manifest.json"))
		if err := json.Unmarshal(data, &ref); err != nil {
			t.Fatal(err)
		}
		g := p.Manifest
		if g.Status != ref.Status || g.OverBudget != ref.OverBudget || strings.Join(g.DoD, ",") != strings.Join(ref.DoD, ",") {
			t.Fatalf("%s manifest status/dod/over differ: %+v vs %+v", name, g, ref)
		}
		for k, v := range ref.Sizes {
			want, _ := os.ReadFile(filepath.Join("testdata", "golden", name, k))
			// Sizes count the stamp line; only the generator name in it differs.
			if g.Sizes[k]-stampChars(p.Files[k]) != v-stampChars(string(want)) {
				t.Fatalf("%s size %s: go %d python %d", name, k, g.Sizes[k], v)
			}
		}
		if len(g.References) != len(ref.References) {
			t.Fatalf("%s refs %d vs %d", name, len(g.References), len(ref.References))
		}
		for i := range ref.References {
			if g.References[i] != ref.References[i] {
				t.Fatalf("%s ref %d: %+v vs %+v", name, i, g.References[i], ref.References[i])
			}
		}
	}
}

func stampChars(s string) int {
	n := 0
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(l, "> Generated ") {
			n += runeLen(strings.TrimRight(l, "\r"))
		}
	}
	return n
}

// dropStamp removes the generated line, the only intended difference (CP-NFR-05).
func dropStamp(s string) string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if !strings.HasPrefix(l, "> Generated ") {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func TestExpandRange(t *testing.T) {
	for in, want := range map[string]string{
		"X-01..05":    "X-01,X-02,X-03,X-04,X-05",
		"X-01..X-03":  "X-01,X-02,X-03",
		"ALG-01.1..3": "ALG-01.1,ALG-01.2,ALG-01.3",
		"API-01":      "API-01",
	} {
		if got := strings.Join(ExpandRange(in), ","); got != want {
			t.Fatalf("%s → %s want %s", in, got, want)
		}
	}
}

func TestTokenBoundaries(t *testing.T) {
	// ALG-03 never matches ALG-03.2; a dash or word rune before an ID breaks it.
	if hasToken("see ALG-03.2 here", "ALG-03") {
		t.Fatal("ALG-03 matched ALG-03.2")
	}
	if !hasToken("ALG-03. end", "ALG-03") || !hasToken("(ALG-03)", "ALG-03") {
		t.Fatal("bounded ALG-03 missed")
	}
	if hasToken("XALG-03", "ALG-03") || hasToken("Ä-ALG-03", "ALG-03") {
		t.Fatal("unbounded prefix matched")
	}
}

func TestAcceptanceLines(t *testing.T) {
	s := demo(t)
	it, _ := s.FindItem("DM-001")
	acs := s.Acceptance(it)
	if len(acs) != 1 || acs[0].ID != "AC-1" || !strings.HasPrefix(acs[0].Text, "GIVEN a note") {
		t.Fatalf("%+v", acs)
	}
	bold := Item{Text: "- **AC-2** GIVEN x, WHEN y, THEN z."}
	if got := s.Acceptance(bold); len(got) != 1 || got[0].ID != "AC-2" {
		t.Fatalf("%+v", got)
	}
}

func TestTrailersNameItemAndEverySource(t *testing.T) {
	s := demo(t)
	p, err := s.Build("DM-001", 0)
	if err != nil {
		t.Fatal(err)
	}
	tr := p.Trailers()
	if !strings.HasPrefix(tr, "Item: DM-001\nSpec-Source: ") {
		t.Fatal(tr)
	}
	if got := strings.Count(tr, "Spec-Source: "); got != len(s.keys) {
		t.Fatalf("%d source trailers, %d loaded docs", got, len(s.keys))
	}
}

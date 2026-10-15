// Package gate implements mechanical lifecycle checks (F30–F32, F37–F39).
package gate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/evidence"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
	ws "github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Axis is one review question (F38). Mixing them lets one mask the other.
type Axis string

const (
	AxisSpec      Axis = "spec"
	AxisStandards Axis = "standards"
)

// CheckEnter validates prerequisites before a phase worker runs.
func CheckEnter(workspace, product string, id phase.ID) error {
	switch id {
	case phase.Implementation:
		if err := RequireSpec(workspace, product); err != nil {
			return err
		}
		if err := freeze.RequireValid(workspace, product); err != nil {
			return err
		}
		if err := RequireTests(workspace, product); err != nil {
			return err
		}
	case phase.QA:
		if err := RequireSpec(workspace, product); err != nil {
			return fmt.Errorf("qa requires spec: %w", err)
		}
	}
	return nil
}

// CheckLeave validates that the phase produced its required durable output
// before the orchestrator advances.
func CheckLeave(workspace, product string, id phase.ID) error {
	switch id {
	case phase.HumanDecision, phase.Sealed, phase.Implementation:
		return nil
	case phase.TestAuthoring:
		if err := RequireTests(workspace, product); err != nil {
			return err
		}
		return RequireRequirementCoverage(workspace, product)
	case phase.Spec:
		if err := requirePrimaryNonEmpty(workspace, product, id); err != nil {
			return err
		}
		return RequireSpecShape(workspace, product)
	}
	return requirePrimaryNonEmpty(workspace, product, id)
}

func requirePrimaryNonEmpty(workspace, product string, id phase.ID) error {
	rel, ok := artifact.Path(product, id)
	if !ok {
		return nil
	}
	okNonEmpty, err := artifact.NonEmpty(workspace, rel)
	if err != nil {
		return err
	}
	if !okNonEmpty {
		return fmt.Errorf("artifact required after %s: %s missing or empty", id, rel)
	}
	return nil
}

// RequireSpec enforces F31 — non-empty PRD before impl (and related gates). An item-bound
// contract is the spec in a dev repo: the spec repo owns the PRD, the worker gets the item pack.
func RequireSpec(workspace, product string) error {
	if ItemBound(workspace) {
		return nil
	}
	rel, ok := artifact.Path(product, phase.Spec)
	if !ok {
		return fmt.Errorf("spec-before-impl: no PRD path registered")
	}
	okNonEmpty, err := artifact.NonEmpty(workspace, rel)
	if err != nil {
		return err
	}
	if !okNonEmpty {
		return fmt.Errorf("spec-before-impl: %s missing or empty — run spec phase first", rel)
	}
	return nil
}

// RequireSpecShape enforces F37 — PRD is a machine-checkable contract.
func RequireSpecShape(workspace, product string) error {
	rel, ok := artifact.Path(product, phase.Spec)
	if !ok {
		return fmt.Errorf("spec-shape: no PRD path registered")
	}
	data, err := artifact.Read(workspace, rel)
	if err != nil {
		return fmt.Errorf("spec-shape: %w", err)
	}
	return spec.ShapeError(spec.Parse(string(data)))
}

// RequireRequirementCoverage enforces F37 — each requirement id appears in tests. With an
// item-bound contract the item's behavior criteria are the behavioral shape: each has its test.
func RequireRequirementCoverage(workspace, product string) error {
	if con, ok, err := contract.Load(workspace); err != nil {
		return fmt.Errorf("requirement-coverage: %w", err)
	} else if ok && con.Item != "" {
		return requireItemCoverage(workspace, product, con)
	}
	rel, ok := artifact.Path(product, phase.Spec)
	if !ok {
		return nil
	}
	data, err := artifact.Read(workspace, rel)
	if err != nil {
		return fmt.Errorf("requirement-coverage: %w", err)
	}
	doc := spec.Parse(string(data))
	if len(doc.IDs()) == 0 {
		return fmt.Errorf("requirement-coverage: PRD has no ### Requirement: ids")
	}
	hay, err := collectTestHaystack(workspace, product)
	if err != nil {
		return err
	}
	missing := spec.Uncovered(doc, hay)
	if len(missing) > 0 {
		return fmt.Errorf("requirement-coverage: tests do not mention %s", strings.Join(missing, ", "))
	}
	return nil
}

// ItemBound reports whether .dudes/contract.json binds a backlog item (F71).
func ItemBound(workspace string) bool {
	con, ok, err := contract.Load(workspace)
	return err == nil && ok && con.Item != ""
}

// NamesTests reports whether .dudes/contract.json names the files of its acceptance tests (F88):
// they sit where the language needs them, and the contract's red-first task writes them.
func NamesTests(workspace string) bool {
	con, ok, err := contract.Load(workspace)
	return err == nil && ok && con.NamesTests()
}

func requireItemCoverage(workspace, product string, con *contract.Contract) error {
	hay, err := collectTestHaystack(workspace, product)
	if err != nil {
		return err
	}
	// the files the contract names count wherever they are and whatever they are called (F88)
	for _, rel := range con.ExistingTests(workspace) {
		if data, err := artifact.Read(workspace, rel); err == nil {
			hay += string(data) + "\n"
		}
	}
	var missing []string
	for _, a := range con.Acceptance {
		if strings.EqualFold(a.Kind, "behavior") && !strings.Contains(hay, spec.AcceptanceTestName(a.ID)) {
			missing = append(missing, a.ID+" ("+spec.AcceptanceTestName(a.ID)+")")
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("requirement-coverage: item %s criteria without a test: %s", con.Item, strings.Join(missing, ", "))
	}
	return nil
}

func collectTestHaystack(workspaceRoot, product string) (string, error) {
	var b strings.Builder
	for _, rel := range testCandidates(product) {
		if !artifact.Exists(workspaceRoot, rel) {
			continue
		}
		data, err := artifact.Read(workspaceRoot, rel)
		if err != nil {
			return "", err
		}
		b.Write(data)
		b.WriteByte('\n')
	}
	_ = filepath.Walk(workspaceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			if info != nil && info.IsDir() && (info.Name() == ".dudes" || ws.Foreign(workspaceRoot, path)) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") || strings.HasSuffix(info.Name(), "_test.py") || strings.HasSuffix(info.Name(), ".test.ts") {
			data, err := os.ReadFile(path)
			if err == nil {
				b.Write(data)
				b.WriteByte('\n')
			}
		}
		return nil
	})
	return b.String(), nil
}

func testCandidates(product string) []string {
	if product == "" {
		product = "product"
	}
	out := []string{
		filepath.ToSlash(filepath.Join("tests", product+"_smoke_test.go")),
		"tests/smoke_test.go",
	}
	if rel, ok := artifact.Path(product, phase.TestAuthoring); ok {
		out = append([]string{rel}, out...)
	}
	return out
}

// RequireTests enforces F32 — test authoring left runnable evidence. A contract that names its
// tests writes them in its first task, red: that task's gate proves they run and fail, so the
// gate has nothing to find before it (F88).
func RequireTests(workspace, product string) error {
	if NamesTests(workspace) {
		return nil
	}
	if product == "" {
		product = "product"
	}
	for _, rel := range testCandidates(product) {
		ok, err := artifact.NonEmpty(workspace, rel)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
	}
	found := false
	_ = filepath.Walk(workspace, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil || info.IsDir() {
			if info != nil && info.IsDir() && (info.Name() == ".dudes" || ws.Foreign(workspace, path)) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") && info.Size() > 0 {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if found {
		return nil
	}
	return fmt.Errorf("test-before-impl: no non-empty test file (expected tests/%s_smoke_test.go)", product)
}

// SpecPath returns the authoritative PRD relative path (F12).
func SpecPath(product string) (string, bool) {
	return artifact.Path(product, phase.Spec)
}

// Finding is one mechanical convergence issue (F34 / F38).
type Finding struct {
	Axis  Axis     `json:"axis"`
	Phase phase.ID `json:"phase"`
	Path  string   `json:"path,omitempty"`
	Msg   string   `json:"msg"`
}

func (f Finding) Error() string {
	axis := string(f.Axis)
	if axis == "" {
		axis = "standards"
	}
	if f.Path != "" {
		return fmt.Sprintf("%s/%s: %s (%s)", axis, f.Phase, f.Msg, f.Path)
	}
	return fmt.Sprintf("%s/%s: %s", axis, f.Phase, f.Msg)
}

// Report is a two-axis converge result (F38).
type Report struct {
	Spec      []Finding `json:"spec"`
	Standards []Finding `json:"standards"`
	Tasks     string    `json:"tasks,omitempty"`
}

// OK is true when both axes are clean.
func (r Report) OK() bool {
	return len(r.Spec) == 0 && len(r.Standards) == 0
}

// All concatenates spec then standards.
func (r Report) All() []Finding {
	out := make([]Finding, 0, len(r.Spec)+len(r.Standards))
	out = append(out, r.Spec...)
	out = append(out, r.Standards...)
	return out
}

// requiredBeforeHuman are artifacts that must exist before seal consideration.
var requiredBeforeHuman = []phase.ID{
	phase.Idea,
	phase.Strategy,
	phase.Spec,
	phase.TestAuthoring,
	phase.QA,
	phase.Documentation,
}

// Assess returns the two-axis report (F38) and writes {product}_TASKS.md (F39).
func Assess(workspaceRoot, product string) Report {
	var r Report
	add := func(axis Axis, id phase.ID, path, msg string) {
		f := Finding{Axis: axis, Phase: id, Path: path, Msg: msg}
		if axis == AxisSpec {
			r.Spec = append(r.Spec, f)
		} else {
			r.Standards = append(r.Standards, f)
		}
	}

	for _, id := range requiredBeforeHuman {
		if id == phase.TestAuthoring {
			if err := RequireTests(workspaceRoot, product); err != nil {
				rel, _ := artifact.Path(product, id)
				add(AxisStandards, id, rel, err.Error())
			}
			continue
		}
		rel, ok := artifact.Path(product, id)
		if !ok {
			continue
		}
		ne, err := artifact.NonEmpty(workspaceRoot, rel)
		if err != nil {
			add(AxisStandards, id, rel, err.Error())
			continue
		}
		if !ne {
			add(AxisStandards, id, rel, "missing or empty")
		}
	}
	ta := artifact.TAPath(product)
	if ne, err := artifact.NonEmpty(workspaceRoot, ta); err != nil {
		add(AxisStandards, phase.Spec, ta, err.Error())
	} else if !ne {
		add(AxisStandards, phase.Spec, ta, "missing or empty (F13 one-spec-set TA)")
	}
	if err := freeze.RequireValid(workspaceRoot, product); err != nil {
		add(AxisSpec, phase.Spec, freeze.Path(), err.Error())
	}
	if s, ok, err := itempack.OpenCommitted(workspaceRoot, product); err != nil {
		add(AxisSpec, phase.Spec, itempack.ProfilePath(), err.Error())
	} else if ok {
		for _, f := range s.Check() {
			add(AxisSpec, phase.Spec, itempack.ProfilePath(), "consistency "+f.String())
		}
	}
	prd, _ := artifact.Path(product, phase.Spec)
	if ne, err := artifact.NonEmpty(workspaceRoot, prd); err == nil && ne {
		if err := RequireSpecShape(workspaceRoot, product); err != nil {
			add(AxisSpec, phase.Spec, prd, err.Error())
		} else if err := RequireRequirementCoverage(workspaceRoot, product); err != nil {
			add(AxisSpec, phase.TestAuthoring, prd, err.Error())
		}
	}
	if log, ok, err := testrun.ReadLastLog(workspaceRoot); err == nil && ok {
		if strings.Contains(log, "# ok: false") {
			add(AxisStandards, phase.Implementation, evidence.Source(), "last framework test run recorded FAIL")
		}
		tests, _ := artifact.Path(product, phase.TestAuthoring)
		if ev, err := evidence.Derive(workspaceRoot, product, prd, tests); err == nil {
			if miss := ev.UnmetSoftware(); len(miss) > 0 {
				add(AxisSpec, phase.QA, evidence.Source(), "software evidence not green: "+strings.Join(miss, ", "))
			}
		}
	}

	r.Tasks = artifact.TasksPath(product)
	_ = writeTasks(workspaceRoot, product, r)
	return r
}

func writeTasks(workspaceRoot, product string, r Report) error {
	var b strings.Builder
	b.WriteString("# Tasks — " + product + "\n\n")
	b.WriteString("Generated by `dudes check` (F39 converge-to-tasks). Human seal still required.\n\n")
	if r.OK() {
		b.WriteString("## Unmet\n\n- [x] none — both review axes are clean\n")
		return artifact.Write(workspaceRoot, artifact.TasksPath(product), []byte(b.String()))
	}
	b.WriteString("## Unmet\n\n")
	for _, f := range r.All() {
		fmt.Fprintf(&b, "- [ ] (%s) %s\n", f.Axis, f.Error())
	}
	b.WriteString("\n## Repair brief\n\n")
	b.WriteString("- **Current:** see findings above\n")
	b.WriteString("- **Expected:** spec shape + requirement coverage + required artifacts\n")
	b.WriteString("- **Unchanged:** frozen PRD hash unless spec phase is re-run\n")
	return artifact.Write(workspaceRoot, artifact.TasksPath(product), []byte(b.String()))
}

// Converge reports mechanical findings for a release candidate (F34).
// Empty slice = ready for human decision / seal path.
func Converge(workspace, product string) []Finding {
	return Assess(workspace, product).All()
}

// ReadyForHuman is true when Assess finds nothing on either axis.
func ReadyForHuman(workspace, product string) bool {
	return Assess(workspace, product).OK()
}

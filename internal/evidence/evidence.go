// Package evidence derives requirement → test → run rows (F67) from the spec and the recorded test runs.
package evidence

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const (
	ClassSoftware = "software"
	ClassField    = "field"
	StatusPlanned = "planned"
	StatusRed     = "red"
	StatusGreen   = "green"
	StatusBlocked = "blocked"
)

// Row is one traceability cell.
type Row struct {
	Requirement string    `json:"requirement"`
	Test        string    `json:"test,omitempty"`
	Commit      string    `json:"commit,omitempty"`
	RunRef      string    `json:"run_ref,omitempty"`
	Status      string    `json:"status"`
	Class       string    `json:"class"`
	Variant     string    `json:"variant,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}

// File is the evidence of one product: computed, never stored.
type File struct {
	Product string `json:"product"`
	Rows    []Row  `json:"rows"`
}

// Source names where the results come from: the test-run notes (testrun.NotesRef).
func Source() string { return "refs/notes/" + testrun.NotesRef }

// Derive computes the evidence (F67): rows planned from the frozen PRD (with the contract's variant
// and field rows), results from the newest recorded test run in HEAD's history (a git note on the
// tested commit). Computed, never stored — like derived status (F70). Rows exist once tests were
// authored from the PRD: no PRD or no authored tests → no rows.
func Derive(workspaceRoot, product, prdRel, testsRel string) (*File, error) {
	f := &File{Product: product}
	con, ok, err := contract.Load(workspaceRoot)
	if err != nil {
		return nil, err
	}
	// Tests are authored when the test phase's artifact exists — or, for a contract that names its
	// tests (F88), when one of those files does.
	named := ok && len(con.ExistingTests(workspaceRoot)) > 0
	if !named {
		if testsRel == "" {
			return f, nil
		}
		testsAbs, err := workspace.Resolve(workspaceRoot, testsRel)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(testsAbs); err != nil {
			return f, nil // no authored tests yet
		}
	}
	switch {
	case ok && con.Item != "":
		f = PlanItemRows(product, con) // item-bound: one row per acceptance criterion
	default:
		if f, err = planFromPRD(workspaceRoot, product, prdRel, con); err != nil || len(f.Rows) == 0 {
			return f, err
		}
	}
	if ok {
		for _, a := range con.Acceptance {
			if strings.EqualFold(a.Kind, "field") {
				f.EnsureFieldRow(a.ID, a.Text)
			}
		}
	}
	log, commit, found, err := testrun.LastRun(workspaceRoot)
	if err != nil {
		return nil, err
	}
	acc, err := testrun.Acceptance(workspaceRoot)
	if err != nil {
		return nil, err
	}
	// The newer run is applied last. An acceptance rerun follows its gate on the same commit, so it
	// wins unless the unit run tested a later commit.
	older, newer := map[string]testrun.Outcome{}, map[string]testrun.Outcome{}
	before := map[string]bool{} // commit → strictly before the unit run's
	for name, o := range acc {
		b, seen := before[o.Commit]
		if !seen {
			b = found && o.Commit != commit && gitx.IsAncestor(workspaceRoot, o.Commit, commit)
			before[o.Commit] = b
		}
		if b {
			older[name] = o
		} else {
			newer[name] = o
		}
	}
	f.ApplyResults(older, AcceptanceSource())
	if found {
		f.ApplyRun(log, commit, Source())
	}
	f.ApplyResults(newer, AcceptanceSource())
	return f, nil
}

// AcceptanceSource names the acceptance-run notes (testrun.AcceptanceNotesRef).
func AcceptanceSource() string { return "refs/notes/" + testrun.AcceptanceNotesRef }

// ApplyResults updates software rows from per-test outcomes (testrun.Acceptance): pass is green,
// fail and skip are red — a skipped criterion is not green. Rows whose test has no outcome stay as
// they are.
func (f *File) ApplyResults(results map[string]testrun.Outcome, runRef string) {
	if f == nil {
		return
	}
	now := time.Now().UTC()
	for i := range f.Rows {
		o, ok := results[f.Rows[i].Test]
		if f.Rows[i].Class != ClassSoftware || f.Rows[i].Test == "" || !ok {
			continue
		}
		f.Rows[i].Commit = o.Commit
		f.Rows[i].RunRef = runRef
		f.Rows[i].UpdatedAt = now
		if o.Result == testrun.Pass {
			f.Rows[i].Status = StatusGreen
		} else {
			f.Rows[i].Status = StatusRed
		}
	}
}

func planFromPRD(workspaceRoot, product, prdRel string, con *contract.Contract) (*File, error) {
	f := &File{Product: product}
	if prdRel == "" {
		return f, nil
	}
	prdAbs, err := workspace.Resolve(workspaceRoot, prdRel)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(prdAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return f, nil
		}
		return nil, err
	}
	variant := "standard"
	if con != nil && con.Variant != "" {
		variant = con.Variant
	}
	return PlanSoftwareRows(product, variant, spec.Parse(string(data))), nil
}

// PlanItemRows plans one software row per behavior criterion of an item-bound contract; its test
// is the one test authoring generated for it (spec.AcceptanceTestName).
func PlanItemRows(product string, con *contract.Contract) *File {
	f := &File{Product: product}
	now := time.Now().UTC()
	for _, a := range con.Acceptance {
		if !strings.EqualFold(a.Kind, "behavior") {
			continue
		}
		f.Rows = append(f.Rows, Row{
			Requirement: a.ID,
			Test:        spec.AcceptanceTestName(a.ID),
			Status:      StatusPlanned,
			Class:       ClassSoftware,
			Variant:     con.Variant,
			UpdatedAt:   now,
		})
	}
	return f
}

// PlanSoftwareRows creates planned software rows from a parsed PRD (not field hypotheses).
func PlanSoftwareRows(product, variant string, d *spec.Document) *File {
	f := &File{Product: product}
	if d == nil {
		return f
	}
	now := time.Now().UTC()
	for _, id := range d.IDs() {
		f.Rows = append(f.Rows, Row{
			Requirement: id,
			Test:        "TestReq_" + safeIdent(id),
			Status:      StatusPlanned,
			Class:       ClassSoftware,
			Variant:     variant,
			UpdatedAt:   now,
		})
	}
	return f
}

func safeIdent(id string) string {
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
	return s
}

// ApplyRun updates software rows from a test log and commit (F67).
// Only rows whose test name appears in the log are updated — a global PASS
// must not green-wash unmentioned requirements.
func (f *File) ApplyRun(logText, commit, runRef string) {
	if f == nil {
		return
	}
	now := time.Now().UTC()
	ok := strings.Contains(logText, "# ok: true") || strings.Contains(logText, "PASS")
	fail := strings.Contains(logText, "# ok: false") || strings.Contains(logText, "FAIL")
	for i := range f.Rows {
		if f.Rows[i].Class != ClassSoftware {
			continue
		}
		name := f.Rows[i].Test
		if name == "" || !strings.Contains(logText, name) {
			continue
		}
		f.Rows[i].Commit = commit
		f.Rows[i].RunRef = runRef
		f.Rows[i].UpdatedAt = now
		switch {
		case fail:
			f.Rows[i].Status = StatusRed
		case ok:
			f.Rows[i].Status = StatusGreen
		}
	}
}

// UnmetSoftware IDs still not green.
func (f *File) UnmetSoftware() []string {
	if f == nil {
		return nil
	}
	var out []string
	for _, r := range f.Rows {
		if r.Class == ClassSoftware && r.Status != StatusGreen {
			out = append(out, r.Requirement)
		}
	}
	return out
}

// EnsureFieldRow records a business hypothesis that CI must not green-wash.
func (f *File) EnsureFieldRow(id, text string) {
	if f == nil || id == "" {
		return
	}
	for _, r := range f.Rows {
		if r.Requirement == id && r.Class == ClassField {
			return
		}
	}
	f.Rows = append(f.Rows, Row{
		Requirement: id,
		Status:      StatusBlocked,
		Class:       ClassField,
		RunRef:      text,
		UpdatedAt:   time.Now().UTC(),
	})
}

// FormatError is a mechanical fail when software rows are not green.
func (f *File) FormatError() error {
	miss := f.UnmetSoftware()
	if len(miss) == 0 {
		return nil
	}
	return fmt.Errorf("evidence: software rows not green: %s", strings.Join(miss, ", "))
}

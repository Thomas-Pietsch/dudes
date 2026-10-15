package itempack_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// sampleSet loads the sample product (examples/sample) the way an instance workspace would: its
// profile, its spec-set registry, its documents. The roots live in this repository, so nothing
// here ever skips. It returns the set and the sample's directory.
func sampleSet(t *testing.T) (*itempack.Set, string) {
	t.Helper()
	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..", "..")
	ex := filepath.Join(root, "examples", "sample")

	var set specset.Set
	readJSON(t, filepath.Join(ex, "specset.json"), &set)
	for _, r := range set.Roots {
		if _, err := os.Stat(filepath.Join(root, r)); err != nil {
			t.Fatalf("sample root %s not present: %v", r, err)
		}
	}
	data, err := os.ReadFile(filepath.Join(ex, "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	prof, err := itempack.ParseProfile(data)
	if err != nil {
		t.Fatal(err)
	}
	docs, err := itempack.SpecSetDocs(root, &set)
	if err != nil {
		t.Fatal(err)
	}
	s, err := itempack.New(prof, docs)
	if err != nil {
		t.Fatal(err)
	}
	return s, ex
}

// Conformance against the sample product: zero consistency findings and every work item packable.
func TestSampleSpecsConform(t *testing.T) {
	s, _ := sampleSet(t)
	for _, key := range []string{"PS", "PRD", "TA", "UX", "QR", "BACKLOG", "ROADMAP"} {
		if !s.Has(key) {
			t.Fatalf("%s not loaded — check examples/sample/specset.json", key)
		}
	}
	if f := s.Check(); len(f) != 0 {
		for _, x := range f {
			t.Error(x)
		}
		t.Fatalf("sample spec-set: %d consistency finding(s)", len(f))
	}
	packed := 0
	for _, it := range s.Items() {
		if it.Kind != itempack.KindItem {
			continue
		}
		p, err := s.Build(it.ID, 0)
		if err != nil {
			t.Fatalf("%s not ready: %v", it.ID, err)
		}
		if len(p.Manifest.References) == 0 {
			t.Fatalf("%s: pack has no references", it.ID)
		}
		packed++
	}
	if packed == 0 {
		t.Fatal("no packable work items")
	}
	t.Logf("sample: 0 findings, %d items packable", packed)
}

// A task's refs narrow the item pack (F65). A ref the item does not make would hand the worker an
// empty spec — a slicing is refused for it (F88) — so the sample's contract must not carry one.
func TestSampleContractRefsAreTheItemsOwn(t *testing.T) {
	s, ex := sampleSet(t)
	var c contract.Contract
	readJSON(t, filepath.Join(ex, "contract.json"), &c)
	it, ok := s.FindItem(c.Item)
	if !ok {
		t.Fatalf("%s is not a work item of the sample", c.Item)
	}
	_, byDoc := s.Targets(it)
	have := map[string]bool{}
	for doc, targets := range byDoc {
		for _, target := range targets {
			have[doc+":"+target] = true
		}
	}
	n := 0
	for _, tk := range c.Tasks {
		for _, r := range tk.Refs {
			n++
			if !have[strings.Trim(strings.TrimSpace(r), "[]")] {
				t.Errorf("task %s: %s is not a reference of %s", tk.ID, r, c.Item)
			}
		}
	}
	if n == 0 {
		t.Fatal("the sample contract narrows no task by refs")
	}
}

// In a dev repository of several products, a work item finds its product by its ID family: the
// profile's owner that gives the backlog its IDs (F88). Every item of the sample is the sample's; a
// criterion, a TA ID or another product's item is not.
func TestSampleProfileOwnsItsItems(t *testing.T) {
	s, _ := sampleSet(t)
	for _, it := range s.Items() {
		if !s.Profile.OwnsItem(it.ID) || !s.Profile.OwnsItem(strings.ToLower(it.ID)) {
			t.Errorf("%s is not the sample's", it.ID)
		}
	}
	for _, id := range []string{"AC-SHELF-01", "API-01", "BL-002", "BK-1", "BK-0012"} {
		if s.Profile.OwnsItem(id) {
			t.Errorf("%s counts as a work item of the sample", id)
		}
	}
	if got := s.Profile.ItemFamily(); got != `BK-\d{3}` {
		t.Errorf("item family %q", got)
	}
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatal(err)
	}
}

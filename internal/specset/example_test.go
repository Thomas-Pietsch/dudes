package specset_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

func TestSampleSpecsetKeepsNames(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(this), "..", "..", "examples", "sample", "specset.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s specset.Set
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	if s.Product != "bookshelf" {
		t.Fatalf("product=%s", s.Product)
	}
	found := false
	for _, e := range s.Entries {
		if e.Path == "BOOKSHELF_PRD.md" && e.Root == "pm" {
			found = true
		}
		if e.Path == "product_PRD.md" {
			t.Fatal("must not rename the product's own files")
		}
	}
	if !found {
		t.Fatal("expected original PRD filename")
	}
}

func TestExampleSpecsetsRegisterCompanions(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..", "..", "examples")
	wantKind := map[string]string{
		"UX":      specset.KindUX,
		"ROADMAP": specset.KindRoadmap,
		"BACKLOG": specset.KindBacklog,
	}
	paths, err := filepath.Glob(filepath.Join(root, "*", "specset.json"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no example spec-set under examples/: %v", err)
	}
	for _, path := range paths {
		product := filepath.Base(filepath.Dir(path))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var s specset.Set
		if err := json.Unmarshal(data, &s); err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, e := range s.Entries {
			got[e.ID] = e.Kind
			if !specset.KnownKind(e.Kind) {
				t.Fatalf("%s: unknown kind %q on %s", product, e.Kind, e.ID)
			}
		}
		for id, kind := range wantKind {
			if got[id] != kind {
				t.Fatalf("%s: %s kind=%q want %q", product, id, got[id], kind)
			}
		}
	}
}

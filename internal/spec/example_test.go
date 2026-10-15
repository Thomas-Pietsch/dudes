package spec_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/spec"
)

// The sample product's PRD is written in the shape the engine reads (F37): every requirement holds
// a SHALL line and a complete GIVEN/WHEN/THEN scenario.
func TestSamplePRDHasTheShape(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	data, err := os.ReadFile(filepath.Join(filepath.Dir(this), "..", "..", "examples", "sample", "pm", "BOOKSHELF_PRD.md"))
	if err != nil {
		t.Fatal(err)
	}
	d := spec.Parse(string(data))
	if err := spec.ShapeError(d); err != nil {
		t.Fatal(err)
	}
	if got := d.IDs(); len(got) != 3 || got[0] != "REQ-ADD" {
		t.Fatalf("ids=%v", got)
	}
}

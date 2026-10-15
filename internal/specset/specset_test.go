package specset_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

func TestEngineSetAndSaveLoad(t *testing.T) {
	dir := t.TempDir()
	s := specset.EngineSet("demo")
	if len(s.Entries) < 5 {
		t.Fatalf("entries=%d", len(s.Entries))
	}
	if err := specset.Save(dir, s); err != nil {
		t.Fatal(err)
	}
	got, ok, err := specset.Load(dir, "demo")
	if err != nil || !ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	if got.Product != "demo" {
		t.Fatalf("%+v", got)
	}
	abs, err := got.ResolveAbs(dir, got.Entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Dir(abs) != dir && filepath.Dir(abs) != filepath.Clean(dir) {
		// idea file is in dir
	}
	_ = os.WriteFile(abs, []byte("x"), 0o644)
}

func TestNamedSiblingRoot(t *testing.T) {
	ws := t.TempDir()
	sib := t.TempDir()
	if err := os.WriteFile(filepath.Join(sib, "KEEP_NAME_PRD.md"), []byte("# keep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(ws, sib)
	if err != nil {
		t.Fatal(err)
	}
	s := &specset.Set{
		Product: "keep",
		Roots:   map[string]string{"pm": filepath.ToSlash(rel)},
		Entries: []specset.Entry{{ID: "PRD", Kind: "prd", Root: "pm", Path: "KEEP_NAME_PRD.md"}},
	}
	abs, err := s.ResolveAbs(ws, s.Entries[0])
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(abs) != "KEEP_NAME_PRD.md" {
		t.Fatalf("renamed? %s", abs)
	}
}

func TestKnownKind(t *testing.T) {
	if !specset.KnownKind("roadmap") || !specset.KnownKind("BACKLOG") || !specset.KnownKind("ux") {
		t.Fatal("companions must be known kinds")
	}
	if specset.KnownKind("company") {
		t.Fatal("company is not a spec-set kind")
	}
}

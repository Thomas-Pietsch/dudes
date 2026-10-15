package freeze_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/freeze"
)

// A spec_freeze.json from before freezes moved into git becomes a note once, then the file is gone.
func TestLoadMigratesSpecFreezeJSON(t *testing.T) {
	dir := gitRepo(t)
	legacy := filepath.Join(dir, ".dudes", "spec_freeze.json")
	if err := os.MkdirAll(filepath.Dir(legacy), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(`{"product":"demo","prd":"demo_PRD.md","sha256":"abc"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	rec, ok, err := freeze.Load(dir)
	if err != nil || !ok || rec.Product != "demo" || rec.SHA256 != "abc" {
		t.Fatalf("migrate: %+v ok=%v err=%v", rec, ok, err)
	}
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Fatal("spec_freeze.json should be gone after the migration")
	}
	if rec, ok, _ := freeze.Load(dir); !ok || rec.SHA256 != "abc" {
		t.Fatal("migrated freeze should load from git")
	}
}

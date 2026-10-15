package freeze_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// F69: with a profile, freeze refuses an inconsistent spec-set.
func TestFreezeRefusesInconsistentSpecSet(t *testing.T) {
	dir := gitRepo(t)
	src := filepath.Join("..", "itempack", "testdata", "demo")
	specs := filepath.Join(dir, "specs")
	_ = os.MkdirAll(specs, 0o755)
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	set := &specset.Set{Product: "demo", Roots: map[string]string{"specs": "specs"}}
	for _, e := range entries {
		data, _ := os.ReadFile(filepath.Join(src, e.Name()))
		_ = os.WriteFile(filepath.Join(specs, e.Name()), data, 0o644)
		if e.Name() == "profile.json" {
			continue
		}
		key := strings.TrimSuffix(strings.TrimPrefix(e.Name(), "DEMO_"), ".md")
		set.Entries = append(set.Entries, specset.Entry{ID: key, Kind: "other", Root: "specs", Path: e.Name()})
	}
	if err := specset.Save(dir, set); err != nil {
		t.Fatal(err)
	}
	prof, _ := os.ReadFile(filepath.Join(src, "profile.json"))
	_ = artifact.Write(dir, ".dudes/profile.json", prof)
	commit(t, dir, "spec")

	if _, err := freeze.WriteAll(dir, "demo", "run-1"); err != nil {
		t.Fatalf("consistent spec-set must freeze: %v", err)
	}
	backlog := filepath.Join(specs, "DEMO_BACKLOG.md")
	data, _ := os.ReadFile(backlog)
	_ = os.WriteFile(backlog, []byte(strings.ReplaceAll(string(data), "[[TA:SLO-01]]", "[[TA:SLO-09]]")), 0o644)
	commit(t, dir, "broken backlog")
	_, err = freeze.WriteAll(dir, "demo", "run-2")
	if err == nil || !strings.Contains(err.Error(), "consistency") {
		t.Fatalf("expected consistency refusal, got %v", err)
	}
}

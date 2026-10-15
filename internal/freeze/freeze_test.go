package freeze_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

func TestWriteRequireAndDrift(t *testing.T) {
	dir := gitRepo(t)
	if err := freeze.RequireValid(dir, "demo"); err == nil {
		t.Fatal("expected missing freeze")
	}
	_ = artifact.Write(dir, "demo_PRD.md", []byte("# PRD\nv1\n"))
	commit(t, dir, "spec v1")
	rec, err := freeze.Write(dir, "demo", "run-1")
	if err != nil || rec.SHA256 == "" {
		t.Fatalf("%v %+v", err, rec)
	}
	if err := freeze.RequireValid(dir, "demo"); err != nil {
		t.Fatal(err)
	}
	_ = artifact.Write(dir, "demo_PRD.md", []byte("# PRD\nv2\n"))
	if err := freeze.RequireValid(dir, "demo"); err != nil {
		t.Fatalf("an uncommitted edit is work in progress, not drift: %v", err)
	}
	commit(t, dir, "spec v2")
	if err := freeze.RequireValid(dir, "demo"); err == nil {
		t.Fatal("expected drift after the change was committed")
	}
}

func TestWriteAllIncludesCompanion(t *testing.T) {
	dir := gitRepo(t)
	_ = artifact.Write(dir, "demo_PRD.md", []byte("# PRD\nv1\n"))
	_ = artifact.Write(dir, "demo_TA.md", []byte("# TA\n"))
	commit(t, dir, "spec")
	rec, err := freeze.WriteAll(dir, "demo", "run-2")
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Inputs) == 0 {
		t.Fatal("expected inputs")
	}
	foundTA := false
	for _, in := range rec.Inputs {
		if in.Path == "demo_TA.md" {
			foundTA = true
		}
	}
	if !foundTA {
		t.Fatalf("inputs=%+v", rec.Inputs)
	}
	if err := freeze.RequireValid(dir, "demo"); err != nil {
		t.Fatal(err)
	}
	_ = artifact.Write(dir, "demo_TA.md", []byte("# TA\nv2\n"))
	commit(t, dir, "TA v2")
	if err := freeze.RequireValid(dir, "demo"); err == nil {
		t.Fatal("expected companion drift")
	}
}

func TestWriteAllSkipsMissingCompanions(t *testing.T) {
	dir := gitRepo(t)
	_ = artifact.Write(dir, "demo_PRD.md", []byte("# PRD\nv1\n"))
	s := specset.EngineSet("demo")
	s.Entries = append(s.Entries,
		specset.Entry{ID: "UX", Kind: specset.KindUX, Root: "workspace", Path: "demo_UX.md"},
		specset.Entry{ID: "ROADMAP", Kind: specset.KindRoadmap, Root: "workspace", Path: "demo_ROADMAP.md"},
		specset.Entry{ID: "BACKLOG", Kind: specset.KindBacklog, Root: "workspace", Path: "demo_BACKLOG.md"},
	)
	if err := specset.Save(dir, s); err != nil {
		t.Fatal(err)
	}
	commit(t, dir, "spec")
	rec, err := freeze.WriteAll(dir, "demo", "run-3")
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range rec.Inputs {
		if in.Path == "demo_UX.md" || in.Path == "demo_ROADMAP.md" || in.Path == "demo_BACKLOG.md" {
			t.Fatalf("missing companion should be skipped: %+v", in)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "demo_UX.md"), []byte("# UX\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	commit(t, dir, "UX")
	rec, err = freeze.WriteAll(dir, "demo", "run-4")
	if err != nil {
		t.Fatal(err)
	}
	foundUX := false
	for _, in := range rec.Inputs {
		if in.Path == "demo_UX.md" {
			foundUX = true
		}
	}
	if !foundUX {
		t.Fatalf("existing UX must join freeze: %+v", rec.Inputs)
	}
}

func TestWriteAllUsesSpecsetPRD(t *testing.T) {
	ws := gitRepo(t)
	pm := t.TempDir()
	if err := os.WriteFile(filepath.Join(pm, "KEEP_NAME_PRD.md"), []byte("# instance PRD\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(ws, pm)
	if err != nil {
		t.Fatal(err)
	}
	s := &specset.Set{
		Product: "keep",
		Roots:   map[string]string{"pm": filepath.ToSlash(rel)},
		Entries: []specset.Entry{{ID: "PRD", Kind: specset.KindPRD, Root: "pm", Path: "KEEP_NAME_PRD.md"}},
	}
	if err := specset.Save(ws, s); err != nil {
		t.Fatal(err)
	}
	rec, err := freeze.WriteAll(ws, "keep", "run-specset")
	if err != nil {
		t.Fatal(err)
	}
	if rec.PRD != "KEEP_NAME_PRD.md" {
		t.Fatalf("prd=%s", rec.PRD)
	}
	if rec.PRDRoot != "pm" {
		t.Fatalf("prd_root=%s", rec.PRDRoot)
	}
	found := false
	for _, in := range rec.Inputs {
		if in.Path == "KEEP_NAME_PRD.md" {
			if in.Root != "pm" {
				t.Fatalf("prd input root=%s", in.Root)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("inputs=%+v", rec.Inputs)
	}
	if err := freeze.RequireValid(ws, "keep"); err != nil {
		t.Fatal(err)
	}
}

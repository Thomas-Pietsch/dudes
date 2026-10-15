package recipe

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestLoadMissing(t *testing.T) {
	dir := t.TempDir()
	r, ok, err := Load(dir, phase.Idea)
	if err != nil || ok || r != nil {
		t.Fatalf("want missing ok=false, got r=%v ok=%v err=%v", r, ok, err)
	}
}

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	recipes := filepath.Join(dir, Dir)
	if err := os.MkdirAll(recipes, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phase":"idea","worker":"ollama","instructions":"Be brief.","pack":true,"tools":["none"],"skills":["tdd"],"rules":["CORE.md"]}`
	if err := os.WriteFile(filepath.Join(recipes, "idea.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r, ok, err := Load(dir, phase.Idea)
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if r.Instructions != "Be brief." || !r.Pack || r.Worker != "ollama" {
		t.Fatalf("unexpected recipe: %+v", r)
	}
	if len(r.Skills) != 1 || r.Skills[0] != "tdd" || len(r.Rules) != 1 {
		t.Fatalf("skills/rules: %+v", r)
	}
	if r.TaskOverride() != "Be brief." {
		t.Fatalf("TaskOverride: %q", r.TaskOverride())
	}
}

func TestLoadPhaseMismatch(t *testing.T) {
	dir := t.TempDir()
	recipes := filepath.Join(dir, Dir)
	if err := os.MkdirAll(recipes, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phase":"strategy","worker":"ollama"}`
	if err := os.WriteFile(filepath.Join(recipes, "idea.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(dir, phase.Idea)
	if err == nil {
		t.Fatal("expected phase mismatch error")
	}
}

func TestLoadPiWorker(t *testing.T) {
	dir := t.TempDir()
	recipes := filepath.Join(dir, Dir)
	if err := os.MkdirAll(recipes, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phase":"implementation","worker":"pi","tools":["edit"]}`
	if err := os.WriteFile(filepath.Join(recipes, "implementation.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r, ok, err := Load(dir, phase.Implementation)
	if err != nil || !ok || r.Worker != "pi" {
		t.Fatalf("ok=%v err=%v r=%+v", ok, err, r)
	}
}

func TestLoadBadWorker(t *testing.T) {
	dir := t.TempDir()
	recipes := filepath.Join(dir, Dir)
	if err := os.MkdirAll(recipes, 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"phase":"idea","worker":"chatgpt"}`
	if err := os.WriteFile(filepath.Join(recipes, "idea.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, err := Load(dir, phase.Idea)
	if err == nil {
		t.Fatal("expected bad worker error")
	}
}

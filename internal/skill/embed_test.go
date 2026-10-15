package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// An instance workspace holds only the company's own skills; the framework's level skills come
// from the binary — and the framework folder never does.
func TestLoadFromBinary(t *testing.T) {
	ws := t.TempDir()
	for _, name := range []string{"tdd", "research", "work-item"} {
		if body, ok, err := Load(ws, name); err != nil || !ok || body == "" {
			t.Errorf("%s from the binary: ok=%v err=%v", name, ok, err)
		}
	}
	if _, ok, _ := Load(ws, "maintain"); ok {
		t.Error("a framework skill must not come from the binary")
	}
	own := filepath.Join(ws, Dir, "lab-check")
	if err := os.MkdirAll(own, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(own, FileName), []byte("---\nname: lab-check\n---\nown\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if body, ok, _ := Load(ws, "lab-check"); !ok || !strings.Contains(body, "own") {
		t.Errorf("the instance's own skill: ok=%v", ok)
	}
}

// Every skill a shipped contract or recipe names resolves from an instance workspace — a rename
// that leaves a dangling name fails here, not in a worker run.
func TestNamedSkillsResolve(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..", "..")
	files, _ := filepath.Glob(filepath.Join(root, "examples", "*", "contract.json"))
	recipes, _ := filepath.Glob(filepath.Join(root, "recipes", "*.json"))
	ws := t.TempDir()
	n := 0
	for _, f := range append(files, recipes...) {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			Skills []string `json:"skills"`
			Tasks  []struct {
				Skills []string `json:"skills"`
			} `json:"tasks"`
		}
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		names := doc.Skills
		for _, tk := range doc.Tasks {
			names = append(names, tk.Skills...)
		}
		for _, name := range names {
			n++
			if _, err := LoadMany(ws, []string{name}); err != nil {
				t.Errorf("%s names %q: %v", filepath.Base(filepath.Dir(f))+"/"+filepath.Base(f), name, err)
			}
		}
	}
	if n == 0 {
		t.Fatal("no skill names found in examples or recipes")
	}
}

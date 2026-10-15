package skill

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadMany(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, Dir, "tdd")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, FileName), []byte("# TDD\n\nWrite a failing test first.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := LoadMany(dir, []string{"tdd"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "failing test") || !strings.Contains(body, "Skill: tdd") {
		t.Fatalf("body=%s", body)
	}
}

func TestRejectPathEscape(t *testing.T) {
	if _, _, err := Load(t.TempDir(), "../secret"); err == nil {
		t.Fatal("expected invalid name")
	}
	if ValidName("tdd") != true || ValidName("TDD") {
		t.Fatal("name rules")
	}
}

func TestMissingNamedSkillErrors(t *testing.T) {
	if _, err := LoadMany(t.TempDir(), []string{"nope"}); err == nil {
		t.Fatal("expected missing skill error")
	}
}

func TestLoadRules(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "STD.md"), []byte("use tabs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	body, err := LoadRules(dir, []string{"STD.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, "use tabs") {
		t.Fatalf("body=%s", body)
	}
}

func TestShippedDisciplineSkills(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(this), "..", "..")
	for _, name := range []string{"tdd", "diagnosing-bugs", "code-review", "codebase-design", "writing-for-agents", "delimited-output", "work-item"} {
		body, ok, err := Load(root, name)
		if err != nil || !ok || body == "" {
			t.Fatalf("skill %s: ok=%v err=%v", name, ok, err)
		}
	}
}

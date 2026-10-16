package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A workspace is its repository's top level (P): inside another repository, git would put the
// engine's commits and notes in the outer one, so the engine refuses and says where they would land.
func TestWorkspaceMustBeItsRepositoryTopLevel(t *testing.T) {
	outer := t.TempDir()
	gitIn(t, outer, "init", "-q", "-b", "main")
	nested := filepath.Join(outer, "sub", "ws")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	cfgFor := func(dir string) string {
		cfg := filepath.Join(t.TempDir(), "dudes.json")
		body := `{"workspace": "` + filepath.ToSlash(dir) + `", "product": "demo"}`
		if err := os.WriteFile(cfg, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return cfg
	}

	for _, cmd := range []string{"status", "init"} {
		code, _, errOut := captured(func() int { return run([]string{"-config", cfgFor(nested), cmd}) })
		if code != 1 || !strings.Contains(errOut, "workspace "+nested) || !strings.Contains(errOut, "inside the repository") || !strings.Contains(errOut, "commits and notes would land in") {
			t.Fatalf("%s in a nested workspace = %d, %q", cmd, code, errOut)
		}
		// both paths, as the file system has them: the outer top level is named too
		top, _ := filepath.EvalSymlinks(outer)
		if !strings.Contains(strings.ToLower(filepath.ToSlash(errOut)), strings.ToLower(filepath.ToSlash(top))) {
			t.Fatalf("the refusal names no outer repository %s: %q", top, errOut)
		}
	}
	if _, err := os.Stat(filepath.Join(nested, ".dudes")); err == nil {
		t.Fatal("a refused command wrote engine state into the nested folder")
	}

	// the top level itself works
	if code, out, errOut := captured(func() int { return run([]string{"-config", cfgFor(outer), "status"}) }); code != 0 || !strings.Contains(out, "product:   demo") {
		t.Fatalf("status at the top level = %d, %q, %q", code, out, errOut)
	}
	// so does a folder that is a repository of its own, inside another one
	inner := filepath.Join(outer, "own")
	if err := os.MkdirAll(inner, 0o755); err != nil {
		t.Fatal(err)
	}
	gitIn(t, inner, "init", "-q", "-b", "main")
	if code, _, errOut := captured(func() int { return run([]string{"-config", cfgFor(inner), "status"}) }); code != 0 {
		t.Fatalf("status in a repository of its own = %d, %q", code, errOut)
	}
	// a link to the top level is the top level
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(outer, link); err == nil {
		if code, _, errOut := captured(func() int { return run([]string{"-config", cfgFor(link), "status"}) }); code != 0 {
			t.Fatalf("status through a link to the top level = %d, %q", code, errOut)
		}
	}
}

// What needs no repository stays allowed in a nested folder: help, version, phases, and a folder
// with no repository at all. doctor still runs and reports the finding as its own failed check.
func TestNestedWorkspaceLeavesRepositoryFreeCommandsAlone(t *testing.T) {
	outer := t.TempDir()
	gitIn(t, outer, "init", "-q", "-b", "main")
	nested := filepath.Join(outer, "ws")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(t.TempDir(), "dudes.json")
	if err := os.WriteFile(cfg, []byte(`{"workspace": "`+filepath.ToSlash(nested)+`", "product": "demo"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"version"}, {"help"}, {"phases"}, {"-version"}} {
		if code, _, errOut := captured(func() int { return run(append([]string{"-config", cfg}, args...)) }); code != 0 {
			t.Errorf("%v in a nested workspace = %d, %q", args, code, errOut)
		}
	}
	code, out, _ := captured(func() int { return run([]string{"-config", cfg, "doctor"}) })
	if code != 1 || !strings.Contains(out, "inside the repository") || !strings.Contains(out, "doctor: FAIL") {
		t.Errorf("doctor in a nested workspace = %d, %q", code, out)
	}

	// no repository above: the engine does not ask for one here (doctor reports its absence)
	bare := t.TempDir()
	if err := os.WriteFile(cfg, []byte(`{"workspace": "`+filepath.ToSlash(bare)+`", "product": "demo"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := captured(func() int { return run([]string{"-config", cfg, "status"}) }); code != 0 {
		t.Errorf("status in a folder with no repository = %d, %q", code, errOut)
	}
}

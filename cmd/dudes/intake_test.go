package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// The bench fixture is a small spec repo: intake binds a named item, runs a named goal, refuses a
// second item while a run is under way, and answers anything else with the menu.
func TestIntake(t *testing.T) {
	ws := t.TempDir()
	src := filepath.Join("..", "..", "bench", "workspace")
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		data, err := os.ReadFile(p)
		if err == nil {
			_ = os.MkdirAll(filepath.Join(ws, filepath.Dir(rel)), 0o755)
			err = os.WriteFile(filepath.Join(ws, rel), data, 0o644)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(filepath.Join(ws, "dudes"), filepath.Join(ws, ".dudes")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ws, "Makefile"), []byte("hello: ## Say hello\n\t@echo hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"config", "user.email", "t@t"}, {"config", "user.name", "t"},
		{"config", "core.autocrlf", "false"}, {"add", "-A"}, {"commit", "-qm", "spec"}} {
		if out, err := exec.Command("git", append([]string{"-C", ws}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	cfgPath := filepath.Join(ws, ".dudes", "dudes.json")
	if err := os.WriteFile(cfgPath, []byte(fmt.Sprintf(`{"workspace":%q,"product":"bench"}`, filepath.ToSlash(ws))), 0o644); err != nil {
		t.Fatal(err)
	}

	if code := cmdDo(cfgPath, []string{"something", "vague"}); code != 2 {
		t.Fatalf("free text: exit %d, want the menu (2)", code)
	}
	// -pick: the model's choice is validated and printed, never run; a non-entry falls back to the menu.
	for _, reply := range []string{"`BL-002`", "build a whole new product"} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{"response": reply, "done": true})
		}))
		withModel := fmt.Sprintf(`{"workspace":%q,"product":"bench","ollama":{"base_url":%q,"model":"m"}}`, filepath.ToSlash(ws), srv.URL)
		if err := os.WriteFile(cfgPath, []byte(withModel), 0o644); err != nil {
			t.Fatal(err)
		}
		if code := cmdDo(cfgPath, []string{"-pick", "sort", "versions", "properly"}); code != 2 {
			t.Fatalf("pick %q: exit %d, want 2 (nothing started)", reply, code)
		}
		srv.Close()
		if _, ok, _ := contract.Load(ws); ok {
			t.Fatalf("pick %q bound an item", reply)
		}
	}
	if _, err := exec.LookPath("make"); err == nil {
		if code := cmdDo(cfgPath, []string{"hello"}); code != 0 {
			t.Fatalf("goal: exit %d", code)
		}
	}
	if code := cmdDo(cfgPath, []string{"build", "bl-001", "please"}); code != 0 {
		t.Fatalf("item: exit %d", code)
	}
	con, ok, err := contract.Load(ws)
	if err != nil || !ok || con.Item != "BL-001" || len(con.Acceptance) != 1 {
		t.Fatalf("contract: %+v %v %v", con, ok, err)
	}
	st, ok, err := runstate.Load(ws)
	if err != nil || !ok || st.Phase != phase.TestAuthoring {
		t.Fatalf("run: %+v %v %v", st, ok, err)
	}
	if code := cmdDo(cfgPath, []string{"BL-002"}); code != 1 {
		t.Fatalf("second item during a run: exit %d, want refusal", code)
	}
	if code := cmdDo(cfgPath, []string{"BL-001"}); code != 0 {
		t.Fatalf("the bound item again: exit %d", code)
	}
}

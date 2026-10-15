package goals_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/goals"
)

const mk = "# comment ## not a goal\n" +
	"VERSION := $(shell sh ci/version.sh) ## not a goal either\n" +
	"NOW  = sed -n '/^\\#\\# Now/p' TODO.md\n" +
	".PHONY: help build test\n" +
	"help: ## List the goals\n" +
	"build: web ## Full build: UI + binary\n" +
	"test:  ##   Run the unit tests  \n" +
	"secret:\n" +
	"docker-%: ## pattern rule\n"

func write(t *testing.T, name string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(mk), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestListParsesDocumentedTargetsOnly(t *testing.T) {
	gs, err := goals.List(write(t, "makefile"))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, g := range gs {
		names = append(names, g.Name+"="+g.Desc)
	}
	got := strings.Join(names, "|")
	want := "help=List the goals|build=Full build: UI + binary|test=Run the unit tests"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestListWithoutMakefile(t *testing.T) {
	gs, err := goals.List(t.TempDir())
	if err != nil || gs != nil {
		t.Fatalf("got %v, %v", gs, err)
	}
}

func TestSelectNamesEveryMissingGoal(t *testing.T) {
	gs, _ := goals.List(write(t, "Makefile"))
	if _, err := goals.Select(gs, []string{"test", "lint", "deploy"}); err == nil ||
		!strings.Contains(err.Error(), "lint, deploy") {
		t.Fatalf("want error naming lint, deploy; got %v", err)
	}
	sel, err := goals.Select(gs, []string{"test", "build"})
	if err != nil || len(sel) != 2 || sel[0].Name != "test" {
		t.Fatalf("got %v, %v", sel, err)
	}
	if b := goals.Block(sel); !strings.Contains(b, "- `make test` — Run the unit tests") {
		t.Fatalf("block: %q", b)
	}
}

func TestHas(t *testing.T) {
	dir := write(t, "Makefile")
	if !goals.Has(dir, "test") || goals.Has(dir, "secret") {
		t.Fatal("Has: want test documented, secret not")
	}
}

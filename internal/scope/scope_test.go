package scope_test

import (
	"os"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/scope"
)

func TestWriteItemAndPatterns(t *testing.T) {
	dir := t.TempDir()
	if err := scope.Write(dir, " X-012 ", []string{"app/*", " ", "shared/*"}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(scope.Path(dir))
	if err != nil {
		t.Fatal(err)
	}
	want := "# Scope of the bound work item — written by dudes, enforced by hooks/ (F78)\n" +
		"# item: X-012\napp/*\nshared/*\n"
	if string(data) != want {
		t.Fatalf("got %q", data)
	}
}

func TestWriteNothingRemovesScope(t *testing.T) {
	dir := t.TempDir()
	if err := scope.Write(dir, "X-001", nil); err != nil {
		t.Fatal(err)
	}
	if err := scope.Write(dir, "", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(scope.Path(dir)); !os.IsNotExist(err) {
		t.Fatalf("scope file should be gone, stat err = %v", err)
	}
	if err := scope.Remove(dir); err != nil {
		t.Fatalf("remove of missing file: %v", err)
	}
}

// The engine and the pre-commit hook agree: `*` crosses directories.
func TestInside(t *testing.T) {
	pats := []string{"# item: X-001", "app/internal/config/*", "tests/*", "makefile"}
	for path, want := range map[string]bool{
		"app/internal/config/config.go":     true,
		"app/internal/config/sub/x_test.go": true,
		"tests/product_smoke_test.go":       true,
		"makefile":                          true,
		"go.mod":                            false,
		"app/internal/db/db.go":             false,
		"makefile.bak":                      false,
	} {
		if got := scope.Inside(path, pats); got != want {
			t.Errorf("Inside(%q) = %v, want %v", path, got, want)
		}
	}
	if !scope.Inside("anything", nil) {
		t.Error("no patterns: everything is inside")
	}
}

// A `!` pattern takes files out of the scope: the implementation task's writable is the slice's
// scope minus the tests and hooks the red-first task wrote (F88).
func TestInsideExclusions(t *testing.T) {
	pats := []string{"# item: X-001", "app/internal/store/*", "!app/internal/store/x001_test.go", "!app/internal/store/hook.go", " "}
	for path, want := range map[string]bool{
		"app/internal/store/store.go":     true,
		"app/internal/store/new_file.go":  true,
		"app/internal/store/x001_test.go": false,
		"app/internal/store/hook.go":      false,
		"app/api/api.go":                  false,
	} {
		if got := scope.Inside(path, pats); got != want {
			t.Errorf("Inside(%q) = %v, want %v", path, got, want)
		}
	}
	// only exclusions: everything else is inside
	if !scope.Inside("main.go", []string{"!go.mod"}) || scope.Inside("go.mod", []string{"!go.mod"}) {
		t.Error("only exclusions: everything but them is inside")
	}
	if got := scope.Included(pats); len(got) != 1 || got[0] != "app/internal/store/*" {
		t.Errorf("Included = %v", got)
	}
}

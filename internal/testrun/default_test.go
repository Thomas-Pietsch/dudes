package testrun_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

func TestDefaultForUsesTheRepoTestGoal(t *testing.T) {
	dir := t.TempDir()
	if got := strings.Join(testrun.DefaultFor(dir), " "); got != "go test -v ./..." {
		t.Fatalf("no makefile: got %q", got)
	}
	mk := "build: ## Build\ntest: build ## Run the tests\n"
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), []byte(mk), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(testrun.DefaultFor(dir), " "); got != "make test" {
		t.Fatalf("documented test goal: got %q", got)
	}
}

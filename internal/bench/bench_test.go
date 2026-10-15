package bench_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/bench"
)

func TestPrepare(t *testing.T) {
	dir := t.TempDir()
	if err := bench.Prepare(dir, "hello"); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.mod", "dudes.json", ".git"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

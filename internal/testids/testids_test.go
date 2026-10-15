package testids_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/testids"
)

func TestRunFindsUnnamedAcceptanceTests(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}
	dir := t.TempDir()
	e2e := filepath.Join(dir, "e2e")
	if err := os.MkdirAll(e2e, 0o755); err != nil {
		t.Fatal(err)
	}
	named := "test(\"BK-001 book is saved on the shelf\", async () => {})\n"
	spec := named +
		"test.skip(\"footer is shown on every page\", async () => {})\n" +
		"test.describe(\"BK-002 shelf\", () => {})\n"
	if err := os.WriteFile(filepath.Join(e2e, "shelf.spec.ts"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := testids.Run(dir, "BK-?[0-9]{3}", []string{"e2e"})
	if err != nil {
		t.Fatal(err)
	}
	if res.OK || res.Missing != 1 {
		t.Fatalf("want 1 unnamed acceptance test, got ok=%v missing=%d\n%s", res.OK, res.Missing, res.Output)
	}

	if err := os.WriteFile(filepath.Join(e2e, "shelf.spec.ts"), []byte(named), 0o644); err != nil {
		t.Fatal(err)
	}
	if res, err = testids.Run(dir, "BK-?[0-9]{3}", []string{"e2e"}); err != nil || !res.OK {
		t.Fatalf("named test only: ok=%v err=%v\n%s", res != nil && res.OK, err, res.Output)
	}
}

func TestRunNeedsPatternAndPaths(t *testing.T) {
	if _, err := testids.Run(t.TempDir(), "", []string{"e2e"}); err == nil {
		t.Fatal("want error without ID pattern")
	}
	if _, err := testids.Run(t.TempDir(), "BK-[0-9]+", nil); err == nil {
		t.Fatal("want error without paths")
	}
}

package doctor

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
)

func TestRunProductAndWorkspace(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v (%s)", err, out)
	}
	cfg := config.Default()
	cfg.Workspace = dir
	cfg.Product = "demo"
	checks := Run(cfg)
	if Failed(checks) {
		t.Fatalf("unexpected fail: %+v", checks)
	}
	found := false
	for _, c := range checks {
		if c.Name == "product" && c.Level == OK {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing product ok: %+v", checks)
	}
	_ = os.WriteFile(filepath.Join(dir, "x"), []byte("1"), 0o644)
}

// The freeze line is a verification, not a look at the recorded hash: with the frozen PRD or a frozen
// input changed (committed), doctor must say so, as `dudes check` and the implementation gate do.
func TestCheckFreezeVerifiesTheHashes(t *testing.T) {
	for _, all := range []bool{false, true} {
		name := "PRD only"
		if all {
			name = "PRD and inputs"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			gitIn(t, dir, "init", "-q")
			gitIn(t, dir, "config", "user.email", "dudes@test")
			gitIn(t, dir, "config", "user.name", "dudes-test")
			put := func(rel, body string) {
				t.Helper()
				if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			commit := func(msg string) {
				t.Helper()
				gitIn(t, dir, "add", "demo_PRD.md", "demo_TA.md")
				gitIn(t, dir, "commit", "-q", "-m", msg)
			}

			if c := checkFreeze(dir, "demo"); c.Level != OK || c.Message != "(none)" {
				t.Fatalf("no freeze: want ok (none), got %+v", c)
			}

			put("demo_PRD.md", "# PRD\nv1\n")
			put("demo_TA.md", "# TA\nv1\n")
			commit("spec v1")
			write := freeze.Write
			if all {
				write = freeze.WriteAll
			}
			if _, err := write(dir, "demo", "run-1"); err != nil {
				t.Fatal(err)
			}
			c := checkFreeze(dir, "demo")
			if c.Level != OK || !strings.Contains(c.Message, "verified") {
				t.Fatalf("a freeze that holds: want ok and verified, got %+v", c)
			}
			if all != strings.Contains(c.Message, "inputs") {
				t.Fatalf("the message names the inputs of a freeze with inputs only (all=%v): %+v", all, c)
			}

			put("demo_PRD.md", "# PRD\nv2\n")
			if c := checkFreeze(dir, "demo"); c.Level != OK {
				t.Fatalf("an uncommitted edit is work in progress, not drift: %+v", c)
			}
			commit("spec v2")
			c = checkFreeze(dir, "demo")
			if c.Level != Fail || !strings.Contains(c.Message, "PRD changed since freeze") {
				t.Fatalf("the PRD changed after the freeze: want fail naming it, got %+v", c)
			}
			cfg := config.Default()
			cfg.Workspace, cfg.Product = dir, "demo"
			if !Failed([]Check{c}) {
				t.Fatalf("a freeze that does not hold must fail the doctor: %+v", c)
			}
			var line *Check
			for _, ch := range Run(cfg) {
				if ch.Name == "freeze" {
					line = &ch
				}
			}
			if line == nil || line.Level != Fail {
				t.Fatalf("Run must verify the freeze of the configured product: %+v", line)
			}

			// back to the frozen text: the freeze holds again
			put("demo_PRD.md", "# PRD\nv1\n")
			commit("spec back to v1")
			if c := checkFreeze(dir, "demo"); c.Level != OK {
				t.Fatalf("the frozen text is back: want ok, got %+v", c)
			}
			if all {
				put("demo_TA.md", "# TA\nv2\n")
				commit("TA v2")
				if c := checkFreeze(dir, "demo"); c.Level != Fail || !strings.Contains(c.Message, "demo_TA.md changed since freeze") {
					t.Fatalf("a frozen input changed: want fail naming it, got %+v", c)
				}
			}
		})
	}
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

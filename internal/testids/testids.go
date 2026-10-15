// Package testids runs the acceptance-test ID check (F77, CORE §6 "No unspecified work"): every
// acceptance test (BDD / E2E scenario) names the requirement it proves. The check itself is the
// shell script beside this file — one source for the engine, `make test-ids` and instance CIs.
package testids

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed check-test-ids.sh
var script []byte

// Result of one check.
type Result struct {
	OK      bool   // every acceptance test names an ID
	Missing int    // unnamed acceptance tests
	Output  string // the script's report
}

// Run checks the acceptance tests under paths (relative to dir) against the ID pattern.
func Run(dir, id string, paths []string) (*Result, error) {
	if strings.TrimSpace(id) == "" || len(paths) == 0 {
		return nil, fmt.Errorf("testids: need an ID pattern and at least one acceptance-test path")
	}
	sh, err := exec.LookPath("sh")
	if err != nil {
		return nil, fmt.Errorf("testids: sh not found (Git Bash on Windows): %w", err)
	}
	f, err := os.CreateTemp("", "dudes-test-ids-*.sh")
	if err != nil {
		return nil, fmt.Errorf("testids: %w", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(script); err != nil {
		f.Close()
		return nil, fmt.Errorf("testids: %w", err)
	}
	f.Close()

	// The pattern travels as an environment variable: on Windows the MSYS runtime expands braces
	// in the arguments a non-MSYS program passes to sh (`[0-9]{3}` would arrive as `[0-9]3`).
	args := append([]string{"-c", `exec sh "$0" "$DUDES_TEST_ID" "$@"`, filepath.ToSlash(f.Name())}, paths...)
	cmd := exec.Command(sh, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DUDES_TEST_ID="+id)
	out, err := cmd.CombinedOutput()
	res := &Result{OK: err == nil, Output: string(out)}
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("testids: %w", err)
		}
		for _, l := range strings.Split(res.Output, "\n") {
			if strings.HasPrefix(l, "   ") {
				res.Missing++
			}
		}
	}
	return res, nil
}

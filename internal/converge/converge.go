// Package converge runs re-analysis passes until a pass produces no outcome (F72).
//
// Pass 1 is the work itself. Every further pass is a fresh worker invoke that
// re-reads the task from the beginning, compares it with the current state and
// does only what is still missing. A pass that changes nothing ends the loop:
// the task is done, not just prompted once. The loop is bounded.
package converge

import (
	"errors"
	"fmt"
)

// DefaultMaxPasses applies when convergence is on and no maximum is configured.
const DefaultMaxPasses = 5

// ErrNotConverged means the last allowed pass still produced an outcome.
var ErrNotConverged = errors.New("not converged")

// ErrStop ends the loop early without a verdict (e.g. the worker escalated).
var ErrStop = errors.New("stopped")

// Instruction is appended to the task for pass n ≥ 2.
func Instruction(n int) string {
	return fmt.Sprintf(`

## Convergence pass %d (fresh start)

Re-read the task and the specification from the beginning. Do not rely on what an earlier pass intended.
Compare them with the current repository state. Do only what is still missing or wrong.
If everything is done, change nothing — an unchanged repository ends the loop.
`, n)
}

// Run calls pass(n) for n = 2..max until a pass reports no outcome.
// It returns the number of passes made, counting the caller's pass 1.
func Run(max int, pass func(n int) (outcome bool, err error)) (int, error) {
	if max < 2 {
		return 1, nil
	}
	for n := 2; n <= max; n++ {
		outcome, err := pass(n)
		if err != nil {
			return n, err
		}
		if !outcome {
			return n, nil
		}
	}
	return max, ErrNotConverged
}

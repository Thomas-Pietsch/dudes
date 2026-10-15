package converge

import (
	"errors"
	"strings"
	"testing"
)

func TestRunStopsAtFirstPassWithoutOutcome(t *testing.T) {
	var seen []int
	n, err := Run(5, func(n int) (bool, error) {
		seen = append(seen, n)
		return n < 3, nil // pass 2 changes something, pass 3 does not
	})
	if err != nil || n != 3 || len(seen) != 2 {
		t.Fatalf("n=%d err=%v seen=%v", n, err, seen)
	}
}

func TestRunIsBounded(t *testing.T) {
	n, err := Run(4, func(int) (bool, error) { return true, nil })
	if !errors.Is(err, ErrNotConverged) || n != 4 {
		t.Fatalf("n=%d err=%v", n, err)
	}
}

func TestRunOffOrSinglePass(t *testing.T) {
	called := false
	if n, err := Run(1, func(int) (bool, error) { called = true; return true, nil }); n != 1 || err != nil || called {
		t.Fatalf("max 1 must not re-run: n=%d err=%v", n, err)
	}
}

func TestRunPropagatesStop(t *testing.T) {
	if _, err := Run(5, func(int) (bool, error) { return false, ErrStop }); !errors.Is(err, ErrStop) {
		t.Fatal(err)
	}
}

func TestInstructionStartsFresh(t *testing.T) {
	if s := Instruction(2); !strings.Contains(s, "pass 2") || !strings.Contains(s, "from the beginning") {
		t.Fatal(s)
	}
}

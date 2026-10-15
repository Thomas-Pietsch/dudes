package itempack_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/itempack"
)

const pack = `# Rules for X-001

rule text

# Task X-001 — Skeleton

task text

# Contracts for X-001

> stamp

## PRD

### [[PRD:§2.1]]

prd 2.1

## TA

### [[TA:§3]]

## 3. Packages

package list

### [[TA:§4.2]]

schema text

## SC

### [[SC:M-03]]

mitigation
`

// A task reads the rules, the task and only its references; a quoted heading inside a
// reference does not end it.
func TestNarrow(t *testing.T) {
	got := itempack.Narrow(pack, []string{"TA:§4.2", "[[SC:M-03]]"})
	for _, want := range []string{"rule text", "task text", "# Contracts for X-001", "## TA", "schema text", "## SC", "mitigation"} {
		if !strings.Contains(got, want) {
			t.Errorf("narrowed pack misses %q:\n%s", want, got)
		}
	}
	for _, gone := range []string{"## PRD", "prd 2.1", "[[TA:§3]]", "Packages", "package list"} {
		if strings.Contains(got, gone) {
			t.Errorf("narrowed pack keeps %q:\n%s", gone, got)
		}
	}
	if itempack.Narrow(pack, nil) != pack {
		t.Error("no refs: the pack unchanged")
	}
}

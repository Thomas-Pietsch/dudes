// Package goals reads a repo's goal interface: its self-documented make goals (CORE §13, F75).
// A goal is a makefile target with a `## <what it does>` comment on its line — the same lines
// `make help` lists. The makefile is parsed, not run: no shell, no colours, no side effects.
package goals

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Goal is one documented make target.
type Goal struct {
	Name string
	Desc string
}

// files in GNU make's lookup order.
var files = []string{"GNUmakefile", "makefile", "Makefile"}

// target: deps ## description — `:=` / `::=` assignments are not targets.
var line = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9_.-]*)[ \t]*:([^=#][^#]*)?##[ \t]*(.+)$`)

// List returns the documented goals of the makefile in dir, in file order.
// No makefile → nil, nil.
func List(dir string) ([]Goal, error) {
	for _, name := range files {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("goals: %w", err)
		}
		defer f.Close()
		var out []Goal
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			if m := line.FindStringSubmatch(sc.Text()); m != nil {
				out = append(out, Goal{Name: m[1], Desc: strings.TrimSpace(m[3])})
			}
		}
		if err := sc.Err(); err != nil {
			return nil, fmt.Errorf("goals: read %s: %w", name, err)
		}
		return out, nil
	}
	return nil, nil
}

// Has reports whether dir's makefile documents goal name.
func Has(dir, name string) bool {
	gs, err := List(dir)
	if err != nil {
		return false
	}
	_, ok := find(gs, name)
	return ok
}

// Select returns the named goals in the given order. A name the makefile does not document is an
// error naming every missing goal — the gate stops, the worker never improvises the command.
func Select(gs []Goal, names []string) ([]Goal, error) {
	var out []Goal
	var missing []string
	for _, n := range names {
		n = strings.TrimSpace(n)
		if g, ok := find(gs, n); ok {
			out = append(out, g)
		} else if n != "" {
			missing = append(missing, n)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("goals not documented in the makefile (target with `## description`): %s", strings.Join(missing, ", "))
	}
	return out, nil
}

// Block renders goals as the prompt section a worker gets: its goals and nothing else (CORE §2.4).
func Block(gs []Goal) string {
	if len(gs) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## Goals you may run\n\nUse only these; run them as `make <goal>`:\n\n")
	for _, g := range gs {
		fmt.Fprintf(&b, "- `make %s` — %s\n", g.Name, g.Desc)
	}
	return b.String()
}

func find(gs []Goal, name string) (Goal, bool) {
	for _, g := range gs {
		if g.Name == name {
			return g, true
		}
	}
	return Goal{}, false
}

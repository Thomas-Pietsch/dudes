package handoff

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// The item branch (F88): the engine's commits on it say what they do in the item's terms — the
// slice, the task, the pass — and carry the task as a trailer. When the role approves a task its
// commits land as one, so the branch merges as it is and its code and tests are the engine's commits.

// RoleTag is the engine's role tag on its commits (DUDES §13).
const RoleTag = "[🔹engine]"

// The trailers of a task's commits: the task, the passes an approved task's commit holds, and the
// guard's revert of a pass (F86).
const (
	TaskTrailer     = "Task"
	PassesTrailer   = "Passes"
	RevertedTrailer = "Dudes-Reverted"
)

// Names name the commits of one task of an item-bound contract.
type Names struct {
	Scope string // the slice as people write it, "BK-001a"; the item when the contract is the whole item
	Task  string // the task's ID
	kind  string // "test" for a red-first task, "feat" for the code
	what  string // "red tests for AC-1 and AC-3"
}

// NamesFor names the commits of task taskID of c; ok is false when c binds no item or has no such
// task — the engine's lifecycle commits keep their phase subjects then.
func NamesFor(c *contract.Contract, taskID string) (Names, bool) {
	if c == nil || c.Item == "" {
		return Names{}, false
	}
	tk, ok := c.TaskByID(taskID)
	if !ok {
		return Names{}, false
	}
	n := Names{Scope: SliceName(c), Task: tk.ID, kind: "feat", what: "code"}
	if tk.Expect == "fail" {
		n.kind, n.what = "test", "red tests"
	}
	if crit := Criteria(c); len(crit) > 0 {
		n.what += " for " + listAnd(crit)
	}
	return n, true
}

// Pass is the subject of the task's pass p: "test(BK-001a): red tests for AC-1 and AC-3, pass 2 [🔹engine]".
func (n Names) Pass(p int) string {
	return fmt.Sprintf("%s(%s): %s, pass %d %s", n.kind, n.Scope, n.what, p, RoleTag)
}

// Landed is the subject of the task's approved commit: "feat(BK-001a): code for AC-1 and AC-3 [🔹engine]".
func (n Names) Landed() string {
	return fmt.Sprintf("%s(%s): %s %s", n.kind, n.Scope, n.what, RoleTag)
}

// Revert is the subject of the guard's revert of the task's passes from..to, which broke tests that
// were green before.
func (n Names) Revert(from, to int, broken []string) string {
	passes := fmt.Sprintf("pass %d", to)
	if from < to {
		passes = fmt.Sprintf("passes %d–%d", from, to)
	}
	return fmt.Sprintf("revert(%s): %s of the %s broke %s, green before %s", n.Scope, passes, n.what, strings.Join(broken, ", "), RoleTag)
}

// Trailers are the task's trailers after the item's own (Item, Spec-Source): the task, and passes when
// the commit holds an approved task.
func (n Names) Trailers(item string, passes int) string {
	t := TaskTrailer + ": " + n.Task
	if passes > 0 {
		t += "\n" + PassesTrailer + ": " + strconv.Itoa(passes)
	}
	return joinLines(item, t)
}

// StartSubject is the subject of a slice's first commit, the one its run's notes start on:
// "chore(BK-001a): start the run for AC-1 and AC-3 of BK-001 [🔹engine]".
func StartSubject(c *contract.Contract) string {
	what := "start the run"
	if crit := Criteria(c); len(crit) > 0 {
		what += " for " + listAnd(crit) + " of " + c.Item
	}
	return fmt.Sprintf("chore(%s): %s %s", SliceName(c), what, RoleTag)
}

// StartMessage is the whole message of a slice's first commit: the subject, why the commit is there —
// the run's notes start on it, so they stay off the base other branches share — the pins, and the
// item's trailers.
func StartMessage(c *contract.Contract, pins []Pin, trailers string) string {
	body := "The run's state and its spec freeze are git notes on this commit."
	var at []string
	for _, p := range pins {
		at = append(at, fmt.Sprintf("%s at %s (%s)", p.Root, p.Ref, shortRev(p.Commit)))
	}
	if len(at) > 0 {
		body += " The spec as pinned: " + listAnd(at) + "."
	}
	msg := StartSubject(c) + "\n\n" + body
	if trailers != "" {
		msg += "\n\n" + trailers
	}
	return msg
}

// Criteria are the contract's behaviour criteria as the item writes them: "bk-001-ac-1" → "AC-1".
func Criteria(c *contract.Contract) []string {
	prefix := strings.ToLower(c.Item) + "-"
	var out []string
	for _, a := range c.Acceptance {
		if !strings.EqualFold(a.Kind, "behavior") {
			continue
		}
		id := strings.TrimPrefix(a.ID, prefix)
		if first, _, _ := strings.Cut(strings.TrimSpace(a.Text), " "); strings.EqualFold(first, id) {
			id = first // the item's own spelling, which the projection put first
		} else {
			id = strings.ToUpper(id)
		}
		out = append(out, id)
	}
	return out
}

// SliceName is the slice as people write it: the item and the slice's suffix ("BK-001a" for the contract
// bk-001a of BK-001), else the item.
func SliceName(c *contract.Contract) string {
	if low := strings.ToLower(c.Item); strings.HasPrefix(c.ID, low) {
		return c.Item + c.ID[len(low):]
	}
	return c.Item
}

// TaskRun is the run of commits at the tip of a branch that one task made — its passes, the guard's
// reverts, its earlier approved commit — and the commit they start from.
type TaskRun struct {
	Commits []gitx.Entry // newest first
	Base    string       // the commit before them; "" when the walk found none
	Passes  int          // the passes they hold
}

// TaskCommits reads task's run from a first-parent walk, newest first (gitx.FirstParents): the commits
// whose Task trailer names it, up to the first that does not.
func TaskCommits(log []gitx.Entry, task string) TaskRun {
	var r TaskRun
	if task == "" {
		return r
	}
	for _, e := range log {
		if e.Trailer(TaskTrailer) != task {
			r.Base = e.SHA
			break
		}
		r.Commits = append(r.Commits, e)
		switch n, err := strconv.Atoi(e.Trailer(PassesTrailer)); {
		case err == nil && n > 0:
			r.Passes += n
		case e.Trailer(RevertedTrailer) == "":
			r.Passes++
		}
	}
	return r
}

func shortRev(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func joinLines(a, b string) string {
	switch {
	case a == "":
		return b
	case b == "":
		return a
	}
	return a + "\n" + b
}

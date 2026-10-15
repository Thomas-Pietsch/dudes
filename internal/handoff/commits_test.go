package handoff_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
)

func sliceContract() *contract.Contract {
	return &contract.Contract{ID: "bk-001a", Product: "bookshelf", Item: "BK-001",
		Acceptance: []contract.Acceptance{
			{ID: "bk-001-ac-1", Kind: "behavior", Text: "AC-1 GIVEN an empty shelf …"},
			{ID: "bk-001-ac-3", Kind: "behavior"},
			{ID: "bk-001-field", Kind: "field", Text: "people keep using it"},
		},
		Tasks: []contract.Task{{ID: "bk-001a-tests", Expect: "fail"}, {ID: "bk-001a-code"}},
	}
}

// The engine's commits on an item branch say what they do in the item's terms: the slice, the task,
// its criteria, the pass — with the role tag the hooks ask for (F88).
func TestTaskNames(t *testing.T) {
	c := sliceContract()
	red, ok := handoff.NamesFor(c, "bk-001a-tests")
	if !ok {
		t.Fatal("no names for the red task")
	}
	code, _ := handoff.NamesFor(c, "bk-001a-code")
	for _, x := range []struct{ got, want string }{
		{red.Pass(2), "test(BK-001a): red tests for AC-1 and AC-3, pass 2 [🔹engine]"},
		{red.Landed(), "test(BK-001a): red tests for AC-1 and AC-3 [🔹engine]"},
		{code.Pass(1), "feat(BK-001a): code for AC-1 and AC-3, pass 1 [🔹engine]"},
		{code.Landed(), "feat(BK-001a): code for AC-1 and AC-3 [🔹engine]"},
		{code.Revert(3, 3, []string{"TestBK001AC1"}), "revert(BK-001a): pass 3 of the code for AC-1 and AC-3 broke TestBK001AC1, green before [🔹engine]"},
		{code.Revert(2, 4, []string{"TestBK001AC1", "TestBK001AC3"}), "revert(BK-001a): passes 2–4 of the code for AC-1 and AC-3 broke TestBK001AC1, TestBK001AC3, green before [🔹engine]"},
		{code.Trailers("Item: BK-001\nSpec-Source: P.md abc", 3), "Item: BK-001\nSpec-Source: P.md abc\nTask: bk-001a-code\nPasses: 3"},
		{code.Trailers("", 0), "Task: bk-001a-code"},
		{handoff.StartSubject(c), "chore(BK-001a): start the run for AC-1 and AC-3 of BK-001 [🔹engine]"},
		{handoff.StartMessage(c, []handoff.Pin{{Root: "spec", Ref: "release/bookshelf-v1.0", Commit: "0123456789abcdef"}}, "Item: BK-001"),
			"chore(BK-001a): start the run for AC-1 and AC-3 of BK-001 [🔹engine]\n\n" +
				"The run's state and its spec freeze are git notes on this commit. The spec as pinned: spec at release/bookshelf-v1.0 (0123456).\n\nItem: BK-001"},
	} {
		if x.got != x.want {
			t.Errorf("got  %q\nwant %q", x.got, x.want)
		}
	}
	// the whole item's contract: the item is the scope
	whole := sliceContract()
	whole.ID = "bk-001"
	if n, _ := handoff.NamesFor(whole, "bk-001a-code"); n.Scope != "BK-001" || handoff.SliceName(&contract.Contract{ID: "s", Item: "BK-001"}) != "BK-001" {
		t.Errorf("scope %q", n.Scope)
	}
	for _, c := range []*contract.Contract{nil, {ID: "x", Tasks: []contract.Task{{ID: "t"}}}, sliceContract()} {
		if _, ok := handoff.NamesFor(c, "t"); ok {
			t.Errorf("names for %+v: a lifecycle commit keeps its phase subject", c)
		}
	}
}

// A task's commits are the run at the tip of the branch whose Task trailer names it: its passes, the
// guard's revert (no pass), an earlier approved commit with the passes it holds.
func TestTaskCommits(t *testing.T) {
	e := func(sha string, kv ...string) gitx.Entry {
		x := gitx.Entry{SHA: sha, Trailers: map[string][]string{}}
		for i := 0; i+1 < len(kv); i += 2 {
			x.Trailers[kv[i]] = append(x.Trailers[kv[i]], kv[i+1])
		}
		return x
	}
	log := []gitx.Entry{
		e("c3", "Task", "bk-001a-code"),
		e("r", "Task", "bk-001a-code", "Dudes-Reverted", "c2"),
		e("c2", "Task", "bk-001a-code"),
		e("c1", "Task", "bk-001a-code"),
		e("t", "Task", "bk-001a-tests", "Passes", "2"),
		e("s", "Item", "BK-001"),
	}
	code := handoff.TaskCommits(log, "bk-001a-code")
	if len(code.Commits) != 4 || code.Base != "t" || code.Passes != 3 {
		t.Errorf("code %+v", code)
	}
	if tests := handoff.TaskCommits(log, "bk-001a-tests"); len(tests.Commits) != 0 || tests.Base != "c3" {
		t.Errorf("a task whose commits are not at the tip: %+v", tests)
	}
	if tests := handoff.TaskCommits(log[4:], "bk-001a-tests"); len(tests.Commits) != 1 || tests.Passes != 2 || tests.Base != "s" {
		t.Errorf("an approved task: %+v", tests)
	}
	redo := append([]gitx.Entry{e("c5", "Task", "bk-001a-code"), e("L", "Task", "bk-001a-code", "Passes", "4")}, log[4:]...)
	if r := handoff.TaskCommits(redo, "bk-001a-code"); r.Passes != 5 || len(r.Commits) != 2 {
		t.Errorf("a redo after the approval: %+v", r)
	}
	if r := handoff.TaskCommits(log, ""); len(r.Commits) != 0 {
		t.Errorf("no task: %+v", r)
	}
}

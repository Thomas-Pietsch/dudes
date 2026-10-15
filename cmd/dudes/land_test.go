package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// landWorker stands in for Aider across one slice's review: the red tests, then the same tests once
// more after a redo; the code wrong (its gate red), right, broken after a redo (the guard reverts it,
// F86), right again, and once more after the slice ended. LAND_PASSES counts its invocations.
const landWorker = `#!/bin/sh
while [ $# -gt 0 ]; do
	if [ "$1" = "--message-file" ]; then msg=$2; shift; fi
	shift
done
mkdir -p semver
n=$(cat "$LAND_PASSES" 2>/dev/null || echo 0)
n=$((n + 1))
echo $n >"$LAND_PASSES"
if grep -q 'red first' "$msg"; then
cat > semver/bl002_test.go <<'EOF'
package semver

import "testing"

func TestBL002AC1(t *testing.T) {
	if got, _ := Compare("v1.10.0", "1.9.3"); got != 1 {
		t.Fatalf("Compare = %d, want 1", got)
	}
}

func TestBL002AC3(t *testing.T) {
	if _, err := Parse("1.x.0"); err == nil {
		t.Fatal("no error for 1.x.0")
	}
}
EOF
printf 'package semver\n\nvar hook = func() {}\n' > semver/hook.go
if [ $n -gt 1 ]; then printf '\n// redone after the review\n' >> semver/bl002_test.go; fi
exit 0
fi
case $n in
3) printf 'package semver\n\ntype Version struct{ Major int }\n\nfunc Parse(s string) (Version, error) { return Version{}, nil }\n\nfunc Compare(a, b string) (int, error) { return 0, nil }\n' > semver/semver.go ;;
5) printf 'package semver\n\nfunc Compare(a, b string) (int, error) { return undefined, nil }\n' > semver/semver.go ;;
*) cat > semver/semver.go <<'EOF'
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct{ Major, Minor, Patch int }

func Parse(s string) (Version, error) {
	p := strings.Split(strings.TrimPrefix(s, "v"), ".")
	if len(p) != 3 {
		return Version{}, fmt.Errorf("ERR_VERSION: %q", s)
	}
	var n [3]int
	for i, x := range p {
		v, err := strconv.Atoi(x)
		if err != nil {
			return Version{}, fmt.Errorf("ERR_VERSION: %q", s)
		}
		n[i] = v
	}
	return Version{n[0], n[1], n[2]}, nil
}

func Compare(a, b string) (int, error) {
	x, err := Parse(a)
	if err != nil {
		return 0, err
	}
	y, err := Parse(b)
	if err != nil {
		return 0, err
	}
	for _, d := range []int{x.Major - y.Major, x.Minor - y.Minor, x.Patch - y.Patch} {
		if d != 0 {
			return map[bool]int{true: 1, false: -1}[d > 0], nil
		}
	}
	return 0, nil
}
EOF
;;
esac
if [ $n -ge 6 ]; then printf '\n// redone after the review: %s\n' $n >> semver/semver.go; fi
echo done
`

// The item branch merges as it is (F88). A slice runs under the framework's hooks through what a review
// brings — a redo of the red tests, a failed pass, a redo of the code whose pass the guard reverts —
// and every engine commit names the slice, the task and the pass. Each approval lands its task as one
// commit and keeps its passes under a local ref; the last approval ends the run, and nothing writes into
// the branch after its tasks. A redo opens the slice again, and its second landing keeps both sets of
// passes. git merge --no-ff into development, under the hooks, brings in the product's files and three
// commits: the run's anchor, the tests, the code. The run's notes come along on them: there, the run
// reads as ended — dudes step does nothing, and intake binds the next item.
func TestItemBranchMergesAsItIs(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	f := newHandoffFixture(t)
	hooks, err := filepath.Abs(filepath.Join("..", "..", "hooks"))
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, f.dev, "config", "core.hooksPath", filepath.ToSlash(hooks)) // the framework's hooks, in this fixture only
	gitIn(t, f.dev, "checkout", "-q", "-b", "development")
	base := gitIn(t, f.dev, "rev-parse", "HEAD")

	dir := t.TempDir()
	t.Setenv("LAND_PASSES", filepath.ToSlash(filepath.Join(dir, "passes")))
	script := filepath.Join(dir, "worker.sh")
	if err := os.WriteFile(script, []byte(landWorker), 0o755); err != nil {
		t.Fatal(err)
	}
	user := filepath.Join(f.parent, "worker.json")
	if err := os.WriteFile(user, []byte(`{"aider": {"launch": ["sh", "`+filepath.ToSlash(script)+`"], "timeout_sec": 300}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := prepareSlice(user, f.dev, sliceArgs{slicing: f.slicing(t, "bl-002a", nil)}); err != nil {
		t.Fatal(err)
	}
	wt := f.worktree("bl-002a")
	cfgPath := filepath.Join(wt, ".dudes", "dudes.json")
	step := func(want int) {
		t.Helper()
		if code := cmdStep(cfgPath, false, false, false); code != want {
			t.Fatalf("step: exit %d, want %d", code, want)
		}
	}
	subjects := func(from string) []string {
		return strings.Split(gitIn(t, wt, "log", "--format=%s", from+"..HEAD"), "\n")
	}
	commits := func(from string) []string { return strings.Fields(gitIn(t, wt, "log", "--format=%H", from+"..HEAD")) }
	// kept checks that ref keeps the commits of a landed task: it points at their head, and each is in it
	kept := func(ref string, shas []string) {
		t.Helper()
		if got := gitIn(t, wt, "rev-parse", ref); len(shas) == 0 || got != shas[0] {
			t.Fatalf("%s is %s, want the task's old head %v", ref, got, shas)
		}
		for _, sha := range shas {
			if err := exec.Command("git", "-C", wt, "merge-base", "--is-ancestor", sha, ref).Run(); err != nil {
				t.Errorf("%s does not keep %s", ref, sha)
			}
		}
	}
	ended := func(want bool) {
		t.Helper()
		if st, _, _ := runstate.Load(wt); (st.Status == runstate.StatusEnded) != want {
			t.Fatalf("run status %s, ended=%v wanted", st.Status, want)
		}
	}
	start := gitIn(t, wt, "rev-parse", "HEAD")

	// the red tests, sent back once: two passes, named so; approved, they land as one commit
	step(0)
	if code := cmdRedo(cfgPath, []string{"bl-002a-tests", "name", "the", "version", "in", "the", "failure"}); code != 0 {
		t.Fatalf("redo: exit %d", code)
	}
	step(0)
	red := "test(BL-002a): red tests for AC-1 and AC-3"
	if got := strings.Join(subjects(start), "\n"); got != red+", pass 2 [🔹engine]\n"+red+", pass 1 [🔹engine]" {
		t.Errorf("the red task's commits:\n%s", got)
	}
	redPasses := commits(start)
	if code := cmdDecide(cfgPath, []string{"bl-002a-tests"}); code != 0 {
		t.Fatalf("decide: exit %d", code)
	}
	if got := subjects(start); len(got) != 1 || got[0] != red+" [🔹engine]" {
		t.Fatalf("the approved red task: %q", got)
	}
	kept("refs/dudes/passes/bl-002a/bl-002a-tests/1", redPasses)
	ended(false) // the code task is open
	landed := gitIn(t, wt, "log", "-1", "--format=%B")
	for _, want := range []string{"\nItem: BL-002\n", "\nSpec-Source: BENCH_PRD.md ", "\nTask: bl-002a-tests\n", "\nPasses: 2"} {
		if !strings.Contains(landed, want) {
			t.Errorf("the landed red commit lacks %q:\n%s", want, landed)
		}
	}
	redSHA := gitIn(t, wt, "rev-parse", "HEAD")

	// the code: a red gate, green, a redo whose pass broke the build and was reverted, green again
	step(1)
	step(0)
	if code := cmdRedo(cfgPath, []string{"bl-002a-code", "say", "what", "changed"}); code != 0 {
		t.Fatalf("redo: exit %d", code)
	}
	step(1)
	step(0)
	code := "feat(BL-002a): code for AC-1 and AC-3"
	want := []string{code + ", pass 4 [🔹engine]",
		"revert(BL-002a): pass 3 of the code for AC-1 and AC-3 broke TestBL002AC1, TestBL002AC3, green before [🔹engine]",
		code + ", pass 3 [🔹engine]", code + ", pass 2 [🔹engine]", code + ", pass 1 [🔹engine]"}
	if got := subjects(redSHA); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("the code task's commits:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	codePasses := commits(redSHA)
	for _, sha := range codePasses {
		if msg := gitIn(t, wt, "log", "-1", "--format=%B", sha); !strings.Contains(msg, "\nItem: BL-002") || !strings.Contains(msg, "\nTask: bl-002a-code") {
			t.Errorf("%s lacks the item or the task:\n%s", sha, msg)
		}
	}
	if code := cmdDecide(cfgPath, []string{"bl-002a-code"}); code != 0 {
		t.Fatalf("decide: exit %d", code)
	}
	if got := subjects(start); strings.Join(got, "\n") != code+" [🔹engine]\n"+red+" [🔹engine]" {
		t.Fatalf("the approved slice: %q", got)
	}
	if !strings.Contains(gitIn(t, wt, "log", "-1", "--format=%B"), "\nPasses: 4") {
		t.Error("the landed code commit does not hold its four passes")
	}
	// the passes stay the record, notes and all: the failed first pass still has its gate run
	kept("refs/dudes/passes/bl-002a/bl-002a-code/1", codePasses)
	if note := gitIn(t, wt, "notes", "--ref", testrun.NotesRef, "show", codePasses[len(codePasses)-1]); !strings.Contains(note, "# ok: false") {
		t.Errorf("the failed pass's gate run is gone:\n%s", note)
	}
	ended(true)

	// the slice ends with its tasks: a step writes nothing, the state stays on the tip
	tip := gitIn(t, wt, "rev-parse", "HEAD")
	step(0)
	if gitIn(t, wt, "rev-parse", "HEAD") != tip || gitIn(t, wt, "status", "--porcelain") != "" {
		t.Error("a step after the slice's last task changed the branch")
	}

	// sent back once more: the slice opens again, and the second landing keeps both sets of passes
	if code := cmdRedo(cfgPath, []string{"bl-002a-code", "name", "the", "rule"}); code != 0 {
		t.Fatalf("redo after the end: exit %d", code)
	}
	ended(false)
	step(0)
	redo := commits(redSHA)
	if code := cmdDecide(cfgPath, []string{"bl-002a-code"}); code != 0 {
		t.Fatalf("decide: exit %d", code)
	}
	kept("refs/dudes/passes/bl-002a/bl-002a-code/1", codePasses)
	kept("refs/dudes/passes/bl-002a/bl-002a-code/2", redo)
	tip = gitIn(t, wt, "rev-parse", "HEAD")
	if got := subjects(start); strings.Join(got, "\n") != code+" [🔹engine]\n"+red+" [🔹engine]" || !strings.Contains(gitIn(t, wt, "log", "-1", "--format=%B"), "\nPasses: 5") {
		t.Fatalf("the second landing: %q", got)
	}
	ended(true)
	st, ok, _ := runstate.Load(wt)
	if !ok || st.Phase != phase.Implementation || !st.HasApproved("bl-002a-code") || !st.HasApproved("bl-002a-tests") {
		t.Errorf("run state after the slice %+v", st)
	}
	if acc, err := testrun.Acceptance(wt); err != nil || acc["TestBL002AC1"].Result != testrun.Pass || acc["TestBL002AC1"].Commit != tip {
		t.Errorf("the acceptance run did not follow the landed commit: %v %v", acc, err)
	}
	// the ledger stays where its events happened: every worker call and every answer once
	ledger, err := events.Ledger(wt)
	count := map[string]int{}
	for _, ev := range ledger {
		count[ev.Type]++
	}
	if err != nil || count["worker"] != 7 || count["decide"] != 3 || count["redo"] != 3 {
		t.Errorf("the ledger after three landings: %v %v", count, err)
	}

	// the merge, under the hooks: git's own message has no role tag and is refused
	if out, err := exec.Command("git", "-C", f.dev, "merge", "--no-ff", "feature/bl-002a").CombinedOutput(); err == nil || !strings.Contains(string(out), "Missing role tag") {
		t.Fatalf("a merge without a role tag: %v\n%s", err, out)
	}
	gitIn(t, f.dev, "merge", "--abort")
	gitIn(t, f.dev, "merge", "--no-ff", "-m", "feat(BL-002a): AC-1 and AC-3 of BL-002 [🔹Developer]", "feature/bl-002a")
	if got := strings.Fields(gitIn(t, f.dev, "diff", "--name-only", base, "HEAD")); strings.Join(got, " ") != "semver/bl002_test.go semver/hook.go semver/semver.go" {
		t.Errorf("the merge brought in %v, want the product's files only", got)
	}
	merged := strings.Split(gitIn(t, f.dev, "log", "--format=%s", base+"..HEAD"), "\n")
	if len(merged) != 4 || merged[1] != code+" [🔹engine]" || merged[2] != red+" [🔹engine]" || !strings.HasPrefix(merged[3], "chore(BL-002a): start the run") {
		t.Errorf("development holds:\n%s", strings.Join(merged, "\n"))
	}
	if gitIn(t, f.dev, "status", "--porcelain") != "" {
		t.Error("the merge left changes")
	}
	// the notes came along: the freeze on the anchor, the acceptance run on the code
	if fr, ok, err := freeze.Load(f.dev); err != nil || !ok || fr.Product != "bench" {
		t.Errorf("no freeze in development's history: %+v %v %v", fr, ok, err)
	}
	if acc, err := testrun.Acceptance(f.dev); err != nil || acc["TestBL002AC3"].Result != testrun.Pass {
		t.Errorf("no acceptance run in development's history: %v %v", acc, err)
	}

	// on development the slice's run reads as ended: dudes status says so, a step does nothing, and
	// intake (F84) binds the next item instead of taking the slice for a run under way
	profile, err := os.ReadFile(filepath.Join("..", "..", "bench", "workspace", "dudes", "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	var entries []string
	for _, doc := range []string{"PRD", "TA", "QR", "BACKLOG", "ROADMAP"} {
		entries = append(entries, `{"id": "`+doc+`", "kind": "`+strings.ToLower(doc)+`", "root": "spec", "path": "BENCH_`+doc+`.md"}`)
	}
	for name, body := range map[string]string{
		"profile.json": string(profile),
		"specset.json": `{"product": "bench", "roots": {"spec": "../spec"}, "entries": [` + strings.Join(entries, ", ") + `]}`,
		"dudes.json":   `{"workspace": "` + filepath.ToSlash(f.dev) + `", "product": "bench"}`,
	} {
		if err := os.MkdirAll(filepath.Join(f.dev, ".dudes"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(f.dev, ".dudes", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfgDev := filepath.Join(f.dev, ".dudes", "dudes.json")
	_, out, _ := captured(func() int { return cmdStatus(cfgDev, true) })
	var status struct {
		Run struct {
			RunID  string `json:"run_id"`
			Status string `json:"status"`
		} `json:"run"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil || status.Run.Status != string(runstate.StatusEnded) {
		t.Errorf("dudes status on development: %v\n%s", err, out)
	}
	merge := gitIn(t, f.dev, "rev-parse", "HEAD")
	if code, out, _ := captured(func() int { return cmdStep(cfgDev, false, false, false) }); code != 0 || !strings.Contains(out, "no run is under way") || gitIn(t, f.dev, "rev-parse", "HEAD") != merge {
		t.Errorf("dudes step on development: exit %d\n%s", code, out)
	}
	if code, out, errOut := captured(func() int { return cmdDo(cfgDev, []string{"BL-001"}) }); code != 0 {
		t.Errorf("intake on development: exit %d\n%s%s", code, out, errOut)
	}
	if st, ok, _ := runstate.Load(f.dev); !ok || st.RunID == status.Run.RunID || st.Status != runstate.StatusRunning || st.Phase != phase.TestAuthoring {
		t.Errorf("intake did not start the next item's run: %+v", st)
	}

	// every engine commit passes hooks/commit-msg on development and, item bound, on the item branch
	if err := os.MkdirAll(filepath.Join(wt, ".dudes"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(wt, ".dudes", "writable"), []byte("# item: BL-002\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(filepath.Join(wt, ".dudes", "writable"))
	for _, sha := range strings.Fields(gitIn(t, f.dev, "log", "--format=%H", "--no-merges", base+"..HEAD")) {
		msg := filepath.Join(dir, "msg-"+sha)
		if err := os.WriteFile(msg, []byte(gitIn(t, f.dev, "log", "-1", "--format=%B", sha)+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		for _, where := range []string{f.dev, wt} {
			hook := exec.Command("sh", filepath.Join(hooks, "commit-msg"), msg)
			hook.Dir = where
			if out, err := hook.CombinedOutput(); err != nil {
				t.Errorf("commit-msg refuses %s in %s: %v\n%s", sha, where, err, out)
			}
		}
	}
}

// The bridge (F88): a dev repo whose Makefile includes the framework's make/bridge.mk by its path runs
// a slice through its goals — slice, review, approve both tasks, land — with a worker that stands in
// for Aider. development then holds the product's files and the slice's commits under the role's
// merge, and the slice's worktree, pin and branch are gone; its kept passes stay.
func TestBridgeGoalsLandASlice(t *testing.T) {
	for _, tool := range []string{"make", "sh"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skip("no " + tool)
		}
	}
	f := newHandoffFixture(t)
	framework, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	bin := filepath.Join(dir, "dudes")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	script := filepath.Join(dir, "worker.sh")
	if err := os.WriteFile(script, []byte(fakeWorker), 0o755); err != nil {
		t.Fatal(err)
	}
	user := filepath.Join(dir, "user.json")
	if err := os.WriteFile(user, []byte(`{"aider": {"launch": ["sh", "`+filepath.ToSlash(script)+`"], "timeout_sec": 300}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, f.dev, "config", "core.hooksPath", filepath.ToSlash(filepath.Join(framework, "hooks")))
	gitIn(t, f.dev, "checkout", "-q", "-b", "development")
	if err := os.WriteFile(filepath.Join(f.dev, "Makefile"), []byte("-include "+filepath.ToSlash(framework)+"/make/bridge.mk\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, f.dev, "add", "Makefile")
	gitIn(t, f.dev, "commit", "-qm", "chore(make): the engine's goals [🔹Developer]")
	base := gitIn(t, f.dev, "rev-parse", "HEAD")
	mk := func(goal string, vars ...string) string {
		t.Helper()
		args := append([]string{"-C", f.dev, goal, "DUDES_BIN=" + filepath.ToSlash(bin), "DUDES_CONFIG=" + filepath.ToSlash(user)}, vars...)
		out, err := exec.Command("make", args...).CombinedOutput()
		if err != nil {
			t.Fatalf("make %s: %v\n%s", goal, err, out)
		}
		return string(out)
	}

	mk("slice", "ITEM=BL-002", "SLICING="+filepath.ToSlash(f.slicing(t, "bl-002a", nil)))
	if out := mk("review", "SLICE=bl-002a"); !strings.Contains(out, "bl-002a-tests") || !strings.Contains(out, "semver/bl002_test.go") {
		t.Errorf("make review shows no review packet of the red task:\n%s", out)
	}
	mk("approve", "SLICE=bl-002a", "TASK=bl-002a-tests", "ROLE=QA Engineer")
	mk("approve", "SLICE=bl-002a", "TASK=bl-002a-code", "ROLE=Developer")
	mk("land", "SLICE=bl-002a", "ROLE=Developer")

	subjects := strings.Split(gitIn(t, f.dev, "log", "--format=%s", base+"..HEAD"), "\n")
	want := []string{"feat(BL-002a): AC-1 and AC-3 of BL-002, from feature/bl-002a [🔹Developer]",
		"feat(BL-002a): code for AC-1 and AC-3 [🔹engine]", "test(BL-002a): red tests for AC-1 and AC-3 [🔹engine]"}
	if len(subjects) != 4 || strings.Join(subjects[:3], "\n") != strings.Join(want, "\n") || !strings.HasPrefix(subjects[3], "chore(BL-002a): start the run") {
		t.Errorf("development holds:\n%s", strings.Join(subjects, "\n"))
	}
	if msg := gitIn(t, f.dev, "log", "-1", "--format=%B"); !strings.Contains(msg, "\nItem: BL-002") {
		t.Errorf("the merge does not name the item:\n%s", msg)
	}
	if got := strings.Fields(gitIn(t, f.dev, "diff", "--name-only", base, "HEAD")); strings.Join(got, " ") != "semver/bl002_test.go semver/hook.go semver/semver.go" {
		t.Errorf("the merge brought in %v", got)
	}
	for _, gone := range []string{f.worktree("bl-002a"), filepath.Join(f.spec, ".worktrees", "spec-bl-002a")} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Errorf("%s is still there (%v)", gone, err)
		}
	}
	if b := gitIn(t, f.dev, "branch", "--list", "feature/bl-002a"); b != "" {
		t.Errorf("the slice's branch is still there: %s", b)
	}
	if refs := gitIn(t, f.dev, "for-each-ref", "--format=%(refname)", "refs/dudes/passes/"); strings.Count(refs, "\n")+1 != 2 {
		t.Errorf("kept passes %q, want the two tasks'", refs)
	}
	st, ok, _ := runstate.Load(f.dev)
	if !ok || st.Status != runstate.StatusEnded || st.DecidedBy["bl-002a-tests"].Role != "QA Engineer" || st.DecidedBy["bl-002a-code"].Role != "Developer" {
		t.Errorf("the run on development: %+v", st)
	}
	if gitIn(t, f.dev, "status", "--porcelain") != "" {
		t.Error("the landing left changes")
	}
}

// An approved task's commits that another branch or a remote holds stay as they are: a rewrite would
// fork that history. Once the item branch alone holds them, they land as one commit.
func TestLandTaskLeavesSharedCommits(t *testing.T) {
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q", "-b", "feature/bk-001a")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	if err := os.WriteFile(filepath.Join(ws, ".git", "info", "exclude"), []byte("/.dudes/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := gitx.EmptyCommit(ws, "chore(BK-001a): start the run [🔹engine]\n\nItem: BK-001"); err != nil {
		t.Fatal(err)
	}
	start := gitIn(t, ws, "rev-parse", "HEAD")
	con := &contract.Contract{ID: "bk-001a", Product: "bookshelf", Item: "BK-001", Tests: []string{"x_test.go"},
		Acceptance: []contract.Acceptance{{ID: "bk-001-ac-2", Kind: "behavior", Text: "AC-2 …"}},
		Tasks:      []contract.Task{{ID: "bk-001a-code", Phase: phase.Implementation, Approval: "after"}}}
	if err := contract.Save(ws, con); err != nil {
		t.Fatal(err)
	}
	names, _ := handoff.NamesFor(con, "bk-001a-code")
	for i, body := range []string{"package x\n", "package x // two\n"} {
		if err := os.WriteFile(filepath.Join(ws, "x.go"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := gitx.AddCommit(ws, []string{"x.go"}, names.Pass(i+1)+"\n\nItem: BK-001\nSpec-Source: P_PRD.md abc\nTask: bk-001a-code"); err != nil {
			t.Fatal(err)
		}
	}
	head := gitIn(t, ws, "rev-parse", "HEAD")
	gitIn(t, ws, "branch", "shared")
	cfg := config.Config{Workspace: ws, Product: "bookshelf"}
	st := &runstate.State{RunID: "r", Phase: phase.Implementation, Completed: []string{"bk-001a-code"}, Approved: []string{"bk-001a-code"}}

	landTask(cfg, st, "bk-001a-code")
	if gitIn(t, ws, "rev-parse", "HEAD") != head {
		t.Fatal("commits another branch holds were rewritten")
	}
	gitIn(t, ws, "branch", "-D", "shared")
	landTask(cfg, st, "bk-001a-code")
	msg := gitIn(t, ws, "log", "-1", "--format=%B")
	if gitIn(t, ws, "rev-parse", "HEAD~1") != start || !strings.HasPrefix(msg, "feat(BK-001a): code for AC-2 [🔹engine]\n\n") ||
		!strings.Contains(msg, "Item: BK-001\nSpec-Source: P_PRD.md abc\nTask: bk-001a-code\nPasses: 2") {
		t.Fatalf("not landed as one commit on the start:\n%s", msg)
	}
	landed := gitIn(t, ws, "rev-parse", "HEAD")
	if got := gitIn(t, ws, "rev-parse", "refs/dudes/passes/bk-001a/bk-001a-code/1"); got != head {
		t.Errorf("the passes are kept at %s, want the old head %s", got, head)
	}
	if st.Status != runstate.StatusEnded {
		t.Errorf("the slice's last approval left the run %s", st.Status)
	}
	landTask(cfg, st, "bk-001a-code") // landed already: nothing to do
	if gitIn(t, ws, "rev-parse", "HEAD") != landed {
		t.Error("a landed task was landed again")
	}
	if refs, _ := gitx.Refs(ws, "refs/dudes/passes"); len(refs) != 1 {
		t.Errorf("passes refs %v after a landing that did nothing", refs)
	}
}

// A stub worker's file is no output a task asked for: outside the task's scope the stub writes
// nothing, so it never lands on an item branch — under the hooks a refused commit left it staged.
func TestStubWritesNothingOutsideTheTask(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	f := newHandoffFixture(t)
	hooks, err := filepath.Abs(filepath.Join("..", "..", "hooks"))
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, f.dev, "config", "core.hooksPath", filepath.ToSlash(hooks))
	if _, err := prepareSlice(f.cfg, f.dev, sliceArgs{slicing: f.slicing(t, "bl-002a", nil)}); err != nil {
		t.Fatal(err)
	}
	wt := f.worktree("bl-002a")
	head := gitIn(t, wt, "rev-parse", "HEAD")
	cmdStep(filepath.Join(wt, ".dudes", "dudes.json"), true, false, false) // the gate refuses: no red test
	if _, err := os.Stat(filepath.Join(wt, "cmd", "bench", "main.go")); !os.IsNotExist(err) {
		t.Errorf("the stub wrote outside the task (%v)", err)
	}
	if gitIn(t, wt, "rev-parse", "HEAD") != head || gitIn(t, wt, "status", "--porcelain") != "" {
		t.Error("the stub step left a commit or a change on the item branch")
	}
}

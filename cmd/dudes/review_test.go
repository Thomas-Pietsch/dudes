package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// reviewWorker stands in for Aider in the review loop: it writes what the task asks, as fakeWorker does,
// and prints what Aider prints — its reply with the code in fenced blocks, and its own report of each
// model call's tokens. A redo's feedback changes the tests it writes.
const reviewWorker = `#!/bin/sh
while [ $# -gt 0 ]; do
	if [ "$1" = "--message-file" ]; then msg=$2; shift; fi
	shift
done
mkdir -p semver
echo "Aider v0.86.2"
echo "Model: ollama/m with whole edit format"
if grep -qF 'and **fails**' "$msg"; then
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
if grep -q 'rejection feedback' "$msg"; then
	printf '\n// TestBL002AC3 uses the input AC-3 names.\n' >> semver/bl002_test.go
fi
printf 'package semver\n\nvar hook = func() {}\n' > semver/hook.go
printf 'package semver\n' > semver/stray.go
echo "The tests of AC-1 and AC-3: red until the code exists."
echo
echo "semver/bl002_test.go"
echo '` + "```go" + `'
cat semver/bl002_test.go
echo '` + "```" + `'
echo "Tokens: 2.1k sent, 302 received."
echo "Applied edit to semver/bl002_test.go"
else
cat > semver/semver.go <<'EOF'
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
echo "Parse and Compare, as the tests ask."
echo '` + "```go" + `'
cat semver/semver.go
echo '` + "```" + `'
echo "Tokens: 9.8k sent, 1.2k received."
echo "The edit did not apply; once more."
echo "Tokens: 1.1k sent, 85 received."
echo "Applied edit to semver/semver.go"
fi
`

// sliceRun is the fixture's slice bl-002a, prepared, its code worker a script that stands in for Aider.
type sliceRun struct {
	f       *handoffFixture
	wt, cfg string // the slice's worktree and its run's config
}

func newSliceRun(t *testing.T, worker string) *sliceRun {
	t.Helper()
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	isolateGit(t) // the decider is the repository's identity, whatever this machine's git says
	f := newHandoffFixture(t)
	script := filepath.Join(t.TempDir(), "worker.sh")
	if err := os.WriteFile(script, []byte(worker), 0o755); err != nil {
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
	return &sliceRun{f: f, wt: wt, cfg: filepath.Join(wt, ".dudes", "dudes.json")}
}

// dudes runs a dudes command on the slice's run: its exit code and what it printed.
func (s *sliceRun) dudes(args ...string) (int, string) {
	code, out, errOut := captured(func() int { return run(append([]string{"-config", s.cfg}, args...)) })
	return code, out + errOut
}

// The review loop of a slice (F88), with a worker that stands in for Aider: at each pause after a
// task's run, dudes status carries what the reviewer decides on — the task's diff of its writable paths
// since it first ran, the gate run with the tests that failed or passed, the worker's last reply with
// its code shortened — in text and in -json; a redo and a decision name who answered, in the event and
// in the run state's note; every worker call is an event with Aider's report of its tokens and the time.
// Once the slice's worktree is gone, its ledger — the worker calls, the answers — reads from the repository.
func TestReviewLoop(t *testing.T) {
	s := newSliceRun(t, reviewWorker)
	f, wt := s.f, s.wt
	dudes := s.dudes
	step := func() {
		t.Helper()
		if code, out := dudes("step"); code != 0 {
			t.Fatalf("step: exit %d\n%s", code, out)
		}
	}
	status := func() (string, map[string]any) {
		t.Helper()
		code, text := dudes("status")
		if code != 0 {
			t.Fatalf("status: exit %d\n%s", code, text)
		}
		code, js := dudes("status", "-json")
		var doc map[string]any
		if err := json.Unmarshal([]byte(js), &doc); code != 0 || err != nil {
			t.Fatalf("status -json: exit %d, %v\n%s", code, err, js)
		}
		run, _ := doc["run"].(map[string]any)
		pk, _ := run["review"].(map[string]any)
		return text, pk
	}
	has := func(text string, wants ...string) {
		t.Helper()
		for _, w := range wants {
			if !strings.Contains(text, w) {
				t.Errorf("lacks %q:\n%s", w, text)
			}
		}
	}
	part := func(pk map[string]any, keys ...string) any {
		var v any = pk
		for _, k := range keys {
			m, _ := v.(map[string]any)
			v = m[k]
		}
		return v
	}
	workerEvents := func() []events.Usage {
		t.Helper()
		evs, err := events.ReadAll(wt)
		if err != nil {
			t.Fatal(err)
		}
		var us []events.Usage
		for _, ev := range evs {
			if ev.Type == "worker" && ev.Usage != nil {
				us = append(us, *ev.Usage)
			}
		}
		return us
	}

	// the red-first task ran: its packet shows its tests and hook, the red gate, the reply
	step()
	text, pk := status()
	has(text,
		"review:    bl-002a-tests ran and waits for your answer — dudes -config ",
		"decide bl-002a-tests [-by <role>] | dudes -config ",
		`redo bl-002a-tests [-by <role>] "<feedback>"`,
		"diff:      2 file(s), +",
		"its writable paths: semver/bl002_test.go semver/hook.go",
		"  semver/bl002_test.go +15 -0\n",
		"    +func TestBL002AC1(t *testing.T) {\n",
		"gate:      [go test -count=1 ./semver/...] on ",
		": FAIL (exit 1, ",
		"the task is red first (expect: fail)",
		"  failed: 2 — TestBL002AC1, TestBL002AC3\n",
		"reply:     the worker's last reply, its code blocks shortened (the whole: ",
		"replies/bl-002a-tests.md; the message it got: ",
		"    The tests of AC-1 and AC-3: red until the code exists.\n",
		"    … 15 line(s)\n",
		"    Tokens: 2.1k sent, 302 received.\n",
	)
	if strings.Contains(text, "stray.go") || strings.Contains(text, "missing:") {
		t.Errorf("the packet shows a discarded file, or misses a part:\n%s", text)
	}
	if pk == nil {
		t.Fatal("status -json carries no review")
	}
	if part(pk, "task") != "bl-002a-tests" || part(pk, "gate", "ok") != false || part(pk, "gate", "expect") != "fail" {
		t.Errorf("review %v", pk)
	}
	if files, _ := part(pk, "diff", "files").([]any); len(files) != 2 {
		t.Errorf("diff files %v", files)
	}
	if reply, _ := part(pk, "reply", "text").(string); !strings.Contains(reply, "… 15 line(s)") || strings.Contains(reply, "func TestBL002AC1") {
		t.Errorf("the reply's code is not shortened: %q", reply)
	}
	if diff, _ := part(pk, "diff", "text").(string); !strings.Contains(diff, "+func TestBL002AC3(t *testing.T) {") {
		t.Errorf("the diff lacks the test: %q", diff)
	}
	// the bench reads the run's fields as the first line of that name (scripts/bench-models.sh field):
	// the packet, sorted before them, holds none of them
	field := func(js, key string) string {
		for _, l := range strings.Split(js, "\n") {
			if v, ok := strings.CutPrefix(strings.TrimSpace(l), `"`+key+`": `); ok {
				return strings.Trim(v, `",`)
			}
		}
		return ""
	}
	if _, js := dudes("status", "-json"); field(js, "status") != "waiting_decision" || field(js, "task_id") != "bl-002a-tests" || field(js, "tasks_done") != "1" {
		t.Errorf("the bench would read status %q task_id %q tasks_done %q", field(js, "status"), field(js, "task_id"), field(js, "tasks_done"))
	}
	if us := workerEvents(); len(us) != 1 || us[0] != (events.Usage{Task: "bl-002a-tests", Worker: "aider", Model: "ollama/m", PromptTokens: 2100, CompletionTokens: 302, Counted: "aider", DurationMs: us[0].DurationMs}) {
		t.Errorf("worker events %+v", us)
	}

	// a role sends it back: the event, the output and the run state's note name who
	code, out := dudes("redo", "bl-002a-tests", "-by", "Developer", "AC-3 names the input 1.x.0: say so")
	if code != 0 || !strings.Contains(out, "sent bl-002a-tests back, by Developer (dudes-test <dudes@test>)") {
		t.Fatalf("redo: exit %d\n%s", code, out)
	}
	developer := runstate.Decider{Identity: "dudes-test <dudes@test>", Role: "Developer"}
	if st, _, _ := runstate.Load(wt); st.DecidedBy["bl-002a-tests"] != developer {
		t.Errorf("the run state's decider %+v", st.DecidedBy)
	}
	if note := gitIn(t, wt, "notes", "--ref", runstate.NotesRef, "show", "HEAD"); !strings.Contains(note, `"role": "Developer"`) || !strings.Contains(note, `"identity": "dudes-test <dudes@test>"`) {
		t.Errorf("the note made by redo:\n%s", note)
	}
	evs, _ := events.ReadAll(wt)
	if last := evs[len(evs)-1]; last.Type != "redo" || last.By != "dudes-test <dudes@test>" || last.Role != "Developer" {
		t.Errorf("redo event %+v", last)
	}
	if _, out := dudes("events", "-n", "1"); !strings.Contains(out, "bl-002a-tests AC-3 names the input 1.x.0: say so — by Developer (dudes-test <dudes@test>)") {
		t.Errorf("dudes events:\n%s", out)
	}

	// it runs again: its diff still starts where it first ran, over both of its commits
	step()
	text, _ = status()
	has(text, "review:    bl-002a-tests ran", "  semver/bl002_test.go +17 -0\n", "    +// TestBL002AC3 uses the input AC-3 names.\n")
	if n := strings.Count(text, "  commit "); n != 2 {
		t.Errorf("%d commits listed, want the task's two:\n%s", n, text)
	}

	// approved without a role: the git identity alone
	if code, out := dudes("decide", "bl-002a-tests"); code != 0 || !strings.Contains(out, "decided bl-002a-tests, by dudes-test <dudes@test> — status running") {
		t.Fatalf("decide: exit %d\n%s", code, out)
	}
	if st, _, _ := runstate.Load(wt); st.DecidedBy["bl-002a-tests"] != (runstate.Decider{Identity: "dudes-test <dudes@test>"}) {
		t.Errorf("after the decision: %+v", st.DecidedBy)
	}
	if text, pk := status(); pk != nil || strings.Contains(text, "review:") {
		t.Errorf("a packet while nothing waits for review:\n%s", text)
	}

	// the code task ran: its code alone, the gate green with the slice's acceptance tests passed
	step()
	text, pk = status()
	has(text,
		"review:    bl-002a-code ran and waits for your answer",
		"diff:      1 file(s), +",
		"  semver/semver.go +",
		"gate:      [go test -count=1 ./semver/... ./api/...] on ",
		": PASS (exit 0",
		"  passed: 2 — TestBL002AC1, TestBL002AC3\n",
		"  failed: 0\n",
		"    Parse and Compare, as the tests ask.\n",
	)
	if strings.Contains(text, "tampered") || strings.Contains(text, "red first") {
		t.Errorf("the code task's packet:\n%s", text)
	}
	if passed, _ := part(pk, "gate", "passed").([]any); len(passed) != 2 {
		t.Errorf("passed %v", passed)
	}
	us := workerEvents()
	if len(us) != 3 {
		t.Fatalf("worker events %+v, want one per call", us)
	}
	if code := us[2]; code.Task != "bl-002a-code" || code.PromptTokens != 10900 || code.CompletionTokens != 1285 || code.Counted != "aider" {
		t.Errorf("the code task's call, its two model calls summed: %+v", code)
	}
	if _, out := dudes("events", "-n", "50"); !strings.Contains(out, "bl-002a-code aider: 10900 prompt + 1285 completion tokens (aider), ") {
		t.Errorf("dudes events:\n%s", out)
	}

	// the slice is approved and its worktree goes, then its branch: the ledger stays in the repository
	if code, out := dudes("decide", "bl-002a-code", "-by", "Developer", "the code meets AC-1 and AC-3"); code != 0 {
		t.Fatalf("decide: exit %d\n%s", code, out)
	}
	gitIn(t, f.dev, "worktree", "remove", "--force", wt)
	if _, err := os.Stat(wt); !os.IsNotExist(err) {
		t.Fatalf("the worktree is still there (%v)", err)
	}
	root := filepath.Join(f.parent, "root.json") // dudes run from the dev repository's root
	if err := os.WriteFile(root, []byte(`{"workspace": "`+filepath.ToSlash(f.dev)+`", "product": "bench"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ledger := func() []events.Event {
		t.Helper()
		code, js, errOut := captured(func() int { return run([]string{"-config", root, "events", "-n", "100", "-json"}) })
		var evs []events.Event
		if err := json.Unmarshal([]byte(js), &evs); code != 0 || err != nil {
			t.Fatalf("events -json: exit %d, %v\n%s%s", code, err, js, errOut)
		}
		return evs
	}
	evs = ledger()
	var types []string
	for _, ev := range evs {
		types = append(types, ev.Type)
	}
	if got := strings.Join(types, " "); got != "worker redo worker decide worker decide" {
		t.Fatalf("the ledger from the repository: %s, want each worker call and each answer, in order", got)
	}
	redo, last := evs[1], evs[5]
	if redo.Detail != "bl-002a-tests AC-3 names the input 1.x.0: say so" || redo.By != "dudes-test <dudes@test>" || redo.Role != "Developer" {
		t.Errorf("the redo, with who and why: %+v", redo)
	}
	if u := evs[4].Usage; u == nil || u.Task != "bl-002a-code" || u.PromptTokens != 10900 || u.CompletionTokens != 1285 || u.DurationMs <= 0 {
		t.Errorf("the code task's call: %+v", u)
	}
	if last.Detail != "bl-002a-code the code meets AC-1 and AC-3" || last.Role != "Developer" || evs[3].Role != "" {
		t.Errorf("the decisions: %+v, %+v", evs[3], last)
	}
	_, text, _ = captured(func() int { return run([]string{"-config", root, "events"}) })
	has(text, "events: .dudes/events.jsonl is not here — the ledger in refs/notes/dudes/ledger",
		"bl-002a-tests AC-3 names the input 1.x.0: say so — by Developer (dudes-test <dudes@test>)")
	gitIn(t, f.dev, "branch", "-D", "feature/bl-002a")
	if n := len(ledger()); n != 6 {
		t.Errorf("after the branch went, %d ledger events, want 6", n)
	}
}

// fixWorker stands in for Aider on the red-first task: it stops at a gap in the spec (BLOCKED:) until
// the answer is in its message, then writes the tests.
const fixWorker = `#!/bin/sh
while [ $# -gt 0 ]; do
	if [ "$1" = "--message-file" ]; then msg=$2; shift; fi
	shift
done
if ! grep -q 'Human decision on your escalation' "$msg"; then
	echo "BLOCKED: AC-3 names no input that is not a version"
	exit 0
fi
mkdir -p semver
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
echo "The tests of AC-1 and AC-3, with 1.x.0 as the answer names it."
`

// An answer to a worker's escalation is no approval (F74, F88): the task runs again with the answer,
// then pauses for its review after the run, as its approval: after asks — before, the answer counted as
// the approval, and the task's work went on unreviewed.
func TestEscalationAnswerKeepsTheReview(t *testing.T) {
	s := newSliceRun(t, fixWorker)
	if code, out := s.dudes("step"); code != 0 {
		t.Fatalf("step: exit %d\n%s", code, out)
	}
	st, _, _ := runstate.Load(s.wt)
	if st.Status != runstate.StatusWaitingDecision || st.Escalation == "" || st.TaskID != "bl-002a-tests" {
		t.Fatalf("the worker did not escalate: %+v", st)
	}
	if _, text := s.dudes("status"); strings.Contains(text, "review:") {
		t.Errorf("a review packet for an escalation:\n%s", text)
	}

	code, out := s.dudes("decide", "bl-002a-tests", "-by", "Developer", "1.x.0 is the input that is not a version")
	if code != 0 || !strings.Contains(out, "answered the escalation of bl-002a-tests, by Developer (dudes-test <dudes@test>)") {
		t.Fatalf("decide: exit %d\n%s", code, out)
	}
	st, _, _ = runstate.Load(s.wt)
	if st.HasApproved("bl-002a-tests") || st.Status != runstate.StatusRunning || st.DecidedBy["bl-002a-tests"].Role != "Developer" {
		t.Errorf("after the answer: %+v", st)
	}

	if code, out := s.dudes("step"); code != 0 {
		t.Fatalf("step: exit %d\n%s", code, out)
	}
	if msg, _ := os.ReadFile(filepath.Join(s.wt, ".dudes", "worker_message.md")); !strings.Contains(string(msg), "1.x.0 is the input that is not a version") {
		t.Errorf("the worker did not get the answer:\n%s", msg)
	}
	st, _, _ = runstate.Load(s.wt)
	if st.Status != runstate.StatusWaitingDecision || st.TaskID != "bl-002a-tests" || !st.HasCompleted("bl-002a-tests") || st.HasApproved("bl-002a-tests") || st.Escalation != "" {
		t.Fatalf("the task ran and went on without its review: %+v", st)
	}
	if _, text := s.dudes("status"); !strings.Contains(text, "review:    bl-002a-tests ran and waits for your answer") {
		t.Errorf("no review packet:\n%s", text)
	}
}

// The text worker's call is an event with Ollama's own counts of its tokens, and its reply is kept.
func TestTextWorkerTokens(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"response": "# Idea\n\nA CLI that greets.\n", "done": true, "prompt_eval_count": 412, "eval_count": 37})
	}))
	defer srv.Close()
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	cfgPath := filepath.Join(ws, "dudes.json")
	if err := os.WriteFile(cfgPath, []byte(`{"workspace": "`+filepath.ToSlash(ws)+`", "product": "hello", "ollama": {"base_url": "`+srv.URL+`", "model": "m"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, out, errOut := captured(func() int { return cmdInit(cfgPath, "a CLI that greets") }); code != 0 {
		t.Fatalf("init: %d\n%s%s", code, out, errOut)
	}
	if code, out, errOut := captured(func() int { return cmdStep(cfgPath, false, false, true) }); code != 0 {
		t.Fatalf("step: %d\n%s%s", code, out, errOut)
	}
	evs, err := events.ReadAll(ws)
	if err != nil {
		t.Fatal(err)
	}
	var calls []events.Event
	for _, ev := range evs {
		if ev.Type == "worker" {
			calls = append(calls, ev)
		}
	}
	if len(calls) != 1 || calls[0].Phase != phase.Idea || calls[0].Usage == nil {
		t.Fatalf("worker events %+v", calls)
	}
	if u := *calls[0].Usage; u.Worker != "ollama" || u.Model != "m" || u.PromptTokens != 412 || u.CompletionTokens != 37 || u.Counted != "ollama" {
		t.Errorf("usage %+v", u)
	}
	if !strings.HasPrefix(calls[0].Detail, "ollama: 412 prompt + 37 completion tokens (ollama), ") {
		t.Errorf("detail %q", calls[0].Detail)
	}
	if reply, err := os.ReadFile(filepath.Join(ws, filepath.FromSlash(replyRel("idea")))); err != nil || !strings.Contains(string(reply), "A CLI that greets.") {
		t.Errorf("the reply kept: %q %v", reply, err)
	}
}

// A long diff is cut to its bound, with the git command that prints the whole; paths outside the task's
// patterns and the engine's .dudes/ stay out of it.
func TestTaskDiffBounded(t *testing.T) {
	ws := t.TempDir()
	gitIn(t, ws, "init", "-q")
	gitIn(t, ws, "config", "user.email", "dudes@test")
	gitIn(t, ws, "config", "user.name", "dudes-test")
	gitIn(t, ws, "config", "core.autocrlf", "false")
	gitIn(t, ws, "commit", "-q", "--allow-empty", "-m", "base")
	base := gitIn(t, ws, "rev-parse", "HEAD")
	var big strings.Builder
	for i := range 1000 {
		big.WriteString("// line " + strings.Repeat("x", i%7) + "\n")
	}
	for rel, body := range map[string]string{"pkg/big.go": big.String(), "pkg/x_test.go": "package pkg\n", "other/y.go": "package other\n", ".dudes/state.txt": "engine\n"} {
		p := filepath.Join(ws, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	gitIn(t, ws, "add", "-A")
	gitIn(t, ws, "commit", "-q", "-m", "work")
	head := gitIn(t, ws, "rev-parse", "HEAD")

	d, why := taskDiff(ws, base, head, []string{"pkg/*", "!pkg/x_test.go"})
	if why != "" || d == nil {
		t.Fatalf("taskDiff: %q", why)
	}
	if len(d.Files) != 1 || d.Files[0].Path != "pkg/big.go" || d.Files[0].Added != 1000 {
		t.Errorf("files %+v", d.Files)
	}
	if !d.Cut || strings.Count(d.Text, "\n") != reviewDiffLines || d.Lines != 1006 { // a new file's six header lines
		t.Errorf("cut %v, %d lines shown of %d", d.Cut, strings.Count(d.Text, "\n"), d.Lines)
	}
	if want := gitCmd(ws, "diff", shortSHA(base), shortSHA(head), "--", "pkg/big.go"); d.Full != want {
		t.Errorf("the whole: %q, want %q", d.Full, want)
	}
	var out strings.Builder
	(&reviewPacket{Task: "t", Answer: []string{"approve", "send back"}, Diff: d}).print(&out)
	if want := "  (cut: 400 of 1006 lines shown — the whole: " + d.Full + ")\n"; !strings.Contains(out.String(), want) {
		t.Errorf("lacks %q", want)
	}
	if _, why := taskDiff(ws, "", head, nil); !strings.Contains(why, "did not record where the task started") {
		t.Errorf("no recorded start: %q", why)
	}
}

// -by names the role anywhere after the command; without a value it is refused, not taken as the note.
func TestAnswerArgs(t *testing.T) {
	for _, c := range []struct {
		words            []string
		task, note, role string
		fail             bool
	}{
		{[]string{"t1"}, "t1", "", "", false},
		{[]string{"t1", "tests", "read", "fine"}, "t1", "tests read fine", "", false},
		{[]string{"t1", "-by", "Developer", "a swallowed error"}, "t1", "a swallowed error", "Developer", false},
		{[]string{"-by=QA Engineer", "t1", "the path is relative"}, "t1", "the path is relative", "QA Engineer", false},
		{[]string{"t1", "off by one", "--by", "DevOps"}, "t1", "off by one", "DevOps", false},
		{[]string{"t1", "-by"}, "", "", "", true},
		{[]string{"t1", "-by=", "x"}, "", "", "", true},
	} {
		task, note, role, err := answerArgs(c.words)
		if (err != nil) != c.fail || task != c.task || note != c.note || role != c.role {
			t.Errorf("%q: got %q %q %q %v", c.words, task, note, role, err)
		}
	}
}

// A packet's text is cut at a line's end within its bounds, and says how long the whole is.
func TestClip(t *testing.T) {
	if e := clip("", 3, 100); e.Text != "" || e.Cut || e.Lines != 0 {
		t.Errorf("empty: %+v", e)
	}
	if e := clip("a\r\nb\nc\n", 3, 100); e.Text != "a\nb\nc\n" || e.Cut || e.Lines != 3 {
		t.Errorf("fits: %+v", e)
	}
	if e := clip("a\nb\nc\nd\n", 2, 100); e.Text != "a\nb\n" || !e.Cut || e.Lines != 4 {
		t.Errorf("lines: %+v", e)
	}
	if e := clip("aaaa\nbbbb\ncccc", 10, 11); e.Text != "aaaa\nbbbb\n" || !e.Cut || e.Lines != 3 {
		t.Errorf("bytes: %+v", e)
	}
	if e := clip("ääää\nb", 10, 5); e.Text != "ää\n" || !e.Cut || !utf8.ValidString(e.Text) {
		t.Errorf("a first line over the bound is cut between characters: %+v", e)
	}
}

// A reply's code blocks shrink to their length; the words around them stay.
func TestShortenCode(t *testing.T) {
	reply := "Done.\n\nsemver/semver.go\n```go\npackage semver\n\nfunc A() {}\n```\nApplied edit\n```\nunclosed\n"
	want := "Done.\n\nsemver/semver.go\n```go\n… 3 line(s)\n```\nApplied edit\n```\n… 2 line(s), the block does not end\n"
	if got := shortenCode(reply); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

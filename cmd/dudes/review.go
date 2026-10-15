package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// The review packet (F88): when a task waits for approval after its run, dudes status carries what the
// reviewer decides on — the task's diff of its writable paths since it first ran, the gate run with the
// tests that passed and failed, the worker's last reply — each cut to a bound, with the command or file
// that holds the whole. The first real item's first review cost 128k tokens, most of them spent
// reading these three by hand from git and .dudes/.

// The packet's bounds: what one review reads at once.
const (
	reviewDiffLines, reviewDiffBytes   = 400, 32 << 10
	reviewLogLines, reviewLogBytes     = 60, gateLogMax
	reviewReplyLines, reviewReplyBytes = 60, gateLogMax
	reviewNames                        = 40 // names per test list
	reviewCommits                      = 20
)

// reviewPacket is what a reviewer decides on. Missing says what could not be read, and why.
type reviewPacket struct {
	Task    string       `json:"task"`
	Answer  []string     `json:"answer"` // the two commands that answer: approve, send back
	Diff    *reviewDiff  `json:"diff,omitempty"`
	Gate    *reviewGate  `json:"gate,omitempty"`
	Reply   *reviewReply `json:"reply,omitempty"`
	Missing []string     `json:"missing,omitempty"`
}

// excerpt is a text cut to a bound at a line's end: the head shown, whether it was cut, how many
// lines the whole has, and where the whole is.
type excerpt struct {
	Text  string `json:"text"`
	Cut   bool   `json:"cut,omitempty"`
	Lines int    `json:"lines"`
	Full  string `json:"full,omitempty"`
}

// reviewDiff is the task's change: from the commit it first ran on to HEAD, its writable paths only,
// and the commits that made it, newest first.
type reviewDiff struct {
	Base     string       `json:"base"`
	Head     string       `json:"head"`
	Writable []string     `json:"writable,omitempty"`
	Files    []reviewFile `json:"files"`
	Commits  []string     `json:"commits,omitempty"`
	excerpt
}

type reviewFile struct {
	Path    string `json:"path"`
	Added   int    `json:"added"`
	Deleted int    `json:"deleted"`
	Binary  bool   `json:"binary,omitempty"`
}

// reviewGate is the test run recorded on HEAD: the gate as it ran, its outcome, its tests by result —
// from the gate's log and the acceptance rerun on the same commit — and its log.
type reviewGate struct {
	Command  string     `json:"command"`
	TaskGate string     `json:"task_gate,omitempty"` // the task's own gate, when the run on HEAD was another command
	Commit   string     `json:"commit"`
	OK       bool       `json:"ok"`
	Exit     int        `json:"exit"`
	Duration string     `json:"duration,omitempty"`
	Expect   string     `json:"expect,omitempty"` // "fail": the task is red first
	Passed   []string   `json:"passed"`           // top-level tests, at most reviewNames per list
	Failed   []string   `json:"failed"`
	Skipped  []string   `json:"skipped,omitempty"`
	Counts   testCounts `json:"counts"` // the whole lists' lengths
	excerpt
}

type testCounts struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// reviewReply is the worker's last reply on the task, its code blocks shortened to their length: the
// diff shows the code as committed. Sent is the message the worker got.
type reviewReply struct {
	Sent string `json:"sent,omitempty"`
	excerpt
}

// review is the packet of the task the run waits on after its run; nil when it waits on no such task
// (no pause, a pause before a task, an escalation). configPath goes into the answer commands.
func review(configPath string, cfg config.Config, st *runstate.State) *reviewPacket {
	if st == nil || st.Status != runstate.StatusWaitingDecision || st.Escalation != "" {
		return nil
	}
	con, ok, err := contract.Load(cfg.Workspace)
	if err != nil || !ok {
		return nil
	}
	tk, ok := con.TaskByID(st.TaskID)
	if !ok || !tk.NeedsAfterDecision() || !st.HasCompleted(tk.ID) || st.HasApproved(tk.ID) {
		return nil
	}
	p := &reviewPacket{Task: tk.ID, Answer: []string{
		dudesCmd(configPath, "decide", tk.ID, "[-by <role>]"),
		dudesCmd(configPath, "redo", tk.ID, "[-by <role>]", `"<feedback>"`),
	}}
	head, _ := gitx.HeadFull(cfg.Workspace)
	var why string
	if p.Diff, why = taskDiff(cfg.Workspace, st.TaskBase[tk.ID], head, writable(con, st)); why != "" {
		p.Missing = append(p.Missing, "diff: "+why)
	}
	if p.Gate, why = gateRun(cfg, st, tk, head); why != "" {
		p.Missing = append(p.Missing, "gate: "+why)
	}
	if p.Reply, why = lastReply(cfg.Workspace, tk.ID); why != "" {
		p.Missing = append(p.Missing, "reply: "+why)
	}
	return p
}

// taskDiff is the change from base, the commit the task first ran on, to head, in the task's writable
// paths (pats; none → every path but the engine's .dudes/).
func taskDiff(ws, base, head string, pats []string) (*reviewDiff, string) {
	if base == "" {
		return nil, "the run did not record where the task started (it started before the review packet existed) — see git log"
	}
	if head == "" {
		return nil, "the workspace has no commit"
	}
	changes, err := gitx.Changes(ws, base, head)
	if err != nil {
		return nil, err.Error()
	}
	d := &reviewDiff{Base: base, Head: head, Writable: pats, Files: []reviewFile{}}
	var paths []string
	for _, c := range changes {
		if strings.HasPrefix(c.Path, ".dudes/") || !scope.Inside(c.Path, pats) {
			continue
		}
		d.Files = append(d.Files, reviewFile{Path: c.Path, Added: c.Added, Deleted: c.Deleted, Binary: c.Binary})
		paths = append(paths, c.Path)
	}
	if len(paths) == 0 {
		return d, ""
	}
	text, err := gitx.Diff(ws, base, head, paths)
	if err != nil {
		return d, err.Error()
	}
	d.excerpt = clip(text, reviewDiffLines, reviewDiffBytes)
	d.Full = gitCmd(ws, append([]string{"diff", shortSHA(base), shortSHA(head), "--"}, paths...)...)
	if commits, err := gitx.Subjects(ws, base, head, paths); err == nil {
		d.Commits = commits
		if len(commits) > reviewCommits {
			d.Commits = append(commits[:reviewCommits:reviewCommits], fmt.Sprintf("… %d more: %s", len(commits)-reviewCommits,
				gitCmd(ws, "log", "--oneline", shortSHA(base)+".."+shortSHA(head))))
		}
	}
	return d, ""
}

// gateRun is the test run recorded on HEAD, with the acceptance rerun on the same commit.
func gateRun(cfg config.Config, st *runstate.State, tk contract.Task, head string) (*reviewGate, string) {
	ws := cfg.Workspace
	log, commit, ok, err := testrun.LastRun(ws)
	if err != nil {
		return nil, err.Error()
	}
	if !ok || commit != head {
		return nil, "no test run is recorded on HEAD"
	}
	rec := testrun.ParseRecord(log)
	g := &reviewGate{Command: rec.Args, Commit: commit, OK: rec.OK, Exit: rec.Exit, Duration: rec.Duration, Expect: tk.Expect}
	if d, err := time.ParseDuration(rec.Duration); err == nil {
		g.Duration = d.Round(10 * time.Millisecond).String()
	}
	if cmd, _ := gateCommand(cfg, st); fmt.Sprint(cfg.Launched(cmd)) != rec.Args {
		g.TaskGate = fmt.Sprint(cfg.Launched(cmd))
	}
	results := testrun.Results(rec.Output)
	if c, note, ok, err := gitx.LatestNote(ws, testrun.AcceptanceNotesRef, 1); err == nil && ok && c == commit {
		rank := map[string]int{testrun.Pass: 1, testrun.Skip: 2, testrun.Fail: 3}
		for name, res := range testrun.Results(note) {
			if rank[res] > rank[results[name]] {
				results[name] = res
			}
		}
	}
	for name, res := range results {
		switch {
		case res == testrun.Fail:
			g.Failed = append(g.Failed, name)
		case res == testrun.Skip:
			g.Skipped = append(g.Skipped, name)
		case !strings.Contains(name, "/"): // a passing test's subtests passed with it
			g.Passed = append(g.Passed, name)
		}
	}
	g.Counts = testCounts{len(g.Passed), len(g.Failed), len(g.Skipped)}
	g.Passed, g.Failed, g.Skipped = firstNames(g.Passed), firstNames(g.Failed), firstNames(g.Skipped)
	if g.Passed == nil {
		g.Passed = []string{}
	}
	if g.Failed == nil {
		g.Failed = []string{}
	}
	g.excerpt = clip(rec.Output, reviewLogLines, reviewLogBytes)
	g.Full = gitCmd(ws, "notes", "--ref", testrun.NotesRef, "show", shortSHA(commit))
	return g, ""
}

// firstNames is names sorted, at most reviewNames of them.
func firstNames(names []string) []string {
	sort.Strings(names)
	if len(names) > reviewNames {
		return names[:reviewNames]
	}
	return names
}

// lastReply is the worker's last reply on task, as recordCall kept it.
func lastReply(ws, task string) (*reviewReply, string) {
	p := filepath.Join(ws, filepath.FromSlash(replyRel(task)))
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, "none is kept for " + task + " (a stub worker writes none)"
	}
	r := &reviewReply{excerpt: clip(shortenCode(string(data)), reviewReplyLines, reviewReplyBytes)}
	r.Full = shownPath(p)
	if msg := filepath.Join(ws, ".dudes", "worker_message.md"); fileExists(msg) {
		r.Sent = shownPath(msg)
	}
	return r, ""
}

// shortenCode replaces the lines of each fenced code block in a worker's reply with their count: what
// the reply says besides the code is what the diff cannot show.
func shortenCode(reply string) string {
	var b strings.Builder
	open, n := false, 0
	for _, l := range strings.Split(strings.ReplaceAll(reply, "\r\n", "\n"), "\n") {
		fence := strings.HasPrefix(strings.TrimSpace(l), "```")
		switch {
		case fence && !open:
			open, n = true, 0
			b.WriteString(l + "\n")
		case fence:
			open = false
			fmt.Fprintf(&b, "… %d line(s)\n%s\n", n, l)
		case open:
			n++
		default:
			b.WriteString(l + "\n")
		}
	}
	if open {
		fmt.Fprintf(&b, "… %d line(s), the block does not end\n", n)
	}
	return b.String()
}

// clip cuts text to its first maxLines lines and at most maxBytes bytes, at a line's end; a first line
// longer than maxBytes is cut at a character's end.
func clip(text string, maxLines, maxBytes int) excerpt {
	text = strings.TrimRight(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if text == "" {
		return excerpt{}
	}
	lines := strings.Split(text, "\n")
	e := excerpt{Lines: len(lines)}
	var b strings.Builder
	for i, l := range lines {
		if i == maxLines || b.Len()+len(l)+1 > maxBytes {
			if i == 0 {
				cut := min(len(l), maxBytes)
				for cut > 0 && cut < len(l) && !utf8.RuneStart(l[cut]) {
					cut--
				}
				b.WriteString(l[:cut] + "\n")
			}
			e.Cut = true
			break
		}
		b.WriteString(l + "\n")
	}
	e.Text = b.String()
	return e
}

// print writes the packet as dudes status shows it: a header line per part, its text indented below.
func (p *reviewPacket) print(w io.Writer) {
	fmt.Fprintf(w, "review:    %s ran and waits for your answer — %s | %s\n", p.Task, p.Answer[0], p.Answer[1])
	for _, m := range p.Missing {
		fmt.Fprintf(w, "  missing: %s\n", m)
	}
	if d := p.Diff; d != nil {
		added, deleted := 0, 0
		for _, f := range d.Files {
			added, deleted = added+f.Added, deleted+f.Deleted
		}
		fmt.Fprintf(w, "diff:      %d file(s), +%d -%d since %s, where the task started", len(d.Files), added, deleted, shortSHA(d.Base))
		if len(d.Writable) > 0 {
			fmt.Fprintf(w, " — its writable paths: %s", strings.Join(d.Writable, " "))
		}
		fmt.Fprintln(w)
		for _, f := range d.Files {
			if f.Binary {
				fmt.Fprintf(w, "  %s (binary)\n", f.Path)
			} else {
				fmt.Fprintf(w, "  %s +%d -%d\n", f.Path, f.Added, f.Deleted)
			}
		}
		for _, c := range d.Commits {
			fmt.Fprintf(w, "  commit %s\n", c)
		}
		printExcerpt(w, d.excerpt)
	}
	if g := p.Gate; g != nil {
		outcome := "PASS"
		if !g.OK {
			outcome = "FAIL"
		}
		fmt.Fprintf(w, "gate:      %s on %s: %s (exit %d", g.Command, shortSHA(g.Commit), outcome, g.Exit)
		if g.Duration != "" {
			fmt.Fprintf(w, ", %s", g.Duration)
		}
		fmt.Fprint(w, ")")
		if g.Expect == "fail" {
			fmt.Fprint(w, " — the task is red first (expect: fail)")
		}
		if g.TaskGate != "" {
			fmt.Fprintf(w, " — not the task's gate, %s", g.TaskGate)
		}
		fmt.Fprintln(w)
		printNames(w, "passed", g.Passed, g.Counts.Passed)
		printNames(w, "failed", g.Failed, g.Counts.Failed)
		if g.Counts.Skipped > 0 {
			printNames(w, "skipped", g.Skipped, g.Counts.Skipped)
		}
		printExcerpt(w, g.excerpt)
	}
	if r := p.Reply; r != nil {
		fmt.Fprintf(w, "reply:     the worker's last reply, its code blocks shortened (the whole: %s", r.Full)
		if r.Sent != "" {
			fmt.Fprintf(w, "; the message it got: %s", r.Sent)
		}
		fmt.Fprintln(w, ")")
		printExcerpt(w, r.excerpt)
	}
}

func printNames(w io.Writer, label string, names []string, total int) {
	fmt.Fprintf(w, "  %s: %d", label, total)
	if len(names) > 0 {
		fmt.Fprintf(w, " — %s", strings.Join(names, ", "))
		if total > len(names) {
			fmt.Fprintf(w, ", and %d more", total-len(names))
		}
	}
	fmt.Fprintln(w)
}

// printExcerpt writes e's text indented, and where the whole is when it was cut.
func printExcerpt(w io.Writer, e excerpt) {
	if e.Text != "" {
		for _, l := range strings.Split(strings.TrimSuffix(e.Text, "\n"), "\n") {
			fmt.Fprintf(w, "    %s\n", l)
		}
	}
	if e.Cut {
		fmt.Fprintf(w, "  (cut: %d of %d lines shown — the whole: %s)\n", strings.Count(e.Text, "\n"), e.Lines, e.Full)
	}
}

// dudesCmd is a dudes command line as the user types it, with -config when it is not the default.
func dudesCmd(configPath string, args ...string) string {
	words := []string{"dudes"}
	if configPath != "" && configPath != "dudes.json" {
		words = append(words, "-config", shellWord(shownPath(configPath)))
	}
	return strings.Join(append(words, args...), " ")
}

// gitCmd is a git command line run in ws, as the user types it from the current directory.
func gitCmd(ws string, args ...string) string {
	words := []string{"git"}
	if dir := shownPath(ws); dir != "." {
		words = append(words, "-C", shellWord(dir))
	}
	for _, a := range args {
		words = append(words, shellWord(a))
	}
	return strings.Join(words, " ")
}

// shownPath is p as the user reaches it from the current directory: relative when it is inside it.
func shownPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.ToSlash(p)
	}
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, abs); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(abs)
}

// shellWord is s quoted for a shell when it holds anything but plain path characters.
func shellWord(s string) string {
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("_./:@%+=-", r)) {
			return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
		}
	}
	return s
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

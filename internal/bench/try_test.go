package bench_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// holdEnv turns this test binary into a stand-in for a shell that sits in a folder: started with the
// folder as its working directory it announces itself and waits until its stdin closes. On Windows a
// process in a folder keeps it from being moved or deleted ("Device or resource busy"), as a PowerShell
// or cmd window left in an old tour does.
const holdEnv = "DUDES_TRY_HOLD"

func TestMain(m *testing.M) {
	if os.Getenv(holdEnv) != "" {
		_, _ = os.Stdout.WriteString("holding\n")
		_, _ = io.Copy(io.Discard, os.Stdin)
		return
	}
	os.Exit(m.Run())
}

// The five-minute tour (scripts/try.sh, make try) is the first thing a newcomer runs, so it is held
// to what it promises: a scratch repository of its own for every tour, the whole stub lifecycle with
// no model, one refusal that is the engine's own, a stop at the human gate and a printed seal command
// that works. It never dead-ends (after a tour that failed, beside folders that are not tours, with a
// shell still in an old one) and it never touches the framework's repository: each tour here runs in
// a clone of its own, a git repository with the script in scripts/ and .dudes/ ignored as the real
// one has, and nothing in the clone's refs, notes or commits may change. The script runs as make runs
// it, with a built engine; the engine is the real one, or a stand-in where the test needs one.
func TestTryTour(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not on PATH")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	t.Cleanup(cancel)

	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	bin := filepath.ToSlash(filepath.Join(t.TempDir(), "dudes"+exe))
	if out, err := exec.CommandContext(ctx, "go", "build", "-o", bin, "github.com/Thomas-Pietsch/dudes/cmd/dudes").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "try.sh"))
	if err != nil {
		t.Fatal(err)
	}
	watch(t, script)
	src, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}

	// Stand-in engines: shell scripts the tour starts as it starts the real one.
	// failing is an engine that dies at once, as one killed half-way does.
	failing := writeExe(t, "failing", "echo 'ENGINE RAN'\nexit 1\n")
	// nogit asks git where it is after its repository has vanished, as the engine did in a tour whose
	// `git init` had made nothing, and reports whether it can see a repository above the tour.
	nogit := writeExe(t, "nogit", `rm -rf .git
if top=$(git rev-parse --show-toplevel 2>/dev/null); then echo "FOUND A REPOSITORY ABOVE: $top"; else echo "NO REPOSITORY ABOVE"; fi
exit 1
`)
	// lenient is the real engine, except that it lets every step through.
	lenient := writeExe(t, "lenient", "[ \"$1\" = step ] && exit 0\nexec \""+bin+"\" \"$@\"\n")

	// clone makes what the framework's own checkout is to the tour. Its path holds a space, as a
	// clone under a Windows user folder may.
	clone := func(t *testing.T) string {
		t.Helper()
		dir := filepath.Join(t.TempDir(), "a clone")
		writeFile(t, filepath.Join(dir, "scripts", "try.sh"), string(src))
		writeFile(t, filepath.Join(dir, ".gitignore"), ".dudes/\n")
		git(t, dir, "init", "-q")
		git(t, dir, "config", "user.name", "the clone")
		git(t, dir, "config", "user.email", "clone@example.com")
		git(t, dir, "add", ".gitignore", "scripts/try.sh")
		git(t, dir, "commit", "-q", "-m", "chore: the clone")
		return dir
	}
	// tour runs the script in a clone as `make try` does: the engine in BIN, the folder for the tours
	// the default (.dudes/try) unless one is given.
	tour := func(repo, engine string, env []string, folder ...string) (string, error) {
		cmd := exec.CommandContext(ctx, sh, append([]string{"scripts/try.sh"}, folder...)...)
		cmd.Dir = repo
		cmd.Env = append(append(os.Environ(), "BIN="+engine), env...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	t.Run("one real tour, after a tour that failed, beside what is not a tour", func(t *testing.T) {
		t.Parallel()
		repo := clone(t)
		folder := filepath.Join(repo, ".dudes", "try")
		cloneBefore, treeBefore := history(t, repo), git(t, repo, "status", "--porcelain")

		// A machine that has run the tour before: a file and a folder of the user's own in the
		// tours' folder, and a tour of the earlier layout, a repository in the folder itself.
		keep := filepath.Join(folder, "keep.txt")
		writeFile(t, keep, "mine\n")
		note := filepath.Join(folder, "notes", "todo.md")
		writeFile(t, note, "mine too\n")
		writeFile(t, filepath.Join(folder, "go.mod"), "module try.local\n\ngo 1.22\n")
		git(t, folder, "init", "-q")
		git(t, folder, "config", "user.name", "an earlier tour")
		git(t, folder, "config", "user.email", "try@example.com")
		writeFile(t, filepath.Join(folder, ".git", "dudes-try"), "")
		git(t, folder, "add", "go.mod")
		git(t, folder, "commit", "-q", "-m", "chore: try fixture")
		earlierBefore := history(t, folder)
		// and a tour whose engine died: it leaves its repository behind
		if out, err := tour(repo, failing, nil); err == nil || !strings.Contains(out, "ENGINE RAN") {
			t.Fatalf("the stand-in engine should have run and failed: %v\n%s", err, out)
		}
		partial := tours(t, folder)
		if len(partial) != 1 {
			t.Fatalf("a failed tour leaves one tour behind, got %v", partial)
		}

		decoy := filepath.Join(t.TempDir(), "decoy.git") // a hook's GIT_DIR must not turn the tour's repository into another one
		start := time.Now()
		out, err := tour(repo, bin, []string{"GIT_DIR=" + decoy})
		if err != nil {
			t.Fatalf("the tour failed: %v\n%s", err, out)
		}
		t.Logf("the tour took %s", time.Since(start).Round(100*time.Millisecond))
		now := tours(t, folder)
		if len(now) != 1 {
			t.Fatalf("one tour is left after the next, got %v\n%s", now, out)
		}
		dir := now[0]

		t.Run("the retry after a partial tour goes through", func(t *testing.T) {
			if now[0] == partial[0] {
				t.Errorf("the retry must make a tour of its own, not reuse %s", partial[0])
			}
			if _, err := os.Stat(partial[0]); !os.IsNotExist(err) {
				t.Errorf("the earlier tour should be cleared, whole: %v", err)
			}
			if left, _ := filepath.Glob(filepath.Join(folder, ".gone-*")); len(left) > 0 {
				t.Errorf("a half-deleted tour is left behind: %v", left)
			}
			if strings.Contains(out, "not touching") {
				t.Errorf("the tour refused a folder:\n%s", out)
			}
		})

		t.Run("what is not a tour is left alone", func(t *testing.T) {
			for path, body := range map[string]string{keep: "mine\n", note: "mine too\n"} {
				if got, err := os.ReadFile(path); err != nil || string(got) != body {
					t.Errorf("%s was touched: %q, %v", path, got, err)
				}
			}
			if got := history(t, folder); got != earlierBefore {
				t.Errorf("the repository of an earlier tour was touched:\nbefore %s\nafter  %s", earlierBefore, got)
			}
		})

		t.Run("the clone's git is untouched", func(t *testing.T) {
			if got := history(t, repo); got != cloneBefore {
				t.Errorf("the tour changed the clone's refs or commits:\nbefore %s\nafter  %s", cloneBefore, got)
			}
			if notes := git(t, repo, "for-each-ref", "refs/notes"); notes != "" {
				t.Errorf("the tour wrote notes into the clone:\n%s", notes)
			}
			if got := git(t, repo, "status", "--porcelain"); got != treeBefore {
				t.Errorf("the tour changed the clone's work tree:\n%s", got)
			}
			if _, err := os.Stat(decoy); !os.IsNotExist(err) {
				t.Errorf("GIT_DIR redirected the scratch repository: %v", err)
			}
		})

		t.Run("the engine refuses the edited spec once, and the tour goes on", func(t *testing.T) {
			const refusal = "spec-freeze: PRD changed since freeze"
			if n := strings.Count(out, refusal); n != 1 {
				t.Errorf("the refusal is shown once, got %d times:\n%s", n, out)
			}
			froze, refused, implemented := strings.Index(out, "froze spec hello_PRD.md"), strings.Index(out, refusal), strings.Index(out, "at phase implementation")
			if froze < 0 || refused < froze || implemented < refused {
				t.Errorf("the order is: frozen (%d), refused (%d), then the implementation (%d)", froze, refused, implemented)
			}
			if !strings.Contains(out, "dudes: gate: "+refusal) {
				t.Errorf("the refusal is the engine's own words, `dudes: gate: ...`:\n%s", out)
			}
			// the engine's gate said it, and kept it on record: once
			var failed []string
			data, err := os.ReadFile(filepath.Join(dir, ".dudes", "events.jsonl"))
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(data), "\n") {
				var ev struct{ Type, Phase, Detail string }
				if json.Unmarshal([]byte(line), &ev) == nil && ev.Type == "gate_fail" {
					failed = append(failed, ev.Phase+": "+ev.Detail)
				}
			}
			if len(failed) != 1 || !strings.HasPrefix(failed[0], "implementation: "+refusal) {
				t.Errorf("the engine's event log should hold exactly one gate failure, at the implementation: %q", failed)
			}
			// the record: the edit, its revert, and only then the code
			subjects := strings.Split(strings.TrimSpace(git(t, dir, "log", "--reverse", "--format=%s")), "\n")
			at := func(part string) int {
				for i, s := range subjects {
					if strings.Contains(s, part) {
						return i
					}
				}
				return -1
			}
			spec, tests, edit, revert, impl, qa, docs := at("chore(phase): spec"), at("chore(phase): test_authoring"), at("docs(spec): "), at("Revert "), at("chore(phase): implementation"), at("chore(phase): qa"), at("chore(phase): documentation")
			if !(0 < spec && spec < tests && tests < edit && edit < revert && revert < impl && impl < qa && qa < docs) || docs != len(subjects)-1 {
				t.Errorf("the record should run spec, tests, edit, revert, implementation, qa, documentation:\n%s", strings.Join(subjects, "\n"))
			}
			if strings.Contains(strings.Join(subjects, "\n"), "human") {
				t.Errorf("the dudes must not seal:\n%s", strings.Join(subjects, "\n"))
			}
			if prd := git(t, dir, "show", "HEAD:hello_PRD.md"); strings.Contains(prd, "exclamation") {
				t.Errorf("the edit was not taken back:\n%s", prd)
			}
		})

		t.Run("the printed seal command works", func(t *testing.T) {
			m := regexp.MustCompile(`(?m)^  (cd .+ seal)$`).FindStringSubmatch(out)
			if m == nil {
				t.Fatalf("no seal command printed:\n%s", out)
			}
			if !strings.Contains(m[1], filepath.Base(dir)) {
				t.Errorf("the seal command should name this tour, %s: %s", dir, m[1])
			}
			cmd := exec.CommandContext(ctx, sh, "-c", m[1])
			cmd.Dir = repo
			if b, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("the printed seal command failed: %v\n%s\n%s", err, m[1], b)
			}
			if first, _, _ := strings.Cut(git(t, dir, "log", "--format=%s"), "\n"); !strings.Contains(first, "seal") || !strings.Contains(first, "human") {
				t.Errorf("sealing should add the human's commit, newest commit %q", first)
			}
			if tags := git(t, dir, "tag", "--list", "dudes/seal/*"); strings.TrimSpace(tags) == "" {
				t.Errorf("the run should end in a seal tag")
			}
		})
	})

	t.Run("git finds no repository above the tour, even when the tour's own is gone", func(t *testing.T) {
		t.Parallel()
		repo := clone(t)
		before := history(t, repo)
		out, err := tour(repo, nogit, nil)
		if err == nil {
			t.Errorf("the stand-in engine fails, so does the tour:\n%s", out)
		}
		if !strings.Contains(out, "NO REPOSITORY ABOVE") {
			t.Errorf("git must not climb out of the tour into the clone:\n%s", out)
		}
		if got := history(t, repo); got != before || git(t, repo, "for-each-ref", "refs/notes") != "" {
			t.Errorf("the clone's git changed:\nbefore %s\nafter  %s", before, got)
		}
	})

	t.Run("a repository that git did not make is caught before the engine runs", func(t *testing.T) {
		t.Parallel()
		repo := clone(t)
		realGit, err := exec.LookPath("git")
		if err != nil {
			t.Fatal(err)
		}
		// a git whose `init` claims success and makes nothing
		shim := t.TempDir()
		writeExe(t, "git", "[ \"$1\" = init ] && exit 0\nexec \""+filepath.ToSlash(realGit)+"\" \"$@\"\n", shim)
		out, err := tour(repo, failing, []string{"PATH=" + shim + string(os.PathListSeparator) + os.Getenv("PATH")})
		if err == nil || !strings.Contains(out, "git did not make a repository") {
			t.Errorf("the tour should stop and say so: %v\n%s", err, out)
		}
		if strings.Contains(out, "ENGINE RAN") {
			t.Errorf("the engine ran without a repository:\n%s", out)
		}
		if left := tours(t, filepath.Join(repo, ".dudes", "try")); len(left) > 0 {
			t.Errorf("nothing should be left behind: %v", left)
		}
	})

	t.Run("a gate that does not refuse is not faked", func(t *testing.T) {
		t.Parallel()
		out, err := tour(clone(t), lenient, nil)
		if err == nil {
			t.Errorf("the tour must fail when the engine lets the edited spec through:\n%s", out)
		}
		if strings.Contains(out, "PRD changed since freeze") || strings.Contains(out, "refused (exit") {
			t.Errorf("the tour printed a refusal the engine did not give:\n%s", out)
		}
		if !strings.Contains(out, "let the changed frozen spec through") {
			t.Errorf("the tour should say what went wrong:\n%s", out)
		}
	})

	t.Run("an earlier tour that a shell still sits in does not stop the next one", func(t *testing.T) {
		t.Parallel()
		repo := clone(t)
		folder := filepath.Join(repo, ".dudes", "try")
		if _, err := tour(repo, failing, nil); err == nil {
			t.Fatal("the stand-in engine fails")
		}
		earlier := tours(t, folder)
		if len(earlier) != 1 {
			t.Fatalf("the failed tour is left behind: %v", earlier)
		}
		hold := exec.Command(os.Args[0])
		hold.Dir = earlier[0]
		hold.Env = append(os.Environ(), holdEnv+"=1")
		stdin, err := hold.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		stdout, err := hold.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		if err := hold.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = stdin.Close(); _ = hold.Wait() })
		if line, _ := bufio.NewReader(stdout).ReadString('\n'); line != "holding\n" {
			t.Fatalf("the stand-in shell did not start: %q", line)
		}

		out, err := tour(repo, failing, nil)
		if err == nil || !strings.Contains(out, "ENGINE RAN") {
			t.Errorf("the next tour should get as far as its engine: %v\n%s", err, out)
		}
		if runtime.GOOS == "windows" { // only there does a process in a folder keep it from being moved
			if !strings.Contains(out, "a shell is still in") {
				t.Errorf("a folder in use is named, plainly:\n%s", out)
			}
			for _, name := range []string{"go.mod", filepath.Join(".git", "dudes-try")} {
				if _, err := os.Stat(filepath.Join(earlier[0], name)); err != nil {
					t.Errorf("a folder in use is left whole, not half deleted: %v", err)
				}
			}
		}
	})

	t.Run("a folder of one's own, and a file where a folder should be", func(t *testing.T) {
		t.Parallel()
		repo := clone(t)
		out, err := tour(repo, failing, nil, "tours/")
		if err == nil || !regexp.MustCompile(`scratch repo tours/\d{8}-\d{6} `).MatchString(out) {
			t.Errorf("the tour goes into the folder it was given, named without a double slash: %v\n%s", err, out)
		}
		if got := tours(t, filepath.Join(repo, "tours")); len(got) != 1 {
			t.Errorf("the tour is made in the given folder: %v", got)
		}
		file := filepath.Join(repo, "mine")
		writeFile(t, file, "mine\n")
		out, err = tour(repo, failing, nil, "mine")
		if err == nil || !strings.Contains(out, "not a folder") || strings.Contains(out, "ENGINE RAN") {
			t.Errorf("a file in the way is named, and nothing runs: %v\n%s", err, out)
		}
		if got, _ := os.ReadFile(file); string(got) != "mine\n" {
			t.Errorf("the file was touched: %q", got)
		}
	})
}

// tours lists the tours in a folder: the folders that carry the mark of the script.
func tours(t *testing.T, folder string) []string {
	t.Helper()
	marks, err := filepath.Glob(filepath.Join(folder, "*", ".git", "dudes-try"))
	if err != nil {
		t.Fatal(err)
	}
	var dirs []string
	for _, m := range marks {
		dirs = append(dirs, filepath.Dir(filepath.Dir(m)))
	}
	return dirs
}

// history is what a tour could change in a repository's record: every ref, notes included, and every
// commit reachable from one.
func history(t *testing.T, repo string) string {
	t.Helper()
	return git(t, repo, "for-each-ref") + "--\n" + git(t, repo, "log", "--all", "--format=%H %s")
}

// git runs git in dir and returns its output. A hook's repository variables do not reach it: a test
// must never act on the repository that runs it.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	for _, kv := range os.Environ() {
		if name, _, _ := strings.Cut(kv, "="); !strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			cmd.Env = append(cmd.Env, kv)
		}
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// watch reads what the tour is made of, the script and the Go sources of the engine it builds, so
// that go's test cache counts them as inputs: it cannot see what a subprocess or `go build` reads,
// and an edit would otherwise pass on an old result.
func watch(t *testing.T, script string) {
	t.Helper()
	root := filepath.Join(filepath.Dir(script), "..")
	paths := []string{script, filepath.Join(root, "go.mod")}
	for _, dir := range []string{"cmd", "internal", "skills"} {
		err := filepath.WalkDir(filepath.Join(root, dir), func(p string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(p, ".go") {
				paths = append(paths, p)
			}
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, p := range paths {
		if _, err := os.ReadFile(p); err != nil {
			t.Fatal(err)
		}
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeExe writes a shell script a tour can start as an engine, or as git, and returns its path as a
// shell takes it. The folder is the test's own unless one is given.
func writeExe(t *testing.T, name, body string, folder ...string) string {
	t.Helper()
	dir := t.TempDir()
	if len(folder) > 0 {
		dir = folder[0]
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(path)
}

package bench_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A worker container, and a test the gate runs in one, get the repository read-write (scripts/worker.sh),
// .git included. What they write into the git dir's hooks or config, the engine's next git command runs
// on the host. So the script mounts both read-only over the writable mount. A fake docker that prints
// its arguments, one per line, shows the mounts; the script is the real one.
func TestWorkerGitDirReadOnly(t *testing.T) {
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("sh not on PATH")
	}
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "worker.sh"))
	if err != nil {
		t.Fatal(err)
	}
	watch(t, script)

	fake := t.TempDir()
	writeExe(t, "docker", "printf '%s\n' \"$@\"\n", fake)
	path := fake + string(os.PathListSeparator) + os.Getenv("PATH")

	// run starts the script in dir as the engine does and returns the arguments docker got.
	run := func(t *testing.T, dir string) []string {
		t.Helper()
		cmd := exec.Command(sh, filepath.ToSlash(script), "true")
		cmd.Dir = dir
		for _, kv := range os.Environ() {
			if name, _, _ := strings.Cut(kv, "="); !strings.HasPrefix(strings.ToUpper(name), "GIT_") && !strings.EqualFold(name, "PATH") {
				cmd.Env = append(cmd.Env, kv)
			}
		}
		cmd.Env = append(cmd.Env, "PATH="+path)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("worker.sh in %s: %v\n%s", dir, err, out)
		}
		return strings.Split(strings.TrimSpace(strings.ReplaceAll(string(out), "\r", "")), "\n")
	}
	// shellPath is a folder as the script spells it (Git Bash: /c/…), so the test compares like with like.
	shellPath := func(t *testing.T, p string) string {
		t.Helper()
		out, err := exec.Command(sh, "-c", `cd "$1" && pwd`, "sh", filepath.ToSlash(p)).Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	// mounts are the -v values, in order.
	mounts := func(args []string) []string {
		var m []string
		for i, a := range args {
			if a == "-v" && i+1 < len(args) {
				m = append(m, args[i+1])
			}
		}
		return m
	}
	has := func(args []string, want string) bool {
		for _, m := range mounts(args) {
			if m == want {
				return true
			}
		}
		return false
	}
	readOnly := func(t *testing.T, args []string, gitDir string) {
		t.Helper()
		for _, name := range []string{"hooks", "config"} {
			p := gitDir + "/" + name
			if !has(args, p+":"+p+":ro") {
				t.Errorf("no read-only mount of %s; mounts: %q", p, mounts(args))
			}
		}
	}

	repo := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "init", "-q")
	git(t, repo, "-c", "user.name=t", "-c", "user.email=t@localhost", "commit", "-q", "--allow-empty", "-m", "first")
	gitDir := shellPath(t, filepath.Join(repo, ".git"))

	t.Run("repository", func(t *testing.T) {
		args := run(t, repo)
		readOnly(t, args, gitDir)
		if has(args, gitDir+":"+gitDir) {
			t.Errorf("a plain repository's git dir is mounted a second time: %q", mounts(args))
		}
	})

	t.Run("missing hooks directory is created", func(t *testing.T) {
		hooks := filepath.Join(repo, ".git", "hooks")
		if err := os.RemoveAll(hooks); err != nil {
			t.Fatal(err)
		}
		readOnly(t, run(t, repo), gitDir)
		if fi, err := os.Stat(hooks); err != nil || !fi.IsDir() {
			t.Errorf("hooks directory not created: %v", err)
		}
	})

	t.Run("worktree", func(t *testing.T) {
		wt := filepath.Join(t.TempDir(), "wt")
		git(t, repo, "worktree", "add", "-q", wt, "-b", "side")
		// a worktree with its own config.worktree
		git(t, repo, "config", "extensions.worktreeConfig", "true")
		git(t, wt, "config", "--worktree", "core.autocrlf", "false")
		own := strings.TrimSpace(git(t, wt, "rev-parse", "--absolute-git-dir")) // as git spells it, which the script keeps for this one
		if err := os.RemoveAll(filepath.Join(repo, ".git", "hooks")); err != nil {
			t.Fatal(err)
		}

		args := run(t, wt)
		readOnly(t, args, gitDir) // the common dir's, not the worktree's own
		if !has(args, own+"/config.worktree:"+own+"/config.worktree:ro") {
			t.Errorf("no read-only mount of %s/config.worktree; mounts: %q", own, mounts(args))
		}
		for _, m := range mounts(args) {
			if strings.HasPrefix(m, own+"/hooks") || strings.HasPrefix(m, own+"/config:") {
				t.Errorf("mount of the worktree's own hooks or config, which do not exist: %s", m)
			}
		}
		if !has(args, gitDir+":"+gitDir) {
			t.Errorf("common git dir not mounted: %q", mounts(args))
		}
		if fi, err := os.Stat(filepath.Join(repo, ".git", "hooks")); err != nil || !fi.IsDir() {
			t.Errorf("common hooks directory not created: %v", err)
		}
	})
}

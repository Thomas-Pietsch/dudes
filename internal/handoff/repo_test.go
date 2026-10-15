package handoff_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/handoff"
)

func gitT(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v in %s: %v\n%s", args, dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

func newRepo(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitT(t, dir, "init", "-q", "-b", "main")
	gitT(t, dir, "config", "user.email", "dudes@test")
	gitT(t, dir, "config", "user.name", "dudes-test")
	gitT(t, dir, "config", "core.autocrlf", "false")
}

func writeFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const profileJSON = `{"profile_version":1,"product":"P","owners":[{"pattern":"AC-\\d+","doc":"PRD"}],
 "backlog":{"doc":"BACKLOG","items_section":"## 4.","item_heading":"^#### ([A-Z]+-\\d{3}) · (.*)$","dod_field":"^\\*\\*DoD\\.\\*\\*(.*)$"},
 "status":{"doc":"QR","done":["x"],"blocked":["y"],"open":["z"]},"roadmap":{"doc":"ROADMAP","milestone_row":"^\\| (R\\d)"}}`

// dudes/ is read as a commit holds it: an uncommitted edit is work in progress, not the repository's word.
func TestReadRepo(t *testing.T) {
	dir := t.TempDir()
	newRepo(t, dir)
	writeFile(t, dir, "dudes/profile.json", profileJSON)
	writeFile(t, dir, "dudes/specset.json", `{"product":"p","roots":{"pm":{"repo":"../pm","pin":"release/p-v1.0"}},
 "entries":[{"id":"PRD","kind":"prd","root":"pm","path":"P_PRD.md"},{"id":"X","kind":"other","root":"workspace","path":"x.md"}]}`)
	writeFile(t, dir, "dudes/config.json", `{"product":"p","aider":{"model":"ollama/m"}}`)
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-qm", "engine-ready")
	rev := gitT(t, dir, "rev-parse", "HEAD")
	writeFile(t, dir, "dudes/config.json", `{"product":"edited-after-the-commit"}`)

	r, err := handoff.ReadRepo(dir, rev)
	if err != nil {
		t.Fatal(err)
	}
	if r.SpecSet.Product != "p" || r.SpecSet.Roots["pm"].Repo != "../pm" || r.SpecSet.Roots["pm"].Pin != "release/p-v1.0" || len(r.SpecSet.Entries) != 2 {
		t.Errorf("specset %+v", r.SpecSet)
	}
	if !strings.Contains(string(r.Config), "ollama/m") || strings.Contains(string(r.Config), "edited-after") {
		t.Errorf("config %s", r.Config)
	}
	if names := r.SpecSet.RootNames(); len(names) != 1 || names[0] != "pm" {
		t.Errorf("root names %v", names)
	}

	// the run's spec-set: the roots point at where their pins were checked out
	ws := filepath.Join(dir, ".worktrees", "s")
	set := r.SpecSet.Specset("p", ws, map[string]string{"pm": filepath.Join(dir, "..", "pm", ".worktrees", "spec-s")})
	if set.Product != "p" || set.Roots["pm"] != "../../../pm/.worktrees/spec-s" || len(set.Entries) != 2 {
		t.Errorf("run spec-set %+v", set)
	}

	// config.json is optional
	gitT(t, dir, "rm", "-qf", "dudes/config.json")
	gitT(t, dir, "commit", "-qm", "no config")
	if r, err = handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD")); err != nil || r.Config != nil {
		t.Errorf("a repository without config.json: %+v %v", r, err)
	}
	// the others are not
	gitT(t, dir, "rm", "-q", "dudes/profile.json")
	gitT(t, dir, "commit", "-qm", "no profile")
	if _, err = handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD")); err == nil || !strings.Contains(err.Error(), "not engine-ready") {
		t.Errorf("a repository without profile.json: %v", err)
	}
}

func TestReadRepoRefusesAWrongSpecset(t *testing.T) {
	for name, c := range map[string]struct{ specset, want string }{
		"root without a pin":  {`{"roots":{"pm":{"repo":"../pm"}},"entries":[]}`, "ref it is pinned at"},
		"root as a path":      {`{"roots":{"pm":"../pm"},"entries":[]}`, "specset.json"},
		"workspace as a root": {`{"roots":{"workspace":{"repo":"x","pin":"y"}},"entries":[]}`, "workspace"},
		"undeclared root":     {`{"roots":{},"entries":[{"id":"PRD","kind":"prd","root":"pm","path":"P.md"}]}`, "does not declare"},
	} {
		dir := t.TempDir()
		newRepo(t, dir)
		writeFile(t, dir, "dudes/profile.json", profileJSON)
		writeFile(t, dir, "dudes/specset.json", c.specset)
		gitT(t, dir, "add", "-A")
		gitT(t, dir, "commit", "-qm", "x")
		if _, err := handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD")); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want an error naming %q", name, err, c.want)
		}
	}
}

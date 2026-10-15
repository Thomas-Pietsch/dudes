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

// productProfile is a profile whose backlog owns the ID family family ("BK-\\d{3}"), "" for none.
func productProfile(family string) string {
	owners := `{"pattern":"AC-\\d+","doc":"PRD"}`
	if family != "" {
		owners += `,{"pattern":"` + family + `","doc":"BACKLOG"}`
	}
	return strings.Replace(profileJSON, `[{"pattern":"AC-\\d+","doc":"PRD"}]`, "["+owners+"]", 1)
}

// dudes/ is read as a commit holds it: an uncommitted edit is work in progress, not the repository's word.
func TestReadRepo(t *testing.T) {
	dir := t.TempDir()
	newRepo(t, dir)
	writeFile(t, dir, "dudes/profile.json", profileJSON)
	writeFile(t, dir, "dudes/specset.json", `{"product":"p","roots":{"pm":{"repo":"../pm","pin":"release/p-v1.0"}},
 "entries":[{"id":"PRD","kind":"prd","root":"pm","path":"P_PRD.md"},{"id":"X","kind":"other","root":"workspace","path":"x.md"}]}`)
	writeFile(t, dir, "dudes/config.json", `{"product":"p","aider":{"model":"ollama/m"},"test_command":["go","test","./..."]}`)
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-qm", "engine-ready")
	rev := gitT(t, dir, "rev-parse", "HEAD")
	writeFile(t, dir, "dudes/config.json", `{"product":"edited-after-the-commit"}`)

	r, err := handoff.ReadRepo(dir, rev, "P-001") // one product: every item is its
	if err != nil {
		t.Fatal(err)
	}
	if r.SpecSet.Product != "p" || r.SpecSet.Roots["pm"].Repo != "../pm" || r.SpecSet.Roots["pm"].Pin != "release/p-v1.0" || len(r.SpecSet.Entries) != 2 {
		t.Errorf("specset %+v", r.SpecSet)
	}
	if len(r.Configs) != 1 || r.Configs[0].Path != "dudes/config.json" || !strings.Contains(string(r.Configs[0].Data), "ollama/m") || strings.Contains(string(r.Configs[0].Data), "edited-after") {
		t.Errorf("configs %+v", r.Configs)
	}
	if r.Product != "" || len(r.Products) != 0 || r.ProductName() != "p" || r.Where() != "dudes" || strings.Join(r.TestCommand(), " ") != "go test ./..." {
		t.Errorf("one product: %q %v name %q at %s test %v", r.Product, r.Products, r.ProductName(), r.Where(), r.TestCommand())
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
	if r, err = handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD"), "P-001"); err != nil || len(r.Configs) != 0 || r.ProductName() != "p" {
		t.Errorf("a repository without config.json: %+v %v", r, err)
	}
	// the others are not
	gitT(t, dir, "rm", "-q", "dudes/profile.json")
	gitT(t, dir, "commit", "-qm", "no profile")
	if _, err = handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD"), "P-001"); err == nil || !strings.Contains(err.Error(), "not engine-ready") {
		t.Errorf("a repository without profile.json: %v", err)
	}
	gitT(t, dir, "rm", "-qr", "dudes")
	gitT(t, dir, "commit", "-qm", "no dudes/")
	if _, err = handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD"), "P-001"); err == nil || !strings.Contains(err.Error(), "not engine-ready") {
		t.Errorf("a repository without dudes/: %v", err)
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
		if _, err := handoff.ReadRepo(dir, gitT(t, dir, "rev-parse", "HEAD"), "P-001"); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want an error naming %q", name, err, c.want)
		}
	}
}

// severalProducts is a dev repository of three products, each with its folder under dudes/ beside the
// shared settings: bookshelf (BK-nnn, with its own settings), tracker (TR-nnn) and notes, whose profile
// names no ID family for its backlog. It returns the repository and its commit.
func severalProducts(t *testing.T, mutate func(dir string)) (string, string) {
	t.Helper()
	dir := t.TempDir()
	newRepo(t, dir)
	specset := func(p string) string {
		return `{"roots":{"spec":{"repo":"../spec","pin":"release/` + p + `-v1.0"}},"entries":[{"id":"PRD","kind":"prd","root":"spec","path":"` + strings.ToUpper(p) + `_PRD.md"}]}`
	}
	writeFile(t, dir, "dudes/config.json", `{"aider":{"model":"ollama/m"},"test_command":["go","test","./..."]}`)
	writeFile(t, dir, "dudes/bookshelf/profile.json", productProfile(`BK-\\d{3}`))
	writeFile(t, dir, "dudes/bookshelf/specset.json", specset("bookshelf"))
	writeFile(t, dir, "dudes/bookshelf/config.json", `{"product":"bookshelf","test_command":["go","test","./bookshelf/..."]}`)
	writeFile(t, dir, "dudes/tracker/profile.json", productProfile(`TR-\\d{3}`))
	writeFile(t, dir, "dudes/tracker/specset.json", specset("tracker"))
	writeFile(t, dir, "dudes/notes/profile.json", productProfile(""))
	writeFile(t, dir, "dudes/notes/specset.json", specset("notes"))
	writeFile(t, dir, "dudes/README.md", "what the folders are\n")
	if mutate != nil {
		mutate(dir)
	}
	gitT(t, dir, "add", "-A")
	gitT(t, dir, "commit", "-qm", "three products")
	return dir, gitT(t, dir, "rev-parse", "HEAD")
}

// A dev repository of several products keeps a folder per product: the item's ID finds its product,
// whose profile, spec-set (pinned at the product's own release) and settings over the shared ones
// the slice gets (F88).
func TestReadRepoSeveralProducts(t *testing.T) {
	dir, rev := severalProducts(t, nil)
	r, err := handoff.ReadRepo(dir, rev, "bk-001")
	if err != nil {
		t.Fatal(err)
	}
	if r.Product != "bookshelf" || r.ProductName() != "bookshelf" || r.Where() != "dudes/bookshelf" || strings.Join(r.Products, " ") != "bookshelf notes tracker" {
		t.Errorf("product %q (%q at %s) of %v", r.Product, r.ProductName(), r.Where(), r.Products)
	}
	if !strings.Contains(string(r.Profile), `BK-\\d{3}`) || r.SpecSet.Roots["spec"].Pin != "release/bookshelf-v1.0" {
		t.Errorf("profile %s spec-set %+v", r.Profile, r.SpecSet)
	}
	if len(r.Configs) != 2 || r.Configs[0].Path != "dudes/config.json" || r.Configs[1].Path != "dudes/bookshelf/config.json" {
		t.Errorf("configs %+v: the shared one, then the product's", r.Configs)
	}
	if got := strings.Join(r.TestCommand(), " "); got != "go test ./bookshelf/..." {
		t.Errorf("test command %q: the product's own over the shared one", got)
	}

	r, err = handoff.ReadRepo(dir, rev, "TR-002")
	if err != nil {
		t.Fatal(err)
	}
	if r.Product != "tracker" || r.SpecSet.Roots["spec"].Pin != "release/tracker-v1.0" || len(r.Configs) != 1 || strings.Join(r.TestCommand(), " ") != "go test ./..." {
		t.Errorf("tracker: %q %+v %+v %v", r.Product, r.SpecSet, r.Configs, r.TestCommand())
	}

	// an item no product claims names the products the repository knows
	_, err = handoff.ReadRepo(dir, rev, "ZZ-001")
	for _, want := range []string{"ZZ-001", `bookshelf (BK-\d{3})`, `tracker (TR-\d{3})`, "notes (no owner gives its backlog its IDs)"} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("an unknown item: %v, want %q", err, want)
		}
	}
}

// What cannot be one product's files is refused, naming what is wrong: an item two products claim,
// both layouts at once, a product folder without its profile or with a name that is no product, and a
// file that names another product than its folder.
func TestReadRepoSeveralProductsRefusals(t *testing.T) {
	for name, c := range map[string]struct {
		mutate func(dir string)
		item   string
		want   []string
	}{
		"an item two products claim": {func(dir string) {
			writeFile(t, dir, "dudes/shelf2/profile.json", productProfile(`BK-\\d{3}`))
			writeFile(t, dir, "dudes/shelf2/specset.json", `{"roots":{},"entries":[]}`)
		}, "BK-001", []string{"BK-001 fits the work items of bookshelf and shelf2", "tracker (TR-"}},
		"both layouts": {func(dir string) {
			writeFile(t, dir, "dudes/profile.json", profileJSON)
		}, "BK-001", []string{"one layout", "bookshelf, notes, tracker"}},
		"a folder without its profile": {func(dir string) {
			writeFile(t, dir, "dudes/later/specset.json", `{"roots":{},"entries":[]}`)
		}, "BK-001", []string{"dudes/later/profile.json", "not engine-ready"}},
		"a folder that is no product's name": {func(dir string) {
			writeFile(t, dir, "dudes/Big Shelf/profile.json", productProfile(`BS-\\d{3}`))
		}, "BK-001", []string{"dudes/Big Shelf/", "slug"}},
		"shared settings that name a product": {func(dir string) {
			writeFile(t, dir, "dudes/config.json", `{"product":"bookshelf"}`)
		}, "TR-001", []string{"dudes/config.json sets product bookshelf", "the folder is the product"}},
		"a spec-set that names another product": {func(dir string) {
			writeFile(t, dir, "dudes/tracker/specset.json", `{"product":"bookshelf","roots":{},"entries":[]}`)
		}, "TR-001", []string{"dudes/tracker/specset.json names product bookshelf"}},
	} {
		dir, rev := severalProducts(t, c.mutate)
		_, err := handoff.ReadRepo(dir, rev, c.item)
		for _, want := range c.want {
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Errorf("%s: %v, want %q", name, err, want)
			}
		}
	}
}

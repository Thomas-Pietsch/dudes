package gitx_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
)

// What changed between two commits, limited to the paths given (F88): the files with their line
// counts, the patch, and the commits that touched them. A path with a pathspec character in its name
// is taken literally.
func TestChangesDiffSubjects(t *testing.T) {
	dir, base := repoWithCommit(t)
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("semver/semver.go", "package semver\n\nfunc Compare() int { return 1 }\n")
	write("semver/[ab].txt", "a file named with brackets\n")
	write("semver/a.txt", "a file the brackets match as a pattern\n")
	gitOut(t, dir, "add", "-A")
	gitOut(t, dir, "commit", "-qm", "the code")
	write("api/api.go", "package api\n")
	if err := os.WriteFile(filepath.Join(dir, "logo.bin"), []byte{0, 1, 2, 0, 255}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dir, "dudes", "config.json")); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "add", "-A")
	gitOut(t, dir, "commit", "-qm", "elsewhere")
	head := gitOut(t, dir, "rev-parse", "HEAD")

	changes, err := gitx.Changes(dir, base, head)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]gitx.FileChange{}
	for _, c := range changes {
		got[c.Path] = c
	}
	if len(got) != 6 || got["semver/semver.go"].Added != 3 || got["dudes/config.json"].Deleted != 1 || !got["logo.bin"].Binary || got["semver/[ab].txt"].Added != 1 {
		t.Errorf("changes %+v", changes)
	}

	diff, err := gitx.Diff(dir, base, head, []string{"semver/[ab].txt"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(diff, "+a file named with brackets") || strings.Contains(diff, "as a pattern") || strings.Contains(diff, "semver.go") {
		t.Errorf("the brackets are a name, not a pattern:\n%s", diff)
	}
	subjects, err := gitx.Subjects(dir, base, head, []string{"semver/semver.go"})
	if err != nil || len(subjects) != 1 || !strings.HasSuffix(subjects[0], " the code") {
		t.Errorf("subjects %q %v", subjects, err)
	}
	if all, _ := gitx.Subjects(dir, base, head, nil); len(all) != 2 || !strings.HasSuffix(all[0], " elsewhere") {
		t.Errorf("every commit, newest first: %q", all)
	}
}

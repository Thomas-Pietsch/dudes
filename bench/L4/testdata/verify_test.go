// Hidden verification of L4 (BL-004, TA ALG-03 on top of ALG-01 and ALG-02): the worker never sees
// this file. It checks what the item's criteria imply but its tests need not spell out: the committed
// golden file byte for byte, exactly one trailing newline, a `v` prefix kept as given, a candidate's
// placement, an empty notes file, the error's name and line, the contract across packages, and that
// semver and release kept ALG-01 and ALG-02 (the L2 and L3 cases).
package verify_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"bench/notes"
	"bench/release"
	"bench/semver"
)

// The committed input renders to the committed golden file, byte for byte (the item's AC-1).
func TestGoldenVerify(t *testing.T) {
	in, err := os.ReadFile("../testdata/notes.txt")
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../testdata/release-notes.md")
	if err != nil {
		t.Fatal(err)
	}
	ns, err := notes.Parse("notes.txt", strings.NewReader(string(in)))
	if err != nil {
		t.Fatalf("Parse(notes.txt): %v", err)
	}
	got, err := notes.Render(ns)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if got != string(want) {
		t.Errorf("Render differs from testdata/release-notes.md\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

// Render on inputs the item's tests do not name: the exact bytes, one trailing newline (ALG-03.3 to ALG-03.5).
func TestRenderVerify(t *testing.T) {
	note := func(tag, text string) notes.Note { return notes.Note{Tag: tag, Text: text} }
	for _, c := range []struct {
		name  string
		notes []notes.Note
		want  string
	}{
		{"an empty notes file", nil, "# Release notes\n"},
		{"a v prefix kept as given, numbers not strings", []notes.Note{note("1.2.0", "Patch"), note("v1.10.0", "Minor"), note("v1.9.3", "Fix")},
			"# Release notes\n\nLatest release: v1.10.0\n\n## v1.10.0\n\n- Minor\n\n## v1.9.3\n\n- Fix\n\n## 1.2.0\n\n- Patch\n"},
		{"a candidate above the finals, the latest line names the final", []notes.Note{note("v1.5.0", "Final"), note("v2.0.0-rc.1", "Candidate")},
			"# Release notes\n\nLatest release: v1.5.0\n\n## v2.0.0-rc.1 (candidate)\n\n- Candidate\n\n## v1.5.0\n\n- Final\n"},
		{"only candidates, by number", []notes.Note{note("1.2.0-rc.9", "Nine"), note("1.2.0-rc.10", "Ten")},
			"# Release notes\n\nLatest release: 1.2.0-rc.10\n\n## 1.2.0-rc.10 (candidate)\n\n- Ten\n\n## 1.2.0-rc.9 (candidate)\n\n- Nine\n"},
		{"the notes of one tag in input order", []notes.Note{note("v1.0.0", "First"), note("v1.1.0", "Other"), note("v1.0.0", "Second")},
			"# Release notes\n\nLatest release: v1.1.0\n\n## v1.1.0\n\n- Other\n\n## v1.0.0\n\n- First\n- Second\n"},
	} {
		got, err := notes.Render(c.notes)
		if err != nil {
			t.Errorf("%s: Render: %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: Render =\n%q\nwant\n%q", c.name, got, c.want)
		}
		if !strings.HasSuffix(got, "\n") || strings.HasSuffix(got, "\n\n") {
			t.Errorf("%s: the file ends with %q, want exactly one newline", c.name, got[max(0, len(got)-2):])
		}
	}
}

// Parse refuses with the input's name and the line (ALG-03.2); empty lines count and are skipped (ALG-03.1).
func TestParseVerify(t *testing.T) {
	for _, c := range []struct{ name, in, prefix string }{
		{"notes.txt", "v1.0.0 First\n1.x.0 Broken\n", "ERR_NOTES: notes.txt:2:"}, // the item's AC-2
		{"notes.txt", "v1.0.0 First\n\nv1.1.0\n", "ERR_NOTES: notes.txt:3:"},     // no text; the empty line counts
		{"CHANGES", "nightly build\n", "ERR_NOTES: CHANGES:1:"},                  // the name as given
	} {
		_, err := notes.Parse(c.name, strings.NewReader(c.in))
		if err == nil || !strings.HasPrefix(err.Error(), c.prefix) {
			t.Errorf("Parse(%q, %q) error %v, want the prefix %q", c.name, c.in, err, c.prefix)
		}
	}
	for _, c := range []struct {
		in   string
		want []notes.Note
	}{
		{"", nil},     // an empty notes file
		{"\n\n", nil}, // only empty lines
		{"v1.0.0 Last line without a newline", []notes.Note{{Tag: "v1.0.0", Text: "Last line without a newline"}}},
		{"\nv1.0.0  Blanks around the text \n\n1.0.0-rc.1 Candidate\n", []notes.Note{{Tag: "v1.0.0", Text: "Blanks around the text"}, {Tag: "1.0.0-rc.1", Text: "Candidate"}}},
	} {
		got, err := notes.Parse("notes.txt", strings.NewReader(c.in))
		if err != nil {
			t.Errorf("Parse(%q): %v", c.in, err)
			continue
		}
		if len(got) != len(c.want) {
			t.Errorf("Parse(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("Parse(%q)[%d] = %v, want %v", c.in, i, got[i], c.want[i])
			}
		}
	}
}

// notes orders and classifies through semver and release (ALG-03.6): the package imports both.
func TestOneParserVerify(t *testing.T) {
	entries, err := os.ReadDir("../notes")
	if err != nil {
		t.Fatal(err)
	}
	imports := map[string]bool{}
	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join("../notes", e.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range f.Imports {
			imports[strings.Trim(imp.Path.Value, `"`)] = true
		}
	}
	for _, want := range []string{"bench/semver", "bench/release"} {
		if !imports[want] {
			t.Errorf("package notes does not import %s (ALG-03.6: one parser)", want)
		}
	}
}

// semver keeps ALG-01 (the L2 cases and L3's IsCandidate): the item did not touch it.
func TestSemverKeptVerify(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"v1.10.0", "1.9.3", 1}, {"1.2.3", "1.2.10", -1}, {"2.0.0", "10.0.0", -1}, {"1.2.0-rc.2", "1.2.0-rc.10", -1},
		{"1.2.0-rc.10", "1.2.0", -1}, {"1.2.0", "1.2.0-rc.1", 1}, {"v2.0.0", "2.0.0", 0}, {"1.2.3", "1.2.3", 0},
	} {
		if got, err := semver.Compare(c.a, c.b); err != nil || got != c.want {
			t.Errorf("Compare(%q, %q) = %d, %v; want %d", c.a, c.b, got, err, c.want)
		}
	}
	for _, in := range []string{"1.x.0", "1.2", "1.2.3.4", "-1.2.3", "1.2.3-beta.1", "1.2.3-rc.", ""} {
		if _, err := semver.Parse(in); err == nil || !strings.HasPrefix(err.Error(), "ERR_VERSION:") || !strings.Contains(err.Error(), in) {
			t.Errorf("Parse(%q) error %v, want ERR_VERSION: quoting the input", in, err)
		}
	}
	for in, want := range map[string]bool{"1.2.0-rc.1": true, "v2.0.0-rc.10": true, "1.2.0": false, "v1.0.0": false, "1.2.0-rc.": false, "nightly": false} {
		if got := semver.IsCandidate(in); got != want {
			t.Errorf("IsCandidate(%q) = %v, want %v", in, got, want)
		}
	}
}

// release keeps ALG-02 (the L3 cases): the item did not touch it.
func TestReleaseKeptVerify(t *testing.T) {
	for _, c := range []struct {
		tags []string
		want string
	}{
		{[]string{"v1.9.0", "v1.10.0-rc.1", "nightly", "1.10.0"}, "1.10.0"},
		{[]string{"v2.0.0-rc.1", "v2.0.0-rc.2"}, "v2.0.0-rc.2"},
		{[]string{"v3.0.0-rc.1", "v2.9.9"}, "v2.9.9"},
		{[]string{"1.2.0-rc.10", "1.2.0-rc.9"}, "1.2.0-rc.10"},
		{[]string{"release-1", "v1.0.0", "1.0", "latest"}, "v1.0.0"},
		{[]string{"v1.10.0", "v1.9.0", "v1.2.0"}, "v1.10.0"},
	} {
		if got, err := release.Latest(c.tags); err != nil || got != c.want {
			t.Errorf("Latest(%q) = %q, %v; want %q", c.tags, got, err, c.want)
		}
	}
	for _, tags := range [][]string{nil, {}, {"nightly", "1.x.0", "v1.2"}} {
		if _, err := release.Latest(tags); err == nil || !strings.HasPrefix(err.Error(), "ERR_NO_RELEASE:") {
			t.Errorf("Latest(%q) error %v, want ERR_NO_RELEASE:", tags, err)
		}
	}
}

// Package notes renders a release-notes file from tagged notes per TA ALG-03, on top of semver
// (ALG-01) and release (ALG-02). The reference solution of L4: the worker never sees it.
package notes

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"

	"bench/release"
	"bench/semver"
)

// Note is one line of a notes file: a tag and its text (ALG-03.1).
type Note struct {
	Tag  string
	Text string
}

// Parse reads notes from r, one per line as `<tag> <text>` (ALG-03.1); name names the input in an
// error (ALG-03.2).
func Parse(name string, r io.Reader) ([]Note, error) {
	var notes []Note
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		tag, text, _ := strings.Cut(line, " ")
		if text = strings.TrimSpace(text); text == "" {
			return nil, fmt.Errorf("ERR_NOTES: %s:%d: no text after the tag", name, n)
		}
		if _, err := semver.Parse(tag); err != nil {
			return nil, fmt.Errorf("ERR_NOTES: %s:%d: %v", name, n, err)
		}
		notes = append(notes, Note{Tag: tag, Text: text})
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("ERR_NOTES: %s: %v", name, err)
	}
	return notes, nil
}

// Render writes the release-notes file of notes per ALG-03.3 to ALG-03.5.
func Render(notes []Note) (string, error) {
	var tags []string
	texts := map[string][]string{}
	for _, n := range notes {
		if _, seen := texts[n.Tag]; !seen {
			if _, err := semver.Parse(n.Tag); err != nil {
				return "", fmt.Errorf("ERR_NOTES: %v", err)
			}
			tags = append(tags, n.Tag)
		}
		texts[n.Tag] = append(texts[n.Tag], n.Text)
	}
	var b strings.Builder
	b.WriteString("# Release notes\n")
	if len(tags) == 0 {
		return b.String(), nil
	}
	sort.SliceStable(tags, func(i, j int) bool { // newest first; equal tags keep their order (ALG-03.3)
		c, _ := semver.Compare(tags[i], tags[j])
		return c > 0
	})
	latest, err := release.Latest(tags) // ALG-03.4: a final release beats every candidate
	if err != nil {
		return "", fmt.Errorf("ERR_NOTES: %v", err)
	}
	fmt.Fprintf(&b, "\nLatest release: %s\n", latest)
	for _, t := range tags {
		b.WriteString("\n## " + t)
		if semver.IsCandidate(t) {
			b.WriteString(" (candidate)")
		}
		b.WriteString("\n\n")
		for _, text := range texts[t] {
			b.WriteString("- " + text + "\n")
		}
	}
	return b.String(), nil
}

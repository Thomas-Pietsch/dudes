package itempack

import (
	"regexp"
	"strings"
)

// docHeading is a document group line of the contracts part ("## TA"); quoted section
// headings inside a reference ("## 15. Error Catalogue") are not.
var docHeading = regexp.MustCompile(`^## [A-Z][A-Z0-9_]*$`)

// Narrow keeps the rules and the task of a PACK.md and, of its contracts, only the named
// references ("TA:§20.4", brackets optional) — the part of the item one task works on (F65).
// No refs → the pack unchanged.
func Narrow(pack string, refs []string) string {
	if len(refs) == 0 {
		return pack
	}
	want := map[string]bool{}
	for _, r := range refs {
		want["[["+strings.Trim(strings.TrimSpace(r), "[]")+"]]"] = true
	}
	i := strings.Index(pack, "\n# Contracts for ")
	if i < 0 {
		return pack
	}
	head, ctx := pack[:i+1], pack[i+1:]
	var out []string
	keep, doc := true, "" // doc: a group heading, written with its first kept reference
	for _, line := range strings.Split(ctx, "\n") {
		switch {
		case docHeading.MatchString(line):
			doc, keep = line, false
			continue
		case strings.HasPrefix(line, "### [["):
			keep = want[strings.TrimPrefix(line, "### ")]
			if keep && doc != "" {
				out = append(out, doc, "")
				doc = ""
			}
		}
		if keep {
			out = append(out, line)
		}
	}
	return head + strings.Join(out, "\n")
}

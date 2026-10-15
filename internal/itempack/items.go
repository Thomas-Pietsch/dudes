package itempack

import (
	"fmt"
	"regexp"
	"strings"
)

// Item kinds.
const (
	KindItem    = "item"
	KindOutline = "outline"
)

// Derived statuses (CP-FR-06). Never stored in a spec.
const (
	StatusOpen      = "open"
	StatusDone      = "done"
	StatusBlocked   = "blocked"
	StatusUntracked = "untracked"
	StatusOutline   = "outline"
)

// Item is one work item (or outline row) of the backlog.
type Item struct {
	ID      string
	Title   string
	Release string // "" when none
	Text    string
	Kind    string
}

// idsRe finds prefixed numeric IDs with optional range or slash list (AC-X-08/09).
var idsRe = regexp.MustCompile(`(([A-Z][A-Z0-9]*(?:-[A-Z0-9]+)*-)(\d+)((?:\.\.\d+)|(?:/\d+)+)?)(?:$|[^\p{L}\p{N}_])`)

// ExpandIDs returns the owner-family IDs in text, ranges and slash lists expanded.
func (s *Set) ExpandIDs(text string) []string {
	var ids []string
	scanBounded(idsRe, text, func(m []int) {
		pre, first := text[m[4]:m[5]], text[m[6]:m[7]]
		tail := ""
		if m[8] >= 0 {
			tail = text[m[8]:m[9]]
		}
		nums := []string{first}
		switch {
		case strings.HasPrefix(tail, ".."):
			var to int
			fmt.Sscanf(tail[2:], "%d", &to)
			var from int
			fmt.Sscanf(first, "%d", &from)
			nums = nil
			for n := from; n <= to; n++ {
				nums = append(nums, fmt.Sprintf("%0*d", len(first), n))
			}
		case strings.HasPrefix(tail, "/"):
			for _, x := range strings.Split(tail[1:], "/") {
				nums = append(nums, leftPad(x, len(first)))
			}
		}
		for _, n := range nums {
			ids = append(ids, pre+n)
		}
	})
	out := ids[:0]
	for _, id := range ids {
		if s.OwnerOf(id) != "" {
			out = append(out, id)
		}
	}
	return out
}

func leftPad(x string, width int) string {
	for len(x) < width {
		x = "0" + x
	}
	return x
}

// Items lists work items and outline rows of the backlog (CP-FR-01).
func (s *Set) Items() []Item {
	b := s.Profile.Backlog
	d, ok := s.docs[b.Doc]
	if !ok {
		return nil
	}
	var items []Item
	release := ""
	inside := false
	for i, line := range d.lines {
		if strings.HasPrefix(line, "## ") {
			inside = strings.HasPrefix(line, b.ItemsSection)
		}
		if !inside {
			continue
		}
		if s.re.releaseHead != nil {
			if m := s.re.releaseHead.FindStringSubmatch(line); m != nil {
				release = m[1]
			}
		}
		if m := s.re.itemHeading.FindStringSubmatch(line); m != nil {
			text := d.section(i)
			rel := ""
			if s.re.releaseField != nil {
				if r := s.re.releaseField.FindStringSubmatch(text); r != nil {
					rel = r[1]
				}
			}
			items = append(items, Item{ID: m[1], Title: trimSpace(m[2]), Release: rel, Text: text, Kind: KindItem})
			continue
		}
		if s.re.outlineRow != nil && release != "" {
			if m := s.re.outlineRow.FindStringSubmatch(line); m != nil {
				items = append(items, Item{ID: m[1], Title: trimSpace(m[2]), Release: release, Text: line, Kind: KindOutline})
			}
		}
	}
	return items
}

// FindItem returns the packable work item with id.
func (s *Set) FindItem(id string) (Item, bool) {
	for _, it := range s.Items() {
		if it.ID == id && it.Kind == KindItem {
			return it, true
		}
	}
	return Item{}, false
}

// DoDIDs are the QR row IDs named in the item's DoD field; backlog-owned IDs are ignored.
func (s *Set) DoDIDs(it Item) []string {
	m := s.re.dodField.FindStringSubmatch(it.Text)
	if m == nil {
		return []string{}
	}
	seen := map[string]bool{}
	out := []string{}
	for _, id := range s.ExpandIDs(m[1]) {
		if seen[id] || s.OwnerOf(id) == s.Profile.Backlog.Doc {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	return out
}

// StatusIndex maps every QR-row ID to the status marks of its row.
func (s *Set) StatusIndex() map[string][]string {
	st := s.Profile.Status
	marks := append(append(append([]string{}, st.Done...), st.Blocked...), st.Open...)
	idx := map[string][]string{}
	d, ok := s.docs[st.Doc]
	if !ok {
		return idx
	}
	for _, line := range d.lines {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := strings.Split(strings.Trim(trimSpace(line), "|"), "|")
		var found []string
		for _, c := range cells[1:] {
			c = trimSpace(c)
			for _, mk := range marks {
				if strings.HasPrefix(c, mk) {
					found = append(found, mk)
				}
			}
		}
		if len(found) > 0 {
			for _, id := range s.ExpandIDs(line) {
				idx[id] = append(idx[id], found...)
			}
			continue
		}
		// No status mark: a not-applicable mark speaks only for the IDs the row is about (first cell).
		if na := naMark(cells[1:], st.NotApplicable); na != "" {
			for _, id := range s.ExpandIDs(cells[0]) {
				idx[id] = append(idx[id], na)
			}
		}
	}
	return idx
}

func naMark(cells, na []string) string {
	for _, c := range cells {
		c = trimSpace(c)
		for _, mk := range na {
			if c == mk {
				return mk
			}
		}
	}
	return ""
}

// realMarks drops not-applicable marks; nil when the ID has nothing else (it is out of scope).
func realMarks(marks, na []string) []string {
	var out []string
	for _, mk := range marks {
		if !containsString(na, mk) {
			out = append(out, mk)
		}
	}
	return out
}

// DeriveStatus computes done / blocked / open / untracked / outline (CP-FR-06).
// idx may be nil.
func (s *Set) DeriveStatus(it Item, idx map[string][]string) (status string, dod, missing []string) {
	if idx == nil {
		idx = s.StatusIndex()
	}
	dod = s.DoDIDs(it)
	missing = []string{}
	if it.Kind == KindOutline {
		return StatusOutline, dod, missing
	}
	for _, id := range dod {
		if _, ok := idx[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return StatusUntracked, dod, missing
	}
	st := s.Profile.Status
	allDone := len(dod) > 0
	for _, id := range dod {
		for _, mk := range realMarks(idx[id], st.NotApplicable) { // only not-applicable: out of scope
			if containsString(st.Blocked, mk) {
				return StatusBlocked, dod, missing
			}
			if !containsString(st.Done, mk) {
				allDone = false
			}
		}
	}
	if allDone {
		return StatusDone, dod, missing
	}
	return StatusOpen, dod, missing
}

// Criterion is one acceptance criterion line of a work item.
type Criterion struct {
	ID   string
	Text string
}

// Acceptance returns the item's AC lines (backlog.ac_line).
func (s *Set) Acceptance(it Item) []Criterion {
	var out []Criterion
	for _, line := range strings.Split(it.Text, "\n") {
		if m := s.re.acLine.FindStringSubmatch(line); m != nil {
			out = append(out, Criterion{ID: m[1], Text: trimSpace(m[2])})
		}
	}
	return out
}

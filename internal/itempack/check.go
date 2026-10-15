package itempack

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	germanQuoteRe = regexp.MustCompile(`„([^“]+)“`)
	plainQuoteRe  = regexp.MustCompile(`"([^"]{3,})"`)
	capWordRe     = regexp.MustCompile(`[A-ZÄÖÜ][a-zäöüß]+ [a-zäöüßA-ZÄÖÜ{]`)
	umlautRe      = regexp.MustCompile(`[äöüß]`)
)

// Finding is one consistency violation: "<code> <location> <message>".
type Finding struct {
	Code string
	Msg  string
}

func (f Finding) String() string { return f.Code + " " + f.Msg }

// Check runs consistency rules C1–C8 (CP-FR-07). Zero findings = consistent.
func (s *Set) Check() []Finding {
	p, cfg := s.Profile, s.Profile.Consistency
	var out []Finding
	add := func(code, format string, a ...any) { out = append(out, Finding{code, fmt.Sprintf(format, a...)}) }

	// C1 — every reference resolves (inline code holds examples, skipped).
	for _, key := range s.keys {
		for n, line := range s.docs[key].lines {
			for _, m := range refRe.FindAllStringSubmatch(inlineCode.ReplaceAllString(line, ""), -1) {
				if _, err := s.Resolve(m[1], m[2]); err != nil {
					add("C1", "%s:%d %v", key, n+1, err)
				}
			}
		}
	}

	// C2 — every referenced owner-family ID is defined exactly once in its owner.
	shared := map[string]bool{}
	for _, d := range p.SharedDocuments {
		shared[d] = true
	}
	referenced := map[string]string{}
	for _, key := range s.keys {
		if shared[key] {
			continue
		}
		for _, line := range s.docs[key].lines {
			if cfg.SkipLineMarker != "" && strings.Contains(strings.ToLower(line), cfg.SkipLineMarker) {
				continue
			}
			scanBounded(s.re.ids, line, func(m []int) {
				id := line[m[2]:m[3]]
				if _, ok := referenced[id]; !ok {
					referenced[id] = key
				}
			})
		}
	}
	ids := make([]string, 0, len(referenced))
	for id := range referenced {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		own := s.OwnerOf(id)
		d, ok := s.docs[own]
		if !ok {
			continue
		}
		found := s.definitions(d, id)
		if len(found) == 0 {
			add("C2", "%s referenced in %s but not defined in %s", id, referenced[id], own)
		} else if len(found) > 1 && !hasAnyPrefix(id, p.MultiDefinitionAllowed) {
			add("C2", "%s defined %dx in %s", id, len(found), own)
		}
	}

	// C3 — work items hold no route shape, no DDL, no UI copy without a UX anchor.
	b := p.Backlog
	if d, ok := s.docs[b.Doc]; ok {
		inside := false
		for n, line := range d.lines {
			if strings.HasPrefix(line, "## ") {
				inside = strings.HasPrefix(line, b.ItemsSection)
			}
			if !inside || strings.HasPrefix(line, "|") {
				continue
			}
			bare := refRe.ReplaceAllString(line, "")
			if s.re.routeShape != nil && s.re.routeShape.MatchString(bare) {
				add("C3", "%s:%d route shape outside [[…]]", b.Doc, n+1)
			}
			if cfg.DDLMarker != "" && strings.Contains(bare, cfg.DDLMarker) {
				add("C3", "%s:%d DDL in work item", b.Doc, n+1)
			}
			plain := inlineCode.ReplaceAllString(bare, "")
			var quoted []string
			for _, m := range germanQuoteRe.FindAllStringSubmatch(plain, -1) {
				quoted = append(quoted, m[1])
			}
			for _, m := range plainQuoteRe.FindAllStringSubmatch(plain, -1) {
				if capWordRe.MatchString(m[1]) || umlautRe.MatchString(m[1]) {
					quoted = append(quoted, m[1])
				}
			}
			if len(quoted) > 0 && !strings.Contains(line, cfg.UIAnchor) {
				add("C3", "%s:%d UI string „%s“ without UX anchor", b.Doc, n+1, firstRunes(quoted[0], 40))
			}
		}
	}

	// C4 — the PRD holds no route table row and no DDL.
	if d, ok := s.docs["PRD"]; ok {
		for n, line := range d.lines {
			if (s.re.prdForbidden != nil && s.re.prdForbidden.MatchString(line)) || (cfg.DDLMarker != "" && strings.Contains(line, cfg.DDLMarker)) {
				add("C4", "PRD:%d route table row or DDL", n+1)
			}
		}
	}

	// C5 — every route named anywhere exists in its owner.
	if s.re.routePattern != nil {
		ownerText := ""
		if d, ok := s.docs[cfg.RouteOwner]; ok {
			ownerText = strings.Join(d.lines, "\n")
		}
		for _, key := range s.keys {
			for n, line := range s.docs[key].lines {
				for _, m := range s.re.routePattern.FindAllStringSubmatch(line, -1) {
					path := m[0]
					if len(m) > 1 {
						path = m[1]
					}
					if !strings.Contains(ownerText, path) {
						add("C5", "%s:%d route %s not in %s", key, n+1, path, cfg.RouteOwner)
					}
				}
			}
		}
	}

	// C6 / C7 — releases are ROADMAP milestones, and every milestone has items.
	releases := map[string]bool{}
	if d, ok := s.docs[p.Roadmap.Doc]; ok && s.re.milestoneRow != nil {
		for _, m := range s.re.milestoneRow.FindAllStringSubmatch(strings.Join(d.lines, "\n"), -1) {
			releases[m[len(m)-1]] = true
		}
	}
	items := s.Items()
	for _, it := range items {
		if it.Kind == KindItem && it.Release == "" {
			add("C6", "%s has no release field", it.ID)
		} else if it.Release != "" && !releases[it.Release] {
			add("C6", "%s release %s is not a ROADMAP milestone", it.ID, it.Release)
		}
	}
	rels := make([]string, 0, len(releases))
	for r := range releases {
		rels = append(rels, r)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		has := false
		for _, it := range items {
			if it.Release == rel {
				has = true
				break
			}
		}
		if !has {
			add("C7", "milestone %s has no work item", rel)
		}
	}

	// C8 — every DoD ID of a work item has a QR row.
	idx := s.StatusIndex()
	for _, it := range items {
		if it.Kind != KindItem {
			continue
		}
		for _, id := range s.DoDIDs(it) {
			if _, ok := idx[id]; !ok {
				add("C8", "%s DoD %s has no QR row", it.ID, id)
			}
		}
	}
	return out
}

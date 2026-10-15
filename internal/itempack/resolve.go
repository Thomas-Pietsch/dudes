package itempack

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ResolveError is a reference that does not resolve (CP-FR-03).
type ResolveError struct{ Msg string }

func (e *ResolveError) Error() string { return e.Msg }

func resolveErr(format string, a ...any) error { return &ResolveError{Msg: fmt.Sprintf(format, a...)} }

var (
	headNumRe   = regexp.MustCompile(`^[\d.]+[a-z]?[\s\v\x{85}\p{Z}]*`)
	leadIDRe    = regexp.MustCompile(`[A-Z][A-Z-]*-[\d.]+`)
	boldItemRe  = regexp.MustCompile(`^[\s\v\x{85}\p{Z}]*[-*][\s\v\x{85}\p{Z}]+\*\*([^*]+)\*\*`)
	boldParaRe  = regexp.MustCompile(`^\*\*([A-Z][\p{L}\p{N}_.-]*\p{Nd})(?:$|[^\p{L}\p{N}_])`)
	ddlTableRow = regexp.MustCompile("\\| `([a-z_]+)`")
)

type definition struct {
	line int
	text string
}

// definitions returns every definition of ident in doc (CP-FR-02 a–c, h).
func (s *Set) definitions(d *document, ident string) []definition {
	var found []definition
	for i, line := range d.lines {
		if d.fenced[i] {
			continue
		}
		if d.level[i] > 0 && hasToken(firstRunes(d.headText[i], 120), ident) {
			first := strings.TrimLeft(headNumRe.ReplaceAllString(d.headText[i], ""), "*` ")
			lead := ""
			if strings.Contains(first, " · ") {
				lead = strings.SplitN(first, " · ", 2)[0]
			}
			if strings.HasPrefix(first, ident) || containsString(leadIDRe.FindAllString(lead, -1), ident) {
				found = append(found, definition{i, d.section(i)})
				continue
			}
		}
		if strings.HasPrefix(line, "|") {
			parts := strings.Split(line, "|")
			cell := trimSpace(strings.Trim(trimSpace(parts[1]), "*`"))
			if cell == ident || (strings.HasPrefix(cell, ident) && !wordOrDot(cell[len(ident):])) {
				hdr := d.tableHeader(i)
				found = append(found, definition{i, strings.Join(append(hdr, line), "\n")})
				continue
			}
		}
		if m := boldItemRe.FindStringSubmatch(line); m != nil {
			if f := strings.Fields(m[1]); len(f) > 0 && strings.TrimRight(f[0], ":") == ident {
				found = append(found, definition{i, line})
				continue
			}
		}
		if m := boldParaRe.FindStringSubmatch(line); m != nil && m[1] == ident {
			block := []string{line}
			for _, nxt := range d.lines[i+1:] {
				if strings.HasPrefix(nxt, "**") || strings.HasPrefix(nxt, "#") || strings.HasPrefix(nxt, "---") {
					break
				}
				block = append(block, nxt)
			}
			found = append(found, definition{i, strings.TrimRightFunc(strings.Join(block, "\n"), unicode.IsSpace)})
		}
	}
	return found
}

func wordOrDot(rest string) bool {
	if rest == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(rest)
	return r == '.' || isWord(r)
}

func containsString(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// ExpandRange turns 'X-01..05' or 'X-01..X-05' into each ID, zero padding kept.
// Anything else is returned unchanged.
func ExpandRange(spec string) []string {
	for off := 0; ; {
		k := strings.Index(spec[off:], "..")
		if k < 0 {
			return []string{spec}
		}
		idx := off + k
		left, right := spec[:idx], spec[idx+2:]
		j := len(left)
		for j > 0 && left[j-1] >= '0' && left[j-1] <= '9' {
			j--
		}
		if j < len(left) {
			pre, a := left[:j], left[j:]
			b := ""
			if pre != "" && strings.HasPrefix(right, pre) && allDigits(right[len(pre):]) {
				b = right[len(pre):]
			} else if allDigits(right) {
				b = right
			}
			if b != "" {
				from, err1 := strconv.Atoi(a)
				to, err2 := strconv.Atoi(b)
				if err1 == nil && err2 == nil {
					out := []string{}
					for n := from; n <= to; n++ {
						out = append(out, fmt.Sprintf("%s%0*d", pre, len(a), n))
					}
					return out
				}
			}
		}
		off = idx + 1
	}
}

func allDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// ddl returns CREATE TABLE / TRIGGER / INDEX statements of table as one sql block (CP-FR-02 f).
func (d *document) ddl(table string) []string {
	text := strings.Join(d.lines, "\n")
	q := regexp.QuoteMeta(table)
	var out []string
	for _, pat := range []string{
		`(?s)CREATE TABLE ` + q + ` \(.*?\n\);`,
		`(?s)CREATE TRIGGER \w+ (?:BEFORE|AFTER) (?:UPDATE|DELETE|INSERT)(?: OF \w+)? ON ` + q + `\b.*?END;`,
		`(?s)CREATE (?:UNIQUE )?INDEX \w+ ON ` + q + `\b[^;]*;`,
	} {
		out = append(out, regexp.MustCompile(pat).FindAllString(text, -1)...)
	}
	if len(out) == 0 {
		return nil
	}
	return []string{"```sql\n" + strings.Join(out, "\n\n") + "\n```"}
}

// Resolve returns the definition texts of one reference target in doc (CP-FR-02).
func (s *Set) Resolve(doc, target string) ([]string, error) {
	key := doc + ":" + target
	if r, ok := s.cache[key]; ok {
		return r.parts, r.err
	}
	parts, err := s.resolve(doc, target)
	s.cache[key] = resolved{parts, err}
	return parts, err
}

func (s *Set) resolve(doc, target string) ([]string, error) {
	d, ok := s.docs[doc]
	if !ok {
		return nil, resolveErr("unknown or missing document %s", doc)
	}
	if strings.HasPrefix(target, "§") {
		num := strings.TrimPrefix(target, "§")
		for i := range d.lines {
			if d.fenced[i] || d.level[i] == 0 {
				continue
			}
			t := trimSpace(d.headText[i]) + " "
			if strings.HasPrefix(t, num) {
				r, _ := utf8.DecodeRuneInString(t[len(num):])
				if r == '.' || unicode.IsSpace(r) {
					return []string{d.section(i)}, nil
				}
			}
		}
		return nil, resolveErr("%s %s not found", doc, target)
	}
	var out []string
	p := s.Profile
	for _, ident := range ExpandRange(trimSpace(target)) {
		found := s.definitions(d, ident)
		if len(found) == 0 {
			return nil, resolveErr("%s:%s not defined", doc, ident)
		}
		if hasAnyPrefix(ident, p.MultiDefinitionAllowed) {
			for _, f := range found {
				out = append(out, f.text)
			}
		} else {
			out = append(out, found[0].text)
		}
		if p.TableDDL.Doc != "" && doc == p.TableDDL.Doc && p.TableDDL.IDPrefix != "" && strings.HasPrefix(ident, p.TableDDL.IDPrefix) {
			rows := strings.Split(found[0].text, "\n")
			if m := ddlTableRow.FindStringSubmatch(rows[len(rows)-1]); m != nil {
				out = append(out, d.ddl(m[1])...)
			}
		}
	}
	return mergeRows(out), nil
}

func hasAnyPrefix(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// mergeRows emits rows of the same table (range references) under one header.
func mergeRows(parts []string) []string {
	var merged []string
	for _, part := range parts {
		lines := strings.Split(part, "\n")
		if len(merged) > 0 && len(lines) == 3 && tableSepRe.MatchString(lines[1]) {
			prev := strings.Split(merged[len(merged)-1], "\n")
			if len(prev) >= 2 && prev[0] == lines[0] && prev[1] == lines[1] {
				merged[len(merged)-1] += "\n" + lines[2]
				continue
			}
		}
		merged = append(merged, part)
	}
	return merged
}

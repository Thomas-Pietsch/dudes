package itempack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// RE2 has no look-around. An ID token must not be preceded by a word rune or
// '-' (checked in code) and must not be followed by a word rune or '.digit'
// (consumed by idTrail; scanning resumes at the end of group 1).
const idTrail = `(?:$|[^\p{L}\p{N}_.]|\.$|\.[^\p{Nd}])`

var (
	refRe      = regexp.MustCompile(`\[\[([A-Z]+):([^\]]+)\]\]`)
	headRe     = regexp.MustCompile(`^(#{1,6})[\s\v\x{85}\p{Z}]+(.*)$`)
	tableSepRe = regexp.MustCompile(`^\|[\s:|-]+\|$`)
	inlineCode = regexp.MustCompile("`[^`]*`")
)

// Doc names one spec-set file. Key is the DOC key used in references.
type Doc struct {
	Key  string
	Path string // absolute
}

type document struct {
	key, name string
	sum       string
	lines     []string
	fenced    []bool
	level     []int    // heading level, 0 = not a heading (fences ignored, as in _section)
	headText  []string // heading text after the hashes
}

// Set is a loaded spec-set under one profile.
type Set struct {
	Profile *Profile
	re      *compiled
	keys    []string // load order
	docs    map[string]*document
	cache   map[string]resolved
}

type resolved struct {
	parts []string
	err   error
}

// New loads the listed documents from disk. A listed file that does not exist is not
// loaded (CP-FR-03); only a reference into it is an error.
func New(p *Profile, docs []Doc) (*Set, error) {
	return newWith(p, docs, os.ReadFile)
}

func newWith(p *Profile, docs []Doc, read func(string) ([]byte, error)) (*Set, error) {
	c, err := p.compile()
	if err != nil {
		return nil, err
	}
	s := &Set{Profile: p, re: c, docs: map[string]*document{}, cache: map[string]resolved{}}
	for _, d := range docs {
		data, err := read(d.Path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("itempack: %s: %w", d.Key, err)
		}
		if _, dup := s.docs[d.Key]; dup {
			return nil, fmt.Errorf("itempack: duplicate document key %s", d.Key)
		}
		h := sha256.Sum256(data)
		doc := &document{
			key:   d.Key,
			name:  filepath.Base(d.Path),
			sum:   hex.EncodeToString(h[:]),
			lines: strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n"),
		}
		doc.index()
		s.keys = append(s.keys, d.Key)
		s.docs[d.Key] = doc
	}
	return s, nil
}

// Open loads profile + spec-set from the workspace, the spec documents as they are on disk — what
// a spec author checks before committing (`dudes item check`). ok=false without a profile.
func Open(workspaceRoot, product string) (*Set, bool, error) {
	return open(workspaceRoot, product, os.ReadFile)
}

// OpenCommitted is Open with every spec document as its repository's HEAD holds it: what a run
// freezes, checks and hands to a worker, never a working copy someone is still editing.
func OpenCommitted(workspaceRoot, product string) (*Set, bool, error) {
	return open(workspaceRoot, product, gitx.ReadCommitted)
}

func open(workspaceRoot, product string, read func(string) ([]byte, error)) (*Set, bool, error) {
	p, ok, err := LoadProfile(workspaceRoot)
	if err != nil || !ok {
		return nil, ok, err
	}
	set, sok, err := specset.Load(workspaceRoot, product)
	if err != nil {
		return nil, true, err
	}
	if !sok {
		return nil, true, fmt.Errorf("itempack: %s needs %s (DOC keys are spec-set entry IDs)", profileRel, specset.Path())
	}
	docs, err := SpecSetDocs(workspaceRoot, set)
	if err != nil {
		return nil, true, err
	}
	s, err := newWith(p, docs, read)
	return s, true, err
}

// SpecSetDocs maps every spec-set entry to a Doc. Entries whose root is missing
// count as absent files.
func SpecSetDocs(workspaceRoot string, set *specset.Set) ([]Doc, error) {
	var out []Doc
	for _, e := range set.Entries {
		abs, err := set.ResolveAbs(workspaceRoot, e)
		if err != nil {
			if strings.Contains(err.Error(), "escapes") {
				return nil, err
			}
			continue
		}
		out = append(out, Doc{Key: e.ID, Path: abs})
	}
	return out, nil
}

// Has reports whether document key is loaded.
func (s *Set) Has(key string) bool { _, ok := s.docs[key]; return ok }

// Sources maps every loaded file name to its SHA-256 (staleness, CP-FR-05).
func (s *Set) Sources() map[string]string {
	out := map[string]string{}
	for _, k := range s.keys {
		out[s.docs[k].name] = s.docs[k].sum
	}
	return out
}

// OwnerOf returns the owner document of ident, or "".
func (s *Set) OwnerOf(ident string) string {
	for i, re := range s.re.owners {
		if re.MatchString(ident) {
			return s.Profile.Owners[i].Doc
		}
	}
	return ""
}

func (d *document) index() {
	n := len(d.lines)
	d.fenced = make([]bool, n)
	d.level = make([]int, n)
	d.headText = make([]string, n)
	inside := false
	for i, line := range d.lines {
		if strings.HasPrefix(strings.TrimLeftFunc(line, unicode.IsSpace), "```") {
			inside = !inside
			d.fenced[i] = true
		} else if inside {
			d.fenced[i] = true
		}
		if m := headRe.FindStringSubmatch(line); m != nil {
			d.level[i] = len(m[1])
			d.headText[i] = m[2]
		}
	}
}

// section is the heading at i plus everything up to the next heading of equal or higher level.
func (d *document) section(i int) string {
	level := d.level[i]
	if level == 0 {
		return strings.TrimRightFunc(d.lines[i], unicode.IsSpace)
	}
	end := i + 1
	for end < len(d.lines) && (d.level[end] == 0 || d.level[end] > level) {
		end++
	}
	return strings.TrimRightFunc(strings.Join(d.lines[i:end], "\n"), unicode.IsSpace)
}

func (d *document) tableHeader(i int) []string {
	j := i
	for j > 0 && strings.HasPrefix(d.lines[j-1], "|") {
		j--
	}
	if j+1 < len(d.lines) && tableSepRe.MatchString(d.lines[j+1]) {
		return []string{d.lines[j], d.lines[j+1]}
	}
	return nil
}

// --- text helpers (Python str semantics) ---

func isWord(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) }

// boundedBefore: s[:i] does not end in a word rune or '-'.
func boundedBefore(s string, i int) bool {
	if i == 0 {
		return true
	}
	r, _ := utf8.DecodeLastRuneInString(s[:i])
	return !isWord(r) && r != '-'
}

// boundedAfter: s[i:] does not start with a word rune or '.digit'.
func boundedAfter(s string, i int) bool {
	if i >= len(s) {
		return true
	}
	r, n := utf8.DecodeRuneInString(s[i:])
	if isWord(r) {
		return false
	}
	if r == '.' && i+n < len(s) {
		r2, _ := utf8.DecodeRuneInString(s[i+n:])
		return !unicode.IsDigit(r2)
	}
	return true
}

// hasToken reports whether ident occurs in s as a bounded token.
func hasToken(s, ident string) bool {
	for pos := 0; pos <= len(s); {
		k := strings.Index(s[pos:], ident)
		if k < 0 {
			return false
		}
		st := pos + k
		if boundedBefore(s, st) && boundedAfter(s, st+len(ident)) {
			return true
		}
		_, n := utf8.DecodeRuneInString(s[st:])
		pos = st + n
	}
	return false
}

// scanBounded emulates Python finditer with a (?<![\w-]) look-behind: re's
// group 1 is the token; matches whose start follows a word rune or '-' are
// skipped one rune at a time. fn receives submatch indices into s.
func scanBounded(re *regexp.Regexp, s string, fn func(m []int)) {
	for pos := 0; pos <= len(s); {
		loc := re.FindStringSubmatchIndex(s[pos:])
		if loc == nil {
			return
		}
		for k := range loc {
			if loc[k] >= 0 {
				loc[k] += pos
			}
		}
		if !boundedBefore(s, loc[2]) {
			_, n := utf8.DecodeRuneInString(s[loc[2]:])
			if n == 0 {
				return
			}
			pos = loc[2] + n
			continue
		}
		fn(loc)
		if loc[3] > pos {
			pos = loc[3]
		} else {
			pos++
		}
	}
}

func firstRunes(s string, n int) string {
	i := 0
	for k := range s {
		if i == n {
			return s[:k]
		}
		i++
	}
	return s
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }

func trimSpace(s string) string { return strings.TrimFunc(s, unicode.IsSpace) }

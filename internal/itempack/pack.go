package itempack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/version"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Schema is the manifest format identifier (kept for kit interop).
const Schema = "context-pack/1"

const packsRel = ".dudes/items"

// FileNames in render order.
var FileNames = []string{"RULES.md", "TASK.md", "CONTEXT.md", "PACK.md"}

// now is replaced in tests.
var now = time.Now

// Ref is one resolved reference in the manifest.
type Ref struct {
	Ref     string `json:"ref"`
	Doc     string `json:"doc"`
	ID      string `json:"id"`
	Chars   int    `json:"chars"`
	Trimmed bool   `json:"trimmed"`
}

// Manifest is the machine interface of one pack (CP-FR-05).
type Manifest struct {
	Schema     string            `json:"schema"`
	Generator  string            `json:"generator"`
	Generated  string            `json:"generated"`
	Product    string            `json:"product"`
	Item       string            `json:"item"`
	Title      string            `json:"title"`
	Release    *string           `json:"release"`
	Status     string            `json:"status"`
	DoD        []string          `json:"dod"`
	DoDMissing []string          `json:"dod_missing"`
	Budget     *int              `json:"budget"`
	OverBudget bool              `json:"over_budget"`
	Sizes      map[string]int    `json:"sizes"`
	References []Ref             `json:"references"`
	Sources    map[string]string `json:"sources"`
}

// Pack is one rendered item pack. Nothing is written until Write.
type Pack struct {
	Item     string
	Files    map[string]string
	Manifest Manifest
}

type packRef struct {
	ref, doc, id, text string
	trimmed            bool
}

// Build assembles the pack of one work item (CP-FR-01..05). budget ≤ 0 means none.
// Any unresolved reference fails the build: the item is not ready.
func (s *Set) Build(itemID string, budget int) (*Pack, error) {
	it, ok := s.FindItem(itemID)
	if !ok {
		for _, o := range s.Items() {
			if o.ID == itemID {
				return nil, resolveErr("work item %s is an outline row — write it out before packing", itemID)
			}
		}
		return nil, resolveErr("work item %s not found", itemID)
	}
	var refs []*packRef
	var errs []string
	seen := map[string]bool{}
	for _, m := range refRe.FindAllStringSubmatch(it.Text, -1) {
		doc, target := m[1], m[2]
		key := doc + ":" + target
		if seen[key] {
			continue
		}
		seen[key] = true
		parts, err := s.Resolve(doc, target)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		refs = append(refs, &packRef{ref: "[[" + key + "]]", doc: doc, id: target, text: strings.Join(parts, "\n\n")})
	}
	if len(errs) > 0 {
		return nil, resolveErr("unresolved references: %s", strings.Join(errs, "; "))
	}
	rules, task := s.rulesText(), it.Text
	over := false
	var budgetPtr *int
	if budget > 0 {
		budgetPtr = &budget
		fixed := runeLen(rules) + runeLen(task)
		total := func() int {
			n := fixed
			for _, r := range refs {
				n += runeLen(r.text)
			}
			return n
		}
		order := append([]*packRef{}, refs...)
		sort.SliceStable(order, func(i, j int) bool { return runeLen(order[i].text) > runeLen(order[j].text) })
		for _, r := range order {
			if total() <= budget {
				break
			}
			lines := strings.Split(r.text, "\n")
			if len(lines) > 12 {
				lines = lines[:12]
			}
			r.text = strings.Join(lines, "\n") + fmt.Sprintf("\n\n> *Shortened to fit the budget of %d characters; full text: %s.*", budget, r.ref)
			r.trimmed = true
		}
		over = total() > budget
	}
	status, dod, missing := s.DeriveStatus(it, nil)
	date := now().Format("2006-01-02")
	stamp := fmt.Sprintf("> Generated %s by dudes %s. Build artifact — do not edit; regenerate.", date, version.Version)

	ctx := []string{"# Contracts for " + itemID, "", stamp, ""}
	for _, doc := range s.contextDocs(refs) {
		var group []*packRef
		for _, r := range refs {
			if r.doc == doc {
				group = append(group, r)
			}
		}
		if len(group) == 0 {
			continue
		}
		ctx = append(ctx, "## "+doc, "")
		for _, r := range group {
			ctx = append(ctx, "### "+r.ref, "", r.text, "")
		}
	}
	files := map[string]string{
		"RULES.md":   strings.Join([]string{"# Rules for " + itemID, "", stamp, "", rules, ""}, "\n"),
		"TASK.md":    strings.Join([]string{"# Task " + itemID + " — " + it.Title, "", stamp, "", task, ""}, "\n"),
		"CONTEXT.md": strings.Join(ctx, "\n"),
	}
	files["PACK.md"] = strings.Join([]string{files["RULES.md"], files["TASK.md"], files["CONTEXT.md"]}, "\n\n")

	var release *string
	if it.Release != "" {
		r := it.Release
		release = &r
	}
	man := Manifest{
		Schema: Schema, Generator: "dudes " + version.Version, Generated: date,
		Product: s.Profile.Product, Item: itemID, Title: it.Title, Release: release,
		Status: status, DoD: dod, DoDMissing: missing, Budget: budgetPtr, OverBudget: over,
		Sizes: map[string]int{}, References: []Ref{}, Sources: s.Sources(),
	}
	for _, n := range FileNames {
		man.Sizes[n] = runeLen(files[n])
	}
	for _, r := range refs {
		man.References = append(man.References, Ref{Ref: r.ref, Doc: r.doc, ID: r.id, Chars: runeLen(r.text), Trimmed: r.trimmed})
	}
	return &Pack{Item: itemID, Files: files, Manifest: man}, nil
}

// contextDocs is the profile context_order, then any other referenced doc alphabetically.
func (s *Set) contextDocs(refs []*packRef) []string {
	order := append([]string{}, s.Profile.ContextOrder...)
	in := map[string]bool{}
	for _, d := range order {
		in[d] = true
	}
	var rest []string
	for _, r := range refs {
		if !in[r.doc] {
			in[r.doc] = true
			rest = append(rest, r.doc)
		}
	}
	sort.Strings(rest)
	return append(order, rest...)
}

func (s *Set) rulesText() string {
	var out []string
	for _, r := range s.Profile.RulesSections {
		d, ok := s.docs[r.Doc]
		if !ok {
			continue
		}
		for i, line := range d.lines {
			if strings.HasPrefix(line, r.HeadingPrefix) {
				out = append(out, d.section(i))
				break
			}
		}
	}
	return strings.Join(out, "\n\n")
}

// Dir is the pack folder of item relative to the workspace.
func Dir(itemID string) string { return packsRel + "/" + itemID }

// Write stores the pack under <workspace>/.dudes/items/<ID>/ and returns that path (relative).
func Write(workspaceRoot string, p *Pack) (string, error) {
	rel := Dir(p.Item)
	abs, err := workspace.Resolve(workspaceRoot, rel)
	if err != nil {
		return "", err
	}
	if err := WriteDir(abs, p); err != nil {
		return "", err
	}
	return rel, nil
}

// WriteDir writes the five pack files into dir.
func WriteDir(dir string, p *Pack) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, n := range FileNames {
		body := p.Files[n]
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		if err := os.WriteFile(filepath.Join(dir, n), []byte(body), 0o644); err != nil {
			return err
		}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p.Manifest); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), buf.Bytes(), 0o644)
}

// Targets groups the item's reference targets by document, in context order,
// first appearance and without duplicates. Nothing is resolved.
func (s *Set) Targets(it Item) (docs []string, byDoc map[string][]string) {
	byDoc = map[string][]string{}
	seen := map[string]bool{}
	var refs []*packRef
	for _, m := range refRe.FindAllStringSubmatch(it.Text, -1) {
		key := m[1] + ":" + m[2]
		if seen[key] {
			continue
		}
		seen[key] = true
		byDoc[m[1]] = append(byDoc[m[1]], m[2])
		refs = append(refs, &packRef{doc: m[1]})
	}
	for _, d := range s.contextDocs(refs) {
		if len(byDoc[d]) > 0 {
			docs = append(docs, d)
		}
	}
	return docs, byDoc
}

// Trailers records the item and every source hash as git commit trailers.
// Git is the archive: the pack itself is a build artifact and is not kept.
func (p *Pack) Trailers() string {
	names := make([]string, 0, len(p.Manifest.Sources))
	for n := range p.Manifest.Sources {
		names = append(names, n)
	}
	sort.Strings(names)
	lines := []string{"Item: " + p.Item}
	for _, n := range names {
		lines = append(lines, "Spec-Source: "+n+" "+p.Manifest.Sources[n])
	}
	return strings.Join(lines, "\n")
}

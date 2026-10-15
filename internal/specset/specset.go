// Package specset registers instance spec artifacts without renaming them (F57).
package specset

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/product"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

const relPath = ".dudes/specset.json"

// Instance companion kinds (CORE §6, F57). "other" is an escape hatch.
const (
	KindBM      = "bm"
	KindPS      = "ps"
	KindPRD     = "prd"
	KindTA      = "ta"
	KindSC      = "sc"
	KindUX      = "ux"
	KindQR      = "qr"
	KindRoadmap = "roadmap"
	KindBacklog = "backlog"
	KindIdea    = "idea"
	KindOther   = "other"
)

// KnownKind reports whether kind is a named F57 kind (including other).
func KnownKind(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case KindBM, KindPS, KindPRD, KindTA, KindSC, KindUX, KindQR, KindRoadmap, KindBacklog, KindIdea, KindOther:
		return true
	default:
		return false
	}
}

// Entry is one registered document. Path keeps the instance filename.
// The file may not exist yet (specs in work); freeze hashes it when it appears.
type Entry struct {
	ID   string `json:"id"`
	Kind string `json:"kind"` // bm, ps, prd, ta, sc, ux, qr, roadmap, backlog, idea, other
	Root string `json:"root"` // "workspace" or a key in Roots
	Path string `json:"path"` // original relative path under that root
}

// Set is the product's document inventory.
type Set struct {
	Product string            `json:"product"`
	Roots   map[string]string `json:"roots,omitempty"` // alias → path relative to workspace (siblings allowed)
	Entries []Entry           `json:"entries"`
}

// Path is the registry file relative to workspace.
func Path() string { return relPath }

// EngineSet is the F13 default (workspace-only names).
func EngineSet(productName string) *Set {
	if productName == "" {
		productName = "product"
	}
	paths := product.SpecSet(productName)
	kinds := []string{"idea", "ps", "prd", "ta", "qr"}
	s := &Set{Product: productName, Roots: map[string]string{}, Entries: nil}
	for i, p := range paths {
		kind := "other"
		if i < len(kinds) {
			kind = kinds[i]
		}
		s.Entries = append(s.Entries, Entry{ID: strings.ToUpper(kind), Kind: kind, Root: "workspace", Path: p})
	}
	return s
}

// Load reads the registry. Missing file → EngineSet, ok=false.
func Load(workspaceRoot, productName string) (*Set, bool, error) {
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return EngineSet(productName), false, nil
		}
		return nil, false, fmt.Errorf("specset: %w", err)
	}
	var s Set
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, false, fmt.Errorf("specset: parse: %w", err)
	}
	if s.Product == "" {
		s.Product = productName
	}
	if s.Roots == nil {
		s.Roots = map[string]string{}
	}
	return &s, true, nil
}

// Save writes the registry.
func Save(workspaceRoot string, s *Set) error {
	if s == nil {
		return fmt.Errorf("specset: nil")
	}
	abs, err := workspace.Resolve(workspaceRoot, relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(abs, append(data, '\n'), 0o644)
}

// RootAbs resolves a named root. "workspace" is the engine workspace.
// Other aliases may point at sibling directories (explicit allow-list, F57).
func (s *Set) RootAbs(workspaceRoot, name string) (string, error) {
	if name == "" || name == "workspace" {
		return filepath.Abs(workspaceRoot)
	}
	rel, ok := s.Roots[name]
	if !ok || strings.TrimSpace(rel) == "" {
		return "", fmt.Errorf("specset: unknown root %q", name)
	}
	ws, err := filepath.Abs(workspaceRoot)
	if err != nil {
		return "", err
	}
	p := rel
	if !filepath.IsAbs(p) {
		p = filepath.Join(ws, filepath.FromSlash(rel))
	}
	p = filepath.Clean(p)
	fi, err := os.Stat(p)
	if err != nil {
		return "", fmt.Errorf("specset: root %q: %w", name, err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("specset: root %q is not a directory", name)
	}
	return p, nil
}

// ResolveAbs returns the absolute path of an entry, contained under its named root.
func (s *Set) ResolveAbs(workspaceRoot string, e Entry) (string, error) {
	rootAbs, err := s.RootAbs(workspaceRoot, e.Root)
	if err != nil {
		return "", err
	}
	return workspace.Resolve(rootAbs, e.Path)
}

// Find returns the entry with id (case-insensitive).
func (s *Set) Find(id string) (Entry, bool) {
	want := strings.ToLower(strings.TrimSpace(id))
	for _, e := range s.Entries {
		if strings.ToLower(e.ID) == want || strings.ToLower(e.Kind) == want {
			return e, true
		}
	}
	return Entry{}, false
}

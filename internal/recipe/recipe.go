// Package recipe loads thin per-phase worker recipes (F27).
// Format is JSON (zero deps); one file per phase under recipes/<phase>.json.
package recipe

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
)

// Dir is the workspace-relative recipes folder.
const Dir = "recipes"

// Recipe is a declarative phase pack: instructions + worker kind + optional flags.
// One recipe = one bounded invoke, then exit (Goose-style, orchestrated by Go).
type Recipe struct {
	Phase        phase.ID `json:"phase"`
	Worker       string   `json:"worker"` // ollama | aider | stub (empty = engine default)
	Instructions string   `json:"instructions"`
	Tools        []string `json:"tools"`  // advisory allow-list; empty = engine default
	Skills       []string `json:"skills"` // skill names under skills/<name>/SKILL.md (F48)
	Rules        []string `json:"rules"`  // workspace-relative markdown injected this phase only (F49)
	Goals        []string `json:"goals"`  // make goals this phase's worker may run; their ## lines go into the prompt (F75)
	Pack         bool     `json:"pack"`   // if true, refresh context pack before this phase
}

// Path returns recipes/<id>.json relative to workspace.
func Path(id phase.ID) string {
	return filepath.ToSlash(filepath.Join(Dir, string(id)+".json"))
}

// Load reads a recipe for id. ok=false if file missing.
func Load(root string, id phase.ID) (*Recipe, bool, error) {
	rel := Path(id)
	abs, err := workspace.Resolve(root, rel)
	if err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("recipe: read %s: %w", rel, err)
	}
	var r Recipe
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, false, fmt.Errorf("recipe: parse %s: %w", rel, err)
	}
	if r.Phase == "" {
		r.Phase = id
	}
	if r.Phase != id {
		return nil, false, fmt.Errorf("recipe: %s declares phase %q, expected %q", rel, r.Phase, id)
	}
	r.Worker = strings.ToLower(strings.TrimSpace(r.Worker))
	switch r.Worker {
	case "", "ollama", "aider", "stub", "pi", "human":
	default:
		return nil, false, fmt.Errorf("recipe: %s unknown worker %q (want ollama|aider|stub|pi|human)", rel, r.Worker)
	}
	return &r, true, nil
}

// TaskOverride returns instructions if non-empty.
func (r *Recipe) TaskOverride() string {
	if r == nil {
		return ""
	}
	return strings.TrimSpace(r.Instructions)
}

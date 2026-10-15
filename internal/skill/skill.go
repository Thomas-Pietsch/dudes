// Package skill loads portable discipline units (F48, TREND §6).
// Skills carry discipline, never phase sequence.
package skill

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/compact"
	"github.com/Thomas-Pietsch/dudes/internal/workspace"
	"github.com/Thomas-Pietsch/dudes/skills"
)

const (
	// Dir is the workspace-relative skills folder.
	Dir = "skills"
	// FileName is the Agent Skills standard entry file.
	FileName = "SKILL.md"
	maxBody  = 8000
	// FrameworkFolder holds the framework world's skills (CORE §5): DevOps on DUDES, never a worker's.
	FrameworkFolder = "framework"
)

// reName is the agentskills.io name: lowercase letters, digits, single hyphens, no hyphen at
// either end; at most 64 characters (ValidName).
var reName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// RelPath returns where skill name lives under root: skills/<name>/SKILL.md in an instance repo
// (the repo is one level, its skills sit directly in skills/), skills/<folder>/<name>/SKILL.md in
// the framework repo, which groups its skills by folder (framework, dev, spec, lead, shared —
// CORE §5). Not found → the flat path, for the error message.
func RelPath(root, name string) string {
	flat := filepath.ToSlash(filepath.Join(Dir, name, FileName))
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(flat))); err == nil {
		return flat
	}
	if m, _ := filepath.Glob(filepath.Join(root, Dir, "*", name, FileName)); len(m) == 1 {
		if rel, err := filepath.Rel(root, m[0]); err == nil {
			return filepath.ToSlash(rel)
		}
	}
	return flat
}

// ValidName reports a path-safe skill id.
func ValidName(name string) bool {
	return len(name) <= 64 && reName.MatchString(name)
}

// Load reads one skill body: from the workspace first — an instance's own skills, or the framework
// repo when it is the workspace — else from the framework's level skills built into the binary
// (package skills), so an instance repo needs no copy of them. ok=false if missing.
func Load(root, name string) (body string, ok bool, err error) {
	name = strings.TrimSpace(name)
	if !ValidName(name) {
		return "", false, fmt.Errorf("skill: invalid name %q", name)
	}
	rel := RelPath(root, name)
	abs, err := workspace.Resolve(root, rel)
	if err != nil {
		return "", false, err
	}
	data, err := os.ReadFile(abs)
	if err == nil {
		return compact.Text(string(data), maxBody), true, nil
	}
	if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("skill: read %s: %w", rel, err)
	}
	if m, _ := fs.Glob(skills.FS, "*/"+name+"/"+FileName); len(m) == 1 {
		if data, err := fs.ReadFile(skills.FS, m[0]); err == nil {
			return compact.Text(string(data), maxBody), true, nil
		}
	}
	return "", false, nil
}

// LoadMany concatenates the skills a recipe or contract task gives a worker. Missing names are an
// error (a recipe that names a skill must ship it), and so is a framework skill.
func LoadMany(root string, names []string) (string, error) {
	var parts []string
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		body, ok, err := Load(root, name)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("skill %q not found (expected %s or %s/<folder>/%s/%s)", name, RelPath(root, name), Dir, name, FileName)
		}
		if strings.HasPrefix(RelPath(root, name), Dir+"/"+FrameworkFolder+"/") {
			return "", fmt.Errorf("skill %q is a framework skill (%s/%s/) — never a worker's (CORE §5)", name, Dir, FrameworkFolder)
		}
		parts = append(parts, fmt.Sprintf("### Skill: %s\n\n%s", name, body))
	}
	return strings.Join(parts, "\n\n"), nil
}

// LoadRules reads workspace-relative markdown files named by a recipe (F49).
func LoadRules(root string, rels []string) (string, error) {
	var parts []string
	for _, rel := range rels {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		abs, err := workspace.Resolve(root, rel)
		if err != nil {
			return "", fmt.Errorf("rules: %s: %w", rel, err)
		}
		data, err := os.ReadFile(abs)
		if err != nil {
			return "", fmt.Errorf("rules: read %s: %w", rel, err)
		}
		parts = append(parts, fmt.Sprintf("### Standard: %s\n\n%s", rel, compact.Text(string(data), maxBody)))
	}
	return strings.Join(parts, "\n\n"), nil
}

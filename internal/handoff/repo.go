package handoff

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// Dir is the engine-ready repository's tracked folder: what its work items need from the engine.
const Dir = "dudes"

// Root is one spec root of a dev repository: the spec repository, where it sits relative to the dev
// repository, and the ref a slice's run pins it at — a release branch or a tag, the spec as frozen.
type Root struct {
	Repo string `json:"repo"`
	Pin  string `json:"pin"`
}

// SpecSet is dudes/specset.json: the spec-set of .dudes/specset.json (F57), its roots named with
// the repository and the pin instead of a path.
type SpecSet struct {
	Product string          `json:"product,omitempty"`
	Roots   map[string]Root `json:"roots"`
	Entries []specset.Entry `json:"entries"`
}

// Repo is the tracked dudes/ folder of an engine-ready dev repository, as one commit holds it:
//
//	dudes/profile.json   the spec format (F68)
//	dudes/specset.json   the spec roots and their documents
//	dudes/config.json    the engine's settings for this repository — the worker, the gate image,
//	                     the formatter — without the workspace and without machine values
//
// Machine values, above all the model host, never come from a repository.
type Repo struct {
	Profile []byte
	SpecSet SpecSet
	Config  []byte // nil: the repository sets nothing
}

// ReadRepo reads dudes/ as commit rev of the repository at dir holds it: tracked, so committed — an
// uncommitted edit is work in progress, as for the freeze (F59).
func ReadRepo(dir, rev string) (*Repo, error) {
	read := func(name string, required bool) ([]byte, error) {
		data, err := gitx.ShowFile(dir, rev, Dir+"/"+name)
		if err == nil {
			return data, nil
		}
		if !required && os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s/%s is not in %s — the repository is not engine-ready (F88): it tracks no %s", Dir, name, rev, name)
	}
	var r Repo
	var err error
	if r.Profile, err = read("profile.json", true); err != nil {
		return nil, err
	}
	if _, err := itempack.ParseProfile(r.Profile); err != nil {
		return nil, fmt.Errorf("%s/profile.json: %w", Dir, err)
	}
	data, err := read("specset.json", true)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &r.SpecSet); err != nil {
		return nil, fmt.Errorf("%s/specset.json: %w", Dir, err)
	}
	if err := r.SpecSet.validate(); err != nil {
		return nil, fmt.Errorf("%s/specset.json: %w", Dir, err)
	}
	if r.Config, err = read("config.json", false); err != nil {
		return nil, err
	}
	return &r, nil
}

func (s SpecSet) validate() error {
	for name, r := range s.Roots {
		if name == "" || name == "workspace" {
			return fmt.Errorf("root name %q: \"workspace\" is the dev repository itself", name)
		}
		if strings.TrimSpace(r.Repo) == "" || strings.TrimSpace(r.Pin) == "" {
			return fmt.Errorf("root %s needs its repository and the ref it is pinned at: {\"repo\": \"../spec-repo\", \"pin\": \"release/<product>-v1.0\"}", name)
		}
	}
	for _, e := range s.Entries {
		if e.ID == "" || e.Path == "" {
			return fmt.Errorf("entry %+v needs an id and a path", e)
		}
		if _, ok := s.Roots[e.Root]; !ok && e.Root != "" && e.Root != "workspace" {
			return fmt.Errorf("entry %s names root %q, which the file does not declare", e.ID, e.Root)
		}
	}
	return nil
}

// RootNames are the declared roots, sorted: pins are made in a stable order.
func (s SpecSet) RootNames() []string {
	var names []string
	for n := range s.Roots {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Specset is .dudes/specset.json for a run: the entries as the repository tracks them, each root
// pointing at where its pin was checked out — a path relative to the workspace when there is one.
func (s SpecSet) Specset(product, workspace string, pins map[string]string) *specset.Set {
	out := &specset.Set{Product: product, Roots: map[string]string{}, Entries: s.Entries}
	for name, abs := range pins {
		out.Roots[name] = relative(workspace, abs)
	}
	return out
}

// relative is target as a slash-separated path from base, or target itself when no relative path
// exists (another drive).
func relative(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(target)
}

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
	"github.com/Thomas-Pietsch/dudes/internal/product"
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

// Repo is the tracked dudes/ folder of an engine-ready dev repository, as one commit holds it, for
// the product of one work item. A repository of one product keeps its files in dudes/ itself:
//
//	dudes/profile.json   the spec format (F68)
//	dudes/specset.json   the spec roots and their documents
//	dudes/config.json    the engine's settings for this repository — the worker, the gate image,
//	                     the formatter — without the workspace and without machine values
//
// A repository of several products keeps a folder per product, named for it, beside the settings
// they share:
//
//	dudes/config.json              the settings every product shares
//	dudes/<product>/profile.json   the product's spec format; its owners give its work items' IDs
//	dudes/<product>/specset.json   its spec roots, each pinned at the product's own release
//	dudes/<product>/config.json    its own settings, over the shared ones
//
// The work item's ID finds its product. Machine values, above all the model host, never come from a
// repository.
type Repo struct {
	Product  string   // the product's folder; "" in a repository of one product
	Products []string // the product folders, sorted; none in a repository of one product
	Profile  []byte
	SpecSet  SpecSet
	Configs  []File // the settings, the shared ones first: each applies over the one before
}

// File is one tracked file of dudes/: its path in the repository and what the commit holds.
type File struct {
	Path string
	Data []byte
}

// ReadRepo reads dudes/ as commit rev of the repository at dir holds it — tracked, so committed: an
// uncommitted edit is work in progress, as for the freeze (F59) — for work item item. In a repository
// of several products the item's product is the one whose profile gives its backlog the item's ID; an
// item no product claims, or more than one, is refused with the products the repository knows.
func ReadRepo(dir, rev, item string) (*Repo, error) {
	files, err := gitx.Files(dir, rev, Dir)
	if err != nil {
		return nil, err
	}
	single, folders := false, map[string]bool{}
	for _, f := range files {
		parts := strings.Split(strings.TrimPrefix(f, Dir+"/"), "/")
		switch {
		case len(parts) == 1 && (parts[0] == "profile.json" || parts[0] == "specset.json"):
			single = true
		case len(parts) == 2 && (parts[1] == "profile.json" || parts[1] == "specset.json"):
			folders[parts[0]] = true
		}
	}
	var products []string
	for p := range folders {
		products = append(products, p)
	}
	sort.Strings(products)
	switch {
	case single && len(products) > 0:
		return nil, fmt.Errorf("%s/ in %s holds the files of one product and the folders of several (%s): keep one layout (F88)", Dir, rev, strings.Join(products, ", "))
	case single:
		return readProduct(dir, rev, "", nil)
	case len(products) == 0:
		return nil, fmt.Errorf("%s/profile.json is not in %s, and no %s/<product>/profile.json either — the repository is not engine-ready (F88): it tracks no profile.json", Dir, rev, Dir)
	}
	var known, claim []string
	for _, p := range products {
		if err := product.ValidateSlug(p); err != nil {
			return nil, fmt.Errorf("%s/%s/: a product folder is named for its product: %w", Dir, p, err)
		}
		data, err := gitx.ShowFile(dir, rev, Dir+"/"+p+"/profile.json")
		if err != nil {
			return nil, fmt.Errorf("%s/%s/profile.json is not in %s — the repository is not engine-ready (F88): every product folder tracks its profile.json", Dir, p, rev)
		}
		prof, err := itempack.ParseProfile(data)
		if err != nil {
			return nil, fmt.Errorf("%s/%s/profile.json: %w", Dir, p, err)
		}
		family := prof.ItemFamily()
		if family == "" {
			family = "no owner gives its backlog its IDs"
		}
		known = append(known, p+" ("+family+")")
		if prof.OwnsItem(item) {
			claim = append(claim, p)
		}
	}
	switch len(claim) {
	case 0:
		return nil, fmt.Errorf("work item %s is no product's of this repository — it knows %s", item, strings.Join(known, ", "))
	case 1:
		return readProduct(dir, rev, claim[0], products)
	}
	return nil, fmt.Errorf("work item %s fits the work items of %s: an item is one product's — the repository knows %s", item, strings.Join(claim, " and "), strings.Join(known, ", "))
}

// readProduct reads the files of one product: those of dudes/ itself (name ""), or those of its folder
// beside the shared config.json.
func readProduct(dir, rev, name string, products []string) (*Repo, error) {
	at := Dir + "/"
	if name != "" {
		at += name + "/"
	}
	read := func(path string, required bool) ([]byte, error) {
		data, err := gitx.ShowFile(dir, rev, path)
		if err == nil {
			return data, nil
		}
		if !required && os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("%s is not in %s — the repository is not engine-ready (F88): it tracks no %s", path, rev, filepath.Base(path))
	}
	r := Repo{Product: name, Products: products}
	var err error
	if r.Profile, err = read(at+"profile.json", true); err != nil {
		return nil, err
	}
	if _, err := itempack.ParseProfile(r.Profile); err != nil {
		return nil, fmt.Errorf("%sprofile.json: %w", at, err)
	}
	data, err := read(at+"specset.json", true)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &r.SpecSet); err != nil {
		return nil, fmt.Errorf("%sspecset.json: %w", at, err)
	}
	if err := r.SpecSet.validate(); err != nil {
		return nil, fmt.Errorf("%sspecset.json: %w", at, err)
	}
	cfgs := []string{Dir + "/config.json"}
	if name != "" {
		cfgs = append(cfgs, at+"config.json")
	}
	for _, path := range cfgs {
		data, err := read(path, false)
		if err != nil {
			return nil, err
		}
		if data != nil {
			r.Configs = append(r.Configs, File{Path: path, Data: data})
		}
	}
	if name != "" {
		if err := r.agree(); err != nil {
			return nil, err
		}
	}
	return &r, nil
}

// agree checks, in a repository of several products, that no file names a product other than its
// folder's: the shared settings name none, the product's own files name their folder or nothing.
func (r *Repo) agree() error {
	if p := strings.TrimSpace(r.SpecSet.Product); p != "" && p != r.Product {
		return fmt.Errorf("%s/%s/specset.json names product %s: in a repository of several products the folder is the product", Dir, r.Product, p)
	}
	for _, f := range r.Configs {
		p := configProduct(f.Data)
		switch {
		case p == "":
		case f.Path == Dir+"/config.json":
			return fmt.Errorf("%s sets product %s, for every product: in a repository of several products the folder is the product (%s)", f.Path, p, strings.Join(r.Products, ", "))
		case p != r.Product:
			return fmt.Errorf("%s names product %s: in a repository of several products the folder is the product", f.Path, p)
		}
	}
	return nil
}

// ProductName is the repository's product for this item: its folder in a repository of several, else
// the one its config.json names, else its specset.json's — never the user's, whose dudes.json belongs to
// whatever product that file was written for.
func (r *Repo) ProductName() string {
	if r.Product != "" {
		return r.Product
	}
	for i := len(r.Configs) - 1; i >= 0; i-- {
		if p := configProduct(r.Configs[i].Data); p != "" {
			return p
		}
	}
	return strings.TrimSpace(r.SpecSet.Product)
}

// TestCommand is the test_command the repository's settings name — the product's own over the shared
// ones — the default of a slicing that names none.
func (r *Repo) TestCommand() []string {
	for i := len(r.Configs) - 1; i >= 0; i-- {
		var c struct {
			TestCommand []string `json:"test_command"`
		}
		if json.Unmarshal(r.Configs[i].Data, &c) == nil && len(c.TestCommand) > 0 {
			return c.TestCommand
		}
	}
	return nil
}

// Where names the folder the product's files sit in: "dudes" or "dudes/<product>".
func (r *Repo) Where() string {
	if r.Product == "" {
		return Dir
	}
	return Dir + "/" + r.Product
}

func configProduct(data []byte) string {
	var c struct {
		Product string `json:"product"`
	}
	_ = json.Unmarshal(data, &c)
	return strings.TrimSpace(c.Product)
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

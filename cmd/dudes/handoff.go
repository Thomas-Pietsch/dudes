package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/handoff"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
)

// Item handoff (F88): `dudes do <ITEM> -slice <file.json>` prepares one slice of a work item for the
// engine from the repository's tracked dudes/ folder — a worktree on its own branch, the spec roots
// pinned beside it, the run's config and contract, the freeze — and names the command that starts the
// work. Preparation is process, so the engine does it: a model call per item for it would be advisory
// and cost tokens (CORE §13).

const doSliceUsage = "usage: dudes do <ITEM> -slice <file.json> [-base <ref>] [-run]"

// sliceArgs are the arguments of `dudes do … -slice`.
type sliceArgs struct {
	item, slicing, base string
	run                 bool // -run: run the prepared slice to its first pause
}

// parseSliceArgs reads the words after `dudes do`. -slice switches intake (F84) to item handoff;
// without it the words are the prompt, unchanged — and so are -base and -run, which mean something
// only with -slice.
func parseSliceArgs(words []string) (a sliceArgs, slice bool, err error) {
	var rest []string
	for i := 0; i < len(words); i++ {
		name, val, hasVal := strings.Cut(words[i], "=")
		switch name {
		case "-slice", "--slice", "-base", "--base":
		case "-run", "--run":
			a.run = true
			continue
		default:
			rest = append(rest, words[i])
			continue
		}
		if !hasVal {
			if i+1 >= len(words) {
				return a, true, fmt.Errorf("%s needs a value — %s", name, doSliceUsage)
			}
			i++
			val = words[i]
		}
		if strings.HasSuffix(name, "slice") {
			a.slicing, slice = val, true
		} else {
			a.base = val
		}
	}
	if !slice {
		return sliceArgs{}, false, nil
	}
	if len(rest) > 1 {
		return a, true, fmt.Errorf("one item per slice, got %q — %s", strings.Join(rest, " "), doSliceUsage)
	}
	if len(rest) == 1 {
		a.item = rest[0]
	}
	return a, true, nil
}

// cmdDoSlice prepares the slice in the repository the command runs in and prints what it made and
// the next command; with -run it runs the slice to its first pause, as dudes run does. Exit 1: nothing
// was made (or it was taken back), or the run stopped on an error; 2: a usage error.
func cmdDoSlice(configPath string, a sliceArgs) int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	res, err := prepareSlice(configPath, dir, a)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	res.print(os.Stdout)
	if !a.run {
		return 0
	}
	return cmdRun(filepath.Join(dir, filepath.FromSlash(res.Config)), "", false, false, false, "")
}

// sliceResult is what a preparation made.
type sliceResult struct {
	Name, Item   string
	Product      string
	Files        string // where the product's files sit: dudes, or dudes/<product> in a repository of several
	Branch, Base string
	BaseSHA      string
	Worktree     string // relative to the repository's root
	Config       string // the run's config, likewise
	Pins         []handoff.Pin
	Tasks        []contract.Task
}

func (r *sliceResult) print(w io.Writer) {
	fmt.Fprintf(w, "prepared %s (item %s of %s, from %s/): %s on %s from %s (%s)\n", r.Name, r.Item, r.Product, r.Files, r.Worktree, r.Branch, r.Base, shortSHA(r.BaseSHA))
	for _, p := range r.Pins {
		fmt.Fprintf(w, "  spec %s pinned at %s (%s)\n", p.Root, p.Ref, shortSHA(p.Commit))
	}
	var tasks []string
	for _, t := range r.Tasks {
		if t.Expect == "fail" {
			t.ID += " (red first)"
		}
		tasks = append(tasks, t.ID)
	}
	fmt.Fprintf(w, "  tasks %s\n", strings.Join(tasks, ", "))
	fmt.Fprintf(w, "next: dudes -config %s step\n", r.Config)
}

// prepareSlice makes the slice, all or nothing: it checks everything it can before it creates
// anything, and a failure after that takes back the worktree, the branch and the pins.
func prepareSlice(configPath, dir string, a sliceArgs) (*sliceResult, error) {
	root, err := gitx.Toplevel(dir)
	if err != nil {
		return nil, err
	}
	if !handoff.SameDir(root, dir) {
		return nil, fmt.Errorf("run it from the dev repository's root (%s), not from %s", root, dir)
	}
	sl, err := handoff.ReadSlicing(a.slicing)
	if err != nil {
		return nil, err
	}
	itemArg := a.item
	switch {
	case itemArg == "":
		itemArg = sl.Item
	case sl.Item != "" && !strings.EqualFold(itemArg, sl.Item):
		return nil, fmt.Errorf("the command names %s, the slicing %s", itemArg, sl.Item)
	}
	if itemArg == "" {
		return nil, fmt.Errorf("no work item: name it on the command line or in the slicing (\"item\") — %s", doSliceUsage)
	}

	// The user's own settings and the machine's model host: before anything is created, and never
	// from the repository (CORE §13).
	user, _, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	host, err := config.ModelHost(user.Ollama.BaseURL)
	if err != nil {
		return nil, err
	}

	plan, err := handoff.NewPlan(root, sl, a.base)
	if err != nil {
		return nil, err
	}
	// the item's product and its files: in a repository of several products, the one whose work items
	// the ID belongs to — refused, with the products it knows, before anything is made
	repo, err := handoff.ReadRepo(root, plan.BaseSHA, itemArg)
	if err != nil {
		return nil, err
	}
	cfg := user
	for _, f := range repo.Configs { // the shared settings, then the product's own
		var ignored []string
		if cfg, ignored, err = cfg.Overlay(f.Data); err != nil {
			return nil, fmt.Errorf("%s: %w", f.Path, err)
		}
		for _, k := range ignored {
			why := "that is the machine's, not the repository's"
			switch {
			case k == "seal.require_signature" && user.Seal.RequireSignature:
				why = "your own dudes.json asks for signed seals, and a repository cannot waive that"
			case k == "seal.require_signature":
				why = handoff.Dir + "/config.json asks for signed seals, and a product cannot waive that"
			}
			fmt.Fprintf(os.Stderr, "dudes: warning: %s sets %s — ignored: %s\n", f.Path, k, why)
		}
	}
	product, repoTest := repo.ProductName(), repo.TestCommand()
	if product == "" {
		return nil, fmt.Errorf("%s names no product: set \"product\" in %s/specset.json or %s/config.json", handoff.Dir, handoff.Dir, handoff.Dir)
	}
	cfg.Product = product
	cfg.Ollama.BaseURL = host
	cfg.Worktree = false // the slice is the worktree; a second one per run would hide it
	cfg.Workspace = filepath.ToSlash(plan.Worktree)

	if err := plan.PlanPins(repo.SpecSet.Roots, repo.SpecSet.RootNames()); err != nil {
		return nil, err
	}
	if err := plan.Create(); err != nil {
		return nil, err
	}
	res, err := populateSlice(plan, repo, sl, cfg, itemArg, repoTest)
	if err != nil {
		for _, e := range plan.Rollback() {
			fmt.Fprintf(os.Stderr, "dudes: warning: could not take back everything: %v\n", e)
		}
		return nil, err
	}
	res.Worktree = relTo(root, plan.Worktree)
	res.Config = relTo(root, filepath.Join(plan.Worktree, ".dudes", "dudes.json"))
	res.Base, res.BaseSHA, res.Branch, res.Pins = plan.Base, plan.BaseSHA, plan.Branch, plan.Pins
	return res, nil
}

// populateSlice fills the worktree the plan made: the run's .dudes/ (profile, spec-set whose roots
// are the pins, config), the contract, then the run itself — a start commit of its own, so the run's
// notes do not land on a commit other worktrees share (the run state and the freeze are notes on
// HEAD) — init and freeze.
func populateSlice(plan *handoff.Plan, repo *handoff.Repo, sl *handoff.Slicing, cfg config.Config, itemArg string, repoTest []string) (*sliceResult, error) {
	ws := plan.Worktree
	if err := gitx.Exclude(ws, "/.dudes/", ".dudes/x"); err != nil { // the run config holds the model host
		return nil, err
	}
	dd := filepath.Join(ws, ".dudes")
	if err := os.MkdirAll(dd, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dd, "profile.json"), repo.Profile, 0o644); err != nil {
		return nil, err
	}
	if err := specset.Save(ws, repo.SpecSet.Specset(cfg.Product, ws, plan.PinPaths())); err != nil {
		return nil, err
	}
	cfgPath := filepath.Join(dd, "dudes.json")
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		return nil, err
	}
	if err := os.WriteFile(cfgPath, buf.Bytes(), 0o644); err != nil {
		return nil, err
	}

	s, ok, err := itempack.Open(ws, cfg.Product)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("%s/profile.json was not written", handoff.Dir)
	}
	item, ok := canonicalItem(s, itemArg)
	if !ok {
		return nil, fmt.Errorf("work item %s not found in the spec as pinned (outline rows are not packable)", itemArg)
	}
	sources, acc, picked, err := projectItem(cfg, s, item, sl.Acceptance)
	if err != nil {
		return nil, err
	}
	if err := checkRefs(s, item, sl.Refs); err != nil {
		return nil, err
	}
	c := contract.Project(cfg.Product, item, sources, acc, nil)
	var criteria []string
	for _, p := range picked {
		criteria = append(criteria, p.ID)
	}
	if err := sl.Apply(c, criteria, repoTest); err != nil {
		return nil, err
	}
	if err := contract.Save(ws, c); err != nil {
		return nil, err
	}

	// What the run would refuse later, refused now: an item that is not ready, a spec-set with findings.
	in, err := itemSpec(cfg)
	if err != nil {
		return nil, err
	}
	if err := freeze.RequireConsistent(ws, cfg.Product); err != nil {
		return nil, err
	}
	if err := gitx.EmptyCommit(ws, handoff.StartMessage(c, plan.Pins, in.trailers)); err != nil {
		return nil, err
	}
	if code := cmdInit(cfgPath, sl.Name(item)); code != 0 {
		return nil, fmt.Errorf("init failed")
	}
	if code := cmdFreeze(cfgPath); code != 0 {
		return nil, fmt.Errorf("freeze failed")
	}
	return &sliceResult{Name: sl.Name(item), Item: item, Product: cfg.Product, Files: repo.Where(), Tasks: c.Tasks}, nil
}

// canonicalItem finds the work item named in any case ("bk-001") and returns its ID as the backlog writes it.
func canonicalItem(s *itempack.Set, id string) (string, bool) {
	for _, it := range s.Items() {
		if it.Kind == itempack.KindItem && strings.EqualFold(it.ID, id) {
			return it.ID, true
		}
	}
	return "", false
}

// checkRefs refuses a slicing reference the item does not make: Narrow would then hand a worker an
// empty spec.
func checkRefs(s *itempack.Set, item string, refs []string) error {
	if len(refs) == 0 {
		return nil
	}
	it, _ := s.FindItem(item)
	_, byDoc := s.Targets(it)
	have := map[string]bool{}
	var list []string
	for doc, targets := range byDoc {
		for _, t := range targets {
			have[doc+":"+t] = true
			list = append(list, doc+":"+t)
		}
	}
	sort.Strings(list)
	for _, r := range refs {
		if !have[strings.Trim(strings.TrimSpace(r), "[]")] {
			return fmt.Errorf("slicing reference %s is not a reference of %s (it has %s)", r, item, nonzero(strings.Join(list, ", "), "none"))
		}
	}
	return nil
}

func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// relTo is target as a slash-separated path from base, or target itself when it has no relative form.
func relTo(base, target string) string {
	if rel, err := filepath.Rel(base, target); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(target)
}

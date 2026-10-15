package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

const itemUsage = "usage: dudes item check | list | board | pack <ID>… [-all] [-budget N] | contract <ID>"

// cmdItem is the work-item context surface (F68–F71). Exit codes: 0 clean,
// 1 findings / not ready, 2 could not run (F47: found problems ≠ failed to run).
func cmdItem(configPath string, args []string, asJSON bool) int {
	asJSON = flaggedJSON(asJSON, args)
	args = withoutJSONFlag(args)
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return 2
	}
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "dudes: "+itemUsage)
		return 2
	}
	s, code := openItems(cfg)
	if code != 0 {
		return 2
	}
	switch args[0] {
	case "check":
		f := s.Check()
		if asJSON {
			rows := make([]map[string]string, 0, len(f))
			for _, x := range f {
				rows = append(rows, map[string]string{"code": x.Code, "msg": x.Msg})
			}
			if c := emitJSON(map[string]any{"ok": len(f) == 0, "n": len(f), "findings": rows}); c != 0 {
				return 2
			}
		} else {
			for _, x := range f {
				fmt.Println(x)
			}
			fmt.Printf("consistency: %d finding(s)\n", len(f))
		}
		if len(f) > 0 {
			return 1
		}
		return 0
	case "list":
		results := acceptanceResults(cfg)
		for _, it := range s.Items() {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", it.ID, nonzero(it.Release, "-"), it.Kind, itemAC(s, it, results), it.Title)
		}
		return 0
	case "board":
		printBoard(os.Stdout, s, acceptanceResults(cfg))
		return 0
	case "pack":
		return itemPack(cfg, s, args[1:])
	case "contract":
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, "dudes: usage: dudes item contract <ID>")
			return 2
		}
		return itemContract(cfg, s, args[1])
	default:
		fmt.Fprintln(os.Stderr, "dudes: "+itemUsage)
		return 2
	}
}

func openItems(cfg config.Config) (*itempack.Set, int) {
	s, ok, err := itempack.Open(cfg.Workspace, cfg.Product)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return nil, 1
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "dudes: no %s — copy the instance profile (e.g. examples/sample/profile.json)\n", itempack.ProfilePath())
		return nil, 1
	}
	return s, 0
}

// acceptanceResults are the per-test results of the acceptance runs recorded in the dev workspace
// (testrun.Acceptance: the newest run per test); empty when none is recorded.
func acceptanceResults(cfg config.Config) map[string]string {
	acc, err := testrun.Acceptance(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: acceptance runs not read: %v\n", err)
	}
	results := map[string]string{}
	for name, o := range acc {
		results[name] = o.Result
	}
	return results
}

// acceptanceID is the contract ID of an item's criterion ("BK-001", "AC-2" → "bk-001-ac-2"); its
// test is spec.AcceptanceTestName of it.
func acceptanceID(item, ac string) string { return spec.AcceptanceID(item, ac) }

// acTally counts an item's acceptance tests that pass in the recorded acceptance run. recorded is
// false when none of them ran there: the item's status is then the QR's alone.
type acTally struct {
	green, total int
	recorded     bool
}

func (t acTally) String() string {
	if !t.recorded {
		return "-"
	}
	return fmt.Sprintf("AC %d/%d", t.green, t.total)
}

func itemAC(s *itempack.Set, it itempack.Item, results map[string]string) acTally {
	var t acTally
	for _, c := range s.Acceptance(it) {
		t.total++
		res, ok := results[spec.AcceptanceTestName(acceptanceID(it.ID, c.ID))]
		t.recorded = t.recorded || ok
		if res == testrun.Pass {
			t.green++
		}
	}
	return t
}

// itemStatus is the derived status (F70) with engine evidence: an item its QR DoD IDs mark done is
// done only when all its acceptance tests pass in the recorded acceptance run. Without a recorded
// run the status is the QR's, as before.
func itemStatus(s *itempack.Set, it itempack.Item, idx map[string][]string, results map[string]string) (string, acTally) {
	st, _, _ := s.DeriveStatus(it, idx)
	t := itemAC(s, it, results)
	if st == itempack.StatusDone && t.recorded && t.green < t.total {
		st = itempack.StatusOpen
	}
	return st, t
}

func printBoard(w io.Writer, s *itempack.Set, results map[string]string) {
	items := s.Items()
	idx := s.StatusIndex()
	rels := map[string]bool{}
	for _, it := range items {
		if it.Release != "" {
			rels[it.Release] = true
		}
	}
	var order []string
	for r := range rels {
		order = append(order, r)
	}
	sort.Strings(order)
	for _, rel := range order {
		fmt.Fprintf(w, "## %s\n\n| Item | Title | DoD | AC | Status |\n|---|---|---|---|---|\n", rel)
		for _, it := range items {
			if it.Release != rel {
				continue
			}
			st, ac := itemStatus(s, it, idx, results)
			title := []rune(it.Title)
			if len(title) > 60 {
				title = title[:60]
			}
			fmt.Fprintf(w, "| %s | %s | %d | %s | %s |\n", it.ID, string(title), len(s.DoDIDs(it)), ac, st)
		}
		fmt.Fprintln(w)
	}
}

func itemPack(cfg config.Config, s *itempack.Set, args []string) int {
	budget := cfg.ContextBudget // F73 default; -budget overrides
	all := false
	var ids []string
	for i := 0; i < len(args); i++ {
		switch a := args[i]; a {
		case "-all", "--all":
			all = true
		case "-budget", "--budget":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "dudes: -budget needs a number of characters")
				return 2
			}
			i++
			n, err := strconv.Atoi(args[i])
			if err != nil || n < 1 {
				fmt.Fprintf(os.Stderr, "dudes: bad -budget %q\n", args[i])
				return 2
			}
			budget = n
		default:
			ids = append(ids, a)
		}
	}
	if all {
		ids = nil
		for _, it := range s.Items() {
			if it.Kind == itempack.KindItem {
				ids = append(ids, it.ID)
			}
		}
	}
	if len(ids) == 0 {
		fmt.Fprintln(os.Stderr, "dudes: "+itemUsage)
		return 2
	}
	rc := 0
	for _, id := range ids {
		p, err := s.Build(id, budget)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: not ready: %v\n", id, err)
			rc = 2
			continue
		}
		rel, err := itempack.Write(cfg.Workspace, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		m := p.Manifest
		over := ""
		if m.OverBudget {
			over = "  OVER BUDGET"
		}
		fmt.Printf("%s  refs=%d  chars=%d  status=%s%s\n", rel, len(m.References), m.Sizes["PACK.md"], m.Status, over)
		if len(m.References) == 0 {
			fmt.Fprintf(os.Stderr, "warning: %s has no [[…]] references — the pack holds only rules and task\n", id)
		}
	}
	return rc
}

// itemContract projects .dudes/contract.json from one work item (F71).
func itemContract(cfg config.Config, s *itempack.Set, id string) int {
	sources, acc, _, err := projectItem(cfg, s, id, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	prev, _, err := contract.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	c := contract.Project(cfg.Product, id, sources, acc, prev)
	if err := contract.Save(cfg.Workspace, c); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	fmt.Printf("wrote %s id=%s item=%s  sources=%d  acceptance=%d\n", contract.Path(), c.ID, id, len(c.Sources), len(c.Acceptance))
	if prev != nil && (len(c.Constraints) > 0 || len(c.Tasks) > 0) {
		fmt.Printf("kept %d constraint(s) and %d task(s) from the previous contract %s — edit if they no longer apply\n", len(c.Constraints), len(c.Tasks), prev.ID)
	}
	return 0
}

// projectItem is what a work item contributes to its contract (F71): the documents it references as
// sources and its behavior criteria as acceptance — all of them, or only those named in only (a
// slice's, F88: their IDs as the item writes them, in any case; the result follows the order of
// only). picked are the criteria taken, as the item writes them.
func projectItem(cfg config.Config, s *itempack.Set, id string, only []string) (sources []contract.Source, acc []contract.Acceptance, picked []itempack.Criterion, err error) {
	it, ok := s.FindItem(id)
	if !ok {
		return nil, nil, nil, fmt.Errorf("work item %s not found (outline rows are not packable)", id)
	}
	if _, err := s.Build(id, 0); err != nil {
		return nil, nil, nil, fmt.Errorf("%s not ready: %w", id, err)
	}
	set, _, err := specset.Load(cfg.Workspace, cfg.Product)
	if err != nil {
		return nil, nil, nil, err
	}
	backlog := s.Profile.Backlog.Doc
	addSource := func(doc, section string) {
		e, ok := set.Find(doc)
		if !ok {
			return
		}
		sources = append(sources, contract.Source{ID: strings.ToLower(doc), Root: e.Root, Path: e.Path, Section: section})
	}
	addSource(backlog, id)
	docs, byDoc := s.Targets(it)
	for _, d := range docs {
		if d == backlog {
			continue
		}
		addSource(d, strings.Join(byDoc[d], " "))
	}
	picked = s.Acceptance(it)
	if len(only) > 0 {
		if picked, err = pickCriteria(id, picked, only); err != nil {
			return nil, nil, nil, err
		}
	}
	for _, c := range picked {
		acc = append(acc, contract.Acceptance{
			ID: acceptanceID(id, c.ID), Kind: "behavior", Source: strings.ToLower(backlog), Text: c.ID + " " + c.Text,
		})
	}
	return sources, acc, picked, nil
}

// pickCriteria are the item's criteria named in only, in that order; a name the item does not have
// is an error that lists the ones it has.
func pickCriteria(item string, have []itempack.Criterion, only []string) ([]itempack.Criterion, error) {
	var out []itempack.Criterion
	for _, want := range only {
		found := false
		for _, c := range have {
			if strings.EqualFold(c.ID, want) {
				out, found = append(out, c), true
				break
			}
		}
		if !found {
			var ids []string
			for _, c := range have {
				ids = append(ids, c.ID)
			}
			return nil, fmt.Errorf("work item %s has no criterion %s (it has %s)", item, want, nonzero(strings.Join(ids, ", "), "none"))
		}
	}
	return out, nil
}

// itemInput is the worker's spec input for an item-bound contract.
type itemInput struct {
	rel, body string // PACK.md path and text
	trailers  string // git trailers: item + source hashes (git is the archive)
}

// itemSpec renders the item pack when the contract is item-bound (F71).
// It is the worker's spec input instead of the whole PRD. nil = not item-bound.
// An item whose references do not resolve is not ready and stops the step;
// a pack that does not fit context_budget after trimming stops it too (F73).
func itemSpec(cfg config.Config) (*itemInput, error) {
	con, cok, err := contract.Load(cfg.Workspace)
	if err != nil || !cok || con.Item == "" {
		return nil, err
	}
	s, pok, err := itempack.OpenCommitted(cfg.Workspace, cfg.Product) // the run's spec: committed
	if err != nil {
		return nil, err
	}
	if !pok {
		return nil, fmt.Errorf("contract item %s needs %s", con.Item, itempack.ProfilePath())
	}
	p, err := s.Build(con.Item, cfg.ContextBudget)
	if err != nil {
		return nil, fmt.Errorf("item %s not ready: %w", con.Item, err)
	}
	m := p.Manifest
	if m.OverBudget {
		return nil, fmt.Errorf("item %s: pack is %d chars after trimming, over context_budget %d — use a larger worker or split the item", con.Item, m.Sizes["PACK.md"], cfg.ContextBudget)
	}
	if trimmed := countTrimmed(m.References); trimmed > 0 {
		fmt.Printf("context_budget %d: %d reference(s) shortened to fit\n", cfg.ContextBudget, trimmed)
	}
	dir, err := itempack.Write(cfg.Workspace, p)
	if err != nil {
		return nil, err
	}
	return &itemInput{rel: dir + "/PACK.md", body: p.Files["PACK.md"], trailers: p.Trailers()}, nil
}

func countTrimmed(refs []itempack.Ref) int {
	n := 0
	for _, r := range refs {
		if r.Trimmed {
			n++
		}
	}
	return n
}

func withoutJSONFlag(args []string) []string {
	out := args[:0:0]
	for _, a := range args {
		if a != "-json" && a != "--json" {
			out = append(out, a)
		}
	}
	return out
}

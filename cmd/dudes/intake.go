package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/goals"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// Intake maps a prompt onto what the engine can run — deterministically: a prompt naming one work
// item binds it and starts its run at test_authoring; a prompt that is a documented make goal runs
// that goal; anything else gets the menu (open items, goals, a new idea's full lifecycle) and starts
// nothing. With -pick the text model chooses one entry of that menu for free text: the engine
// validates the choice and prints it; the human confirms by running it — the model picks, it never
// invents a flow.
//
// With -slice the work item comes with the role's slicing: `dudes do <ITEM> -slice <file.json>`
// prepares the slice for the engine — worktree, pinned spec, contract, freeze (item handoff, F88).
func cmdDo(configPath string, words []string) int {
	if a, slice, err := parseSliceArgs(words); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 2
	} else if slice {
		return cmdDoSlice(configPath, a)
	}
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	pick := len(words) > 0 && (words[0] == "-pick" || words[0] == "--pick")
	if pick {
		words = words[1:]
	}
	text := strings.TrimSpace(strings.Join(words, " "))
	if text == "" {
		fmt.Fprintln(os.Stderr, `dudes: usage: dudes do <prompt> — "BK-001", a make goal, or free text for the menu; dudes do <ITEM> -slice <file.json> prepares a slice (F88)`)
		return 2
	}
	s, ok, err := itempack.Open(cfg.Workspace, cfg.Product)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		s = nil // no profile: no items, goals and the lifecycle remain
	}
	ids := namedItems(s, text)
	if len(ids) == 1 {
		return bindItem(cfg, configPath, s, ids[0], text)
	}
	if len(ids) == 0 && !strings.ContainsAny(text, " \t") && goals.Has(cfg.Workspace, text) {
		fmt.Printf("intake: goal %s\n", text)
		cmd := exec.Command("make", text)
		cmd.Dir, cmd.Stdout, cmd.Stderr = cfg.Workspace, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: make %s: %v\n", text, err)
			return 1
		}
		return 0
	}
	menu := intakeMenu(cfg, s, text)
	if len(ids) > 1 {
		fmt.Printf("intake: %q names %d items (%s) — one item per run\n", text, len(ids), strings.Join(ids, ", "))
	} else if pick {
		return pickFromMenu(cfg, menu, text)
	} else {
		fmt.Printf("intake: %q names no work item and no goal — pick one:\n", text)
	}
	printMenu(menu)
	return 2
}

// menuEntry is one thing intake can start: its key (an item ID, a goal, or "new") and its line.
type menuEntry struct{ key, line string }

func printMenu(menu []menuEntry) {
	for _, e := range menu {
		fmt.Println(e.line)
	}
}

// namedItems are the work items (not outline rows) whose ID the text names, case-insensitively.
func namedItems(s *itempack.Set, text string) []string {
	if s == nil {
		return nil
	}
	var out []string
	for _, it := range s.Items() {
		if it.Kind == itempack.KindOutline {
			continue
		}
		if regexp.MustCompile(`(?i)(^|[^\w-])` + regexp.QuoteMeta(it.ID) + `($|[^\w-])`).MatchString(text) {
			out = append(out, it.ID)
		}
	}
	return out
}

// bindItem projects the item's contract (as `dudes item contract`) and starts its run — unless a run
// is under way: then intake reports it and changes nothing.
func bindItem(cfg config.Config, configPath string, s *itempack.Set, id, text string) int {
	st, running, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if running && st.Status != runstate.StatusSealed && st.Status != runstate.StatusFailed && st.Status != runstate.StatusEnded {
		con, _, _ := contract.Load(cfg.Workspace)
		if con != nil && con.Item == id {
			fmt.Printf("intake: %s is bound, run %s at %s — next: dudes step\n", id, st.RunID, st.Phase)
			return 0
		}
		fmt.Fprintf(os.Stderr, "dudes: run %s is under way at %s — finish it or `dudes reset -yes` before binding %s\n", st.RunID, st.Phase, id)
		return 1
	}
	fmt.Printf("intake: work item %s\n", id)
	if code := itemContract(cfg, s, id); code != 0 {
		return code
	}
	return cmdInit(configPath, text)
}

// intakeMenu is what intake can start: open items with their derived status, the goals, a new idea.
func intakeMenu(cfg config.Config, s *itempack.Set, text string) []menuEntry {
	var out []menuEntry
	if s != nil {
		idx := s.StatusIndex()
		for _, it := range s.Items() {
			if it.Kind == itempack.KindOutline {
				continue
			}
			if status, _, _ := s.DeriveStatus(it, idx); status != itempack.StatusDone {
				out = append(out, menuEntry{it.ID, fmt.Sprintf("  item  dudes do %-12s %s (%s)", it.ID, it.Title, status)})
			}
		}
	}
	gs, _ := goals.List(cfg.Workspace)
	for _, g := range gs {
		out = append(out, menuEntry{g.Name, fmt.Sprintf("  goal  dudes do %-12s %s", g.Name, g.Desc)})
	}
	return append(out, menuEntry{"new", fmt.Sprintf("  new   dudes init -prompt %q   (full lifecycle from the idea phase)", text)})
}

// pickFromMenu asks the text model for one menu key and prints the validated choice; nothing starts.
func pickFromMenu(cfg config.Config, menu []menuEntry, text string) int {
	var b strings.Builder
	b.WriteString("Route the request to exactly one entry of the menu. Reply with the entry's key only, nothing else. If no entry fits, reply NONE.\n\nRequest: " + text + "\n\nMenu:\n")
	for _, e := range menu {
		fmt.Fprintf(&b, "- key `%s` — %s\n", e.key, strings.TrimSpace(e.line))
	}
	out, err := cfg.OllamaClient().Generate(context.Background(), b.String())
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: intake pick: %v\n", err)
		return 1
	}
	answer := strings.Trim(strings.TrimSpace(out), "`.'\" ")
	for _, e := range menu {
		if strings.EqualFold(answer, e.key) {
			fmt.Printf("intake: the model picks — run it to confirm:\n%s\n", e.line)
			return 2
		}
	}
	if i := strings.IndexByte(answer, '\n'); i >= 0 {
		answer = answer[:i]
	}
	fmt.Printf("intake: the model named no entry of the menu (%q) — pick one:\n", answer)
	printMenu(menu)
	return 2
}

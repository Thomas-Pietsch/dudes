package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/bench"
	"github.com/Thomas-Pietsch/dudes/internal/compact"
	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/doctor"
	"github.com/Thomas-Pietsch/dudes/internal/events"
	"github.com/Thomas-Pietsch/dudes/internal/extract"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/gate"
	"github.com/Thomas-Pietsch/dudes/internal/gitx"
	"github.com/Thomas-Pietsch/dudes/internal/humangate"
	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/ollama"
	"github.com/Thomas-Pietsch/dudes/internal/packer"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
	"github.com/Thomas-Pietsch/dudes/internal/pi"
	"github.com/Thomas-Pietsch/dudes/internal/product"
	"github.com/Thomas-Pietsch/dudes/internal/recipe"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
	"github.com/Thomas-Pietsch/dudes/internal/skill"
	"github.com/Thomas-Pietsch/dudes/internal/spec"
	"github.com/Thomas-Pietsch/dudes/internal/specset"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
	"github.com/Thomas-Pietsch/dudes/internal/version"
	"github.com/Thomas-Pietsch/dudes/internal/worktree"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("dudes", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	configPath := fs.String("config", "dudes.json", "path to engine config JSON")
	showVersion := fs.Bool("version", false, "print version and exit")
	prompt := fs.String("prompt", "", "user prompt when starting a run (init)")
	useStub := fs.Bool("stub", false, "use stub artifacts / pack (no Ollama/Repomix)")
	doFix := fs.Bool("fix", false, "with test: on failure, re-invoke Ollama with test log only")
	doPack := fs.Bool("pack", false, "with step/run: refresh phase-scoped context pack before worker")
	skipTest := fs.Bool("skip-test", false, "with step/run: do not run test_command after implementation")
	wantJSON := fs.Bool("json", false, "machine-readable JSON on stdout (status, check, events, bench)")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	// a -config that names no file is a typo (a slice that is not there), not a wish for the defaults
	named := false
	fs.Visit(func(f *flag.Flag) { named = named || f.Name == "config" })
	if _, err := os.Stat(*configPath); named && err != nil {
		fmt.Fprintf(os.Stderr, "dudes: -config %s: %v\n", *configPath, err)
		return 1
	}

	if *showVersion {
		fmt.Println(version.Version)
		return 0
	}

	rest := fs.Args()
	if len(rest) == 0 {
		return cmdStatus(*configPath, *wantJSON)
	}

	switch rest[0] {
	case "version":
		fmt.Println(version.Version)
		return 0
	case "help", "-h":
		printUsage(fs)
		return 0
	case "status":
		return cmdStatus(*configPath, flaggedJSON(*wantJSON, rest[1:]))
	case "init":
		p := *prompt
		args := rest[1:]
		if len(args) >= 2 && (args[0] == "-prompt" || args[0] == "--prompt") {
			p = args[1]
			args = args[2:]
		}
		if len(args) > 0 {
			p = strings.Join(args, " ")
		}
		return cmdInit(*configPath, p)
	case "do":
		return cmdDo(*configPath, rest[1:])
	case "reset":
		yes := false
		for _, a := range rest[1:] {
			if a == "-yes" || a == "--yes" {
				yes = true
			}
		}
		return cmdReset(*configPath, yes)
	case "run":
		stub := *useStub
		withPack := *doPack
		noTest := *skipTest
		p := *prompt
		until := ""
		args := rest[1:]
		out := args[:0]
		for i := 0; i < len(args); i++ {
			a := args[i]
			switch a {
			case "-stub", "--stub":
				stub = true
			case "-pack", "--pack":
				withPack = true
			case "-skip-test", "--skip-test":
				noTest = true
			case "-prompt", "--prompt":
				if i+1 < len(args) {
					i++
					p = args[i]
				}
			case "-until", "--until":
				if i+1 < len(args) {
					i++
					until = args[i]
				}
			default:
				out = append(out, a)
			}
		}
		if len(out) > 0 && p == "" {
			p = strings.Join(out, " ")
		}
		return cmdRun(*configPath, p, stub, withPack, noTest, until)
	case "advance":
		return cmdAdvance(*configPath)
	case "step":
		stub := *useStub
		withPack := *doPack
		noTest := *skipTest
		args := rest[1:]
		for _, a := range args {
			switch a {
			case "-stub", "--stub":
				stub = true
			case "-pack", "--pack":
				withPack = true
			case "-skip-test", "--skip-test":
				noTest = true
			}
		}
		return cmdStep(*configPath, stub, withPack, noTest)
	case "pack":
		stub := *useStub
		args := rest[1:]
		for _, a := range args {
			if a == "-stub" || a == "--stub" {
				stub = true
			}
		}
		return cmdPack(*configPath, stub)
	case "check":
		return cmdCheck(*configPath, flaggedJSON(*wantJSON, rest[1:]))
	case "doctor":
		return cmdDoctor(*configPath)
	case "events":
		n := 20
		args := rest[1:]
		for i := 0; i < len(args); i++ {
			a := args[i]
			if (a == "-n" || a == "--n") && i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &n)
			}
		}
		return cmdEvents(*configPath, n, flaggedJSON(*wantJSON, args))
	case "bench":
		return cmdBench(flaggedJSON(*wantJSON, rest[1:]))
	case "freeze":
		return cmdFreeze(*configPath)
	case "test":
		fix := *doFix
		args := rest[1:]
		for _, a := range args {
			if a == "-fix" || a == "--fix" {
				fix = true
			}
		}
		return cmdTest(*configPath, fix)
	case "seal":
		release := ""
		if len(rest) > 1 {
			release = rest[1]
		}
		return cmdSeal(*configPath, release)
	case "reject":
		return cmdReject(*configPath, strings.Join(rest[1:], " "))
	case "verify":
		return cmdVerify(*configPath, rest[1:])
	case "phases":
		return cmdPhases()
	case "specset":
		return cmdSpecset(*configPath, rest[1:])
	case "contract":
		return cmdContract(*configPath)
	case "evidence":
		return cmdEvidenceShow(*configPath)
	case "decide":
		return cmdDecide(*configPath, rest[1:])
	case "redo":
		return cmdRedo(*configPath, rest[1:])
	case "land":
		return cmdLand(rest[1:])
	case "item":
		return cmdItem(*configPath, rest[1:], *wantJSON)
	default:
		fmt.Fprintf(os.Stderr, "dudes: unknown command %q\n", rest[0])
		printUsage(fs)
		return 2
	}
}

func loadCfg(configPath string) (config.Config, int) {
	cfg, _, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return config.Config{}, 1
	}
	return cfg, 0
}

func flaggedJSON(global bool, args []string) bool {
	if global {
		return true
	}
	for _, a := range args {
		if a == "-json" || a == "--json" {
			return true
		}
	}
	return false
}

func emitJSON(v any) int {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false) // a diff's `&&` and `<-` stay as they are, not escaped for HTML
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: json: %v\n", err)
		return 1
	}
	return 0
}

func cmdStatus(configPath string, asJSON bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	out := map[string]any{
		"version":   version.Version,
		"workspace": cfg.Workspace,
		"product":   cfg.Product,
		"ollama":    nonzero(cfg.Ollama.BaseURL, "http://127.0.0.1:11434") + " / " + nonzero(cfg.Ollama.Model, "qwen2.5-coder:7b"),
		"aider":     nonzero(cfg.Aider.Bin, "aider") + " / " + nonzero(cfg.Aider.Model, "(default)"),
		"packer":    nonzero(cfg.Packer.Bin, "repomix"),
		"test_cmd":  cfg.TestCommand,
		"events":    events.Path(),
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		out["run"] = nil
		if asJSON {
			return emitJSON(out)
		}
		fmt.Printf("dudes %s\n", version.Version)
		fmt.Printf("workspace: %s\n", cfg.Workspace)
		fmt.Printf("product:   %s\n", cfg.Product)
		fmt.Println("run:       (none — run: dudes init [-prompt \"...\"])")
		return 0
	}
	run := map[string]any{
		"run_id":  st.RunID,
		"phase":   st.Phase,
		"status":  st.Status,
		"cycle":   st.Cycle,
		"prompt":  st.Prompt,
		"state":   runstate.Path(cfg.Workspace),
		"updated": st.UpdatedAt.Format(time.RFC3339),
	}
	if st.Feedback != "" {
		run["feedback"] = st.Feedback
	}
	if st.TaskID != "" {
		run["task_id"] = st.TaskID
	}
	if len(st.Approved) > 0 {
		run["approved"] = st.Approved
	}
	if st.Escalation != "" {
		run["escalation"] = st.Escalation
		run["answer"] = escalationAnswerCmd(configPath, cfg, st)
	}
	done, total := contractProgress(cfg, st)
	if total > 0 {
		run["tasks_done"], run["tasks_total"] = done, total
	}
	if rel, ok := gate.SpecPath(cfg.Product); ok {
		run["spec"] = rel
		run["spec_exists"] = artifact.Exists(cfg.Workspace, rel)
	}
	if fr, ok, _ := freeze.Load(cfg.Workspace); ok {
		run["freeze_sha"] = trimSHA(fr.SHA256)
	}
	if rel, ok := artifact.Path(cfg.Product, st.Phase); ok {
		run["artifact"] = rel
		run["artifact_exists"] = artifact.Exists(cfg.Workspace, rel)
	}
	if _, rok, err := recipe.Load(cfg.Workspace, st.Phase); err == nil && rok {
		run["recipe"] = recipe.Path(st.Phase)
	}
	pk := review(configPath, cfg, st) // F88: what a task's reviewer decides on
	if pk != nil {
		run["review"] = pk
	}
	out["run"] = run
	if asJSON {
		return emitJSON(out)
	}

	fmt.Printf("dudes %s\n", version.Version)
	fmt.Printf("workspace: %s\n", cfg.Workspace)
	fmt.Printf("product:   %s\n", cfg.Product)
	fmt.Printf("ollama:    %s\n", out["ollama"])
	fmt.Printf("aider:     %s\n", out["aider"])
	pc := cfg.PackerRunConfig().Resolve()
	fmt.Printf("packer:    %s → %s\n", nonzero(cfg.Packer.Bin, "repomix"), pc.Output)
	fmt.Printf("test_cmd:  %v\n", cfg.TestCommand)
	fmt.Printf("run_id:    %s\n", st.RunID)
	fmt.Printf("phase:     %s\n", st.Phase)
	fmt.Printf("status:    %s\n", st.Status)
	if st.Cycle > 0 {
		fmt.Printf("cycle:     %d\n", st.Cycle)
	}
	if st.Prompt != "" {
		fmt.Printf("prompt:    %s\n", st.Prompt)
	}
	if st.Feedback != "" {
		fmt.Printf("feedback:  %s\n", st.Feedback)
	}
	if st.Phase == phase.HumanDecision {
		fmt.Println("human:     waiting — dudes seal | dudes reject <text>")
	}
	if st.Status == runstate.StatusWaitingDecision {
		fmt.Printf("decision:  waiting — dudes decide %s\n", st.TaskID)
		if st.Escalation != "" { // the worker's question, and how to answer it (F74)
			fmt.Printf("escalation: %s\n", st.Escalation)
			fmt.Printf("answer:    %s\n", escalationAnswerCmd(configPath, cfg, st))
		}
	}
	if st.TaskID != "" {
		fmt.Printf("task:      %s\n", st.TaskID)
	}
	if total > 0 {
		fmt.Printf("tasks:     %d/%d done\n", done, total)
	}
	if rel, ok := gate.SpecPath(cfg.Product); ok {
		fmt.Printf("spec:      %s", rel)
		if artifact.Exists(cfg.Workspace, rel) {
			fmt.Print(" (exists)")
		}
		fmt.Println()
	}
	if fr, ok, _ := freeze.Load(cfg.Workspace); ok {
		fmt.Printf("freeze:    %s sha=%s…\n", fr.FrozenAt.Format(time.RFC3339), trimSHA(fr.SHA256))
	} else {
		fmt.Println("freeze:    (none)")
	}
	fmt.Printf("events:    %s\n", events.Path())
	if rel, ok := artifact.Path(cfg.Product, st.Phase); ok {
		fmt.Printf("artifact:  %s", rel)
		if artifact.Exists(cfg.Workspace, rel) {
			fmt.Print(" (exists)")
		}
		fmt.Println()
	}
	if _, rok, err := recipe.Load(cfg.Workspace, st.Phase); err == nil && rok {
		fmt.Printf("recipe:    %s\n", recipe.Path(st.Phase))
	}
	fmt.Printf("updated:   %s\n", st.UpdatedAt.Format(time.RFC3339))
	fmt.Printf("state:     %s\n", runstate.Path(cfg.Workspace))
	if pk != nil {
		pk.print(os.Stdout)
	}
	return 0
}

// contractProgress counts the contract's tasks the run has completed, over all phases (F65).
// No contract or no tasks → 0, 0.
func contractProgress(cfg config.Config, st *runstate.State) (done, total int) {
	con, ok, err := contract.Load(cfg.Workspace)
	if err != nil || !ok || con == nil {
		return 0, 0
	}
	for _, t := range con.Tasks {
		if st.HasCompleted(t.ID) {
			done++
		}
	}
	return done, len(con.Tasks)
}

func nonzero(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func cmdInit(configPath, prompt string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	reg := phase.Default()
	id := fmt.Sprintf("run-%d", time.Now().Unix())
	st := runstate.NewRun(id, prompt, reg)
	// A bound item is already specified: idea, strategy and spec belong to the spec repo (§2.4).
	switch {
	case gate.ItemBound(cfg.Workspace) && gate.NamesTests(cfg.Workspace):
		// The slice's tests live where the language needs them and its red-first task writes them
		// (F88): no placeholder per criterion, so no test_authoring.
		st.Phase = phase.Implementation
		fmt.Println("item-bound, tests named by the contract: starting at implementation, where its red-first task writes them")
	case gate.ItemBound(cfg.Workspace):
		st.Phase = phase.TestAuthoring
		fmt.Println("item-bound: idea, strategy and spec are the spec repo's — starting at test_authoring")
	}
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "init", st.Phase, st.Prompt)
	if _, ok, err := specset.Load(cfg.Workspace, cfg.Product); err == nil && !ok {
		_ = specset.Save(cfg.Workspace, specset.EngineSet(cfg.Product))
	}
	fmt.Printf("initialized %s at phase %s\n", st.RunID, st.Phase)
	fmt.Printf("state: %s\n", runstate.Path(cfg.Workspace))
	fmt.Printf("specset: %v\n", product.SpecSet(cfg.Product))
	return 0
}

func cmdReset(configPath string, yes bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Println("reset: no run state")
		return 0
	}
	if !yes {
		fmt.Fprintf(os.Stderr, "dudes: will clear run %s (phase %s). Re-run with -yes to confirm.\n", st.RunID, st.Phase)
		return 2
	}
	runID := st.RunID
	if err := runstate.Clear(cfg.Workspace); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, runID, "reset", st.Phase, "run state reset")
	fmt.Printf("reset: run %s ended — reset note in %s (earlier states stay in its history; events/freeze/packs kept)\n", runID, runstate.Path(cfg.Workspace))
	return 0
}

func cmdAdvance(configPath string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes init")
		return 1
	}
	if st.Status == runstate.StatusWaitingDecision {
		fmt.Fprintf(os.Stderr, "dudes: waiting for decision — run: dudes decide %s\n", st.TaskID)
		return 1
	}
	if st.Phase == phase.HumanDecision {
		fmt.Fprintln(os.Stderr, "dudes: waiting for human — run: dudes seal | dudes reject <text>")
		return 1
	}
	if st.Phase == phase.Sealed {
		fmt.Fprintln(os.Stderr, "dudes: already sealed")
		return 1
	}
	if st.Status == runstate.StatusEnded {
		fmt.Fprintln(os.Stderr, "dudes: "+endedRun(cfg, st))
		return 1
	}
	return advanceLocked(cfg, st)
}

func cmdStep(configPath string, stub, withPack, skipTest bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes init")
		return 1
	}
	return stepOnce(cfg, st, stub, withPack, skipTest)
}

func cmdRun(configPath, prompt string, stub, withPack, skipTest bool, until string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	var untilPhase phase.ID
	if until != "" {
		untilPhase = phase.ID(until)
		if !phase.Default().Contains(untilPhase) {
			fmt.Fprintf(os.Stderr, "dudes: unknown -until phase %q (see: dudes phases)\n", until)
			return 2
		}
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		if prompt == "" {
			fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes run -prompt \"...\" (or dudes init)")
			return 1
		}
		if code := cmdInit(configPath, prompt); code != 0 {
			return code
		}
		st, ok, err = runstate.Load(cfg.Workspace)
		if err != nil || !ok {
			fmt.Fprintf(os.Stderr, "dudes: failed to load state after init: %v\n", err)
			return 1
		}
	} else if prompt != "" && st.Prompt == "" {
		st.Prompt = prompt
		_ = runstate.Save(cfg.Workspace, st)
	}

	const maxSteps = 64
	for i := 0; i < maxSteps; i++ {
		st, ok, err = runstate.Load(cfg.Workspace)
		if err != nil || !ok {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		if untilPhase != "" && st.Phase == untilPhase {
			fmt.Printf("run: stopped before %s (-until)\n", untilPhase)
			return 0
		}
		if st.Status == runstate.StatusWaitingDecision {
			fmt.Printf("run: waiting for decision — dudes decide %s\n", st.TaskID)
			return 0
		}
		if st.Phase == phase.HumanDecision {
			if findings := gate.Converge(cfg.Workspace, cfg.Product); len(findings) > 0 {
				fmt.Fprintln(os.Stderr, "dudes: converge failed at human_decision:")
				for _, f := range findings {
					fmt.Fprintf(os.Stderr, "  - %s\n", f.Error())
				}
				return 1
			}
			fmt.Println("run: waiting for human — dudes seal | dudes reject <text>")
			return 0
		}
		if st.Phase == phase.Sealed {
			fmt.Println("run: already sealed")
			return 0
		}
		if st.Status == runstate.StatusEnded {
			fmt.Println("run: " + endedRun(cfg, st))
			return 0
		}
		fmt.Printf("run: step %d at phase %s\n", i+1, st.Phase)
		if code := stepOnce(cfg, st, stub, withPack, skipTest); code != 0 {
			return code
		}
	}
	fmt.Fprintln(os.Stderr, "dudes: run exceeded max steps — stopping")
	return 1
}

func stepOnce(cfg config.Config, st *runstate.State, stub, withPack, skipTest bool) int {
	if st.Status == runstate.StatusWaitingDecision {
		fmt.Fprintf(os.Stderr, "dudes: waiting for decision — run: dudes decide %s\n", st.TaskID)
		return 1
	}
	if st.Phase == phase.HumanDecision {
		fmt.Fprintln(os.Stderr, "dudes: waiting for human — run: dudes seal | dudes reject <text>")
		return 1
	}
	if st.Phase == phase.Sealed {
		fmt.Fprintln(os.Stderr, "dudes: already sealed")
		return 1
	}
	// An item slice ends with its last task (F88): its QR is the spec repository's, and a later phase
	// would commit the engine's own QR and docs into the branch the role merges.
	if st.Status == runstate.StatusEnded {
		fmt.Println(endedRun(cfg, st))
		return 0
	}

	con, _, err := contract.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: contract: %v\n", err)
		return 1
	}
	if firstTaskInPhase(con, st) {
		events.Log(cfg.Workspace, st.RunID, "phase_enter", st.Phase, "")
		if err := gate.CheckEnter(cfg.Workspace, cfg.Product, st.Phase); err != nil {
			events.Log(cfg.Workspace, st.RunID, "gate_fail", st.Phase, err.Error())
			fmt.Fprintf(os.Stderr, "dudes: gate: %v\n", err)
			return 1
		}
	}

	rec, hasRecipe, err := recipe.Load(cfg.Workspace, st.Phase)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: recipe: %v\n", err)
		return 1
	}
	taskOverride := ""
	discipline := ""
	tools := []string{}
	worker := ""
	if hasRecipe {
		if err := applyGoals(cfg, rec); err != nil {
			events.Log(cfg.Workspace, st.RunID, "gate_fail", st.Phase, err.Error())
			fmt.Fprintf(os.Stderr, "dudes: gate: %v\n", err)
			return 1
		}
		taskOverride = rec.TaskOverride()
		tools = rec.Tools
		worker = rec.Worker
		if rec.Pack {
			withPack = true
		}
		if rec.Worker == "stub" {
			stub = true
		}
		d, err := loadDiscipline(cfg, rec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		discipline = d
		fmt.Printf("recipe: %s (worker=%s)\n", recipe.Path(st.Phase), nonzero(rec.Worker, "default"))
		events.Log(cfg.Workspace, st.RunID, "recipe", st.Phase, recipe.Path(st.Phase))
	}

	if withPack {
		if code := runPack(cfg, st.Phase, stub); code != 0 {
			return code
		}
	}

	if con != nil && len(con.TasksFor(st.Phase)) > 0 {
		return stepContractTask(cfg, st, con, rec, hasRecipe, stub, skipTest, discipline)
	}

	packCtx := loadPackContext(cfg)
	if code := executePhaseWork(cfg, st, stub, skipTest, packCtx, taskOverride, discipline, tools, worker); code == paused {
		return 0
	} else if code != 0 {
		return code
	}
	if err := gate.CheckLeave(cfg.Workspace, cfg.Product, st.Phase); err != nil {
		events.Log(cfg.Workspace, st.RunID, "gate_fail", st.Phase, err.Error())
		fmt.Fprintf(os.Stderr, "dudes: gate: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "phase_leave", st.Phase, "")
	return advanceLocked(cfg, st)
}

func runTestsGate(cfg config.Config, st *runstate.State) int {
	fmt.Println("gate: running framework tests after implementation…")
	cmd, slice := gateCommand(cfg, st)
	res, err := testrun.Run(cfg.Workspace, cfg.Launched(cmd))
	if err != nil && res == nil {
		fmt.Fprintf(os.Stderr, "dudes: test gate: %v\n", err)
		return 1
	}
	fmt.Printf("test: %v (exit %d, %s)\n", res.Args, res.ExitCode, res.Duration)
	warnUnrecorded(err)
	if expectRed(cfg, st) {
		switch {
		case res.OK && st != nil && st.Redo[st.TaskID] != "":
			// A red test sent back after its implementation ran may pass: the review decides.
			fmt.Println("tests: PASS — a red test redone after its implementation; the review decides")
			return 0
		case res.OK:
			fmt.Fprintln(os.Stderr, "dudes: test gate FAILED — the tests pass before the implementation exists, so they do not test the criterion yet (expect: fail)")
			return 1
		case !redTests(cfg, res.Output):
			fmt.Print(res.Output)
			fmt.Fprintln(os.Stderr, "dudes: test gate FAILED — red must be a failing test, not a build error (expect: fail)")
			return 1
		}
		fmt.Println("tests: FAIL, as expected — the criterion's test is red; the implementation task turns it green")
		return 0
	}
	if res.OK {
		if skipped := skippedAcceptance(cfg, cmd); len(skipped) > 0 {
			fmt.Fprintf(os.Stderr, "dudes: test gate FAILED — acceptance test(s) skipped: %s. %s\n", strings.Join(skipped, ", "), skipRule)
			return 1
		}
		fmt.Println("tests: PASS")
		if slice {
			fmt.Println("task slice green — the item's acceptance runs with the full suite (the phase's last task)")
			return 0
		}
		return acceptanceGate(cfg, st)
	}
	fmt.Print(res.Output)
	fmt.Fprintln(os.Stderr, "dudes: test gate FAILED — fix then re-run step (or dudes test -fix). Phase not advanced.")
	return 1
}

// gateCommand is the current task's own test command when it has one — a slice of the item
// (F65) — else the engine's test_command, the full suite.
func gateCommand(cfg config.Config, st *runstate.State) (cmd []string, slice bool) {
	if con, ok, _ := contract.Load(cfg.Workspace); ok && st != nil {
		if tk, ok := con.TaskByID(st.TaskID); ok && len(tk.Test) > 0 {
			return tk.Test, true
		}
	}
	return cfg.TestCommand, false
}

// skippedAcceptance reruns a green `go test` gate verbosely over the bound item's acceptance tests
// and returns those that reported SKIP — a t.Skip hides the gap the test exists to show (an
// instance trial: negative controls that skipped on a wrong fixture path). Tests a slice filters
// out with -skip do not run, so they do not count. The rerun is recorded under
// testrun.AcceptanceNotesRef: it names each passing test, so item status and evidence read it.
// Other toolchains: nothing to check yet.
func skippedAcceptance(cfg config.Config, cmd []string) []string {
	if len(cmd) < 2 || cmd[0] != "go" || cmd[1] != "test" {
		return nil
	}
	_, acs, err := itemAcceptance(cfg)
	if err != nil || len(acs) == 0 {
		return nil
	}
	var names []string
	for _, ac := range acs {
		names = append(names, spec.AcceptanceTestName(ac.ID))
	}
	verbose := append([]string{"go", "test", "-v", "-run", "^(" + strings.Join(names, "|") + ")$"}, cmd[2:]...)
	res, err := testrun.RunRecorded(cfg.Workspace, cfg.Launched(verbose), testrun.AcceptanceNotesRef)
	if res == nil {
		return nil
	}
	warnUnrecorded(err)
	results := testrun.Results(res.Output)
	var skipped []string
	for _, n := range names {
		if results[n] == testrun.Skip {
			skipped = append(skipped, n)
		}
	}
	return skipped
}

// expectRed reports whether the current task is red first (expect: fail): its gate must fail.
func expectRed(cfg config.Config, st *runstate.State) bool {
	if con, ok, _ := contract.Load(cfg.Workspace); ok && st != nil {
		if tk, ok := con.TaskByID(st.TaskID); ok {
			return tk.Expect == "fail"
		}
	}
	return false
}

// redTests tells failing tests from tests that do not build. With a bound item, red means one of its
// acceptance tests ran and reported `--- FAIL:` — a package of the code that does not exist yet may
// still fail to set up (a gate that names `./app/...` before the implementation). Without an item:
// no package marked `[build failed]` / `[setup failed]` (Go; other toolchains count as failing).
func redTests(cfg config.Config, output string) bool {
	if _, acs, err := itemAcceptance(cfg); err == nil && len(acs) > 0 {
		for _, ac := range acs {
			if strings.Contains(output, "--- FAIL: "+spec.AcceptanceTestName(ac.ID)+" ") {
				return true
			}
		}
		return false
	}
	return !strings.Contains(output, "[build failed]") && !strings.Contains(output, "[setup failed]")
}

// warnUnrecorded reports a test run that ran but could not be recorded as a git note — evidence
// then does not see it.
func warnUnrecorded(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: test run not recorded: %v\n", err)
	}
}

func cmdCheck(configPath string, asJSON bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	rep := gate.Assess(cfg.Workspace, cfg.Product)
	payload := map[string]any{
		"ok":          rep.OK(),
		"spec":        rep.Spec,
		"standards":   rep.Standards,
		"tasks":       rep.Tasks,
		"spec_n":      len(rep.Spec),
		"standards_n": len(rep.Standards),
	}
	if asJSON {
		code := emitJSON(payload)
		if !rep.OK() && code == 0 {
			return 1
		}
		return code
	}
	if rep.OK() {
		fmt.Println("check: PASS — spec and standards axes clean")
		fmt.Printf("tasks: %s\n", rep.Tasks)
		return 0
	}
	fmt.Printf("check: FAIL (spec %d, standards %d)\n", len(rep.Spec), len(rep.Standards))
	if len(rep.Spec) > 0 {
		fmt.Println("spec:")
		for _, f := range rep.Spec {
			fmt.Printf("  - %s\n", f.Error())
		}
	}
	if len(rep.Standards) > 0 {
		fmt.Println("standards:")
		for _, f := range rep.Standards {
			fmt.Printf("  - %s\n", f.Error())
		}
	}
	fmt.Printf("tasks: %s (converge-to-tasks)\n", rep.Tasks)
	return 1
}

func cmdDoctor(configPath string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	checks := doctor.Run(cfg)
	fail := false
	warn := false
	for _, c := range checks {
		fmt.Printf("%-12s %-4s %s\n", c.Name, c.Level, c.Message)
		switch c.Level {
		case doctor.Fail:
			fail = true
		case doctor.Warn:
			warn = true
		}
	}
	if fail {
		fmt.Println("doctor: FAIL")
		return 1
	}
	if warn {
		fmt.Println("doctor: OK (with warnings)")
		return 0
	}
	fmt.Println("doctor: OK")
	return 0
}

func cmdFreeze(configPath string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	runID := ""
	if st, ok, _ := runstate.Load(cfg.Workspace); ok && st != nil {
		runID = st.RunID
	}
	rec, err := freeze.WriteAll(cfg.Workspace, cfg.Product, runID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, runID, "spec_freeze", phase.Spec, rec.SHA256)
	fmt.Printf("froze %s sha256=%s at %s (%d inputs)\n", rec.PRD, rec.SHA256, rec.FrozenAt.Format(time.RFC3339), len(rec.Inputs))
	return 0
}

func trimSHA(s string) string {
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

func cmdPack(configPath string, stub bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	id := phase.Idea
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if ok {
		id = st.Phase
	}
	return runPack(cfg, id, stub)
}

func runPack(cfg config.Config, id phase.ID, stub bool) int {
	pcfg := cfg.PackerRunConfig()
	includes := packer.IncludesFor(id, cfg.Product)
	fmt.Printf("pack: phase %s includes %v\n", id, includes)

	var rel string
	var err error
	if stub {
		rel, err = packer.WriteStubPack(cfg.Workspace, includes, pcfg.Output)
	} else {
		if err := packer.LookPath(pcfg.Bin); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		rel, err = packer.Run(context.Background(), cfg.Workspace, includes, pcfg)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	src := "repomix"
	if stub {
		src = "stub"
	}
	fmt.Printf("wrote %s via %s\n", rel, src)
	return 0
}

func loadPackContext(cfg config.Config) string {
	rel := cfg.PackerRunConfig().Resolve().Output
	body, ok, err := packer.Read(cfg.Workspace, rel)
	if err != nil || !ok {
		return ""
	}
	return packer.TruncateForPrompt(body, 24000)
}

// runImplementation invokes the code worker (Aider default, optional Pi) or writes a stub entrypoint.
// With convergence on (F72) the worker is re-invoked with a fresh start until a pass changes nothing.
func runImplementation(cfg config.Config, st *runstate.State, stub bool, packCtx, taskOverride, discipline string, tools []string, worker string) int {
	prdRel, _ := gate.SpecPath(cfg.Product)
	prd := ""
	trailers := ""
	if item, err := itemSpec(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	} else if item != nil {
		prdRel, prd, trailers = item.rel, item.body, item.trailers
		fmt.Printf("spec input: item pack %s (F71)\n", item.rel)
		if con, ok, _ := contract.Load(cfg.Workspace); ok {
			if tk, ok := con.TaskByID(st.TaskID); ok && len(tk.Refs) > 0 {
				prd = itempack.Narrow(prd, tk.Refs)
				fmt.Printf("spec input narrowed to the task's %d reference(s): %d of %d chars\n", len(tk.Refs), len(prd), len(item.body))
			}
		}
	} else if prdRel != "" && artifact.Exists(cfg.Workspace, prdRel) {
		data, err := artifact.Read(cfg.Workspace, prdRel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		prd = string(data)
	}
	msg := aider.Message{Product: cfg.Product, Intent: st.Prompt, Task: taskOverride, Feedback: st.TaskFeedback(),
		SpecRel: prdRel, Spec: prd, Pack: packCtx, Discipline: discipline}
	if cmd, slice := gateCommand(cfg, st); slice {
		msg.DoneCmd = cmd
	}
	msg.Red = expectRed(cfg, st)
	con, _, _ := contract.Load(cfg.Workspace)
	scopePats := writable(con, st)
	msg.Writable = scopePats
	if con != nil {
		msg.Constraints = con.Constraints
	}
	gateCmd, _ := gateCommand(cfg, st)
	msg.Tests, msg.ReadOnlyTests = taskTests(cfg, scopePats, gateCmd)
	gateLog, gateFiles, gateRefs := gateFailure(cfg.Workspace, cfg.Launched(gateCmd), scopePats)
	msg.GateLog = gateLog
	if gateLog != "" {
		fmt.Printf("gate: last run on this commit failed — its output, %d named file(s) to edit and %d to read go to the worker\n", len(gateFiles), len(gateRefs))
	}
	message := msg.String() + escalationAnswer(st)
	worker = strings.ToLower(strings.TrimSpace(worker))
	ws := cfg.Workspace
	if worktree.Want(cfg.Worktree, tools) {
		abs, branch, err := worktree.Add(cfg.Workspace, st.RunID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: worktree warning: %v — using workspace\n", err)
		} else {
			ws = abs
			fmt.Printf("worktree: %s (%s)\n", abs, branch)
		}
	}
	subject, trailers := passSubjects(cfg, st, ws, trailers)
	clearScope, err := bindScope(cfg, ws, st)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	defer clearScope()
	logPromptSize(cfg, st, phase.Implementation, message)

	if stub || worker == "stub" {
		// the stub's file is no worker's output a task asked for: outside the task's scope it is not
		// written, so it never lands on an item branch (F88)
		if rel := aider.StubImplRel(cfg.Product); len(scopePats) > 0 && !scope.Inside(rel, scopePats) {
			fmt.Printf("stub: %s is outside the task's writable paths — nothing written\n", rel)
			return 0
		}
		rel, err := aider.WriteStubImpl(ws, cfg.Product)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: stub impl: %v\n", err)
			return 1
		}
		if err := gitx.AddCommit(ws, []string{rel}, withTrailers(subject(1, 1), trailers)); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: git: %v\n", err)
			return 1
		}
		fmt.Printf("wrote %s via stub and committed\n", rel)
		return 0
	}

	var invoke func(msg string) (string, error)
	if worker == "pi" {
		pcfg := cfg.PiRunConfig()
		if err := pi.LookPath(pcfg.Bin); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		invoke = func(msg string) (string, error) {
			fmt.Println("implementation: invoking pi (fresh process)…")
			return pi.Run(context.Background(), ws, msg, tools, pcfg)
		}
	} else {
		acfg := cfg.AiderRunConfig()
		acfg.Extra = append(acfg.Extra, aider.ReadFlags(ws, tools)...)
		acfg.Files = append([]string{}, msg.Tests...)
		for _, f := range msg.ReadOnlyTests { // written by the red-first task: read, never change
			acfg.Extra = append(acfg.Extra, "--read", f)
		}
		for _, f := range gateFiles {
			if !slices.Contains(acfg.Files, f) {
				acfg.Files = append(acfg.Files, f)
			}
		}
		for _, f := range gateRefs { // another task's file the gate names: read-only
			acfg.Extra = append(acfg.Extra, "--read", f)
		}
		if code, skipped := scopeCode(ws, scopePats, acfg.Files); len(code)+skipped > 0 {
			acfg.Files = append(acfg.Files, code...)
			fmt.Printf("scope: %d existing file(s) go to the worker", len(code))
			if skipped > 0 {
				fmt.Printf(", %d over the %d KB budget left out", skipped, scopeCodeBudget>>10)
			}
			fmt.Println()
		}
		flags, err := aider.IgnoreFlags(ws, scopePats)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		acfg.Extra = append(acfg.Extra, flags...)
		if err := aider.LookPath(acfg.Exe()); err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		invoke = func(msg string) (string, error) {
			fmt.Println("implementation: invoking aider (fresh process, --no-auto-commits)…")
			return aider.Run(context.Background(), ws, msg, acfg)
		}
	}
	if worker == "pi" { // each call's tokens and time in the run's events, its reply kept (F88)
		invoke = metered(cfg, st, "pi", "", invoke)
	} else {
		invoke = metered(cfg, st, "aider", cfg.Aider.Model, invoke)
	}
	return implementConverged(cfg, st, ws, scopePats, invoke, message, subject, trailers)
}

func phaseBody(cfg config.Config, st *runstate.State, cur phase.ID, stub bool, packCtx, taskOverride, discipline string) (body, source string, err error) {
	if cur == phase.TestAuthoring {
		return authoredTests(cfg, st)
	}
	if stub {
		return artifact.StubContent(cur, cfg.Product, st.Prompt), "stub", nil
	}
	switch cur {
	case phase.Idea, phase.Strategy, phase.Spec, phase.QA, phase.Documentation:
		prior := priorMarkdown(cfg, cur)
		specRel, _ := gate.SpecPath(cfg.Product)
		if cur == phase.QA || cur == phase.Documentation {
			item, err := itemSpec(cfg)
			if err != nil {
				return "", "", err
			}
			if item != nil {
				specRel, prior = item.rel, item.body
			}
		}
		prompt := ollama.PhasePrompt(cur, cfg.Product, st.Prompt, prior, packCtx, specRel, taskOverride, st.TaskFeedback(), discipline) + escalationAnswer(st)
		logPromptSize(cfg, st, cur, prompt)
		client := cfg.OllamaClient()
		ctx := context.Background()
		start := time.Now()
		out, usage, err := client.GenerateUsage(ctx, prompt)
		recordCall(cfg, st, ollamaUsage(client.Model(), usage, err), time.Since(start), out, err)
		if err != nil {
			return "", "", fmt.Errorf("ollama worker (use -stub offline): %w", err)
		}
		if e, ok := extract.FindEscalation(out); ok {
			return "", "", &escalationError{e}
		}
		return extract.Artifact(out), "ollama", nil
	default:
		return artifact.StubContent(cur, cfg.Product, st.Prompt), "stub", nil
	}
}

// itemAcceptance is the bound work item and its behavior criteria — each gets its own failing
// test file at test_authoring (spec.ItemTestFile); empty when no item is bound.
func itemAcceptance(cfg config.Config) (item string, acs []spec.Criterion, err error) {
	con, ok, err := contract.Load(cfg.Workspace)
	if err != nil || !ok || con.Item == "" {
		return "", nil, err
	}
	for _, a := range con.Acceptance {
		if strings.EqualFold(a.Kind, "behavior") {
			acs = append(acs, spec.Criterion{ID: a.ID, Text: a.Text})
		}
	}
	return con.Item, acs, nil
}

func authoredTests(cfg config.Config, st *runstate.State) (body, source string, err error) {
	// Item-bound: one failing test per acceptance criterion of the item (tests first); the
	// primary artifact lists their files, task.go writes them.
	if item, acs, err := itemAcceptance(cfg); err != nil {
		return "", "", err
	} else if len(acs) > 0 {
		return spec.ItemTestIndex(item, acs), "item " + item, nil
	}
	prdRel, ok := gate.SpecPath(cfg.Product)
	if !ok || !artifact.Exists(cfg.Workspace, prdRel) {
		return artifact.StubContent(phase.TestAuthoring, cfg.Product, st.Prompt), "stub", nil
	}
	data, err := artifact.Read(cfg.Workspace, prdRel)
	if err != nil {
		return "", "", err
	}
	// Evidence rows follow from the PRD and these tests (evidence.Derive) — nothing to store here.
	return spec.GoTests(cfg.Product, spec.Parse(string(data))), "spec", nil
}

func priorMarkdown(cfg config.Config, cur phase.ID) string {
	// Prefer previous lifecycle artifact when present (durable handoff).
	order := []phase.ID{phase.Idea, phase.Strategy, phase.Spec}
	var last string
	for _, id := range order {
		if id == cur {
			break
		}
		rel, ok := artifact.Path(cfg.Product, id)
		if !ok || !artifact.Exists(cfg.Workspace, rel) {
			continue
		}
		data, err := artifact.Read(cfg.Workspace, rel)
		if err == nil && len(data) > 0 {
			last = string(data)
		}
	}
	return compact.Text(last, compact.DefaultMax)
}

func cmdEvents(configPath string, n int, asJSON bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	evs, err := events.Tail(cfg.Workspace, n)
	gone := err == nil && !events.Logged(cfg.Workspace)
	if gone { // the log went with its workspace (a slice's worktree): the ledger is in git (F88)
		evs, err = events.Ledger(cfg.Workspace)
		evs = events.Last(evs, n)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if asJSON {
		if evs == nil {
			evs = []events.Event{}
		}
		return emitJSON(evs)
	}
	if gone {
		fmt.Printf("events: %s is not here — the ledger in refs/notes/%s: worker calls, decide, redo\n", events.Path(), events.LedgerRef)
	}
	if len(evs) == 0 {
		fmt.Println("events: (none)")
		return 0
	}
	for _, ev := range evs {
		detail := ev.Detail
		if len(detail) > 80 {
			detail = detail[:77] + "…"
		}
		if ev.By != "" { // who answered a pause (F88)
			detail += " — by " + runstate.Decider{Identity: ev.By, Role: ev.Role}.String()
		}
		fmt.Printf("%s  %-12s %-16s %s\n",
			ev.Time.Format(time.RFC3339),
			ev.Type,
			ev.Phase,
			detail,
		)
	}
	return 0
}

func cmdTest(configPath string, fix bool) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	res, err := testrun.Run(cfg.Workspace, cfg.Launched(cfg.TestCommand))
	if err != nil && res == nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	fmt.Printf("test: %v (exit %d, %s)\n", res.Args, res.ExitCode, res.Duration)
	if err == nil {
		fmt.Printf("log:  git notes --ref %s show HEAD\n", testrun.NotesRef)
	}
	if res.Output != "" {
		fmt.Print(res.Output)
		if !strings.HasSuffix(res.Output, "\n") {
			fmt.Println()
		}
	}
	warnUnrecorded(err)
	if res.OK {
		fmt.Println("tests: PASS")
		st, _, _ := runstate.Load(cfg.Workspace) // nil without a run: the check still reports
		return acceptanceGate(cfg, st)
	}
	fmt.Println("tests: FAIL")
	if !fix {
		return 1
	}

	fmt.Println("fix: invoking ollama with test log only (no chat history)…")
	log, ok, err := testrun.ReadLastLog(cfg.Workspace)
	if err != nil || !ok {
		fmt.Fprintf(os.Stderr, "dudes: no test log for fix: %v\n", err)
		return 1
	}
	client := cfg.OllamaClient()
	out, err := client.Generate(context.Background(), ollama.FixFromTestPrompt(log))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: fix worker: %v\n", err)
		return 1
	}
	rel := cfg.Product + "_FIX.md"
	if err := artifact.Write(cfg.Workspace, rel, []byte(out+"\n")); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	msg := "chore(phase): fix-from-test [🔹engine]"
	if err := gitx.AddCommit(cfg.Workspace, []string{rel}, msg); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: git: %v\n", err)
		return 1
	}
	fmt.Printf("wrote %s (suggestion only — apply via code worker / step)\n", rel)
	return 1
}

// cmdSeal records the human seal. With a release version the tag is the git contract's
// <product>-vX.Y.Z (CORE §10) — refused before sealing if invalid or taken; without one the run
// gets dudes/seal/<product>/<run>. The seal commit names the sealer (Sealed-by) and the tag is
// signed when the sealer's git can sign (user.signingkey); with seal.require_signature an unsigned
// seal is refused before any file, commit or tag is made (F89).
func cmdSeal(configPath, release string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	tag := ""
	if release != "" {
		t, err := gitx.ReleaseTagName(cfg.Product, release)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
			return 1
		}
		if gitx.TagExists(cfg.Workspace, t) {
			fmt.Fprintf(os.Stderr, "dudes: %s exists — a release is sealed once; the next one is a new X.Y.Z\n", t)
			return 1
		}
		tag = t
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes init")
		return 1
	}
	if findings := gate.Converge(cfg.Workspace, cfg.Product); len(findings) > 0 {
		fmt.Fprintln(os.Stderr, "dudes: cannot seal — converge FAIL:")
		for _, f := range findings {
			fmt.Fprintf(os.Stderr, "  - %s\n", f.Error())
		}
		return 1
	}
	if err := humangate.ApplySeal(st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if tag == "" {
		tag = gitx.SealTagName(cfg.Product, st.RunID)
		if gitx.TagExists(cfg.Workspace, tag) {
			fmt.Fprintf(os.Stderr, "dudes: %s exists — this run is sealed already, or the tag was made by hand; nothing was changed\n", tag)
			return 1
		}
	}
	// Who seals, and whether the tag will be signed: settled before the first file, commit or tag.
	sealer, err := gitx.Identity(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: cannot seal — %v\n", err)
		return 1
	}
	signing := gitx.CanSign(cfg.Workspace)
	if !signing.OK && cfg.Seal.RequireSignature {
		fmt.Fprintf(os.Stderr, "dudes: cannot seal — seal.require_signature is on, and the seal would be unsigned: %s\n", signing.Why)
		fmt.Fprintln(os.Stderr, "  how: git config user.signingkey <key> (gpg.format openpgp, ssh or x509) — or set \"require_signature\" to false under \"seal\" in dudes.json")
		return 1
	}
	rel := humangate.SealArtifactRel(cfg.Product)
	body := humangate.SealMarkdown(cfg.Product, st)
	if err := artifact.Write(cfg.Workspace, rel, []byte(body)); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	before, err := commitSeal(cfg.Workspace, rel, humangate.SealMessage(sealer))
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: git: %v\n", err)
		return 1
	}
	// From here to the tag a failure takes the seal commit back: a seal is all there is, or nothing.
	abort := func(why, fix string) int {
		undoSealCommit(cfg.Workspace, before, rel)
		fmt.Fprintf(os.Stderr, "dudes: cannot seal — %s\n  nothing is sealed (the seal commit is taken back); %s, then run dudes seal again\n", why, fix)
		return 1
	}
	// The run's state goes on the seal commit itself, before a worktree merge moves HEAD: dudes verify
	// reads it there.
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		return abort(fmt.Sprintf("the run state could not be saved: %v", err), "fix that")
	}
	msg := fmt.Sprintf("DUDES human seal for %s (%s)", cfg.Product, st.RunID)
	if signing.OK {
		// A key that is set and cannot sign is no reason to seal unsigned: the sealer meant a signed seal.
		if err := gitx.SignedTag(cfg.Workspace, tag, msg); err != nil {
			return abort(fmt.Sprintf("git could not sign the seal tag: %v", err), "fix the signing setup")
		}
		fmt.Printf("tag: %s (signed, %s)\n", tag, signing.Format)
	} else {
		if err := gitx.AnnotatedTag(cfg.Workspace, tag, msg); err != nil {
			return abort(fmt.Sprintf("git could not make the seal tag: %v", err), "fix that")
		}
		fmt.Printf("tag: %s\n", tag)
		fmt.Printf("unsigned: %s\n", signing.Why)
	}
	events.Log(cfg.Workspace, st.RunID, "seal_tag", phase.Sealed, tag)
	events.Log(cfg.Workspace, st.RunID, "seal", phase.Sealed, rel)
	if abs, err := worktree.Abs(cfg.Workspace, st.RunID); err == nil {
		if _, err := os.Stat(abs); err == nil {
			if err := worktree.MergeIntoHEAD(cfg.Workspace, st.RunID); err != nil {
				fmt.Fprintf(os.Stderr, "dudes: worktree merge warning: %v\n", err)
			} else if err := worktree.Remove(cfg.Workspace, st.RunID); err != nil {
				fmt.Fprintf(os.Stderr, "dudes: worktree remove warning: %v\n", err)
			}
		}
	}
	fmt.Printf("sealed %s — wrote %s\n", st.RunID, rel)
	return 0
}

// commitSeal makes the seal commit: the seal file with the message. When the file is already as
// committed there is nothing to stage, and the commit is empty — the seal is a commit of its own either
// way, the one the tag goes on and the trailer is in. before is where HEAD was.
func commitSeal(ws, rel, message string) (before string, err error) {
	before, _ = gitx.HeadFull(ws)
	if err := gitx.AddCommit(ws, []string{rel}, message); err != nil {
		return before, err
	}
	if head, _ := gitx.HeadFull(ws); head == before {
		return before, gitx.EmptyCommit(ws, message)
	}
	return before, nil
}

// undoSealCommit takes back the seal commit made a moment ago — HEAD goes back to before, and the
// seal file to what it was — so a seal that could not be signed leaves no trace.
func undoSealCommit(ws, before, rel string) {
	if head, _ := gitx.HeadFull(ws); before == "" || head == before {
		return // no commit was made
	}
	if err := gitx.ResetSoft(ws, before); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: %v — the seal commit is still in the history\n", err)
		return
	}
	if err := gitx.Discard(ws, []string{rel}); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: warning: %v — %s is still the seal's\n", err, rel)
	}
}

// cmdVerify checks a seal from git alone (F89): the tag's signature by git's own verifier, the seal
// commit with its Sealed-by trailer, the run's state note. The tag is the newest seal tag reachable
// from HEAD unless one is named. Exit 0 only when every check holds.
func cmdVerify(configPath string, args []string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "usage: dudes verify [tag]")
		return 1
	}
	var tag string
	if len(args) == 1 {
		tag = args[0]
	} else {
		var err error
		if tag, err = humangate.LatestSealTag(cfg.Workspace); err != nil {
			fmt.Printf("NOT verified: %v\n", err)
			return 1
		}
	}
	rep := humangate.Verify(cfg.Workspace, tag)
	fmt.Print(rep)
	if !rep.OK() {
		return 1
	}
	return 0
}

func cmdReject(configPath, feedback string) int {
	cfg, code := loadCfg(configPath)
	if code != 0 {
		return code
	}
	st, ok, err := runstate.Load(cfg.Workspace)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "dudes: no run state — run: dudes init")
		return 1
	}
	feedback = strings.TrimSpace(feedback)
	if strings.HasPrefix(feedback, "-reason ") || strings.HasPrefix(feedback, "--reason ") {
		parts := strings.SplitN(feedback, " ", 2)
		if len(parts) == 2 {
			feedback = strings.TrimSpace(parts[1])
		}
	}
	if err := humangate.ApplyReject(st, feedback); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	rel := humangate.FeedbackArtifactRel(cfg.Product)
	body := humangate.FeedbackMarkdown(cfg.Product, st)
	if err := artifact.Write(cfg.Workspace, rel, []byte(body)); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	if err := gitx.AddCommit(cfg.Workspace, []string{rel}, "chore(phase): reject [🔹human]"); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: git: %v\n", err)
		return 1
	}
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "reject", st.Phase, st.Feedback)
	fmt.Printf("rejected — cycle %d, resumed at %s\n", st.Cycle, st.Phase)
	fmt.Printf("wrote %s\n", rel)
	return 0
}

func advanceLocked(cfg config.Config, st *runstate.State) int {
	// a slice never leaves implementation: its QR is the spec repository's, and a later phase would
	// commit the engine's own QR and docs into the branch the role merges (F88)
	if con, ok, _ := contract.Load(cfg.Workspace); ok && con.NamesTests() && st.Phase == phase.Implementation {
		fmt.Fprintf(os.Stderr, "dudes: slice %s ends with its tasks — it does not leave %s\n", con.ID, st.Phase)
		return 1
	}
	reg := phase.Default()
	next, ok := reg.Next(st.Phase)
	if !ok {
		fmt.Fprintf(os.Stderr, "dudes: already at terminal phase %s\n", st.Phase)
		return 1
	}
	if next == phase.HumanDecision {
		if findings := gate.Converge(cfg.Workspace, cfg.Product); len(findings) > 0 {
			fmt.Fprintln(os.Stderr, "dudes: converge blocked advance to human_decision:")
			for _, f := range findings {
				fmt.Fprintf(os.Stderr, "  - %s\n", f.Error())
			}
			return 1
		}
	}
	prev := st.Phase
	st.Phase = next
	switch next {
	case phase.HumanDecision:
		st.Status = runstate.StatusWaiting
	case phase.Sealed:
		st.Status = runstate.StatusSealed
	default:
		st.Status = runstate.StatusRunning
	}
	if err := runstate.Save(cfg.Workspace, st); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: %v\n", err)
		return 1
	}
	events.Log(cfg.Workspace, st.RunID, "advance", st.Phase, string(prev)+"→"+string(st.Phase))
	fmt.Printf("advanced %s → %s\n", prev, st.Phase)
	return 0
}

func cmdPhases() int {
	for _, d := range phase.Default().Order() {
		fmt.Printf("%s\t%s\n", d.ID, d.Capability)
	}
	return 0
}

func loadDiscipline(cfg config.Config, rec *recipe.Recipe) (string, error) {
	if rec == nil {
		return "", nil
	}
	var parts []string
	if len(rec.Skills) > 0 {
		s, err := skill.LoadMany(cfg.Workspace, rec.Skills)
		if err != nil {
			return "", err
		}
		if s != "" {
			parts = append(parts, s)
		}
	}
	if len(rec.Rules) > 0 {
		s, err := skill.LoadRules(cfg.Workspace, rec.Rules)
		if err != nil {
			return "", err
		}
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}

func cmdBench(asJSON bool) int {
	dir, err := os.MkdirTemp("", "dudes-bench-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: bench temp: %v\n", err)
		return 1
	}
	product := "hello"
	if err := bench.Prepare(dir, product); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: bench prepare: %v\n", err)
		return 1
	}
	cfgPath := filepath.Join(dir, "dudes.json")
	started := time.Now()
	code := cmdRun(cfgPath, "hello CLI that greets", true, false, true, "")
	rep, err := bench.Summarize(dir, product, started, true, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dudes: bench summarize: %v\n", err)
		return 1
	}
	if code != 0 {
		rep.OK = false
	}
	if err := bench.WriteLast(dir, rep); err != nil {
		fmt.Fprintf(os.Stderr, "dudes: bench write: %v\n", err)
	}
	if cwd, err := os.Getwd(); err == nil {
		_ = bench.WriteLast(cwd, rep)
	}
	if asJSON {
		if err := emitJSON(rep); err != 0 {
			return err
		}
		if !rep.OK {
			return 1
		}
		return 0
	}
	fmt.Printf("bench: workspace %s\n", dir)
	fmt.Printf("bench: duration %dms phases_left %d gate_fails %d final %s check=%v ok=%v\n",
		rep.DurationMs, rep.PhaseLeaves, rep.GateFails, rep.FinalPhase, rep.CheckOK, rep.OK)
	fmt.Printf("bench: report .dudes/bench/last.json\n")
	if !rep.OK {
		return 1
	}
	return 0
}

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, `dudes — DUDES process engine (Go)

Usage:
  dudes status [-json]                 show run state. A task that ran and waits for approval (approval:
                                       after) adds its review packet (F88): the task's diff of its writable
                                       paths since it first ran, the gate run on HEAD with the tests that
                                       passed and failed, the worker's last reply — each cut to a bound,
                                       with the git command or file that holds the whole. A worker's
                                       escalation shows its question and the answer command
  dudes doctor                         probe git/ollama/aider/packer/config
  dudes events [-n N] [-json]          show last N run events (.dudes/events.jsonl). Each worker call of a step is a
                                       worker event: its local tokens (Ollama's counts, or Aider's own
                                       report) and its duration; a decide or redo names who answered.
                                       Those three are the ledger, kept in git too (refs/notes/dudes/ledger):
                                       without the file (a slice's worktree removed) events reads them there
  dudes init [-prompt text] [text]     start run at first phase
  dudes do [-pick] <prompt>            intake: a named work item binds and starts at test_authoring,
                                       a make goal runs; anything else prints the menu (exit 2);
                                       -pick: the text model suggests one entry, you run it
  dudes do <ITEM> -slice <file.json> [-base <ref>] [-run]
                                       item handoff (F88), from the dev repo's root: prepares one slice
                                       of the item — worktree .worktrees/<id> on feature/<id> from -base
                                       (default: the current branch's upstream, else HEAD; refused when
                                       either exists), each spec root pinned as a detached worktree
                                       <root repo>/.worktrees/spec-<id>, the run's .dudes/, a red-test
                                       task and a code task, init, freeze — and prints the next command;
                                       -run then runs the slice to its first pause.
                                       Reads the repo's tracked dudes/: profile.json; specset.json
                                       (roots {"<name>": {"repo": "../spec", "pin": "release/x-v1.0"}},
                                       entries as in .dudes/specset.json); config.json (a dudes.json
                                       without workspace and machine values). A repo of several products
                                       keeps them per product in dudes/<product>/, its config.json over
                                       the shared dudes/config.json; the item's ID finds its product
                                       (its profile's ID family). The model host comes from
                                       DUDES_OLLAMA_URL, else your dudes.json, else OLLAMA_API_BASE,
                                       never from the repo. The slicing: {"id": "bk-001a", "item",
                                       "acceptance": ["AC-2"], "scope": ["pkg/*"], "tests": ["pkg/x_test.go"],
                                       "hooks": [..], "test_command": [..], "gate_command": [..],
                                       "tools": [..], "refs": ["TA:§3"], "constraints": [..], "later"}
  dudes reset -yes                     clear run state (keep events/freeze/packs)
  dudes run [-prompt text] [-stub] [-pack] [-skip-test] [-until phase]
                                       init if needed; step until human_decision (or -until)
  dudes step [-stub] [-pack] [-skip-test]
                                       one phase: worker + commit + advance
  dudes check [-json]                  two-axis converge (spec vs standards) + TASKS.md
  dudes bench [-json]                  hermetic stub run; write .dudes/bench/last.json
  dudes freeze                         freeze PRD + all spec-set/contract inputs (repo/rev)
  dudes specset [init]                 show or write .dudes/specset.json (original names)
  dudes contract                       show/validate .dudes/contract.json
  dudes evidence                       show requirement→test→run rows
  dudes decide <task-id> [-by <role>] [note]
                                       approve a waiting in-phase task (F66). On a slice's branch (do -slice)
                                       a finished task lands as one commit, its passes kept under
                                       refs/dudes/passes/<slice>/<task>/<n>; once every task is decided the
                                       run ends (status ended) and the branch merges as it is (F88)
  dudes redo <task-id> [-by <role>] <feedback>
                                       send a task back: its next run gets the feedback (a finished one
                                       reopens, an ended slice runs again). Both record who answered, in the
                                       event and the run state: your git identity, and the role -by names (F88)
  dudes land <slice> -by <role>        from the dev repo's root, once the slice's run ended: git merge --no-ff
                                       of feature/<slice> into the branch checked out there, the role's tag
                                       and the item on the merge commit (hooks run; a stopped merge is taken
                                       back), then its spec pins, worktree and branch go (F88); refused
                                       first, naming the files, when the worktree or a pin has a change no
                                       commit holds
  dudes item check [-json]             spec-set consistency C1–C8 (needs .dudes/profile.json)
  dudes item list | board              work items + status from QR and acceptance runs (F70)
  dudes item pack <ID>… [-all] [-budget N]
                                       work-item context pack → .dudes/items/<ID>/ (F68)
  dudes item contract <ID>             project .dudes/contract.json from the item (F71)
  dudes pack [-stub]                   phase-scoped Repomix context pack → .dudes/
  dudes advance                        advance phase only (blocked at human_decision / decide)
  dudes test [-fix]                    framework-owned tests; -fix re-invokes Ollama with log only
  dudes seal [X.Y.Z]                   human seal (only at human_decision); annotated tag <product>-vX.Y.Z
                                       with a release version, else dudes/seal/<product>/<run>. The seal
                                       commit names you (Sealed-by); the tag is signed (git tag -s) when
                                       your git has user.signingkey (gpg.format openpgp, ssh or x509),
                                       else unsigned and it says so; seal.require_signature (dudes.json)
                                       refuses an unsigned seal before anything is made
  dudes verify [tag]                   check a seal, from git alone: git tag -v accepts the tag's signature,
                                       it points at the seal commit with its Sealed-by trailer, the run's
                                       state note says sealed. The tag defaults to the newest one reachable
                                       from HEAD that points at a seal commit. Exit 0 only when every
                                       check holds
  dudes reject <text>                  human reject + resume at implementation
  dudes phases                         list lifecycle registry
  dudes version                        print version

  Optional recipes/<phase>.json override worker instructions (F27).
  recipes may name skills/<name>/SKILL.md (F48) and rules files (F49).
  Spec phase writes a SHALL + GIVEN/WHEN/THEN PRD (F37) + TA and auto-freezes all inputs (F59).
  Events append to .dudes/events.jsonl. Impl requires matching freeze.
  Large prior/pack context is head/tail compacted (F28-lite).
  Example increment: examples/sample/ (a made-up product; it modifies no repo).

`)
	fs.PrintDefaults()
}

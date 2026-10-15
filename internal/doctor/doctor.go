// Package doctor probes local engine dependencies (config, git, ollama, aider, packer).
package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
	"github.com/Thomas-Pietsch/dudes/internal/config"
	"github.com/Thomas-Pietsch/dudes/internal/freeze"
	"github.com/Thomas-Pietsch/dudes/internal/packer"
	"github.com/Thomas-Pietsch/dudes/internal/pi"
	"github.com/Thomas-Pietsch/dudes/internal/product"
	"github.com/Thomas-Pietsch/dudes/internal/runstate"
)

// Level is check severity for reporting.
type Level string

const (
	OK   Level = "ok"
	Warn Level = "warn"
	Fail Level = "fail"
)

// Check is one probe result.
type Check struct {
	Name    string
	Level   Level
	Message string
}

// Run probes the environment for cfg. Does not mutate workspace beyond ensuring .dudes exists.
func Run(cfg config.Config) []Check {
	var out []Check
	out = append(out, checkProduct(cfg.Product))
	out = append(out, checkWorkspace(cfg.Workspace))
	out = append(out, checkGit(cfg.Workspace))
	out = append(out, checkOllama(cfg))
	out = append(out, checkAider(cfg))
	out = append(out, checkAiderModel(cfg))
	out = append(out, checkPi(cfg))
	out = append(out, checkPacker(cfg))
	out = append(out, checkTestCmd(cfg))
	out = append(out, checkState(cfg.Workspace))
	out = append(out, checkFreeze(cfg.Workspace, cfg.Product))
	return out
}

func checkProduct(slug string) Check {
	if err := product.ValidateSlug(slug); err != nil {
		return Check{"product", Fail, err.Error()}
	}
	return Check{"product", OK, slug}
}

func checkWorkspace(root string) Check {
	abs, err := filepath.Abs(root)
	if err != nil {
		return Check{"workspace", Fail, err.Error()}
	}
	fi, err := os.Stat(abs)
	if err != nil || !fi.IsDir() {
		return Check{"workspace", Fail, fmt.Sprintf("not a directory: %s", abs)}
	}
	probe := filepath.Join(abs, ".dudes")
	if err := os.MkdirAll(probe, 0o755); err != nil {
		return Check{"workspace", Fail, fmt.Sprintf("not writable: %v", err)}
	}
	return Check{"workspace", OK, abs}
}

func checkGit(root string) Check {
	cmd := exec.Command("git", "-C", root, "rev-parse", "--is-inside-work-tree")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Check{"git", Fail, fmt.Sprintf("not a git repo: %v (%s)", err, trim(string(out)))}
	}
	head, err := exec.Command("git", "-C", root, "rev-parse", "--short", "HEAD").Output()
	if err != nil {
		return Check{"git", Warn, "repo ok, but no HEAD yet"}
	}
	return Check{"git", OK, "HEAD " + trim(string(head))}
}

func checkOllama(cfg config.Config) Check {
	client := cfg.OllamaClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	models, err := client.ListModels(ctx)
	if err != nil {
		return Check{"ollama", Warn, fmt.Sprintf("unreachable (%s): %v — use -stub offline", client.BaseURL(), err)}
	}
	want := client.Model()
	found := false
	for _, m := range models {
		if m == want || strings.HasPrefix(m, want) {
			found = true
			break
		}
	}
	if !found {
		return Check{"ollama", Warn, fmt.Sprintf("up at %s but model %q not listed (%d models)", client.BaseURL(), want, len(models))}
	}
	return Check{"ollama", OK, fmt.Sprintf("%s / %s", client.BaseURL(), want)}
}

// checkAider starts the code worker the way the engine does — through aider.launch, so Docker,
// the worker image and Aider itself are probed in one go — and asks for its version.
func checkAider(cfg config.Config) Check {
	acfg := cfg.AiderRunConfig()
	if err := aider.LookPath(acfg.Exe()); err != nil {
		return Check{"aider", Warn, err.Error() + " — impl needs -stub or install aider"}
	}
	bin := nonzeroBin(acfg.Bin, "aider")
	cmdline := append(append([]string{}, acfg.Launch...), bin)
	args := append(append([]string{}, cmdline[1:]...), "--version")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, cmdline[0], args...)
	cmd.Dir = cfg.Workspace
	out, err := cmd.CombinedOutput()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	last := strings.TrimSpace(lines[len(lines)-1])
	if err != nil {
		return Check{"aider", Warn, fmt.Sprintf("%s does not start: %v %s", strings.Join(cmdline, " "), err, last)}
	}
	return Check{"aider", OK, fmt.Sprintf("%s (%s)", last, strings.Join(cmdline, " "))}
}

// checkAiderModel: the code worker's Ollama model is served by the model host (ollama.base_url).
func checkAiderModel(cfg config.Config) Check {
	model := cfg.AiderRunConfig().Model
	name, ok := strings.CutPrefix(model, "ollama_chat/")
	if !ok {
		if name, ok = strings.CutPrefix(model, "ollama/"); !ok {
			return Check{"aider model", OK, nonzeroBin(model, "(aider default)") + " — not an Ollama model, not probed"}
		}
	}
	client := cfg.OllamaClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	models, err := client.ListModels(ctx)
	if err != nil {
		return Check{"aider model", Warn, fmt.Sprintf("model host %s unreachable: %v", client.BaseURL(), err)}
	}
	for _, m := range models {
		if m == name || strings.HasPrefix(m, name+":") {
			return Check{"aider model", OK, fmt.Sprintf("%s on %s", name, client.BaseURL())}
		}
	}
	return Check{"aider model", Warn, fmt.Sprintf("%s not served by %s (%d models)", name, client.BaseURL(), len(models))}
}

func checkPi(cfg config.Config) Check {
	pcfg := cfg.PiRunConfig()
	if err := pi.LookPath(pcfg.Bin); err != nil {
		return Check{"pi", Warn, err.Error() + " — optional F55 worker"}
	}
	return Check{"pi", OK, nonzeroBin(pcfg.Bin, "pi")}
}

func nonzeroBin(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func checkPacker(cfg config.Config) Check {
	pcfg := cfg.PackerRunConfig().Resolve()
	if err := packer.LookPath(pcfg.Bin); err != nil {
		return Check{"packer", Warn, err.Error() + " — pack needs -stub or install repomix"}
	}
	return Check{"packer", OK, pcfg.Bin}
}

func checkTestCmd(cfg config.Config) Check {
	if len(cfg.TestCommand) == 0 {
		return Check{"test_command", Fail, "empty"}
	}
	cmd := cfg.Launched(cfg.TestCommand) // with test_launch the launcher is on the host, the rest in its image
	if _, err := exec.LookPath(cmd[0]); err != nil {
		return Check{"test_command", Warn, fmt.Sprintf("%q not on PATH", cmd[0])}
	}
	return Check{"test_command", OK, fmt.Sprintf("%v", cmd)}
}

func checkState(root string) Check {
	st, ok, err := runstate.Load(root)
	if err != nil {
		return Check{"run_state", Fail, err.Error()}
	}
	if !ok {
		return Check{"run_state", OK, "(none)"}
	}
	return Check{"run_state", OK, fmt.Sprintf("%s @ %s", st.RunID, st.Phase)}
}

// checkFreeze reports the freeze in force and verifies it against the committed spec inputs — the
// check the gates make (freeze.RequireValid), not a look at the recorded hash: a freeze whose PRD
// changed since must not read ok. A change that is not committed yet is work in progress, not drift.
func checkFreeze(root, productSlug string) Check {
	fr, ok, err := freeze.Load(root)
	if err != nil {
		return Check{"freeze", Fail, err.Error()}
	}
	if !ok {
		return Check{"freeze", OK, "(none)"}
	}
	if err := freeze.RequireValid(root, productSlug); err != nil {
		return Check{"freeze", Fail, err.Error()}
	}
	sha := fr.SHA256
	if len(sha) > 12 {
		sha = sha[:12]
	}
	msg := fmt.Sprintf("%s sha=%s… verified against HEAD", fr.PRD, sha)
	if n := len(fr.Inputs); n > 0 {
		msg += fmt.Sprintf(" (%d inputs)", n)
	}
	return Check{"freeze", OK, msg}
}

// Failed reports whether any check is Fail.
func Failed(checks []Check) bool {
	for _, c := range checks {
		if c.Level == Fail {
			return true
		}
	}
	return false
}

func trim(s string) string {
	return strings.TrimSpace(s)
}

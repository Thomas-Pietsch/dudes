// Package config loads engine settings from disk.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/aider"
	"github.com/Thomas-Pietsch/dudes/internal/ollama"
	"github.com/Thomas-Pietsch/dudes/internal/packer"
	"github.com/Thomas-Pietsch/dudes/internal/pi"
	"github.com/Thomas-Pietsch/dudes/internal/product"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// Config is the engine configuration surface.
type Config struct {
	Workspace   string       `json:"workspace"`
	Product     string       `json:"product"`
	Ollama      OllamaConfig `json:"ollama"`
	Aider       AiderConfig  `json:"aider"`
	Packer      PackerConfig `json:"packer"`
	Pi          PiConfig     `json:"pi"`
	TestCommand []string     `json:"test_command"`
	// TestLaunch is the command the test gate runs through — the product's toolchain image, e.g.
	// sh scripts/worker.sh --image golang:1.26-bookworm (CORE §10): a Linux product is tested under
	// Linux. Empty → on the engine's host.
	TestLaunch []string `json:"test_launch,omitempty"`
	// Format runs after each worker pass on the changed files inside the task's scope that end in
	// one of Ext (appended as arguments), through TestLaunch — e.g. goimports: an unused or missing
	// import cost a step in nearly every slice of an instance trial. Empty → no formatting.
	Format   FormatConfig `json:"format,omitempty"`
	Worktree bool         `json:"worktree"`
	// ContextBudget caps the spec input a worker receives, in characters (F73). 0 = no cap.
	ContextBudget int               `json:"context_budget"`
	Convergence   ConvergenceConfig `json:"convergence"`
	// Seal is what the human seal demands of its sealer (F89).
	Seal SealConfig `json:"seal"`
}

// SealConfig holds the demands on the human seal (F89).
type SealConfig struct {
	// RequireSignature refuses an unsigned seal: `dudes seal` stops, before any commit or tag is made,
	// when the sealer's git has no signing key (user.signingkey). Off by default: an unsigned seal
	// still works, and says so.
	RequireSignature bool `json:"require_signature"`
}

// ConvergenceConfig toggles re-analysis passes after the work (F72). Off by
// default: a strong model often finishes in one pass; switch it on for small
// local workers that leave work undone.
type ConvergenceConfig struct {
	Enabled   bool `json:"enabled"`
	MaxPasses int  `json:"max_passes"` // 0 = converge.DefaultMaxPasses
}

// OllamaConfig is optional; empty fields use ollama defaults.
type OllamaConfig struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	// TimeoutSec per generate call (0 = default).
	TimeoutSec int `json:"timeout_sec"`
}

// AiderConfig is optional; empty fields use aider defaults.
type AiderConfig struct {
	Bin        string   `json:"bin"`
	Model      string   `json:"model"`
	TimeoutSec int      `json:"timeout_sec"`
	Extra      []string `json:"extra"`
	// Launch starts the worker through another command — ["sh", "<dudes>/scripts/worker.sh"] runs it
	// in the pinned worker image (CORE §10). Empty → bin runs on the host.
	Launch []string `json:"launch"`
}

// PiConfig is optional; empty fields use pi defaults (F55).
type PiConfig struct {
	Bin        string   `json:"bin"`
	TimeoutSec int      `json:"timeout_sec"`
	Extra      []string `json:"extra"`
	// Launch starts Pi through another command — ["sh", "<dudes>/scripts/worker.sh", "--image", …]
	// runs it in the Pi worker image. Pi has no approval gate: it runs only so. Empty → bin runs on
	// the host.
	Launch []string `json:"launch"`
	// Tools are Pi's own tool names (--tools). Empty → the task's tools list, as before.
	Tools []string `json:"tools"`
}

// PackerConfig is optional; empty fields use Repomix defaults.
type PackerConfig struct {
	Bin        string   `json:"bin"`
	Output     string   `json:"output"`
	TimeoutSec int      `json:"timeout_sec"`
	Compress   *bool    `json:"compress"` // nil = true
	Extra      []string `json:"extra"`
}

// Default returns safe defaults.
func Default() Config {
	return Config{
		Workspace:   ".",
		Product:     "product",
		Ollama:      OllamaConfig{},
		Aider:       AiderConfig{},
		Packer:      PackerConfig{},
		TestCommand: testrun.DefaultFor("."),
	}
}

// Load reads JSON from path. Missing file → Default() and ok=false.
func Load(path string) (Config, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if err := product.ValidateSlug(cfg.Product); err != nil {
				return Config{}, false, err
			}
			return cfg, false, nil
		}
		return Config{}, false, fmt.Errorf("read config: %w", err)
	}
	cfg := Default()
	cfg.TestCommand = nil // resolved below against the configured workspace
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, false, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Workspace == "" {
		cfg.Workspace = "."
	}
	if cfg.Product == "" {
		cfg.Product = "product"
	}
	if err := product.ValidateSlug(cfg.Product); err != nil {
		return Config{}, false, fmt.Errorf("config: %w", err)
	}
	if len(cfg.TestCommand) == 0 {
		cfg.TestCommand = testrun.DefaultFor(cfg.Workspace)
	}
	return cfg, true, nil
}

// OllamaURLEnv names the model host in the user's environment: a machine value, so it comes from
// there or from the user's own dudes.json, never from a repo (F88). Aider reads OLLAMA_API_BASE.
const OllamaURLEnv = "DUDES_OLLAMA_URL"

const aiderURLEnv = "OLLAMA_API_BASE"

// ModelHost is the model host of a prepared run: DUDES_OLLAMA_URL, else the base_url of the user's own
// dudes.json (configured), else the OLLAMA_API_BASE that Aider reads. None set is an error — the
// engine never guesses a host, and a repo's dudes/ does not name one.
func ModelHost(configured string) (string, error) {
	for _, url := range []string{os.Getenv(OllamaURLEnv), configured, os.Getenv(aiderURLEnv)} {
		if url != "" {
			return url, nil
		}
	}
	return "", fmt.Errorf("no model host: set %s (or ollama.base_url in your own dudes.json) — a repository's dudes/ never names the machine's model host", OllamaURLEnv)
}

// Overlay applies a repository's tracked config (dudes/config.json, F88) onto c: the keys it sets
// win, the rest stays as c has it. What a repo must not decide is dropped and named in ignored: the
// workspace (the item's worktree), the machine values (ollama.base_url, the model host) and a waiver
// of the sealer's own demand for signed seals (a repository can switch seal.require_signature on, F89).
func (c Config) Overlay(data []byte) (merged Config, ignored []string, err error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return c, nil, fmt.Errorf("config: parse: %w", err)
	}
	if _, ok := raw["workspace"]; ok {
		delete(raw, "workspace")
		ignored = append(ignored, "workspace")
	}
	if ol, ok := raw["ollama"]; ok {
		var o map[string]json.RawMessage
		if json.Unmarshal(ol, &o) == nil {
			if _, ok := o["base_url"]; ok {
				delete(o, "base_url")
				ignored = append(ignored, "ollama.base_url")
				raw["ollama"], _ = json.Marshal(o)
			}
		}
	}
	// a copy that shares no slice with c, then the repo's keys over it
	base, err := json.Marshal(c)
	if err != nil {
		return c, nil, err
	}
	if err := json.Unmarshal(base, &merged); err != nil {
		return c, nil, err
	}
	repo, err := json.Marshal(raw)
	if err != nil {
		return c, nil, err
	}
	// strict: a key the engine does not know is a typo in a file people edit by hand, and a silently
	// ignored test_launch or format would run the gate where the repository did not mean it
	dec := json.NewDecoder(bytes.NewReader(repo))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&merged); err != nil {
		return c, nil, fmt.Errorf("config: parse: %w", err)
	}
	if c.Seal.RequireSignature && !merged.Seal.RequireSignature {
		merged.Seal.RequireSignature = true
		ignored = append(ignored, "seal.require_signature")
	}
	if merged.Product == "" {
		merged.Product = "product"
	}
	if err := product.ValidateSlug(merged.Product); err != nil {
		return c, nil, fmt.Errorf("config: %w", err)
	}
	return merged, ignored, nil
}

// OllamaClient builds an ollama client from config.
func (c Config) OllamaClient() *ollama.Client {
	cfg := ollama.Config{
		BaseURL: c.Ollama.BaseURL,
		Model:   c.Ollama.Model,
	}
	if c.Ollama.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(c.Ollama.TimeoutSec) * time.Second
	}
	return ollama.New(cfg)
}

// AiderRunConfig builds aider.Config from engine settings.
func (c Config) AiderRunConfig() aider.Config {
	cfg := aider.Config{
		Bin:    c.Aider.Bin,
		Model:  c.Aider.Model,
		Extra:  append([]string{}, c.Aider.Extra...),
		Launch: c.Aider.Launch,
	}
	if c.Aider.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(c.Aider.TimeoutSec) * time.Second
	}
	// One Ollama connection for every worker: Aider reaches Ollama through OLLAMA_API_BASE,
	// so it follows ollama.base_url instead of the shell's environment.
	if c.Ollama.BaseURL != "" {
		cfg.Env = append(cfg.Env, "OLLAMA_API_BASE="+c.Ollama.BaseURL)
	}
	return cfg
}

// PackerRunConfig builds packer.Config from engine settings.
func (c Config) PackerRunConfig() packer.Config {
	cfg := packer.Config{
		Bin:    c.Packer.Bin,
		Output: c.Packer.Output,
		Extra:  append([]string{}, c.Packer.Extra...),
	}
	if c.Packer.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(c.Packer.TimeoutSec) * time.Second
	}
	if c.Packer.Compress == nil {
		cfg.Compress = true
	} else {
		cfg.Compress = *c.Packer.Compress
	}
	return cfg
}

// PiRunConfig builds pi.Config from engine settings.
func (c Config) PiRunConfig() pi.Config {
	cfg := pi.Config{
		Bin:    c.Pi.Bin,
		Extra:  append([]string{}, c.Pi.Extra...),
		Launch: c.Pi.Launch,
		Tools:  append([]string{}, c.Pi.Tools...),
	}
	if c.Pi.TimeoutSec > 0 {
		cfg.Timeout = time.Duration(c.Pi.TimeoutSec) * time.Second
	}
	return cfg
}

// Launched is cmd run through TestLaunch — the test gate as it is executed and recorded.
func (c Config) Launched(cmd []string) []string {
	if len(c.TestLaunch) == 0 {
		return cmd
	}
	return append(append([]string{}, c.TestLaunch...), cmd...)
}

// FormatConfig is the formatter the engine runs on a worker's changed files (Format).
type FormatConfig struct {
	Command []string `json:"command,omitempty"`
	Ext     []string `json:"ext,omitempty"` // e.g. [".go"]; empty → every changed file
}

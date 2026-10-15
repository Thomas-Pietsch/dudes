package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
)

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Workspace != "." {
		t.Fatalf("workspace: got %q want .", cfg.Workspace)
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, loaded, err := config.Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if loaded {
		t.Fatal("expected loaded=false")
	}
	if cfg.Workspace != "." {
		t.Fatalf("workspace: got %q", cfg.Workspace)
	}
}

func TestLoadOK(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dudes.json")
	if err := os.WriteFile(path, []byte(`{"workspace":"C:\\proj"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, loaded, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded {
		t.Fatal("expected loaded=true")
	}
	if cfg.Workspace != `C:\proj` {
		t.Fatalf("workspace: got %q", cfg.Workspace)
	}
}

// Aider reaches the same Ollama as the text workers: ollama.base_url, not the shell.
func TestAiderFollowsOllamaBaseURL(t *testing.T) {
	cfg := config.Default()
	cfg.Ollama.BaseURL = "http://192.0.2.10:11434"
	env := cfg.AiderRunConfig().Env
	if len(env) != 1 || env[0] != "OLLAMA_API_BASE=http://192.0.2.10:11434" {
		t.Fatalf("env %v", env)
	}
}

// The model host is the machine's (F88): the environment, then the user's own dudes.json, then the
// variable Aider reads. None is an error that names the one to set.
func TestModelHost(t *testing.T) {
	t.Setenv(config.OllamaURLEnv, "")
	t.Setenv("OLLAMA_API_BASE", "")
	if _, err := config.ModelHost(""); err == nil || !strings.Contains(err.Error(), config.OllamaURLEnv) {
		t.Fatalf("none set: %v", err)
	}
	t.Setenv("OLLAMA_API_BASE", "http://aider")
	if url, err := config.ModelHost(""); err != nil || url != "http://aider" {
		t.Fatalf("aider's variable: %q %v", url, err)
	}
	if url, err := config.ModelHost("http://file"); err != nil || url != "http://file" {
		t.Fatalf("the user's file over aider's variable: %q %v", url, err)
	}
	t.Setenv(config.OllamaURLEnv, "http://env")
	if url, err := config.ModelHost("http://file"); err != nil || url != "http://env" {
		t.Fatalf("the environment over the file: %q %v", url, err)
	}
}

// A repository's tracked config (dudes/config.json) overlays the user's: its keys win, the rest stays —
// except what a repository must not decide, the workspace and the machine's model host.
func TestOverlay(t *testing.T) {
	user := config.Default()
	user.Workspace = "C:/somewhere"
	user.Product = "demo"
	user.Ollama = config.OllamaConfig{BaseURL: "http://own-host:11434", Model: "own-model", TimeoutSec: 600}
	user.Aider = config.AiderConfig{Bin: "aider", Launch: []string{"sh", "../../../scripts/worker.sh"}, Extra: []string{"--chat-language", "English"}}
	user.TestCommand = []string{"go", "test", "./..."}

	repo := `{
  "workspace": "/repo/says/this",
  "product": "bookshelf",
  "ollama": {"base_url": "http://repo-host:11434", "model": "qwen3.8:27b"},
  "aider": {"model": "ollama_chat/qwen3.8:27b", "extra": ["--edit-format", "whole"]},
  "test_launch": ["sh", "worker.sh", "--image", "golang:1.26"],
  "format": {"command": ["goimports", "-w"], "ext": [".go"]}
}`
	got, ignored, err := user.Overlay([]byte(repo))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(ignored, ",") != "workspace,ollama.base_url" {
		t.Errorf("ignored %v", ignored)
	}
	if got.Workspace != "C:/somewhere" || got.Ollama.BaseURL != "http://own-host:11434" {
		t.Errorf("the repository decided %q / %q", got.Workspace, got.Ollama.BaseURL)
	}
	if got.Product != "bookshelf" || got.Ollama.Model != "qwen3.8:27b" || got.Ollama.TimeoutSec != 600 {
		t.Errorf("product %q ollama %+v: the repository's keys win, the user's others stay", got.Product, got.Ollama)
	}
	if got.Aider.Model != "ollama_chat/qwen3.8:27b" || got.Aider.Bin != "aider" || strings.Join(got.Aider.Launch, " ") != "sh ../../../scripts/worker.sh" {
		t.Errorf("aider %+v", got.Aider)
	}
	if strings.Join(got.Aider.Extra, " ") != "--edit-format whole" { // a list is replaced, not merged
		t.Errorf("aider extra %v", got.Aider.Extra)
	}
	if strings.Join(got.TestCommand, " ") != "go test ./..." || len(got.TestLaunch) != 4 || strings.Join(got.Format.Command, " ") != "goimports -w" {
		t.Errorf("test command %v launch %v format %+v", got.TestCommand, got.TestLaunch, got.Format)
	}
	// the user's config is untouched, its slices unshared
	if user.Product != "demo" || user.Ollama.Model != "own-model" || strings.Join(user.Aider.Extra, " ") != "--chat-language English" {
		t.Errorf("the user's config changed: %+v", user)
	}

	if _, _, err := user.Overlay([]byte(`{"product": "Not A Slug"}`)); err == nil {
		t.Error("an invalid product")
	}
	if _, _, err := user.Overlay([]byte(`{`)); err == nil {
		t.Error("broken JSON")
	}
	if _, _, err := user.Overlay([]byte(`{"test_cmd": ["make", "test"]}`)); err == nil || !strings.Contains(err.Error(), "test_cmd") {
		t.Errorf("a misspelt key must not be ignored: %v", err)
	}
}

// The test gate runs through test_launch — the product's toolchain image — when one is set.
func TestTestLaunch(t *testing.T) {
	p := filepath.Join(t.TempDir(), "dudes.json")
	if err := os.WriteFile(p, []byte(`{"test_command":["go","test","./..."],"test_launch":["sh","worker.sh","--image","golang:1.26"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := config.Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(cfg.Launched(cfg.TestCommand), " "); got != "sh worker.sh --image golang:1.26 go test ./..." {
		t.Fatalf("launched: %q", got)
	}
	cfg.TestLaunch = nil
	if got := strings.Join(cfg.Launched(cfg.TestCommand), " "); got != "go test ./..." {
		t.Fatalf("on the host: %q", got)
	}
}

// A signed seal is asked for in dudes.json (F89): off by default, and the example file names the key.
func TestSealRequireSignature(t *testing.T) {
	if config.Default().Seal.RequireSignature {
		t.Fatal("an unsigned seal must work by default")
	}
	p := filepath.Join(t.TempDir(), "dudes.json")
	if err := os.WriteFile(p, []byte(`{"seal":{"require_signature":true}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, _, err := config.Load(p)
	if err != nil || !cfg.Seal.RequireSignature {
		t.Fatalf("require_signature: %+v %v", cfg.Seal, err)
	}

	example, err := os.ReadFile(filepath.Join("..", "..", "dudes.example.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(example, &raw); err != nil {
		t.Fatal(err)
	}
	var seal struct {
		RequireSignature *bool `json:"require_signature"`
	}
	if err := json.Unmarshal(raw["seal"], &seal); err != nil || seal.RequireSignature == nil || *seal.RequireSignature {
		t.Fatalf(`dudes.example.json should hold "seal": {"require_signature": false}, got %s (%v)`, raw["seal"], err)
	}
}

// A repository's tracked config can ask for signed seals, never waive what the sealer asks for.
func TestOverlaySealRequireSignature(t *testing.T) {
	for _, tc := range []struct {
		name        string
		user        bool
		repo        string
		want        bool
		wantIgnored string
	}{
		{"neither asks", false, `{"product": "demo"}`, false, ""},
		{"the repository asks", false, `{"seal": {"require_signature": true}}`, true, ""},
		{"the sealer asks, the repository is silent", true, `{"product": "demo"}`, true, ""},
		{"the sealer asks, the repository agrees", true, `{"seal": {"require_signature": true}}`, true, ""},
		{"the sealer asks, the repository waives it", true, `{"seal": {"require_signature": false}}`, true, "seal.require_signature"},
	} {
		user := config.Default()
		user.Seal.RequireSignature = tc.user
		got, ignored, err := user.Overlay([]byte(tc.repo))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if got.Seal.RequireSignature != tc.want || strings.Join(ignored, ",") != tc.wantIgnored {
			t.Errorf("%s: require_signature=%v ignored=%v, want %v and %q", tc.name, got.Seal.RequireSignature, ignored, tc.want, tc.wantIgnored)
		}
		if user.Seal.RequireSignature != tc.user {
			t.Errorf("%s: the user's config changed", tc.name)
		}
	}
}

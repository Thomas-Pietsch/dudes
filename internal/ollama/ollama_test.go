package ollama_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/ollama"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestGenerateSingleShot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" {
			t.Fatalf("path %s", r.URL.Path)
		}
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["stream"] != false {
			t.Fatalf("stream=%v", req["stream"])
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"response": "# PRD\n\nok\n"})
	}))
	defer srv.Close()

	c := ollama.New(ollama.Config{BaseURL: srv.URL, Model: "test"})
	out, err := c.Generate(context.Background(), "hello")
	if err != nil {
		t.Fatal(err)
	}
	if out != "# PRD\n\nok" {
		t.Fatalf("got %q", out)
	}
}

// A call's cost is Ollama's own count of its tokens, from the generate reply (F88).
func TestGenerateUsage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"response": "# QR\n", "done": true, "prompt_eval_count": 1834, "eval_count": 211, "total_duration": 4200000000})
	}))
	defer srv.Close()
	c := ollama.New(ollama.Config{BaseURL: srv.URL, Model: "test"})
	out, u, err := c.GenerateUsage(context.Background(), "hello")
	if err != nil || out != "# QR" {
		t.Fatalf("out %q err %v", out, err)
	}
	if u.PromptTokens != 1834 || u.CompletionTokens != 211 {
		t.Fatalf("usage %+v", u)
	}
}

func TestPhasePromptIncludesPrior(t *testing.T) {
	p := ollama.PhasePrompt(phase.Spec, "demo", "build x", "## PS\nprior", "pack-body", "demo_PRD.md", "", "", "")
	if !strings.Contains(p, "prior") || !strings.Contains(p, "PRD") || !strings.Contains(p, "pack-body") || !strings.Contains(p, "demo_PRD.md") || !strings.Contains(p, "BEGIN ARTIFACT") || !strings.Contains(p, "### Requirement:") {
		t.Fatalf("prompt=%s", p)
	}
}

func TestPhasePromptDiscipline(t *testing.T) {
	p := ollama.PhasePrompt(phase.Idea, "demo", "x", "", "", "", "", "", "### Skill: tdd\n\nRed-green.")
	if !strings.Contains(p, "Skill: tdd") || !strings.Contains(p, "this phase only") {
		t.Fatalf("prompt=%s", p)
	}
}

func TestPhasePromptTaskOverride(t *testing.T) {
	p := ollama.PhasePrompt(phase.Idea, "demo", "x", "", "", "", "Custom recipe task.", "", "")
	if !strings.Contains(p, "Custom recipe task.") || strings.Contains(p, "idea capture") {
		t.Fatalf("prompt=%s", p)
	}
}

func TestPhasePromptFeedback(t *testing.T) {
	p := ollama.PhasePrompt(phase.QA, "demo", "x", "", "", "demo_PRD.md", "", "CLI flag missing", "")
	if !strings.Contains(p, "Human rejection feedback") || !strings.Contains(p, "CLI flag missing") {
		t.Fatalf("prompt=%s", p)
	}
}

func TestListModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Fatalf("path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"models": []map[string]string{{"name": "qwen2.5-coder:7b"}},
		})
	}))
	defer srv.Close()
	c := ollama.New(ollama.Config{BaseURL: srv.URL, Model: "qwen2.5-coder:7b"})
	names, err := c.ListModels(context.Background())
	if err != nil || len(names) != 1 || names[0] != "qwen2.5-coder:7b" {
		t.Fatalf("names=%v err=%v", names, err)
	}
}

func TestFixPromptUsesOnlyTestOutput(t *testing.T) {
	p := ollama.FixFromTestPrompt("FAIL: boom")
	if !strings.Contains(p, "FAIL: boom") || !strings.Contains(p, "ONLY") {
		t.Fatalf("prompt=%s", p)
	}
}

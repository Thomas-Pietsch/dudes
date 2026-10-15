// Package ollama is a bounded text worker: one HTTP generate call = fresh context (reset).
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Thomas-Pietsch/dudes/internal/redact"
)

// Config for the local OpenAI-compatible-ish Ollama generate API.
type Config struct {
	BaseURL string        // e.g. http://127.0.0.1:11434
	Model   string        // e.g. qwen2.5-coder:7b
	Timeout time.Duration // per request
}

// DefaultConfig returns local Ollama defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL: "http://127.0.0.1:11434",
		Model:   "qwen2.5-coder:7b",
		Timeout: 10 * time.Minute,
	}
}

// Client talks to Ollama /api/generate (non-streaming).
type Client struct {
	cfg    Config
	client *http.Client
}

// New creates a client. Empty BaseURL/Model filled from defaults.
func New(cfg Config) *Client {
	d := DefaultConfig()
	if cfg.BaseURL == "" {
		cfg.BaseURL = d.BaseURL
	}
	if cfg.Model == "" {
		cfg.Model = d.Model
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = d.Timeout
	}
	return &Client{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

type generateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResp struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
	// Ollama's own counts of the call: the prompt's tokens it evaluated, and the reply's.
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}

// Usage is what one generate call cost, as Ollama counts it: the tokens of the prompt it evaluated
// (prompt_eval_count — a prompt it has cached counts less) and of the reply (eval_count).
type Usage struct {
	PromptTokens, CompletionTokens int
}

// Generate runs a single-shot prompt with no chat history (phase context reset).
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	out, _, err := c.GenerateUsage(ctx, prompt)
	return out, err
}

// GenerateUsage is Generate, with Ollama's counts of the call's tokens (F88).
func (c *Client) GenerateUsage(ctx context.Context, prompt string) (string, Usage, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", Usage{}, fmt.Errorf("empty prompt")
	}
	body, err := json.Marshal(generateReq{
		Model:  c.cfg.Model,
		Prompt: prompt,
		Stream: false,
	})
	if err != nil {
		return "", Usage{}, err
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", Usage{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return "", Usage{}, fmt.Errorf("ollama request: %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return "", Usage{}, err
	}
	if res.StatusCode >= 300 {
		return "", Usage{}, fmt.Errorf("ollama HTTP %d: %s", res.StatusCode, truncate(redact.String(string(data)), 400))
	}
	var parsed generateResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", Usage{}, fmt.Errorf("ollama decode: %w", err)
	}
	u := Usage{PromptTokens: parsed.PromptEvalCount, CompletionTokens: parsed.EvalCount}
	if parsed.Error != "" {
		return "", u, fmt.Errorf("ollama: %s", redact.String(parsed.Error))
	}
	out := strings.TrimSpace(parsed.Response)
	if out == "" {
		return "", u, fmt.Errorf("ollama: empty response")
	}
	return out, u, nil
}

// BaseURL returns the configured Ollama base URL.
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// Model returns the configured model name.
func (c *Client) Model() string { return c.cfg.Model }

type tagsResp struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// ListModels probes GET /api/tags (connectivity + installed models).
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama tags: %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama HTTP %d: %s", res.StatusCode, truncate(redact.String(string(data)), 200))
	}
	var parsed tagsResp
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("ollama tags decode: %w", err)
	}
	names := make([]string, 0, len(parsed.Models))
	for _, m := range parsed.Models {
		if m.Name != "" {
			names = append(names, m.Name)
		}
	}
	return names, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

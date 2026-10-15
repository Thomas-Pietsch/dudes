package redact_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/redact"
)

func TestRedactAPIKey(t *testing.T) {
	in := `failed api_key=sk-secretvalue123456 password=hunter2`
	out := redact.String(in)
	if strings.Contains(out, "sk-secretvalue") || strings.Contains(out, "hunter2") {
		t.Fatalf("not redacted: %s", out)
	}
	if !strings.Contains(out, "[REDACTED]") {
		t.Fatalf("missing marker: %s", out)
	}
}

func TestRedactBearerAndGitHub(t *testing.T) {
	in := "Authorization: Bearer abcdefghijklmnop ghp_abcdefghijklmnopqrstuvwxyz012345"
	out := redact.String(in)
	if strings.Contains(out, "abcdefghijklmnop") || strings.Contains(out, "ghp_abcd") {
		t.Fatalf("got %s", out)
	}
}

func TestEmpty(t *testing.T) {
	if redact.String("") != "" {
		t.Fatal("empty")
	}
}

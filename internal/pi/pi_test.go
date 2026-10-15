package pi_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/pi"
)

func TestLookPathMissing(t *testing.T) {
	if err := pi.LookPath("dudes-pi-binary-that-does-not-exist"); err == nil {
		t.Fatal("expected missing binary")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := pi.DefaultConfig()
	if cfg.Bin != "pi" {
		t.Fatalf("bin=%s", cfg.Bin)
	}
}

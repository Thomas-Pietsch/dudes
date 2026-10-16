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

func TestExeIsTheLauncherElseTheBinary(t *testing.T) {
	c := pi.Config{Bin: "pi", Launch: []string{"sh", "worker.sh"}}
	if c.Exe() != "sh" {
		t.Fatalf("exe=%s", c.Exe())
	}
	if (pi.Config{}).Exe() != "pi" {
		t.Fatal("default exe must be pi")
	}
}

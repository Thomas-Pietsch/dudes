package doctor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/config"
)

// doctor starts the worker through its launcher, as the engine does, and reports its version.
func TestCheckAiderThroughLauncher(t *testing.T) {
	dir := t.TempDir()
	launcher := filepath.Join(dir, "launch")
	body := "#!/bin/sh\n[ \"$1\" = aider ] && [ \"$2\" = --version ] && echo aider 9.9.9\n"
	if runtime.GOOS == "windows" {
		launcher += ".bat"
		body = "@echo off\r\nif \"%1\"==\"aider\" if \"%2\"==\"--version\" echo aider 9.9.9\r\n"
	}
	if err := os.WriteFile(launcher, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Workspace = dir
	cfg.Aider.Bin = "aider"
	cfg.Aider.Launch = []string{launcher}
	c := checkAider(cfg)
	if c.Level != OK || !strings.Contains(c.Message, "aider 9.9.9") {
		t.Fatalf("check = %+v", c)
	}
	cfg.Aider.Launch = []string{filepath.Join(dir, "missing-launcher")}
	if c := checkAider(cfg); c.Level == OK {
		t.Fatalf("a missing launcher must not be ok: %+v", c)
	}
}

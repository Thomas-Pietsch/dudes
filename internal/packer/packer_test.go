package packer_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/packer"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestIncludesForImpl(t *testing.T) {
	inc := packer.IncludesFor(phase.Implementation, "demo")
	joined := strings.Join(inc, ",")
	if !strings.Contains(joined, "cmd/**") || !strings.Contains(joined, "demo_PRD.md") {
		t.Fatalf("includes=%v", inc)
	}
}

func TestWriteStubPack(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.MkdirAll(filepath.Join(dir, "cmd", "x"), 0o755)
	_ = os.WriteFile(filepath.Join(dir, "cmd", "x", "main.go"), []byte("package main\n"), 0o644)

	rel, err := packer.WriteStubPack(dir, []string{"go.mod", "cmd/**"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if rel != packer.DefaultOutputRel {
		t.Fatalf("rel=%s", rel)
	}
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	body := string(data)
	if !strings.Contains(body, "go.mod") || !strings.Contains(body, "cmd/x/main.go") {
		t.Fatalf("body=%s", body)
	}
}

func TestRunWithFakeBinary(t *testing.T) {
	dir := t.TempDir()
	bin := writeFakeRepomix(t, dir)
	outRel := ".dudes/context_pack.md"
	rel, err := packer.Run(context.Background(), dir, []string{"*.md"}, packer.Config{
		Bin:      bin,
		Output:   outRel,
		Compress: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if rel != outRel {
		t.Fatalf("rel=%s", rel)
	}
	if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(outRel))); err != nil {
		t.Fatal(err)
	}
}

func writeFakeRepomix(t *testing.T, dir string) string {
	t.Helper()
	var path, body string
	if runtime.GOOS == "windows" {
		path = filepath.Join(dir, "fake-repomix.bat")
		// Write to the --output path: parse is hard in bat; write fixed relative path.
		body = "@echo off\r\nmkdir .dudes 2>nul\r\necho fake pack>.dudes\\context_pack.md\r\n"
	} else {
		path = filepath.Join(dir, "fake-repomix")
		body = "#!/bin/sh\nmkdir -p .dudes\necho fake pack > .dudes/context_pack.md\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

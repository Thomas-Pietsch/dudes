package extract_test

import (
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/extract"
)

func TestArtifactDelimited(t *testing.T) {
	raw := "Sure!\n" + extract.BeginMark + "\n# PRD\n\nok\n" + extract.EndMark + "\nbye"
	out := extract.Artifact(raw)
	if !strings.HasPrefix(out, "# PRD") || strings.Contains(out, "Sure") || strings.Contains(out, "bye") {
		t.Fatalf("got %q", out)
	}
}

func TestArtifactFence(t *testing.T) {
	raw := "Here you go:\n```markdown\n# Idea\n\nx\n```\n"
	out := extract.Artifact(raw)
	if !strings.Contains(out, "# Idea") || strings.Contains(out, "Here you go") {
		t.Fatalf("got %q", out)
	}
}

func TestArtifactPreambleStrip(t *testing.T) {
	raw := "Of course, here is the document:\n# Title\n\nbody\n"
	out := extract.Artifact(raw)
	if !strings.HasPrefix(out, "# Title") {
		t.Fatalf("got %q", out)
	}
}

func TestControlJSON(t *testing.T) {
	var dest struct {
		OK   bool   `json:"ok"`
		File string `json:"file"`
	}
	raw := "result:\n```json\n{\"ok\":true,\"file\":\"a.md\"}\n```\n"
	if err := extract.ControlJSON(raw, &dest); err != nil {
		t.Fatal(err)
	}
	if !dest.OK || dest.File != "a.md" {
		t.Fatalf("%+v", dest)
	}
}

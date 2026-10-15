package artifact_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/artifact"
	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestPath(t *testing.T) {
	rel, ok := artifact.Path("demo", phase.Strategy)
	if !ok || rel != "demo_PS.md" {
		t.Fatalf("got %q ok=%v", rel, ok)
	}
	if _, ok := artifact.Path("demo", phase.Sealed); ok {
		t.Fatal("sealed should have no artifact")
	}
}

func TestWriteRejectsEscape(t *testing.T) {
	dir := t.TempDir()
	if err := artifact.Write(dir, "../nope.md", []byte("x")); err == nil {
		t.Fatal("expected escape error")
	}
}

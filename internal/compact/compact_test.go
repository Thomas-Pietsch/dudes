package compact

import (
	"strings"
	"testing"
)

func TestTextShortUnchanged(t *testing.T) {
	s := "hello"
	if Text(s, 100) != s {
		t.Fatal("short text changed")
	}
}

func TestTextHeadTail(t *testing.T) {
	s := strings.Repeat("A", 1000) + "MID" + strings.Repeat("Z", 1000)
	out := Text(s, 200)
	if len(out) > 200+10 { // small slack for marker math
		t.Fatalf("len=%d", len(out))
	}
	if !strings.HasPrefix(out, "AAA") || !strings.HasSuffix(out, "ZZZ") {
		t.Fatalf("head/tail missing: %q…%q", out[:20], out[len(out)-20:])
	}
	if !strings.Contains(out, "middle omitted") {
		t.Fatalf("missing marker: %s", out)
	}
	if strings.Contains(out, "MID") {
		t.Fatal("middle should be omitted")
	}
}

func TestNote(t *testing.T) {
	if Note(10, 100) != "" {
		t.Fatal("expected empty note")
	}
	if Note(1000, 100) == "" {
		t.Fatal("expected note")
	}
}

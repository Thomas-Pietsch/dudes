package phase_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/phase"
)

func TestDefaultFirstAndChain(t *testing.T) {
	r := phase.Default()
	if r.First() != phase.Idea {
		t.Fatalf("first: %s", r.First())
	}
	cur := r.First()
	seen := map[phase.ID]bool{cur: true}
	for {
		next, ok := r.Next(cur)
		if !ok {
			break
		}
		if seen[next] {
			t.Fatalf("cycle at %s", next)
		}
		seen[next] = true
		cur = next
	}
	if cur != phase.Sealed {
		t.Fatalf("last: got %s want sealed", cur)
	}
	if len(seen) != len(phase.DefaultOrder) {
		t.Fatalf("visited %d want %d", len(seen), len(phase.DefaultOrder))
	}
}

func TestNextUnknown(t *testing.T) {
	r := phase.Default()
	if _, ok := r.Next("nope"); ok {
		t.Fatal("expected false")
	}
}

func TestParse(t *testing.T) {
	r := phase.Default()
	id, err := r.Parse("qa")
	if err != nil || id != phase.QA {
		t.Fatalf("parse qa: %v %s", err, id)
	}
	if _, err := r.Parse("nope"); err == nil {
		t.Fatal("expected error")
	}
}

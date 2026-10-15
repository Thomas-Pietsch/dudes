package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/itempack"
	"github.com/Thomas-Pietsch/dudes/internal/testrun"
)

// demoItems is the itempack DEMO spec set with every QR mark set to qrMark.
func demoItems(t *testing.T, qrMark string) *itempack.Set {
	t.Helper()
	src := filepath.Join("..", "..", "internal", "itempack", "testdata", "demo")
	dir := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if e.Name() == "DEMO_QR.md" {
			data = []byte(strings.ReplaceAll(string(data), "⬜", qrMark))
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(filepath.Join(dir, "profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	p, err := itempack.ParseProfile(data)
	if err != nil {
		t.Fatal(err)
	}
	var docs []itempack.Doc
	for key, name := range map[string]string{"STRATEGY": "STRATEGY.md", "PS": "DEMO_PS.md", "PRD": "DEMO_PRD.md",
		"TA": "DEMO_TA.md", "UX": "DEMO_UX.md", "ROADMAP": "DEMO_ROADMAP.md", "BACKLOG": "DEMO_BACKLOG.md",
		"SC": "DEMO_SC.md", "QR": "DEMO_QR.md", "BM": "DEMO_BM.md"} {
		docs = append(docs, itempack.Doc{Key: key, Path: filepath.Join(dir, name)})
	}
	s, err := itempack.New(p, docs)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// An item is done when its QR DoD IDs are green and all its acceptance tests pass in the recorded
// acceptance run; without a recorded run its status is the QR's alone.
func TestItemStatusFromAcceptanceEvidence(t *testing.T) {
	for _, c := range []struct {
		name, qr string
		results  map[string]string
		status   string
		ac       string
	}{
		{"no evidence, QR green", "🟢", nil, itempack.StatusDone, "-"},
		{"no evidence, QR open", "⬜", nil, itempack.StatusOpen, "-"},
		{"other item's tests only", "🟢", map[string]string{"TestDM002AC1": testrun.Pass}, itempack.StatusDone, "-"},
		{"AC green, QR green", "🟢", map[string]string{"TestDM001AC1": testrun.Pass}, itempack.StatusDone, "AC 1/1"},
		{"AC skipped, QR green", "🟢", map[string]string{"TestDM001AC1": testrun.Skip}, itempack.StatusOpen, "AC 0/1"},
		{"AC failed, QR green", "🟢", map[string]string{"TestDM001AC1": testrun.Fail}, itempack.StatusOpen, "AC 0/1"},
		{"AC green, QR open", "⬜", map[string]string{"TestDM001AC1": testrun.Pass}, itempack.StatusOpen, "AC 1/1"},
	} {
		s := demoItems(t, c.qr)
		it, ok := s.FindItem("DM-001")
		if !ok {
			t.Fatal("DM-001 missing")
		}
		st, ac := itemStatus(s, it, nil, c.results)
		if st != c.status || ac.String() != c.ac {
			t.Errorf("%s: status %s %s, want %s %s", c.name, st, ac, c.status, c.ac)
		}
	}
	var b bytes.Buffer
	printBoard(&b, demoItems(t, "🟢"), map[string]string{"TestDM001AC1": testrun.Fail})
	if !strings.Contains(b.String(), "| Item | Title | DoD | AC | Status |") || !strings.Contains(b.String(), "| AC 0/1 | open |") {
		t.Fatalf("board:\n%s", b.String())
	}
}

package contract_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/contract"
	"github.com/Thomas-Pietsch/dudes/internal/scope"
)

func TestSampleExampleValidates(t *testing.T) {
	_, this, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(this), "..", "..", "examples", "sample", "contract.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var c contract.Contract
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.ID != "bk-001" || c.Product != "bookshelf" || c.Item != "BK-001" {
		t.Fatalf("%+v", c)
	}
	// the red-first task pauses for a human once its tests are written: nobody codes before they are read
	red, ok := c.TaskByID("test-add")
	if !ok || red.Expect != "fail" || !red.NeedsAfterDecision() {
		t.Fatal("test-add must be red first and pause for a human")
	}
	// the code task cannot change the tests it must pass
	code, ok := c.TaskByID("impl-add")
	if !ok || !code.NeedsAfterDecision() {
		t.Fatal("impl-add must exist and pause for a human")
	}
	for _, test := range red.Writable {
		if scope.Inside(test, code.Writable) {
			t.Fatalf("impl-add may change %s, a test file of test-add", test)
		}
	}
	var field bool
	for _, a := range c.Acceptance {
		if a.Kind == "field" {
			field = true
		}
	}
	if !field {
		t.Fatal("expected a field acceptance row")
	}
}

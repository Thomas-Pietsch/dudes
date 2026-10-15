package product_test

import (
	"testing"

	"github.com/Thomas-Pietsch/dudes/internal/product"
)

func TestValidateSlug(t *testing.T) {
	if err := product.ValidateSlug("demo"); err != nil {
		t.Fatal(err)
	}
	if err := product.ValidateSlug("Demo"); err == nil {
		t.Fatal("uppercase")
	}
	if err := product.ValidateSlug("../x"); err == nil {
		t.Fatal("path")
	}
	if err := product.ValidateSlug(""); err == nil {
		t.Fatal("empty")
	}
}

func TestSpecSet(t *testing.T) {
	s := product.SpecSet("demo")
	if len(s) != 5 || s[2] != "demo_PRD.md" || s[3] != "demo_TA.md" {
		t.Fatalf("%v", s)
	}
}

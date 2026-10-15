// Package product validates and names the single product identity (F13).
package product

import (
	"fmt"
	"regexp"
	"strings"
)

var slugRE = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)

// ValidateSlug ensures one stable product id (no path sprawl).
func ValidateSlug(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("product: empty")
	}
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return fmt.Errorf("product: must not contain path separators")
	}
	if !slugRE.MatchString(name) {
		return fmt.Errorf("product %q: use lowercase slug [a-z][a-z0-9_-]{0,63}", name)
	}
	return nil
}

// SpecSet returns the canonical one-spec-set relative paths for a product (F13).
func SpecSet(name string) []string {
	if name == "" {
		name = "product"
	}
	return []string{
		name + "_IDEA.md",
		name + "_PS.md",
		name + "_PRD.md",
		name + "_TA.md",
		name + "_QR.md",
	}
}

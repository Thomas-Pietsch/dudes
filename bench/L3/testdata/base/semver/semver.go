// Package semver orders versions per TA ALG-01 — the L3 start: L2's behaviour, already built.
package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version is a parsed MAJOR.MINOR.PATCH with an optional release candidate number (-1: none).
type Version struct{ Major, Minor, Patch, RC int }

var form = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-rc\.(\d+))?$`)

// Parse reads s per ALG-01.1.
func Parse(s string) (Version, error) {
	m := form.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("ERR_VERSION: %q", s)
	}
	n := func(x string) int { v, _ := strconv.Atoi(x); return v }
	v := Version{n(m[1]), n(m[2]), n(m[3]), -1}
	if m[4] != "" {
		v.RC = n(m[4])
	}
	return v, nil
}

// Compare orders a and b per ALG-01.2–4.
func Compare(a, b string) (int, error) {
	x, err := Parse(a)
	if err != nil {
		return 0, err
	}
	y, err := Parse(b)
	if err != nil {
		return 0, err
	}
	for _, p := range [][2]int{{x.Major, y.Major}, {x.Minor, y.Minor}, {x.Patch, y.Patch}, {rank(x), rank(y)}} {
		if p[0] < p[1] {
			return -1, nil
		}
		if p[0] > p[1] {
			return 1, nil
		}
	}
	return 0, nil
}

// rank puts a release after all of its candidates.
func rank(v Version) int {
	if v.RC < 0 {
		return 1 << 30
	}
	return v.RC
}

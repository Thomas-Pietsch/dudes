// Hidden verification of L3 (BL-003, TA ALG-02 on top of ALG-01): the worker never sees this file.
// It checks every rule of the new package and that semver kept its behaviour.
package verify_test

import (
	"strings"
	"testing"

	"bench/release"
	"bench/semver"
)

func TestLatestVerify(t *testing.T) {
	for _, c := range []struct {
		tags []string
		want string
	}{
		{[]string{"v1.9.0", "v1.10.0-rc.1", "nightly", "1.10.0"}, "1.10.0"}, // the item's AC-1
		{[]string{"v2.0.0-rc.1", "v2.0.0-rc.2"}, "v2.0.0-rc.2"},             // ALG-02.2 only candidates
		{[]string{"v3.0.0-rc.1", "v2.9.9"}, "v2.9.9"},                       // ALG-02.2 a final beats a higher candidate
		{[]string{"1.2.0-rc.10", "1.2.0-rc.9"}, "1.2.0-rc.10"},              // ALG-01.3 candidates by number
		{[]string{"release-1", "v1.0.0", "1.0", "latest"}, "v1.0.0"},        // ALG-02.1 skip, ALG-02.3 as given
		{[]string{"v1.10.0", "v1.9.0", "v1.2.0"}, "v1.10.0"},                // ALG-01.2 numbers, not strings
	} {
		got, err := release.Latest(c.tags)
		if err != nil || got != c.want {
			t.Errorf("Latest(%q) = %q, %v; want %q", c.tags, got, err, c.want)
		}
	}
	for _, tags := range [][]string{nil, {}, {"nightly", "1.x.0", "v1.2"}} {
		if _, err := release.Latest(tags); err == nil || !strings.HasPrefix(err.Error(), "ERR_NO_RELEASE:") {
			t.Errorf("Latest(%q) error %v, want ERR_NO_RELEASE:", tags, err)
		}
	}
}

func TestIsCandidateVerify(t *testing.T) {
	for in, want := range map[string]bool{"1.2.0-rc.1": true, "v2.0.0-rc.10": true, "1.2.0": false, "v1.0.0": false, "1.2.0-rc.": false, "nightly": false} {
		if got := semver.IsCandidate(in); got != want {
			t.Errorf("IsCandidate(%q) = %v, want %v", in, got, want)
		}
	}
}

// semver keeps ALG-01 (the L2 cases).
func TestSemverKeptVerify(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want int
	}{
		{"v1.10.0", "1.9.3", 1}, {"1.2.3", "1.2.10", -1}, {"1.2.0-rc.2", "1.2.0-rc.10", -1},
		{"1.2.0-rc.10", "1.2.0", -1}, {"v2.0.0", "2.0.0", 0},
	} {
		if got, err := semver.Compare(c.a, c.b); err != nil || got != c.want {
			t.Errorf("Compare(%q, %q) = %d, %v; want %d", c.a, c.b, got, err, c.want)
		}
	}
	for _, in := range []string{"1.x.0", "1.2", "-1.2.3", "1.2.3-rc."} {
		if _, err := semver.Parse(in); err == nil || !strings.HasPrefix(err.Error(), "ERR_VERSION:") {
			t.Errorf("Parse(%q) error %v, want ERR_VERSION:", in, err)
		}
	}
}

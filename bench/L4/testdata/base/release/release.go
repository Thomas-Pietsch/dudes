// Package release names the latest release among tags per TA ALG-02. The L4 start: L3's behaviour,
// already built. The item does not touch it.
package release

import (
	"fmt"

	"bench/semver"
)

// Latest returns the tag of the latest release per ALG-02.1 to ALG-02.4: the highest final release,
// or the highest candidate when no final release exists, as the tag was given.
func Latest(tags []string) (string, error) {
	best, bestFinal := "", false
	for _, t := range tags {
		if _, err := semver.Parse(t); err != nil {
			continue // ALG-02.1: a tag that does not parse is skipped
		}
		final := !semver.IsCandidate(t)
		switch {
		case best == "", final && !bestFinal:
			best, bestFinal = t, final // the first usable tag, or a final release over a candidate (ALG-02.2)
		case final != bestFinal:
			// a candidate never replaces a final release
		default:
			if c, _ := semver.Compare(t, best); c > 0 {
				best = t
			}
		}
	}
	if best == "" {
		return "", fmt.Errorf("ERR_NO_RELEASE: no tag of %d parses", len(tags))
	}
	return best, nil
}

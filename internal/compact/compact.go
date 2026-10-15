// Package compact trims long text for worker prompts (F28-lite).
// Prefer phase context reset (F21); this is a secondary safety net for oversized
// durable artifacts / packs injected into a single invoke.
package compact

import "fmt"

// DefaultMax is the default rune/byte budget for injected context blobs.
const DefaultMax = 16000

const marker = "\n\n…[dudes compact: middle omitted]…\n\n"

// Text keeps the head and tail of s when longer than max bytes.
// max <= 0 uses DefaultMax.
func Text(s string, max int) string {
	if max <= 0 {
		max = DefaultMax
	}
	if len(s) <= max {
		return s
	}
	overhead := len(marker)
	if max <= overhead+32 {
		return s[:max] + "…"
	}
	keep := (max - overhead) / 2
	head := s[:keep]
	tail := s[len(s)-keep:]
	return head + marker + tail
}

// Note returns a short human hint when compaction applied (empty if not).
func Note(originalLen, max int) string {
	if max <= 0 {
		max = DefaultMax
	}
	if originalLen <= max {
		return ""
	}
	return fmt.Sprintf("compacted %d→~%d bytes", originalLen, max)
}

// Package redact masks common secret patterns in logs and error text (F44).
package redact

import (
	"regexp"
)

var patterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)bearer\s+[a-z0-9\-._~+/]+=*`),
	regexp.MustCompile(`\b(sk-[a-zA-Z0-9]{10,})\b`),
	regexp.MustCompile(`\b(ghp_[a-zA-Z0-9]{20,})\b`),
	regexp.MustCompile(`\b(gho_[a-zA-Z0-9]{20,})\b`),
	regexp.MustCompile(`\b(xox[baprs]-[a-zA-Z0-9-]{10,})\b`),
	regexp.MustCompile(`\b(AKIA[0-9A-Z]{12,})\b`),
	regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|passwd)\s*[=:]\s*["']?[^\s"'\\]+`),
}

const replacement = "[REDACTED]"

// String returns s with likely secrets replaced.
func String(s string) string {
	if s == "" {
		return s
	}
	out := s
	for _, re := range patterns {
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			// Keep key labels when present: "api_key=…" → "api_key=[REDACTED]"
			if i := indexSep(m); i >= 0 {
				return m[:i+1] + replacement
			}
			return replacement
		})
	}
	return out
}

func indexSep(m string) int {
	for i, c := range m {
		if c == '=' || c == ':' {
			return i
		}
	}
	return -1
}

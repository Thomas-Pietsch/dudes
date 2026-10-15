// Package extract cleans worker free-form output into durable artifacts (F25).
package extract

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const (
	BeginMark = "===BEGIN ARTIFACT==="
	EndMark   = "===END ARTIFACT==="
)

var fenceRE = regexp.MustCompile("(?s)```(?:markdown|md|text)?\\s*\\n(.*?)```")

// Artifact prefers delimited blocks, then fenced code, else trimmed raw text.
func Artifact(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if body, ok := between(raw, BeginMark, EndMark); ok {
		return strings.TrimSpace(body) + "\n"
	}
	if m := fenceRE.FindStringSubmatch(raw); len(m) == 2 {
		return strings.TrimSpace(m[1]) + "\n"
	}
	// Drop common one-line preambles ("Sure, here is…") when followed by Markdown heading.
	if i := strings.Index(raw, "\n#"); i > 0 && i < 200 {
		rest := strings.TrimSpace(raw[i+1:])
		if strings.HasPrefix(rest, "#") {
			return rest + "\n"
		}
	}
	if !strings.HasSuffix(raw, "\n") {
		raw += "\n"
	}
	return raw
}

// ControlJSON unmarshals a small JSON object from raw or a ```json fence (control plane only).
func ControlJSON(raw string, dest any) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("empty control JSON")
	}
	if body, ok := between(raw, BeginMark, EndMark); ok {
		raw = strings.TrimSpace(body)
	}
	jsonFence := regexp.MustCompile("(?s)```json\\s*\\n(.*?)```")
	if m := jsonFence.FindStringSubmatch(raw); len(m) == 2 {
		raw = strings.TrimSpace(m[1])
	}
	// Find first { … last }
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return fmt.Errorf("no JSON object found")
	}
	raw = raw[start : end+1]
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return fmt.Errorf("control JSON: %w", err)
	}
	return nil
}

func between(s, begin, end string) (string, bool) {
	i := strings.Index(s, begin)
	if i < 0 {
		return "", false
	}
	i += len(begin)
	j := strings.Index(s[i:], end)
	if j < 0 {
		return "", false
	}
	return s[i : i+j], true
}

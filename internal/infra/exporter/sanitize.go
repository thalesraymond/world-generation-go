package exporter

import (
	"regexp"
	"strings"
)

var (
	illegalChars  = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	collapseSpace = regexp.MustCompile(`\s+`)
)

type nameTracker struct {
	used map[string]int
}

func newNameTracker() *nameTracker {
	return &nameTracker{used: make(map[string]int)}
}

func (t *nameTracker) sanitize(name string) string {
	cleaned := illegalChars.ReplaceAllString(name, "")
	cleaned = collapseSpace.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	if cleaned == "" {
		cleaned = "unnamed"
	}

	lower := strings.ToLower(cleaned)
	if count, exists := t.used[lower]; exists {
		t.used[lower] = count + 1
		return cleaned
	}
	t.used[lower] = 1
	return cleaned
}

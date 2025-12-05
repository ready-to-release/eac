package stream

import (
	"regexp"
	"strings"
	"sync"
)

// Filter determines what output lines to display.
type Filter struct {
	// Patterns to always show
	importantPatterns []*regexp.Regexp

	// Patterns to always hide
	noisePatterns []*regexp.Regexp

	// Recent duplicates tracking (for deduplication)
	recentLines map[string]int
	maxRecent   int
	mu          sync.Mutex
}

// NewFilter creates a filter with sensible defaults for Go build/test output.
func NewFilter() *Filter {
	return &Filter{
		importantPatterns: []*regexp.Regexp{
			regexp.MustCompile(`(?i)error`),
			regexp.MustCompile(`(?i)failed`),
			regexp.MustCompile(`(?i)warning`),
			regexp.MustCompile(`^panic:`),
			regexp.MustCompile(`^fatal:`),
			regexp.MustCompile(`undefined:`),
			regexp.MustCompile(`cannot find`),
			regexp.MustCompile(`not found`),
			regexp.MustCompile(`^FAIL\s`),
		},
		noisePatterns: []*regexp.Regexp{
			regexp.MustCompile(`^\s*$`),            // Empty lines
			regexp.MustCompile(`^go: downloading`), // Download messages
			regexp.MustCompile(`^=== RUN`),         // Test run markers
			regexp.MustCompile(`^=== PAUSE`),       // Test pause markers
			regexp.MustCompile(`^=== CONT`),        // Test continue markers
			regexp.MustCompile(`^--- PASS:`),       // Passing test markers
			regexp.MustCompile(`^PASS$`),           // Final pass marker
			regexp.MustCompile(`^ok\s+\S+\s+\d`),   // Package OK line
			regexp.MustCompile(`^coverage:`),       // Coverage line
			regexp.MustCompile(`^\?\s+\S+\s+\[no test files\]`), // No test files
		},
		recentLines: make(map[string]int),
		maxRecent:   100,
	}
}

// ShouldShow returns true if the line should be displayed.
func (f *Filter) ShouldShow(text string) bool {
	trimmed := strings.TrimSpace(text)

	// Empty lines are never shown
	if trimmed == "" {
		return false
	}

	// Always show important patterns (errors, warnings, etc.)
	for _, p := range f.importantPatterns {
		if p.MatchString(trimmed) {
			return true
		}
	}

	// Never show noise patterns
	for _, p := range f.noisePatterns {
		if p.MatchString(trimmed) {
			return false
		}
	}

	// Deduplicate recent lines
	f.mu.Lock()
	defer f.mu.Unlock()

	if count, ok := f.recentLines[trimmed]; ok && count > 2 {
		return false // Skip if seen more than twice recently
	}

	f.trackRecent(trimmed)
	return true
}

func (f *Filter) trackRecent(text string) {
	f.recentLines[text]++

	// Cleanup if too many entries
	if len(f.recentLines) > f.maxRecent {
		f.recentLines = make(map[string]int)
	}
}

// Reset clears the deduplication state.
func (f *Filter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.recentLines = make(map[string]int)
}

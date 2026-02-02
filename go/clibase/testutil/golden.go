// Package testutil provides shared test utilities for command testing.
package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// UpdateGolden is a flag that controls whether golden files should be updated.
// Run tests with -update-golden to update golden files.
var UpdateGolden = flag.Bool("update-golden", false, "update golden files")

// GoldenFile represents a golden file for comparison testing.
type GoldenFile struct {
	// Dir is the directory containing golden files (defaults to "testdata")
	Dir string
	// Name is the golden file name without extension
	Name string
	// Extension is the file extension (defaults to ".golden")
	Extension string
}

// NewGolden creates a new GoldenFile with sensible defaults.
func NewGolden(name string) *GoldenFile {
	return &GoldenFile{
		Dir:       "testdata",
		Name:      name,
		Extension: ".golden",
	}
}

// WithDir sets a custom directory for the golden file.
func (g *GoldenFile) WithDir(dir string) *GoldenFile {
	g.Dir = dir
	return g
}

// WithExtension sets a custom extension for the golden file.
func (g *GoldenFile) WithExtension(ext string) *GoldenFile {
	g.Extension = ext
	return g
}

// Path returns the full path to the golden file.
func (g *GoldenFile) Path() string {
	return filepath.Join(g.Dir, g.Name+g.Extension)
}

// Assert compares actual output against the golden file.
// If -update-golden flag is set, updates the golden file instead.
func (g *GoldenFile) Assert(t *testing.T, actual string) {
	t.Helper()

	path := g.Path()

	if *UpdateGolden {
		g.update(t, actual)
		return
	}

	expected, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("golden file %s does not exist, run with -update-golden to create it", path)
		}
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}

	expectedStr := string(expected)
	if actual != expectedStr {
		t.Errorf("output does not match golden file %s\n%s",
			path, formatDiff(expectedStr, actual))
	}
}

// AssertTrimmed compares actual output against the golden file after trimming whitespace.
func (g *GoldenFile) AssertTrimmed(t *testing.T, actual string) {
	t.Helper()
	g.Assert(t, strings.TrimSpace(actual))
}

// update writes the actual output to the golden file.
func (g *GoldenFile) update(t *testing.T, actual string) {
	t.Helper()

	path := g.Path()
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("failed to create golden file directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(actual), 0o644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}

	t.Logf("updated golden file: %s", path)
}

// Read reads the golden file content.
func (g *GoldenFile) Read(t *testing.T) string {
	t.Helper()

	path := g.Path()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return string(content)
}

// Exists returns true if the golden file exists.
func (g *GoldenFile) Exists() bool {
	_, err := os.Stat(g.Path())
	return err == nil
}

// formatDiff formats a simple diff between expected and actual strings.
func formatDiff(expected, actual string) string {
	var sb strings.Builder

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	sb.WriteString("\n--- expected ---\n")
	for i, line := range expectedLines {
		sb.WriteString(formatLine(i+1, line))
	}

	sb.WriteString("\n--- actual ---\n")
	for i, line := range actualLines {
		sb.WriteString(formatLine(i+1, line))
	}

	// Show line-by-line differences
	sb.WriteString("\n--- differences ---\n")
	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	diffCount := 0
	for i := 0; i < maxLines; i++ {
		var expLine, actLine string
		if i < len(expectedLines) {
			expLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actLine = actualLines[i]
		}

		if expLine != actLine {
			diffCount++
			if diffCount <= 10 { // Limit output
				sb.WriteString(formatLineDiff(i+1, expLine, actLine))
			}
		}
	}

	if diffCount > 10 {
		sb.WriteString("... and more differences\n")
	}

	sb.WriteString(strings.Repeat("-", 40) + "\n")
	return sb.String()
}

// formatLine formats a single line with line number.
func formatLine(num int, line string) string {
	return strings.ReplaceAll(line, "\t", "→") + "\n"
}

// formatLineDiff formats a line difference.
func formatLineDiff(num int, expected, actual string) string {
	var sb strings.Builder
	sb.WriteString("Line ")
	sb.WriteString(strings.Repeat(" ", 4-len(string(rune(num)))))
	sb.WriteString(string(rune('0' + num)))
	sb.WriteString(":\n")
	sb.WriteString("  - ")
	sb.WriteString(strings.ReplaceAll(expected, "\t", "→"))
	sb.WriteString("\n")
	sb.WriteString("  + ")
	sb.WriteString(strings.ReplaceAll(actual, "\t", "→"))
	sb.WriteString("\n")
	return sb.String()
}

// AssertGolden is a convenience function for simple golden file assertions.
func AssertGolden(t *testing.T, name, actual string) {
	t.Helper()
	NewGolden(name).Assert(t, actual)
}

// AssertGoldenTrimmed is a convenience function for trimmed golden file assertions.
func AssertGoldenTrimmed(t *testing.T, name, actual string) {
	t.Helper()
	NewGolden(name).AssertTrimmed(t, actual)
}

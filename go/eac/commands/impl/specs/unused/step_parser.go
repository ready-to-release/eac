package unused

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// StepDefinition represents a registered godog step.
type StepDefinition struct {
	Pattern      string         // raw regex pattern string
	File         string         // source file path
	Line         int            // line number
	Regex        *regexp.Regexp // compiled regex
	CompileErr   error          // error if regex failed to compile
	IsFromShared bool           // true if from internal/steps.go
}

// ParseStepDefinitions extracts all sc.Step() patterns from the given Go files.
func ParseStepDefinitions(files []string, isShared bool) ([]StepDefinition, error) {
	var allSteps []StepDefinition

	for _, file := range files {
		steps, err := parseStepFile(file, isShared)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		allSteps = append(allSteps, steps...)
	}

	return allSteps, nil
}

// parseStepFile parses a single Go file for step definitions.
func parseStepFile(filePath string, isShared bool) ([]StepDefinition, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var steps []StepDefinition
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Pattern to match sc.Step(`pattern`, ...) or sc.Step("pattern", ...)
	// Handles both backtick and double-quoted strings
	stepPattern := regexp.MustCompile(`\.Step\s*\(\s*` + "`" + `([^` + "`" + `]+)` + "`")
	stepPatternDoubleQuote := regexp.MustCompile(`\.Step\s*\(\s*"([^"\\]*(?:\\.[^"\\]*)*)"`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip comments
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Try backtick pattern first (more common for regex)
		if matches := stepPattern.FindStringSubmatch(line); len(matches) >= 2 {
			pattern := matches[1]
			step := createStepDefinition(pattern, filePath, lineNum, isShared)
			steps = append(steps, step)
			continue
		}

		// Try double-quote pattern
		if matches := stepPatternDoubleQuote.FindStringSubmatch(line); len(matches) >= 2 {
			pattern := unescapeDoubleQuoted(matches[1])
			step := createStepDefinition(pattern, filePath, lineNum, isShared)
			steps = append(steps, step)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

// createStepDefinition creates a StepDefinition and attempts to compile the regex.
func createStepDefinition(pattern, file string, line int, isShared bool) StepDefinition {
	step := StepDefinition{
		Pattern:      pattern,
		File:         file,
		Line:         line,
		IsFromShared: isShared,
	}

	// Try to compile the regex
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		step.CompileErr = err
	} else {
		step.Regex = compiled
	}

	return step
}

// unescapeDoubleQuoted handles escape sequences in double-quoted strings.
func unescapeDoubleQuoted(s string) string {
	// Handle common escape sequences
	s = strings.ReplaceAll(s, `\\`, `\`)
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}

// RelativePath returns the file path relative to the given base.
func (s StepDefinition) RelativePath(base string) string {
	rel, err := RelPath(base, s.File)
	if err != nil {
		return s.File
	}
	return rel
}

// RelPath is a helper to get relative path, handling errors.
func RelPath(base, target string) (string, error) {
	// Normalize paths for Windows/Unix compatibility
	base = strings.ReplaceAll(base, "\\", "/")
	target = strings.ReplaceAll(target, "\\", "/")

	if strings.HasPrefix(target, base) {
		rel := strings.TrimPrefix(target, base)
		rel = strings.TrimPrefix(rel, "/")
		return rel, nil
	}

	// Fallback: just return target as-is
	return target, nil
}

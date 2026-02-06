// Package gowork provides utilities for Go workspace module discovery.
package gowork

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParseModules reads go.work at repoRoot and returns absolute paths to all
// workspace module directories.
func ParseModules(repoRoot string) ([]string, error) {
	goWorkPath := filepath.Join(repoRoot, "go.work")
	f, err := os.Open(goWorkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.work: %w", err)
	}
	defer f.Close()

	var modules []string
	inUseBlock := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "use (" {
			inUseBlock = true
			continue
		}
		if inUseBlock && line == ")" {
			inUseBlock = false
			continue
		}
		if inUseBlock && line != "" {
			// Strip leading "./"
			modPath := strings.TrimPrefix(line, "./")
			modules = append(modules, filepath.Join(repoRoot, modPath))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read go.work: %w", err)
	}

	return modules, nil
}

// IndentLines prefixes each non-empty line in text with prefix.
func IndentLines(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

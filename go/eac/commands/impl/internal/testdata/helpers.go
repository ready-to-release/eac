// Package testdata provides shared test data preparation functions
// used by get-tests, show-tests, get-suite, and show-suite commands.
package testdata

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindRepoRoot finds the repository root by looking for .git directory
func FindRepoRoot(startPath string) (string, error) {
	current := startPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("not in a git repository")
		}
		current = parent
	}
}

// NormalizePathSeparators converts backslashes to forward slashes
func NormalizePathSeparators(path string) string {
	result := make([]byte, len(path))
	for i := 0; i < len(path); i++ {
		if path[i] == '\\' {
			result[i] = '/'
		} else {
			result[i] = path[i]
		}
	}
	return string(result)
}

// SplitPath splits a path by forward slashes
func SplitPath(path string) []string {
	parts := []string{}
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

// FilterTagsByPrefix returns tags that start with the given prefix
func FilterTagsByPrefix(tags []string, prefix string) []string {
	result := []string{}
	for _, tag := range tags {
		if len(tag) >= len(prefix) && tag[:len(prefix)] == prefix {
			result = append(result, tag)
		}
	}
	return result
}

// FilterTagsByPatterns returns tags that exactly match any of the given patterns
func FilterTagsByPatterns(tags []string, patterns []string) []string {
	result := []string{}
	for _, tag := range tags {
		for _, pattern := range patterns {
			if tag == pattern {
				result = append(result, tag)
				break
			}
		}
	}
	return result
}

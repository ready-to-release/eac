// Package testjson provides utilities for parsing Go test JSON output
package testjson

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
)

// GoTestEvent represents a single event from `go test -json` output.
type GoTestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`  // "run", "output", "pass", "fail", "skip"
	Package string  `json:"Package"` // Package name
	Test    string  `json:"Test"`    // Test name (empty for package-level)
	Output  string  `json:"Output"`  // Test output line
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// ParseJSONFile parses a Go test JSON file and returns all events.
func ParseJSONFile(jsonPath string) ([]GoTestEvent, error) {
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var events []GoTestEvent
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		var event GoTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue // Skip malformed lines
		}
		events = append(events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

// CountTestResults counts passed, failed, and skipped tests from events.
func CountTestResults(events []GoTestEvent) (passed, failed, skipped int) {
	for _, event := range events {
		// Only count test-level events (not package-level)
		if event.Test != "" {
			switch event.Action {
			case "pass":
				passed++
			case "fail":
				failed++
			case "skip":
				skipped++
			}
		}
	}
	return passed, failed, skipped
}

// ExtractFailedTests returns a map of failed tests with their output.
// Parent tests are excluded when their subtests are already included,
// since the parent failure is just a consequence of the subtest failure.
func ExtractFailedTests(events []GoTestEvent) map[string][]string {
	testOutputs := make(map[string][]string)
	failedTests := make(map[string]bool)

	for _, event := range events {
		if event.Test == "" {
			continue
		}

		key := event.Package + "::" + event.Test

		// Mark test as failed
		if event.Action == "fail" {
			failedTests[key] = true
		}

		// Collect output for all tests
		if event.Action == "output" {
			if trimmed := strings.TrimSpace(event.Output); trimmed != "" {
				// Skip ANSI color codes and test runner output
				if !strings.HasPrefix(trimmed, "===") && !strings.HasPrefix(trimmed, "---") {
					testOutputs[key] = append(testOutputs[key], trimmed)
				}
			}
		}
	}

	// Filter out parent tests when subtests exist
	// e.g., if "TestFoo/subtest" failed, don't also report "TestFoo"
	for key := range failedTests {
		for otherKey := range failedTests {
			if otherKey != key && strings.HasPrefix(otherKey, key+"/") {
				// key is a parent of otherKey, remove the parent
				delete(failedTests, key)
				break
			}
		}
	}

	// Return only outputs for failed tests
	result := make(map[string][]string)
	for key := range failedTests {
		if outputs, exists := testOutputs[key]; exists {
			result[key] = outputs
		}
	}

	return result
}

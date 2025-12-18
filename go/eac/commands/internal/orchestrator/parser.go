package orchestrator

import (
	"encoding/json"
	"os"
	"strings"
)

// goTestEvent represents a single event from go test -json output
type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// maxTailLines is the number of test output lines to show
const maxTailLines = 5

// parseLogForIssues scans a log file for warnings and errors
// Handles both plain text logs and Go test JSON output
// Returns actual test assertion output, filtering out framework summary lines
func parseLogForIssues(logPath string) (warnings []string, errors []string) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, nil
	}

	// Extract all content lines from JSON or plain text
	rawLines := strings.Split(string(data), "\n")
	var testOutputLines []string
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as JSON (go test -json output)
		var content string
		if strings.HasPrefix(line, "{") {
			var event goTestEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				content = strings.TrimSpace(event.Output)
				if content == "" {
					continue
				}
			} else {
				content = line
			}
		} else {
			content = line
		}

		// Only include actual test output, not framework lines
		if isTestOutput(content) {
			testOutputLines = append(testOutputLines, content)
		}
	}

	// Return the last N test output lines
	if len(testOutputLines) > maxTailLines {
		errors = testOutputLines[len(testOutputLines)-maxTailLines:]
	} else {
		errors = testOutputLines
	}

	return warnings, errors
}

// isTestOutput returns true if the line is actual test output (assertions, errors)
// rather than framework summary lines
func isTestOutput(content string) bool {
	// Skip empty or whitespace-only
	if strings.TrimSpace(content) == "" {
		return false
	}

	// Skip test run markers
	if strings.HasPrefix(content, "=== RUN") ||
		strings.HasPrefix(content, "--- PASS") ||
		strings.HasPrefix(content, "--- FAIL") ||
		strings.HasPrefix(content, "PASS") ||
		strings.HasPrefix(content, "FAIL") ||
		strings.HasPrefix(content, "ok ") ||
		strings.HasPrefix(content, "?") {
		return false
	}

	// Skip godog/cucumber summary lines
	if strings.Contains(content, " scenarios (") ||
		strings.Contains(content, " steps (") ||
		strings.HasPrefix(content, "Feature:") ||
		strings.HasPrefix(content, "Scenario:") {
		return false
	}

	// Skip godog step definition references (Given/When/Then ... # path)
	if (strings.HasPrefix(content, "Given ") ||
		strings.HasPrefix(content, "When ") ||
		strings.HasPrefix(content, "Then ") ||
		strings.HasPrefix(content, "And ")) &&
		strings.Contains(content, " # ") {
		return false
	}

	// Skip feature file line references (e.g., "feature:17", "specification.feature:17")
	if strings.Contains(content, "feature:") && !strings.Contains(content, "Error") {
		return false
	}

	// Skip test function name lines (e.g., "TestRepositoryFeatures/No_files...")
	if strings.HasPrefix(content, "Test") && strings.Contains(content, "/") {
		return false
	}

	// Skip lines ending with test timing (e.g., "(0.08s)")
	if strings.HasSuffix(content, "s)") && strings.Contains(content, "(0.") {
		return false
	}

	// Skip pure timing lines (e.g., "86.741ms", "1.234s")
	if isTimingLine(content) {
		return false
	}

	// Skip pure number lines (e.g., " 5" from godog summary counts)
	if isPureNumber(content) {
		return false
	}

	// Skip "Failed steps:" header
	if content == "Failed steps:" {
		return false
	}

	// Include lines that look like test assertions (file.go:123: message)
	// Include lines with "Error:" or "error:"
	// Include indented continuation lines (start with spaces or tabs)
	// Include everything else that's not filtered above
	return true
}

// isTimingLine checks if a line is just a timing value
func isTimingLine(content string) bool {
	content = strings.TrimSpace(content)
	// Match patterns like "86.741ms", "1.234s", "2m30s"
	if len(content) < 20 && !strings.Contains(content, " ") {
		if strings.HasSuffix(content, "ms") ||
			strings.HasSuffix(content, "s") ||
			strings.HasSuffix(content, "m") {
			// Check if rest is numeric/timing
			for _, r := range content[:len(content)-1] {
				if r != '.' && r != 'm' && r != 's' && (r < '0' || r > '9') {
					return false
				}
			}
			return true
		}
	}
	return false
}

// isPureNumber checks if a line is just a number (e.g., step counts from godog)
func isPureNumber(content string) bool {
	content = strings.TrimSpace(content)
	if content == "" {
		return false
	}
	for _, r := range content {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

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

// parseLogForIssues scans a log file for warnings and errors
// Handles both plain text logs and Go test JSON output
func parseLogForIssues(logPath string) (warnings []string, errors []string) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse as JSON (go test -json output)
		var content string
		if strings.HasPrefix(line, "{") {
			var event goTestEvent
			if err := json.Unmarshal([]byte(line), &event); err == nil {
				// Use the Output field from JSON
				content = strings.TrimSpace(event.Output)
				if content == "" {
					continue
				}
			} else {
				// Not valid JSON, use the line as-is
				content = line
			}
		} else {
			content = line
		}

		lowerContent := strings.ToLower(content)

		// Check for errors (but skip test names that contain "error" or "failed")
		isTestOutput := strings.HasPrefix(content, "=== RUN") ||
			strings.HasPrefix(content, "--- PASS") ||
			strings.HasPrefix(content, "--- FAIL")

		if !isTestOutput {
			if strings.Contains(lowerContent, "error:") ||
				strings.Contains(content, "❌") ||
				strings.Contains(lowerContent, "fatal:") {
				errors = append(errors, content)
			}
		}

		// Check for warnings (but not if also an error)
		// Look for WARNING (case-insensitive) but not in test names
		if !isTestOutput &&
			(strings.Contains(lowerContent, "warning:") || strings.Contains(content, "WARNING:")) &&
			!strings.Contains(lowerContent, "error:") {
			warnings = append(warnings, content)
		}
	}

	return warnings, errors
}

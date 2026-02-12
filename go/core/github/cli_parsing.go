package github

import (
	"encoding/json"
	"fmt"
	"strings"
)

// parseFlagValue extracts the value of a named flag from a CLI argument list.
// Returns the value and true if found, or empty string and false if not found.
// Example: parseFlagValue([]string{"-w", "ci.yml", "--json", "..."}, "-w") returns ("ci.yml", true)
func parseFlagValue(args []string, flag string) (string, bool) {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1], true
		}
	}
	return "", false
}

// parseWorkflowRuns unmarshals JSON data into a slice of WorkflowRun.
func parseWorkflowRuns(data []byte) ([]WorkflowRun, error) {
	var runs []WorkflowRun
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, fmt.Errorf("failed to parse runs: %w", err)
	}
	return runs, nil
}

// parseReleases unmarshals JSON data into a slice of Release.
func parseReleases(data []byte) ([]Release, error) {
	var releases []Release
	if err := json.Unmarshal(data, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse releases: %w", err)
	}
	return releases, nil
}

// filterNonEmptyLines splits text by newlines and returns non-empty lines.
func filterNonEmptyLines(text string) []string {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

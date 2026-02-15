package testing

import (
	"encoding/json"
	"os"
)

// manualTestFile is the structure of a manual-tests.json file.
type manualTestFile struct {
	Tests []manualTestEntry `json:"tests"`
}

type manualTestEntry struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	DurationMs int64  `json:"duration_ms,omitempty"`
	Error      string `json:"error,omitempty"`
}

// parseManualTestFile parses a manual-tests.json file and returns test entries.
func parseManualTestFile(path, module string) []TestEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var file manualTestFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil
	}

	var entries []TestEntry
	for _, test := range file.Tests {
		entries = append(entries, TestEntry{
			Name:       test.Name,
			Module:     module,
			Package:    "manual",
			Type:       "manual",
			Suite:      "manual",
			Status:     test.Status,
			DurationMs: test.DurationMs,
			Error:      test.Error,
		})
	}

	return entries
}

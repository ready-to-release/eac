package testview

import (
	"encoding/json"
	"os"
)

// ctrfReport is a minimal struct for parsing unit.json (CTRF) files.
type ctrfReport struct {
	Results struct {
		Tests []ctrfTest `json:"tests"`
	} `json:"results"`
}

type ctrfTest struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Duration int64  `json:"duration"` // milliseconds
	Suite    string `json:"suite,omitempty"`
	FilePath string `json:"filePath,omitempty"`
}

// parseCTRFFile parses a unit.json (CTRF) file and returns test entries.
func parseCTRFFile(path, module string) []TestEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var report ctrfReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil
	}

	var entries []TestEntry
	for _, test := range report.Results.Tests {
		status := normalizeCTRFStatus(test.Status)
		suite := test.Suite
		if suite == "" {
			suite = "unit"
		}

		entries = append(entries, TestEntry{
			Name:       test.Name,
			Module:     module,
			Package:    test.FilePath,
			Type:       "gotest",
			Suite:      suite,
			Status:     status,
			DurationMs: test.Duration,
			FilePath:   test.FilePath,
		})
	}

	return entries
}

// normalizeCTRFStatus maps CTRF status values to our constants.
func normalizeCTRFStatus(status string) string {
	switch status {
	case "passed":
		return StatusPassed
	case "failed":
		return StatusFailed
	case "skipped", "pending":
		return StatusSkipped
	default:
		return status
	}
}

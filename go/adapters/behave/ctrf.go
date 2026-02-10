package behave

import (
	"encoding/json"
	"fmt"

	"github.com/ready-to-release/eac/go/core/ctrf"
)

// behaveFeature represents a feature in behave's JSON output.
type behaveFeature struct {
	Name     string           `json:"name"`
	Keyword  string           `json:"keyword"`
	Status   string           `json:"status"`
	Elements []behaveScenario `json:"elements"`
}

// behaveScenario represents a scenario in behave's JSON output.
type behaveScenario struct {
	Name    string       `json:"name"`
	Keyword string       `json:"keyword"`
	Status  string       `json:"status"`
	Steps   []behaveStep `json:"steps"`
	Type    string       `json:"type"`
}

// behaveStep represents a step in behave's JSON output.
type behaveStep struct {
	Name     string        `json:"name"`
	Keyword  string        `json:"keyword"`
	Status   string        `json:"status"`
	Duration float64       `json:"duration"`
	Result   behaveResult  `json:"result"`
}

// behaveResult represents a step result.
type behaveResult struct {
	Status       string  `json:"status"`
	Duration     float64 `json:"duration"`
	ErrorMessage string  `json:"error_message"`
}

// durationMs converts seconds to milliseconds, ensuring minimum 1ms for non-zero.
func durationMs(seconds float64) int64 {
	if seconds <= 0 {
		return 0
	}
	ms := int64(seconds * 1000)
	if ms == 0 {
		return 1
	}
	return ms
}

// convertBehaveJSONToCTRF converts behave JSON output to CTRF format.
func convertBehaveJSONToCTRF(jsonData []byte) *ctrf.Report {
	var features []behaveFeature
	if err := json.Unmarshal(jsonData, &features); err != nil {
		return nil
	}

	report := ctrf.NewReport("behave")

	for _, feature := range features {
		for _, scenario := range feature.Elements {
			if scenario.Type == "background" {
				continue // Skip background sections
			}

			test := ctrf.Test{
				Name:  fmt.Sprintf("%s: %s", feature.Name, scenario.Name),
				Suite: feature.Name,
			}

			// Calculate total duration from steps
			var totalDuration float64
			for _, step := range scenario.Steps {
				totalDuration += step.Duration
			}
			test.Duration = durationMs(totalDuration)

			switch scenario.Status {
			case "passed":
				test.Status = ctrf.StatusPassed
			case "failed":
				test.Status = ctrf.StatusFailed
				// Find the first failed step for error message
				for _, step := range scenario.Steps {
					if step.Status == "failed" || step.Result.Status == "failed" {
						test.Message = step.Result.ErrorMessage
						break
					}
				}
			case "skipped":
				test.Status = ctrf.StatusSkipped
			case "undefined":
				test.Status = ctrf.StatusPending
			default:
				test.Status = ctrf.StatusOther
			}

			report.AddTest(test)
		}
	}

	report.Finalize()
	return report
}

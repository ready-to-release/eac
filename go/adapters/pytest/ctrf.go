package pytest

import (
	"encoding/json"

	"github.com/ready-to-release/eac/go/core/ctrf"
)

// pytestJSONReport represents pytest-json-report output format.
type pytestJSONReport struct {
	Summary pytestSummary `json:"summary"`
	Tests   []pytestTest  `json:"tests"`
}

type pytestSummary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
	Total   int `json:"total"`
}

type pytestTest struct {
	NodeID   string        `json:"nodeid"`
	Outcome  string        `json:"outcome"`
	Duration float64       `json:"duration"`
	Setup    pytestPhase   `json:"setup"`
	Call     pytestPhase   `json:"call"`
	Teardown pytestPhase   `json:"teardown"`
	Keywords []string      `json:"keywords"`
}

type pytestPhase struct {
	Duration float64       `json:"duration"`
	Outcome  string        `json:"outcome"`
	Crash    *pytestCrash  `json:"crash"`
	Longrepr string        `json:"longrepr"`
}

type pytestCrash struct {
	Path    string `json:"path"`
	Lineno  int    `json:"lineno"`
	Message string `json:"message"`
}

// durationMs converts seconds to milliseconds, ensuring minimum 1ms for non-zero durations.
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

// convertPytestJSONToCTRF converts pytest-json-report output to CTRF format.
func convertPytestJSONToCTRF(jsonData []byte) *ctrf.Report {
	var report pytestJSONReport
	if err := json.Unmarshal(jsonData, &report); err != nil {
		return nil
	}

	ctrfReport := ctrf.NewReport("pytest")

	for _, t := range report.Tests {
		test := ctrf.Test{
			Name:     t.NodeID,
			Duration: durationMs(t.Duration),
		}

		switch t.Outcome {
		case "passed":
			test.Status = ctrf.StatusPassed
		case "failed":
			test.Status = ctrf.StatusFailed
			if t.Call.Crash != nil {
				test.Message = t.Call.Crash.Message
			}
			if t.Call.Longrepr != "" {
				test.Trace = t.Call.Longrepr
			}
		case "skipped":
			test.Status = ctrf.StatusSkipped
		default:
			test.Status = ctrf.StatusOther
		}

		ctrfReport.AddTest(test)
	}

	ctrfReport.Finalize()
	return ctrfReport
}

package gotest

import (
	"strings"
	"time"

	"github.com/ready-to-release/eac/go/clibase/ctrf"
	"github.com/ready-to-release/eac/go/clibase/testrunners"
)

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

// convertGoTestEventsToCTRF converts go test -json events to CTRF format.
func convertGoTestEventsToCTRF(events []testrunners.TestEvent) *ctrf.Report {
	report := ctrf.NewReport("go-test")

	type testState struct {
		output  []string
		elapsed float64
	}
	tests := make(map[string]*testState)

	var startTime, stopTime time.Time
	for _, event := range events {
		if event.Time != "" {
			if t, err := time.Parse(time.RFC3339Nano, event.Time); err == nil {
				if startTime.IsZero() || t.Before(startTime) {
					startTime = t
				}
				if t.After(stopTime) {
					stopTime = t
				}
			}
		}

		if event.Test == "" {
			continue
		}

		if tests[event.Test] == nil {
			tests[event.Test] = &testState{}
		}

		switch event.Action {
		case "output":
			tests[event.Test].output = append(tests[event.Test].output, event.Output)
		case "pass":
			tests[event.Test].elapsed = event.Elapsed
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusPassed,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
			})
		case "fail":
			tests[event.Test].elapsed = event.Elapsed
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusFailed,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
				Trace:    strings.Join(tests[event.Test].output, ""),
			})
		case "skip":
			report.AddTest(ctrf.Test{
				Name:     event.Test,
				Status:   ctrf.StatusSkipped,
				Duration: durationMs(event.Elapsed),
				Suite:    event.Package,
			})
		}
	}

	if !startTime.IsZero() && !stopTime.IsZero() {
		report.SetTimes(startTime.UnixMilli(), stopTime.UnixMilli())
	} else {
		report.Finalize()
	}
	return report
}

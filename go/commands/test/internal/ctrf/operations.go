package ctrf

import (
	"encoding/json"
	"os"
	"time"
)

// NewReport creates a new CTRF report with current time as start.
func NewReport(toolName string) *Report {
	return &Report{
		Results: Results{
			Tool:  Tool{Name: toolName},
			Tests: make([]Test, 0),
			Summary: Summary{
				Start: time.Now().UnixMilli(),
			},
		},
	}
}

// NewEmptyReport creates a new CTRF report for aggregation (no preset times).
func NewEmptyReport(toolName string) *Report {
	return &Report{
		Results: Results{
			Tool:  Tool{Name: toolName},
			Tests: make([]Test, 0),
			Summary: Summary{
				Start: 1<<62 - 1, // Max value so first merge sets it
				Stop:  0,         // Min value so first merge sets it
			},
		},
	}
}

// SetTimes sets the start and stop times explicitly.
func (r *Report) SetTimes(start, stop int64) {
	r.Results.Summary.Start = start
	r.Results.Summary.Stop = stop
}

// AddTest adds a test result to the report.
func (r *Report) AddTest(test Test) {
	r.Results.Tests = append(r.Results.Tests, test)

	// Update summary counts
	switch test.Status {
	case StatusPassed:
		r.Results.Summary.Passed++
	case StatusFailed:
		r.Results.Summary.Failed++
	case StatusSkipped:
		r.Results.Summary.Skipped++
	case StatusPending:
		r.Results.Summary.Pending++
	default:
		r.Results.Summary.Other++
	}
	r.Results.Summary.Tests++
}

// Finalize sets the stop time.
func (r *Report) Finalize() {
	r.Results.Summary.Stop = time.Now().UnixMilli()
}

// ToJSON serializes the report to JSON.
func (r *Report) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// Merge combines another report into this one.
func (r *Report) Merge(other *Report) {
	for _, test := range other.Results.Tests {
		r.AddTest(test)
	}
	// Update start time to earliest
	if other.Results.Summary.Start < r.Results.Summary.Start {
		r.Results.Summary.Start = other.Results.Summary.Start
	}
	// Update stop time to latest
	if other.Results.Summary.Stop > r.Results.Summary.Stop {
		r.Results.Summary.Stop = other.Results.Summary.Stop
	}
}

// ParseFile reads and parses a CTRF JSON file.
func ParseFile(path string) (*Report, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Parse parses CTRF JSON data.
func Parse(data []byte) (*Report, error) {
	var report Report
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

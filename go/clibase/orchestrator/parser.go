package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LogEvent represents a structured log event (compatible with go test -json).
type LogEvent struct {
	Time    string  `json:"Time,omitempty"`
	Action  string  `json:"Action"`
	Package string  `json:"Package,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// testState tracks output for a single test.
type testState struct {
	output []string
	failed bool
}

// maxTailLines is the number of output lines to show for failures.
const maxTailLines = 10

// parseLogForIssues extracts error output from failed tests.
// Checks structured output files first (cucumber.json, unit.json),
// then falls back to parsing the log file for JSON events.
func parseLogForIssues(logPath string) (warnings, errors []string) {
	dir := filepath.Dir(logPath)

	// Try structured output files first - these are authoritative
	// 1. cucumber.json (godog BDD tests)
	cucumberPath := filepath.Join(dir, "cucumber.json")
	if warnings, errors = parseCucumberJSON(cucumberPath); len(errors) > 0 || len(warnings) > 0 {
		return warnings, errors
	}

	// 2. unit.json (CTRF format from gotest/mocha)
	unitPath := filepath.Join(dir, "unit.json")
	if warnings, errors = parseCTRFJSON(unitPath); len(errors) > 0 || len(warnings) > 0 {
		return warnings, errors
	}

	// Fallback: parse log file for JSON events (legacy support)
	file, err := os.Open(logPath)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	return parseJSONLog(file)
}

// parseJSONLog parses JSON log output using Action field as source of truth.
func parseJSONLog(r io.Reader) (warnings, errors []string) {
	tests := make(map[string]*testState)
	var packageOutput []string
	packageFailed := false

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip non-JSON lines (go: downloading, etc.)
		if !strings.HasPrefix(line, "{") {
			continue
		}

		var event LogEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		key := event.Package
		if event.Test != "" {
			key = event.Package + "/" + event.Test
		}

		switch event.Action {
		case "run":
			tests[key] = &testState{}

		case "output":
			output := strings.TrimRight(event.Output, "\n")
			if output == "" {
				continue
			}
			if event.Test != "" {
				if t, ok := tests[key]; ok {
					t.output = append(t.output, output)
				}
			} else {
				packageOutput = append(packageOutput, output)
			}

		case "fail":
			if event.Test != "" {
				if t, ok := tests[key]; ok {
					t.failed = true
				}
			} else {
				packageFailed = true
			}

		case "error":
			// Direct error event (for non-test tools using JSONLogWriter)
			if event.Output != "" {
				errors = append(errors, strings.TrimRight(event.Output, "\n"))
			}

		case "warning":
			// Direct warning event
			if event.Output != "" {
				warnings = append(warnings, strings.TrimRight(event.Output, "\n"))
			}
		}
	}

	// Collect output only from failed tests
	for _, t := range tests {
		if t.failed {
			errors = append(errors, t.output...)
		}
	}

	// If package failed but no specific test failures, use package output
	if packageFailed && len(errors) == 0 {
		errors = packageOutput
	}

	// Limit output size
	if len(errors) > maxTailLines {
		errors = errors[len(errors)-maxTailLines:]
	}

	return warnings, errors
}

// cucumberReport represents the structure of cucumber.json output.
type cucumberReport []cucumberFeature

// cucumberFeature represents a feature in cucumber.json.
type cucumberFeature struct {
	Elements []cucumberScenario `json:"elements"`
}

// cucumberScenario represents a scenario in cucumber.json.
type cucumberScenario struct {
	Name  string           `json:"name"`
	Steps []cucumberStep   `json:"steps"`
}

// cucumberStep represents a step in cucumber.json.
type cucumberStep struct {
	Result cucumberStepResult `json:"result"`
}

// cucumberStepResult represents a step result in cucumber.json.
type cucumberStepResult struct {
	Status string `json:"status"`
	Error  string `json:"error_message,omitempty"`
}

// ctrfReport represents the structure of unit.json (CTRF format).
type ctrfReport struct {
	Results ctrfResults `json:"results"`
}

type ctrfResults struct {
	Tests []ctrfTest `json:"tests"`
}

type ctrfTest struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Trace   string `json:"trace,omitempty"`
}

// parseCTRFJSON extracts errors from a unit.json file (CTRF format).
// Returns test names with their error messages/traces for failed tests.
func parseCTRFJSON(ctrfPath string) (warnings, errors []string) {
	data, err := os.ReadFile(ctrfPath)
	if err != nil {
		return nil, nil // File doesn't exist or can't be read
	}

	var report ctrfReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, nil // Invalid JSON
	}

	for _, test := range report.Results.Tests {
		if test.Status == "failed" {
			// Prefer Trace (full error output), fall back to Message
			errDetail := test.Trace
			if errDetail == "" {
				errDetail = test.Message
			}
			if errDetail != "" {
				errors = append(errors, fmt.Sprintf("%s: %s", test.Name, errDetail))
			} else {
				errors = append(errors, fmt.Sprintf("%s: failed", test.Name))
			}
		}
	}

	// Limit output size
	if len(errors) > maxTailLines {
		errors = errors[len(errors)-maxTailLines:]
	}

	return warnings, errors
}

// parseCucumberJSON extracts errors from a cucumber.json file.
// Returns scenario names with their error messages for all failed steps.
func parseCucumberJSON(cucumberPath string) (warnings, errors []string) {
	data, err := os.ReadFile(cucumberPath)
	if err != nil {
		return nil, nil // File doesn't exist or can't be read
	}

	var report cucumberReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, nil // Invalid JSON
	}

	for i := range report {
		feature := &report[i]
		for j := range feature.Elements {
			scenario := &feature.Elements[j]
			// Collect all failed steps
			for _, step := range scenario.Steps {
				if step.Result.Status == "failed" && step.Result.Error != "" {
					errors = append(errors, fmt.Sprintf("%s: %s", scenario.Name, step.Result.Error))
				}
			}
		}
	}

	// Limit output size
	if len(errors) > maxTailLines {
		errors = errors[len(errors)-maxTailLines:]
	}

	return warnings, errors
}

// JSONLogWriter wraps an io.Writer to output JSON log events.
// Use this for non-test workers to produce parseable output.
type JSONLogWriter struct {
	w   io.Writer
	pkg string
	enc *json.Encoder
}

// NewJSONLogWriter creates a writer that outputs JSON log events.
func NewJSONLogWriter(w io.Writer, pkg string) *JSONLogWriter {
	return &JSONLogWriter{
		w:   w,
		pkg: pkg,
		enc: json.NewEncoder(w),
	}
}

// Write implements io.Writer, treating each write as an output event.
func (j *JSONLogWriter) Write(p []byte) (n int, err error) {
	event := LogEvent{
		Action:  "output",
		Package: j.pkg,
		Output:  string(p),
	}
	if err := j.enc.Encode(event); err != nil {
		return 0, err
	}
	return len(p), nil
}

// Info writes an info-level output event.
func (j *JSONLogWriter) Info(msg string) {
	_ = j.enc.Encode(LogEvent{Action: "output", Package: j.pkg, Output: msg + "\n"}) //nolint:errcheck // best-effort logging
}

// Error writes an error event.
func (j *JSONLogWriter) Error(msg string) {
	_ = j.enc.Encode(LogEvent{Action: "error", Package: j.pkg, Output: msg + "\n"}) //nolint:errcheck // best-effort logging
}

// Warning writes a warning event.
func (j *JSONLogWriter) Warning(msg string) {
	_ = j.enc.Encode(LogEvent{Action: "warning", Package: j.pkg, Output: msg + "\n"}) //nolint:errcheck // best-effort logging
}

// Fail marks the package as failed.
func (j *JSONLogWriter) Fail() {
	_ = j.enc.Encode(LogEvent{Action: "fail", Package: j.pkg}) //nolint:errcheck // best-effort logging
}

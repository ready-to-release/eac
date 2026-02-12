package testrunners

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewStreamingRunner(t *testing.T) {
	t.Run("creates runner with both writers", func(t *testing.T) {
		tui := &bytes.Buffer{}
		log := &bytes.Buffer{}
		r := NewStreamingRunner(tui, log)

		if r == nil {
			t.Fatal("expected non-nil runner")
		}
		if r.tuiWriter != tui {
			t.Error("tuiWriter not set correctly")
		}
		if r.logWriter != log {
			t.Error("logWriter not set correctly")
		}
		if len(r.events) != 0 {
			t.Errorf("events should be empty, got %d", len(r.events))
		}
	})

	t.Run("creates runner with nil writers", func(t *testing.T) {
		r := NewStreamingRunner(nil, nil)

		if r == nil {
			t.Fatal("expected non-nil runner")
		}
		if r.tuiWriter != nil {
			t.Error("tuiWriter should be nil")
		}
		if r.logWriter != nil {
			t.Error("logWriter should be nil")
		}
	})
}

func TestProcessJSONOutput(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantEvents     int
		wantTUIOutput  string
		wantTUIAbsent  []string // substrings that must NOT appear in TUI output
		wantTUIContain []string // substrings that must appear in TUI output
	}{
		{
			name:       "valid pass event",
			input:      `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"pkg/foo","Test":"TestOne","Elapsed":0.5}`,
			wantEvents: 1,
		},
		{
			name:       "valid fail event",
			input:      `{"Time":"2024-01-01T00:00:00Z","Action":"fail","Package":"pkg/foo","Test":"TestTwo","Elapsed":1.2}`,
			wantEvents: 1,
		},
		{
			name:       "valid skip event",
			input:      `{"Time":"2024-01-01T00:00:00Z","Action":"skip","Package":"pkg/foo","Test":"TestThree"}`,
			wantEvents: 1,
		},
		{
			name:           "output event writes to TUI",
			input:          `{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestOne","Output":"--- PASS: TestOne (0.00s)\n"}`,
			wantEvents:     1,
			wantTUIContain: []string{"--- PASS: TestOne"},
		},
		{
			name:          "output event with empty output is skipped",
			input:         `{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestOne","Output":"   \n"}`,
			wantEvents:    1,
			wantTUIAbsent: []string{"PASS", "FAIL"},
		},
		{
			name:          "output event with empty string is skipped",
			input:         `{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestOne","Output":""}`,
			wantEvents:    1,
			wantTUIOutput: "",
		},
		{
			name:           "non-output action does not write to TUI",
			input:          `{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestOne"}`,
			wantEvents:     1,
			wantTUIAbsent:  []string{"TestOne"},
			wantTUIContain: nil,
		},
		{
			name: "multiple events parsed in order",
			input: strings.Join([]string{
				`{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestA"}`,
				`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Test":"TestA","Output":"hello\n"}`,
				`{"Time":"2024-01-01T00:00:02Z","Action":"pass","Package":"pkg/foo","Test":"TestA","Elapsed":0.1}`,
			}, "\n"),
			wantEvents:     3,
			wantTUIContain: []string{"hello"},
		},
		{
			name:           "malformed JSON written raw to TUI",
			input:          `this is not json at all`,
			wantEvents:     0,
			wantTUIContain: []string{"this is not json at all"},
		},
		{
			name:           "partial JSON written raw to TUI",
			input:          `{"Time":"2024-01-01T00:00:00Z","Action":`,
			wantEvents:     0,
			wantTUIContain: []string{`{"Time":"2024-01-01T00:00:00Z","Action":`},
		},
		{
			name:       "empty input produces no events",
			input:      "",
			wantEvents: 0,
		},
		{
			name: "mixed valid and invalid lines",
			input: strings.Join([]string{
				`{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"pkg/foo","Test":"TestA","Elapsed":0.1}`,
				`BUILD FAILED: some error`,
				`{"Time":"2024-01-01T00:00:01Z","Action":"fail","Package":"pkg/foo","Test":"TestB","Elapsed":0.2}`,
			}, "\n"),
			wantEvents:     2,
			wantTUIContain: []string{"BUILD FAILED: some error"},
		},
		{
			name:       "package-level event with no test name",
			input:      `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"pkg/foo","Elapsed":1.5}`,
			wantEvents: 1,
		},
		{
			name: "run, pause, cont sequence",
			input: strings.Join([]string{
				`{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestParallel"}`,
				`{"Time":"2024-01-01T00:00:00Z","Action":"pause","Package":"pkg/foo","Test":"TestParallel"}`,
				`{"Time":"2024-01-01T00:00:01Z","Action":"cont","Package":"pkg/foo","Test":"TestParallel"}`,
				`{"Time":"2024-01-01T00:00:02Z","Action":"pass","Package":"pkg/foo","Test":"TestParallel","Elapsed":2.0}`,
			}, "\n"),
			wantEvents: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tui := &bytes.Buffer{}
			r := NewStreamingRunner(tui, nil)

			reader := strings.NewReader(tt.input)
			r.processJSONOutput(reader)

			// Check event count
			events := r.GetEvents()
			if len(events) != tt.wantEvents {
				t.Errorf("got %d events, want %d", len(events), tt.wantEvents)
				for i, e := range events {
					t.Logf("  event[%d]: Action=%q Package=%q Test=%q", i, e.Action, e.Package, e.Test)
				}
			}

			// Check TUI output contains expected substrings
			tuiOutput := tui.String()
			for _, want := range tt.wantTUIContain {
				if !strings.Contains(tuiOutput, want) {
					t.Errorf("TUI output %q does not contain %q", tuiOutput, want)
				}
			}

			// Check TUI output does not contain unwanted substrings
			for _, absent := range tt.wantTUIAbsent {
				if strings.Contains(tuiOutput, absent) {
					t.Errorf("TUI output %q should not contain %q", tuiOutput, absent)
				}
			}

			// Check exact TUI output when specified
			if tt.wantTUIOutput != "" && tuiOutput != tt.wantTUIOutput {
				t.Errorf("TUI output = %q, want %q", tuiOutput, tt.wantTUIOutput)
			}
		})
	}
}

func TestProcessJSONOutputNilTUIWriter(t *testing.T) {
	t.Run("malformed JSON with nil tuiWriter does not panic", func(t *testing.T) {
		r := NewStreamingRunner(nil, nil)
		reader := strings.NewReader("not json\n")

		// Should not panic
		r.processJSONOutput(reader)

		if len(r.GetEvents()) != 0 {
			t.Error("expected no events for malformed JSON")
		}
	})

	t.Run("valid output event with nil tuiWriter does not panic", func(t *testing.T) {
		r := NewStreamingRunner(nil, nil)
		reader := strings.NewReader(`{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestOne","Output":"hello\n"}`)

		// Should not panic
		r.processJSONOutput(reader)

		events := r.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Action != "output" {
			t.Errorf("event action = %q, want %q", events[0].Action, "output")
		}
	})
}

func TestProcessStderr(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		wantTUIContain []string
	}{
		{
			name:           "single error line",
			input:          "build error: missing import\n",
			wantTUIContain: []string{"build error: missing import"},
		},
		{
			name: "multiple error lines",
			input: strings.Join([]string{
				"error line 1",
				"error line 2",
			}, "\n"),
			wantTUIContain: []string{"error line 1", "error line 2"},
		},
		{
			name:           "empty input",
			input:          "",
			wantTUIContain: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tui := &bytes.Buffer{}
			r := NewStreamingRunner(tui, nil)

			reader := strings.NewReader(tt.input)
			r.processStderr(reader)

			tuiOutput := tui.String()
			for _, want := range tt.wantTUIContain {
				if !strings.Contains(tuiOutput, want) {
					t.Errorf("TUI output %q does not contain %q", tuiOutput, want)
				}
			}
		})
	}
}

func TestProcessStderrNilTUIWriter(t *testing.T) {
	r := NewStreamingRunner(nil, nil)
	reader := strings.NewReader("error output\n")

	// Should not panic with nil tuiWriter
	r.processStderr(reader)
}

func TestCountResults(t *testing.T) {
	tests := []struct {
		name        string
		events      []TestEvent
		wantPassed  int
		wantFailed  int
		wantSkipped int
	}{
		{
			name:        "empty events",
			events:      nil,
			wantPassed:  0,
			wantFailed:  0,
			wantSkipped: 0,
		},
		{
			name: "all passing",
			events: []TestEvent{
				{Action: "pass", Test: "TestA", Package: "pkg/foo"},
				{Action: "pass", Test: "TestB", Package: "pkg/foo"},
			},
			wantPassed:  2,
			wantFailed:  0,
			wantSkipped: 0,
		},
		{
			name: "all failing",
			events: []TestEvent{
				{Action: "fail", Test: "TestA", Package: "pkg/foo"},
				{Action: "fail", Test: "TestB", Package: "pkg/foo"},
			},
			wantPassed:  0,
			wantFailed:  2,
			wantSkipped: 0,
		},
		{
			name: "all skipped",
			events: []TestEvent{
				{Action: "skip", Test: "TestA", Package: "pkg/foo"},
			},
			wantPassed:  0,
			wantFailed:  0,
			wantSkipped: 1,
		},
		{
			name: "mixed results",
			events: []TestEvent{
				{Action: "pass", Test: "TestA", Package: "pkg/foo"},
				{Action: "fail", Test: "TestB", Package: "pkg/foo"},
				{Action: "skip", Test: "TestC", Package: "pkg/foo"},
				{Action: "pass", Test: "TestD", Package: "pkg/foo"},
			},
			wantPassed:  2,
			wantFailed:  1,
			wantSkipped: 1,
		},
		{
			name: "package-level events are not counted",
			events: []TestEvent{
				{Action: "pass", Test: "", Package: "pkg/foo"},
				{Action: "fail", Test: "", Package: "pkg/bar"},
				{Action: "pass", Test: "TestA", Package: "pkg/foo"},
			},
			wantPassed:  1,
			wantFailed:  0,
			wantSkipped: 0,
		},
		{
			name: "non-terminal actions are not counted",
			events: []TestEvent{
				{Action: "run", Test: "TestA", Package: "pkg/foo"},
				{Action: "output", Test: "TestA", Package: "pkg/foo"},
				{Action: "pause", Test: "TestA", Package: "pkg/foo"},
				{Action: "cont", Test: "TestA", Package: "pkg/foo"},
				{Action: "pass", Test: "TestA", Package: "pkg/foo"},
			},
			wantPassed:  1,
			wantFailed:  0,
			wantSkipped: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewStreamingRunner(nil, nil)
			r.mu.Lock()
			r.events = tt.events
			r.mu.Unlock()

			passed, failed, skipped := r.CountResults()
			if passed != tt.wantPassed {
				t.Errorf("passed = %d, want %d", passed, tt.wantPassed)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %d, want %d", failed, tt.wantFailed)
			}
			if skipped != tt.wantSkipped {
				t.Errorf("skipped = %d, want %d", skipped, tt.wantSkipped)
			}
		})
	}
}

func TestGetEvents(t *testing.T) {
	t.Run("returns copy of events", func(t *testing.T) {
		r := NewStreamingRunner(nil, nil)
		r.mu.Lock()
		r.events = []TestEvent{
			{Action: "pass", Test: "TestA", Package: "pkg/foo"},
			{Action: "fail", Test: "TestB", Package: "pkg/bar"},
		}
		r.mu.Unlock()

		got := r.GetEvents()
		if len(got) != 2 {
			t.Fatalf("expected 2 events, got %d", len(got))
		}
		if got[0].Test != "TestA" || got[0].Action != "pass" {
			t.Errorf("event[0] = %+v, want pass/TestA", got[0])
		}
		if got[1].Test != "TestB" || got[1].Action != "fail" {
			t.Errorf("event[1] = %+v, want fail/TestB", got[1])
		}

		// Modify returned slice; original should be unaffected
		got[0].Test = "MODIFIED"
		original := r.GetEvents()
		if original[0].Test != "TestA" {
			t.Error("GetEvents() did not return a defensive copy; mutation leaked to original")
		}
	})

	t.Run("returns empty slice when no events", func(t *testing.T) {
		r := NewStreamingRunner(nil, nil)
		got := r.GetEvents()
		if got == nil {
			t.Error("expected non-nil empty slice, got nil")
		}
		if len(got) != 0 {
			t.Errorf("expected 0 events, got %d", len(got))
		}
	})
}

func TestProcessJSONOutputEventFields(t *testing.T) {
	t.Run("all fields parsed correctly", func(t *testing.T) {
		input := `{"Time":"2024-06-15T10:30:00.123Z","Action":"pass","Package":"github.com/foo/bar","Test":"TestSomething","Output":"ok","Elapsed":1.234}`
		r := NewStreamingRunner(&bytes.Buffer{}, nil)
		r.processJSONOutput(strings.NewReader(input))

		events := r.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		e := events[0]
		if e.Time != "2024-06-15T10:30:00.123Z" {
			t.Errorf("Time = %q, want %q", e.Time, "2024-06-15T10:30:00.123Z")
		}
		if e.Action != "pass" {
			t.Errorf("Action = %q, want %q", e.Action, "pass")
		}
		if e.Package != "github.com/foo/bar" {
			t.Errorf("Package = %q, want %q", e.Package, "github.com/foo/bar")
		}
		if e.Test != "TestSomething" {
			t.Errorf("Test = %q, want %q", e.Test, "TestSomething")
		}
		if e.Output != "ok" {
			t.Errorf("Output = %q, want %q", e.Output, "ok")
		}
		if e.Elapsed != 1.234 {
			t.Errorf("Elapsed = %f, want %f", e.Elapsed, 1.234)
		}
	})

	t.Run("minimal JSON fields", func(t *testing.T) {
		input := `{"Action":"run"}`
		r := NewStreamingRunner(&bytes.Buffer{}, nil)
		r.processJSONOutput(strings.NewReader(input))

		events := r.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Action != "run" {
			t.Errorf("Action = %q, want %q", events[0].Action, "run")
		}
		if events[0].Elapsed != 0 {
			t.Errorf("Elapsed = %f, want 0 (omitempty default)", events[0].Elapsed)
		}
	})

	t.Run("extra unknown JSON fields are ignored", func(t *testing.T) {
		input := `{"Action":"pass","Test":"TestA","Package":"pkg","UnknownField":"value","Another":123}`
		r := NewStreamingRunner(&bytes.Buffer{}, nil)
		r.processJSONOutput(strings.NewReader(input))

		events := r.GetEvents()
		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}
		if events[0].Action != "pass" {
			t.Errorf("Action = %q, want %q", events[0].Action, "pass")
		}
	})
}

func TestProcessJSONOutputIntegrationWithCountResults(t *testing.T) {
	// Simulate a realistic go test -json output stream
	input := strings.Join([]string{
		`{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestAlpha"}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestAlpha","Output":"=== RUN   TestAlpha\n"}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestAlpha","Output":"--- PASS: TestAlpha (0.00s)\n"}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"pkg/foo","Test":"TestAlpha","Elapsed":0.0}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"run","Package":"pkg/foo","Test":"TestBeta"}`,
		`{"Time":"2024-01-01T00:00:00Z","Action":"output","Package":"pkg/foo","Test":"TestBeta","Output":"=== RUN   TestBeta\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Test":"TestBeta","Output":"    beta_test.go:10: assertion failed\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Test":"TestBeta","Output":"--- FAIL: TestBeta (1.00s)\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"fail","Package":"pkg/foo","Test":"TestBeta","Elapsed":1.0}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"run","Package":"pkg/foo","Test":"TestGamma"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Test":"TestGamma","Output":"=== RUN   TestGamma\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Test":"TestGamma","Output":"--- SKIP: TestGamma (0.00s)\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"skip","Package":"pkg/foo","Test":"TestGamma","Elapsed":0.0}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"output","Package":"pkg/foo","Output":"FAIL\n"}`,
		`{"Time":"2024-01-01T00:00:01Z","Action":"fail","Package":"pkg/foo","Elapsed":1.5}`,
	}, "\n")

	tui := &bytes.Buffer{}
	r := NewStreamingRunner(tui, nil)
	r.processJSONOutput(strings.NewReader(input))

	// Verify counts
	passed, failed, skipped := r.CountResults()
	if passed != 1 {
		t.Errorf("passed = %d, want 1", passed)
	}
	if failed != 1 {
		t.Errorf("failed = %d, want 1", failed)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1", skipped)
	}

	// Verify total events
	events := r.GetEvents()
	if len(events) != 15 {
		t.Errorf("total events = %d, want 15", len(events))
	}

	// Verify TUI output includes test output lines
	tuiOutput := tui.String()
	if !strings.Contains(tuiOutput, "--- PASS: TestAlpha") {
		t.Error("TUI output missing PASS line for TestAlpha")
	}
	if !strings.Contains(tuiOutput, "assertion failed") {
		t.Error("TUI output missing assertion failure for TestBeta")
	}
	if !strings.Contains(tuiOutput, "--- FAIL: TestBeta") {
		t.Error("TUI output missing FAIL line for TestBeta")
	}
	if !strings.Contains(tuiOutput, "--- SKIP: TestGamma") {
		t.Error("TUI output missing SKIP line for TestGamma")
	}
}

func TestProcessStderrAppendNewline(t *testing.T) {
	// Verify that processStderr appends a newline after each line
	tui := &bytes.Buffer{}
	r := NewStreamingRunner(tui, nil)

	input := strings.Join([]string{
		"line one",
		"line two",
	}, "\n")
	r.processStderr(strings.NewReader(input))

	output := tui.String()
	if !strings.Contains(output, "line one\n") {
		t.Errorf("expected newline after 'line one' in output %q", output)
	}
	if !strings.Contains(output, "line two\n") {
		t.Errorf("expected newline after 'line two' in output %q", output)
	}
}

func TestProcessJSONOutputMalformedAppendNewline(t *testing.T) {
	// Verify that malformed (non-JSON) lines get a newline appended
	tui := &bytes.Buffer{}
	r := NewStreamingRunner(tui, nil)

	input := "not-json-line"
	r.processJSONOutput(strings.NewReader(input))

	output := tui.String()
	if output != "not-json-line\n" {
		t.Errorf("expected %q, got %q", "not-json-line\n", output)
	}
}

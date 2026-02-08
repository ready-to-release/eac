package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestConsoleObserverImplementsExecutionObserver(t *testing.T) {
	obs := NewConsoleObserver()

	// Verify it implements ExecutionObserver
	var _ core.ExecutionObserver = obs
}

func TestConsoleObserverImplementsWriterFactory(t *testing.T) {
	obs := NewConsoleObserver()

	// Verify it implements WriterFactory
	var _ core.WriterFactory = obs
}

func TestConsoleObserverPhaseStarted(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.PhaseStartedEvent{
		Phase:     core.PhaseRun,
		Time:      now,
		TotalWork: 10,
	})

	output := buf.String()
	if !strings.Contains(output, "Run") {
		t.Errorf("Expected output to contain 'Run', got %q", output)
	}
}

func TestConsoleObserverPhaseCompleted(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.PhaseCompletedEvent{
		Phase:    core.PhaseRun,
		Time:     now,
		Success:  true,
		Summary:  "All passed",
		Duration: 5 * time.Second,
	})

	output := buf.String()
	if !strings.Contains(output, "Run") {
		t.Errorf("Expected output to contain 'Run', got %q", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("Expected output to contain 'OK', got %q", output)
	}
}

func TestConsoleObserverPhaseCompletedFailed(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.PhaseCompletedEvent{
		Phase:    core.PhaseRun,
		Time:     now,
		Success:  false,
		Summary:  "Tests failed",
		Duration: 5 * time.Second,
	})

	output := buf.String()
	if !strings.Contains(output, "FAILED") {
		t.Errorf("Expected output to contain 'FAILED', got %q", output)
	}
}

func TestConsoleObserverUnitStarted(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.UnitStartedEvent{
		Time: now,
		ID:   "build:core:go:go",
	})

	output := buf.String()
	if !strings.Contains(output, "START") {
		t.Errorf("Expected output to contain 'START', got %q", output)
	}
	if !strings.Contains(output, "build:core:go:go") {
		t.Errorf("Expected output to contain unit ID, got %q", output)
	}
}

func TestConsoleObserverUnitCompletedSuccess(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.UnitCompletedEvent{
		Time:     now,
		ID:       "build:core:go:go",
		ExitCode: 0,
		Duration: 2 * time.Second,
	})

	output := buf.String()
	if !strings.Contains(output, "OK") {
		t.Errorf("Expected output to contain 'OK', got %q", output)
	}
}

func TestConsoleObserverUnitCompletedFailed(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.UnitCompletedEvent{
		Time:     now,
		ID:       "build:core:go:go",
		ExitCode: 1,
		Duration: 2 * time.Second,
	})

	output := buf.String()
	if !strings.Contains(output, "FAIL") {
		t.Errorf("Expected output to contain 'FAIL', got %q", output)
	}
}

func TestConsoleObserverUnitCompletedCached(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.UnitCompletedEvent{
		Time:     now,
		ID:       "build:core:go:go",
		ExitCode: -1,
		Duration: 0,
	})

	output := buf.String()
	if !strings.Contains(output, "CACHED") {
		t.Errorf("Expected output to contain 'CACHED', got %q", output)
	}
}

func TestConsoleObserverSummaryReady(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()
	obs.OnEvent(core.SummaryReadyEvent{
		Time:      now,
		Success:   true,
		TotalTime: 30 * time.Second,
		Passed:    8,
		Failed:    0,
		Cached:    2,
		NextSteps: "Run: eac release",
	})

	output := buf.String()
	if !strings.Contains(output, "Summary") {
		t.Errorf("Expected output to contain 'Summary', got %q", output)
	}
	if !strings.Contains(output, "Passed: 8") {
		t.Errorf("Expected output to contain 'Passed: 8', got %q", output)
	}
	if !strings.Contains(output, "eac release") {
		t.Errorf("Expected output to contain next steps, got %q", output)
	}
}

func TestConsoleObserverNewWriter(t *testing.T) {
	obs := NewConsoleObserver()

	var buf bytes.Buffer
	writer := obs.NewWriter("test-unit", &buf)

	// Write something and verify it goes to the log writer
	n, err := writer.Write([]byte("test output\n"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != 12 {
		t.Errorf("Expected 12 bytes written, got %d", n)
	}

	// Close the writer
	if err := writer.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Verify the buffer received the data
	if buf.String() != "test output\n" {
		t.Errorf("Expected 'test output\\n', got %q", buf.String())
	}
}

func TestConsoleObserverTracksProgress(t *testing.T) {
	var buf bytes.Buffer
	obs := NewConsoleObserverWithWriter(&buf)

	now := time.Now()

	// Queue a unit
	obs.OnEvent(core.UnitQueuedEvent{
		Time:        now,
		ID:          "build:core:go:go",
		DisplayName: "core:go",
		Weight:      4,
	})

	// Start it
	obs.OnEvent(core.UnitStartedEvent{
		Time: now,
		ID:   "build:core:go:go",
	})

	// Complete it
	obs.OnEvent(core.UnitCompletedEvent{
		Time:     now,
		ID:       "build:core:go:go",
		ExitCode: 0,
		Duration: 2 * time.Second,
	})

	output := buf.String()
	if !strings.Contains(output, "START") {
		t.Errorf("Expected output to contain 'START', got %q", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("Expected output to contain 'OK', got %q", output)
	}
}

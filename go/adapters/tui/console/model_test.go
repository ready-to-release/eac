package console

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestListenForLines_StopsOnDoneChan verifies that listenForLines returns
// linesDoneMsg when the done channel is closed, even if lineChan is still open.
func TestListenForLines_StopsOnDoneChan(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		lineChan,
		nil,
		doneChan, // NEW: done channel parameter
		false,
		true,
		nil,
	)

	// Start listener in goroutine
	resultChan := make(chan tea.Msg, 1)
	go func() {
		cmd := model.listenForLines()
		resultChan <- cmd()
	}()

	// Give the goroutine time to start blocking on select
	time.Sleep(10 * time.Millisecond)

	// Close done channel - listener should return
	close(doneChan)

	select {
	case msg := <-resultChan:
		if _, ok := msg.(linesDoneMsg); !ok {
			t.Errorf("expected linesDoneMsg, got %T", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("listener did not return after done channel closed")
	}
}

// TestListenForStatus_StopsOnDoneChan verifies that listenForStatus returns
// statusDoneMsg when the done channel is closed, even if statusChan is still open.
func TestListenForStatus_StopsOnDoneChan(t *testing.T) {
	statusChan := make(chan Status, 10)
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		nil,
		statusChan,
		doneChan, // NEW: done channel parameter
		false,
		true,
		nil,
	)

	// Start listener in goroutine
	resultChan := make(chan tea.Msg, 1)
	go func() {
		cmd := model.listenForStatus()
		resultChan <- cmd()
	}()

	// Give the goroutine time to start blocking on select
	time.Sleep(10 * time.Millisecond)

	// Close done channel - listener should return
	close(doneChan)

	select {
	case msg := <-resultChan:
		if _, ok := msg.(statusDoneMsg); !ok {
			t.Errorf("expected statusDoneMsg, got %T", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("listener did not return after done channel closed")
	}
}

// TestListenForLines_ReturnsLineWhenAvailable verifies that listenForLines
// still returns lines normally when the done channel is open.
func TestListenForLines_ReturnsLineWhenAvailable(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		lineChan,
		nil,
		doneChan,
		false,
		true,
		nil,
	)

	// Send a line
	expectedLine := Line{Text: "test line", Level: LevelInfo}
	lineChan <- expectedLine

	// Start listener
	cmd := model.listenForLines()
	msg := cmd()

	lineMsg, ok := msg.(lineMsg)
	if !ok {
		t.Fatalf("expected lineMsg, got %T", msg)
	}
	if Line(lineMsg).Text != expectedLine.Text {
		t.Errorf("expected line text %q, got %q", expectedLine.Text, Line(lineMsg).Text)
	}
}

// TestListenForStatus_ReturnsStatusWhenAvailable verifies that listenForStatus
// still returns status updates normally when the done channel is open.
func TestListenForStatus_ReturnsStatusWhenAvailable(t *testing.T) {
	statusChan := make(chan Status, 10)
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		nil,
		statusChan,
		doneChan,
		false,
		true,
		nil,
	)

	// Send a status
	expectedStatus := Status{Phase: "test", Total: 5}
	statusChan <- expectedStatus

	// Start listener
	cmd := model.listenForStatus()
	msg := cmd()

	statusMsg, ok := msg.(statusMsg)
	if !ok {
		t.Fatalf("expected statusMsg, got %T", msg)
	}
	if Status(statusMsg).Total != expectedStatus.Total {
		t.Errorf("expected status Total %d, got %d", expectedStatus.Total, Status(statusMsg).Total)
	}
}

// TestListenForLines_NilLineChanReturnsImmediately verifies that listenForLines
// returns immediately when lineChan is nil.
func TestListenForLines_NilLineChanReturnsImmediately(t *testing.T) {
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		nil, // nil line channel
		nil,
		doneChan,
		false,
		true,
		nil,
	)

	// Start listener
	cmd := model.listenForLines()
	msg := cmd()

	if _, ok := msg.(linesDoneMsg); !ok {
		t.Errorf("expected linesDoneMsg for nil lineChan, got %T", msg)
	}
}

// TestListenForStatus_NilStatusChanReturnsImmediately verifies that listenForStatus
// returns immediately when statusChan is nil.
func TestListenForStatus_NilStatusChanReturnsImmediately(t *testing.T) {
	doneChan := make(chan struct{})

	model := NewModel(
		20,
		"Test",
		nil,
		nil, // nil status channel
		doneChan,
		false,
		true,
		nil,
	)

	// Start listener
	cmd := model.listenForStatus()
	msg := cmd()

	if _, ok := msg.(statusDoneMsg); !ok {
		t.Errorf("expected statusDoneMsg for nil statusChan, got %T", msg)
	}
}

// TestGetCapacityInfo verifies that GetCapacityInfo correctly combines
// running count from UoW states with roof and pressureTarget from status.
func TestGetCapacityInfo(t *testing.T) {
	tests := []struct {
		name           string
		uowStates      map[string]*UoWState
		roof           int
		pressureTarget int
		wantRunning    int
		wantRoof       int
		wantPressure   int
	}{
		{
			name:           "empty state",
			uowStates:      map[string]*UoWState{},
			roof:           24,
			pressureTarget: 24,
			wantRunning:    0,
			wantRoof:       24,
			wantPressure:   24,
		},
		{
			name: "running weight from UoW states",
			uowStates: map[string]*UoWState{
				"m1:c1": {Status: UoWRunning, Weight: 5},
				"m1:c2": {Status: UoWRunning, Weight: 3},
				"m2:c1": {Status: UoWComplete, Weight: 2}, // Not counted in running
			},
			roof:           24,
			pressureTarget: 16,
			wantRunning:    8, // 5 + 3
			wantRoof:       24,
			wantPressure:   16,
		},
		{
			name: "under pressure scenario (24/16)",
			uowStates: map[string]*UoWState{
				"m1:c1": {Status: UoWRunning, Weight: 1},
				"m1:c2": {Status: UoWRunning, Weight: 1},
				"m1:c3": {Status: UoWRunning, Weight: 1},
				"m1:c4": {Status: UoWRunning, Weight: 1},
				"m2:c1": {Status: UoWRunning, Weight: 1},
				"m2:c2": {Status: UoWRunning, Weight: 1},
				"m2:c3": {Status: UoWRunning, Weight: 1},
				"m2:c4": {Status: UoWRunning, Weight: 1},
				"m3:c1": {Status: UoWRunning, Weight: 1},
				"m3:c2": {Status: UoWRunning, Weight: 1},
				"m3:c3": {Status: UoWRunning, Weight: 1},
				"m3:c4": {Status: UoWRunning, Weight: 1},
				"m4:c1": {Status: UoWRunning, Weight: 1},
				"m4:c2": {Status: UoWRunning, Weight: 1},
				"m4:c3": {Status: UoWRunning, Weight: 1},
				"m4:c4": {Status: UoWRunning, Weight: 1},
				"m5:c1": {Status: UoWRunning, Weight: 1},
				"m5:c2": {Status: UoWRunning, Weight: 1},
				"m5:c3": {Status: UoWRunning, Weight: 1},
				"m5:c4": {Status: UoWRunning, Weight: 1},
				"m6:c1": {Status: UoWRunning, Weight: 1},
				"m6:c2": {Status: UoWRunning, Weight: 1},
				"m6:c3": {Status: UoWRunning, Weight: 1},
				"m6:c4": {Status: UoWRunning, Weight: 1},
			},
			roof:           24,
			pressureTarget: 16,
			wantRunning:    24,
			wantRoof:       24,
			wantPressure:   16,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				uowStates:      tt.uowStates,
				roof:           tt.roof,
				pressureTarget: tt.pressureTarget,
			}

			info := m.GetCapacityInfo()

			if info.Running != tt.wantRunning {
				t.Errorf("Running: got %d, want %d", info.Running, tt.wantRunning)
			}
			if info.Roof != tt.wantRoof {
				t.Errorf("Roof: got %d, want %d", info.Roof, tt.wantRoof)
			}
			if info.PressureTarget != tt.wantPressure {
				t.Errorf("PressureTarget: got %d, want %d", info.PressureTarget, tt.wantPressure)
			}
		})
	}
}

// TestGetCapacityInfo_UnderPressure verifies the IsUnderPressure detection.
func TestGetCapacityInfo_UnderPressure(t *testing.T) {
	m := Model{
		uowStates: map[string]*UoWState{
			"m1:c1": {Status: UoWRunning, Weight: 24},
		},
		roof:           24,
		pressureTarget: 16,
	}

	info := m.GetCapacityInfo()

	if !info.IsUnderPressure() {
		t.Error("expected IsUnderPressure() to be true when pressureTarget < roof")
	}
}

package console

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestListenForLines_StopsOnDoneChan verifies that listenForLines returns
// linesDoneMsg when the done channel is closed, even if lineChan is still open.
func TestListenForLines_StopsOnDoneChan(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

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

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		StatusChan:   statusChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

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
// returns a batchLineMsg containing the line when the done channel is open.
func TestListenForLines_ReturnsLineWhenAvailable(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

	// Send a line
	expectedLine := Line{Text: "test line", Level: LevelInfo}
	lineChan <- expectedLine

	// Start listener
	cmd := model.listenForLines()
	msg := cmd()

	batch, ok := msg.(batchLineMsg)
	if !ok {
		t.Fatalf("expected batchLineMsg, got %T", msg)
	}
	if len(batch) == 0 {
		t.Fatal("expected non-empty batch")
	}
	if batch[0].Text != expectedLine.Text {
		t.Errorf("expected line text %q, got %q", expectedLine.Text, batch[0].Text)
	}
}

// TestListenForLines_BatchesMutipleLines verifies that listenForLines
// drains multiple available lines into a single batch.
func TestListenForLines_BatchesMutipleLines(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

	// Send multiple lines before calling listener
	for i := 0; i < 5; i++ {
		lineChan <- Line{Text: fmt.Sprintf("line %d", i), Level: LevelInfo}
	}

	// Start listener - should batch all 5 lines
	cmd := model.listenForLines()
	msg := cmd()

	batch, ok := msg.(batchLineMsg)
	if !ok {
		t.Fatalf("expected batchLineMsg, got %T", msg)
	}
	if len(batch) != 5 {
		t.Errorf("expected 5 lines in batch, got %d", len(batch))
	}
	for i, line := range batch {
		expected := fmt.Sprintf("line %d", i)
		if line.Text != expected {
			t.Errorf("batch[%d].Text = %q, want %q", i, line.Text, expected)
		}
	}
}

// TestListenForStatus_ReturnsStatusWhenAvailable verifies that listenForStatus
// still returns status updates normally when the done channel is open.
func TestListenForStatus_ReturnsStatusWhenAvailable(t *testing.T) {
	statusChan := make(chan Status, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		StatusChan:   statusChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

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

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

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

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

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
				Execution: ExecutionState{
					UoWStates:      tt.uowStates,
					Roof:           tt.roof,
					PressureTarget: tt.pressureTarget,
				},
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

// TestListenForLines_DrainsLineChanOnDoneChan verifies that when doneChan closes
// while lineChan still has buffered lines, listenForLines drains those lines
// into a batchLineMsg instead of discarding them.
func TestListenForLines_DrainsLineChanOnDoneChan(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

	// Buffer lines BEFORE closing doneChan
	for i := 0; i < 5; i++ {
		lineChan <- Line{Text: fmt.Sprintf("buffered line %d", i), Level: LevelInfo}
	}

	// Close doneChan while lines are still buffered
	close(doneChan)

	// listenForLines should drain the buffered lines instead of returning linesDoneMsg
	cmd := model.listenForLines()
	msg := cmd()

	batch, ok := msg.(batchLineMsg)
	if !ok {
		t.Fatalf("expected batchLineMsg (drain), got %T", msg)
	}
	if len(batch) != 5 {
		t.Errorf("expected 5 drained lines, got %d", len(batch))
	}
	for i, line := range batch {
		expected := fmt.Sprintf("buffered line %d", i)
		if line.Text != expected {
			t.Errorf("batch[%d].Text = %q, want %q", i, line.Text, expected)
		}
	}

	// Next call should return linesDoneMsg since lineChan is now empty
	cmd2 := model.listenForLines()
	msg2 := cmd2()
	if _, ok := msg2.(linesDoneMsg); !ok {
		t.Errorf("expected linesDoneMsg on second call, got %T", msg2)
	}
}

// TestUoWCompleteMsg_NoAutoAdvanceOnFailure verifies that when a UoW fails
// (ExitCode > 0), the active tab does NOT auto-advance to the next running tab.
func TestUoWCompleteMsg_NoAutoAdvanceOnFailure(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

	// Register two UoWs
	model.GetOrCreateUoWStateByID("build:mod1:comp1:go", "mod1-go", 1)
	model.GetOrCreateUoWStateByID("build:mod2:comp2:go", "mod2-go", 1)
	model.MarkUoWRunning("build:mod1:comp1:go")
	model.MarkUoWRunning("build:mod2:comp2:go")

	// Set first tab as active (effective default)
	model.Interaction.ActiveTab = "build:mod1:comp1:go"

	// Complete first UoW with FAILURE (exit code 1)
	completeMsg := UoWCompleteMsg{
		Moniker:  "build:mod1:comp1:go",
		ExitCode: 1,
	}
	updatedModel, _ := model.Update(completeMsg)
	m := updatedModel.(Model)

	// Active tab should NOT have changed - failed tab stays focused
	if m.Interaction.ActiveTab != "build:mod1:comp1:go" {
		t.Errorf("expected active tab to stay on failed UoW %q, got %q",
			"build:mod1:comp1:go", m.Interaction.ActiveTab)
	}
}

// TestUoWCompleteMsg_AutoAdvanceOnSuccess verifies that when a UoW succeeds
// (ExitCode == 0), the active tab auto-advances to the next running tab.
func TestUoWCompleteMsg_AutoAdvanceOnSuccess(t *testing.T) {
	lineChan := make(chan Line, 10)
	doneChan := make(chan struct{})

	model := NewModel(ModelOptions{
		Height:       20,
		RunPhaseName: "Test",
		LineChan:     lineChan,
		DoneChan:     doneChan,
		SkipTUIDelay: true,
	})

	// Register two UoWs
	model.GetOrCreateUoWStateByID("build:mod1:comp1:go", "mod1-go", 1)
	model.GetOrCreateUoWStateByID("build:mod2:comp2:go", "mod2-go", 1)
	model.MarkUoWRunning("build:mod1:comp1:go")
	model.MarkUoWRunning("build:mod2:comp2:go")

	// Set first tab as active
	model.Interaction.ActiveTab = "build:mod1:comp1:go"

	// Complete first UoW with SUCCESS (exit code 0)
	completeMsg := UoWCompleteMsg{
		Moniker:  "build:mod1:comp1:go",
		ExitCode: 0,
	}
	updatedModel, _ := model.Update(completeMsg)
	m := updatedModel.(Model)

	// Active tab should have advanced to the next running UoW
	if m.Interaction.ActiveTab != "build:mod2:comp2:go" {
		t.Errorf("expected active tab to advance to %q, got %q",
			"build:mod2:comp2:go", m.Interaction.ActiveTab)
	}
}

// TestGetCapacityInfo_UnderPressure verifies the IsUnderPressure detection.
func TestGetCapacityInfo_UnderPressure(t *testing.T) {
	m := Model{
		Execution: ExecutionState{
			UoWStates: map[string]*UoWState{
				"m1:c1": {Status: UoWRunning, Weight: 24},
			},
			Roof:           24,
			PressureTarget: 16,
		},
	}

	info := m.GetCapacityInfo()

	if !info.IsUnderPressure() {
		t.Error("expected IsUnderPressure() to be true when pressureTarget < roof")
	}
}

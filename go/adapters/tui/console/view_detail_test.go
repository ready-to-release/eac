package console

import (
	"strings"
	"testing"
	"time"

	zone "github.com/lrstanley/bubblezone"
)

func TestRenderDetailPaneLineCount(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("mod:comp:tool", UoWRunning)

	result := m.renderDetailPane(100)
	lines := strings.Split(result, "\n")

	if len(lines) != detailPaneHeight {
		t.Errorf("renderDetailPane produced %d lines, want %d", len(lines), detailPaneHeight)
	}
}

func TestRenderDetailPaneEmptyState(t *testing.T) {
	zone.NewGlobal()
	m := Model{
		Display: DisplayConfig{
			Width:        120,
			Height:       40,
			RunPhaseName: "Building",
		},
		Execution: ExecutionState{
			Panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
			UoWStates: make(map[string]*UoWState),
		},
		Resources: ResourceState{
			Catalog: newTestCatalog(),
		},
	}

	result := m.renderDetailPane(100)
	lines := strings.Split(result, "\n")

	if len(lines) != detailPaneHeight {
		t.Errorf("empty state: renderDetailPane produced %d lines, want %d", len(lines), detailPaneHeight)
	}

	// Should contain "---" for empty values
	if !strings.Contains(result, "---") {
		t.Error("empty state should contain '---' placeholder values")
	}

	// Should contain "Detail" in header
	if !strings.Contains(result, "Detail") {
		t.Error("empty state should contain 'Detail' in header")
	}
}

func TestRenderDetailPaneRunningState(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWRunning)
	state := m.Execution.UoWStates["core:api-server:lint:golint"]
	state.Module = "core"
	state.Component = "api-server"
	state.Tool = "golint"
	state.StartTime = time.Now().Add(-12 * time.Second)

	result := m.renderDetailPane(100)

	// Should contain module, component, tool names
	if !strings.Contains(result, "core") {
		t.Error("should contain module name 'core'")
	}
	if !strings.Contains(result, "api-server") {
		t.Error("should contain component name 'api-server'")
	}
	if !strings.Contains(result, "golint") {
		t.Error("should contain tool name 'golint'")
	}

	// Should contain RUN status
	if !strings.Contains(result, "RUN") {
		t.Error("should contain 'RUN' status text")
	}
}

func TestRenderDetailPaneCompletedState(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWComplete)
	state := m.Execution.UoWStates["core:api-server:lint:golint"]
	state.Module = "core"
	state.Component = "api-server"
	state.Tool = "golint"
	state.StartTime = time.Now().Add(-5 * time.Second)
	state.EndTime = time.Now()

	result := m.renderDetailPane(100)

	if !strings.Contains(result, "DONE") {
		t.Error("completed state should contain 'DONE'")
	}
}

func TestRenderDetailPaneCachedState(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWSkipped)

	result := m.renderDetailPane(100)

	if !strings.Contains(result, "CACHED") {
		t.Error("cached state should contain 'CACHED'")
	}
}

func TestRenderDetailPaneFailedState(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWFailed)
	state := m.Execution.UoWStates["core:api-server:lint:golint"]
	state.ExitCode = 1

	result := m.renderDetailPane(100)

	if !strings.Contains(result, "FAIL") {
		t.Error("failed state should contain 'FAIL'")
	}
}

func TestRenderDetailPaneNarrowWidth(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWRunning)

	// Narrow width — should not panic
	result := m.renderDetailPane(50)
	lines := strings.Split(result, "\n")

	if len(lines) != detailPaneHeight {
		t.Errorf("narrow width: renderDetailPane produced %d lines, want %d", len(lines), detailPaneHeight)
	}
}

func TestRenderDetailPaneASCIIMode(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWRunning)
	m.Display.AsciiMode = true
	state := m.Execution.UoWStates["core:api-server:lint:golint"]
	state.Container = true

	result := m.renderDetailPane(100)

	// ASCII borders
	if !strings.Contains(result, "+") {
		t.Error("ASCII mode should use '+' for corners")
	}
	if !strings.Contains(result, "-") {
		t.Error("ASCII mode should use '-' for horizontal borders")
	}
	// Container type should be text
	if !strings.Contains(result, "docker") {
		t.Error("ASCII mode container type should be 'docker'")
	}
}

func TestLayoutMetricsDetailPaneHeight(t *testing.T) {
	zone.NewGlobal()

	tests := []struct {
		name         string
		height       int
		phaseStatus  PhaseStatus
		expectDetail int
	}{
		{
			name:         "active phase, sufficient height",
			height:       40,
			phaseStatus:  PhaseActive,
			expectDetail: detailPaneHeight,
		},
		{
			name:         "pending phase, no detail",
			height:       40,
			phaseStatus:  PhasePending,
			expectDetail: 0,
		},
		{
			name:         "active phase, tiny terminal",
			height:       14, // RemainingHeight = 9, which is < detailPaneHeight+5=11
			phaseStatus:  PhaseActive,
			expectDetail: 0,
		},
		{
			name:         "active phase, borderline height",
			height:       16, // RemainingHeight = 11, which == detailPaneHeight+5=11
			phaseStatus:  PhaseActive,
			expectDetail: detailPaneHeight,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{
				Display: DisplayConfig{
					Width:  120,
					Height: tt.height,
				},
				Execution: ExecutionState{
					Panes:     [3]*Pane{{}, {Status: tt.phaseStatus}, {}},
					UoWStates: make(map[string]*UoWState),
				},
				Resources: ResourceState{
					Catalog: newTestCatalog(),
				},
			}
			metrics := m.computeLayoutMetrics()
			if metrics.DetailPaneHeight != tt.expectDetail {
				t.Errorf("DetailPaneHeight = %d, want %d (RemainingHeight=%d)",
					metrics.DetailPaneHeight, tt.expectDetail, metrics.RemainingHeight)
			}
		})
	}
}

func TestRenderDetailPaneProgressLamps(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWRunning)
	m.Execution.Total = 10
	m.Execution.Completed = 5

	result := m.renderDetailPane(100)
	lines := strings.Split(result, "\n")

	if len(lines) != detailPaneHeight {
		t.Errorf("renderDetailPane produced %d lines, want %d", len(lines), detailPaneHeight)
	}

	// Should contain "Progress" label
	if !strings.Contains(result, "Progress") {
		t.Error("should contain 'Progress' label")
	}

	// Should contain progress count
	if !strings.Contains(result, "5/10") {
		t.Error("should contain progress count '5/10'")
	}
}

func TestRenderDetailPaneNewCellOrder(t *testing.T) {
	zone.NewGlobal()
	m := newTestModelWithUoW("core:api-server:lint:golint", UoWRunning)
	state := m.Execution.UoWStates["core:api-server:lint:golint"]
	state.Module = "core"
	state.Component = "api-server"
	state.Tool = "golint"

	result := m.renderDetailPane(100)

	// Row 1 should contain Module, Component, Type, Tool, UoW ID
	if !strings.Contains(result, "Module") {
		t.Error("should contain 'Module' label")
	}
	if !strings.Contains(result, "Component") {
		t.Error("should contain 'Component' label")
	}
	if !strings.Contains(result, "Type") {
		t.Error("should contain 'Type' label")
	}
	if !strings.Contains(result, "Tool") {
		t.Error("should contain 'Tool' label")
	}
	if !strings.Contains(result, "UoW ID") {
		t.Error("should contain 'UoW ID' label")
	}

	// Row 2 should contain Progress, State, Duration, Phase
	if !strings.Contains(result, "State") {
		t.Error("should contain 'State' label")
	}
	if !strings.Contains(result, "Duration") {
		t.Error("should contain 'Duration' label")
	}
	if !strings.Contains(result, "Phase") {
		t.Error("should contain 'Phase' label")
	}
}

// --- Helpers ---

func newTestModelWithUoW(moniker string, status UoWStatus) Model {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	m := Model{
		Display: DisplayConfig{
			Width:         120,
			Height:        40,
			RunPhaseName:  "Building",
			BufferSizeUoW: 50,
		},
		Execution: ExecutionState{
			Panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
			UoWStates: make(map[string]*UoWState),
			UoWOrder:  []string{moniker},
			Total:     10,
			Completed: 3,
		},
		Interaction: InteractionState{
			ActiveTab: moniker,
		},
		Resources: ResourceState{
			Catalog: catalog,
		},
	}

	parts := strings.Split(moniker, ":")
	state := &UoWState{
		Moniker:   moniker,
		Status:    status,
		Buffer:    NewRingBuffer(50),
		StartTime: time.Now(),
	}
	if len(parts) >= 3 {
		state.Module = parts[0]
		state.Component = parts[1]
		state.Tool = parts[len(parts)-1]
	}
	m.Execution.UoWStates[moniker] = state

	return m
}

func newTestCatalog() *WidgetCatalog {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)
	return catalog
}

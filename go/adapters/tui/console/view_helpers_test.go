package console

import (
	"strings"
	"testing"

	zone "github.com/lrstanley/bubblezone"
)

// TestHelpTextMap validates that all expected zone IDs have help text.
func TestHelpTextMap(t *testing.T) {
	expectedZones := []string{
		// Row 1
		"res-timer",
		"res-cpu",
		"res-mem",
		"res-dmem",
		// Row 2 - Scheduler lamps
		"res-host",
		"res-docker",
		// Row 2 - Weight counters
		"res-host-weight",
		"res-docker-weight",
		// Row 2/3 - Progress
		"res-counters",
		// Row 2/3 - Resources
		"res-jobs",
		"res-native",
		"res-ctools",
		// Controls
		"freeze-button",
	}

	for _, zoneID := range expectedZones {
		t.Run(zoneID, func(t *testing.T) {
			helpText, ok := HelpTextMap[zoneID]
			if !ok {
				t.Errorf("HelpTextMap missing entry for zone %q", zoneID)
				return
			}
			if helpText == "" {
				t.Errorf("HelpTextMap has empty help text for zone %q", zoneID)
			}
		})
	}
}

// TestRenderSelectedHelp validates the help text rendering.
func TestRenderSelectedHelp(t *testing.T) {
	m := Model{width: 120}

	tests := []struct {
		zoneID      string
		helpText    string
		expectLabel string
	}{
		{"res-timer", "Elapsed time since execution started", "Timer:"},
		{"res-cpu", "CPU usage per core", "CPU:"},
		{"res-mem", "Memory usage", "Memory:"},
		{"res-dmem", "Docker memory pool", "Docker Mem:"},
		{"res-host", "Host scheduler lamps", "Host:"},
		{"res-docker", "Docker scheduler lamps", "Docker:"},
		{"res-host-weight", "Host weight counter", "Host Weight:"},
		{"res-docker-weight", "Docker weight counter", "Docker Weight:"},
		{"res-counters", "Progress counters", "Counters:"},
		{"res-jobs", "Running containers", "Containers:"},
		{"res-native", "System tools", "Native:"},
		{"res-ctools", "Container tools", "Tools:"},
		{"freeze-button", "Click to pause auto-exit", "Freeze:"},
	}

	for _, tt := range tests {
		t.Run(tt.zoneID, func(t *testing.T) {
			result := m.renderSelectedHelp(tt.zoneID, tt.helpText, 100)

			// Result should contain the label
			if !strings.Contains(result, tt.expectLabel) {
				t.Errorf("renderSelectedHelp(%q, %q) = %q, expected to contain %q",
					tt.zoneID, tt.helpText, result, tt.expectLabel)
			}

			// Result should contain the help text
			if !strings.Contains(result, tt.helpText) {
				t.Errorf("renderSelectedHelp(%q, %q) = %q, expected to contain help text",
					tt.zoneID, tt.helpText, result)
			}
		})
	}
}

func TestGetModuleName(t *testing.T) {
	tests := []struct {
		name     string
		moniker  string
		expected string
	}{
		{
			name:     "simple module:component",
			moniker:  "eac-cli:impl/build",
			expected: "eac-cli",
		},
		{
			name:     "module with single component",
			moniker:  "core:go",
			expected: "core",
		},
		{
			name:     "module with nested component path",
			moniker:  "my-module:path/to/component",
			expected: "my-module",
		},
		{
			name:     "no colon - return as is",
			moniker:  "simple-name",
			expected: "simple-name",
		},
		{
			name:     "empty string",
			moniker:  "",
			expected: "",
		},
		{
			name:     "multiple colons - first part only",
			moniker:  "module:comp:extra",
			expected: "module",
		},
		{
			name:     "colon at start",
			moniker:  ":component",
			expected: "",
		},
		{
			name:     "colon at end",
			moniker:  "module:",
			expected: "module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getModuleName(tt.moniker)
			if result != tt.expected {
				t.Errorf("getModuleName(%q) = %q, want %q", tt.moniker, result, tt.expected)
			}
		})
	}
}

// TestLayoutMetricsConsistency validates that calculateLayoutMetrics() produces
// values that match what viewPanes() actually renders.
// This test catches layout drift where click detection becomes misaligned with rendering.
func TestLayoutMetricsConsistency(t *testing.T) {
	// Initialize bubblezone manager (required for rendering)
	zone.NewGlobal()

	tests := []struct {
		name         string
		setupModel   func() Model
		description  string
	}{
		{
			name: "minimal model - no resources or selected pane",
			setupModel: func() Model {
				m := Model{
					width:  120,
					height: 40,
					panes:  [3]*Pane{{}, {Status: PhaseActive}, {}},
				}
				return m
			},
			description: "Without scheduler, resources and selected panes should not appear",
		},
		{
			name: "with scheduler - resources and selected panes visible",
			setupModel: func() Model {
				m := Model{
					width:  120,
					height: 40,
					panes:  [3]*Pane{{}, {Status: PhaseActive}, {}},
					locks: []LockStatus{
						{Name: "component-scheduler", Capacity: 4, Used: 2},
						{Name: "docker-scheduler", Capacity: 6, Used: 3, Waiting: 1},
					},
				}
				return m
			},
			description: "With scheduler active, both panes should appear and be counted",
		},
		{
			name: "with summary data",
			setupModel: func() Model {
				m := Model{
					width:       120,
					height:      40,
					panes:       [3]*Pane{{}, {Status: PhaseActive}, {}},
					summaryData: &SummaryData{Success: true},
				}
				return m
			},
			description: "Summary data should reserve space at bottom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setupModel()
			metrics := m.calculateLayoutMetrics()

			// Validate ComponentsStart matches actual rendered line count
			// by counting lines before the side-by-side layout

			// Count init line (always 1)
			expectedLines := 1

			// Count resources pane lines
			if resourcesPane := m.renderResourcesPane(); resourcesPane != "" {
				expectedLines += strings.Count(resourcesPane, "\n") + 1
			}

			// Count selected pane lines
			if selectedPane := m.renderSelectedPane(); selectedPane != "" {
				expectedLines += strings.Count(selectedPane, "\n") + 1
			}

			// Add 1 for components panel header
			expectedLines += 1

			if metrics.ComponentsStart != expectedLines {
				t.Errorf("ComponentsStart mismatch: metrics=%d, expected=%d\n%s",
					metrics.ComponentsStart, expectedLines, tt.description)
			}

			// Validate individual components sum correctly
			computedStart := metrics.InitLines + metrics.ResourcesLines + metrics.SelectedLines + 1 // +1 for panel header
			if computedStart != metrics.ComponentsStart {
				t.Errorf("ComponentsStart (%d) != sum of components (%d+%d+%d+1=%d)",
					metrics.ComponentsStart,
					metrics.InitLines, metrics.ResourcesLines, metrics.SelectedLines,
					computedStart)
			}

			// Validate RemainingHeight is positive and reasonable
			if metrics.RemainingHeight < 5 {
				t.Errorf("RemainingHeight too small: %d (min 5)", metrics.RemainingHeight)
			}
		})
	}
}

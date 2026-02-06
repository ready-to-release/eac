package console

import (
	"strings"
	"testing"
	"time"

	zone "github.com/lrstanley/bubblezone"
)

// TestCatalogHelpText validates that all registered zone IDs have help text via the catalog.
func TestCatalogHelpText(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	for _, zoneID := range catalog.AllZoneIDs() {
		t.Run(zoneID, func(t *testing.T) {
			helpText, ok := catalog.HelpText(zoneID)
			if !ok {
				t.Errorf("catalog.HelpText(%q) returned false", zoneID)
				return
			}
			if helpText == "" {
				t.Errorf("catalog.HelpText(%q) returned empty help text", zoneID)
			}
		})
	}
}

// TestRenderSelectedHelp validates the help text rendering.
func TestRenderSelectedHelp(t *testing.T) {
	m := Model{width: 120}

	tests := []struct {
		elementName string
		helpText    string
		expectLabel string
	}{
		{"CPU", "CPU pressure lamps", "CPU:"},
		{"Progress", "Overall progress bar", "Progress:"},
		{"Memory", "Memory usage", "Memory:"},
		{"Docker Mem", "Docker memory pool", "Docker Mem:"},
		{"Host", "Host scheduler lamps", "Host:"},
		{"Docker", "Docker scheduler lamps", "Docker:"},
		{"Slots", "Scheduler slots", "Slots:"},
		{"Counters", "Progress counters", "Counters:"},
		{"Freeze", "Click to pause auto-exit", "Freeze:"},
	}

	for _, tt := range tests {
		t.Run(tt.elementName, func(t *testing.T) {
			result := m.renderSelectedHelp(tt.elementName, tt.helpText, 100)

			// Result should contain the label
			if !strings.Contains(result, tt.expectLabel) {
				t.Errorf("renderSelectedHelp(%q, %q) = %q, expected to contain %q",
					tt.elementName, tt.helpText, result, tt.expectLabel)
			}

			// Result should contain the help text
			if !strings.Contains(result, tt.helpText) {
				t.Errorf("renderSelectedHelp(%q, %q) = %q, expected to contain help text",
					tt.elementName, tt.helpText, result)
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

func TestExtractToolFromMoniker(t *testing.T) {
	tests := []struct {
		name     string
		moniker  string
		expected string
	}{
		{
			name:     "build UoW - 4 parts",
			moniker:  "build:eac-cli:go:go",
			expected: "go",
		},
		{
			name:     "build UoW - docker handler",
			moniker:  "build:books:dockerfile:buildx",
			expected: "buildx",
		},
		{
			name:     "test UoW - with extra (testname)",
			moniker:  "test:eac-cli:go:gotest:impl-build",
			expected: "gotest",
		},
		{
			name:     "test UoW - godog with spec",
			moniker:  "test:eac-cli:gherkin:godog:build-module",
			expected: "godog",
		},
		{
			name:     "lint UoW",
			moniker:  "lint:core:go:golangci-lint",
			expected: "golangci-lint",
		},
		{
			name:     "scan UoW - with category extra",
			moniker:  "scan:core:go:trivy-sbom:sbom",
			expected: "trivy-sbom",
		},
		{
			name:     "too few parts",
			moniker:  "build:module:component",
			expected: "",
		},
		{
			name:     "empty string",
			moniker:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractToolFromMoniker(tt.moniker)
			if result != tt.expected {
				t.Errorf("extractToolFromMoniker(%q) = %q, want %q", tt.moniker, result, tt.expected)
			}
		})
	}
}

// TestLayoutMetricsConsistency validates that calculateLayoutMetrics() produces
// correct fixed values for the 2-section layout (side-by-side + status bar).
func TestLayoutMetricsConsistency(t *testing.T) {
	// Initialize bubblezone manager (required for rendering)
	zone.NewGlobal()

	tests := []struct {
		name       string
		width      int
		height     int
	}{
		{"standard terminal", 120, 40},
		{"small terminal", 80, 24},
		{"tall terminal", 120, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := NewWidgetCatalog()
			RegisterAllWidgets(catalog)
			m := Model{
				width:     tt.width,
				height:    tt.height,
				panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
				uowStates: make(map[string]*UoWState),
				catalog:   catalog,
			}
			metrics := m.calculateLayoutMetrics()

			// ComponentsStart is always 1 (panel header only, no resources pane)
			if metrics.ComponentsStart != 1 {
				t.Errorf("ComponentsStart should always be 1, got %d", metrics.ComponentsStart)
			}

			// SummaryLines is always 5 (header + data + help + lamps + footer)
			if metrics.SummaryLines != 5 {
				t.Errorf("SummaryLines should be 5, got %d", metrics.SummaryLines)
			}

			// RemainingHeight = height - SummaryLines, min 5
			expectedRemaining := tt.height - 5
			if expectedRemaining < 5 {
				expectedRemaining = 5
			}
			if metrics.RemainingHeight != expectedRemaining {
				t.Errorf("RemainingHeight: got %d, want %d", metrics.RemainingHeight, expectedRemaining)
			}
		})
	}
}

// === Tab View Mode Tests ===

func TestTabViewModeString(t *testing.T) {
	tests := []struct {
		mode     TabViewMode
		expected string
	}{
		{TabViewName, "Name"},
		{TabViewModule, "Module"},
		{TabViewType, "Type"},
		{TabViewTool, "Tool"},
		{TabViewExec, "Mode"},
		{TabViewState, "State"},
		{TabViewMode(99), "Name"}, // unknown defaults to Name
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.expected {
			t.Errorf("TabViewMode(%d).String() = %q, want %q", tt.mode, got, tt.expected)
		}
	}
}

func TestTabViewModeCycling(t *testing.T) {
	// Cycling through all modes should wrap back to Name
	mode := TabViewName
	for i := 0; i < tabViewModeCount; i++ {
		mode = (mode + 1) % tabViewModeCount
	}
	if mode != TabViewName {
		t.Errorf("after cycling through all modes, got %v, want TabViewName", mode)
	}
}

func TestTabLabelForViewMode(t *testing.T) {
	inst := TabInstance{
		DisplayName:   "go",
		Module:        "eac-cli",
		Component:     "go",
		Tool:          "gotest",
		ComponentType: "go-lib",
		Container:     false,
	}

	tests := []struct {
		mode     TabViewMode
		expected string
	}{
		{TabViewName, "go"},
		{TabViewModule, "eac-cli"},
		{TabViewType, "go-lib"},
		{TabViewTool, "gotest"},
		{TabViewExec, "host"},
		{TabViewState, "pending"}, // Status defaults to UoWPending (zero value)
	}
	for _, tt := range tests {
		got := tabLabelForViewMode(inst, tt.mode)
		if got != tt.expected {
			t.Errorf("tabLabelForViewMode(%v) = %q, want %q", tt.mode, got, tt.expected)
		}
	}

	// Container mode
	containerInst := inst
	containerInst.Container = true
	got := tabLabelForViewMode(containerInst, TabViewExec)
	if got != "container" {
		t.Errorf("tabLabelForViewMode(TabViewExec) with Container=true = %q, want %q", got, "container")
	}

	// Fallback to DisplayName when field is empty
	emptyInst := TabInstance{DisplayName: "fallback"}
	for _, mode := range []TabViewMode{TabViewModule, TabViewType, TabViewTool} {
		got := tabLabelForViewMode(emptyInst, mode)
		if got != "fallback" {
			t.Errorf("tabLabelForViewMode(%v) with empty fields = %q, want %q", mode, got, "fallback")
		}
	}
}

func TestFormatTabState(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		inst     TabInstance
		expected string
	}{
		{
			name:     "pending",
			inst:     TabInstance{Status: UoWPending},
			expected: "pending",
		},
		{
			name:     "running with start time",
			inst:     TabInstance{Status: UoWRunning, StartTime: now.Add(-5 * time.Second)},
			expected: "run 5s",
		},
		{
			name:     "running without start time",
			inst:     TabInstance{Status: UoWRunning},
			expected: "running",
		},
		{
			name:     "complete with timing",
			inst:     TabInstance{Status: UoWComplete, StartTime: now.Add(-12 * time.Second), EndTime: now},
			expected: "done 12s",
		},
		{
			name:     "complete without timing",
			inst:     TabInstance{Status: UoWComplete},
			expected: "done",
		},
		{
			name:     "skipped/cached",
			inst:     TabInstance{Status: UoWSkipped},
			expected: "cached",
		},
		{
			name:     "failed with exit code",
			inst:     TabInstance{Status: UoWFailed, ExitCode: 1},
			expected: "fail:1",
		},
		{
			name:     "failed without exit code",
			inst:     TabInstance{Status: UoWFailed},
			expected: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTabState(tt.inst)
			if got != tt.expected {
				t.Errorf("formatTabState() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFormatElapsed(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{500 * time.Millisecond, "<1s"},
		{1 * time.Second, "1s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m"},
		{90 * time.Second, "1m30s"},
		{120 * time.Second, "2m"},
		{150 * time.Second, "2m30s"},
	}
	for _, tt := range tests {
		got := formatElapsed(tt.d)
		if got != tt.expected {
			t.Errorf("formatElapsed(%v) = %q, want %q", tt.d, got, tt.expected)
		}
	}
}

// === Dynamic Sizing Tests ===

func TestComponentsWidth(t *testing.T) {
	tests := []struct {
		name          string
		tabWidth      int
		paneWidthCols int
		termWidth     int
		expectMin     int
		expectMax     int
	}{
		{
			name:          "default 4-col at 15w",
			tabWidth:      15, paneWidthCols: 4, termWidth: 200,
			expectMin:     65, expectMax: 65, // 4*15 + 3*1 + 2 = 65
		},
		{
			name:          "2-col at 10w",
			tabWidth:      10, paneWidthCols: 2, termWidth: 200,
			expectMin:     23, expectMax: 23, // 2*10 + 1*1 + 2 = 23
		},
		{
			name:          "minimum width enforced",
			tabWidth:      10, paneWidthCols: 1, termWidth: 200,
			expectMin:     20, expectMax: 20, // 1*10 + 0 + 2 = 12 → clamped to 20
		},
		{
			name:          "max clamp snaps to exact columns",
			tabWidth:      30, paneWidthCols: 6, termWidth: 100,
			expectMin:     32, expectMax: 32, // max=60, cols=(60-2+1)/31=1, 1*30+0+2=32
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{tabWidth: tt.tabWidth, paneWidthCols: tt.paneWidthCols, width: tt.termWidth}
			got := m.ComponentsWidth()
			if got < tt.expectMin || got > tt.expectMax {
				t.Errorf("ComponentsWidth() = %d, want [%d, %d]", got, tt.expectMin, tt.expectMax)
			}
		})
	}
}

func TestMaxPaneWidthCols(t *testing.T) {
	tests := []struct {
		name      string
		tabWidth  int
		termWidth int
		expected  int
	}{
		{"wide terminal", 15, 200, 7},          // min(160,120)=120, (120-2+1)/(15+1) = 7.4 → 7
		{"standard terminal", 15, 120, 4},      // min(80,72)=72, (72-2+1)/(15+1) = 4.4 → 4
		{"narrow terminal", 15, 80, 2},          // min(40,48)=40, (40-2+1)/(15+1) = 2.4 → 2
		{"very narrow", 15, 50, 1},              // min(10,30)=10, 10 < 15+2, returns 1
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Model{tabWidth: tt.tabWidth, width: tt.termWidth}
			got := m.maxPaneWidthCols()
			if got != tt.expected {
				t.Errorf("maxPaneWidthCols() = %d, want %d", got, tt.expected)
			}
		})
	}
}

package console

import (
	"strings"
	"testing"
	"time"

	zone "github.com/lrstanley/bubblezone"
)

func TestCatalogRegistration(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	// Every registered widget should have a non-empty ElementName
	for _, id := range catalog.AllZoneIDs() {
		name := catalog.ElementName(id)
		if name == "" {
			t.Errorf("widget %q has empty ElementName", id)
		}
	}
}

func TestCatalogHelpTextComplete(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	// Every zone-enabled widget should have help text
	for _, id := range catalog.AllZoneIDs() {
		helpText, ok := catalog.HelpText(id)
		if !ok || helpText == "" {
			t.Errorf("widget %q missing help text", id)
		}
	}
}

func TestCatalogRenderNotNil(t *testing.T) {
	zone.NewGlobal()
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	snap := WidgetSnapshot{AsciiMode: true}
	for _, id := range catalog.AllZoneIDs() {
		w := catalog.Get(id)
		if w.Render == nil {
			t.Errorf("widget %q has nil Render function", id)
		}
		// Just verify it doesn't panic - some widgets return empty for zero state
		_ = catalog.RenderWidget(id, snap)
	}
}

func TestCatalogDuplicatePanics(t *testing.T) {
	catalog := NewWidgetCatalog()
	catalog.Register(&Widget{ID: "test", Render: func(snap WidgetSnapshot) string { return "x" }})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	catalog.Register(&Widget{ID: "test", Render: func(snap WidgetSnapshot) string { return "y" }})
}

func TestCatalogTabWidgetDuplicatePanics(t *testing.T) {
	catalog := NewWidgetCatalog()
	catalog.RegisterTabWidget(&TabWidgetDef{
		ID:     "tab1",
		Render: func(instance TabInstance, sizing TabSizing) string { return "x" },
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate tab widget registration")
		}
	}()
	catalog.RegisterTabWidget(&TabWidgetDef{
		ID:     "tab2",
		Render: func(instance TabInstance, sizing TabSizing) string { return "y" },
	})
}

func TestWidgetSnapshotConstruction(t *testing.T) {
	zone.NewGlobal()
	m := Model{
		width:     120,
		height:    40,
		panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
		startTime: time.Now(),
		uowStates: make(map[string]*UoWState),
	}
	snap := m.buildWidgetSnapshot()

	if snap.Width != 120 {
		t.Errorf("snap.Width = %d, want 120", snap.Width)
	}
	if snap.Elapsed < 0 {
		t.Error("snap.Elapsed is negative")
	}
	if snap.RunPhaseActive != true {
		t.Error("snap.RunPhaseActive should be true when PhaseRun is active")
	}
}

func TestAllWidgetsZoned(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	for _, id := range catalog.order {
		w := catalog.Get(id)
		if !w.ZoneEnabled {
			t.Errorf("widget %q has ZoneEnabled=false, all widgets should be zoned", id)
		}
	}
}

func TestTabWidgetRegistered(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	if catalog.TabWidget == nil {
		t.Fatal("tab template widget not registered")
	}
	if catalog.TabWidget.ElementName == "" {
		t.Error("tab widget has empty ElementName")
	}
	if catalog.TabWidget.HelpText == "" {
		t.Error("tab widget has empty HelpText")
	}
	if catalog.TabWidget.Render == nil {
		t.Error("tab widget has nil Render function")
	}
	if len(catalog.TabWidget.ColorMap) == 0 {
		t.Error("tab widget has empty ColorMap")
	}
}

func TestTabWidgetRender(t *testing.T) {
	zone.NewGlobal()
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	sizing := ComputeTabSizing(66, 4, 0, false)

	tests := []struct {
		name     string
		instance TabInstance
	}{
		{
			name: "pending tab",
			instance: TabInstance{
				Moniker: "build:core:go:go", DisplayName: "go",
				Status: UoWPending, Weight: 3,
			},
		},
		{
			name: "running active tab",
			instance: TabInstance{
				Moniker: "build:eac-cli:go:go", DisplayName: "eac-cli",
				Status: UoWRunning, Weight: 8, IsActive: true,
			},
		},
		{
			name: "hovered tab with long name",
			instance: TabInstance{
				Moniker: "test:core:gherkin:godog:config", DisplayName: "very-long-component-name",
				Status: UoWComplete, Weight: 2, IsHovered: true,
			},
		},
		{
			name: "failed tab",
			instance: TabInstance{
				Moniker: "lint:core:go:golangci-lint", DisplayName: "golangci",
				Status: UoWFailed, Weight: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := catalog.RenderTab(tt.instance, sizing)
			if result == "" {
				t.Error("tab rendered empty string")
			}
		})
	}
}

func TestComputeTabSizing(t *testing.T) {
	tests := []struct {
		name       string
		panelWidth int
		tabWidth   int
		expectCols int
	}{
		{
			name:       "standard 15w tab at 66w panel",
			panelWidth: 66, tabWidth: 15,
			expectCols: 4, // (66-2+1)/(15+1) = 4.06 → 4
		},
		{
			name:       "wide panel 15w tabs at 120w",
			panelWidth: 120, tabWidth: 15,
			expectCols: 7, // (120-2+1)/(15+1) = 7.4 → 7
		},
		{
			name:       "narrow panel 15w tabs at 40w",
			panelWidth: 40, tabWidth: 15,
			expectCols: 2, // (40-2+1)/(15+1) = 2.4 → 2
		},
		{
			name:       "minimum tab width enforced",
			panelWidth: 50, tabWidth: 3,
			expectCols: 6, // clamped to 7, (50-2+1)/(7+1) = 6.1 → 6
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sizing := ComputeTabSizing(tt.panelWidth, tt.tabWidth, 0, false)
			if sizing.TabColumns != tt.expectCols {
				t.Errorf("TabColumns = %d, want %d for panel=%d tabWidth=%d",
					sizing.TabColumns, tt.expectCols, tt.panelWidth, tt.tabWidth)
			}
			if sizing.LabelWidth != sizing.TabWidth-sizing.BadgeWidth {
				t.Errorf("LabelWidth = %d, want TabWidth(%d) - BadgeWidth(%d) = %d",
					sizing.LabelWidth, sizing.TabWidth, sizing.BadgeWidth, sizing.TabWidth-sizing.BadgeWidth)
			}
		})
	}
}

func TestTabSizingConsistency(t *testing.T) {
	// Verify that TabWidth is passed through unchanged (above minimum)
	for tw := 10; tw <= 30; tw++ {
		for pw := 40; pw <= 160; pw += 10 {
			sizing := ComputeTabSizing(pw, tw, 0, false)
			if sizing.TabWidth != tw {
				t.Errorf("tabWidth=%d panel=%d: TabWidth=%d, want %d",
					tw, pw, sizing.TabWidth, tw)
			}
			// Columns should fit within panel
			reconstructed := sizing.TabColumns*sizing.TabWidth + (sizing.TabColumns-1)*tabGap + tabBorder
			if reconstructed > pw {
				t.Errorf("tabWidth=%d panel=%d cols=%d: reconstructed=%d > panelWidth=%d",
					tw, pw, sizing.TabColumns, reconstructed, pw)
			}
		}
	}
}

func TestCatalogHelpTextForMoniker(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	// Tab moniker should return tab widget help text
	helpText, ok := catalog.HelpText("build:eac-cli:go:go")
	if !ok {
		t.Error("expected help text for moniker-like zone ID")
	}
	if helpText == "" {
		t.Error("expected non-empty help text for moniker-like zone ID")
	}

	// Non-moniker, non-registered should return false
	_, ok = catalog.HelpText("unknown-zone")
	if ok {
		t.Error("expected no help text for unknown zone ID")
	}
}

func TestCatalogElementNameForMoniker(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	name := catalog.ElementName("build:eac-cli:go:go")
	if name != "Component" {
		t.Errorf("ElementName for moniker = %q, want %q", name, "Component")
	}

	name = catalog.ElementName("unknown-zone")
	if name != "" {
		t.Errorf("ElementName for unknown zone = %q, want empty", name)
	}
}

func TestCatalogGetUnknownWidget(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	w := catalog.Get("nonexistent")
	if w != nil {
		t.Error("expected nil for nonexistent widget")
	}

	result := catalog.RenderWidget("nonexistent", WidgetSnapshot{})
	if result != "" {
		t.Errorf("RenderWidget for nonexistent = %q, want empty", result)
	}
}

func TestCatalogRenderTabWithNoTemplate(t *testing.T) {
	catalog := NewWidgetCatalog()
	// Don't register tab widget

	result := catalog.RenderTab(TabInstance{Moniker: "test:a:b:c"}, TabSizing{TabWidth: 15, LabelWidth: 12, BadgeWidth: 3})
	if result != "" {
		t.Errorf("RenderTab with no template = %q, want empty", result)
	}
}

func TestAllZoneIDsCount(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	ids := catalog.AllZoneIDs()
	// We registered 11 singleton widgets
	if len(ids) != 11 {
		t.Errorf("AllZoneIDs() returned %d IDs, want 11: %v", len(ids), ids)
	}
}

func TestWidgetSnapshotWeightedProgress(t *testing.T) {
	zone.NewGlobal()
	m := Model{
		width:     120,
		height:    40,
		panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
		startTime: time.Now(),
		uowStates: map[string]*UoWState{
			"a": {Status: UoWComplete, Weight: 4},
			"b": {Status: UoWRunning, Weight: 2},
			"c": {Status: UoWPending, Weight: 1},
			"d": {Status: UoWFailed, Weight: 3},
		},
	}
	snap := m.buildWidgetSnapshot()

	// TotalWeight = 4 + 2 + 1 + 3 = 10
	if snap.TotalWeight != 10 {
		t.Errorf("TotalWeight = %d, want 10", snap.TotalWeight)
	}
	// FinalizedWeight = 4 (complete) + 3 (failed) = 7 (running + pending don't count)
	if snap.FinalizedWeight != 7 {
		t.Errorf("FinalizedWeight = %d, want 7", snap.FinalizedWeight)
	}
}

func TestComputeTabSizingBounds(t *testing.T) {
	// tabWidth below minimum (badgeWidth+4=7) should clamp up
	sizing := ComputeTabSizing(100, 3, 0, false)
	if sizing.TabWidth < 7 {
		t.Errorf("tabWidth 3 should clamp to minimum 7, got %d", sizing.TabWidth)
	}

	// At least 1 column must always be returned
	sizing = ComputeTabSizing(10, 30, 0, false)
	if sizing.TabColumns < 1 {
		t.Errorf("should always have at least 1 column, got %d", sizing.TabColumns)
	}
}

func TestComputeTabSizingMarqueeAndAscii(t *testing.T) {
	sizing := ComputeTabSizing(66, 15, 42, true)
	if sizing.MarqueePos != 42 {
		t.Errorf("MarqueePos = %d, want 42", sizing.MarqueePos)
	}
	if !sizing.AsciiMode {
		t.Error("AsciiMode should be true")
	}
}

// TestCatalogZoneIDsMatchExpected ensures the catalog produces the same zone IDs
// that TestCatalogHelpText validates via catalog.AllZoneIDs().
func TestCatalogZoneIDsMatchExpected(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	expected := []string{
		"res-cpu",
		"res-slots",
		"res-counters",
		"progress-count",
		"status-text",
		"res-mem",
		"res-dmem",
		"res-host",
		"res-docker",
		"res-progress",
		"freeze-button",
	}

	ids := catalog.AllZoneIDs()
	if len(ids) != len(expected) {
		t.Fatalf("AllZoneIDs() returned %d IDs, want %d: %v", len(ids), len(expected), ids)
	}

	// Build a set for membership check (order may differ)
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}
	for _, exp := range expected {
		if !idSet[exp] {
			t.Errorf("expected zone ID %q not found in catalog", exp)
		}
	}
}

func TestWidgetSnapshotZeroWeight(t *testing.T) {
	zone.NewGlobal()
	m := Model{
		width:     120,
		height:    40,
		panes:     [3]*Pane{{}, {Status: PhaseActive}, {}},
		startTime: time.Now(),
		uowStates: map[string]*UoWState{
			"a": {Status: UoWComplete, Weight: 0}, // Zero weight treated as 1
			"b": {Status: UoWPending, Weight: 0},
		},
	}
	snap := m.buildWidgetSnapshot()

	if snap.TotalWeight != 2 {
		t.Errorf("TotalWeight = %d, want 2 (zero weights treated as 1)", snap.TotalWeight)
	}
	if snap.FinalizedWeight != 1 {
		t.Errorf("FinalizedWeight = %d, want 1", snap.FinalizedWeight)
	}
}

// Verify status-text widget handles the icon/label shadowing correctly
func TestStatusTextWidgetSuccess(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	snap := WidgetSnapshot{
		SummaryData: &SummaryData{Success: true, TotalTime: 5 * time.Second},
	}
	result := catalog.RenderWidget("status-text", snap)
	if !strings.Contains(result, "Complete") {
		t.Errorf("expected 'Complete' in status text, got %q", result)
	}
}

func TestStatusTextWidgetFailure(t *testing.T) {
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	snap := WidgetSnapshot{
		SummaryData: &SummaryData{Success: false, TotalTime: 3 * time.Second},
	}
	result := catalog.RenderWidget("status-text", snap)
	if !strings.Contains(result, "Failed") {
		t.Errorf("expected 'Failed' in status text, got %q", result)
	}
}

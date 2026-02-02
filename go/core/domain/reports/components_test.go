package reports

import (
	"testing"
)

func TestFilterComponents_NoFilters(t *testing.T) {
	components := []*ComponentInfo{
		{Moniker: "mod1", Component: "go", Type: "go"},
		{Moniker: "mod2", Component: "ts", Type: "typescript"},
	}

	filters := ComponentFilters{}
	result := FilterComponents(components, filters)

	if len(result) != 2 {
		t.Errorf("expected 2 components, got %d", len(result))
	}
}

func TestFilterComponents_ModuleFilter(t *testing.T) {
	components := []*ComponentInfo{
		{Moniker: "mod1", Component: "go", Type: "go"},
		{Moniker: "mod2", Component: "ts", Type: "typescript"},
	}

	filters := ComponentFilters{Module: "mod1"}
	result := FilterComponents(components, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 component, got %d", len(result))
	}
	if result[0].Moniker != "mod1" {
		t.Errorf("expected mod1, got %s", result[0].Moniker)
	}
}

func TestFilterComponents_TypeFilter(t *testing.T) {
	components := []*ComponentInfo{
		{Moniker: "mod1", Component: "go", Type: "go"},
		{Moniker: "mod2", Component: "ts", Type: "typescript"},
	}

	filters := ComponentFilters{Type: "go"}
	result := FilterComponents(components, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 component, got %d", len(result))
	}
	if result[0].Type != "go" {
		t.Errorf("expected go type, got %s", result[0].Type)
	}
}

func TestFilterComponents_BuildableFilter(t *testing.T) {
	components := []*ComponentInfo{
		{
			Moniker:   "mod1",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Build: &PhaseInfo{Enabled: true, Tool: "go"},
			},
		},
		{
			Moniker:   "mod2",
			Component: "md",
			Type:      "markdown",
			Phases:    &ComponentPhases{},
		},
	}

	filters := ComponentFilters{Buildable: true}
	result := FilterComponents(components, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 component, got %d", len(result))
	}
	if result[0].Component != "go" {
		t.Errorf("expected go component, got %s", result[0].Component)
	}
}

func TestFilterComponents_LintableFilter(t *testing.T) {
	components := []*ComponentInfo{
		{
			Moniker:   "mod1",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Lint: &PhaseInfo{Enabled: true, Tools: []string{"golangci-lint"}},
			},
		},
		{
			Moniker:   "mod2",
			Component: "static",
			Type:      "static",
			Phases:    &ComponentPhases{},
		},
	}

	filters := ComponentFilters{Lintable: true}
	result := FilterComponents(components, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 component, got %d", len(result))
	}
}

func TestFilterComponents_CombinedFilters(t *testing.T) {
	components := []*ComponentInfo{
		{
			Moniker:   "mod1",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Build: &PhaseInfo{Enabled: true, Tool: "go"},
				Lint:  &PhaseInfo{Enabled: true, Tools: []string{"golangci-lint"}},
			},
		},
		{
			Moniker:   "mod1",
			Component: "md",
			Type:      "markdown",
			Phases: &ComponentPhases{
				Lint: &PhaseInfo{Enabled: true, Tools: []string{"markdownlint"}},
			},
		},
		{
			Moniker:   "mod2",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Build: &PhaseInfo{Enabled: true, Tool: "go"},
			},
		},
	}

	// Filter for mod1 AND buildable
	filters := ComponentFilters{Module: "mod1", Buildable: true}
	result := FilterComponents(components, filters)

	if len(result) != 1 {
		t.Errorf("expected 1 component, got %d", len(result))
	}
	if result[0].Component != "go" {
		t.Errorf("expected go component, got %s", result[0].Component)
	}
}

func TestComponentFilters_HasFilters(t *testing.T) {
	tests := []struct {
		name     string
		filters  ComponentFilters
		expected bool
	}{
		{"empty", ComponentFilters{}, false},
		{"module only", ComponentFilters{Module: "mod1"}, true},
		{"type only", ComponentFilters{Type: "go"}, true},
		{"buildable only", ComponentFilters{Buildable: true}, true},
		{"lintable only", ComponentFilters{Lintable: true}, true},
		{"testable only", ComponentFilters{Testable: true}, true},
		{"scannable only", ComponentFilters{Scannable: true}, true},
		{"combined", ComponentFilters{Module: "mod1", Buildable: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.filters.HasFilters(); got != tt.expected {
				t.Errorf("HasFilters() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBuildComponentSummary(t *testing.T) {
	components := []*ComponentInfo{
		{
			Moniker:   "mod1",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Build: &PhaseInfo{Enabled: true},
				Lint:  &PhaseInfo{Enabled: true},
				Test:  &PhaseInfo{Enabled: true},
				Scan:  &PhaseInfo{Enabled: true},
			},
		},
		{
			Moniker:   "mod1",
			Component: "md",
			Type:      "markdown",
			Phases: &ComponentPhases{
				Lint: &PhaseInfo{Enabled: true},
			},
		},
		{
			Moniker:   "mod2",
			Component: "go",
			Type:      "go",
			Phases: &ComponentPhases{
				Build: &PhaseInfo{Enabled: true},
				Test:  &PhaseInfo{Enabled: true},
			},
		},
	}

	summary := buildComponentSummary(components)

	if summary.Total != 3 {
		t.Errorf("Total: expected 3, got %d", summary.Total)
	}
	if summary.ByModule["mod1"] != 2 {
		t.Errorf("ByModule[mod1]: expected 2, got %d", summary.ByModule["mod1"])
	}
	if summary.ByModule["mod2"] != 1 {
		t.Errorf("ByModule[mod2]: expected 1, got %d", summary.ByModule["mod2"])
	}
	if summary.ByType["go"] != 2 {
		t.Errorf("ByType[go]: expected 2, got %d", summary.ByType["go"])
	}
	if summary.Buildable != 2 {
		t.Errorf("Buildable: expected 2, got %d", summary.Buildable)
	}
	if summary.Lintable != 2 {
		t.Errorf("Lintable: expected 2, got %d", summary.Lintable)
	}
	if summary.Testable != 2 {
		t.Errorf("Testable: expected 2, got %d", summary.Testable)
	}
	if summary.Scannable != 1 {
		t.Errorf("Scannable: expected 1, got %d", summary.Scannable)
	}
}

package testview

import (
	"os"
	"path/filepath"
	"sort"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	coreoutput "github.com/ready-to-release/eac/go/core/output"
)

// LoadModuleTestView reads all test UoW manifests for a module
// and constructs a TestModuleView by aggregating UoW data
// and locating test output files via artifact references.
func LoadModuleTestView(workspaceRoot, module string) (*TestModuleView, error) {
	reader := coreoutput.NewReader(workspaceRoot)
	uows, err := reader.ListUoWs(core.ActionTest, module)
	if err != nil {
		return nil, err
	}
	if len(uows) == 0 {
		return nil, nil
	}

	view := &TestModuleView{
		Module: module,
		Suites: make(map[string]*SuiteSummary),
	}

	for _, uow := range uows {
		// Aggregate timing
		if view.ExecutedAt.IsZero() || uow.ExecutedAt.Before(view.ExecutedAt) {
			view.ExecutedAt = uow.ExecutedAt
		}
		view.Duration += uow.Duration
		view.UoWCount++
		if uow.ExitCode != 0 {
			view.ExitCode = 1
		}

		// Aggregate tags from each UoW manifest
		if !uow.Tags.IsEmpty() {
			view.Tags = view.Tags.Merge(uow.Tags)
		}

		// Map artifacts to absolute paths
		uowDir := filepath.Join(workspaceRoot, "out", "test", module, uow.DirName())
		for _, art := range uow.Artifacts {
			absPath := filepath.Join(uowDir, art.Path)
			ref := ArtifactRef{
				Type:   art.Type,
				Path:   absPath,
				UoWDir: uow.DirName(),
			}
			view.Artifacts = append(view.Artifacts, ref)

			switch art.Type {
			case ArtifactTypeCucumberReport:
				view.CucumberReports = append(view.CucumberReports, absPath)
			case ArtifactTypeCTRFReport:
				view.CTRFReports = append(view.CTRFReports, absPath)
			case ArtifactTypeCoverage:
				view.CoverageFiles = append(view.CoverageFiles, absPath)
			case ArtifactTypeManualReport:
				// Parse manual test files inline
				entries := parseManualTestFile(absPath, module)
				view.Tests = append(view.Tests, entries...)
			}
		}
	}

	// Parse cucumber reports to extract individual test entries
	for _, cucPath := range view.CucumberReports {
		entries := parseCucumberFile(cucPath, module)
		view.Tests = append(view.Tests, entries...)
	}

	// Parse CTRF reports to extract individual test entries
	for _, ctrfPath := range view.CTRFReports {
		entries := parseCTRFFile(ctrfPath, module)
		view.Tests = append(view.Tests, entries...)
	}

	// Compute summary from test entries
	view.computeSummary()

	// Build suite summaries from test entries
	for i := range view.Tests {
		test := &view.Tests[i]
		suite, exists := view.Suites[test.Suite]
		if !exists {
			suite = &SuiteSummary{}
			view.Suites[test.Suite] = suite
		}
		suite.Total++
		switch test.Status {
		case StatusPassed:
			suite.Passed++
		case StatusFailed:
			suite.Failed++
		case StatusSkipped:
			suite.Skipped++
		}
	}

	return view, nil
}

// LoadAllTestViews loads test views for all modules that have test UoW manifests.
func LoadAllTestViews(workspaceRoot string) ([]*TestModuleView, error) {
	testDir := filepath.Join(workspaceRoot, "out", "test")
	entries, err := os.ReadDir(testDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var views []*TestModuleView
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		view, err := LoadModuleTestView(workspaceRoot, entry.Name())
		if err != nil || view == nil {
			continue
		}
		views = append(views, view)
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].Module < views[j].Module
	})

	return views, nil
}

// LoadTestViewsForModules loads test views for specific modules,
// including transitive dependencies.
func LoadTestViewsForModules(workspaceRoot string, moduleNames []string) ([]*TestModuleView, error) {
	expanded, err := expandModulesWithDependencies(workspaceRoot, moduleNames)
	if err != nil {
		expanded = moduleNames // fallback
	}

	var views []*TestModuleView
	for _, mod := range expanded {
		view, err := LoadModuleTestView(workspaceRoot, mod)
		if err != nil || view == nil {
			continue
		}
		views = append(views, view)
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].Module < views[j].Module
	})

	return views, nil
}

// expandModulesWithDependencies expands a list of module monikers to include
// all their transitive dependencies.
func expandModulesWithDependencies(workspaceRoot string, monikers []string) ([]string, error) {
	registry, err := modules.LoadFromWorkspace(workspaceRoot)
	if err != nil {
		return nil, err
	}

	allModules := make(map[string]bool)
	for _, moniker := range monikers {
		allModules[moniker] = true
		addDepsRecursive(moniker, registry, allModules)
	}

	result := make([]string, 0, len(allModules))
	for moniker := range allModules {
		result = append(result, moniker)
	}

	return result, nil
}

func addDepsRecursive(moniker string, registry *modules.Registry, result map[string]bool) {
	module, exists := registry.Get(moniker)
	if !exists {
		return
	}
	for _, dep := range module.DependsOn {
		if !result[dep] {
			result[dep] = true
			addDepsRecursive(dep, registry, result)
		}
	}
}

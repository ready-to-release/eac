package testdata

import (
	"runtime"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	contractsreports "github.com/ready-to-release/eac/go/core/domain/reports"
	"github.com/ready-to-release/eac/go/core/git"
	"github.com/ready-to-release/eac/go/core/logging"
	"github.com/ready-to-release/eac/go/core/repository"
	"github.com/ready-to-release/eac/go/core/testing"
)

// TestData holds prepared test data with pre-computed aggregations
type TestData struct {
	Tests           []testing.SuiteTestEntry
	TotalCount      int
	OSFilteredCount int    // Tests excluded because they can't run on current OS
	CurrentOS       string // Current OS platform (windows, linux, macos)
	ByType          map[string]int
	ByLevel         map[string]int
	ByModule        map[string]int
}

// GetAllTests discovers all tests and returns enriched test data with aggregations
func GetAllTests(repoRoot string) (*TestData, error) {
	// Load dependencies for discovery options
	moduleReport, _ := contractsreports.GetModuleContracts(repoRoot)
	var moduleRegistry *modules.Registry
	if moduleReport != nil {
		moduleRegistry = moduleReport.Registry
	}

	cfg, _ := config.Load(config.DefaultLoadOptions())
	var environments *config.EnvironmentsConfig
	if cfg != nil {
		environments = cfg.Environments
	}

	// Use unified discovery with full enrichment
	allTests, err := testing.DiscoverAndEnrich(repoRoot, testing.DiscoveryOptions{
		Inferences:     testing.GetGlobalInferences(),
		ModuleRegistry: moduleRegistry,
		Environments:   environments,
	})
	if err != nil {
		return nil, err
	}

	// Calculate OS filtered count using source of truth
	currentOS := mapGOOSToDepTag(runtime.GOOS)
	_, osFilteredCount := testing.FilterByCurrentOS(allTests, currentOS)

	// Build file-module map for conversion
	fileModuleMap, err := BuildFileModuleMap(repoRoot)
	if err != nil {
		fileModuleMap = make(map[string]string)
	}

	// Convert to test entries using canonical conversion
	entries := testing.ConvertToEntries(allTests, fileModuleMap, moduleRegistry, repoRoot)

	// Build aggregations
	byType := make(map[string]int)
	byLevel := make(map[string]int)
	byModule := make(map[string]int)

	for _, entry := range entries {
		byType[entry.Type]++
		for _, level := range entry.Level {
			byLevel[level]++
		}
		if entry.Module != "" {
			byModule[entry.Module]++
		}
	}

	return &TestData{
		Tests:           entries,
		TotalCount:      len(entries),
		OSFilteredCount: osFilteredCount,
		CurrentOS:       currentOS,
		ByType:          byType,
		ByLevel:         byLevel,
		ByModule:        byModule,
	}, nil
}

// mapGOOSToDepTag converts runtime.GOOS to the dependency tag format
func mapGOOSToDepTag(goos string) string {
	switch goos {
	case "darwin":
		return "macos"
	default:
		return goos // linux, windows map directly
	}
}

// BuildFileModuleMap creates a mapping from file paths to module monikers
func BuildFileModuleMap(repoRoot string) (map[string]string, error) {
	// Open git repository
	gitMgr := git.NewManager(logging.C().Zap())
	repo, err := gitMgr.Open(repoRoot)
	if err != nil {
		return nil, err
	}

	files, err := repository.GetRepositoryFilesWithModules(repo, true, false, false)
	if err != nil {
		return nil, err
	}

	fileModuleMap := make(map[string]string)
	for _, file := range files {
		if len(file.Modules) > 0 {
			// Use first module if multiple
			fileModuleMap[file.Name] = file.Modules[0]
		}
	}

	return fileModuleMap, nil
}

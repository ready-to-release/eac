package testdata

import (
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/git"
	"github.com/ready-to-release/eac/go/eac/core/repository"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

// TestData holds prepared test data with pre-computed aggregations
type TestData struct {
	Tests      []testing.SuiteTestEntry
	TotalCount int
	ByType     map[string]int
	ByLevel    map[string]int
	ByModule   map[string]int
}

// GetAllTests discovers all tests and returns enriched test data with aggregations
func GetAllTests(repoRoot string) (*TestData, error) {
	// Discover all tests
	allTests, err := testing.DiscoverAllTests(repoRoot)
	if err != nil {
		return nil, err
	}

	// Load module registry for inference
	moduleReport, err := contractsreports.GetModuleContracts(repoRoot)
	var moduleRegistry *modules.Registry
	if err == nil && moduleReport != nil {
		moduleRegistry = moduleReport.Registry
	}

	// Apply inferences (use global inferences)
	allTests = testing.ApplyInferences(allTests, testing.GetGlobalInferences())

	// Infer system deps from module deps
	if moduleRegistry != nil {
		allTests = testing.InferSystemDepsFromModuleDeps(allTests, moduleRegistry)
	}

	// Infer system deps from environment tags
	cfg, err := config.Load(config.DefaultLoadOptions())
	if err == nil {
		allTests = testing.InferSystemDepsFromEnv(allTests, cfg.Environments)
	}

	// Build file-module map
	fileModuleMap, err := BuildFileModuleMap(repoRoot)
	if err != nil {
		// Non-fatal: use empty map
		fileModuleMap = make(map[string]string)
	}

	// Convert to test entries with all metadata
	entries := ConvertToTestEntries(allTests, fileModuleMap, moduleRegistry, repoRoot)

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
		Tests:      entries,
		TotalCount: len(entries),
		ByType:     byType,
		ByLevel:    byLevel,
		ByModule:   byModule,
	}, nil
}

// BuildFileModuleMap creates a mapping from file paths to module monikers
func BuildFileModuleMap(repoRoot string) (map[string]string, error) {
	// Open git repository
	repo, err := git.Open(repoRoot)
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

// ExtractModuleFromPath looks up the module for a file path using the file-module mapping
func ExtractModuleFromPath(filePath string, fileModuleMap map[string]string, repoRoot string) string {
	if fileModuleMap == nil {
		return ""
	}

	// Normalize separators
	filePath = NormalizePathSeparators(filePath)
	repoRoot = NormalizePathSeparators(repoRoot)

	// Convert absolute path to relative path from repo root
	relativePath := filePath
	if len(filePath) >= len(repoRoot) && filePath[:len(repoRoot)] == repoRoot {
		relativePath = filePath[len(repoRoot):]
		// Trim leading slash
		if len(relativePath) > 0 && relativePath[0] == '/' {
			relativePath = relativePath[1:]
		}
	}

	// For specs/ files, extract module from path structure (specs/MODULE/...)
	if len(relativePath) >= 6 && relativePath[:6] == "specs/" {
		parts := SplitPath(relativePath)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// For go/ files, look up in the file-module map
	if module, found := fileModuleMap[relativePath]; found {
		return module
	}

	// Try direct lookup as fallback
	if module, found := fileModuleMap[filePath]; found {
		return module
	}

	return ""
}

// ConvertToTestEntries converts TestReferences to SuiteTestEntries with metadata
func ConvertToTestEntries(
	tests []testing.TestReference,
	fileModuleMap map[string]string,
	moduleRegistry *modules.Registry,
	repoRoot string,
) []testing.SuiteTestEntry {
	entries := make([]testing.SuiteTestEntry, len(tests))

	for i, test := range tests {
		// Extract module from multiple sources (in priority order):
		// 1. Module dependencies from test tags (@depm:)
		// 2. File path inference
		module := ""
		if len(test.ModuleDependencies) > 0 {
			module = test.ModuleDependencies[0] // Use first module dependency
		}
		if module == "" {
			module = ExtractModuleFromPath(test.FilePath, fileModuleMap, repoRoot)
		}

		// Look up the owning module's type
		moduleType := ""
		if module != "" && moduleRegistry != nil {
			if mod, exists := moduleRegistry.Get(module); exists {
				moduleType = mod.Type
			}
		}

		// Generate test moniker
		moniker := testing.GenerateTestMoniker(test, module)

		// Extract tag categories
		levelTags := FilterTagsByPrefix(test.Tags, "@L")
		verificationTags := FilterTagsByPatterns(test.Tags, []string{"@ov", "@iv", "@pv", "@piv", "@ppv"})
		systemDeps := FilterTagsByPrefix(test.Tags, "@deps:")
		moduleDeps := FilterTagsByPrefix(test.Tags, "@depm:")

		entries[i] = testing.SuiteTestEntry{
			Moniker:          moniker,
			TestName:         test.TestName,
			Type:             test.Type,
			FilePath:         test.FilePath,
			Module:           module,
			ModuleType:       moduleType,
			Level:            levelTags,
			Verification:     verificationTags,
			SystemDeps:       systemDeps,
			ModuleDeps:       moduleDeps,
			IsIgnored:        test.IsIgnored,
			SkipReason:       test.SkipReason,
			IsManual:         test.IsManual,
			RiskControls:     test.RiskControls,
			IsGxP:            test.IsGxP,
			IsCriticalAspect: test.IsCriticalAspect,
		}
	}

	return entries
}

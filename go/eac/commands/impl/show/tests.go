// Command: show tests
// Description: Show all tests in the repository in a human-readable table
package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/commands/internal/render"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	contractsreports "github.com/ready-to-release/eac/go/eac/core/contracts/reports"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	registry.Register(ShowTests)
}

func ShowTests() int {
	// Get repository root
	cwd, err := os.Getwd()
	if err != nil {
		log.Errorf("failed to get current directory: %v", err)
		return 1
	}

	repoRoot, err := findRepoRoot(cwd)
	if err != nil {
		log.Errorf("failed to find repository root: %v", err)
		return 1
	}

	// Discover all tests
	allTests, err := testing.DiscoverAllTests(repoRoot)
	if err != nil {
		log.Errorf("failed to discover tests: %v", err)
		return 1
	}

	// Load module registry for inference
	moduleReport, err := contractsreports.GetModuleContracts(repoRoot)
	var moduleRegistry *modules.Registry
	if err == nil && moduleReport != nil {
		moduleRegistry = moduleReport.Registry
	}

	// Apply inferences
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
	fileModuleMap, err := buildFileModuleMap(repoRoot)
	if err != nil {
		fileModuleMap = make(map[string]string)
	}

	// Convert to test entries with all metadata
	entries := convertToTestEntries(allTests, fileModuleMap, moduleRegistry, repoRoot)

	// Display header
	log.Info("# All Tests\n")
	log.Infof("**Total Tests**: %d  \n", len(entries))

	// Build markdown table
	tb := render.NewTableBuilder().
		WithHeaders("#", "Moniker", "Type", "Module", "Level", "Verification", "System Deps")

	for i, entry := range entries {
		levelStr := strings.Join(entry.Level, ", ")
		verificationStr := strings.Join(entry.Verification, ", ")
		systemDepsStr := strings.Join(entry.SystemDeps, ", ")

		tb.AddRow(
			fmt.Sprintf("%d", i+1),
			entry.Moniker,
			entry.Type,
			entry.Module,
			levelStr,
			verificationStr,
			systemDepsStr,
		)
	}

	log.Info(tb.Build())
	log.Info("")

	// Display summary
	log.Info("## Summary\n")

	// Count by type
	typeCounts := make(map[string]int)
	for _, entry := range entries {
		typeCounts[entry.Type]++
	}
	log.Info("### By Type\n")
	for testType, count := range typeCounts {
		log.Infof("- **%s**: %d tests", testType, count)
	}
	log.Info("")

	// Count by level
	levelCounts := make(map[string]int)
	for _, entry := range entries {
		for _, level := range entry.Level {
			levelCounts[level]++
		}
	}
	log.Info("### By Level\n")
	for _, level := range []string{"@L0", "@L1", "@L2", "@L3", "@L4"} {
		if count, ok := levelCounts[level]; ok {
			log.Infof("- **%s**: %d tests", level, count)
		}
	}
	log.Info("")

	// Count by module
	moduleCounts := make(map[string]int)
	for _, entry := range entries {
		if entry.Module != "" {
			moduleCounts[entry.Module]++
		}
	}
	log.Info("### By Module\n")
	for module, count := range moduleCounts {
		log.Infof("- **%s**: %d tests", module, count)
	}
	log.Info("")

	return 0
}

// findRepoRoot finds the repository root by looking for .git directory
func findRepoRoot(startPath string) (string, error) {
	current := startPath
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("not in a git repository")
		}
		current = parent
	}
}

// convertToTestEntries converts test references to structured test entries
func convertToTestEntries(
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
			module = extractModuleFromPath(test.FilePath, fileModuleMap, repoRoot)
			if module == "unknown" {
				module = ""
			}
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
		levelTags := filterTagsByPrefix(test.Tags, "@L")
		verificationTags := filterTagsByPatterns(test.Tags, []string{"@ov", "@iv", "@pv", "@piv", "@ppv"})
		systemDeps := filterTagsByPrefix(test.Tags, "@deps:")
		moduleDeps := filterTagsByPrefix(test.Tags, "@depm:")

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

// Helper function to extract module from file path
func extractModuleFromPath(filePath string, fileModuleMap map[string]string, repoRoot string) string {
	// Normalize separators and convert absolute to relative path
	filePath = normalizePathSeparators(filePath)
	repoRoot = normalizePathSeparators(repoRoot)

	relativePath := filePath
	if len(filePath) >= len(repoRoot) && filePath[:len(repoRoot)] == repoRoot {
		relativePath = filePath[len(repoRoot):]
		if len(relativePath) > 0 && relativePath[0] == '/' {
			relativePath = relativePath[1:]
		}
	}

	// For specs/ files, extract from path structure
	if len(relativePath) >= 6 && relativePath[:6] == "specs/" {
		parts := splitPath(relativePath)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// For src/ files, look up in map
	if module, found := fileModuleMap[relativePath]; found {
		return module
	}

	return "unknown"
}

func normalizePathSeparators(path string) string {
	return filepath.ToSlash(path)
}

func splitPath(path string) []string {
	return filepath.SplitList(strings.ReplaceAll(path, "/", string(filepath.ListSeparator)))
}

func filterTagsByPrefix(tags []string, prefix string) []string {
	result := []string{}
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			result = append(result, tag)
		}
	}
	return result
}

func filterTagsByPatterns(tags []string, patterns []string) []string {
	result := []string{}
	for _, tag := range tags {
		for _, pattern := range patterns {
			if tag == pattern {
				result = append(result, tag)
				break
			}
		}
	}
	return result
}

package reporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ready-to-release/eac/go/cli/eac/impl/test/internal/cucumber"
	sharedTemplate "github.com/ready-to-release/eac/go/clibase/template"
)

// CollectModuleReports finds and parses all cucumber.json and .log files in the test run directory
// and organizes them by module.
func CollectModuleReports(testRunDir string) ([]ModuleReportData, error) {
	moduleMap := make(map[string]*ModuleReportData)

	// Walk through the test run directory to find all .cucumber.json and .log files
	err := filepath.Walk(testRunDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Handle cucumber.json files (BDD tests)
		if strings.HasSuffix(info.Name(), ".cucumber.json") {
			return processeCucumberReport(path, testRunDir, moduleMap)
		}

		// Handle .log files (Go unit tests), but skip test-suite.log
		if strings.HasSuffix(info.Name(), ".log") && info.Name() != "test-suite.log" {
			return processGoTestLog(path, testRunDir, moduleMap)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	modules := make([]ModuleReportData, 0, len(moduleMap))
	for _, module := range moduleMap {
		// Calculate totals for this module
		totalScenarios := 0
		totalUnitTests := 0
		for _, feature := range module.Features {
			totalScenarios += feature.TestCount
		}
		for _, unitTest := range module.UnitTests {
			totalUnitTests += unitTest.TestCount
		}
		module.TotalScenarios = totalScenarios
		module.TotalUnitTests = totalUnitTests

		modules = append(modules, *module)
	}

	// Sort modules by name
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].ModuleName < modules[j].ModuleName
	})

	return modules, nil
}

// processeCucumberReport processes a single cucumber.json report file.
func processeCucumberReport(path, testRunDir string, moduleMap map[string]*ModuleReportData) error {
	// Parse the cucumber report
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	var report cucumber.CucumberReport
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("failed to parse %s: %w", path, err)
	}

	// Process each feature in the report
	for i := range report {
		feature := &report[i]
		moduleName := feature.GetModule()
		if moduleName == "" {
			// Try to extract module from the file path
			moduleName = extractModuleFromFeaturePath(feature.URI)
		}
		if moduleName == "" {
			moduleName = "unknown"
		}

		// Initialize module data if not exists
		if _, exists := moduleMap[moduleName]; !exists {
			// Get the directory containing this cucumber report for linking
			reportDir, err := filepath.Rel(testRunDir, filepath.Dir(path))
			if err != nil {
				reportDir = filepath.Dir(path) // Fallback to absolute path
			}
			moduleMap[moduleName] = &ModuleReportData{
				ModuleName:    moduleName,
				Features:      []FeatureReportData{},
				UnitTests:     []UnitTestReportData{},
				TestResultDir: reportDir,
			}
		}

		// Count scenarios and their status
		testCount := len(feature.Elements)
		passedCount := 0
		for j := range feature.Elements {
			scenario := &feature.Elements[j]
			if scenario.GetStatus() == "passed" {
				passedCount++
			}
		}

		// Determine overall feature status
		status := "🟢 Passed"
		if passedCount < testCount {
			status = "🔴 Failed"
		}

		// Extract feature tags
		tags := ""
		if len(feature.Tags) > 0 {
			tagList := make([]string, len(feature.Tags))
			for i, tag := range feature.Tags {
				tagList[i] = tag.Name
			}
			tags = strings.Join(tagList, " ")
		}

		// Add feature data
		featureInfo := FeatureReportData{
			Name:          feature.Name,
			URI:           feature.URI,
			NormalizedURI: sharedTemplate.NormalizeSpecPath(feature.URI),
			Tags:          tags,
			TestCount:     testCount,
			PassedCount:   passedCount,
			Status:        status,
		}
		moduleMap[moduleName].Features = append(moduleMap[moduleName].Features, featureInfo)
	}
	return nil
}

// processGoTestLog processes a single Go test log file.
func processGoTestLog(path, testRunDir string, moduleMap map[string]*ModuleReportData) error {
	// Extract module and package from file path
	// Path format: testRunDir/module-name/package-name.log
	relPath, err := filepath.Rel(testRunDir, path)
	if err != nil {
		return nil // Can't determine module from path
	}
	pathParts := strings.Split(filepath.ToSlash(relPath), "/")

	if len(pathParts) < 2 {
		return nil // Skip if not in module subdirectory
	}

	moduleName := pathParts[0]
	packageName := strings.TrimSuffix(pathParts[1], ".log")

	// Convert package file name back to readable format
	// "impl-specs.log" -> "impl/specs"
	packageName = strings.ReplaceAll(packageName, "-", "/")

	// Initialize module data if not exists
	if _, exists := moduleMap[moduleName]; !exists {
		moduleMap[moduleName] = &ModuleReportData{
			ModuleName:    moduleName,
			Features:      []FeatureReportData{},
			UnitTests:     []UnitTestReportData{},
			TestResultDir: moduleName,
		}
	}

	// Parse the test log
	testCount, status, err := parseGoTestLog(path)
	if err != nil {
		// Skip logs that can't be parsed
		return nil
	}

	// Only include if there are actual tests
	if testCount > 0 {
		unitTest := UnitTestReportData{
			Package:   packageName,
			TestCount: testCount,
			Status:    status,
		}
		moduleMap[moduleName].UnitTests = append(moduleMap[moduleName].UnitTests, unitTest)
	}

	return nil
}

// parseGoTestLog parses a Go test log file to extract test count and status.
func parseGoTestLog(logPath string) (int, string, error) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return 0, "", err
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Count top-level tests (lines starting with "=== RUN   Test" at column 0)
	testCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "=== RUN   Test") {
			testCount++
		}
	}

	// Check final status (last non-empty line)
	status := "🟢 Passed"
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "FAIL") {
			status = "🔴 Failed"
		}
		break
	}

	return testCount, status, nil
}

// extractModuleFromFeaturePath extracts module name from feature file path
// Example: "specs/eac-cli/templates/specification.feature" -> "eac-cli".
func extractModuleFromFeaturePath(featurePath string) string {
	// Normalize to forward slashes
	normalized := filepath.ToSlash(featurePath)

	// Remove leading "../" if present
	for strings.HasPrefix(normalized, "../") {
		normalized = strings.TrimPrefix(normalized, "../")
	}

	// Look for "specs/eac-*" or "specs/clie-*" pattern
	if strings.HasPrefix(normalized, "specs/eac-") || strings.HasPrefix(normalized, "specs/clie-") {
		parts := strings.Split(strings.TrimPrefix(normalized, "specs/"), "/")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	// Look for "/go/eac/" or "/go/clie/" pattern
	for _, boundary := range []string{"/go/eac/", "/go/clie/"} {
		if strings.Contains(normalized, boundary) {
			idx := strings.Index(normalized, boundary)
			relativePath := normalized[idx+len(boundary):]
			parts := strings.Split(relativePath, "/")
			if len(parts) >= 1 && parts[0] != "" {
				prefix := "eac"
				if boundary == "/go/clie/" {
					prefix = "clie"
				}
				return prefix + "-" + parts[0]
			}
		}
	}

	return ""
}

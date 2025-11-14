package testing

import (
	"github.com/ready-to-release/eac/src/core/contracts/modules"
)

// SuiteTestEntry represents a single test in a suite report
type SuiteTestEntry struct {
	TestName         string   `yaml:"test_name" json:"test_name" toml:"test_name"`
	Type             string   `yaml:"type" json:"type" toml:"type"`
	FilePath         string   `yaml:"file_path" json:"file_path" toml:"file_path"`
	Module           string   `yaml:"module" json:"module" toml:"module"`
	Level            []string `yaml:"level" json:"level" toml:"level"`
	Verification     []string `yaml:"verification" json:"verification" toml:"verification"`
	SystemDeps       []string `yaml:"system_deps" json:"system_deps" toml:"system_deps"`
	ModuleDeps       []string `yaml:"module_deps" json:"module_deps" toml:"module_deps"`
	ModuleTypes      []string `yaml:"module_types" json:"module_types" toml:"module_types"`
	IsIgnored        bool     `yaml:"is_ignored" json:"is_ignored" toml:"is_ignored"`
	SkipReason       string   `yaml:"skip_reason,omitempty" json:"skip_reason,omitempty" toml:"skip_reason,omitempty"`
	IsManual         bool     `yaml:"is_manual" json:"is_manual" toml:"is_manual"`
	RiskControls     []string `yaml:"risk_controls,omitempty" json:"risk_controls,omitempty" toml:"risk_controls,omitempty"`
	IsGxP            bool     `yaml:"is_gxp" json:"is_gxp" toml:"is_gxp"`
	IsCriticalAspect bool     `yaml:"is_critical_aspect" json:"is_critical_aspect" toml:"is_critical_aspect"`
}

// SuiteReport represents a complete test suite report
type SuiteReport struct {
	SuiteMoniker      string              `yaml:"suite_moniker" json:"suite_moniker" toml:"suite_moniker"`
	SuiteName         string              `yaml:"suite_name" json:"suite_name" toml:"suite_name"`
	Description       string              `yaml:"description" json:"description" toml:"description"`
	ProductionTests   []SuiteTestEntry    `yaml:"production_tests" json:"production_tests" toml:"production_tests"`
	FrameworkTests    []SuiteTestEntry    `yaml:"framework_tests" json:"framework_tests" toml:"framework_tests"`
	TotalDiscovered   int                 `yaml:"total_discovered" json:"total_discovered" toml:"total_discovered"`
	Selectors         []TagSelector       `yaml:"selectors" json:"selectors" toml:"selectors"`
	ValidationErrors  map[string][]string `yaml:"validation_errors,omitempty" json:"validation_errors,omitempty" toml:"validation_errors,omitempty"`
}

// GenerateSuiteReport generates a complete test suite report with all metadata
// This is the canonical data generator used by both `get suite` and `show suite` commands
func GenerateSuiteReport(
	suite *TestSuite,
	repoRoot string,
	moduleRegistry *modules.Registry,
	fileModuleMap map[string]string,
) (*SuiteReport, error) {
	// Phase 1: Discover all tests
	allTests, err := DiscoverAllTests(repoRoot)
	if err != nil {
		return nil, err
	}

	// Phase 2: Apply inferences
	allTests = ApplyInferences(allTests, suite.Inferences)

	// Phase 2.5: Infer system deps from module deps (if registry available)
	if moduleRegistry != nil {
		allTests = InferSystemDepsFromModuleDeps(allTests, moduleRegistry)
	}

	// Phase 3: Select tests for this suite
	selectedTests := suite.SelectTests(allTests)

	// Phase 4: Separate production tests from framework tests
	productionTests := []TestReference{}
	frameworkTests := []TestReference{}
	for _, test := range selectedTests {
		if ShouldSkipValidation(test) {
			frameworkTests = append(frameworkTests, test)
		} else {
			productionTests = append(productionTests, test)
		}
	}

	// Phase 5: Validate post-inference tags
	validationErrors := ValidateAllPostInference(productionTests, repoRoot)

	// Convert test references to suite entries
	productionEntries := convertToSuiteEntries(productionTests, fileModuleMap, moduleRegistry, repoRoot)
	frameworkEntries := convertToSuiteEntries(frameworkTests, fileModuleMap, moduleRegistry, repoRoot)

	report := &SuiteReport{
		SuiteMoniker:     suite.Moniker,
		SuiteName:        suite.Name,
		Description:      suite.Description,
		ProductionTests:  productionEntries,
		FrameworkTests:   frameworkEntries,
		TotalDiscovered:  len(allTests),
		Selectors:        suite.Selectors,
		ValidationErrors: validationErrors,
	}

	return report, nil
}

// convertToSuiteEntries converts TestReferences to SuiteTestEntries with metadata
func convertToSuiteEntries(
	tests []TestReference,
	fileModuleMap map[string]string,
	moduleRegistry *modules.Registry,
	repoRoot string,
) []SuiteTestEntry {
	entries := make([]SuiteTestEntry, len(tests))

	for i, test := range tests {
		// Extract module from file path
		module := ""
		if fileModuleMap != nil {
			if m, exists := fileModuleMap[test.FilePath]; exists {
				module = m
			}
		}

		// Extract tag categories
		levelTags := filterTagsByPrefix(test.Tags, "@L")
		verificationTags := filterTagsByPatterns(test.Tags, []string{"@ov", "@iv", "@pv", "@piv", "@ppv"})
		systemDeps := filterTagsByPrefix(test.Tags, "@deps:")
		moduleDeps := filterTagsByPrefix(test.Tags, "@depm:")

		// Look up module types
		moduleTypes := []string{}
		if moduleRegistry != nil {
			for _, depTag := range moduleDeps {
				moniker := trimPrefix(depTag, "@depm:")
				if mod, exists := moduleRegistry.Get(moniker); exists {
					moduleTypes = append(moduleTypes, mod.Type)
				}
			}
		}

		entries[i] = SuiteTestEntry{
			TestName:         test.TestName,
			Type:             test.Type,
			FilePath:         test.FilePath,
			Module:           module,
			Level:            levelTags,
			Verification:     verificationTags,
			SystemDeps:       systemDeps,
			ModuleDeps:       moduleDeps,
			ModuleTypes:      moduleTypes,
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

// Helper functions for tag filtering
func filterTagsByPrefix(tags []string, prefix string) []string {
	result := []string{}
	for _, tag := range tags {
		if len(tag) >= len(prefix) && tag[:len(prefix)] == prefix {
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

func trimPrefix(s, prefix string) string {
	if len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}

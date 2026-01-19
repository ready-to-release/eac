package testing

import (
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// SuiteTestEntry represents a single test (assertion) in a suite report.
type SuiteTestEntry struct {
	Moniker      string   `yaml:"moniker" json:"moniker" toml:"moniker"`
	TestName     string   `yaml:"test_name" json:"test_name" toml:"test_name"`
	Type         string   `yaml:"type" json:"type" toml:"type"`
	FilePath     string   `yaml:"file_path" json:"file_path" toml:"file_path"`
	Package      string   `yaml:"package" json:"package" toml:"package"` // Feature/package name from path
	Module       string   `yaml:"module" json:"module" toml:"module"`
	ModuleType   string   `yaml:"module_type" json:"module_type" toml:"module_type"`
	Level        []string `yaml:"level" json:"level" toml:"level"`
	Verification []string `yaml:"verification" json:"verification" toml:"verification"`
	SystemDeps   []string `yaml:"system_deps" json:"system_deps" toml:"system_deps"`
	ModuleDeps   []string `yaml:"module_deps" json:"module_deps" toml:"module_deps"`

	// Tag provenance - tracking what was inferred vs explicit
	SourceTags   []string `yaml:"source_tags" json:"source_tags" toml:"source_tags"`       // Original tags from source
	InferredTags []string `yaml:"inferred_tags" json:"inferred_tags" toml:"inferred_tags"` // Tags added by inference
	InferredDeps []string `yaml:"inferred_deps" json:"inferred_deps" toml:"inferred_deps"` // Deps inferred from module type
	InferredDepm []string `yaml:"inferred_depm" json:"inferred_depm" toml:"inferred_depm"` // Module deps inferred from path

	IsIgnored        bool     `yaml:"is_ignored" json:"is_ignored" toml:"is_ignored"`
	SkipReason       string   `yaml:"skip_reason,omitempty" json:"skip_reason,omitempty" toml:"skip_reason,omitempty"`
	IsManual         bool     `yaml:"is_manual" json:"is_manual" toml:"is_manual"`
	IsSequential     bool     `yaml:"is_sequential" json:"is_sequential" toml:"is_sequential"`
	RiskControls     []string `yaml:"risk_controls,omitempty" json:"risk_controls,omitempty" toml:"risk_controls,omitempty"`
	IsGxP            bool     `yaml:"is_gxp" json:"is_gxp" toml:"is_gxp"`
	IsCriticalAspect bool     `yaml:"is_critical_aspect" json:"is_critical_aspect" toml:"is_critical_aspect"`
}

// SuiteReport represents a complete test suite report.
type SuiteReport struct {
	SuiteMoniker     string              `yaml:"suite_moniker" json:"suite_moniker" toml:"suite_moniker"`
	SuiteName        string              `yaml:"suite_name" json:"suite_name" toml:"suite_name"`
	Description      string              `yaml:"description" json:"description" toml:"description"`
	ProductionTests  []SuiteTestEntry    `yaml:"production_tests" json:"production_tests" toml:"production_tests"`
	FrameworkTests   []SuiteTestEntry    `yaml:"framework_tests" json:"framework_tests" toml:"framework_tests"`
	TotalDiscovered  int                 `yaml:"total_discovered" json:"total_discovered" toml:"total_discovered"`
	Selectors        []TagSelector       `yaml:"selectors" json:"selectors" toml:"selectors"`
	ValidationErrors map[string][]string `yaml:"validation_errors,omitempty" json:"validation_errors,omitempty" toml:"validation_errors,omitempty"`
}

// GenerateSuiteReport generates a complete test suite report with all metadata
// This is the canonical data generator used by both `get suite` and `show suite` commands.
func GenerateSuiteReport(
	suite *TestSuite,
	repoRoot string,
	moduleRegistry *modules.Registry,
	fileModuleMap map[string]string,
) (*SuiteReport, error) {
	// Load environments for inference
	cfg, _ := config.Load(config.DefaultLoadOptions())
	var environments *config.EnvironmentsConfig
	if cfg != nil {
		environments = cfg.Environments
	}

	// Use unified discovery with suite-specific inferences
	allTests, err := DiscoverAndEnrich(repoRoot, DiscoveryOptions{
		Inferences:     suite.Inferences,
		ModuleRegistry: moduleRegistry,
		Environments:   environments,
	})
	if err != nil {
		return nil, err
	}

	// Select tests for this suite
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

	// Convert test references to suite entries using canonical conversion
	productionEntries := ConvertToEntries(productionTests, fileModuleMap, moduleRegistry, repoRoot)
	frameworkEntries := ConvertToEntries(frameworkTests, fileModuleMap, moduleRegistry, repoRoot)

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

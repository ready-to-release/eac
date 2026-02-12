package paths

import (
	"path/filepath"
	"testing"
)

// TestDefaultPathConfigMatchesConstants verifies that DefaultPathConfig() values
// match the hardcoded constants used throughout the paths package.
func TestDefaultPathConfigMatchesConstants(t *testing.T) {
	pc := DefaultPathConfig()

	checks := []struct {
		field    string
		got      string
		expected string
	}{
		{"SpecsRoot", pc.SpecsRoot, SpecsDir},
		{"Templates", pc.Templates, TemplatesDir},
		{"OutBuild", pc.OutBuild, OutBuildRelPath},
		{"OutTest", pc.OutTest, OutTestRelPath},
		{"OutLogs", pc.OutLogs, OutLogsRelPath},
		{"OutScan", pc.OutScan, OutSecurityRelPath},
		{"OutTools", pc.OutTools, OutDir + "/" + ToolsDir},
		{"BuildLog", pc.BuildLog, "build.log"},
		{"BuildTiming", pc.BuildTiming, "build-timing.txt"},
		{"TestTiming", pc.TestTiming, "test-timing.txt"},
		{"Specification", pc.Specification, "specification.feature"},
		{"RiskCatalog", pc.RiskCatalog, "controls.catalog.json"},
		{"RiskControlsDir", pc.RiskControlsDir, RiskControlsDir},
		{"TemplateSpecsDir", pc.TemplateSpecsDir, "specs"},
		{"TemplateReportsDir", pc.TemplateReportsDir, "reports"},
		{"TemplateRiskCatalogDir", pc.TemplateRiskCatalogDir, "risk-catalog"},
		{"GodogTest", pc.GodogTest, "godog_test.go"},
	}

	for _, c := range checks {
		t.Run(c.field, func(t *testing.T) {
			if c.got != c.expected {
				t.Errorf("DefaultPathConfig().%s = %q, want %q", c.field, c.got, c.expected)
			}
		})
	}
}

// TestCVariantsMatchHardcoded verifies that C-variant functions with DefaultPathConfig
// produce identical output to the existing hardcoded path functions.
func TestCVariantsMatchHardcoded(t *testing.T) {
	repoRoot := "/repo"
	moniker := "test-module"
	pc := DefaultPathConfig()

	tests := []struct {
		name     string
		cResult  string
		hResult  string
	}{
		{
			name:    "BuildOutputPath",
			cResult: BuildOutputPathC(repoRoot, moniker, pc),
			hResult: BuildOutputPath(repoRoot, moniker),
		},
		{
			name:    "BuildLogPath",
			cResult: BuildLogPathC(repoRoot, moniker, pc),
			hResult: BuildLogPath(repoRoot, moniker),
		},
		{
			name:    "BuildTimingPath",
			cResult: BuildTimingPathC(repoRoot, moniker, pc),
			hResult: BuildTimingPath(repoRoot, moniker),
		},
		{
			name:    "TestModuleDir",
			cResult: TestModuleDirC(repoRoot, moniker, pc),
			hResult: TestModuleDir(repoRoot, moniker),
		},
		{
			name:    "TestModuleTimingPath",
			cResult: TestModuleTimingPathC(repoRoot, moniker, pc),
			hResult: TestModuleTimingPath(repoRoot, moniker),
		},
		{
			name:    "ScanModuleOutputPath",
			cResult: ScanModuleOutputPathC(repoRoot, moniker, pc),
			hResult: SecurityOutputPath(repoRoot, moniker),
		},
		{
			name:    "RiskControlsPath",
			cResult: RiskControlsPathC(repoRoot, pc),
			hResult: RiskControlsPath(repoRoot),
		},
		{
			name:    "RiskCatalogPath",
			cResult: RiskCatalogPathC(repoRoot, pc),
			hResult: RiskCatalogPath(repoRoot),
		},
		{
			name:    "SpecsPath",
			cResult: SpecsPathC(repoRoot, moniker, pc),
			hResult: SpecsPath(repoRoot, moniker),
		},
		{
			name:    "SpecsFeaturePath_WithModule",
			cResult: SpecsFeaturePathC(repoRoot, moniker, "my-feature", pc),
			hResult: SpecsFeaturePath(repoRoot, moniker, "my-feature"),
		},
		{
			name:    "SpecsFeaturePath_NoModule",
			cResult: SpecsFeaturePathC(repoRoot, "", "top-feature", pc),
			hResult: SpecsFeaturePath(repoRoot, "", "top-feature"),
		},
		{
			name:    "TemplatePath_NoSegments",
			cResult: TemplatePathC(repoRoot, pc),
			hResult: TemplatePath(repoRoot),
		},
		{
			name:    "TemplatePath_WithSegments",
			cResult: TemplatePathC(repoRoot, pc, "reports", "release"),
			hResult: TemplatePath(repoRoot, "reports", "release"),
		},
		{
			name:    "TemplateSpecsPath_NoSubpaths",
			cResult: TemplateSpecsPathC(repoRoot, pc),
			hResult: TemplateSpecsPath(repoRoot),
		},
		{
			name:    "TemplateSpecsPath_WithSubpath",
			cResult: TemplateSpecsPathC(repoRoot, pc, "risk-catalog"),
			hResult: TemplateSpecsPath(repoRoot, "risk-catalog"),
		},
		{
			name:    "TemplateReportsPath_NoSubpaths",
			cResult: TemplateReportsPathC(repoRoot, pc),
			hResult: TemplateReportsPath(repoRoot),
		},
		{
			name:    "TemplateReportsPath_WithSubpath",
			cResult: TemplateReportsPathC(repoRoot, pc, "summary.md"),
			hResult: TemplateReportsPath(repoRoot, "summary.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cResult != tt.hResult {
				t.Errorf("%s: C-variant = %q, hardcoded = %q", tt.name, tt.cResult, tt.hResult)
			}
		})
	}
}

// TestCVariantsRelativePaths verifies that C-variant functions produce correct
// relative paths when repoRoot is empty.
func TestCVariantsRelativePaths(t *testing.T) {
	moniker := "core"
	pc := DefaultPathConfig()

	tests := []struct {
		name     string
		result   string
		expected string
	}{
		{
			name:     "SpecsPathC_Relative",
			result:   SpecsPathC("", moniker, pc),
			expected: "specs/core",
		},
		{
			name:     "BuildOutputPathC_Relative",
			result:   BuildOutputPathC("", moniker, pc),
			expected: "out/build/core",
		},
		{
			name:     "BuildLogPathC_Relative",
			result:   BuildLogPathC("", moniker, pc),
			expected: "out/build/core/build.log",
		},
		{
			name:     "BuildTimingPathC_Relative",
			result:   BuildTimingPathC("", moniker, pc),
			expected: "out/build/core/build-timing.txt",
		},
		{
			name:     "TestModuleDirC_Relative",
			result:   TestModuleDirC("", moniker, pc),
			expected: "out/test/core",
		},
		{
			name:     "TestModuleTimingPathC_Relative",
			result:   TestModuleTimingPathC("", moniker, pc),
			expected: "out/test/core/test-timing.txt",
		},
		{
			name:     "ScanModuleOutputPathC_Relative",
			result:   ScanModuleOutputPathC("", moniker, pc),
			expected: "out/scan/core",
		},
		{
			name:     "RiskControlsPathC_Relative",
			result:   RiskControlsPathC("", pc),
			expected: "specs/.risk-controls",
		},
		{
			name:     "SpecsFeaturePathC_Relative_WithModule",
			result:   SpecsFeaturePathC("", moniker, "my-feature", pc),
			expected: "specs/core/my-feature/specification.feature",
		},
		{
			name:     "SpecsFeaturePathC_Relative_NoModule",
			result:   SpecsFeaturePathC("", "", "top-feature", pc),
			expected: "specs/top-feature/specification.feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.result, tt.expected)
			}
		})
	}
}

// TestCVariantsCustomConfig verifies that C-variant functions respect non-default
// PathConfig values (proving configurability works).
func TestCVariantsCustomConfig(t *testing.T) {
	repoRoot := "/repo"
	moniker := "my-mod"
	pc := PathConfig{
		SpecsRoot:              "specifications",
		Templates:              "tmpl",
		OutBuild:               "output/build",
		OutTest:                "output/test",
		OutLogs:                "output/logs",
		OutScan:                "output/security",
		OutTools:               "output/tools",
		BuildLog:               "compile.log",
		BuildTiming:            "compile-timing.txt",
		TestTiming:             "test-duration.txt",
		Specification:          "spec.feature",
		RiskCatalog:            "catalog.json",
		RiskControlsDir:        ".controls",
		TemplateSpecsDir:       "spec-templates",
		TemplateReportsDir:     "report-templates",
		TemplateRiskCatalogDir: "risk",
		GodogTest:              "bdd_test.go",
	}

	tests := []struct {
		name     string
		result   string
		expected string
	}{
		{
			name:     "SpecsPathC",
			result:   SpecsPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "specifications", moniker),
		},
		{
			name:     "BuildOutputPathC",
			result:   BuildOutputPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "build", moniker),
		},
		{
			name:     "BuildLogPathC",
			result:   BuildLogPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "build", moniker, "compile.log"),
		},
		{
			name:     "BuildTimingPathC",
			result:   BuildTimingPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "build", moniker, "compile-timing.txt"),
		},
		{
			name:     "TestModuleDirC",
			result:   TestModuleDirC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "test", moniker),
		},
		{
			name:     "TestModuleTimingPathC",
			result:   TestModuleTimingPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "test", moniker, "test-duration.txt"),
		},
		{
			name:     "ScanModuleOutputPathC",
			result:   ScanModuleOutputPathC(repoRoot, moniker, pc),
			expected: filepath.Join(repoRoot, "output", "security", moniker),
		},
		{
			name:     "RiskControlsPathC",
			result:   RiskControlsPathC(repoRoot, pc),
			expected: filepath.Join(repoRoot, "specifications", ".controls"),
		},
		{
			name:     "RiskCatalogPathC",
			result:   RiskCatalogPathC(repoRoot, pc),
			expected: filepath.Join(repoRoot, "tmpl", "spec-templates", "risk", "catalog.json"),
		},
		{
			name:     "TemplatePathC",
			result:   TemplatePathC(repoRoot, pc, "ai", "prompts"),
			expected: filepath.Join(repoRoot, "tmpl", "ai", "prompts"),
		},
		{
			name:     "TemplateSpecsPathC",
			result:   TemplateSpecsPathC(repoRoot, pc, "my-spec"),
			expected: filepath.Join(repoRoot, "tmpl", "spec-templates", "my-spec"),
		},
		{
			name:     "TemplateReportsPathC",
			result:   TemplateReportsPathC(repoRoot, pc, "summary.md"),
			expected: filepath.Join(repoRoot, "tmpl", "report-templates", "summary.md"),
		},
		{
			name:     "SpecsFeaturePathC_WithModule",
			result:   SpecsFeaturePathC(repoRoot, moniker, "login", pc),
			expected: filepath.Join(repoRoot, "specifications", moniker, "login", "spec.feature"),
		},
		{
			name:     "SpecsFeaturePathC_NoModule",
			result:   SpecsFeaturePathC(repoRoot, "", "login", pc),
			expected: filepath.Join(repoRoot, "specifications", "login", "spec.feature"),
		},
		{
			name:     "IsGodogTestFileC_Match",
			result:   boolToStr(IsGodogTestFileC("bdd_test.go", pc)),
			expected: "true",
		},
		{
			name:     "IsGodogTestFileC_NoMatch",
			result:   boolToStr(IsGodogTestFileC("godog_test.go", pc)),
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.result != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.result, tt.expected)
			}
		})
	}
}

// TestGetPathVariablesC verifies the path variables map.
func TestGetPathVariablesC(t *testing.T) {
	pc := DefaultPathConfig()
	vars := GetPathVariablesC(pc)

	expected := map[string]string{
		"specs_root": "specs",
		"templates":  "templates",
		"out_build":  "out/build",
		"out_test":   "out/test",
		"out_logs":   "out/logs",
		"out_scan":   "out/scan",
		"out_tools":  "out/tools",
	}

	for key, want := range expected {
		if got := vars[key]; got != want {
			t.Errorf("GetPathVariablesC()[%q] = %q, want %q", key, got, want)
		}
	}

	if len(vars) != len(expected) {
		t.Errorf("GetPathVariablesC() returned %d entries, want %d", len(vars), len(expected))
	}
}

func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

package paths

import (
	"path/filepath"
	"testing"
)

// TestNewPathHelpers tests the newly added path helper functions.
func TestNewPathHelpers(t *testing.T) {
	repoRoot := "/repo"
	moniker := "test-module"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "BuildOutputDir",
			fn:       func() string { return BuildOutputDir(repoRoot) },
			expected: filepath.Join(repoRoot, "out", "build"),
		},
		{
			name:     "BuildLogPath",
			fn:       func() string { return BuildLogPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "build", moniker, "build.log"),
		},
		{
			name:     "BuildTimingPath",
			fn:       func() string { return BuildTimingPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "build", moniker, "build-timing.txt"),
		},
		{
			name:     "TestModuleDir",
			fn:       func() string { return TestModuleDir(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "test", moniker),
		},
		{
			name:     "TestModuleTimingPath",
			fn:       func() string { return TestModuleTimingPath(repoRoot, moniker) },
			expected: filepath.Join(repoRoot, "out", "test", moniker, "test-timing.txt"),
		},
		{
			name:     "TestOutputDir",
			fn:       func() string { return TestOutputDir(repoRoot) },
			expected: filepath.Join(repoRoot, "out", "test"),
		},
		{
			name:     "RiskControlsPath",
			fn:       func() string { return RiskControlsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "specs", ".risk-controls"),
		},
		{
			name:     "RiskCatalogPath",
			fn:       func() string { return RiskCatalogPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog", "controls.catalog.json"),
		},
		{
			name:     "TemplateSpecsPath",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog"),
		},
		{
			name:     "TemplateReportsPath",
			fn:       func() string { return TemplateReportsPath(repoRoot, "summary.md") },
			expected: filepath.Join(repoRoot, "templates", "reports", "summary.md"),
		},
		{
			name:     "SpecsFeaturePath_WithModule",
			fn:       func() string { return SpecsFeaturePath(repoRoot, moniker, "my-feature") },
			expected: filepath.Join(repoRoot, "specs", moniker, "my-feature", "specification.feature"),
		},
		{
			name:     "SpecsFeaturePath_NoModule",
			fn:       func() string { return SpecsFeaturePath(repoRoot, "", "top-level-feature") },
			expected: filepath.Join(repoRoot, "specs", "top-level-feature", "specification.feature"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}

// TestPathConstants validates that path constants haven't changed.
func TestPathConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"OutDir", OutDir, "out"},
		{"SpecsDir", SpecsDir, "specs"},
		{"TemplatesDir", TemplatesDir, "templates"},
		{"R2RDir", R2RDir, ".r2r"},
		{"BuildDir", BuildDir, "build"},
		{"TestDir", TestDir, "test"},
		{"LogsDir", LogsDir, "logs"},
		{"RiskControlsDir", RiskControlsDir, ".risk-controls"},
		{"ReleaseDir", ReleaseDir, "release"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, expected %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

// TestTemplatePathVariadic tests the variadic template path functions.
func TestTemplatePathVariadic(t *testing.T) {
	repoRoot := "/repo"

	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{
			name:     "TemplateSpecsPath_NoSubpaths",
			fn:       func() string { return TemplateSpecsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "specs"),
		},
		{
			name:     "TemplateSpecsPath_OneSubpath",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog"),
		},
		{
			name:     "TemplateSpecsPath_MultipleSubpaths",
			fn:       func() string { return TemplateSpecsPath(repoRoot, "risk-catalog", "controls.json") },
			expected: filepath.Join(repoRoot, "templates", "specs", "risk-catalog", "controls.json"),
		},
		{
			name:     "TemplateReportsPath_NoSubpaths",
			fn:       func() string { return TemplateReportsPath(repoRoot) },
			expected: filepath.Join(repoRoot, "templates", "reports"),
		},
		{
			name:     "TemplateReportsPath_OneSubpath",
			fn:       func() string { return TemplateReportsPath(repoRoot, "summary.md") },
			expected: filepath.Join(repoRoot, "templates", "reports", "summary.md"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn()
			if result != tt.expected {
				t.Errorf("%s() = %q, expected %q", tt.name, result, tt.expected)
			}
		})
	}
}

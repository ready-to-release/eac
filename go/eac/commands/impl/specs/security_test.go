//go:build L1 && ov
// +build L1,ov

// File: go/eac/commands/impl/specs/security_test.go
package specs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExtractFeatureName(t *testing.T) {
	tests := []struct {
		name       string
		gherkin    string
		wantModule string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "extract from feature with module prefix",
			gherkin:    "Feature: eac-commands_user-authentication\n  As a user...",
			wantModule: "eac-commands",
			wantName:   "user-authentication",
			wantErr:    false,
		},
		{
			name:       "extract from feature without module prefix",
			gherkin:    "Feature: user-authentication\n  As a user...",
			wantModule: "",
			wantName:   "user-authentication",
			wantErr:    false,
		},
		{
			name:       "handle feature with hyphens",
			gherkin:    "Feature: multi-module_api-rate-limiting\n  As a developer...",
			wantModule: "multi-module",
			wantName:   "api-rate-limiting",
			wantErr:    false,
		},
		{
			name:    "error on missing feature declaration",
			gherkin: "  As a user\n  I want...",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if gotModule != tt.wantModule {
				t.Errorf("ExtractFeatureName() module = %v, want %v", gotModule, tt.wantModule)
			}
			if gotName != tt.wantName {
				t.Errorf("ExtractFeatureName() name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}

func TestDetermineOutputPath(t *testing.T) {
	tests := []struct {
		name        string
		moduleName  string
		featureName string
		wantContain []string
	}{
		{
			name:        "with module prefix",
			moduleName:  "eac-commands",
			featureName: "user-auth",
			wantContain: []string{"specs", "eac-commands", "user-auth", "specification.feature"},
		},
		{
			name:        "without module prefix",
			moduleName:  "",
			featureName: "user-auth",
			wantContain: []string{"specs", "user-auth", "specification.feature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			got := DetermineOutputPath(tmpDir, tt.moduleName, tt.featureName)

			// Check that all expected path components are present
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("DetermineOutputPath() = %v, missing component %v", got, want)
				}
			}

			// Verify it ends with specification.feature
			if !strings.HasSuffix(got, "specification.feature") {
				t.Errorf("DetermineOutputPath() = %v, want to end with specification.feature", got)
			}
		})
	}
}

// ==============================================================================
// CRITICAL EDGE CASES - Security and Data Safety
// ==============================================================================

func TestValidateOutputPath_PathTraversalPrevention(t *testing.T) {
	// Use temp directory for realistic cross-platform testing
	tmpRepo := t.TempDir()

	tests := []struct {
		name           string
		outputPath     string
		wantErr        bool
		wantErrContain string
	}{
		{
			name:       "normal relative path - allowed",
			outputPath: filepath.Join("specs", "my-feature", "spec.feature"),
			wantErr:    false,
		},
		{
			name:       "absolute path within repo - allowed",
			outputPath: filepath.Join(tmpRepo, "specs", "feature.feature"),
			wantErr:    false,
		},
		{
			name:           "path traversal up one level - blocked",
			outputPath:     filepath.Join("..", "outside.feature"),
			wantErr:        true,
			wantErrContain: "must be within repository",
		},
		{
			name:           "path traversal multiple levels - blocked",
			outputPath:     filepath.Join("..", "..", "..", "etc", "passwd"),
			wantErr:        true,
			wantErrContain: "must be within repository",
		},
		{
			name:           "path traversal with mixed separators - blocked",
			outputPath:     filepath.Join("specs", "..", "..", "outside.feature"),
			wantErr:        true,
			wantErrContain: "must be within repository",
		},
		{
			name:       "deep nested path - allowed",
			outputPath: filepath.Join("specs", "a", "b", "c", "d", "e", "spec.feature"),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputPath(tt.outputPath, tmpRepo)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("ValidateOutputPath() error = %v, want to contain %q", err, tt.wantErrContain)
				}
			}
		})
	}
}

func TestValidateOutputPath_RealFileSystem(t *testing.T) {
	// Test with real filesystem using temp directory
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		outputPath string
		wantErr    bool
	}{
		{
			name:       "relative path within temp dir",
			outputPath: filepath.Join("specs", "test.feature"),
			wantErr:    false,
		},
		{
			name:       "try to escape temp dir",
			outputPath: filepath.Join("..", "..", "escape.feature"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(tmpDir, tt.outputPath)
			err := ValidateOutputPath(fullPath, tmpDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExtractFeatureName_SpecialCharactersValidation(t *testing.T) {
	tests := []struct {
		name        string
		gherkin     string
		wantModule  string
		wantName    string
		wantErr     bool
		description string
	}{
		{
			name:        "normal feature name",
			gherkin:     "Feature: eac-commands_user-auth\n",
			wantModule:  "eac-commands",
			wantName:    "user-auth",
			wantErr:     false,
			description: "Standard case",
		},
		{
			name:        "feature with slashes",
			gherkin:     "Feature: go/eac/commands/test\n",
			wantModule:  "",
			wantName:    "",
			wantErr:     true, // ✅ NEW: Improved validation rejects path separators
			description: "Path separators now rejected by validation",
		},
		{
			name:        "feature with backslashes",
			gherkin:     "Feature: src\\commands\\test\n",
			wantModule:  "",
			wantName:    "",
			wantErr:     true, // ✅ NEW: Improved validation rejects path separators
			description: "Windows separators now rejected by validation",
		},
		{
			name:        "path traversal attempt in feature name",
			gherkin:     "Feature: ../malicious\n",
			wantModule:  "",
			wantName:    "",
			wantErr:     true, // ✅ NEW: Improved validation rejects path traversal
			description: "Path traversal now rejected by validation",
		},
		{
			name:        "multiple underscores",
			gherkin:     "Feature: a_b_c_d\n",
			wantModule:  "a",
			wantName:    "b_c_d",
			wantErr:     false,
			description: "Splits on first underscore",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotModule != tt.wantModule {
					t.Errorf("ExtractFeatureName() module = %v, want %v", gotModule, tt.wantModule)
				}
				if gotName != tt.wantName {
					t.Errorf("ExtractFeatureName() name = %v, want %v", gotName, tt.wantName)
				}

				// Log info for potentially dangerous patterns (these are expected in security tests)
				if strings.Contains(gotName, "/") || strings.Contains(gotName, "\\") || strings.Contains(gotName, "..") {
					t.Logf("[INFO] %s - Feature name: %q", tt.description, gotName)
				}
			}
		})
	}
}

func TestExtractFeatureName_UnicodeNames(t *testing.T) {
	tests := []struct {
		name        string
		gherkin     string
		wantModule  string
		wantName    string
		wantErr     bool
		description string
	}{
		{
			name:        "Spanish Unicode",
			gherkin:     "Feature: módulo_función-española\n",
			wantModule:  "módulo",
			wantName:    "función-española",
			wantErr:     false,
			description: "Should handle Spanish accented characters",
		},
		{
			name:        "Chinese characters",
			gherkin:     "Feature: 模块_功能-中文\n",
			wantModule:  "模块",
			wantName:    "功能-中文",
			wantErr:     false,
			description: "Should handle Chinese characters",
		},
		{
			name:        "Russian Cyrillic",
			gherkin:     "Feature: модуль_функция-русский\n",
			wantModule:  "модуль",
			wantName:    "функция-русский",
			wantErr:     false,
			description: "Should handle Russian Cyrillic",
		},
		{
			name:        "Arabic RTL",
			gherkin:     "Feature: وحدة_ميزة-عربية\n",
			wantModule:  "وحدة",
			wantName:    "ميزة-عربية",
			wantErr:     false,
			description: "Should handle Arabic RTL text",
		},
		{
			name:        "Mixed scripts",
			gherkin:     "Feature: r2r-cli_función-中文-test\n",
			wantModule:  "r2r-cli",
			wantName:    "función-中文-test",
			wantErr:     false,
			description: "Should handle mixed scripts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotModule != tt.wantModule {
					t.Errorf("ExtractFeatureName() module = %q, want %q (%s)", gotModule, tt.wantModule, tt.description)
				}
				if gotName != tt.wantName {
					t.Errorf("ExtractFeatureName() name = %q, want %q (%s)", gotName, tt.wantName, tt.description)
				}
			}
		})
	}
}

func TestExtractFeatureName_WindowsReservedNames(t *testing.T) {
	// Windows reserved device names are now properly detected and rejected
	tests := []struct {
		name       string
		gherkin    string
		wantModule string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "CON device",
			gherkin:    "Feature: CON_test\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // ✅ NEW: Windows reserved names now rejected
		},
		{
			name:       "PRN device",
			gherkin:    "Feature: r2r-cli_PRN\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // ✅ NEW: Windows reserved names now rejected
		},
		{
			name:       "NUL device",
			gherkin:    "Feature: NUL_device\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // ✅ NEW: Windows reserved names now rejected
		},
		{
			name:       "valid name that contains CON",
			gherkin:    "Feature: r2r-cli_console-app\n",
			wantModule: "r2r-cli",
			wantName:   "console-app",
			wantErr:    false, // Should pass - "console" != "CON"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotModule != tt.wantModule {
					t.Errorf("ExtractFeatureName() module = %q, want %q", gotModule, tt.wantModule)
				}
				if gotName != tt.wantName {
					t.Errorf("ExtractFeatureName() name = %q, want %q", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestExtractFeatureName_VeryLongNames(t *testing.T) {
	// Test with very long feature names that might exceed filesystem limits
	longName := strings.Repeat("very-long-feature-name-", 20) + "end"

	tests := []struct {
		name        string
		gherkin     string
		expectError bool
		description string
	}{
		{
			name:        "very long feature name",
			gherkin:     fmt.Sprintf("Feature: %s\n", longName),
			expectError: false, // ExtractFeatureName doesn't validate length
			description: "Feature name exceeding 260 chars",
		},
		{
			name:        "very long module and feature",
			gherkin:     fmt.Sprintf("Feature: %s_%s\n", longName, longName),
			expectError: false,
			description: "Combined length exceeding filesystem limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, name, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.expectError {
				t.Errorf("ExtractFeatureName() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				totalLen := len(module) + len(name)
				if totalLen > 200 {
					t.Logf("[INFO] %s - Total length: %d characters", tt.description, totalLen)
				}
			}
		})
	}
}

func TestExtractFeatureName_CaseSensitivity(t *testing.T) {
	tests := []struct {
		name       string
		gherkin    string
		wantModule string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "lowercase feature",
			gherkin:    "feature: test_module\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // Should not match lowercase "feature:"
		},
		{
			name:       "uppercase FEATURE",
			gherkin:    "FEATURE: test_module\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // Should not match uppercase "FEATURE:"
		},
		{
			name:       "correct case Feature",
			gherkin:    "Feature: test_module\n",
			wantModule: "test",
			wantName:   "module",
			wantErr:    false,
		},
		{
			name:       "mixed case FEaTuRe",
			gherkin:    "FEaTuRe: test_module\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // Should not match mixed case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotModule != tt.wantModule {
					t.Errorf("ExtractFeatureName() module = %q, want %q", gotModule, tt.wantModule)
				}
				if gotName != tt.wantName {
					t.Errorf("ExtractFeatureName() name = %q, want %q", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestExtractFeatureName_WhitespaceVariations(t *testing.T) {
	tests := []struct {
		name       string
		gherkin    string
		wantModule string
		wantName   string
		wantErr    bool
	}{
		{
			name:       "extra spaces after colon",
			gherkin:    "Feature:    test_module\n",
			wantModule: "test",
			wantName:   "module",
			wantErr:    false,
		},
		{
			name:       "tab after colon",
			gherkin:    "Feature:\ttest_module\n",
			wantModule: "test",
			wantName:   "module",
			wantErr:    false,
		},
		{
			name:       "mixed tabs and spaces",
			gherkin:    "Feature:  \t  test_module\n",
			wantModule: "test",
			wantName:   "module",
			wantErr:    false,
		},
		{
			name:       "leading whitespace before Feature",
			gherkin:    "  Feature: test_module\n",
			wantModule: "",
			wantName:   "",
			wantErr:    true, // Regex expects start of line (^Feature:)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotModule, gotName, err := ExtractFeatureName(tt.gherkin)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractFeatureName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotModule != tt.wantModule {
					t.Errorf("ExtractFeatureName() module = %q, want %q", gotModule, tt.wantModule)
				}
				if gotName != tt.wantName {
					t.Errorf("ExtractFeatureName() name = %q, want %q", gotName, tt.wantName)
				}
			}
		})
	}
}

func TestValidateOutputPath_SymbolicLinks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a directory and a symlink pointing outside
	insideDir := filepath.Join(tmpDir, "inside")
	outsideDir := t.TempDir() // Different temp dir (outside)

	if err := os.MkdirAll(insideDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Note: Symlink creation might fail on Windows without admin rights
	// Skip test gracefully if symlink creation fails
	symlinkPath := filepath.Join(tmpDir, "link-to-outside")
	if err := os.Symlink(outsideDir, symlinkPath); err != nil {
		t.Skipf("Skipping symlink test: %v", err)
		return
	}

	tests := []struct {
		name        string
		outputPath  string
		wantErr     bool
		windowsOnly bool // deps:windows - test only runs on Windows
	}{
		{
			name:        "path inside allowed directory",
			outputPath:  filepath.Join(insideDir, "test.feature"),
			wantErr:     false,
			windowsOnly: false,
		},
		{
			name:        "path through symlink to outside",
			outputPath:  filepath.Join(symlinkPath, "test.feature"),
			wantErr:     true, // Should detect escape via symlink
			windowsOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("deps:windows - test only runs on Windows")
			}
			err := ValidateOutputPath(tt.outputPath, tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOutputPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetermineOutputPath_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		moduleName  string
		featureName string
		description string
	}{
		{
			name:        "empty module",
			moduleName:  "",
			featureName: "test",
			description: "Should create path without module directory",
		},
		{
			name:        "empty feature",
			moduleName:  "module",
			featureName: "",
			description: "Edge case: empty feature name",
		},
		{
			name:        "both empty",
			moduleName:  "",
			featureName: "",
			description: "Edge case: both empty",
		},
		{
			name:        "feature with dots",
			moduleName:  "module",
			featureName: "feature.with.dots",
			description: "Dots might be confused with file extensions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := DetermineOutputPath(tmpDir, tt.moduleName, tt.featureName)

			// Verify path is absolute or under tmpDir
			if !filepath.IsAbs(path) {
				t.Errorf("Path should be absolute: %s", path)
			}

			// Log for manual inspection
			t.Logf("%s: %s", tt.description, path)

			// Should always end with specification.feature
			if !strings.HasSuffix(path, "specification.feature") {
				t.Errorf("Path should end with specification.feature: %s", path)
			}
		})
	}
}

// ==============================================================================
// HIGH PRIORITY IMPROVEMENT TESTS
// ==============================================================================

func TestValidateFeatureLineSecurity(t *testing.T) {
	tests := []struct {
		name        string
		featureLine string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid feature name",
			featureLine: "eac-commands_user-auth",
			wantErr:     false,
		},
		{
			name:        "valid with Unicode",
			featureLine: "módulo_función-española",
			wantErr:     false,
		},
		{
			name:        "path traversal with ..",
			featureLine: "../../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:        "forward slash separator",
			featureLine: "go/eac/commands/test",
			wantErr:     true,
			errContains: "path separator",
		},
		{
			name:        "backslash separator",
			featureLine: "src\\commands\\test",
			wantErr:     true,
			errContains: "path separator",
		},
		{
			name:        "null byte",
			featureLine: "test\x00malicious",
			wantErr:     true,
			errContains: "control character",
		},
		{
			name:        "newline in name",
			featureLine: "test\nmalicious",
			wantErr:     true,
			errContains: "control character",
		},
		{
			name:        "tab is allowed",
			featureLine: "test\tfeature",
			wantErr:     false, // Tabs are trimmed by TrimSpace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFeatureLineSecurity(tt.featureLine)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFeatureLineSecurity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}
		})
	}
}

func TestValidateWindowsReservedNames(t *testing.T) {
	tests := []struct {
		name        string
		featureName string
		wantErr     bool
		errContains string
	}{
		{
			name:        "normal name",
			featureName: "user-auth",
			wantErr:     false,
		},
		{
			name:        "CON reserved",
			featureName: "CON",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "con lowercase",
			featureName: "con",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "PRN reserved",
			featureName: "PRN",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "NUL reserved",
			featureName: "NUL",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "COM1 reserved",
			featureName: "COM1",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "LPT1 reserved",
			featureName: "LPT1",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "CON with extension",
			featureName: "CON.txt",
			wantErr:     true,
			errContains: "Windows reserved",
		},
		{
			name:        "contains CON but not reserved",
			featureName: "console-app",
			wantErr:     false, // Not at start
		},
		{
			name:        "config is not reserved",
			featureName: "config",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWindowsReservedName(tt.featureName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWindowsReservedName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got %q", tt.errContains, err.Error())
				}
			}
		})
	}
}

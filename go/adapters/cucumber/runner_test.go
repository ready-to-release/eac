//go:build L1
// +build L1

package cucumber

import (
	"testing"
)

func TestParseCucumberPackagePath_ThreeParts(t *testing.T) {
	pkg := parseCucumberPackagePath("login:src/ts:specs/module/login/spec.feature")
	if pkg.displayName != "login:src/ts" {
		t.Errorf("displayName = %q, want %q", pkg.displayName, "login:src/ts")
	}
	if pkg.relPkgPath != "src/ts" {
		t.Errorf("relPkgPath = %q, want %q", pkg.relPkgPath, "src/ts")
	}
	if pkg.relFeatureFile != "specs/module/login/spec.feature" {
		t.Errorf("relFeatureFile = %q, want %q", pkg.relFeatureFile, "specs/module/login/spec.feature")
	}
}

func TestParseCucumberPackagePath_OnePart(t *testing.T) {
	pkg := parseCucumberPackagePath("src/ts")
	if pkg.displayName != "src/ts" {
		t.Errorf("displayName = %q, want %q", pkg.displayName, "src/ts")
	}
	if pkg.relPkgPath != "src/ts" {
		t.Errorf("relPkgPath = %q, want %q", pkg.relPkgPath, "src/ts")
	}
	if pkg.relFeatureFile != "" {
		t.Errorf("relFeatureFile = %q, want empty", pkg.relFeatureFile)
	}
}

func TestParseCucumberPackagePath_Fallback(t *testing.T) {
	pkg := parseCucumberPackagePath("a:b")
	if pkg.displayName != "a:b" {
		t.Errorf("displayName = %q, want %q", pkg.displayName, "a:b")
	}
	if pkg.relPkgPath != "a:b" {
		t.Errorf("relPkgPath = %q, want %q", pkg.relPkgPath, "a:b")
	}
}

func TestExtractTsFeatureFolderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard feature path",
			input:    "specs/my-module/login/specification.feature",
			expected: "login",
		},
		{
			name:     "nested path",
			input:    "specs/module/sub/feature-name/spec.feature",
			expected: "feature-name",
		},
		{
			name:     "single directory",
			input:    "features/spec.feature",
			expected: "features",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractTsFeatureFolderName(tt.input)
			if result != tt.expected {
				t.Errorf("extractTsFeatureFolderName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTsCucumberRunner_TestTypes(t *testing.T) {
	r := &TsCucumberRunner{}
	types := r.TestTypes()
	if len(types) != 1 || types[0] != "tscucumber" {
		t.Errorf("TestTypes() = %v, want [tscucumber]", types)
	}
}

func TestTsCucumberRunner_IsBDD(t *testing.T) {
	r := &TsCucumberRunner{}
	if !r.IsBDD() {
		t.Error("IsBDD() = false, want true")
	}
}

func TestTsCucumberRunner_BuildPackagePath(t *testing.T) {
	r := &TsCucumberRunner{}

	tests := []struct {
		name        string
		testRoot    string
		featurePath string
		expected    string
	}{
		{
			name:     "empty test root",
			testRoot: "",
			expected: "",
		},
		{
			name:     "test root only",
			testRoot: "src/ts",
			expected: "src/ts",
		},
		{
			name:        "with feature path",
			testRoot:    "src/ts",
			featurePath: "specs/module/login/specification.feature",
			expected:    "login:src/ts:specs/module/login/specification.feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.BuildPackagePath(tt.testRoot, tt.featurePath)
			if result != tt.expected {
				t.Errorf("BuildPackagePath(%q, %q) = %q, want %q", tt.testRoot, tt.featurePath, result, tt.expected)
			}
		})
	}
}

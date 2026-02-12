//go:build L1
// +build L1

package gotest

import (
	"testing"

	"github.com/ready-to-release/eac/go/clibase/testrunners"
)

func TestGoTestRunner_TestTypes(t *testing.T) {
	r := &GoTestRunner{}
	types := r.TestTypes()
	if len(types) != 2 {
		t.Fatalf("TestTypes() returned %d items, want 2", len(types))
	}
	if types[0] != "gotest" {
		t.Errorf("TestTypes()[0] = %q, want %q", types[0], "gotest")
	}
	if types[1] != "godog" {
		t.Errorf("TestTypes()[1] = %q, want %q", types[1], "godog")
	}
}

func TestGoTestRunner_IsBDD(t *testing.T) {
	r := &GoTestRunner{}
	if !r.IsBDD() {
		t.Error("IsBDD() = false, want true")
	}
}

func TestGoTestRunner_BuildPackagePath(t *testing.T) {
	r := &GoTestRunner{}
	tests := []struct {
		name        string
		testRoot    string
		featurePath string
		expected    string
	}{
		{
			name:     "unit test path",
			testRoot: "go/core/config",
			expected: "go/core/config",
		},
		{
			name:        "BDD test path",
			testRoot:    "go/specs/repository",
			featurePath: "specs/repository/login/specification.feature",
			expected:    "login:go/specs/repository:specs/repository/login/specification.feature",
		},
		{
			name:        "empty test root with feature",
			testRoot:    "",
			featurePath: "specs/module/feature.feature",
			expected:    "",
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

func TestParseGoTestPackagePath(t *testing.T) {
	tests := []struct {
		name               string
		input              string
		wantDisplayName    string
		wantRelPkgPath     string
		wantRelFeatureFile string
	}{
		{
			name:            "unit test format",
			input:           "go/core/config",
			wantDisplayName: "go/core/config",
			wantRelPkgPath:  "go/core/config",
		},
		{
			name:               "BDD format",
			input:              "login:go/specs/repository:specs/repository/login/spec.feature",
			wantDisplayName:    "login:go/specs/repository",
			wantRelPkgPath:     "go/specs/repository",
			wantRelFeatureFile: "specs/repository/login/spec.feature",
		},
		{
			name:            "fallback two parts",
			input:           "a:b",
			wantDisplayName: "a:b",
			wantRelPkgPath:  "a:b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := parseGoTestPackagePath(tt.input)
			if pkg.displayName != tt.wantDisplayName {
				t.Errorf("displayName = %q, want %q", pkg.displayName, tt.wantDisplayName)
			}
			if pkg.relPkgPath != tt.wantRelPkgPath {
				t.Errorf("relPkgPath = %q, want %q", pkg.relPkgPath, tt.wantRelPkgPath)
			}
			if pkg.relFeatureFile != tt.wantRelFeatureFile {
				t.Errorf("relFeatureFile = %q, want %q", pkg.relFeatureFile, tt.wantRelFeatureFile)
			}
		})
	}
}

func TestExtractFeatureFolderName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard spec path",
			input:    "specs/repository/no-build-tags-in-steps/specification.feature",
			expected: "no-build-tags-in-steps",
		},
		{
			name:     "deeply nested",
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
			result := extractFeatureFolderName(tt.input)
			if result != tt.expected {
				t.Errorf("extractFeatureFolderName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildGoTestArgs(t *testing.T) {
	tests := []struct {
		name           string
		suiteTagFilter string
		coverage       bool
		outputDir      string
		wantContains   []string
		wantAbsent     []string
	}{
		{
			name:           "basic no tags no coverage",
			suiteTagFilter: "",
			wantContains:   []string{"test", "-json", "-v", "."},
			wantAbsent:     []string{"-tags", "-cover"},
		},
		{
			name:           "with L-level tags",
			suiteTagFilter: "@L0,@L1",
			wantContains:   []string{"-tags", "L0,L1"},
		},
		{
			name:           "with deps tags",
			suiteTagFilter: "@deps:docker",
			wantContains:   []string{"-tags", "deps_docker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testrunners.RunConfig{
				Coverage:  tt.coverage,
				OutputDir: tt.outputDir,
			}
			args := buildGoTestArgs(tt.suiteTagFilter, cfg)

			for _, want := range tt.wantContains {
				found := false
				for _, a := range args {
					if a == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("args %v missing %q", args, want)
				}
			}

			for _, absent := range tt.wantAbsent {
				for _, a := range args {
					if a == absent {
						t.Errorf("args %v should not contain %q", args, absent)
					}
				}
			}
		})
	}
}

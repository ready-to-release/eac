//go:build L1
// +build L1

package pytest

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
)

func TestPytestRunner_TestTypes(t *testing.T) {
	r := &PytestRunner{}
	types := r.TestTypes()
	if len(types) != 1 || types[0] != "pytest" {
		t.Errorf("TestTypes() = %v, want [pytest]", types)
	}
}

func TestPytestRunner_IsBDD(t *testing.T) {
	r := &PytestRunner{}
	if r.IsBDD() {
		t.Error("IsBDD() = true, want false")
	}
}

func TestPytestRunner_FindTestRoot(t *testing.T) {
	r := &PytestRunner{}
	// FindTestRoot always returns empty for pytest
	result := r.FindTestRoot("some/path", nil)
	if result != "" {
		t.Errorf("FindTestRoot() = %q, want empty", result)
	}
}

func TestPytestRunner_BuildPackagePath(t *testing.T) {
	r := &PytestRunner{}
	result := r.BuildPackagePath("src/python", "test_file.py")
	if result != "src/python" {
		t.Errorf("BuildPackagePath() = %q, want %q", result, "src/python")
	}
}

func TestFindPythonModuleForPath(t *testing.T) {
	cfg := &config.EACConfig{
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "my-module",
					Components: config.ModuleComponents{
						"python": &config.ComponentEntry{Root: "src/python/my-module"},
					},
				},
				{
					Moniker: "other-module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{Root: "go/other"},
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		relPath  string
		expected string
	}{
		{
			name:     "matching path",
			relPath:  "src/python/my-module/tests",
			expected: "my-module",
		},
		{
			name:     "exact root path",
			relPath:  "src/python/my-module",
			expected: "my-module",
		},
		{
			name:     "non-matching path",
			relPath:  "src/java/something",
			expected: "",
		},
		{
			name:     "partial match should not work",
			relPath:  "src/python/my-module-extra",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findPythonModuleForPath(tt.relPath, cfg)
			if result != tt.expected {
				t.Errorf("findPythonModuleForPath(%q) = %q, want %q", tt.relPath, result, tt.expected)
			}
		})
	}
}

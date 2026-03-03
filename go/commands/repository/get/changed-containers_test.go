package get

import (
	"context"
	"fmt"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func newTestModuleContract(components config.ModuleComponents) *modules.ModuleContract {
	base := domain.BaseContract{
		Moniker:    "oci-tools",
		Components: components,
	}
	return modules.NewModuleContract(base, "")
}

// TestFilterFilesForComponentOwnership tests per-component file filtering.
func TestFilterFilesForComponentOwnership(t *testing.T) {
	mod := newTestModuleContract(config.ModuleComponents{
		"drawio-oci": &config.ComponentEntry{
			Root:     "containers/drawio-oci",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
		"pdf-oci": &config.ComponentEntry{
			Root:     "containers/pdf-oci",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
		"mkdocs-render-oci": &config.ComponentEntry{
			Root:     "containers/mkdocs-render-oci",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
	})

	files := []string{
		"containers/drawio-oci/Dockerfile",
		"containers/drawio-oci/scripts/build.sh",
		"containers/pdf-oci/Dockerfile",
		"containers/mkdocs-render-oci/requirements.txt",
		"go/cli/eac/main.go", // not owned by any container
	}

	tests := []struct {
		name      string
		component string
		expected  int
	}{
		{"drawio gets 2 files", "drawio-oci", 2},
		{"pdf gets 1 file", "pdf-oci", 1},
		{"mkdocs gets 1 file", "mkdocs-render-oci", 1},
		{"unknown gets 0 files", "nonexistent", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterFilesForComponentOwnership(files, mod, tt.component)
			if len(got) != tt.expected {
				t.Errorf("filterFilesForComponentOwnership(%q) returned %d files, expected %d: %v",
					tt.component, len(got), tt.expected, got)
			}
		})
	}
}

// TestDetectSharedFileChanges tests detection of files owned by module but not by any component.
func TestDetectSharedFileChanges(t *testing.T) {
	mod := newTestModuleContract(config.ModuleComponents{
		"drawio-oci": &config.ComponentEntry{
			Root:     "containers/drawio-oci",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
		"pdf-oci": &config.ComponentEntry{
			Root:     "containers/pdf-oci",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
		"shared-scripts": &config.ComponentEntry{
			Root:     "containers/shared",
			Patterns: &config.ComponentPatterns{Source: []string{"**/*"}},
		},
	})

	components := []string{"drawio-oci", "pdf-oci"}

	tests := []struct {
		name     string
		files    []string
		expected int
	}{
		{
			name:     "no shared files",
			files:    []string{"containers/drawio-oci/Dockerfile", "containers/pdf-oci/Dockerfile"},
			expected: 0,
		},
		{
			name:     "shared files detected",
			files:    []string{"containers/drawio-oci/Dockerfile", "containers/shared/build-base.sh"},
			expected: 1,
		},
		{
			name:     "unowned files are not shared",
			files:    []string{"go/cli/eac/main.go", "containers/drawio-oci/Dockerfile"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectSharedFileChanges(tt.files, mod, components)
			if len(got) != tt.expected {
				t.Errorf("detectSharedFileChanges returned %d shared files, expected %d: %v",
					len(got), tt.expected, got)
			}
		})
	}
}

// TestBuildFilteredMatrix tests JSON matrix generation.
func TestBuildFilteredMatrix(t *testing.T) {
	tests := []struct {
		name     string
		changed  []ContainerComponentStatus
		expected string
	}{
		{
			name:     "empty",
			changed:  []ContainerComponentStatus{},
			expected: `[]`,
		},
		{
			name: "single component",
			changed: []ContainerComponentStatus{
				{Name: "drawio-oci", Reason: "no_previous_build"},
			},
			expected: `[{"name":"drawio-oci"}]`,
		},
		{
			name: "multiple components",
			changed: []ContainerComponentStatus{
				{Name: "drawio-oci", Reason: "files_changed"},
				{Name: "pdf-oci", Reason: "no_previous_build"},
			},
			expected: `[{"name":"drawio-oci"},{"name":"pdf-oci"}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildFilteredMatrix(tt.changed)
			if got != tt.expected {
				t.Errorf("buildFilteredMatrix() = %q, expected %q", got, tt.expected)
			}
		})
	}
}

// TestBuildForceAllResult tests the force-all path.
func TestBuildForceAllResult(t *testing.T) {
	components := []string{"drawio-oci", "pdf-oci", "mermaid-oci"}
	result := buildForceAllResult("oci-tools", "abc1234", components)

	if result.Module != "oci-tools" {
		t.Errorf("Module = %q, expected %q", result.Module, "oci-tools")
	}
	if result.HeadSHA != "abc1234" {
		t.Errorf("HeadSHA = %q, expected %q", result.HeadSHA, "abc1234")
	}
	if len(result.ChangedComponents) != 3 {
		t.Fatalf("ChangedComponents count = %d, expected 3", len(result.ChangedComponents))
	}
	for _, c := range result.ChangedComponents {
		if c.Reason != "force_all" {
			t.Errorf("Component %q reason = %q, expected force_all", c.Name, c.Reason)
		}
	}
	if len(result.SkippedComponents) != 0 {
		t.Errorf("SkippedComponents count = %d, expected 0", len(result.SkippedComponents))
	}
	if result.FilteredMatrix == "" {
		t.Error("FilteredMatrix should not be empty")
	}
}

// TestResolveComponentNames tests JSON parsing of component names.
func TestResolveComponentNames(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		expected []string
		wantErr  bool
	}{
		{
			name:     "valid JSON array",
			json:     `[{"name":"drawio-oci"},{"name":"pdf-oci"}]`,
			expected: []string{"drawio-oci", "pdf-oci"},
		},
		{
			name:     "single component",
			json:     `[{"name":"mermaid-oci"}]`,
			expected: []string{"mermaid-oci"},
		},
		{
			name:    "invalid JSON",
			json:    `not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveComponentNames(tt.json, "oci-tools")
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(got) != len(tt.expected) {
				t.Fatalf("Got %d components, expected %d: %v", len(got), len(tt.expected), got)
			}
			for i, name := range got {
				if name != tt.expected[i] {
					t.Errorf("Component[%d] = %q, expected %q", i, name, tt.expected[i])
				}
			}
		})
	}
}

// TestMockContainerRegistryQuerier tests the mock implementation.
func TestMockContainerRegistryQuerier(t *testing.T) {
	mock := &mockContainerRegistryQuerier{
		results: map[string]string{
			"drawio-oci": "abc1234",
			"pdf-oci":    "",
		},
		errors: map[string]error{
			"mermaid-oci": fmt.Errorf("rate limited"),
		},
	}

	t.Run("returns SHA", func(t *testing.T) {
		sha, err := mock.LastBuildSHA(context.Background(), "drawio-oci")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if sha != "abc1234" {
			t.Errorf("SHA = %q, expected %q", sha, "abc1234")
		}
	})

	t.Run("returns empty for no tag", func(t *testing.T) {
		sha, err := mock.LastBuildSHA(context.Background(), "pdf-oci")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if sha != "" {
			t.Errorf("SHA = %q, expected empty", sha)
		}
	})

	t.Run("returns error", func(t *testing.T) {
		_, err := mock.LastBuildSHA(context.Background(), "mermaid-oci")
		if err == nil {
			t.Error("Expected error but got nil")
		}
	})

	t.Run("returns empty for unknown", func(t *testing.T) {
		sha, err := mock.LastBuildSHA(context.Background(), "unknown-oci")
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if sha != "" {
			t.Errorf("SHA = %q, expected empty", sha)
		}
	})
}

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverContainerModules(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create containers directory with various container structures
	containersDir := filepath.Join(tmpDir, "containers")
	os.MkdirAll(containersDir, 0755)

	// Create some -oci containers (should be discovered)
	createContainerDir(t, containersDir, "test-oci")
	createContainerDir(t, containersDir, "another-oci")

	// Create a non-oci container (should NOT be discovered)
	createContainerDir(t, containersDir, "internal-tool")

	// Create a directory without Dockerfile (should NOT be discovered)
	os.MkdirAll(filepath.Join(containersDir, "no-dockerfile-oci"), 0755)

	tests := []struct {
		name             string
		explicitMonikers map[string]bool
		wantCount        int
		wantMonikers     []string
	}{
		{
			name:             "discovers all OCI containers",
			explicitMonikers: map[string]bool{},
			wantCount:        2,
			wantMonikers:     []string{"another-oci", "test-oci"},
		},
		{
			name:             "skips explicitly defined containers",
			explicitMonikers: map[string]bool{"test-oci": true},
			wantCount:        1,
			wantMonikers:     []string{"another-oci"},
		},
		{
			name:             "returns empty when all containers are explicit",
			explicitMonikers: map[string]bool{"test-oci": true, "another-oci": true},
			wantCount:        0,
			wantMonikers:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modules := DiscoverContainerModules(tmpDir, tt.explicitMonikers, "containers")

			if len(modules) != tt.wantCount {
				t.Errorf("got %d modules, want %d", len(modules), tt.wantCount)
			}

			for i, m := range modules {
				if i < len(tt.wantMonikers) && m.Moniker != tt.wantMonikers[i] {
					t.Errorf("module[%d].Moniker = %q, want %q", i, m.Moniker, tt.wantMonikers[i])
				}

				// Verify template is set
				if m.Template != "container" {
					t.Errorf("module[%d].Template = %q, want %q", i, m.Template, "container")
				}

				// Verify name is derived
				if m.Name == "" {
					t.Errorf("module[%d].Name should not be empty", i)
				}
			}
		})
	}
}

func TestDiscoverContainerModules_NoContainersDir(t *testing.T) {
	tmpDir := t.TempDir()
	// Don't create containers directory

	modules := DiscoverContainerModules(tmpDir, map[string]bool{}, "containers")
	if len(modules) != 0 {
		t.Errorf("expected 0 modules when containers dir doesn't exist, got %d", len(modules))
	}
}

func TestDeriveContainerName(t *testing.T) {
	tests := []struct {
		moniker string
		want    string
	}{
		{"pdf-oci", "PDF Container"},
		{"nginx-oci", "Nginx Container"},
		{"mkdocs-dev-oci", "Mkdocs Dev Container"},
		{"pdf-cli-oci", "PDF CLI Container"},
		{"api-gateway-oci", "API Gateway Container"},
	}

	for _, tt := range tests {
		t.Run(tt.moniker, func(t *testing.T) {
			got := deriveContainerName(tt.moniker)
			if got != tt.want {
				t.Errorf("deriveContainerName(%q) = %q, want %q", tt.moniker, got, tt.want)
			}
		})
	}
}

// createContainerDir creates a container directory with a Dockerfile
func createContainerDir(t *testing.T, containersDir, name string) {
	t.Helper()
	dir := filepath.Join(containersDir, name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	dockerfile := filepath.Join(dir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine:latest\n"), 0644); err != nil {
		t.Fatal(err)
	}
}

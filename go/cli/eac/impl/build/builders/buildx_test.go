package builders

import (
	"io"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func TestGetDockerBuildConfig_NamedComponent(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "oci-tools",
			Components: domain.ModuleComponents{
				"pdf-oci": &domain.ComponentEntry{
					DockerBuild: map[string]interface{}{
						"container":  "pdf-oci",
						"context":    "containers/pdf-oci",
						"dockerfile": "containers/pdf-oci/Dockerfile",
					},
				},
				"mkdocs-render-oci": &domain.ComponentEntry{
					DockerBuild: map[string]interface{}{
						"container":  "mkdocs-render-oci",
						"context":    "containers/mkdocs-render-oci",
						"dockerfile": "containers/mkdocs-render-oci/Dockerfile",
					},
				},
			},
		},
	}

	// Should find config for named component "pdf-oci"
	cfg := getDockerBuildConfig(module, "pdf-oci", io.Discard)
	if cfg == nil {
		t.Fatal("expected non-nil config for pdf-oci")
	}
	if cfg.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg.Container)
	}

	// Should find config for named component "mkdocs-render-oci"
	cfg2 := getDockerBuildConfig(module, "mkdocs-render-oci", io.Discard)
	if cfg2 == nil {
		t.Fatal("expected non-nil config for mkdocs-render-oci")
	}
	if cfg2.Container != "mkdocs-render-oci" {
		t.Errorf("expected container=mkdocs-render-oci, got %s", cfg2.Container)
	}

	// Should return nil for non-existent component
	cfg3 := getDockerBuildConfig(module, "nonexistent", io.Discard)
	if cfg3 != nil {
		t.Error("expected nil config for nonexistent component")
	}
}

func TestGetDockerBuildConfig_LegacyDockerfileKey(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "pdf-oci",
			Components: domain.ModuleComponents{
				"dockerfile": &domain.ComponentEntry{
					DockerBuild: map[string]interface{}{
						"container": "pdf-oci",
						"context":   "containers/pdf-oci",
					},
				},
			},
		},
	}

	// With component name "dockerfile", finds via first lookup (named component)
	cfg := getDockerBuildConfig(module, "dockerfile", io.Discard)
	if cfg == nil {
		t.Fatal("expected non-nil config for dockerfile component")
	}
	if cfg.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg.Container)
	}

	// With empty component name, falls back to "dockerfile" key
	cfg2 := getDockerBuildConfig(module, "", io.Discard)
	if cfg2 == nil {
		t.Fatal("expected non-nil config via fallback to 'dockerfile' key")
	}
	if cfg2.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg2.Container)
	}
}

func TestGetDockerBuildConfig_EmptyDockerBuild(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "test-module",
			Components: domain.ModuleComponents{
				"pdf-oci": &domain.ComponentEntry{
					DockerBuild: map[string]interface{}{}, // empty
				},
			},
		},
	}

	// Empty DockerBuild should return nil
	cfg := getDockerBuildConfig(module, "pdf-oci", io.Discard)
	if cfg != nil {
		t.Error("expected nil config for empty DockerBuild map")
	}
}

func TestGetDockerBuildConfig_NilComponent(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "test-module",
			Components: domain.ModuleComponents{
				"pdf-oci": nil, // nil entry
			},
		},
	}

	cfg := getDockerBuildConfig(module, "pdf-oci", io.Discard)
	if cfg != nil {
		t.Error("expected nil config for nil component entry")
	}
}

func TestGetDockerBuildConfig_NamedComponentFallsBackToDockerfile(t *testing.T) {
	// Module has a named component WITHOUT docker_build, but has a "dockerfile" with docker_build.
	// Named lookup should fail, then fall back to "dockerfile" key.
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "mixed-module",
			Components: domain.ModuleComponents{
				"my-comp": &domain.ComponentEntry{
					// No DockerBuild here
				},
				"dockerfile": &domain.ComponentEntry{
					DockerBuild: map[string]interface{}{
						"container": "fallback-container",
						"context":   "containers/fallback",
					},
				},
			},
		},
	}

	// Looking up "my-comp" should fall back to "dockerfile" key
	cfg := getDockerBuildConfig(module, "my-comp", io.Discard)
	if cfg == nil {
		t.Fatal("expected fallback to 'dockerfile' key")
	}
	if cfg.Container != "fallback-container" {
		t.Errorf("expected container=fallback-container, got %s", cfg.Container)
	}
}

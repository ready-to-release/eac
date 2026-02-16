package builders

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func TestGetDockerBuildConfig_NamedComponent(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "oci-tools",
			Components: config.ModuleComponents{
				"pdf-oci": &config.ComponentEntry{
					DockerBuild: &config.DockerBuildConfig{
						Container:  "pdf-oci",
						Context:    "containers/pdf-oci",
						Dockerfile: "containers/pdf-oci/Dockerfile",
					},
				},
				"mkdocs-render-oci": &config.ComponentEntry{
					DockerBuild: &config.DockerBuildConfig{
						Container:  "mkdocs-render-oci",
						Context:    "containers/mkdocs-render-oci",
						Dockerfile: "containers/mkdocs-render-oci/Dockerfile",
					},
				},
			},
		},
	}

	// Should find config for named component "pdf-oci"
	cfg := getDockerBuildConfig(module, "pdf-oci")
	if cfg == nil {
		t.Fatal("expected non-nil config for pdf-oci")
	}
	if cfg.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg.Container)
	}

	// Should find config for named component "mkdocs-render-oci"
	cfg2 := getDockerBuildConfig(module, "mkdocs-render-oci")
	if cfg2 == nil {
		t.Fatal("expected non-nil config for mkdocs-render-oci")
	}
	if cfg2.Container != "mkdocs-render-oci" {
		t.Errorf("expected container=mkdocs-render-oci, got %s", cfg2.Container)
	}

	// Should return nil for non-existent component
	cfg3 := getDockerBuildConfig(module, "nonexistent")
	if cfg3 != nil {
		t.Error("expected nil config for nonexistent component")
	}
}

func TestGetDockerBuildConfig_ContainerComponent(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "pdf-oci",
			Components: config.ModuleComponents{
				"container": &config.ComponentEntry{
					Type: "container",
					DockerBuild: &config.DockerBuildConfig{
						Container: "pdf-oci",
						Context:   "containers/pdf-oci",
					},
				},
			},
		},
	}

	// With component name "container", finds via first lookup (named component)
	cfg := getDockerBuildConfig(module, "container")
	if cfg == nil {
		t.Fatal("expected non-nil config for container component")
	}
	if cfg.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg.Container)
	}

	// With empty component name, falls back to container type
	cfg2 := getDockerBuildConfig(module, "")
	if cfg2 == nil {
		t.Fatal("expected non-nil config via fallback to container type")
	}
	if cfg2.Container != "pdf-oci" {
		t.Errorf("expected container=pdf-oci, got %s", cfg2.Container)
	}
}

func TestGetDockerBuildConfig_EmptyDockerBuild(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "test-module",
			Components: config.ModuleComponents{
				"pdf-oci": &config.ComponentEntry{
					DockerBuild: nil, // nil DockerBuild
				},
			},
		},
	}

	// nil DockerBuild should return nil
	cfg := getDockerBuildConfig(module, "pdf-oci")
	if cfg != nil {
		t.Error("expected nil config for nil DockerBuild")
	}
}

func TestGetDockerBuildConfig_NilComponent(t *testing.T) {
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "test-module",
			Components: config.ModuleComponents{
				"pdf-oci": nil, // nil entry
			},
		},
	}

	cfg := getDockerBuildConfig(module, "pdf-oci")
	if cfg != nil {
		t.Error("expected nil config for nil component entry")
	}
}

func TestGetDockerBuildConfig_NamedComponentFallsBackToContainer(t *testing.T) {
	// Module has a named component WITHOUT docker_build, but has a "container" with docker_build.
	// Named lookup should fail, then fall back to container type.
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "mixed-module",
			Components: config.ModuleComponents{
				"my-comp": &config.ComponentEntry{
					// No DockerBuild here
				},
				"container": &config.ComponentEntry{
					Type: "container",
					DockerBuild: &config.DockerBuildConfig{
						Container: "fallback-container",
						Context:   "containers/fallback",
					},
				},
			},
		},
	}

	// Looking up "my-comp" should fall back to container type
	cfg := getDockerBuildConfig(module, "my-comp")
	if cfg == nil {
		t.Fatal("expected fallback to container type")
	}
	if cfg.Container != "fallback-container" {
		t.Errorf("expected container=fallback-container, got %s", cfg.Container)
	}
}

func TestGetDockerBuildConfig_ContainerTypeFallback(t *testing.T) {
	// Module has a "container" type component with docker_build.
	// Empty component name should fall back to "container" type.
	module := &modules.ModuleContract{
		BaseContract: domain.BaseContract{
			Moniker: "my-oci",
			Components: config.ModuleComponents{
				"container": &config.ComponentEntry{
					Type: "container",
					DockerBuild: &config.DockerBuildConfig{
						Container: "my-oci",
						Context:   "containers/my-oci",
					},
				},
			},
		},
	}

	// With empty component name, should fall back to "container" type
	cfg := getDockerBuildConfig(module, "")
	if cfg == nil {
		t.Fatal("expected non-nil config via fallback to 'container' type")
	}
	if cfg.Container != "my-oci" {
		t.Errorf("expected container=my-oci, got %s", cfg.Container)
	}
}

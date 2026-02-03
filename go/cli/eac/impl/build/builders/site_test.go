package builders

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func TestSiteHandler_Name(t *testing.T) {
	h := &SiteHandler{}
	if got := h.Name(); got != "site" {
		t.Errorf("Name() = %q, want %q", got, "site")
	}
}

func TestSiteHandler_Requirements(t *testing.T) {
	h := &SiteHandler{}
	reqs := h.Requirements()
	if len(reqs) != 1 || reqs[0] != "docker" {
		t.Errorf("Requirements() = %v, want [docker]", reqs)
	}
}

func TestSiteHandler_IsContainer(t *testing.T) {
	h := &SiteHandler{}
	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true (site rendering requires Docker)")
	}
}

func TestSiteHandler_IsHostInstalled(t *testing.T) {
	h := &SiteHandler{}
	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false (site rendering uses Docker)")
	}
}

func TestSiteHandler_ListArtifacts(t *testing.T) {
	h := &SiteHandler{}
	artifacts := h.ListArtifacts(nil, "/workspace")

	expected := []string{"site/"}
	if len(artifacts) != len(expected) {
		t.Fatalf("ListArtifacts() returned %d artifacts, want %d", len(artifacts), len(expected))
	}

	for i, a := range expected {
		if artifacts[i] != a {
			t.Errorf("ListArtifacts()[%d] = %q, want %q", i, artifacts[i], a)
		}
	}
}

func TestSiteHandler_Capabilities(t *testing.T) {
	h := &SiteHandler{}
	caps := h.Capabilities()

	expected := []string{"documentation", "html", "container"}
	if len(caps) != len(expected) {
		t.Fatalf("Capabilities() returned %d capabilities, want %d", len(caps), len(expected))
	}

	for i, c := range expected {
		if caps[i] != c {
			t.Errorf("Capabilities()[%d] = %q, want %q", i, caps[i], c)
		}
	}
}

func TestResolveBookNameForSite(t *testing.T) {
	tests := []struct {
		name          string
		componentName string
		want          string
	}{
		{
			name:          "empty component name",
			componentName: "",
			want:          "site",
		},
		{
			name:          "component name with -site suffix",
			componentName: "docs-site",
			want:          "docs",
		},
		{
			name:          "component name without -site suffix",
			componentName: "docs",
			want:          "docs",
		},
		{
			name:          "short component name (under 5 chars)",
			componentName: "site",
			want:          "site",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal module contract
			base := domain.BaseContract{
				Moniker:    "test-module",
				Components: make(domain.ModuleComponents),
			}
			module := modules.NewModuleContract(base, "/workspace")

			modulePort := adapters.AdaptModule(module)
			got := resolveBookNameForSite(modulePort, tt.componentName)

			if got != tt.want {
				t.Errorf("resolveBookNameForSite(%q) = %q, want %q", tt.componentName, got, tt.want)
			}
		})
	}
}

func TestResolveBookNameForSite_WithConfigOverride(t *testing.T) {
	// Create a module with explicit book config
	base := domain.BaseContract{
		Moniker: "test-module",
		Components: domain.ModuleComponents{
			"docs-site": &domain.ComponentEntry{
				Type: "docs-site",
				Config: map[string]string{
					"book": "custom-book-name",
				},
			},
		},
	}
	module := modules.NewModuleContract(base, "/workspace")

	modulePort := adapters.AdaptModule(module)
	got := resolveBookNameForSite(modulePort, "docs-site")

	want := "custom-book-name"
	if got != want {
		t.Errorf("resolveBookNameForSite() with config override = %q, want %q", got, want)
	}
}

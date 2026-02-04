package builders

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

func TestPDFHandler_Name(t *testing.T) {
	h := &PDFHandler{}
	if got := h.Name(); got != "pdf" {
		t.Errorf("Name() = %q, want %q", got, "pdf")
	}
}

func TestPDFHandler_Requirements(t *testing.T) {
	h := &PDFHandler{}
	reqs := h.Requirements()
	if len(reqs) != 1 || reqs[0] != "docker" {
		t.Errorf("Requirements() = %v, want [docker]", reqs)
	}
}

func TestPDFHandler_IsContainer(t *testing.T) {
	h := &PDFHandler{}
	if !h.IsContainer() {
		t.Error("IsContainer() = false, want true (PDF rendering requires Docker)")
	}
}

func TestPDFHandler_IsHostInstalled(t *testing.T) {
	h := &PDFHandler{}
	if h.IsHostInstalled() {
		t.Error("IsHostInstalled() = true, want false (PDF rendering uses Docker)")
	}
}

func TestPDFHandler_ListArtifacts(t *testing.T) {
	h := &PDFHandler{}
	artifacts := h.ListArtifacts(nil, "/workspace")

	expected := []string{"site/pdf/"}
	if len(artifacts) != len(expected) {
		t.Fatalf("ListArtifacts() returned %d artifacts, want %d", len(artifacts), len(expected))
	}

	for i, a := range expected {
		if artifacts[i] != a {
			t.Errorf("ListArtifacts()[%d] = %q, want %q", i, artifacts[i], a)
		}
	}
}

func TestPDFHandler_Capabilities(t *testing.T) {
	h := &PDFHandler{}
	caps := h.Capabilities()

	expected := []string{"documentation", "pdf", "container"}
	if len(caps) != len(expected) {
		t.Fatalf("Capabilities() returned %d capabilities, want %d", len(caps), len(expected))
	}

	for i, c := range expected {
		if caps[i] != c {
			t.Errorf("Capabilities()[%d] = %q, want %q", i, caps[i], c)
		}
	}
}

func TestResolveBookNameForPDF(t *testing.T) {
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
			name:          "component name with -pdf suffix",
			componentName: "tutorials-pdf",
			want:          "tutorials",
		},
		{
			name:          "component name without -pdf suffix",
			componentName: "docs",
			want:          "docs",
		},
		{
			name:          "short component name (under 4 chars)",
			componentName: "pdf",
			want:          "pdf",
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
			got := resolveBookNameForPDF(modulePort, tt.componentName)

			if got != tt.want {
				t.Errorf("resolveBookNameForPDF(%q) = %q, want %q", tt.componentName, got, tt.want)
			}
		})
	}
}

func TestResolveBookNameForPDF_WithConfigOverride(t *testing.T) {
	// Create a module with explicit book config
	base := domain.BaseContract{
		Moniker: "test-module",
		Components: domain.ModuleComponents{
			"tutorials-pdf": &domain.ComponentEntry{
				Type: "docs-pdf",
				Config: map[string]string{
					"book": "custom-book-name",
				},
			},
		},
	}
	module := modules.NewModuleContract(base, "/workspace")

	modulePort := adapters.AdaptModule(module)
	got := resolveBookNameForPDF(modulePort, "tutorials-pdf")

	want := "custom-book-name"
	if got != want {
		t.Errorf("resolveBookNameForPDF() with config override = %q, want %q", got, want)
	}
}

package serve

import (
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
)

func TestGetServableItems_SiteRender(t *testing.T) {
	// Create a module with site-render component
	module := &config.Module{
		Moniker: "docs",
		Components: config.ModuleComponents{
			"site": &config.ComponentEntry{
				Type: config.ComponentTypeSiteRender,
			},
		},
	}

	items := getServableItems(module)

	if len(items) == 0 {
		t.Fatal("Expected at least one servable item")
	}

	found := false
	for _, item := range items {
		if item.name == "site" && item.isSite {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'site' as a site-render item with isSite=true")
	}
}

func TestGetServableItems_PDFRender(t *testing.T) {
	// Create a module with pdf-render component
	module := &config.Module{
		Moniker: "books",
		Components: config.ModuleComponents{
			"handbook": &config.ComponentEntry{
				Type: config.ComponentTypePdfRender,
			},
		},
	}

	items := getServableItems(module)

	if len(items) == 0 {
		t.Fatal("Expected at least one servable item")
	}

	found := false
	for _, item := range items {
		if item.name == "handbook" && !item.isSite {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'handbook' as a pdf-render item with isSite=false")
	}
}

func TestGetServableItems_Mixed(t *testing.T) {
	// Create a module with both site-render and pdf-render components
	module := &config.Module{
		Moniker: "docs",
		Components: config.ModuleComponents{
			"site": &config.ComponentEntry{
				Type: config.ComponentTypeSiteRender,
			},
			"manual": &config.ComponentEntry{
				Type: config.ComponentTypePdfRender,
			},
		},
	}

	items := getServableItems(module)

	if len(items) != 2 {
		t.Fatalf("Expected 2 servable items, got %d", len(items))
	}

	hasSite := false
	hasPDF := false
	for _, item := range items {
		if item.name == "site" && item.isSite {
			hasSite = true
		}
		if item.name == "manual" && !item.isSite {
			hasPDF = true
		}
	}

	if !hasSite {
		t.Error("Expected to find 'site' item")
	}
	if !hasPDF {
		t.Error("Expected to find 'manual' item")
	}
}

func TestGetServableItems_NoServable(t *testing.T) {
	// Create a module with no servable components
	module := &config.Module{
		Moniker: "gomod",
		Components: config.ModuleComponents{
			"main": &config.ComponentEntry{
				Type: "go",
			},
		},
	}

	items := getServableItems(module)

	if len(items) != 0 {
		t.Errorf("Expected 0 servable items for go module, got %d", len(items))
	}
}

func TestModuleServeConfig_IsSite(t *testing.T) {
	tests := []struct {
		name     string
		config   *ModuleServeConfig
		expected bool
	}{
		{
			name: "Site module",
			config: &ModuleServeConfig{
				ModuleMoniker: "docs",
				ContentPath:   "/workspace/out/build/docs/site/site",
				IsSite:        true,
			},
			expected: true,
		},
		{
			name: "PDF module",
			config: &ModuleServeConfig{
				ModuleMoniker: "books",
				ContentPath:   "/workspace/out/build/books",
				IsSite:        false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config.IsSite != tt.expected {
				t.Errorf("IsSite = %v, want %v", tt.config.IsSite, tt.expected)
			}
		})
	}
}

func TestGetServableItems_DocsSite(t *testing.T) {
	// Create a module with docs-site component (used in docs module)
	module := &config.Module{
		Moniker: "docs",
		Components: config.ModuleComponents{
			"site": &config.ComponentEntry{
				Type: config.ComponentTypeDocsSite,
			},
		},
	}

	items := getServableItems(module)

	if len(items) == 0 {
		t.Fatal("Expected at least one servable item")
	}

	found := false
	for _, item := range items {
		if item.name == "site" && item.isSite {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'site' as a docs-site item with isSite=true")
	}
}

func TestGetServableItems_DocsPDF(t *testing.T) {
	// Create a module with docs-pdf component
	module := &config.Module{
		Moniker: "docs",
		Components: config.ModuleComponents{
			"tutorials": &config.ComponentEntry{
				Type: config.ComponentTypeDocsPdf,
			},
		},
	}

	items := getServableItems(module)

	if len(items) == 0 {
		t.Fatal("Expected at least one servable item")
	}

	found := false
	for _, item := range items {
		if item.name == "tutorials" && !item.isSite {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find 'tutorials' as a docs-pdf item with isSite=false")
	}
}

func TestServableItem(t *testing.T) {
	// Test the servableItem struct
	siteItem := servableItem{
		name:   "site",
		isSite: true,
	}

	if siteItem.name != "site" {
		t.Errorf("name = %q, want %q", siteItem.name, "site")
	}
	if !siteItem.isSite {
		t.Error("isSite should be true for site item")
	}

	pdfItem := servableItem{
		name:   "handbook",
		isSite: false,
	}

	if pdfItem.name != "handbook" {
		t.Errorf("name = %q, want %q", pdfItem.name, "handbook")
	}
	if pdfItem.isSite {
		t.Error("isSite should be false for PDF item")
	}
}

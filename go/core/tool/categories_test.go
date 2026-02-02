package tool

import (
	"sort"
	"testing"
)

func TestScannerToolIDForCategory(t *testing.T) {
	tests := []struct {
		category string
		want     string
	}{
		{"sbom", ToolTrivySBOM},
		{"vuln", ToolTrivyVuln},
		{"secrets", ToolTrivySecrets},
		{"compliance", ToolTrivyCompliance},
		{"iac", ToolTrivyIaC},
		{"sast", ToolSemgrep},
		{"zap", ToolZap},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			got := ScannerToolIDForCategory(tt.category)
			if got != tt.want {
				t.Errorf("ScannerToolIDForCategory(%q) = %q, want %q", tt.category, got, tt.want)
			}
		})
	}
}

func TestServerToolIDForType(t *testing.T) {
	tests := []struct {
		serverType string
		want       string
	}{
		{ToolStaticSite, ToolStaticSite},
		{ToolMkDocsLive, ToolMkDocsLive},
		{ToolStructurizrLite, ToolStructurizrLite},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.serverType, func(t *testing.T) {
			got := ServerToolIDForType(tt.serverType)
			if got != tt.want {
				t.Errorf("ServerToolIDForType(%q) = %q, want %q", tt.serverType, got, tt.want)
			}
		})
	}
}

func TestAllScannerCategories(t *testing.T) {
	categories := AllScannerCategories()

	if len(categories) != 7 {
		t.Errorf("AllScannerCategories() returned %d items, want 7", len(categories))
	}

	// Check all expected categories
	expected := []string{"sbom", "vuln", "secrets", "compliance", "iac", "sast", "zap"}
	catSet := make(map[string]bool)
	for _, cat := range categories {
		catSet[cat] = true
	}

	for _, exp := range expected {
		if !catSet[exp] {
			t.Errorf("AllScannerCategories() missing %q", exp)
		}
	}
}

func TestAllServerTypes(t *testing.T) {
	types := AllServerTypes()

	if len(types) != 3 {
		t.Errorf("AllServerTypes() returned %d items, want 3", len(types))
	}

	// Check all expected server types
	expected := []string{ToolStaticSite, ToolMkDocsLive, ToolStructurizrLite}
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	for _, exp := range expected {
		if !typeSet[exp] {
			t.Errorf("AllServerTypes() missing %q", exp)
		}
	}
}

func TestIsScannerCategory(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"sbom", true},
		{"vuln", true},
		{"secrets", true},
		{"compliance", true},
		{"iac", true},
		{"sast", true},
		{"zap", true},
		{"unknown", false},
		{"SBOM", false}, // case-sensitive
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsScannerCategory(tt.input)
			if got != tt.want {
				t.Errorf("IsScannerCategory(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsServerType(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{ToolStaticSite, true},
		{ToolMkDocsLive, true},
		{ToolStructurizrLite, true},
		{"unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsServerType(tt.input)
			if got != tt.want {
				t.Errorf("IsServerType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCategoryToToolID_Deprecated(t *testing.T) {
	// Test backward compatibility function
	tests := []struct {
		category string
		wantID   string
		wantOK   bool
	}{
		{"sbom", ToolTrivySBOM, true},
		{"vuln", ToolTrivyVuln, true},
		{"unknown", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.category, func(t *testing.T) {
			gotID, gotOK := CategoryToToolID(tt.category)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("CategoryToToolID(%q) = (%q, %v), want (%q, %v)",
					tt.category, gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

func TestCategoryResolverThreadSafety(t *testing.T) {
	// Test concurrent access to category resolver
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = ScannerToolIDForCategory("sbom")
				_ = ServerToolIDForType(ToolStaticSite)
				_ = AllScannerCategories()
				_ = AllServerTypes()
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCategoryMappingConsistency(t *testing.T) {
	// Ensure that AllScannerCategories returns categories that map to tool IDs
	categories := AllScannerCategories()
	sort.Strings(categories)

	for _, cat := range categories {
		toolID := ScannerToolIDForCategory(cat)
		if toolID == "" {
			t.Errorf("Category %q returned by AllScannerCategories() has no tool ID mapping", cat)
		}
	}
}

func TestServerTypeMappingConsistency(t *testing.T) {
	// Ensure that AllServerTypes returns types that map to tool IDs
	types := AllServerTypes()

	for _, serverType := range types {
		toolID := ServerToolIDForType(serverType)
		if toolID == "" {
			t.Errorf("Server type %q returned by AllServerTypes() has no tool ID mapping", serverType)
		}
	}
}

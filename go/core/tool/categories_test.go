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

// resetGlobalCategoryResolver resets the singleton so tests start clean.
func resetGlobalCategoryResolver() {
	globalCategoryResolver.mu.Lock()
	defer globalCategoryResolver.mu.Unlock()
	globalCategoryResolver.scannerMap = nil
	globalCategoryResolver.serverMap = nil
	globalCategoryResolver.initialized = false
}

func TestDefaultScannerCategoryMapMatchesResolver(t *testing.T) {
	// The package-level DefaultScannerCategoryMap should contain every
	// category returned by AllScannerCategories.
	resetGlobalCategoryResolver()
	defer resetGlobalCategoryResolver()

	categories := AllScannerCategories()
	if len(categories) != len(DefaultScannerCategoryMap) {
		t.Errorf("AllScannerCategories() returned %d items but DefaultScannerCategoryMap has %d entries",
			len(categories), len(DefaultScannerCategoryMap))
	}

	for _, cat := range categories {
		if _, ok := DefaultScannerCategoryMap[cat]; !ok {
			t.Errorf("Category %q from resolver is missing in DefaultScannerCategoryMap", cat)
		}
	}
}

func TestDefaultServerTypeMapMatchesResolver(t *testing.T) {
	// The package-level DefaultServerTypeMap should contain every
	// server type returned by AllServerTypes.
	resetGlobalCategoryResolver()
	defer resetGlobalCategoryResolver()

	types := AllServerTypes()
	if len(types) != len(DefaultServerTypeMap) {
		t.Errorf("AllServerTypes() returned %d items but DefaultServerTypeMap has %d entries",
			len(types), len(DefaultServerTypeMap))
	}

	for _, st := range types {
		if _, ok := DefaultServerTypeMap[st]; !ok {
			t.Errorf("Server type %q from resolver is missing in DefaultServerTypeMap", st)
		}
	}
}

func TestOverrideScannerCategoryMap(t *testing.T) {
	// Save and restore state to avoid affecting other tests.
	resetGlobalCategoryResolver()
	defer resetGlobalCategoryResolver()

	custom := map[string]string{
		"custom-scan": "custom-tool-id",
	}
	OverrideScannerCategoryMap(custom)

	// The overridden map should be active.
	got := ScannerToolIDForCategory("custom-scan")
	if got != "custom-tool-id" {
		t.Errorf("after override, ScannerToolIDForCategory(%q) = %q, want %q", "custom-scan", got, "custom-tool-id")
	}

	// Original categories should no longer resolve.
	got = ScannerToolIDForCategory("sbom")
	if got != "" {
		t.Errorf("after override, ScannerToolIDForCategory(%q) = %q, want empty", "sbom", got)
	}

	// AllScannerCategories should reflect the override.
	cats := AllScannerCategories()
	if len(cats) != 1 {
		t.Errorf("after override, AllScannerCategories() returned %d items, want 1", len(cats))
	}
}

func TestOverrideServerTypeMap(t *testing.T) {
	// Save and restore state to avoid affecting other tests.
	resetGlobalCategoryResolver()
	defer resetGlobalCategoryResolver()

	custom := map[string]string{
		"my-server": "my-server-tool",
	}
	OverrideServerTypeMap(custom)

	// The overridden map should be active.
	got := ServerToolIDForType("my-server")
	if got != "my-server-tool" {
		t.Errorf("after override, ServerToolIDForType(%q) = %q, want %q", "my-server", got, "my-server-tool")
	}

	// Original server types should no longer resolve.
	got = ServerToolIDForType(ToolStaticSite)
	if got != "" {
		t.Errorf("after override, ServerToolIDForType(%q) = %q, want empty", ToolStaticSite, got)
	}

	// AllServerTypes should reflect the override.
	types := AllServerTypes()
	if len(types) != 1 {
		t.Errorf("after override, AllServerTypes() returned %d items, want 1", len(types))
	}
}

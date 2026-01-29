package tool

import (
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

func TestNewScanBridge(t *testing.T) {
	bridge := NewScanBridge()

	if bridge == nil {
		t.Fatal("NewScanBridge returned nil")
	}
	if bridge.scannerTools == nil {
		t.Error("scannerTools map not initialized")
	}
}

func TestScanBridge_GetScanner_YAMLTool(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system - use LocalPath for test containers (no version pinning needed)
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "trivy-sbom", // Matches default mapping
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should return YAML tool scanner
	scanner := bridge.GetScanner(ScannerSBOM)
	if scanner == nil {
		t.Fatal("GetScanner returned nil for YAML tool scanner")
	}
}

func TestScanBridge_GetScanner_NotFound(t *testing.T) {
	bridge := NewScanBridge()

	// No tool system configured
	scanner := bridge.GetScanner(ScannerSBOM)
	if scanner != nil {
		t.Error("GetScanner should return nil when no scanner available")
	}
}

func TestScanBridge_GetScanner_CustomMapping(t *testing.T) {
	bridge := NewScanBridge()

	// Register custom mapping
	bridge.SetScannerToolMapping(ScannerSAST, "custom-sast-tool")

	// Set up tool system with the custom tool
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:        "custom-sast-tool",
		Type:      ToolTypeContainer,
		LocalPath: "containers/custom-sast",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Should use custom mapping
	scanner := bridge.GetScanner(ScannerSAST)
	if scanner == nil {
		t.Fatal("GetScanner returned nil for custom mapped scanner")
	}
}

func TestScanBridge_SetToolSystem(t *testing.T) {
	bridge := NewScanBridge()

	registry := NewRegistry()
	resolver := NewResolver(registry)
	executor := &mockExecutor{}

	bridge.SetToolSystem(registry, resolver, executor)

	// Verify tool system is set by registering a tool and retrieving it
	registry.Register(&ToolDefinition{
		ID:    "trivy-vuln",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})

	scanner := bridge.GetScanner(ScannerVuln)
	if scanner == nil {
		t.Error("Tool system not properly configured")
	}
}

func TestScanBridge_HasScanner(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "trivy-sbom",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	registry.Register(&ToolDefinition{
		ID:    "trivy-vuln",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	tests := []struct {
		scannerType ScannerType
		exists      bool
	}{
		{ScannerSBOM, true},     // yaml
		{ScannerVuln, true},     // yaml
		{ScannerSecrets, false}, // not registered
	}

	for _, tt := range tests {
		t.Run(string(tt.scannerType), func(t *testing.T) {
			if got := bridge.HasScanner(tt.scannerType); got != tt.exists {
				t.Errorf("HasScanner(%q) = %v, want %v", tt.scannerType, got, tt.exists)
			}
		})
	}
}

func TestScanBridge_GetAllScannerTypes(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "trivy-sbom",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	registry.Register(&ToolDefinition{
		ID:    "trivy-vuln",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	registry.Register(&ToolDefinition{
		ID:    "trivy-secrets",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	scannerTypes := bridge.GetAllScannerTypes()

	// Should include: sbom, vuln, secrets
	if len(scannerTypes) < 3 {
		t.Errorf("GetAllScannerTypes() returned %d types, want at least 3", len(scannerTypes))
	}
}

func TestScanBridge_GetScannersForModule_NilModule(t *testing.T) {
	bridge := NewScanBridge()

	scanners := bridge.GetScannersForModule(nil, nil)
	if scanners != nil {
		t.Error("GetScannersForModule(nil, nil) should return nil")
	}
}

func TestScanBridge_GetScannersForModule_NilComponentTypes(t *testing.T) {
	bridge := NewScanBridge()

	module := modules.NewModuleContract(contracts.BaseContract{
		Moniker: "test-module",
	}, "/workspace")

	scanners := bridge.GetScannersForModule(module, nil)
	if scanners != nil {
		t.Error("GetScannersForModule with nil componentTypes should return nil")
	}
}

func TestScanBridge_GetScannersForModule(t *testing.T) {
	bridge := NewScanBridge()

	// Create a module with a go component
	module := modules.NewModuleContract(contracts.BaseContract{
		Moniker: "test-module",
		Components: contracts.ModuleComponents{
			"main": &contracts.ComponentEntry{
				Root: ".",
				Type: "go",
			},
		},
	}, "/workspace")

	// Create component types config with scanners
	componentTypes := &config.ComponentTypesConfig{
		ComponentTypes: map[string]*config.ComponentType{
			"go": {
				Scanners: []string{"sbom", "vuln", "secrets"},
			},
		},
	}

	scanners := bridge.GetScannersForModule(module, componentTypes)

	// Should return scanner types for go component
	if len(scanners) == 0 {
		t.Fatal("GetScannersForModule returned empty scanner list")
	}

	// Verify expected scanners
	found := make(map[ScannerType]bool)
	for _, st := range scanners {
		found[st] = true
	}

	if !found[ScannerSBOM] {
		t.Error("ScannerSBOM should be in scanner list")
	}
	if !found[ScannerVuln] {
		t.Error("ScannerVuln should be in scanner list")
	}
	if !found[ScannerSecrets] {
		t.Error("ScannerSecrets should be in scanner list")
	}
}

func TestScanBridge_GetScannersForModule_MultipleComponents(t *testing.T) {
	bridge := NewScanBridge()

	// Create a module with multiple components
	module := modules.NewModuleContract(contracts.BaseContract{
		Moniker: "test-module",
		Components: contracts.ModuleComponents{
			"main": &contracts.ComponentEntry{
				Root: ".",
				Type: "go",
			},
			"config": &contracts.ComponentEntry{
				Root: "config",
				Type: "dockerfile",
			},
		},
	}, "/workspace")

	// Create component types config with different scanners
	componentTypes := &config.ComponentTypesConfig{
		ComponentTypes: map[string]*config.ComponentType{
			"go": {
				Scanners: []string{"sbom", "vuln"},
			},
			"dockerfile": {
				Scanners: []string{"iac", "secrets"},
			},
		},
	}

	scanners := bridge.GetScannersForModule(module, componentTypes)

	// Should return union of scanners from both component types
	// Go: sbom, vuln
	// Dockerfile: iac, secrets
	// Total unique: sbom, vuln, iac, secrets = 4
	if len(scanners) != 4 {
		t.Errorf("GetScannersForModule returned %d scanners, want 4", len(scanners))
	}
}

func TestGlobalScanBridge(t *testing.T) {
	// GlobalScanBridge should return same instance
	bridge1 := GlobalScanBridge()
	bridge2 := GlobalScanBridge()

	if bridge1 != bridge2 {
		t.Error("GlobalScanBridge should return singleton")
	}

	if bridge1 == nil {
		t.Error("GlobalScanBridge should not return nil")
	}
}

func TestParseScannerType(t *testing.T) {
	tests := []struct {
		input    string
		expected ScannerType
		ok       bool
	}{
		{"sbom", ScannerSBOM, true},
		{"vuln", ScannerVuln, true},
		{"secrets", ScannerSecrets, true},
		{"compliance", ScannerCompliance, true},
		{"iac", ScannerIaC, true},
		{"sast", ScannerSAST, true},
		{"zap", ScannerDAST, true},
		{"unknown", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseScannerType(tt.input)
			if ok != tt.ok {
				t.Errorf("ParseScannerType(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if got != tt.expected {
				t.Errorf("ParseScannerType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestScannerTypeConstants(t *testing.T) {
	// Verify all scanner type constants have expected values
	tests := []struct {
		constant ScannerType
		value    string
	}{
		{ScannerSBOM, "sbom"},
		{ScannerVuln, "vuln"},
		{ScannerSecrets, "secrets"},
		{ScannerCompliance, "compliance"},
		{ScannerIaC, "iac"},
		{ScannerSAST, "sast"},
		{ScannerDAST, "zap"},
	}

	for _, tt := range tests {
		if string(tt.constant) != tt.value {
			t.Errorf("Scanner constant %v = %q, want %q", tt.constant, tt.constant, tt.value)
		}
	}
}

func TestScanBridge_DefaultScannerMappings(t *testing.T) {
	bridge := NewScanBridge()

	// Verify default mappings exist
	expected := map[ScannerType]string{
		ScannerSBOM:       "trivy-sbom",
		ScannerVuln:       "trivy-vuln",
		ScannerSecrets:    "trivy-secrets",
		ScannerIaC:        "trivy-iac",
		ScannerCompliance: "trivy-compliance",
		ScannerSAST:       "semgrep",
		ScannerDAST:       "zap",
	}

	for scannerType, expectedToolID := range expected {
		toolID := bridge.GetScannerToolID(scannerType)
		if toolID != expectedToolID {
			t.Errorf("Default mapping for %v = %q, want %q", scannerType, toolID, expectedToolID)
		}
	}
}

func TestScanBridge_Concurrent(t *testing.T) {
	bridge := NewScanBridge()

	// Set up tool system
	registry := NewRegistry()
	registry.Register(&ToolDefinition{
		ID:    "trivy-sbom",
		Type:      ToolTypeContainer,
		LocalPath: "containers/trivy",
	})
	bridge.SetToolSystem(registry, nil, &mockExecutor{})

	// Test concurrent access
	done := make(chan bool)

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetScanner(ScannerSBOM)
		}
		done <- true
	}()

	// Concurrent HasScanner
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.HasScanner(ScannerVuln)
		}
		done <- true
	}()

	// Concurrent GetAllScannerTypes
	go func() {
		for i := 0; i < 100; i++ {
			_ = bridge.GetAllScannerTypes()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done
}

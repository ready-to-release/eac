package design

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateModule(t *testing.T) {
	tests := []struct {
		name        string
		module      string
		expectError bool
	}{
		{
			name:        "valid module",
			module:      "src-cli",
			expectError: false,
		},
		{
			name:        "nonexistent module",
			module:      "nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModule(tt.module)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

func TestListAvailableModules(t *testing.T) {
	modules, err := ListAvailableModules()
	if err != nil {
		t.Fatalf("ListAvailableModules failed: %v", err)
	}

	// Should find at least the src-cli module
	found := false
	for _, m := range modules {
		if m.Name == "src-cli" {
			found = true
			if !m.HasWorkspace {
				t.Error("src-cli module should have workspace.dsl")
			}
			break
		}
	}

	if !found {
		t.Error("src-cli module not found in available modules")
	}
}

func TestGetContainerName(t *testing.T) {
	tests := []struct {
		module   string
		expected string
	}{
		{"src-cli", "structurizr-cli"},
		{"vscode", "structurizr-vscode"},
		{"mcp", "structurizr-mcp"},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := GetContainerName(tt.module)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestGetModulePath(t *testing.T) {
	tests := []struct {
		module string
	}{
		{"src-cli"},
		{"vscode"},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := GetModulePath(tt.module)
			// Path should contain the module name and design directory structure
			expectedComponents := []string{"docs", "reference", "design", tt.module}
			for _, component := range expectedComponents {
				if !containsPathComponent(result, component) {
					t.Errorf("path %s should contain component %s", result, component)
				}
			}
		})
	}
}

// Helper to check if a path contains a specific component
func containsPathComponent(path, component string) bool {
	// Normalize path separators for comparison
	normalizedPath := filepath.ToSlash(path)
	return strings.Contains(normalizedPath, component)
}

func TestModuleExists(t *testing.T) {
	tests := []struct {
		module   string
		expected bool
	}{
		{"src-cli", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := ModuleExists(tt.module)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestModuleInfoGetStatus(t *testing.T) {
	tests := []struct {
		name     string
		info     ModuleInfo
		expected string
	}{
		{
			name: "fully documented",
			info: ModuleInfo{
				HasWorkspace: true,
				HasDocs:      true,
				HasDecisions: true,
			},
			expected: "✅ Ready",
		},
		{
			name: "partial documentation",
			info: ModuleInfo{
				HasWorkspace: true,
				HasDocs:      false,
				HasDecisions: true,
			},
			expected: "⚠️  Partial",
		},
		{
			name: "missing workspace",
			info: ModuleInfo{
				HasWorkspace: false,
			},
			expected: "❌ Missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.info.GetStatus()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDetectBrowser(t *testing.T) {
	command := DetectBrowser()
	if command == "" {
		t.Error("DetectBrowser returned empty string")
	}

	// Should return platform-specific command
	validCommands := []string{"cmd /c start", "open", "xdg-open"}
	found := false
	for _, valid := range validCommands {
		if command == valid {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("DetectBrowser returned unexpected command: %s", command)
	}
}

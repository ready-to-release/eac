package modules

import (
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/internal/contracts"
)

func TestBuildParentChain(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (*ModuleContract, *Registry)
		expected    []string
		expectError bool
		errorMsg    string
	}{
		{
			name: "root level module",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")
				module := createTestModuleWithParent("repository", ".", "/")
				registry.Add(module)
				return module, registry
			},
			expected:    []string{".", "repository"},
			expectError: false,
		},
		{
			name: "one parent level",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				parent := createTestModuleWithParent("claude", ".", "/")
				child := createTestModuleWithParent("claude-agents", "claude", ".claude/agents")

				registry.Add(parent)
				registry.Add(child)

				return child, registry
			},
			expected:    []string{".", "claude", "claude-agents"},
			expectError: false,
		},
		{
			name: "two parent levels",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				root := createTestModuleWithParent("docs", ".", "docs")
				mid := createTestModuleWithParent("docs-api", "docs", "docs/api")
				leaf := createTestModuleWithParent("docs-api-v1", "docs-api", "docs/api/v1")

				registry.Add(root)
				registry.Add(mid)
				registry.Add(leaf)

				return leaf, registry
			},
			expected:    []string{".", "docs", "docs-api", "docs-api-v1"},
			expectError: false,
		},
		{
			name: "missing parent",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")
				module := createTestModuleWithParent("child", "non-existent", "/child")
				registry.Add(module)
				return module, registry
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "parent module 'non-existent' not found",
		},
		{
			name: "circular reference - direct",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				// Create modules with circular reference
				moduleA := createTestModuleWithParent("moduleA", "moduleB", "/a")
				moduleB := createTestModuleWithParent("moduleB", "moduleA", "/b")

				registry.Add(moduleA)
				registry.Add(moduleB)

				return moduleA, registry
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "circular parent chain",
		},
		{
			name: "circular reference - indirect",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				// Create modules with circular reference through 3 modules
				moduleA := createTestModuleWithParent("moduleA", "moduleB", "/a")
				moduleB := createTestModuleWithParent("moduleB", "moduleC", "/b")
				moduleC := createTestModuleWithParent("moduleC", "moduleA", "/c")

				registry.Add(moduleA)
				registry.Add(moduleB)
				registry.Add(moduleC)

				return moduleA, registry
			},
			expected:    nil,
			expectError: true,
			errorMsg:    "circular parent chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, registry := tt.setupFunc()

			chain, err := BuildParentChain(module, registry)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error containing '%s', got no error", tt.errorMsg)
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(chain) != len(tt.expected) {
				t.Errorf("expected chain length %d, got %d", len(tt.expected), len(chain))
				return
			}

			for i, expected := range tt.expected {
				if chain[i] != expected {
					t.Errorf("at position %d: expected '%s', got '%s'", i, expected, chain[i])
				}
			}
		})
	}
}

func TestGetDepth(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func() (*ModuleContract, *Registry)
		expected    int
		expectError bool
	}{
		{
			name: "depth 1 - root level",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")
				module := createTestModuleWithParent("repository", ".", "/")
				registry.Add(module)
				return module, registry
			},
			expected:    1,
			expectError: false,
		},
		{
			name: "depth 2 - one parent",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				parent := createTestModuleWithParent("claude", ".", "/")
				child := createTestModuleWithParent("claude-agents", "claude", ".claude/agents")

				registry.Add(parent)
				registry.Add(child)

				return child, registry
			},
			expected:    2,
			expectError: false,
		},
		{
			name: "depth 3 - two parents",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				root := createTestModuleWithParent("docs", ".", "docs")
				mid := createTestModuleWithParent("docs-api", "docs", "docs/api")
				leaf := createTestModuleWithParent("docs-api-v1", "docs-api", "docs/api/v1")

				registry.Add(root)
				registry.Add(mid)
				registry.Add(leaf)

				return leaf, registry
			},
			expected:    3,
			expectError: false,
		},
		{
			name: "error - missing parent",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")
				module := createTestModuleWithParent("child", "non-existent", "/child")
				registry.Add(module)
				return module, registry
			},
			expected:    0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, registry := tt.setupFunc()

			depth, err := GetDepth(module, registry)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got no error")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if depth != tt.expected {
				t.Errorf("expected depth %d, got %d", tt.expected, depth)
			}
		})
	}
}

func TestValidateParentChain(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*ModuleContract, *Registry)
		expectErr bool
	}{
		{
			name: "valid chain",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				parent := createTestModuleWithParent("claude", ".", "/")
				child := createTestModuleWithParent("claude-agents", "claude", ".claude/agents")

				registry.Add(parent)
				registry.Add(child)

				return child, registry
			},
			expectErr: false,
		},
		{
			name: "invalid - missing parent",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")
				module := createTestModuleWithParent("child", "non-existent", "/child")
				registry.Add(module)
				return module, registry
			},
			expectErr: true,
		},
		{
			name: "invalid - circular reference",
			setupFunc: func() (*ModuleContract, *Registry) {
				registry := NewRegistry("0.1.0", "/test")

				moduleA := createTestModuleWithParent("moduleA", "moduleB", "/a")
				moduleB := createTestModuleWithParent("moduleB", "moduleA", "/b")

				registry.Add(moduleA)
				registry.Add(moduleB)

				return moduleA, registry
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			module, registry := tt.setupFunc()

			err := ValidateParentChain(module, registry)

			if tt.expectErr && err == nil {
				t.Errorf("expected error, got no error")
			}

			if !tt.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Helper function to create test modules with parent
func createTestModuleWithParent(moniker, parent, root string) *ModuleContract {
	base := contracts.BaseContract{
		Moniker: moniker,
		Name:    moniker,
		Type:    "test",
		Parent:  parent,
		Source: contracts.Source{
			Root:     root,
			Includes: []string{"**/*"},
		},
	}
	return NewModuleContract(base, "/test")
}

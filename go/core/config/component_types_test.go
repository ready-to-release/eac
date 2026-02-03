package config

import (
	"testing"
)

func TestComponentTypeGetTools(t *testing.T) {
	tests := []struct {
		name  string
		ct    *ComponentType
		want  []string
	}{
		{
			name: "tools field takes precedence over builder",
			ct: &ComponentType{
				ToolChain:   []string{"preprocess", "mkdocs-pdf"},
				Builder: "legacy-builder",
			},
			want: []string{"preprocess", "mkdocs-pdf"},
		},
		{
			name: "falls back to builder when tools is empty",
			ct: &ComponentType{
				Builder: "go",
			},
			want: []string{"go"},
		},
		{
			name: "returns nil when both tools and builder are empty",
			ct:   &ComponentType{},
			want: nil,
		},
		{
			name: "single tool in tools field",
			ct: &ComponentType{
				ToolChain: []string{"mkdocs-build"},
			},
			want: []string{"mkdocs-build"},
		},
		{
			name: "three tools in chain",
			ct: &ComponentType{
				ToolChain: []string{"step1", "step2", "step3"},
			},
			want: []string{"step1", "step2", "step3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.GetTools()
			if !stringSliceEqual(got, tt.want) {
				t.Errorf("GetTools() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test nil receiver
	t.Run("nil receiver returns nil", func(t *testing.T) {
		var ct *ComponentType
		if got := ct.GetTools(); got != nil {
			t.Errorf("nil.GetTools() = %v, want nil", got)
		}
	})
}

func TestComponentTypeHasToolChain(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "true when tools has multiple entries",
			ct: &ComponentType{
				ToolChain: []string{"preprocess", "mkdocs-pdf"},
			},
			want: true,
		},
		{
			name: "false when tools has single entry",
			ct: &ComponentType{
				ToolChain: []string{"single-tool"},
			},
			want: false,
		},
		{
			name: "false when only builder is set",
			ct: &ComponentType{
				Builder: "go",
			},
			want: false,
		},
		{
			name: "false when neither tools nor builder is set",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "true with three tools",
			ct: &ComponentType{
				ToolChain: []string{"a", "b", "c"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.HasToolChain()
			if got != tt.want {
				t.Errorf("HasToolChain() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test nil receiver
	t.Run("nil receiver returns false", func(t *testing.T) {
		var ct *ComponentType
		if got := ct.HasToolChain(); got != false {
			t.Errorf("nil.HasToolChain() = %v, want false", got)
		}
	})
}

// stringSliceEqual compares two string slices for equality.
func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestNewComponentTypesWithToolChains validates that the new docs-pdf and docs-site
// component types have proper tool chains configured.
func TestNewComponentTypesWithToolChains(t *testing.T) {
	t.Run("docs-pdf has tool chain", func(t *testing.T) {
		ct := &ComponentType{
			ToolChain: []string{"mkdocs-preprocess", "pdf-tool"},
		}
		if !ct.HasToolChain() {
			t.Error("docs-pdf should have tool chain")
		}
		tools := ct.GetTools()
		if len(tools) != 2 {
			t.Errorf("docs-pdf should have 2 tools, got %d", len(tools))
		}
		if tools[0] != "mkdocs-preprocess" {
			t.Errorf("first tool should be mkdocs-preprocess, got %s", tools[0])
		}
		if tools[1] != "pdf-tool" {
			t.Errorf("second tool should be pdf-tool, got %s", tools[1])
		}
	})

	t.Run("docs-site has tool chain", func(t *testing.T) {
		ct := &ComponentType{
			ToolChain: []string{"mkdocs-preprocess", "site-render-tool"},
		}
		if !ct.HasToolChain() {
			t.Error("docs-site should have tool chain")
		}
		tools := ct.GetTools()
		if len(tools) != 2 {
			t.Errorf("docs-site should have 2 tools, got %d", len(tools))
		}
		if tools[0] != "mkdocs-preprocess" {
			t.Errorf("first tool should be mkdocs-preprocess, got %s", tools[0])
		}
		if tools[1] != "site-render-tool" {
			t.Errorf("second tool should be site-render-tool, got %s", tools[1])
		}
	})

	t.Run("single builder has no tool chain", func(t *testing.T) {
		ct := &ComponentType{
			Builder: "go",
		}
		if ct.HasToolChain() {
			t.Error("single builder should not have tool chain")
		}
		tools := ct.GetTools()
		if len(tools) != 1 || tools[0] != "go" {
			t.Errorf("expected [go], got %v", tools)
		}
	})
}

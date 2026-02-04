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

func TestComponentType_GetPool(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want string
	}{
		{
			name: "explicit host pool",
			ct:   &ComponentType{Pool: "host"},
			want: "host",
		},
		{
			name: "explicit docker pool",
			ct:   &ComponentType{Pool: "docker"},
			want: "docker",
		},
		{
			name: "docker requirement implies docker pool",
			ct:   &ComponentType{Requirements: []string{"docker"}},
			want: "docker",
		},
		{
			name: "go requirement defaults to host pool",
			ct:   &ComponentType{Requirements: []string{"go"}},
			want: "host",
		},
		{
			name: "empty component defaults to host",
			ct:   &ComponentType{},
			want: "host",
		},
		{
			name: "explicit pool overrides requirements inference",
			ct:   &ComponentType{Pool: "host", Requirements: []string{"docker"}},
			want: "host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.GetPool()
			if got != tt.want {
				t.Errorf("GetPool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentType_RequiresDocker(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "has docker requirement",
			ct:   &ComponentType{Requirements: []string{"go", "docker"}},
			want: true,
		},
		{
			name: "no docker requirement",
			ct:   &ComponentType{Requirements: []string{"go", "npm"}},
			want: false,
		},
		{
			name: "empty requirements",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "only docker requirement",
			ct:   &ComponentType{Requirements: []string{"docker"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.RequiresDocker()
			if got != tt.want {
				t.Errorf("RequiresDocker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentType_GetDockerWeight(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want int
	}{
		{
			name: "explicit docker weight",
			ct: &ComponentType{
				Pool:      "docker",
				Resources: &ComponentTypeResources{CPUs: 2, DockerWeight: 4},
			},
			want: 4,
		},
		{
			name: "defaults to CPU weight for docker pool",
			ct: &ComponentType{
				Pool:      "docker",
				Resources: &ComponentTypeResources{CPUs: 2},
			},
			want: 2,
		},
		{
			name: "host only returns zero",
			ct: &ComponentType{
				Pool:      "host",
				Resources: &ComponentTypeResources{CPUs: 4, DockerWeight: 2},
			},
			want: 0,
		},
		{
			name: "inferred docker from requirements",
			ct: &ComponentType{
				Requirements: []string{"docker"},
				Resources:    &ComponentTypeResources{CPUs: 3},
			},
			want: 3,
		},
		{
			name: "no resources defaults to 1 for docker pool",
			ct: &ComponentType{
				Pool: "docker",
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.GetDockerWeight()
			if got != tt.want {
				t.Errorf("GetDockerWeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComponentType_GetPoolAllocation(t *testing.T) {
	tests := []struct {
		name            string
		ct              *ComponentType
		wantHostWeight  int
		wantDockerWeight int
		wantIsContainer bool
	}{
		{
			name: "host only component",
			ct: &ComponentType{
				Pool:      "host",
				Resources: &ComponentTypeResources{CPUs: 2},
			},
			wantHostWeight:   2,
			wantDockerWeight: 0,
			wantIsContainer:  false,
		},
		{
			name: "container component",
			ct: &ComponentType{
				Pool:      "docker",
				Resources: &ComponentTypeResources{CPUs: 4, DockerWeight: 4},
			},
			wantHostWeight:   4,
			wantDockerWeight: 4,
			wantIsContainer:  true,
		},
		{
			name: "container with different weights",
			ct: &ComponentType{
				Pool:      "docker",
				Resources: &ComponentTypeResources{CPUs: 2, DockerWeight: 4},
			},
			wantHostWeight:   2,
			wantDockerWeight: 4,
			wantIsContainer:  true,
		},
		{
			name:             "default component",
			ct:               &ComponentType{},
			wantHostWeight:   1,
			wantDockerWeight: 0,
			wantIsContainer:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alloc := tt.ct.GetPoolAllocation()
			if alloc.HostWeight != tt.wantHostWeight {
				t.Errorf("HostWeight = %v, want %v", alloc.HostWeight, tt.wantHostWeight)
			}
			if alloc.DockerWeight != tt.wantDockerWeight {
				t.Errorf("DockerWeight = %v, want %v", alloc.DockerWeight, tt.wantDockerWeight)
			}
			if alloc.IsContainer() != tt.wantIsContainer {
				t.Errorf("IsContainer() = %v, want %v", alloc.IsContainer(), tt.wantIsContainer)
			}
		})
	}
}

func TestComponentType_GetAmp(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want float64
	}{
		{
			name: "returns explicit amp value",
			ct:   &ComponentType{Amp: 1.5},
			want: 1.5,
		},
		{
			name: "returns 1.0 when amp is zero",
			ct:   &ComponentType{Amp: 0},
			want: 1.0,
		},
		{
			name: "returns 1.0 when amp is not set",
			ct:   &ComponentType{},
			want: 1.0,
		},
		{
			name: "returns amp value of 2.0",
			ct:   &ComponentType{Amp: 2.0},
			want: 2.0,
		},
		{
			name: "returns amp value less than 1",
			ct:   &ComponentType{Amp: 0.5},
			want: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.GetAmp()
			if got != tt.want {
				t.Errorf("GetAmp() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test nil receiver
	t.Run("nil receiver returns 1.0", func(t *testing.T) {
		var ct *ComponentType
		if got := ct.GetAmp(); got != 1.0 {
			t.Errorf("nil.GetAmp() = %v, want 1.0", got)
		}
	})
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

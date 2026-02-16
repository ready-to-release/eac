package config

import (
	"testing"
)

func TestComponentType_HasToolChain(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "true when multiple builders",
			ct:   &ComponentType{Builders: []string{"preprocess", "mkdocs-pdf"}},
			want: true,
		},
		{
			name: "false when single builder",
			ct:   &ComponentType{Builders: []string{"go"}},
			want: false,
		},
		{
			name: "false when empty",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "true with three builders",
			ct:   &ComponentType{Builders: []string{"a", "b", "c"}},
			want: true,
		},
		{
			name: "nil receiver returns false",
			ct:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.HasToolChain(); got != tt.want {
				t.Errorf("HasToolChain() = %v, want %v", got, tt.want)
			}
		})
	}
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
			name: "empty component defaults to host",
			ct:   &ComponentType{},
			want: "host",
		},
		{
			name: "unset pool defaults to host",
			ct:   &ComponentType{Builders: []string{"go"}},
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
			name: "docker pool requires docker",
			ct:   &ComponentType{Pool: "docker"},
			want: true,
		},
		{
			name: "host pool does not require docker",
			ct:   &ComponentType{Pool: "host"},
			want: false,
		},
		{
			name: "empty component does not require docker",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "component with builders but no pool does not require docker",
			ct:   &ComponentType{Builders: []string{"go"}},
			want: false,
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
			name: "no resources defaults to 1 for docker pool",
			ct: &ComponentType{
				Pool: "docker",
			},
			want: 1,
		},
		{
			name: "unset pool returns zero",
			ct: &ComponentType{
				Resources: &ComponentTypeResources{CPUs: 3},
			},
			want: 0,
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
		name             string
		ct               *ComponentType
		wantHostWeight   int
		wantDockerWeight int
		wantIsContainer  bool
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

// TestComponentType_IsBuildable tests the IsBuildable method
func TestComponentType_IsBuildable(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "single builder",
			ct:   &ComponentType{Builders: []string{"go"}},
			want: true,
		},
		{
			name: "multiple builders (tool chain)",
			ct:   &ComponentType{Builders: []string{"preprocess", "pdf-oci"}},
			want: true,
		},
		{
			name: "empty builders",
			ct:   &ComponentType{Builders: []string{}},
			want: false,
		},
		{
			name: "nil builders",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsBuildable(); got != tt.want {
				t.Errorf("IsBuildable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_IsLintable tests the IsLintable method
func TestComponentType_IsLintable(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "single linter",
			ct:   &ComponentType{Linters: []string{"golangci-lint"}},
			want: true,
		},
		{
			name: "multiple linters",
			ct:   &ComponentType{Linters: []string{"eslint", "prettier"}},
			want: true,
		},
		{
			name: "empty linters",
			ct:   &ComponentType{Linters: []string{}},
			want: false,
		},
		{
			name: "nil linters",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsLintable(); got != tt.want {
				t.Errorf("IsLintable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_IsTestable tests the IsTestable method
func TestComponentType_IsTestable(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want bool
	}{
		{
			name: "single tester",
			ct:   &ComponentType{Testers: []string{"go"}},
			want: true,
		},
		{
			name: "multiple testers",
			ct:   &ComponentType{Testers: []string{"go", "integration"}},
			want: true,
		},
		{
			name: "empty testers",
			ct:   &ComponentType{Testers: []string{}},
			want: false,
		},
		{
			name: "nil testers",
			ct:   &ComponentType{},
			want: false,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.IsTestable(); got != tt.want {
				t.Errorf("IsTestable() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_GetBuilders tests the GetBuilders method
func TestComponentType_GetBuilders(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want []string
	}{
		{
			name: "single builder",
			ct:   &ComponentType{Builders: []string{"go"}},
			want: []string{"go"},
		},
		{
			name: "multi-step build chain",
			ct:   &ComponentType{Builders: []string{"mkdocs-preprocess", "pdf-oci"}},
			want: []string{"mkdocs-preprocess", "pdf-oci"},
		},
		{
			name: "empty builders",
			ct:   &ComponentType{Builders: []string{}},
			want: []string{},
		},
		{
			name: "nil builders",
			ct:   &ComponentType{},
			want: nil,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.GetBuilders(); !stringSliceEqual(got, tt.want) {
				t.Errorf("GetBuilders() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_GetLinters tests the GetLinters method
func TestComponentType_GetLinters(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want []string
	}{
		{
			name: "single linter",
			ct:   &ComponentType{Linters: []string{"golangci-lint"}},
			want: []string{"golangci-lint"},
		},
		{
			name: "multiple linters",
			ct:   &ComponentType{Linters: []string{"eslint", "prettier"}},
			want: []string{"eslint", "prettier"},
		},
		{
			name: "empty linters",
			ct:   &ComponentType{Linters: []string{}},
			want: []string{},
		},
		{
			name: "nil linters",
			ct:   &ComponentType{},
			want: nil,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.GetLinters(); !stringSliceEqual(got, tt.want) {
				t.Errorf("GetLinters() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_GetTesters tests the GetTesters method
func TestComponentType_GetTesters(t *testing.T) {
	tests := []struct {
		name string
		ct   *ComponentType
		want []string
	}{
		{
			name: "single tester",
			ct:   &ComponentType{Testers: []string{"go"}},
			want: []string{"go"},
		},
		{
			name: "multiple testers",
			ct:   &ComponentType{Testers: []string{"go", "integration"}},
			want: []string{"go", "integration"},
		},
		{
			name: "empty testers",
			ct:   &ComponentType{Testers: []string{}},
			want: []string{},
		},
		{
			name: "nil testers",
			ct:   &ComponentType{},
			want: nil,
		},
		{
			name: "nil receiver",
			ct:   nil,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.GetTesters(); !stringSliceEqual(got, tt.want) {
				t.Errorf("GetTesters() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestComponentType_ToolChainExamples validates multi-step build configurations.
func TestComponentType_ToolChainExamples(t *testing.T) {
	t.Run("docs-pdf has tool chain", func(t *testing.T) {
		ct := &ComponentType{
			Builders: []string{"mkdocs-preprocess", "pdf-oci"},
		}
		if !ct.HasToolChain() {
			t.Error("docs-pdf should have tool chain")
		}
		builders := ct.GetBuilders()
		if len(builders) != 2 {
			t.Errorf("docs-pdf should have 2 builders, got %d", len(builders))
		}
		if builders[0] != "mkdocs-preprocess" {
			t.Errorf("first builder should be mkdocs-preprocess, got %s", builders[0])
		}
		if builders[1] != "pdf-oci" {
			t.Errorf("second builder should be pdf-oci, got %s", builders[1])
		}
	})

	t.Run("docs-site has tool chain", func(t *testing.T) {
		ct := &ComponentType{
			Builders: []string{"mkdocs-preprocess", "mkdocs-render-oci"},
		}
		if !ct.HasToolChain() {
			t.Error("docs-site should have tool chain")
		}
		builders := ct.GetBuilders()
		if len(builders) != 2 {
			t.Errorf("docs-site should have 2 builders, got %d", len(builders))
		}
	})

	t.Run("single builder has no tool chain", func(t *testing.T) {
		ct := &ComponentType{
			Builders: []string{"go"},
		}
		if ct.HasToolChain() {
			t.Error("single builder should not have tool chain")
		}
		builders := ct.GetBuilders()
		if len(builders) != 1 || builders[0] != "go" {
			t.Errorf("expected [go], got %v", builders)
		}
	})
}

func TestComponentType_DeriveName(t *testing.T) {
	goKind := &ComponentType{SourcePrefixes: []string{"go", "src"}}
	tsKind := &ComponentType{SourcePrefixes: []string{"typescript", "src"}}
	structurizrKind := &ComponentType{SourcePrefixes: []string{"specs"}}
	containerKind := &ComponentType{SourcePrefixes: []string{"containers"}}
	bashKind := &ComponentType{SourcePrefixes: []string{"scripts"}}

	tests := []struct {
		name     string
		ct       *ComponentType
		rootPath string
		moniker  string
		want     string
	}{
		// Go kind
		{"go: strip go/ and moniker", goKind, "go/adapters/godog", "adapters", "godog"},
		{"go: strip go/ and moniker prefix", goKind, "go/commands/base", "commands", "base"},
		{"go: exact moniker match", goKind, "go/core", "core", "core"},
		{"go: exact moniker clibase", goKind, "go/clibase", "clibase", "clibase"},
		{"go: no moniker prefix", goKind, "go/cli/eac", "eac", "cli-eac"},
		{"go: strip src/", goKind, "src/adapters/godog", "adapters", "godog"},

		// TypeScript kind
		{"ts: strip typescript/", tsKind, "typescript/vscode-commit", "vscode-commit", "vscode-commit"},

		// Structurizr kind
		{"structurizr: strip specs/ and moniker", structurizrKind, "specs/docs/.design", "docs", "design"},

		// Container kind
		{"container: exact moniker", containerKind, "containers/eac-ext", "eac-ext", "eac-ext"},

		// Bash kind
		{"bash: strip scripts/", bashKind, "scripts/sh/go-invoker", "implicit-cli", "sh-go-invoker"},

		// Edge cases
		{"nil component type", nil, "go/core", "core", "go-core"},
		{"empty root", goKind, "", "core", ""},
		{"no source prefix match", goKind, "contracts/core/0.1.0", "contracts", "core-0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ct.DeriveName(tt.rootPath, tt.moniker)
			if got != tt.want {
				t.Errorf("DeriveName(%q, %q) = %q, want %q", tt.rootPath, tt.moniker, got, tt.want)
			}
		})
	}
}

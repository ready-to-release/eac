package workunit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UnitSpec Type Tests
// =============================================================================

func TestUnitSpec_FieldsExist(t *testing.T) {
	// Verify UnitSpec struct has all expected fields
	spec := UnitSpec{
		ID:            UnitID{Context: ContextBuild, Module: "mod", Component: "comp", Tool: "tool"},
		ComponentType: "go",
		Weight:        1,
		Container:    false,
		DependsOn:     []UnitID{},
		Cached:        false,
		Metadata:      map[string]any{"key": "value"},
	}

	assert.Equal(t, "mod", spec.ID.Module)
	assert.Equal(t, "go", spec.ComponentType)
	assert.Equal(t, 1, spec.Weight)
	assert.False(t, spec.Container)
	assert.Empty(t, spec.DependsOn)
	assert.False(t, spec.Cached)
	assert.Equal(t, "value", spec.Metadata["key"])
}

// =============================================================================
// NewBuildSpec Constructor Tests
// =============================================================================

func TestNewBuildSpec_CreatesSpecWithContextBuild(t *testing.T) {
	spec := NewBuildSpec("core", "go", "go")

	assert.Equal(t, ContextBuild, spec.ID.Context)
}

func TestNewBuildSpec_SetsCorrectFields(t *testing.T) {
	spec := NewBuildSpec("core", "go", "go")

	assert.Equal(t, "core", spec.ID.Module)
	assert.Equal(t, "go", spec.ID.Component)
	assert.Equal(t, "go", spec.ID.Tool)
}

func TestNewBuildSpec_SetsReasonableDefaults(t *testing.T) {
	spec := NewBuildSpec("core", "go", "go")

	assert.Equal(t, 1, spec.Weight, "Weight should default to 1")
	assert.False(t, spec.Cached, "Cached should default to false")
	assert.False(t, spec.Container, "Container should default to false")
	assert.Empty(t, spec.DependsOn, "DependsOn should be empty by default")
}

func TestNewBuildSpec_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		tool      string
	}{
		{
			name:      "standard go build",
			module:    "core",
			component: "go",
			tool:      "go",
		},
		{
			name:      "typescript build",
			module:    "web-app",
			component: "typescript",
			tool:      "tsc",
		},
		{
			name:      "docker build",
			module:    "eac-cli",
			component: "docker",
			tool:      "docker",
		},
		{
			name:      "module with hyphens",
			module:    "my-complex-module",
			component: "go",
			tool:      "go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewBuildSpec(tt.module, tt.component, tt.tool)

			assert.Equal(t, ContextBuild, spec.ID.Context)
			assert.Equal(t, tt.module, spec.ID.Module)
			assert.Equal(t, tt.component, spec.ID.Component)
			assert.Equal(t, tt.tool, spec.ID.Tool)
			assert.Equal(t, 1, spec.Weight)
			assert.False(t, spec.Cached)
		})
	}
}

// =============================================================================
// NewTestSpec Constructor Tests
// =============================================================================

func TestNewTestSpec_CreatesSpecWithContextTest(t *testing.T) {
	spec := NewTestSpec("core", "go", "gotest", "unit")

	assert.Equal(t, ContextTest, spec.ID.Context)
}

func TestNewTestSpec_SetsCorrectFields(t *testing.T) {
	spec := NewTestSpec("core", "go", "gotest", "unit")

	assert.Equal(t, "core", spec.ID.Module)
	assert.Equal(t, "go", spec.ID.Component)
	assert.Equal(t, "gotest", spec.ID.Tool)
}

func TestNewTestSpec_SetsTestsetInExtra(t *testing.T) {
	spec := NewTestSpec("core", "go", "gotest", "unit")

	require.NotNil(t, spec.ID.Extra)
	assert.Equal(t, "unit", spec.ID.Extra["testset"])
}

func TestNewTestSpec_SetsReasonableDefaults(t *testing.T) {
	spec := NewTestSpec("core", "go", "gotest", "unit")

	assert.Equal(t, 1, spec.Weight, "Weight should default to 1")
	assert.False(t, spec.Cached, "Cached should default to false")
	assert.False(t, spec.Container, "Container should default to false")
	assert.Empty(t, spec.DependsOn, "DependsOn should be empty by default")
}

func TestNewTestSpec_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		tool      string
		testset   string
	}{
		{
			name:      "unit tests",
			module:    "core",
			component: "go",
			tool:      "gotest",
			testset:   "unit",
		},
		{
			name:      "integration tests",
			module:    "core",
			component: "go",
			tool:      "gotest",
			testset:   "integration",
		},
		{
			name:      "acceptance tests with gherkin",
			module:    "eac-cli",
			component: "gherkin",
			tool:      "godog",
			testset:   "acceptance",
		},
		{
			name:      "e2e tests",
			module:    "web-app",
			component: "playwright",
			tool:      "playwright",
			testset:   "e2e",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewTestSpec(tt.module, tt.component, tt.tool, tt.testset)

			assert.Equal(t, ContextTest, spec.ID.Context)
			assert.Equal(t, tt.module, spec.ID.Module)
			assert.Equal(t, tt.component, spec.ID.Component)
			assert.Equal(t, tt.tool, spec.ID.Tool)
			assert.Equal(t, tt.testset, spec.ID.Extra["testset"])
			assert.Equal(t, 1, spec.Weight)
			assert.False(t, spec.Cached)
		})
	}
}

// =============================================================================
// NewLintSpec Constructor Tests
// =============================================================================

func TestNewLintSpec_CreatesSpecWithContextLint(t *testing.T) {
	spec := NewLintSpec("core", "go", "golangci-lint")

	assert.Equal(t, ContextLint, spec.ID.Context)
}

func TestNewLintSpec_SetsCorrectFields(t *testing.T) {
	spec := NewLintSpec("core", "go", "golangci-lint")

	assert.Equal(t, "core", spec.ID.Module)
	assert.Equal(t, "go", spec.ID.Component)
	assert.Equal(t, "golangci-lint", spec.ID.Tool)
}

func TestNewLintSpec_SetsReasonableDefaults(t *testing.T) {
	spec := NewLintSpec("core", "go", "golangci-lint")

	assert.Equal(t, 1, spec.Weight, "Weight should default to 1")
	assert.False(t, spec.Cached, "Cached should default to false")
	assert.False(t, spec.Container, "Container should default to false")
	assert.Empty(t, spec.DependsOn, "DependsOn should be empty by default")
}

func TestNewLintSpec_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		provider  string
	}{
		{
			name:      "golangci-lint for go",
			module:    "core",
			component: "go",
			provider:  "golangci-lint",
		},
		{
			name:      "eslint for typescript",
			module:    "web-app",
			component: "typescript",
			provider:  "eslint",
		},
		{
			name:      "yamllint for yaml",
			module:    "configs",
			component: "yaml",
			provider:  "yamllint",
		},
		{
			name:      "markdown lint",
			module:    "docs",
			component: "markdown",
			provider:  "markdownlint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewLintSpec(tt.module, tt.component, tt.provider)

			assert.Equal(t, ContextLint, spec.ID.Context)
			assert.Equal(t, tt.module, spec.ID.Module)
			assert.Equal(t, tt.component, spec.ID.Component)
			assert.Equal(t, tt.provider, spec.ID.Tool)
			assert.Equal(t, 1, spec.Weight)
			assert.False(t, spec.Cached)
		})
	}
}

// =============================================================================
// NewScanSpec Constructor Tests
// =============================================================================

func TestNewScanSpec_CreatesSpecWithContextScan(t *testing.T) {
	spec := NewScanSpec("core", "go", "trivy-vuln")

	assert.Equal(t, ContextScan, spec.ID.Context)
}

func TestNewScanSpec_SetsCorrectFields(t *testing.T) {
	spec := NewScanSpec("core", "go", "trivy-vuln")

	assert.Equal(t, "core", spec.ID.Module)
	assert.Equal(t, "go", spec.ID.Component)
	assert.Equal(t, "trivy-vuln", spec.ID.Tool)
}

func TestNewScanSpec_SetsReasonableDefaults(t *testing.T) {
	spec := NewScanSpec("core", "go", "trivy-vuln")

	assert.Equal(t, 1, spec.Weight, "Weight should default to 1")
	assert.False(t, spec.Cached, "Cached should default to false")
	assert.False(t, spec.Container, "Container should default to false")
	assert.Empty(t, spec.DependsOn, "DependsOn should be empty by default")
}

func TestNewScanSpec_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		component string
		scanner   string
	}{
		{
			name:      "trivy vulnerability scan",
			module:    "core",
			component: "go",
			scanner:   "trivy-vuln",
		},
		{
			name:      "trivy secret scan",
			module:    "core",
			component: "go",
			scanner:   "trivy-secret",
		},
		{
			name:      "semgrep SAST scan",
			module:    "web-app",
			component: "typescript",
			scanner:   "semgrep",
		},
		{
			name:      "zap DAST scan",
			module:    "api-service",
			component: "docker",
			scanner:   "zap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewScanSpec(tt.module, tt.component, tt.scanner)

			assert.Equal(t, ContextScan, spec.ID.Context)
			assert.Equal(t, tt.module, spec.ID.Module)
			assert.Equal(t, tt.component, spec.ID.Component)
			assert.Equal(t, tt.scanner, spec.ID.Tool)
			assert.Equal(t, 1, spec.Weight)
			assert.False(t, spec.Cached)
		})
	}
}

// =============================================================================
// Constructor Consistency Tests
// =============================================================================

func TestConstructors_AllSetMetadataToNonNil(t *testing.T) {
	t.Run("NewBuildSpec", func(t *testing.T) {
		spec := NewBuildSpec("mod", "comp", "tool")
		// Metadata should be initialized (not nil)
		assert.NotNil(t, spec.Metadata, "NewBuildSpec should initialize Metadata")
	})

	t.Run("NewTestSpec", func(t *testing.T) {
		spec := NewTestSpec("mod", "comp", "tool", "testset")
		assert.NotNil(t, spec.Metadata, "NewTestSpec should initialize Metadata")
	})

	t.Run("NewLintSpec", func(t *testing.T) {
		spec := NewLintSpec("mod", "comp", "provider")
		assert.NotNil(t, spec.Metadata, "NewLintSpec should initialize Metadata")
	})

	t.Run("NewScanSpec", func(t *testing.T) {
		spec := NewScanSpec("mod", "comp", "scanner")
		assert.NotNil(t, spec.Metadata, "NewScanSpec should initialize Metadata")
	})
}

func TestConstructors_AllSetComponentType(t *testing.T) {
	t.Run("NewBuildSpec sets ComponentType to component", func(t *testing.T) {
		spec := NewBuildSpec("mod", "go", "go")
		assert.Equal(t, "go", spec.ComponentType)
	})

	t.Run("NewTestSpec sets ComponentType to component", func(t *testing.T) {
		spec := NewTestSpec("mod", "gherkin", "godog", "acceptance")
		assert.Equal(t, "gherkin", spec.ComponentType)
	})

	t.Run("NewLintSpec sets ComponentType to component", func(t *testing.T) {
		spec := NewLintSpec("mod", "typescript", "eslint")
		assert.Equal(t, "typescript", spec.ComponentType)
	})

	t.Run("NewScanSpec sets ComponentType to component", func(t *testing.T) {
		spec := NewScanSpec("mod", "docker", "trivy-vuln")
		assert.Equal(t, "docker", spec.ComponentType)
	})
}

// =============================================================================
// UnitSpec DependsOn Tests
// =============================================================================

func TestUnitSpec_DependsOnCanBeSet(t *testing.T) {
	buildSpec := NewBuildSpec("core", "go", "go")
	testSpec := NewTestSpec("core", "go", "gotest", "unit")

	// Test can depend on build
	testSpec.DependsOn = []UnitID{buildSpec.ID}

	assert.Len(t, testSpec.DependsOn, 1)
	assert.Equal(t, buildSpec.ID, testSpec.DependsOn[0])
}

func TestUnitSpec_DependsOnMultiple(t *testing.T) {
	spec := NewTestSpec("eac-cli", "gherkin", "godog", "integration")

	dep1 := UnitID{Context: ContextBuild, Module: "core", Component: "go", Tool: "go"}
	dep2 := UnitID{Context: ContextBuild, Module: "eac-cli", Component: "go", Tool: "go"}

	spec.DependsOn = []UnitID{dep1, dep2}

	assert.Len(t, spec.DependsOn, 2)
}

// =============================================================================
// UnitSpec Metadata Tests
// =============================================================================

func TestUnitSpec_MetadataCanStoreAnyValues(t *testing.T) {
	spec := NewBuildSpec("mod", "go", "go")

	spec.Metadata["string_key"] = "value"
	spec.Metadata["int_key"] = 42
	spec.Metadata["bool_key"] = true
	spec.Metadata["slice_key"] = []string{"a", "b", "c"}
	spec.Metadata["map_key"] = map[string]int{"x": 1, "y": 2}

	assert.Equal(t, "value", spec.Metadata["string_key"])
	assert.Equal(t, 42, spec.Metadata["int_key"])
	assert.Equal(t, true, spec.Metadata["bool_key"])
	assert.Equal(t, []string{"a", "b", "c"}, spec.Metadata["slice_key"])
	assert.Equal(t, map[string]int{"x": 1, "y": 2}, spec.Metadata["map_key"])
}

// =============================================================================
// UnitSpec Container Tests
// =============================================================================

func TestUnitSpec_ContainerCanBeSet(t *testing.T) {
	spec := NewBuildSpec("eac-cli", "docker", "docker")
	spec.Container = true

	assert.True(t, spec.Container)
}

// =============================================================================
// UnitSpec Weight Tests
// =============================================================================

func TestUnitSpec_WeightCanBeModified(t *testing.T) {
	spec := NewBuildSpec("heavy-module", "go", "go")
	spec.Weight = 5

	assert.Equal(t, 5, spec.Weight)
}

func TestUnitSpec_WeightCanBeZeroForLightTasks(t *testing.T) {
	spec := NewLintSpec("mod", "yaml", "yamllint")
	spec.Weight = 0

	assert.Equal(t, 0, spec.Weight)
}

// =============================================================================
// UnitSpec Cached Tests
// =============================================================================

func TestUnitSpec_CachedCanBeSetToTrue(t *testing.T) {
	spec := NewBuildSpec("mod", "go", "go")
	spec.Cached = true

	assert.True(t, spec.Cached)
}

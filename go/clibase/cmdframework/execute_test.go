package cmdframework

import (
	"context"
	"io"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/workunit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUnitRegistry_RegisterAndRetrieve tests basic registration and retrieval.
func TestUnitRegistry_RegisterAndRetrieve(t *testing.T) {
	// Create a fresh registry for testing
	reg := NewUnitRegistry()

	// Create mock provider and worker
	mockProvider := func(ctx *ExecutionContext) []workunit.UnitSpec {
		return []workunit.UnitSpec{}
	}
	mockWorker := func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
		return 0
	}

	// Register for build command type
	reg.RegisterProvider(CommandTypeBuild, mockProvider)
	reg.RegisterWorker(CommandTypeBuild, mockWorker)

	// Retrieve and verify
	provider := reg.GetProvider(CommandTypeBuild)
	if provider == nil {
		t.Error("Expected provider to be registered, got nil")
	}

	worker := reg.GetWorker(CommandTypeBuild)
	if worker == nil {
		t.Error("Expected worker to be registered, got nil")
	}

	// Verify HasComponents returns true
	if !reg.HasComponents(CommandTypeBuild) {
		t.Error("Expected HasComponents to return true for registered command type")
	}
}

// TestUnitRegistry_UnregisteredCommandType tests behavior for unregistered types.
func TestUnitRegistry_UnregisteredCommandType(t *testing.T) {
	reg := NewUnitRegistry()

	// Should return nil for unregistered command type
	provider := reg.GetProvider(CommandTypeBuild)
	if provider != nil {
		t.Error("Expected nil provider for unregistered command type")
	}

	worker := reg.GetWorker(CommandTypeBuild)
	if worker != nil {
		t.Error("Expected nil worker for unregistered command type")
	}

	// HasComponents should return false
	if reg.HasComponents(CommandTypeBuild) {
		t.Error("Expected HasComponents to return false for unregistered command type")
	}
}

// TestUnitRegistry_PartialRegistration tests that HasComponents requires both.
func TestUnitRegistry_PartialRegistration(t *testing.T) {
	reg := NewUnitRegistry()

	// Register only provider
	mockProvider := func(ctx *ExecutionContext) []workunit.UnitSpec {
		return nil
	}
	reg.RegisterProvider(CommandTypeTest, mockProvider)

	// HasComponents should return false (worker not registered)
	if reg.HasComponents(CommandTypeTest) {
		t.Error("Expected HasComponents to return false when only provider is registered")
	}

	// Now register worker
	mockWorker := func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
		return 0
	}
	reg.RegisterWorker(CommandTypeTest, mockWorker)

	// Now HasComponents should return true
	if !reg.HasComponents(CommandTypeTest) {
		t.Error("Expected HasComponents to return true after both registered")
	}
}

// TestUnitRegistry_ReplaceRegistration tests that registration can be replaced.
func TestUnitRegistry_ReplaceRegistration(t *testing.T) {
	reg := NewUnitRegistry()

	callCount := 0
	provider1 := func(ctx *ExecutionContext) []workunit.UnitSpec {
		callCount = 1
		return nil
	}
	provider2 := func(ctx *ExecutionContext) []workunit.UnitSpec {
		callCount = 2
		return nil
	}

	// Register first provider
	reg.RegisterProvider(CommandTypeScan, provider1)

	// Replace with second provider
	reg.RegisterProvider(CommandTypeScan, provider2)

	// Call the provider and verify it's the second one
	retrieved := reg.GetProvider(CommandTypeScan)
	if retrieved == nil {
		t.Fatal("Expected provider to be registered")
	}
	retrieved(nil)
	if callCount != 2 {
		t.Errorf("Expected second provider to be called (callCount=2), got callCount=%d", callCount)
	}
}

// TestUnitRegistry_AllCommandTypes tests registration for all command types.
func TestUnitRegistry_AllCommandTypes(t *testing.T) {
	reg := NewUnitRegistry()

	commandTypes := []CommandType{CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint}

	for _, cmdType := range commandTypes {
		mockProvider := func(ctx *ExecutionContext) []workunit.UnitSpec {
			return nil
		}
		mockWorker := func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
			return 0
		}

		reg.RegisterProvider(cmdType, mockProvider)
		reg.RegisterWorker(cmdType, mockWorker)

		if !reg.HasComponents(cmdType) {
			t.Errorf("Expected HasComponents to return true for %s", cmdType)
		}
	}
}

// TestUnitRegistry_ConcurrentAccess tests thread safety.
func TestUnitRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewUnitRegistry()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmdType := CommandType([]CommandType{CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint}[idx%4])
			reg.RegisterProvider(cmdType, func(ctx *ExecutionContext) []workunit.UnitSpec {
				return nil
			})
			reg.RegisterWorker(cmdType, func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
				return idx
			})
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmdType := CommandType([]CommandType{CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint}[idx%4])
			_ = reg.GetProvider(cmdType)
			_ = reg.GetWorker(cmdType)
			_ = reg.HasComponents(cmdType)
		}(i)
	}

	wg.Wait()
	// If we get here without a race condition panic, the test passes
}

// --- injectModuleDependencies tests ---

func TestInjectModuleDependencies_Basic(t *testing.T) {
	// Module A depends on Module B. UoWs in A should get DependsOn for B's UoWs.
	reg := modules.NewRegistry("0.1.0", "/test")
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleA",
		DependsOn: []string{"moduleB"},
	}, "/test")))
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleB",
		DependsOn: []string{},
	}, "/test")))

	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "go", Tool: "go-build"}, Index: 0},
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "gherkin", Tool: "godog"}, Index: 1},
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleB", Component: "go", Tool: "go-build"}, Index: 2},
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleB", Component: "docs", Tool: "mkdocs"}, Index: 3},
	}

	result := injectModuleDependencies(work, reg)

	// Module A's UoWs (index 0, 1) should depend on Module B's UoWs (index 2, 3)
	assert.Len(t, result[0].DependsOn, 2, "moduleA:go should depend on 2 B UoWs")
	assert.Len(t, result[1].DependsOn, 2, "moduleA:gherkin should depend on 2 B UoWs")

	// Module B's UoWs should have no deps (it doesn't depend on anything)
	assert.Empty(t, result[2].DependsOn)
	assert.Empty(t, result[3].DependsOn)

	// Verify the dependency IDs match B's UoWs
	depModules := make(map[string]bool)
	for _, dep := range result[0].DependsOn {
		depModules[dep.Module] = true
	}
	assert.True(t, depModules["moduleB"])
}

func TestInjectModuleDependencies_DepNotInBatch(t *testing.T) {
	// Module A depends on Module B, but B is not in the execution batch
	reg := modules.NewRegistry("0.1.0", "/test")
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleA",
		DependsOn: []string{"moduleB"},
	}, "/test")))
	// Note: moduleB is registered but has no UoWs in the work slice

	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "go", Tool: "go-build"}, Index: 0},
	}

	result := injectModuleDependencies(work, reg)

	// No deps injected — moduleB has no UoWs in the batch
	assert.Empty(t, result[0].DependsOn)
}

func TestInjectModuleDependencies_NoDeps(t *testing.T) {
	// Both modules have no dependencies
	reg := modules.NewRegistry("0.1.0", "/test")
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleA",
		DependsOn: []string{},
	}, "/test")))
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleB",
		DependsOn: []string{},
	}, "/test")))

	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "go"}, Index: 0},
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleB", Component: "go"}, Index: 1},
	}

	result := injectModuleDependencies(work, reg)

	assert.Empty(t, result[0].DependsOn)
	assert.Empty(t, result[1].DependsOn)
}

func TestInjectModuleDependencies_NilRegistry(t *testing.T) {
	work := []workunit.UnitSpec{
		{ID: workunit.UnitID{Module: "moduleA", Component: "go"}, Index: 0},
	}

	result := injectModuleDependencies(work, nil)
	assert.Equal(t, work, result, "nil registry should return work unchanged")
}

func TestInjectModuleDependencies_EmptyWork(t *testing.T) {
	reg := modules.NewRegistry("0.1.0", "/test")
	result := injectModuleDependencies(nil, reg)
	assert.Nil(t, result)
}

func TestInjectModuleDependencies_PreservesExistingDeps(t *testing.T) {
	// Module A depends on Module B, and A's UoW already has an intra-module dep
	reg := modules.NewRegistry("0.1.0", "/test")
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleA",
		DependsOn: []string{"moduleB"},
	}, "/test")))
	require.NoError(t, reg.Add(modules.NewModuleContract(domain.BaseContract{
		Moniker:   "moduleB",
		DependsOn: []string{},
	}, "/test")))

	existingDep := workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "static", Tool: "copy"}
	work := []workunit.UnitSpec{
		{
			ID:        workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleA", Component: "go", Tool: "go-build"},
			Index:     0,
			DependsOn: []workunit.UnitID{existingDep},
		},
		{ID: workunit.UnitID{Context: workunit.ContextBuild, Module: "moduleB", Component: "go", Tool: "go-build"}, Index: 1},
	}

	result := injectModuleDependencies(work, reg)

	// Should have existing dep + moduleB's UoW
	assert.Len(t, result[0].DependsOn, 2)
	assert.Equal(t, existingDep, result[0].DependsOn[0], "existing dep should be preserved")
	assert.Equal(t, "moduleB", result[0].DependsOn[1].Module)
}

// TestGlobalRegistryFunctions tests the package-level registry functions.
func TestGlobalRegistryFunctions(t *testing.T) {
	// Reset global registry for test isolation
	registry = NewUnitRegistry()
	defer func() { registry = NewUnitRegistry() }()

	// Test RegisterUnitProvider and GetUnitProvider
	buildCalled := false
	RegisterUnitProvider(CommandTypeBuild, func(ctx *ExecutionContext) []workunit.UnitSpec {
		buildCalled = true
		return nil
	})

	provider := GetUnitProvider(CommandTypeBuild)
	if provider == nil {
		t.Fatal("RegisterUnitProvider should register for CommandTypeBuild")
	}
	provider(nil)
	if !buildCalled {
		t.Error("Build provider was not called")
	}

	// Test RegisterUnitWorker and GetUnitWorker
	RegisterUnitWorker(CommandTypeBuild, func(goCtx context.Context, ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
		return 42
	})

	worker := GetUnitWorker(CommandTypeBuild)
	if worker == nil {
		t.Fatal("RegisterUnitWorker should register for CommandTypeBuild")
	}
	if result := worker(context.Background(), nil, "", "", nil); result != 42 {
		t.Errorf("Expected worker to return 42, got %d", result)
	}

	// Test HasUnitExecution
	if !HasUnitExecution(CommandTypeBuild) {
		t.Error("HasUnitExecution should return true after registering both provider and worker")
	}
}

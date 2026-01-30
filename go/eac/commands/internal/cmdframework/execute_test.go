package cmdframework

import (
	"io"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/workunit"
)

// TestUnitRegistry_RegisterAndRetrieve tests basic registration and retrieval.
func TestUnitRegistry_RegisterAndRetrieve(t *testing.T) {
	// Create a fresh registry for testing
	reg := NewUnitRegistry()

	// Create mock provider and worker
	mockProvider := func(ctx *ExecutionContext) [][]workunit.UnitSpec {
		return [][]workunit.UnitSpec{{}}
	}
	mockWorker := func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
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
	mockProvider := func(ctx *ExecutionContext) [][]workunit.UnitSpec {
		return nil
	}
	reg.RegisterProvider(CommandTypeTest, mockProvider)

	// HasComponents should return false (worker not registered)
	if reg.HasComponents(CommandTypeTest) {
		t.Error("Expected HasComponents to return false when only provider is registered")
	}

	// Now register worker
	mockWorker := func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
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
	provider1 := func(ctx *ExecutionContext) [][]workunit.UnitSpec {
		callCount = 1
		return nil
	}
	provider2 := func(ctx *ExecutionContext) [][]workunit.UnitSpec {
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
		mockProvider := func(ctx *ExecutionContext) [][]workunit.UnitSpec {
			return nil
		}
		mockWorker := func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
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
			reg.RegisterProvider(cmdType, func(ctx *ExecutionContext) [][]workunit.UnitSpec {
				return nil
			})
			reg.RegisterWorker(cmdType, func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
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

// TestExecutionMode_Configuration tests the execution mode configuration.
func TestExecutionMode_Configuration(t *testing.T) {
	tests := []struct {
		cmdType      CommandType
		expectedMode ExecutionMode
	}{
		{CommandTypeBuild, ExecutionModeConfigured},
		{CommandTypeTest, ExecutionModeLayered},
		{CommandTypeScan, ExecutionModeParallel},
		{CommandTypeLint, ExecutionModeParallel},
	}

	for _, tt := range tests {
		t.Run(string(tt.cmdType), func(t *testing.T) {
			mode, exists := executionModeConfig[tt.cmdType]
			if !exists {
				t.Errorf("Expected execution mode config for %s", tt.cmdType)
				return
			}
			if mode != tt.expectedMode {
				t.Errorf("Expected mode %v for %s, got %v", tt.expectedMode, tt.cmdType, mode)
			}
		})
	}
}

// TestGlobalRegistryFunctions tests the package-level registry functions.
func TestGlobalRegistryFunctions(t *testing.T) {
	// Reset global registry for test isolation
	registry = NewUnitRegistry()
	defer func() { registry = NewUnitRegistry() }()

	// Test RegisterUnitProvider and GetUnitProvider
	buildCalled := false
	RegisterUnitProvider(CommandTypeBuild, func(ctx *ExecutionContext) [][]workunit.UnitSpec {
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
	RegisterUnitWorker(CommandTypeBuild, func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
		return 42
	})

	worker := GetUnitWorker(CommandTypeBuild)
	if worker == nil {
		t.Fatal("RegisterUnitWorker should register for CommandTypeBuild")
	}
	if result := worker(nil, "", "", nil); result != 42 {
		t.Errorf("Expected worker to return 42, got %d", result)
	}

	// Test HasUnitExecution
	if !HasUnitExecution(CommandTypeBuild) {
		t.Error("HasUnitExecution should return true after registering both provider and worker")
	}
}

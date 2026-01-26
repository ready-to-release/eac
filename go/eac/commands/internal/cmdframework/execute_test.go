package cmdframework

import (
	"io"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/internal/orchestrator"
)

// TestComponentRegistry_RegisterAndRetrieve tests basic registration and retrieval.
func TestComponentRegistry_RegisterAndRetrieve(t *testing.T) {
	// Create a fresh registry for testing
	reg := NewComponentRegistry()

	// Create mock provider and worker
	mockProvider := func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
		return [][]orchestrator.ComponentWork{{}}
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

// TestComponentRegistry_UnregisteredCommandType tests behavior for unregistered types.
func TestComponentRegistry_UnregisteredCommandType(t *testing.T) {
	reg := NewComponentRegistry()

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

// TestComponentRegistry_PartialRegistration tests that HasComponents requires both.
func TestComponentRegistry_PartialRegistration(t *testing.T) {
	reg := NewComponentRegistry()

	// Register only provider
	mockProvider := func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
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

// TestComponentRegistry_ReplaceRegistration tests that registration can be replaced.
func TestComponentRegistry_ReplaceRegistration(t *testing.T) {
	reg := NewComponentRegistry()

	callCount := 0
	provider1 := func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
		callCount = 1
		return nil
	}
	provider2 := func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
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

// TestComponentRegistry_AllCommandTypes tests registration for all command types.
func TestComponentRegistry_AllCommandTypes(t *testing.T) {
	reg := NewComponentRegistry()

	commandTypes := []CommandType{CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint}

	for _, cmdType := range commandTypes {
		mockProvider := func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
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

// TestComponentRegistry_ConcurrentAccess tests thread safety.
func TestComponentRegistry_ConcurrentAccess(t *testing.T) {
	reg := NewComponentRegistry()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cmdType := CommandType([]CommandType{CommandTypeBuild, CommandTypeTest, CommandTypeScan, CommandTypeLint}[idx%4])
			reg.RegisterProvider(cmdType, func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
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
	registry = NewComponentRegistry()
	defer func() { registry = NewComponentRegistry() }()

	// Test RegisterComponentProvider and GetComponentProvider
	buildCalled := false
	RegisterComponentProvider(CommandTypeBuild, func(ctx *ExecutionContext) [][]orchestrator.ComponentWork {
		buildCalled = true
		return nil
	})

	provider := GetComponentProvider(CommandTypeBuild)
	if provider == nil {
		t.Fatal("RegisterComponentProvider should register for CommandTypeBuild")
	}
	provider(nil)
	if !buildCalled {
		t.Error("Build provider was not called")
	}

	// Test RegisterComponentWorker and GetComponentWorker
	RegisterComponentWorker(CommandTypeBuild, func(ctx *ExecutionContext, module, component string, logWriter io.Writer) int {
		return 42
	})

	worker := GetComponentWorker(CommandTypeBuild)
	if worker == nil {
		t.Fatal("RegisterComponentWorker should register for CommandTypeBuild")
	}
	if result := worker(nil, "", "", nil); result != 42 {
		t.Errorf("Expected worker to return 42, got %d", result)
	}

	// Test HasComponentExecution
	if !HasComponentExecution(CommandTypeBuild) {
		t.Error("HasComponentExecution should return true after registering both provider and worker")
	}
}

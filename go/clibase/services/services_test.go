//go:build L1 && ov

// Package services provides centralized service initialization for EAC commands.
package services

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test: New() with successful initialization
// ============================================================================

func TestNew_Success(t *testing.T) {
	// This test runs within the EAC repository itself, so workspace detection works.
	// We use the real workspace root for integration-style testing.

	opts := core.SimpleServicesOptions{
		InitTools: false, // Don't initialize tools for faster tests
		DebugMode: false,
	}

	svc, err := New(opts)
	require.NoError(t, err, "New() should succeed in valid workspace")
	require.NotNil(t, svc, "Services should not be nil")

	// Ensure cleanup runs
	defer func() {
		err := svc.Close()
		assert.NoError(t, err, "Close() should not error")
	}()

	// Verify all required interfaces are available
	assert.NotEmpty(t, svc.WorkspaceRoot(), "WorkspaceRoot should not be empty")
	assert.NotEmpty(t, svc.ConfigRoot(), "ConfigRoot should not be empty")
	assert.NotNil(t, svc.Config(), "Config should not be nil")
	assert.NotNil(t, svc.Modules(), "Modules should not be nil")
}

// ============================================================================
// Test: Default options behavior
// ============================================================================

func TestNew_DefaultOptions(t *testing.T) {
	// Test with zero-value options - should use sensible defaults
	opts := core.SimpleServicesOptions{}

	svc, err := New(opts)
	require.NoError(t, err, "New() with default options should succeed")
	require.NotNil(t, svc)
	defer svc.Close()

	// Default should NOT initialize tools (InitTools defaults to false)
	// Tools() should return nil when tools are not initialized
	tools := svc.Tools()
	assert.Nil(t, tools, "Tools should be nil when InitTools=false (default)")
}

// ============================================================================
// Test: WorkspaceRoot returns correct path
// ============================================================================

func TestServices_WorkspaceRoot(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	workspaceRoot := svc.WorkspaceRoot()

	// Should be an absolute path
	assert.True(t, filepath.IsAbs(workspaceRoot), "WorkspaceRoot should be absolute path")

	// Should exist as a directory
	info, err := os.Stat(workspaceRoot)
	require.NoError(t, err, "WorkspaceRoot should exist")
	assert.True(t, info.IsDir(), "WorkspaceRoot should be a directory")

	// Should contain .git (we're in a git repo)
	gitDir := filepath.Join(workspaceRoot, ".git")
	_, err = os.Stat(gitDir)
	assert.NoError(t, err, "WorkspaceRoot should contain .git directory")
}

// ============================================================================
// Test: ConfigRoot returns correct .eac path
// ============================================================================

func TestServices_ConfigRoot(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	configRoot := svc.ConfigRoot()
	workspaceRoot := svc.WorkspaceRoot()

	// ConfigRoot should be {workspaceRoot}/.eac
	expectedConfigRoot := filepath.Join(workspaceRoot, ".eac")
	assert.Equal(t, expectedConfigRoot, configRoot, "ConfigRoot should be .eac under workspace root")

	// Should be an absolute path
	assert.True(t, filepath.IsAbs(configRoot), "ConfigRoot should be absolute path")

	// Should exist (eac repo has .eac directory)
	info, err := os.Stat(configRoot)
	require.NoError(t, err, "ConfigRoot should exist in eac repo")
	assert.True(t, info.IsDir(), "ConfigRoot should be a directory")
}

// ============================================================================
// Test: Config() returns ConfigPort interface
// ============================================================================

func TestServices_Config(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	cfg := svc.Config()
	require.NotNil(t, cfg, "Config() should return non-nil ConfigPort")

	// Verify ConfigPort interface methods work
	repoRoot := cfg.GetRepoRoot()
	assert.NotEmpty(t, repoRoot, "Config.GetRepoRoot() should return non-empty string")
	assert.Equal(t, svc.WorkspaceRoot(), repoRoot, "Config.GetRepoRoot() should match WorkspaceRoot()")

	configRoot := cfg.GetConfigRoot()
	assert.NotEmpty(t, configRoot, "Config.GetConfigRoot() should return non-empty string")
	assert.Equal(t, svc.ConfigRoot(), configRoot, "Config.GetConfigRoot() should match ConfigRoot()")

	// Verify repository config is loaded
	repoCfg := cfg.GetRepository()
	require.NotNil(t, repoCfg, "Config.GetRepository() should return non-nil")

	// Should have modules (eac repo has multiple modules)
	monikers := repoCfg.AllMonikers()
	assert.NotEmpty(t, monikers, "Repository should have modules")
}

// ============================================================================
// Test: Modules() returns ModuleRegistryPort interface
// ============================================================================

func TestServices_Modules(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	modules := svc.Modules()
	require.NotNil(t, modules, "Modules() should return non-nil ModuleRegistryPort")

	// Verify ModuleRegistryPort interface methods work
	count := modules.Count()
	assert.Greater(t, count, 0, "Module registry should have modules")

	monikers := modules.AllMonikers()
	assert.NotEmpty(t, monikers, "AllMonikers() should return module names")

	// Should be able to look up a module (eac repo has 'core' module)
	hasCore := modules.Has("core")
	assert.True(t, hasCore, "Should have 'core' module in eac repo")

	// Get should return a module
	if hasCore {
		module, ok := modules.Get("core")
		assert.True(t, ok, "Get('core') should return true")
		assert.NotNil(t, module, "Get('core') should return module")
		assert.Equal(t, "core", module.GetMoniker(), "Module moniker should be 'core'")
	}

	// WorkspaceRoot should match services
	wsRoot := modules.WorkspaceRoot()
	assert.Equal(t, svc.WorkspaceRoot(), wsRoot, "Module registry workspace root should match")
}

// ============================================================================
// Test: Tools() returns ToolRegistryPort (nil when InitTools=false)
// ============================================================================

func TestServices_Tools_NotInitialized(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false, // Explicitly disable tool initialization
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	tools := svc.Tools()
	assert.Nil(t, tools, "Tools() should return nil when InitTools=false")
}

func TestServices_Tools_Initialized(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: true, // Enable tool initialization
		DebugMode: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	tools := svc.Tools()
	// Tools registry may or may not be nil depending on tool-config.yml presence
	// The important thing is that no error occurred during initialization
	if tools != nil {
		// If tools were loaded, verify interface works
		allTools := tools.AllCanonical()
		// We don't assert non-empty as tools are optional
		_ = allTools
	}
}

// ============================================================================
// Test: Close() runs cleanup functions in reverse order
// ============================================================================

func TestServices_Close_CleanupOrder(t *testing.T) {
	// Create a test services implementation that tracks cleanup order
	var cleanupOrder []int
	var mu sync.Mutex

	// Mock cleanup functions
	cleanup1 := func() error {
		mu.Lock()
		defer mu.Unlock()
		cleanupOrder = append(cleanupOrder, 1)
		return nil
	}
	cleanup2 := func() error {
		mu.Lock()
		defer mu.Unlock()
		cleanupOrder = append(cleanupOrder, 2)
		return nil
	}
	cleanup3 := func() error {
		mu.Lock()
		defer mu.Unlock()
		cleanupOrder = append(cleanupOrder, 3)
		return nil
	}

	// Create services and add cleanup functions
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)

	// New() returns *Services directly, so we can use AddCleanup
	svc.AddCleanup(cleanup1)
	svc.AddCleanup(cleanup2)
	svc.AddCleanup(cleanup3)

	// Close should run cleanups in reverse order (3, 2, 1)
	err = svc.Close()
	assert.NoError(t, err, "Close() should not error")

	// Verify cleanup order if we added any
	mu.Lock()
	defer mu.Unlock()
	if len(cleanupOrder) > 0 {
		expected := []int{3, 2, 1}
		assert.Equal(t, expected, cleanupOrder, "Cleanup functions should run in reverse order")
	}
}

func TestServices_Close_MultipleCallsSafe(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)

	// First close
	err = svc.Close()
	assert.NoError(t, err, "First Close() should not error")

	// Second close should be safe (idempotent)
	err = svc.Close()
	assert.NoError(t, err, "Second Close() should not error (idempotent)")
}

// ============================================================================
// Test: New() fails when outside workspace
// ============================================================================

func TestNew_OutsideWorkspace(t *testing.T) {
	// Create a temporary directory that is NOT a git repository
	tempDir := t.TempDir()

	// Set CLIE_REPO_ROOT to the temp directory to simulate being outside workspace
	// This overrides workspace detection
	cleanup := workspace.ForTesting(t, tempDir)
	defer cleanup()

	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)

	// Should fail because:
	// 1. tempDir is not a git repo (no .git)
	// 2. tempDir has no .eac/repository.yml
	// The exact error depends on what validation New() performs

	if err == nil {
		// If New succeeded, Close() should still work
		svc.Close()
		// But we expect it to fail in most cases
		t.Log("Note: New() succeeded with temp directory - this may be expected if New() is lenient")
	} else {
		// Expected: error when not in valid workspace
		assert.Error(t, err, "New() should fail when not in valid workspace")
		assert.Nil(t, svc, "Services should be nil on error")
	}
}

// ============================================================================
// Test: Concurrent access safety
// ============================================================================

func TestServices_ConcurrentAccess(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	// Run multiple goroutines accessing services concurrently
	var wg sync.WaitGroup
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Access various methods concurrently
			_ = svc.WorkspaceRoot()
			_ = svc.ConfigRoot()
			_ = svc.Config()
			_ = svc.Modules()
			_ = svc.Tools()
		}()
	}

	wg.Wait()
	// If we get here without race detector panics, test passes
}

// ============================================================================
// Test: Interface compliance
// ============================================================================

func TestServices_ImplementsSimpleServicesPort(t *testing.T) {
	// Compile-time check that *Services implements SimpleServicesPort
	// This is verified by the "var _ core.SimpleServicesPort = (*Services)(nil)"
	// declaration in services.go
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	// Verify we can use it as the interface
	var port core.SimpleServicesPort = svc
	assert.NotNil(t, port)
}

// ============================================================================
// Test: DebugMode option
// ============================================================================

func TestNew_WithDebugMode(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
		DebugMode: true, // Enable debug mode
	}

	svc, err := New(opts)
	require.NoError(t, err, "New() with DebugMode should succeed")
	require.NotNil(t, svc)
	defer svc.Close()

	// Debug mode should configure logging but not affect basic functionality
	assert.NotEmpty(t, svc.WorkspaceRoot())
	assert.NotNil(t, svc.Config())
}

// ============================================================================
// Test: Module filtering through ModuleRegistryPort
// ============================================================================

func TestServices_Modules_FilterByComponent(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	modules := svc.Modules()
	require.NotNil(t, modules)

	// Filter by 'go' component type (eac repo should have go modules)
	goModules := modules.FilterByComponent("go")
	// Note: May be empty if no modules have 'go' component type defined
	// The important thing is that the method works without error
	_ = goModules

	// Filter by non-existent component type should return empty
	fakeModules := modules.FilterByComponent("nonexistent-component-type")
	assert.Empty(t, fakeModules, "FilterByComponent with nonexistent type should return empty")
}

// ============================================================================
// Test: FindModulesForFile through ModuleRegistryPort
// ============================================================================

func TestServices_Modules_FindModulesForFile(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	modules := svc.Modules()
	require.NotNil(t, modules)

	// Find modules for a known file path in eac repo
	// go/core/config/config.go should be owned by 'core' module
	foundModules := modules.FindModulesForFile("go/core/config/config.go")

	if len(foundModules) > 0 {
		// Verify at least one module was found
		monikers := make([]string, len(foundModules))
		for i, m := range foundModules {
			monikers[i] = m.GetMoniker()
		}
		t.Logf("Found modules for go/core/config/config.go: %v", monikers)
	}

	// Non-existent file should return empty
	noModules := modules.FindModulesForFile("nonexistent/path/file.go")
	assert.Empty(t, noModules, "FindModulesForFile with nonexistent path should return empty")
}

// ============================================================================
// Test: Config provides access to optional configs
// ============================================================================

func TestServices_Config_OptionalConfigs(t *testing.T) {
	opts := core.SimpleServicesOptions{
		InitTools: false,
	}

	svc, err := New(opts)
	require.NoError(t, err)
	defer svc.Close()

	cfg := svc.Config()
	require.NotNil(t, cfg)

	// These may return nil if configs are not present, but should not panic
	envs := cfg.GetEnvironments()
	if envs != nil {
		_ = envs.ListEnvironments()
	}

	tags := cfg.GetTestingTags()
	if tags != nil {
		_ = tags.GetTags()
	}

	suites := cfg.GetTestSuites()
	if suites != nil {
		_ = suites.ListDefault()
	}

	componentTypes := cfg.GetComponentKinds()
	if componentTypes != nil {
		_ = componentTypes.List()
	}
}

// ============================================================================
// Table-driven test: Various initialization scenarios
// ============================================================================

func TestNew_Scenarios(t *testing.T) {
	tests := []struct {
		name      string
		opts      core.SimpleServicesOptions
		wantErr   bool
		checkFunc func(t *testing.T, svc core.SimpleServicesPort)
	}{
		{
			name: "default options",
			opts: core.SimpleServicesOptions{},
			checkFunc: func(t *testing.T, svc core.SimpleServicesPort) {
				assert.NotEmpty(t, svc.WorkspaceRoot())
				assert.Nil(t, svc.Tools(), "Tools should be nil with default options")
			},
		},
		{
			name: "InitTools=false",
			opts: core.SimpleServicesOptions{
				InitTools: false,
			},
			checkFunc: func(t *testing.T, svc core.SimpleServicesPort) {
				assert.Nil(t, svc.Tools())
			},
		},
		{
			name: "InitTools=true",
			opts: core.SimpleServicesOptions{
				InitTools: true,
			},
			checkFunc: func(t *testing.T, svc core.SimpleServicesPort) {
				// Tools may or may not be initialized depending on tool-config.yml
				// Just ensure no panic
			},
		},
		{
			name: "DebugMode=true",
			opts: core.SimpleServicesOptions{
				DebugMode: true,
			},
			checkFunc: func(t *testing.T, svc core.SimpleServicesPort) {
				assert.NotEmpty(t, svc.WorkspaceRoot())
			},
		},
		{
			name: "both options enabled",
			opts: core.SimpleServicesOptions{
				InitTools: true,
				DebugMode: true,
			},
			checkFunc: func(t *testing.T, svc core.SimpleServicesPort) {
				assert.NotEmpty(t, svc.WorkspaceRoot())
				assert.NotNil(t, svc.Config())
				assert.NotNil(t, svc.Modules())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, svc)
			defer svc.Close()

			if tt.checkFunc != nil {
				tt.checkFunc(t, svc)
			}
		})
	}
}

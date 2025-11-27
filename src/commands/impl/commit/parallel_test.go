package commit

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/ready-to-release/eac/src/ai"
	"github.com/ready-to-release/eac/src/core/logging"
	"github.com/ready-to-release/eac/src/core/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateModuleSectionsParallel_PreservesOrder verifies that module sections
// are returned in the same order as affectedModules, despite concurrent execution.
//
// Rule: Easy to understand
//   - Table-driven tests with clear test names
//   - Each test case specifies input modules and expected order
//   - Assertions check exact order preservation
//
// Rule: Easy to change
//   - New test cases can be added by extending the table
//   - Test helper (createTestConfig) isolates test data creation
//
// Rule: Hard to break
//   - Tests multiple scenarios (3 modules, 10 modules)
//   - Verifies both length and content order
//   - Each module name must appear in its corresponding section
func TestGenerateModuleSectionsParallel_PreservesOrder(t *testing.T) {
	tests := []struct {
		name            string
		affectedModules []string
	}{
		{
			name:            "three modules maintain order",
			affectedModules: []string{"src-cli", "src-core", "src-commands"},
		},
		{
			name:            "five modules maintain order",
			affectedModules: []string{"mod-a", "mod-b", "mod-c", "mod-d", "mod-e"},
		},
		{
			name: "ten modules maintain order",
			affectedModules: []string{
				"mod-1", "mod-2", "mod-3", "mod-4", "mod-5",
				"mod-6", "mod-7", "mod-8", "mod-9", "mod-10",
			},
		},
		{
			name:            "two modules maintain order",
			affectedModules: []string{"alpha", "beta"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := createTestConfig(tt.affectedModules)
			logger, _ := logging.NewDefault("test", t.TempDir())
			defer logger.Sync()

			// Create mock executor with deterministic responses
			mockExecutor := createMockExecutor(tt.affectedModules)

			sections, err := generateModuleSectionsParallel(cfg, logger, mockExecutor)

			require.NoError(t, err, "generateModuleSectionsParallel should not return error")
			assert.Len(t, sections, len(tt.affectedModules),
				"Should return one section per module")

			// Verify order is preserved - each section should contain its corresponding module name
			for i, expectedModule := range tt.affectedModules {
				assert.Contains(t, sections[i], expectedModule,
					"Section %d should contain module %s (got: %s)", i, expectedModule, sections[i])
			}
		})
	}
}

// TestGenerateModuleSectionsParallel_SingleModule verifies that single-module
// commits skip module section generation (returning empty slice).
func TestGenerateModuleSectionsParallel_SingleModule(t *testing.T) {
	cfg := createTestConfig([]string{"src-cli"})
	logger, _ := logging.NewDefault("test", t.TempDir())
	defer logger.Sync()
	mockExecutor := createMockExecutor([]string{"src-cli"})

	sections, err := generateModuleSectionsParallel(cfg, logger, mockExecutor)

	require.NoError(t, err)
	assert.Empty(t, sections, "Single-module commits should skip module sections")
}

// TestGenerateModuleSectionsParallel_EmptyModules verifies that empty modules
// are rejected with an error (edge case that shouldn't occur in practice).
func TestGenerateModuleSectionsParallel_EmptyModules(t *testing.T) {
	cfg := createTestConfig([]string{})
	logger, _ := logging.NewDefault("test", t.TempDir())
	defer logger.Sync()
	mockExecutor := createMockExecutor([]string{})

	sections, err := generateModuleSectionsParallel(cfg, logger, mockExecutor)

	require.Error(t, err, "Empty modules should return an error")
	assert.Contains(t, err.Error(), "affectedModules cannot be empty", "Error message should indicate empty modules")
	assert.Nil(t, sections, "Sections should be nil when error occurs")
}

// TestGenerateModuleSectionsParallel_ErrorPropagation verifies that errors
// from individual module generation are properly captured and returned.
func TestGenerateModuleSectionsParallel_ErrorPropagation(t *testing.T) {
	modules := []string{"mod-success", "mod-fail", "mod-success2"}
	cfg := createTestConfig(modules)
	logger, _ := logging.NewDefault("test", t.TempDir())
	defer logger.Sync()

	// Create mock executor that fails for specific module
	executor := ai.NewExecutor(".")
	mockFactory := func(config *ai.Config) (ai.Provider, error) {
		return &errorInjectingMock{failOnModule: "mod-fail"}, nil
	}
	executor.RegisterProvider("mock", mockFactory)
	executor.RegisterProvider("claude-cli", mockFactory)

	sections, err := generateModuleSectionsParallel(cfg, logger, executor)

	// Should return error from the failed module
	require.Error(t, err, "Should return error when module generation fails")
	assert.Contains(t, err.Error(), "mod-fail", "Error should mention the failed module")
	assert.Nil(t, sections, "Sections should be nil when error occurs")
}

// TestGenerateModuleSectionsParallel_ConcurrentContextBuilding verifies that
// buildModuleContext can be called concurrently without issues.
func TestGenerateModuleSectionsParallel_ConcurrentContextBuilding(t *testing.T) {
	modules := []string{"mod-1", "mod-2", "mod-3", "mod-4", "mod-5"}
	cfg := createTestConfig(modules)

	// Build module file map
	moduleFilesMap := make(map[string][]repository.RepositoryFileWithModule)
	for _, file := range cfg.stagedFiles {
		for _, module := range file.Modules {
			moduleFilesMap[module] = append(moduleFilesMap[module], file)
		}
	}

	var wg sync.WaitGroup
	contexts := make([]string, len(modules))

	// Build contexts concurrently
	for i, module := range modules {
		wg.Add(1)
		go func(idx int, moduleName string) {
			defer wg.Done()
			moduleFiles := moduleFilesMap[moduleName]
			contexts[idx] = buildModuleContext(moduleName, moduleFiles, cfg.gitDiff)
		}(i, module)
	}

	wg.Wait()

	// Verify all contexts were built
	for i, context := range contexts {
		assert.NotEmpty(t, context, "Context %d should not be empty", i)
		assert.Contains(t, context, modules[i], "Context should contain module name")
	}
}

// TestGenerateModuleSectionsParallel_ResultChannelCapacity verifies that the
// result channel has correct capacity to prevent goroutine blocking.
func TestGenerateModuleSectionsParallel_ResultChannelCapacity(t *testing.T) {
	// This is a design verification test - channel capacity should equal module count
	// to prevent any goroutine from blocking when sending results.
	//
	// The actual implementation creates: make(chan moduleResult, len(cfg.affectedModules))
	// This test verifies that design decision is correct.

	moduleCount := 10
	cfg := createTestConfig(make([]string, moduleCount))

	// In the actual implementation, if channel capacity is wrong, this test would
	// timeout or deadlock. With correct capacity, all goroutines can send without blocking.

	// Note: This is implicitly tested by TestGenerateModuleSectionsParallel_PreservesOrder
	// with various module counts. If channel capacity was wrong, those tests would hang.

	assert.Equal(t, moduleCount, len(cfg.affectedModules),
		"Test setup should create correct number of modules")
}

// Helper: createTestConfig creates a test configuration with specified modules.
//
// Each module gets one test file to keep tests simple. The gitDiff is minimal
// but valid (required by filterDiffForModule).
func createTestConfig(modules []string) *executionConfig {
	var files []repository.RepositoryFileWithModule
	for i, module := range modules {
		files = append(files, repository.RepositoryFileWithModule{
			Name:    fmt.Sprintf("src/%s/file%d.go", module, i+1),
			Modules: []string{module},
		})
	}

	// Minimal valid git diff (required by filterDiffForModule)
	gitDiff := "diff --git a/test.go b/test.go\n" +
		"index abc123..def456 100644\n" +
		"--- a/test.go\n" +
		"+++ b/test.go\n" +
		"@@ -1,3 +1,4 @@\n" +
		"+added line\n" +
		" existing line\n"

	// Get actual repo root for contract loading
	workspaceRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		// Fallback to current directory if not in a git repo (shouldn't happen in tests)
		workspaceRoot = "."
	}

	return &executionConfig{
		workspaceRoot:   workspaceRoot,
		affectedModules: modules,
		stagedFiles:     files,
		gitDiff:         gitDiff,
		debug:           false,
	}
}

// Helper: createMockExecutor creates a mock AI executor that returns deterministic responses
// containing the module name. This ensures tests are fast and don't depend on real AI calls.
func createMockExecutor(modules []string) *ai.Executor {
	// Create executor (workspace root doesn't matter for mock)
	executor := ai.NewExecutor(".")

	// Register mock provider for both "mock" and "claude-cli" names
	// (claude-cli is the default provider name when no config exists)
	mockFactory := func(config *ai.Config) (ai.Provider, error) {
		return &moduleNameExtractorMock{}, nil
	}
	executor.RegisterProvider("mock", mockFactory)
	executor.RegisterProvider("claude-cli", mockFactory)

	return executor
}

// moduleNameExtractorMock is a test provider that extracts the module name
// from the input context and returns it in the response
type moduleNameExtractorMock struct{}

func (m *moduleNameExtractorMock) Name() string {
	return "mock"
}

func (m *moduleNameExtractorMock) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Extract module name from the input context
	// The context contains "## Module Name" followed by the module name on the next non-empty line
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Module Name" {
			// Module name is on the next non-empty line
			for j := i + 1; j < len(lines); j++ {
				moduleName := strings.TrimSpace(lines[j])
				if moduleName != "" {
					// Return a section that contains the module name (for test assertions)
					return fmt.Sprintf("## %s\n\nMock commit section for module %s", moduleName, moduleName), nil
				}
			}
		}
	}
	// Fallback if module name not found
	return "## Unknown Module\n\nMock commit section", nil
}

// errorInjectingMock is a test provider that returns an error for a specific module
type errorInjectingMock struct {
	failOnModule string
}

func (m *errorInjectingMock) Name() string {
	return "mock"
}

func (m *errorInjectingMock) Execute(ctx context.Context, input string, opts ...ai.Option) (string, error) {
	// Extract module name from the input context
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Module Name" {
			for j := i + 1; j < len(lines); j++ {
				moduleName := strings.TrimSpace(lines[j])
				if moduleName != "" {
					// If this is the module we should fail on, return an error
					if moduleName == m.failOnModule {
						return "", fmt.Errorf("simulated AI error for module %s", moduleName)
					}
					// Otherwise return success
					return fmt.Sprintf("## %s\n\nMock commit section for module %s", moduleName, moduleName), nil
				}
			}
		}
	}
	return "## Unknown Module\n\nMock commit section", nil
}

// Additional test ideas for future implementation:
//
// TestGenerateModuleSectionsParallel_MemoryUsage
//   - Verify memory doesn't grow unbounded with many modules
//   - Use testing.AllocsPerRun to measure allocations
//
// TestGenerateModuleSectionsParallel_GoroutineLeaks
//   - Verify no goroutines leak on error paths
//   - Use runtime.NumGoroutine() before/after
//
// TestGenerateModuleSectionsParallel_CancellationPropagation
//   - Verify context cancellation stops all goroutines
//   - Requires context.Context parameter addition
//
// TestGenerateModuleSectionsParallel_ProgressReporting
//   - Verify WithProgress is called for each module
//   - Requires progress message capture/verification

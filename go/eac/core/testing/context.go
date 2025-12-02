// Package testing provides shared test context for BDD/Godog tests.
//
// The SharedTestContext type provides a unified context that can be shared
// across all test packages without the need for complex synchronization.
// Child packages receive a pointer to the same context, so changes are
// immediately visible everywhere.
//
// Usage:
//
//	// In main test runner
//	ctx := testing.NewSharedTestContext()
//
//	// Pass pointer to child packages
//	childpkg.SetContext(ctx)
//
//	// Changes in child are visible in main
//	childpkg.RunStep() // modifies ctx.CommandOutput
//	fmt.Println(ctx.CommandOutput) // sees the change
package testing

import "fmt"

// SharedTestContext holds state shared between all test step definitions.
// This is a pointer-based shared context - all packages that receive
// a pointer to this context will see changes made by any other package.
type SharedTestContext struct {
	// Command execution results
	CommandOutput string
	ExitCode      int
	CommandError  error

	// Commit-specific fields
	TestCommitMessage string
	AffectedModules   []string

	// Validation fields (shared by commit and specs)
	ValidationErrors []string

	// Reset-specific fields
	ResetTestFile string

	// Specs-specific fields
	TestDescription string
	CreatedFiles    []string
	TestDir         string
	RemovedContracts bool

	// AI test doubles
	TestAIOutput string

	// Test isolation
	OriginalRepoRoot       string
	IsolatedTestProjectDir string

	// TestIsolation provides access to the isolation infrastructure
	// for setting mock AI responses and other test configuration
	Isolation *TestIsolation

	// Command runner function (for child packages to execute commands)
	RunCommand func(cmdLine string) error
}

// NewSharedTestContext creates a new shared test context.
func NewSharedTestContext() *SharedTestContext {
	return &SharedTestContext{
		AffectedModules:  make([]string, 0),
		ValidationErrors: make([]string, 0),
		CreatedFiles:     make([]string, 0),
	}
}

// Reset clears all fields to their zero values.
// Called at the start of each scenario.
func (c *SharedTestContext) Reset() {
	c.CommandOutput = ""
	c.ExitCode = 0
	c.CommandError = nil
	c.TestCommitMessage = ""
	c.AffectedModules = make([]string, 0)
	c.ValidationErrors = make([]string, 0)
	c.ResetTestFile = ""
	c.TestDescription = ""
	c.CreatedFiles = make([]string, 0)
	c.TestDir = ""
	c.RemovedContracts = false
	c.TestAIOutput = ""
	// Note: Don't reset OriginalRepoRoot, IsolatedTestProjectDir, or RunCommand
	// as these are set once at initialization
}

// SetIsolation configures the isolation-related fields.
func (c *SharedTestContext) SetIsolation(originalRepoRoot, isolatedDir string) {
	c.OriginalRepoRoot = originalRepoRoot
	c.IsolatedTestProjectDir = isolatedDir
	c.TestDir = isolatedDir
}

// ClearIsolation clears isolation-related fields.
func (c *SharedTestContext) ClearIsolation() {
	c.IsolatedTestProjectDir = ""
	c.TestDir = ""
	c.Isolation = nil
}

// HasOutput returns true if the context has command output or non-zero exit code.
func (c *SharedTestContext) HasOutput() bool {
	return c.CommandOutput != "" || c.ExitCode != 0
}

// AddValidationError adds an error to the validation errors list.
func (c *SharedTestContext) AddValidationError(err string) {
	c.ValidationErrors = append(c.ValidationErrors, err)
}

// AddCreatedFile tracks a file created during the test.
func (c *SharedTestContext) AddCreatedFile(path string) {
	c.CreatedFiles = append(c.CreatedFiles, path)
}

// AddAffectedModule adds a module to the affected modules list.
func (c *SharedTestContext) AddAffectedModule(module string) {
	c.AffectedModules = append(c.AffectedModules, module)
}

// SetMockAIResponse writes a mock AI response file for the test provider.
// This is used by step definitions to configure AI responses for acceptance tests.
// Returns error if isolation is not set up.
func (c *SharedTestContext) SetMockAIResponse(response string) error {
	if c.Isolation == nil {
		return fmt.Errorf("cannot set mock AI response: isolation not configured")
	}
	return c.Isolation.SetMockAIResponse(response)
}

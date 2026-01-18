// npm.go - Test handler for npm build system
package testers

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
	"github.com/ready-to-release/eac/go/eac/core/platform"
	"github.com/ready-to-release/eac/go/eac/core/testing"
)

func init() {
	// Register handler for "npm" build dependency
	// All npm-based types (vscode-ext, etc.) use this via their build_deps contract
	RegisterSystem("npm", TestNpmModule)
}

// TestNpmModule is the test handler for npm-based modules.
// It runs Mocha tests with tag filtering based on the test suite configuration.
// Tags are embedded in describe() names: describe('@L0 ComponentName', ...)
func TestNpmModule(module *modules.ModuleContract, workspaceRoot, outputDir string, logWriter io.Writer, reportFormat, suiteName string) int {
	Writeln(logWriter, "\n=== Testing typescript: %s ===", module.Moniker)
	Writeln(logWriter, "Suite: %s", suiteName)

	// Get the typescript package root
	tsRoot := module.GetComponentRoot("typescript")
	if tsRoot == "" {
		Writeln(logWriter, "⚠️  No typescript package found, skipping npm tests")
		return 0
	}
	moduleRoot := filepath.Join(workspaceRoot, tsRoot)

	// Check for package.json
	packageJSON := filepath.Join(moduleRoot, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		Writeln(logWriter, "⚠️  No package.json found, skipping npm tests")
		return 0
	}

	// Check for test directory
	testDir := filepath.Join(moduleRoot, "test")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		Writeln(logWriter, "⚠️  No test directory found, skipping npm tests")
		return 0
	}

	// Get suite configuration for tag filtering
	suite, err := testing.GetSuite(suiteName)
	if err != nil {
		Writeln(logWriter, "❌ Failed to get suite '%s': %v", suiteName, err)
		return 1
	}

	// Build grep pattern from suite selectors
	grepPattern := buildMochaGrepPattern(suite)

	Writeln(logWriter, "Module root: %s", moduleRoot)
	Writeln(logWriter, "Test directory: %s", testDir)
	if grepPattern != "" {
		Writeln(logWriter, "Tag filter: %s", grepPattern)
	}
	Writeln(logWriter, "")

	// Build npm test command with tag filtering via Mocha's --grep
	// Use platform-aware command wrapper for Windows compatibility
	args := []string{"test"}
	if grepPattern != "" {
		// Pass grep pattern to Mocha: npm test -- --grep "@L0|@L1"
		args = append(args, "--", "--grep", grepPattern)
	}

	Writeln(logWriter, "Running: npm %s", strings.Join(args, " "))
	Writeln(logWriter, "")

	// Execute npm test
	wrappedName, wrappedArgs := platform.WrapCommand("npm", args...)
	exitCode := RunTestCommand(moduleRoot, logWriter, wrappedName, wrappedArgs...)

	return exitCode
}

// buildMochaGrepPattern converts suite selectors to Mocha --grep pattern.
// Mocha uses regex matching on describe/it names.
// Example: --grep "@L0|@L1" matches tests with @L0 or @L1 in their describe name.
func buildMochaGrepPattern(suite *testing.TestSuite) string {
	if suite == nil || len(suite.Selectors) == 0 {
		return ""
	}

	// Use first selector's any_of_tags as the grep pattern
	// Format: "@L0|@L1|@L2" for OR matching
	selector := suite.Selectors[0]
	if len(selector.AnyOfTags) == 0 {
		return ""
	}

	return strings.Join(selector.AnyOfTags, "|")
}

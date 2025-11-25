// Package tests provides BDD step definitions for the commit command.
//
// This file contains mock setup and cleanup functions for git repository testing.
package tests

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/commit"
	"github.com/ready-to-release/eac/src/core/git"
)

// SetupCommitMocks sets up git and AI mocks for isolated testing.
// Called automatically when @env:isolated-test-project tag is present.
func SetupCommitMocks() error {
	if IsolatedTestProjectDir == "" {
		return nil // Not in isolated mode
	}

	// Set up mock git repository
	TestMockRepo = git.NewMockRepository(IsolatedTestProjectDir).
		WithCurrentBranch("main").
		WithHeadSHA("abc1234567890")
	commit.SetGitRepo(TestMockRepo)

	// Load and set mock AI response
	// Note: filename has typo "reponse" instead of "response" - keeping as-is
	mockResponsePath := filepath.Join(OriginalRepoRoot,
		"src/commands/impl/commit/tests/assets/mock-reponse.txt")
	mockResponse, err := os.ReadFile(mockResponsePath)
	if err == nil {
		commit.SetMockAIResponse(string(mockResponse))
	}

	return nil
}

// CleanupCommitMocks resets all mocks.
func CleanupCommitMocks() {
	commit.ResetGitRepo()
	commit.ResetMockAIResponse()
	TestMockRepo = nil
}

// setupInMemoryGitRepo creates a mock git repository for testing.
func setupInMemoryGitRepo() error {
	if TestMockRepo == nil {
		return SetupCommitMocks()
	}
	return nil
}

// createModuleStructure sets up mock module files.
func createModuleStructure(modules []string) error {
	if TestMockRepo == nil {
		return fmt.Errorf("mock repo not initialized")
	}
	var trackedFiles []string
	for _, mod := range modules {
		trackedFiles = append(trackedFiles, fmt.Sprintf("src/%s/module.yml", mod))
	}
	TestMockRepo.WithTrackedFiles(trackedFiles)
	return nil
}

// commitModuleStructure simulates initial commit.
func commitModuleStructure() error {
	// Mock already has the structure set up
	return nil
}

// stageFileInModule stages a file in a module.
func stageFileInModule(module, filename, content string) error {
	if TestMockRepo == nil {
		return fmt.Errorf("mock repo not initialized")
	}
	filePath := fmt.Sprintf("src/%s/%s", module, filename)

	// Add to staged files
	currentStaged, _ := TestMockRepo.StagedFiles()
	TestMockRepo.WithStagedFiles(append(currentStaged, filePath))

	// Set up mock diff for this file
	diff := fmt.Sprintf(`diff --git a/%s b/%s
new file mode 100644
--- /dev/null
+++ b/%s
@@ -0,0 +1 @@
+%s`, filePath, filePath, filePath, content)

	currentDiff, _ := TestMockRepo.StagedDiff()
	if currentDiff != "" {
		diff = currentDiff + "\n" + diff
	}
	TestMockRepo.WithStagedDiff(diff)

	// Set up mock diff stats
	stats := fmt.Sprintf(" %s | 1 +\n 1 file changed, 1 insertion(+)", filePath)
	TestMockRepo.WithStagedDiffStats(stats)

	return nil
}

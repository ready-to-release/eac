// Package testing provides test utilities including isolated test environments.
//
// The TestIsolation type provides a unified way to create isolated test environments
// for BDD/Godog tests that need to operate on a "fake" repository without affecting
// the real repository.
//
// Key features:
//   - Creates temporary directory for test isolation
//   - Sets CLIE_REPO_ROOT and CLIE_PWD environment variables (no physical .git needed)
//   - Provides hooks for mock injection (git repository, AI responses)
//   - Handles cleanup automatically
//
// Usage:
//
//	isolation := testing.NewTestIsolation().
//	    WithOriginalRepoRoot(originalRoot).
//	    WithCopyContracts(true)
//
//	if err := isolation.Setup(); err != nil {
//	    return err
//	}
//	defer isolation.Cleanup()
//
//	// Access the isolated directory
//	testDir := isolation.IsolatedDir()
//
//	// Get environment variables to pass to subprocesses
//	env := isolation.Environment()
package testing

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/core/domain"
	"github.com/ready-to-release/eac/go/core/paths"
)

// TestIsolation provides an isolated test environment for BDD tests.
// It creates a temporary directory and sets environment variables so that
// repository discovery returns the isolated directory instead of the real repo.
type TestIsolation struct {
	// Configuration (set before Setup)
	originalRepoRoot   string
	copyContracts      bool
	copySpecs          bool
	copyAIContracts    bool
	copyMkdocsConfig   bool
	createMockAIConfig bool
	mockAIResponse     string // AI response to write to mock file

	// State (set by Setup)
	isolatedDir string
	cleanedUp   bool
}

// NewTestIsolation creates a new test isolation builder.
func NewTestIsolation() *TestIsolation {
	return &TestIsolation{}
}

// WithOriginalRepoRoot sets the original repository root path.
// This is needed for copying contracts and other resources to the isolated dir.
func (t *TestIsolation) WithOriginalRepoRoot(root string) *TestIsolation {
	t.originalRepoRoot = root
	return t
}

// WithCopyContracts enables copying the repository config to the isolated dir.
// This is needed for commands that validate against module contracts, tags, etc.
func (t *TestIsolation) WithCopyContracts(copy bool) *TestIsolation {
	t.copyContracts = copy
	return t
}

// WithCopySpecs enables copying the specs directory to the isolated dir.
func (t *TestIsolation) WithCopySpecs(copy bool) *TestIsolation {
	t.copySpecs = copy
	return t
}

// WithCopyAIContracts enables copying the .eac/ai/ directory (AI configs, prompts) to the isolated dir.
// This is needed for commands that use AI prompts from .eac/ai/.
func (t *TestIsolation) WithCopyAIContracts(copy bool) *TestIsolation {
	t.copyAIContracts = copy
	return t
}

// WithCopyMkdocsConfig enables copying MkDocs container templates to the isolated dir.
// This copies containers/mkdocs-render-oci/ and containers/pdf-oci/ which contain the
// mkdocs.yml templates used by the build system.
func (t *TestIsolation) WithCopyMkdocsConfig(copy bool) *TestIsolation {
	t.copyMkdocsConfig = copy
	return t
}

// WithMockAIConfig creates a test ai-provider.yml in the isolated environment.
// This configures the "test" provider which reads mock responses from files.
func (t *TestIsolation) WithMockAIConfig(create bool) *TestIsolation {
	t.createMockAIConfig = create
	return t
}

// WithMockAIResponse sets the mock AI response that will be written to .clie/test/ai-mock.txt.
// The "test" provider reads this file to return predictable responses in acceptance tests.
// Must be called before Setup().
func (t *TestIsolation) WithMockAIResponse(response string) *TestIsolation {
	t.mockAIResponse = response
	return t
}

// Setup creates the isolated test environment.
// Call Cleanup() when done to remove the temporary directory.
func (t *TestIsolation) Setup() error {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "isolated-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	t.isolatedDir = tmpDir

	// Initialize git repository
	if err := t.initGitRepository(); err != nil {
		t.Cleanup()
		return fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Copy repository config if requested
	if t.copyContracts && t.originalRepoRoot != "" {
		srcContracts := filepath.Join(t.originalRepoRoot, domain.EACConfigRelPath)
		dstContracts := filepath.Join(t.isolatedDir, domain.EACConfigRelPath)
		if err := copyDir(srcContracts, dstContracts); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy repository config: %w", err)
		}
	}

	// Copy specs if requested
	if t.copySpecs && t.originalRepoRoot != "" {
		srcSpecs := filepath.Join(t.originalRepoRoot, "specs")
		dstSpecs := filepath.Join(t.isolatedDir, "specs")
		if err := copyDir(srcSpecs, dstSpecs); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy specs: %w", err)
		}
	}

	// Copy AI contracts and config if requested
	if t.copyAIContracts && t.originalRepoRoot != "" {
		// Copy contracts/ directory (JSON schemas)
		srcContracts := filepath.Join(t.originalRepoRoot, "contracts")
		dstContracts := filepath.Join(t.isolatedDir, "contracts")
		if err := copyDir(srcContracts, dstContracts); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy AI contracts: %w", err)
		}

		// Copy .eac/ directory (unified ai-config.yml and prompts)
		srcCLIEEac := filepath.Join(t.originalRepoRoot, ".eac")
		dstCLIEEac := filepath.Join(t.isolatedDir, ".eac")
		if err := copyDir(srcCLIEEac, dstCLIEEac); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy .eac config: %w", err)
		}
	}

	// Copy MkDocs container templates if requested
	// These are used by the build system to generate mkdocs.yml dynamically
	if t.copyMkdocsConfig && t.originalRepoRoot != "" {
		// Copy mkdocs-render-oci container (for HTML builds and serve docs)
		srcSite := filepath.Join(t.originalRepoRoot, "containers", "mkdocs-render-oci")
		if _, err := os.Stat(srcSite); err == nil {
			dstSite := filepath.Join(t.isolatedDir, "containers", "mkdocs-render-oci")
			if err := copyDir(srcSite, dstSite); err != nil {
				t.Cleanup()
				return fmt.Errorf("failed to copy mkdocs-render-oci container: %w", err)
			}
		}
		// Copy pdf-oci container (for PDF builds)
		srcPdf := filepath.Join(t.originalRepoRoot, "containers", "pdf-oci")
		if _, err := os.Stat(srcPdf); err == nil {
			dstPdf := filepath.Join(t.isolatedDir, "containers", "pdf-oci")
			if err := copyDir(srcPdf, dstPdf); err != nil {
				t.Cleanup()
				return fmt.Errorf("failed to copy pdf-oci container: %w", err)
			}
		}
	}

	// Create test AI config if requested
	if t.createMockAIConfig {
		// Use "test" provider which reads mock responses from .clie/test/ai-mock.txt
		testConfig := `# Test AI configuration for acceptance testing
# Uses the "test" provider which reads responses from .clie/test/ai-mock.txt
ai:
  provider: test
  model: test-model
git:
  token: ""
`
		configDir := paths.EACConfigPath(t.isolatedDir)
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to create .eac directory: %w", err)
		}

		// Remove any personal config that may have been copied (it takes precedence)
		personalConfigPath := filepath.Join(configDir, "ai-provider.personal.yml")
		os.Remove(personalConfigPath) // Ignore error - file may not exist

		configPath := filepath.Join(configDir, "ai-provider.yml")
		if err := os.WriteFile(configPath, []byte(testConfig), 0o644); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to create test ai-provider.yml: %w", err)
		}
	}

	// Write mock AI response file if provided
	if t.mockAIResponse != "" {
		mockPath := paths.AITestMockPath(t.isolatedDir)
		mockDir := filepath.Dir(mockPath)
		if err := os.MkdirAll(mockDir, 0o755); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to create .clie/test directory: %w", err)
		}
		if err := os.WriteFile(mockPath, []byte(t.mockAIResponse), 0o644); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to create ai-mock.txt: %w", err)
		}
	}

	return nil
}

// initGitRepository initializes a git repository in the isolated directory using go-git.
// This is required for tests that run git commands like `git add`, `git commit`, etc.
// Uses in-process go-git instead of exec.Command for ~10x speedup on Windows.
func (t *TestIsolation) initGitRepository() error {
	if t.isolatedDir == "" {
		return fmt.Errorf("isolated directory not set")
	}

	// GitInit handles: init, config user.name/email, initial commit, branch rename to "main"
	if _, err := GitInit(t.isolatedDir); err != nil {
		return err
	}

	// Set up a self-referencing "origin" remote for testing work-pull and other remote-dependent commands
	if err := GitAddRemote(t.isolatedDir, "origin", t.isolatedDir); err != nil {
		return err
	}

	return nil
}

// Cleanup removes the temporary directory.
// Safe to call multiple times.
func (t *TestIsolation) Cleanup() {
	if t.cleanedUp || t.isolatedDir == "" {
		return
	}
	os.RemoveAll(t.isolatedDir)
	t.cleanedUp = true
	t.isolatedDir = ""
}

// IsolatedDir returns the path to the isolated test directory.
// Returns empty string if Setup hasn't been called.
func (t *TestIsolation) IsolatedDir() string {
	return t.isolatedDir
}

// Environment returns environment variables that should be set for subprocesses
// to use the isolated directory as the repository root.
func (t *TestIsolation) Environment() []string {
	if t.isolatedDir == "" {
		return nil
	}
	return []string{
		fmt.Sprintf("CLIE_PWD=%s", t.isolatedDir),
		fmt.Sprintf("CLIE_REPO_ROOT=%s", t.isolatedDir),
	}
}

// AppendToEnvironment appends the isolation environment variables to an existing
// environment slice (typically from os.Environ()).
func (t *TestIsolation) AppendToEnvironment(env []string) []string {
	if t.isolatedDir == "" {
		return env
	}
	return append(env, t.Environment()...)
}

// SetMockAIResponse writes a mock AI response file after Setup() has been called.
// This is useful for step definitions that need to configure mock responses
// between scenario setup and command execution.
// Returns error if isolation is not set up.
func (t *TestIsolation) SetMockAIResponse(response string) error {
	if t.isolatedDir == "" {
		return fmt.Errorf("cannot set mock AI response: isolation not set up")
	}

	mockPath := paths.AITestMockPath(t.isolatedDir)
	mockDir := filepath.Dir(mockPath)
	if err := os.MkdirAll(mockDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .clie/test directory: %w", err)
	}
	if err := os.WriteFile(mockPath, []byte(response), 0o644); err != nil {
		return fmt.Errorf("failed to write ai-mock.txt: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory.
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

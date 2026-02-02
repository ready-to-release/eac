// Package testing provides test utilities including isolated test environments.
//
// The TestIsolation type provides a unified way to create isolated test environments
// for BDD/Godog tests that need to operate on a "fake" repository without affecting
// the real repository.
//
// Key features:
//   - Creates temporary directory for test isolation
//   - Sets R2R_REPO_ROOT and R2R_PWD environment variables (no physical .git needed)
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
	"os/exec"
	"path/filepath"
	"strings"

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
// This copies containers/site-render-tool/ and containers/pdf-render-tool/ which contain the
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

// WithMockAIResponse sets the mock AI response that will be written to .r2r/test/ai-mock.txt.
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
		srcR2REac := filepath.Join(t.originalRepoRoot, ".eac")
		dstR2REac := filepath.Join(t.isolatedDir, ".eac")
		if err := copyDir(srcR2REac, dstR2REac); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to copy .eac config: %w", err)
		}
	}

	// Copy MkDocs container templates if requested
	// These are used by the build system to generate mkdocs.yml dynamically
	if t.copyMkdocsConfig && t.originalRepoRoot != "" {
		// Copy site-render-tool container (for HTML builds and serve docs)
		srcSite := filepath.Join(t.originalRepoRoot, "containers", "site-render-tool")
		if _, err := os.Stat(srcSite); err == nil {
			dstSite := filepath.Join(t.isolatedDir, "containers", "site-render-tool")
			if err := copyDir(srcSite, dstSite); err != nil {
				t.Cleanup()
				return fmt.Errorf("failed to copy site-render-tool container: %w", err)
			}
		}
		// Copy pdf-render-tool container (for PDF builds)
		srcPdf := filepath.Join(t.originalRepoRoot, "containers", "pdf-render-tool")
		if _, err := os.Stat(srcPdf); err == nil {
			dstPdf := filepath.Join(t.isolatedDir, "containers", "pdf-render-tool")
			if err := copyDir(srcPdf, dstPdf); err != nil {
				t.Cleanup()
				return fmt.Errorf("failed to copy pdf-render-tool container: %w", err)
			}
		}
	}

	// Create test AI config if requested
	if t.createMockAIConfig {
		// Use "test" provider which reads mock responses from .r2r/test/ai-mock.txt
		testConfig := `# Test AI configuration for acceptance testing
# Uses the "test" provider which reads responses from .r2r/test/ai-mock.txt
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
			return fmt.Errorf("failed to create .r2r/test directory: %w", err)
		}
		if err := os.WriteFile(mockPath, []byte(t.mockAIResponse), 0o644); err != nil {
			t.Cleanup()
			return fmt.Errorf("failed to create ai-mock.txt: %w", err)
		}
	}

	return nil
}

// initGitRepository initializes a git repository in the isolated directory.
// This is required for tests that run git commands like `git add`, `git commit`, etc.
// Ensures the repository is always on "main" branch for consistency across git versions.
func (t *TestIsolation) initGitRepository() error {
	if t.isolatedDir == "" {
		return fmt.Errorf("isolated directory not set")
	}

	// Try to initialize with main branch (git >= 2.28)
	cmd := exec.Command("git", "init", "--initial-branch=main")
	cmd.Dir = t.isolatedDir
	if err := cmd.Run(); err != nil {
		// Fall back to regular init for older git versions
		cmd = exec.Command("git", "init")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git init failed: %w (output: %s)", err, string(output))
		}
	}

	// Configure git user for commits (required for git commit)
	if err := t.gitConfig("user.name", "Test User"); err != nil {
		return err
	}
	if err := t.gitConfig("user.email", "test@example.com"); err != nil {
		return err
	}

	// Check current branch and normalize to "main" if needed
	// This handles older git versions that create "master" by default
	currentBranch, err := t.getCurrentBranch()
	if err != nil {
		return err
	}

	if currentBranch != "main" {
		// We're on wrong branch (probably "master" from old git)
		// Create initial commit first, then rename branch to "main"
		readmePath := filepath.Join(t.isolatedDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0o644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}

		cmd = exec.Command("git", "add", "README.md")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add failed: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit failed: %w (output: %s)", err, string(output))
		}

		// Rename current branch to "main"
		cmd = exec.Command("git", "branch", "-M", "main")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git branch -M main failed: %w (output: %s)", err, string(output))
		}
	} else {
		// Already on main, just create initial commit
		// This allows tests to run `git reset HEAD .` and other HEAD-dependent commands
		readmePath := filepath.Join(t.isolatedDir, "README.md")
		if err := os.WriteFile(readmePath, []byte("# Test Repository\n"), 0o644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}

		cmd = exec.Command("git", "add", "README.md")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git add failed: %w (output: %s)", err, string(output))
		}

		cmd = exec.Command("git", "commit", "-m", "Initial commit")
		cmd.Dir = t.isolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("git commit failed: %w (output: %s)", err, string(output))
		}
	}

	// Set up a self-referencing "origin" remote for testing work-pull and other remote-dependent commands
	// This allows work-pull to fetch from "origin/main" even in isolated tests
	cmd = exec.Command("git", "remote", "add", "origin", t.isolatedDir)
	cmd.Dir = t.isolatedDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git remote add origin failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// gitConfig sets a git configuration value in the isolated repository.
func (t *TestIsolation) gitConfig(key, value string) error {
	cmd := exec.Command("git", "config", key, value)
	cmd.Dir = t.isolatedDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config %s failed: %w (output: %s)", key, err, string(output))
	}
	return nil
}

// getCurrentBranch returns the current branch name in the isolated repository.
func (t *TestIsolation) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = t.isolatedDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w (output: %s)", err, string(output))
	}
	return strings.TrimSpace(string(output)), nil
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
		fmt.Sprintf("R2R_PWD=%s", t.isolatedDir),
		fmt.Sprintf("R2R_REPO_ROOT=%s", t.isolatedDir),
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
		return fmt.Errorf("failed to create .r2r/test directory: %w", err)
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

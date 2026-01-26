package builders

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/core/config"
)

func TestExecutePostBuildSteps_NoConfig(t *testing.T) {
	// When there's no global config, should return 0 (success)
	config.ResetGlobalForTesting()
	defer config.ResetGlobalForTesting()

	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestExecutePostBuildSteps_CopyTo(t *testing.T) {
	// Setup: create temp workspace with output files
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "out", "build", "test-module", "go")
	targetDir := filepath.Join(tmpDir, "target")

	// Create output directory with test files
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	testFile := filepath.Join(outputDir, "app.exe")
	if err := os.WriteFile(testFile, []byte("test binary"), 0o755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create a mock config with post_build configuration
	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test-module",
					Name:    "Test Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/test-module",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyTo: "target",
								},
							},
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	// Verify file was copied
	copiedFile := filepath.Join(targetDir, "app.exe")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist after copy", copiedFile)
	}

	// Verify content is correct
	content, err := os.ReadFile(copiedFile)
	if err != nil {
		t.Fatalf("failed to read copied file: %v", err)
	}
	if string(content) != "test binary" {
		t.Errorf("expected content 'test binary', got %s", string(content))
	}
}

func TestExecutePostBuildSteps_CleansTargetDirectory(t *testing.T) {
	// Setup: create temp workspace with existing stale files in target
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "out", "build", "test-module", "go")
	targetDir := filepath.Join(tmpDir, "target")

	// Create output directory with new file
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}
	newFile := filepath.Join(outputDir, "new.exe")
	if err := os.WriteFile(newFile, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	// Create target directory with stale file
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	staleFile := filepath.Join(targetDir, "stale.exe")
	if err := os.WriteFile(staleFile, []byte("stale binary"), 0o755); err != nil {
		t.Fatalf("failed to create stale file: %v", err)
	}

	// Create mock config
	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test-module",
					Name:    "Test Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/test-module",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyTo: "target",
								},
							},
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	// Verify stale file was removed
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Errorf("expected stale file %s to be removed", staleFile)
	}

	// Verify new file was copied
	copiedFile := filepath.Join(targetDir, "new.exe")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist after copy", copiedFile)
	}
}

func TestExecutePostBuildSteps_ModuleNotFound(t *testing.T) {
	// When module doesn't exist in config, should return 0 (no-op)
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "other-module",
					Name:    "Other Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/other-module",
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("nonexistent-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for nonexistent module, got %d", exitCode)
	}
}

func TestExecutePostBuildSteps_ComponentNotFound(t *testing.T) {
	// When component doesn't exist for module, should return 0 (no-op)
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test-module",
					Name:    "Test Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/test-module",
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "typescript", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for nonexistent component, got %d", exitCode)
	}
}

func TestExecutePostBuildSteps_NoPostBuildConfig(t *testing.T) {
	// When component exists but has no post_build config, should return 0
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test-module",
					Name:    "Test Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/test-module",
							Build: &config.ModuleBuild{
								Handler: "go",
								// No PostBuild config
							},
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode != 0 {
		t.Errorf("expected exit code 0 for component without post_build, got %d", exitCode)
	}
}

func TestExecutePostBuildSteps_PathEscapeAttempt(t *testing.T) {
	// When copy_to tries to escape workspace, should return error
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "test-module",
					Name:    "Test Module",
					Components: config.ModuleComponents{
						"go": &config.ComponentEntry{
							Root: "go/test-module",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyTo: "../outside-workspace", // Trying to escape
								},
							},
						},
					},
				},
			},
		},
	}
	config.SetGlobalForTesting(cfg)
	defer config.ResetGlobalForTesting()

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)

	if exitCode == 0 {
		t.Errorf("expected non-zero exit code for path escape attempt, got 0")
	}

	if !strings.Contains(logBuf.String(), "must be within workspace") {
		t.Errorf("expected error message about workspace, got: %s", logBuf.String())
	}
}

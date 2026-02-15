package builders

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/core/config"
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

func TestExecutePostBuildSteps_FallsBackToSourceWhenFrameworkDirHasOnlyManifest(t *testing.T) {
	// Regression: in-place builders (npm) write output to <component_root>/out/,
	// while the tracker writes uow.manifest.json to the framework output dir.
	// findBuildOutputDir must ignore the manifest and fall back to component source output.
	tmpDir := t.TempDir()

	// Framework output dir: only contains uow.manifest.json (written by tracker)
	frameworkOutputDir := filepath.Join(tmpDir, "out", "build", "vscode-commit", "typescript-npm-build")
	if err := os.MkdirAll(frameworkOutputDir, 0o755); err != nil {
		t.Fatalf("failed to create framework output dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(frameworkOutputDir, "uow.manifest.json"), []byte(`{"exit_code":0}`), 0o644); err != nil {
		t.Fatalf("failed to create manifest: %v", err)
	}

	// Component source output: the real compiled files (npm builds in-place)
	componentRoot := filepath.Join(tmpDir, "typescript", "vscode-commit")
	sourceOutputDir := filepath.Join(componentRoot, "out")
	if err := os.MkdirAll(sourceOutputDir, 0o755); err != nil {
		t.Fatalf("failed to create source output dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceOutputDir, "extension.js"), []byte("compiled js"), 0o644); err != nil {
		t.Fatalf("failed to create extension.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceOutputDir, "extension.js.map"), []byte("source map"), 0o644); err != nil {
		t.Fatalf("failed to create source map: %v", err)
	}

	// Also create package.json in the component root (for copy_files)
	if err := os.WriteFile(filepath.Join(componentRoot, "package.json"), []byte(`{"name":"vscode-commit"}`), 0o644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "vscode-commit",
					Name:    "VSCode Extension - Git Commit",
					Components: config.ModuleComponents{
						"typescript": &config.ComponentEntry{
							Root: "typescript/vscode-commit",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyTo: ".vscode/extensions/vscode-commit/out",
									CopyFiles: []config.CopyFileEntry{
										{
											From: "package.json",
											To:   ".vscode/extensions/vscode-commit/package.json",
										},
									},
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
	exitCode := ExecutePostBuildSteps("vscode-commit", "typescript", tmpDir, frameworkOutputDir, &logBuf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	// Verify extension.js was copied from source output, not the manifest
	targetDir := filepath.Join(tmpDir, ".vscode", "extensions", "vscode-commit", "out")
	content, err := os.ReadFile(filepath.Join(targetDir, "extension.js"))
	if err != nil {
		t.Fatalf("extension.js not found in target: %v", err)
	}
	if string(content) != "compiled js" {
		t.Errorf("expected 'compiled js', got %q", string(content))
	}

	// Verify uow.manifest.json was NOT copied to target
	if _, err := os.Stat(filepath.Join(targetDir, "uow.manifest.json")); !os.IsNotExist(err) {
		t.Errorf("uow.manifest.json should not be copied to post-build target")
	}

	// Verify copy_files: package.json
	pkgContent, err := os.ReadFile(filepath.Join(tmpDir, ".vscode", "extensions", "vscode-commit", "package.json"))
	if err != nil {
		t.Fatalf("package.json not found in target: %v", err)
	}
	if string(pkgContent) != `{"name":"vscode-commit"}` {
		t.Errorf("unexpected package.json content: %q", string(pkgContent))
	}
}

func TestHasNonLogFiles_IgnoresUoWManifest(t *testing.T) {
	// Regression: hasNonLogFiles must not count uow.manifest.json as real build output
	tmpDir := t.TempDir()

	// Dir with only uow.manifest.json
	manifestOnly := filepath.Join(tmpDir, "manifest-only")
	if err := os.MkdirAll(manifestOnly, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestOnly, "uow.manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasNonLogFiles(manifestOnly) {
		t.Error("dir with only uow.manifest.json should return false")
	}

	// Dir with uow.manifest.json AND a log file
	manifestAndLog := filepath.Join(tmpDir, "manifest-and-log")
	if err := os.MkdirAll(manifestAndLog, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestAndLog, "uow.manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestAndLog, "build.log"), []byte("log"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasNonLogFiles(manifestAndLog) {
		t.Error("dir with only uow.manifest.json and log should return false")
	}

	// Dir with uow.manifest.json AND a real file
	manifestAndReal := filepath.Join(tmpDir, "manifest-and-real")
	if err := os.MkdirAll(manifestAndReal, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestAndReal, "uow.manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(manifestAndReal, "app.exe"), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !hasNonLogFiles(manifestAndReal) {
		t.Error("dir with uow.manifest.json and app.exe should return true")
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

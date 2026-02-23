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

func TestExecutePostBuildSteps_GlobPattern(t *testing.T) {
	// Glob copy_files entry copies matched files into target directory
	tmpDir := t.TempDir()

	// Create component source with out/ directory containing compiled files
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

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "vscode-commit",
					Components: config.ModuleComponents{
						"typescript": &config.ComponentEntry{
							Root: "typescript/vscode-commit",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyFiles: []config.CopyFileEntry{
										{
											From: "out/**/*",
											To:   ".vscode/extensions/vscode-commit/out",
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

	outputDir := filepath.Join(tmpDir, "out", "build", "vscode-commit", "typescript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("vscode-commit", "typescript", tmpDir, outputDir, &logBuf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	targetDir := filepath.Join(tmpDir, ".vscode", "extensions", "vscode-commit", "out")
	content, err := os.ReadFile(filepath.Join(targetDir, "extension.js"))
	if err != nil {
		t.Fatalf("extension.js not found in target: %v", err)
	}
	if string(content) != "compiled js" {
		t.Errorf("expected 'compiled js', got %q", string(content))
	}

	mapContent, err := os.ReadFile(filepath.Join(targetDir, "extension.js.map"))
	if err != nil {
		t.Fatalf("extension.js.map not found in target: %v", err)
	}
	if string(mapContent) != "source map" {
		t.Errorf("expected 'source map', got %q", string(mapContent))
	}
}

func TestExecutePostBuildSteps_GlobPreservesSubdirs(t *testing.T) {
	// Glob should preserve subdirectory structure under the static prefix
	tmpDir := t.TempDir()

	componentRoot := filepath.Join(tmpDir, "typescript", "mymodule")
	if err := os.MkdirAll(filepath.Join(componentRoot, "out", "sub", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(componentRoot, "out", "index.js"), []byte("root"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(componentRoot, "out", "sub", "helper.js"), []byte("helper"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(componentRoot, "out", "sub", "deep", "util.js"), []byte("util"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "mymodule",
					Components: config.ModuleComponents{
						"typescript": &config.ComponentEntry{
							Root: "typescript/mymodule",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyFiles: []config.CopyFileEntry{
										{From: "out/**/*", To: "target/out"},
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

	outputDir := filepath.Join(tmpDir, "out", "build", "mymodule", "typescript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("mymodule", "typescript", tmpDir, outputDir, &logBuf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	targetDir := filepath.Join(tmpDir, "target", "out")

	// Check files at each level
	for _, tc := range []struct {
		path    string
		content string
	}{
		{"index.js", "root"},
		{filepath.Join("sub", "helper.js"), "helper"},
		{filepath.Join("sub", "deep", "util.js"), "util"},
	} {
		got, err := os.ReadFile(filepath.Join(targetDir, tc.path))
		if err != nil {
			t.Errorf("file %s not found: %v", tc.path, err)
			continue
		}
		if string(got) != tc.content {
			t.Errorf("file %s: expected %q, got %q", tc.path, tc.content, string(got))
		}
	}
}

func TestExecutePostBuildSteps_GlobNoMatches(t *testing.T) {
	// Glob with no matches should succeed (0 files copied)
	tmpDir := t.TempDir()

	componentRoot := filepath.Join(tmpDir, "typescript", "mymodule")
	if err := os.MkdirAll(componentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	// No out/ directory — glob should match nothing

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "mymodule",
					Components: config.ModuleComponents{
						"typescript": &config.ComponentEntry{
							Root: "typescript/mymodule",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyFiles: []config.CopyFileEntry{
										{From: "out/**/*", To: "target/out"},
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

	outputDir := filepath.Join(tmpDir, "out", "build", "mymodule", "typescript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("mymodule", "typescript", tmpDir, outputDir, &logBuf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	if !strings.Contains(logBuf.String(), "copied 0 files") {
		t.Errorf("expected log to mention 0 files, got: %s", logBuf.String())
	}
}

func TestExecutePostBuildSteps_GlobWithLiteralCopyFiles(t *testing.T) {
	// Test the vscode-commit scenario: glob for out/**/* plus literal for package.json
	tmpDir := t.TempDir()

	componentRoot := filepath.Join(tmpDir, "typescript", "vscode-commit")
	sourceOutputDir := filepath.Join(componentRoot, "out")
	if err := os.MkdirAll(sourceOutputDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceOutputDir, "extension.js"), []byte("compiled js"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(componentRoot, "package.json"), []byte(`{"name":"vscode-commit"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &config.EACConfig{
		RepoRoot: tmpDir,
		Repository: &config.RepositoryConfig{
			Modules: []config.Module{
				{
					Moniker: "vscode-commit",
					Components: config.ModuleComponents{
						"typescript": &config.ComponentEntry{
							Root: "typescript/vscode-commit",
							Build: &config.ModuleBuild{
								PostBuild: &config.PostBuildConfig{
									CopyFiles: []config.CopyFileEntry{
										{
											From: "out/**/*",
											To:   ".vscode/extensions/vscode-commit/out",
										},
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

	outputDir := filepath.Join(tmpDir, "out", "build", "vscode-commit", "typescript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatal(err)
	}

	var logBuf bytes.Buffer
	exitCode := ExecutePostBuildSteps("vscode-commit", "typescript", tmpDir, outputDir, &logBuf)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d; log: %s", exitCode, logBuf.String())
	}

	// Verify glob result
	content, err := os.ReadFile(filepath.Join(tmpDir, ".vscode", "extensions", "vscode-commit", "out", "extension.js"))
	if err != nil {
		t.Fatalf("extension.js not found: %v", err)
	}
	if string(content) != "compiled js" {
		t.Errorf("expected 'compiled js', got %q", string(content))
	}

	// Verify literal result
	pkgContent, err := os.ReadFile(filepath.Join(tmpDir, ".vscode", "extensions", "vscode-commit", "package.json"))
	if err != nil {
		t.Fatalf("package.json not found: %v", err)
	}
	if string(pkgContent) != `{"name":"vscode-commit"}` {
		t.Errorf("unexpected package.json content: %q", string(pkgContent))
	}
}

func TestExecutePostBuildSteps_PathEscapeAttempt(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("failed to create output dir: %v", err)
	}

	// Create component root so copy_files doesn't skip
	componentRoot := filepath.Join(tmpDir, "go", "test-module")
	if err := os.MkdirAll(componentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(componentRoot, "file.txt"), []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("literal target escape", func(t *testing.T) {
		cfg := &config.EACConfig{
			RepoRoot: tmpDir,
			Repository: &config.RepositoryConfig{
				Modules: []config.Module{
					{
						Moniker: "test-module",
						Components: config.ModuleComponents{
							"go": &config.ComponentEntry{
								Root: "go/test-module",
								Build: &config.ModuleBuild{
									PostBuild: &config.PostBuildConfig{
										CopyFiles: []config.CopyFileEntry{
											{From: "file.txt", To: "../outside-workspace/file.txt"},
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
		exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)
		if exitCode == 0 {
			t.Errorf("expected non-zero exit code for path escape attempt")
		}
		if !strings.Contains(logBuf.String(), "must be within workspace") {
			t.Errorf("expected workspace error, got: %s", logBuf.String())
		}
	})

	t.Run("glob target escape", func(t *testing.T) {
		cfg := &config.EACConfig{
			RepoRoot: tmpDir,
			Repository: &config.RepositoryConfig{
				Modules: []config.Module{
					{
						Moniker: "test-module",
						Components: config.ModuleComponents{
							"go": &config.ComponentEntry{
								Root: "go/test-module",
								Build: &config.ModuleBuild{
									PostBuild: &config.PostBuildConfig{
										CopyFiles: []config.CopyFileEntry{
											{From: "*.txt", To: "../outside-workspace"},
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
		exitCode := ExecutePostBuildSteps("test-module", "go", tmpDir, outputDir, &logBuf)
		if exitCode == 0 {
			t.Errorf("expected non-zero exit code for glob path escape attempt")
		}
		if !strings.Contains(logBuf.String(), "must be within workspace") {
			t.Errorf("expected workspace error, got: %s", logBuf.String())
		}
	})
}

func TestGlobStaticPrefix(t *testing.T) {
	tests := []struct {
		pattern  string
		expected string
	}{
		{"out/**/*", "out"},
		{"**/*", ""},
		{"src/lib/**/*.js", "src/lib"},
		{"*.txt", ""},
		{"dist/bundle.js", "dist/bundle.js"},
		{"a/b/c/**", "a/b/c"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := globStaticPrefix(tt.pattern)
			if got != tt.expected {
				t.Errorf("globStaticPrefix(%q) = %q, want %q", tt.pattern, got, tt.expected)
			}
		})
	}
}

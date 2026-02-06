//go:build L0 && ov
// +build L0,ov

// File: go/cli/eac/impl/init/scanner_test.go
package init

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ready-to-release/eac/go/core/environments"
)

// TestScanRepository_SingleModule tests scanning single-module repositories
func TestScanRepository_SingleModule(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		wantModules   int
		wantLanguage  string
		wantBuildTool string
		wantFileCount int
		wantErr       bool
	}{
		{
			name: "single Go module",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "go", "myservice")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "go.mod"),
					[]byte("module github.com/test/myservice\n\ngo 1.21\n"),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "main.go"),
					[]byte("package main\n\nfunc main() {}\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "go",
			wantBuildTool: "go",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single Python module with pyproject.toml",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "python", "myapp")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "pyproject.toml"),
					[]byte("[project]\nname = \"myapp\"\n"),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "main.py"),
					[]byte("print('hello')\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "python",
			wantBuildTool: "pip",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single Python module with setup.py",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "python", "oldapp")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "setup.py"),
					[]byte("from setuptools import setup\nsetup(name='oldapp')\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "python",
			wantBuildTool: "pip",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single Rust module",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "rust", "cli-tool")
				srcDir := filepath.Join(modDir, "src")
				require.NoError(t, os.MkdirAll(srcDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "Cargo.toml"),
					[]byte("[package]\nname = \"cli-tool\"\nversion = \"0.1.0\"\n"),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(srcDir, "main.rs"),
					[]byte("fn main() {}\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "rust",
			wantBuildTool: "cargo",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single TypeScript module",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "typescript", "webapp")
				srcDir := filepath.Join(modDir, "src")
				require.NoError(t, os.MkdirAll(srcDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "package.json"),
					[]byte(`{"name": "webapp", "version": "1.0.0"}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "tsconfig.json"),
					[]byte(`{"compilerOptions": {"target": "ES2020"}}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(srcDir, "index.ts"),
					[]byte("console.log('hello');\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "typescript",
			wantBuildTool: "npm",
			wantFileCount: 2,
			wantErr:       false,
		},
		{
			name: "single .NET module with csproj",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "dotnet", "webapi")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "webapi.csproj"),
					[]byte(`<Project Sdk="Microsoft.NET.Sdk.Web"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "Program.cs"),
					[]byte("var builder = WebApplication.CreateBuilder(args);\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "dotnet",
			wantBuildTool: "dotnet",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single .NET module with fsproj",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "dotnet", "fsharpapp")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "fsharpapp.fsproj"),
					[]byte(`<Project Sdk="Microsoft.NET.Sdk"><PropertyGroup><TargetFramework>net8.0</TargetFramework></PropertyGroup></Project>`),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "dotnet",
			wantBuildTool: "dotnet",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single Java module with pom.xml",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "java", "backend")
				srcDir := filepath.Join(modDir, "src", "main", "java")
				require.NoError(t, os.MkdirAll(srcDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "pom.xml"),
					[]byte(`<?xml version="1.0"?><project><modelVersion>4.0.0</modelVersion><groupId>com.test</groupId><artifactId>backend</artifactId></project>`),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "java",
			wantBuildTool: "maven",
			wantFileCount: 1,
			wantErr:       false,
		},
		{
			name: "single Java module with build.gradle",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "java", "gradleapp")
				srcDir := filepath.Join(modDir, "src", "main", "java")
				require.NoError(t, os.MkdirAll(srcDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "build.gradle"),
					[]byte("plugins { id 'java' }\n"),
					0644,
				))
				return tmpDir
			},
			wantModules:   1,
			wantLanguage:  "java",
			wantBuildTool: "gradle",
			wantFileCount: 1,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := tt.setupFunc(t)

			result, err := ScanRepository(repoRoot)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Modules, tt.wantModules, "expected %d modules", tt.wantModules)

			if len(result.Modules) > 0 {
				mod := result.Modules[0]
				assert.Equal(t, tt.wantLanguage, mod.Language, "expected language %s", tt.wantLanguage)
				assert.Equal(t, tt.wantBuildTool, mod.BuildTool, "expected build tool %s", tt.wantBuildTool)
				assert.Len(t, mod.Files, tt.wantFileCount, "expected %d files", tt.wantFileCount)
				assert.NotEmpty(t, mod.Name, "module name should not be empty")
				assert.NotEmpty(t, mod.Root, "module root should not be empty")
			}
		})
	}
}

// TestScanRepository_MultiModule tests scanning multi-module repositories
func TestScanRepository_MultiModule(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string
		wantModules   int
		wantLanguages map[string]bool // Languages that should be present
		wantErr       bool
	}{
		{
			name: "multi-module Go and TypeScript",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Go module
				goDir := filepath.Join(tmpDir, "go", "api-service")
				require.NoError(t, os.MkdirAll(goDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(goDir, "go.mod"),
					[]byte("module github.com/test/api-service\n\ngo 1.21\n"),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(goDir, "main.go"),
					[]byte("package main\n\nfunc main() {}\n"),
					0644,
				))

				// TypeScript module
				tsDir := filepath.Join(tmpDir, "typescript", "web-app")
				srcDir := filepath.Join(tsDir, "src")
				require.NoError(t, os.MkdirAll(srcDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(tsDir, "package.json"),
					[]byte(`{"name": "web-app", "version": "1.0.0"}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(tsDir, "tsconfig.json"),
					[]byte(`{"compilerOptions": {"target": "ES2020"}}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(srcDir, "index.ts"),
					[]byte("console.log('hello');\n"),
					0644,
				))

				return tmpDir
			},
			wantModules: 2,
			wantLanguages: map[string]bool{
				"go":         true,
				"typescript": true,
			},
			wantErr: false,
		},
		{
			name: "multi-module Go, Python, and .NET",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Go module
				goDir := filepath.Join(tmpDir, "go", "api-service")
				require.NoError(t, os.MkdirAll(goDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(goDir, "go.mod"),
					[]byte("module github.com/test/api-service\n"),
					0644,
				))

				// Python module
				pyDir := filepath.Join(tmpDir, "python", "ml-service")
				require.NoError(t, os.MkdirAll(pyDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(pyDir, "pyproject.toml"),
					[]byte("[project]\nname = \"ml-service\"\n"),
					0644,
				))

				// .NET module
				dotnetDir := filepath.Join(tmpDir, "dotnet", "web-api")
				require.NoError(t, os.MkdirAll(dotnetDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(dotnetDir, "web-api.csproj"),
					[]byte(`<Project Sdk="Microsoft.NET.Sdk.Web"></Project>`),
					0644,
				))

				return tmpDir
			},
			wantModules: 3,
			wantLanguages: map[string]bool{
				"go":     true,
				"python": true,
				"dotnet": true,
			},
			wantErr: false,
		},
		{
			name: "full-stack repository with all languages",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Go
				goDir := filepath.Join(tmpDir, "go", "api-gateway")
				require.NoError(t, os.MkdirAll(goDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(goDir, "go.mod"), []byte("module api\n"), 0644))

				// Python
				pyDir := filepath.Join(tmpDir, "python", "ml-engine")
				require.NoError(t, os.MkdirAll(pyDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(pyDir, "pyproject.toml"), []byte("[project]\n"), 0644))

				// Rust
				rustDir := filepath.Join(tmpDir, "rust", "cli-tools")
				require.NoError(t, os.MkdirAll(rustDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(rustDir, "Cargo.toml"), []byte("[package]\n"), 0644))

				// TypeScript
				tsDir := filepath.Join(tmpDir, "typescript", "web-frontend")
				require.NoError(t, os.MkdirAll(tsDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(tsDir, "package.json"), []byte(`{}`), 0644))
				require.NoError(t, os.WriteFile(filepath.Join(tsDir, "tsconfig.json"), []byte(`{}`), 0644))

				// .NET
				dotnetDir := filepath.Join(tmpDir, "dotnet", "core-services")
				require.NoError(t, os.MkdirAll(dotnetDir, 0755))
				require.NoError(t, os.WriteFile(filepath.Join(dotnetDir, "core-services.csproj"), []byte(`<Project/>`), 0644))

				return tmpDir
			},
			wantModules: 5,
			wantLanguages: map[string]bool{
				"go":         true,
				"python":     true,
				"rust":       true,
				"typescript": true,
				"dotnet":     true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := tt.setupFunc(t)

			result, err := ScanRepository(repoRoot)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Modules, tt.wantModules, "expected %d modules", tt.wantModules)

			// Verify all expected languages are present
			foundLanguages := make(map[string]bool)
			for _, mod := range result.Modules {
				foundLanguages[mod.Language] = true
			}

			for lang := range tt.wantLanguages {
				assert.True(t, foundLanguages[lang], "expected to find language: %s", lang)
			}
		})
	}
}

// TestScanRepository_EdgeCases tests edge cases and error conditions
func TestScanRepository_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T) string
		wantErr   bool
		wantCount int
	}{
		{
			name: "empty repository - no modules",
			setupFunc: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "repository with only README",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(tmpDir, "README.md"),
					[]byte("# Test Repo\n"),
					0644,
				))
				return tmpDir
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "nested modules - should detect both",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Parent module
				parentDir := filepath.Join(tmpDir, "services")
				require.NoError(t, os.MkdirAll(parentDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(parentDir, "go.mod"),
					[]byte("module services\n"),
					0644,
				))

				// Nested module
				nestedDir := filepath.Join(parentDir, "nested")
				require.NoError(t, os.MkdirAll(nestedDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(nestedDir, "go.mod"),
					[]byte("module nested\n"),
					0644,
				))

				return tmpDir
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "vendor directory - should be filtered",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Real module
				modDir := filepath.Join(tmpDir, "go", "app")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "go.mod"),
					[]byte("module app\n"),
					0644,
				))

				// Vendor directory (should be ignored)
				vendorDir := filepath.Join(modDir, "vendor", "github.com", "fake")
				require.NoError(t, os.MkdirAll(vendorDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(vendorDir, "go.mod"),
					[]byte("module fake\n"),
					0644,
				))

				return tmpDir
			},
			wantErr:   false,
			wantCount: 1, // Only the real module, not vendor
		},
		{
			name: "node_modules directory - should be filtered",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Real module
				modDir := filepath.Join(tmpDir, "typescript", "app")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "package.json"),
					[]byte(`{"name": "app"}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "tsconfig.json"),
					[]byte(`{}`),
					0644,
				))

				// node_modules directory (should be ignored)
				nodeModulesDir := filepath.Join(modDir, "node_modules", "@types", "node")
				require.NoError(t, os.MkdirAll(nodeModulesDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(nodeModulesDir, "package.json"),
					[]byte(`{"name": "@types/node"}`),
					0644,
				))

				return tmpDir
			},
			wantErr:   false,
			wantCount: 1, // Only the real module, not node_modules
		},
		{
			name: "target directory (Rust) - should be filtered",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()

				// Real module
				modDir := filepath.Join(tmpDir, "rust", "app")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "Cargo.toml"),
					[]byte("[package]\nname = \"app\"\n"),
					0644,
				))

				// target directory (should be ignored)
				targetDir := filepath.Join(modDir, "target", "debug")
				require.NoError(t, os.MkdirAll(targetDir, 0755))

				return tmpDir
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "JavaScript without TypeScript - plain package.json",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "js", "app")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "package.json"),
					[]byte(`{"name": "app", "version": "1.0.0"}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "index.js"),
					[]byte("console.log('hello');\n"),
					0644,
				))
				return tmpDir
			},
			wantErr:   false,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := tt.setupFunc(t)

			result, err := ScanRepository(repoRoot)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result.Modules, tt.wantCount, "expected %d modules", tt.wantCount)
		})
	}
}

// TestScanRepository_RealTestRepositories tests against actual test repositories
func TestScanRepository_RealTestRepositories(t *testing.T) {
	// Get repo root
	repoRoot := os.Getenv(environments.EnvR2RRepoRoot)
	if repoRoot == "" {
		// Try to find it relative to this file
		cwd, err := os.Getwd()
		require.NoError(t, err)
		// Navigate up from go/eac/commands/impl/init to repo root
		repoRoot = filepath.Join(cwd, "..", "..", "..", "..", "..")
	}

	testReposBase := filepath.Join(repoRoot, "templates", "test-repositories")

	// Skip if test repositories don't exist
	if _, err := os.Stat(testReposBase); os.IsNotExist(err) {
		t.Skip("Test repositories not found, skipping real repository tests")
	}

	tests := []struct {
		name          string
		repoPath      string
		wantModules   int
		wantLanguages map[string]bool
	}{
		{
			name:        "single-go-module",
			repoPath:    filepath.Join(testReposBase, "single-go-module"),
			wantModules: 0, // No go.mod in this test repo structure
		},
		{
			name:        "multi-go-typescript",
			repoPath:    filepath.Join(testReposBase, "multi-go-typescript"),
			wantModules: 2,
			wantLanguages: map[string]bool{
				"go":         true,
				"typescript": true,
			},
		},
		{
			name:        "multi-python-typescript",
			repoPath:    filepath.Join(testReposBase, "multi-python-typescript"),
			wantModules: 2,
			wantLanguages: map[string]bool{
				"python":     true,
				"typescript": true,
			},
		},
		{
			name:        "multi-go-python-dotnet",
			repoPath:    filepath.Join(testReposBase, "multi-go-python-dotnet"),
			wantModules: 3,
			wantLanguages: map[string]bool{
				"go":     true,
				"python": true,
				"dotnet": true,
			},
		},
		{
			name:        "multi-full-stack",
			repoPath:    filepath.Join(testReposBase, "multi-full-stack"),
			wantModules: 5,
			wantLanguages: map[string]bool{
				"go":         true,
				"python":     true,
				"rust":       true,
				"typescript": true,
				"dotnet":     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if specific test repo doesn't exist
			if _, err := os.Stat(tt.repoPath); os.IsNotExist(err) {
				t.Skipf("Test repository not found: %s", tt.repoPath)
			}

			result, err := ScanRepository(tt.repoPath)
			require.NoError(t, err)
			require.NotNil(t, result)

			assert.Len(t, result.Modules, tt.wantModules, "expected %d modules in %s", tt.wantModules, tt.name)

			if tt.wantLanguages != nil {
				foundLanguages := make(map[string]bool)
				for _, mod := range result.Modules {
					foundLanguages[mod.Language] = true
					t.Logf("Found module: %s (language: %s, tool: %s, root: %s)",
						mod.Name, mod.Language, mod.BuildTool, mod.Root)
				}

				for lang := range tt.wantLanguages {
					assert.True(t, foundLanguages[lang], "expected to find language: %s", lang)
				}
			}
		})
	}
}

// TestModuleInfo_NameInference tests that module names are correctly inferred
func TestModuleInfo_NameInference(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) string
		expectedName string
	}{
		{
			name: "Go module name from go.mod",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "go", "myservice")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "go.mod"),
					[]byte("module github.com/org/myservice\n\ngo 1.21\n"),
					0644,
				))
				return tmpDir
			},
			expectedName: "myservice",
		},
		{
			name: "Python module name from pyproject.toml",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "python", "ml-app")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "pyproject.toml"),
					[]byte("[project]\nname = \"ml-app\"\nversion = \"0.1.0\"\n"),
					0644,
				))
				return tmpDir
			},
			expectedName: "ml-app",
		},
		{
			name: "TypeScript module name from package.json",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "ts", "frontend")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "package.json"),
					[]byte(`{"name": "frontend", "version": "1.0.0"}`),
					0644,
				))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "tsconfig.json"),
					[]byte(`{}`),
					0644,
				))
				return tmpDir
			},
			expectedName: "frontend",
		},
		{
			name: "Rust module name from Cargo.toml",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				modDir := filepath.Join(tmpDir, "rust", "cli-tool")
				require.NoError(t, os.MkdirAll(modDir, 0755))
				require.NoError(t, os.WriteFile(
					filepath.Join(modDir, "Cargo.toml"),
					[]byte("[package]\nname = \"cli-tool\"\nversion = \"0.1.0\"\n"),
					0644,
				))
				return tmpDir
			},
			expectedName: "cli-tool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := tt.setupFunc(t)

			result, err := ScanRepository(repoRoot)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Modules, 1)

			assert.Equal(t, tt.expectedName, result.Modules[0].Name)
		})
	}
}

// TestScanRepository_InvalidPath tests error handling for invalid paths
func TestScanRepository_InvalidPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "non-existent directory",
			path:    "/nonexistent/path/that/does/not/exist",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ScanRepository(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

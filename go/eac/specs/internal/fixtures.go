// Package internal provides shared helpers for godog BDD tests.
//
// This file contains fixture helpers for creating complete test environments.
// Fixtures combine templates with additional setup (source files, git state, etc.)
// to create ready-to-use test scenarios.
//
// Design Principles:
// 1. Fixtures are higher-level than templates - they create complete test environments
// 2. Fixtures handle both config files AND source files
// 3. Fixtures can optionally set up git state (staged files, commits)
// 4. Each fixture is documented with what it creates
package internal

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ============================================================================
// Module Fixture Helpers
// ============================================================================

// CreateGoModule creates a complete Go module fixture in the test environment.
// Creates:
//   - Module directory at the specified path
//   - A basic Go source file (main.go or lib.go)
//   - Optionally stages the files in git
//
// Parameters:
//   - modulePath: Path to module root (e.g., "go/test-module")
//   - packageName: Go package name (e.g., "testmodule")
//   - isLibrary: If true, creates a library; if false, creates main package
//   - stage: If true, stages the files in git
func CreateGoModule(ctx *TestContext, modulePath, packageName string, isLibrary, stage bool) error {
	// Create module directory
	if err := CreateDirectory(ctx, modulePath); err != nil {
		return fmt.Errorf("failed to create module directory %s: %w", modulePath, err)
	}

	// Create source file
	var filename, content string
	if isLibrary {
		filename = "lib.go"
		content = fmt.Sprintf("package %s\n\n// Package %s provides library functionality.\n", packageName, packageName)
	} else {
		filename = "main.go"
		content = fmt.Sprintf("package main\n\nfunc main() {\n\t// %s entry point\n}\n", packageName)
	}

	sourceFile := filepath.Join(modulePath, filename)
	if err := CreateFile(ctx, sourceFile, content); err != nil {
		return fmt.Errorf("failed to create source file %s: %w", sourceFile, err)
	}

	// Stage if requested
	if stage && ctx.IsolatedDir != "" {
		cmd := exec.Command("git", "add", sourceFile)
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage %s: %w (output: %s)", sourceFile, err, string(output))
		}
	}

	return nil
}

// ============================================================================
// EAC Configuration Fixture Helpers
// ============================================================================

// SetupEACConfig creates a complete EAC configuration using a named template.
// This is the primary entry point for setting up test fixtures.
//
// Example:
//
//	SetupEACConfig(ctx, "minimal", TemplateParams{
//	    "MODULE_NAME": "test-module",
//	    "MODULE_PATH": "go/test-module",
//	})
func SetupEACConfig(ctx *TestContext, templateName string, params TemplateParams) error {
	// Ensure .r2r/eac directory exists
	if err := CreateDirectory(ctx, ".r2r/eac"); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	// Apply the template
	return ApplyTemplate(ctx, templateName, params)
}

// SetupMinimalEACConfig creates a minimal EAC configuration with a single module.
// This is a convenience wrapper around SetupEACConfig for the most common case.
func SetupMinimalEACConfig(ctx *TestContext, moduleName, modulePath string) error {
	return SetupEACConfig(ctx, "minimal", TemplateParams{
		"MODULE_NAME": moduleName,
		"MODULE_PATH": modulePath,
	})
}

// SetupMinimalGoConfig creates a minimal EAC configuration for Go development.
// Includes module-types and system-dependencies for Go.
func SetupMinimalGoConfig(ctx *TestContext, moduleName, modulePath string) error {
	return SetupEACConfig(ctx, "minimal-go", TemplateParams{
		"MODULE_NAME": moduleName,
		"MODULE_PATH": modulePath,
	})
}

// SetupMultiModuleConfig creates EAC configuration with two modules.
func SetupMultiModuleConfig(ctx *TestContext, module1Name, module1Path, module2Name, module2Path string) error {
	return SetupEACConfig(ctx, "multi-module", TemplateParams{
		"MODULE1_NAME": module1Name,
		"MODULE1_PATH": module1Path,
		"MODULE2_NAME": module2Name,
		"MODULE2_PATH": module2Path,
	})
}

// ============================================================================
// Complete Test Environment Fixtures
// ============================================================================

// SetupGoModuleWithEAC creates a complete test environment with EAC config and a Go module.
// This combines EAC configuration setup with module source file creation.
//
// Creates:
//   - .r2r/eac/modules.yml with the module definition
//   - .r2r/eac/module-types.yml with go-library type
//   - Module directory with Go source file
//   - Optionally stages all files in git
func SetupGoModuleWithEAC(ctx *TestContext, moduleName string, stage bool) error {
	modulePath := "go/" + moduleName

	// Setup EAC config
	if err := SetupMinimalEACConfig(ctx, moduleName, modulePath); err != nil {
		return err
	}

	// Create the Go module source
	if err := CreateGoModule(ctx, modulePath, moduleName, true, stage); err != nil {
		return err
	}

	return nil
}

// SetupTwoGoModulesWithEAC creates a test environment with two Go modules.
func SetupTwoGoModulesWithEAC(ctx *TestContext, module1, module2 string, stage bool) error {
	module1Path := "go/" + module1
	module2Path := "go/" + module2

	// Setup EAC config for two modules
	if err := SetupMultiModuleConfig(ctx, module1, module1Path, module2, module2Path); err != nil {
		return err
	}

	// Create both modules
	if err := CreateGoModule(ctx, module1Path, module1, true, stage); err != nil {
		return err
	}
	if err := CreateGoModule(ctx, module2Path, module2, true, stage); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Individual Config File Helpers
// ============================================================================

// CreateModulesYml creates only the modules.yml file with a single module.
func CreateModulesYml(ctx *TestContext, moduleName, modulePath, moduleType string) error {
	content := fmt.Sprintf(`modules:
  - moniker: %s
    name: %s Module
    type: %s
    files:
      root: %s
`, moduleName, moduleName, moduleType, modulePath)

	return CreateFile(ctx, ".r2r/eac/modules.yml", content)
}

// CreateModuleTypesYml creates module-types.yml with the specified types.
// Common type names: "go-library", "go-command", "docker-image", "mkdocs-book"
func CreateModuleTypesYml(ctx *TestContext, types ...string) error {
	// Map of type definitions
	typeDefinitions := map[string]string{
		"go-library": `  - name: go-library
    description: Go library module
    capabilities:
      - go_module
`,
		"go-command": `  - name: go-command
    description: Go CLI application
    capabilities:
      - go_module
      - executable
`,
		"docker-image": `  - name: docker-image
    description: Docker image
    capabilities:
      - docker
`,
		"mkdocs-book": `  - name: mkdocs-book
    description: MkDocs documentation book
    capabilities:
      - mkdocs
`,
		"gherkin-spec": `  - name: gherkin-spec
    description: Gherkin specification module
    capabilities:
      - gherkin
`,
	}

	var content strings.Builder
	content.WriteString("types:\n")

	for i, t := range types {
		def, ok := typeDefinitions[t]
		if !ok {
			return fmt.Errorf("unknown module type: %s", t)
		}
		if i > 0 {
			content.WriteString("\n")
		}
		content.WriteString(def)
	}

	return CreateFile(ctx, ".r2r/eac/module-types.yml", content.String())
}

// CreateSystemDependenciesYml creates system-dependencies.yml.
func CreateSystemDependenciesYml(ctx *TestContext, withDocker bool) error {
	var content string
	if withDocker {
		content = SystemDependenciesWithDocker
	} else {
		content = SystemDependenciesMinimal
	}
	return CreateFile(ctx, ".r2r/eac/system-dependencies.yml", content)
}

// ============================================================================
// AI Mock Configuration Helpers
// ============================================================================

// SetupMockAI creates the mock AI response file for testing.
func SetupMockAI(ctx *TestContext, response string) error {
	if err := CreateDirectory(ctx, ".r2r/test"); err != nil {
		return err
	}
	return CreateFile(ctx, ".r2r/test/ai-mock.txt", response)
}

// SetupMockAIFromAsset loads a mock response from assets and sets it up.
func SetupMockAIFromAsset(ctx *TestContext, assetPath string) error {
	content, err := LoadAsset(ctx, assetPath)
	if err != nil {
		return err
	}
	return SetupMockAI(ctx, content)
}

// ============================================================================
// Git State Helpers
// ============================================================================

// StageFile stages a single file in git.
func StageFile(ctx *TestContext, path string) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("StageFile requires isolated test environment")
	}

	cmd := exec.Command("git", "add", path)
	cmd.Dir = ctx.IsolatedDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stage %s: %w (output: %s)", path, err, string(output))
	}
	return nil
}

// StageFiles stages multiple files in git.
func StageFiles(ctx *TestContext, paths ...string) error {
	for _, path := range paths {
		if err := StageFile(ctx, path); err != nil {
			return err
		}
	}
	return nil
}

// UnstageAll removes all staged files from git index.
func UnstageAll(ctx *TestContext) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("UnstageAll requires isolated test environment")
	}

	cmd := exec.Command("git", "reset", "HEAD", ".")
	cmd.Dir = ctx.IsolatedDir
	_ = cmd.Run() // Ignore error if nothing to reset
	return nil
}

// ============================================================================
// Test Repository Layout Helpers
// ============================================================================

// CopyTestLayout copies a pre-built test repository layout into the isolated test environment.
// Layouts are stored in templates/test-repositories/<layoutName>/ and contain complete
// directory structures with EAC config, source files, etc.
//
// This is the preferred way to set up test fixtures - using real files rather than
// programmatically generating configs ensures tests match real-world scenarios.
//
// Parameters:
//   - layoutName: Name of the layout directory (e.g., "single-go-module")
//   - stage: If true, stages all files in git after copying
//
// Available layouts:
//   - single-go-module: Single Go library module with minimal EAC config
//   - multi-go-module: Two Go library modules
func CopyTestLayout(ctx *TestContext, layoutName string, stage bool) error {
	if ctx.IsolatedDir == "" {
		return fmt.Errorf("CopyTestLayout requires isolated test environment")
	}
	if ctx.OriginalRepoRoot == "" {
		return fmt.Errorf("CopyTestLayout requires OriginalRepoRoot to be set")
	}

	// Source layout directory
	layoutDir := filepath.Join(ctx.OriginalRepoRoot, "templates", "test-repositories", layoutName)
	if _, err := os.Stat(layoutDir); os.IsNotExist(err) {
		return fmt.Errorf("test layout %q not found at %s", layoutName, layoutDir)
	}

	// Copy all files from layout to isolated dir
	if err := copyDir(layoutDir, ctx.IsolatedDir); err != nil {
		return fmt.Errorf("failed to copy layout %q: %w", layoutName, err)
	}

	// Stage all files if requested
	if stage {
		cmd := exec.Command("git", "add", "-A")
		cmd.Dir = ctx.IsolatedDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to stage layout files: %w (output: %s)", err, string(output))
		}
	}

	return nil
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file from src to dst.
func copyFile(src, dst string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}


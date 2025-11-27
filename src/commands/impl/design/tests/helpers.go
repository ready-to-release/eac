// Package tests provides helper utilities for design command tests.
//
// This file contains reusable test utilities for workspace creation,
// validation, and common assertions.
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidWorkspaceDSL is a minimal valid Structurizr workspace for testing.
const ValidWorkspaceDSL = `workspace "Test Workspace" "A test workspace" {
    model {
        user = person "User" "A user of the system"
        system = softwareSystem "Test System" "The test system" {
            webapp = container "Web Application" "Web app" "Go"
        }

        user -> system "Uses"
        user -> webapp "Interacts with"
    }

    views {
        systemContext system "SystemContext" {
            include *
            autoLayout
        }

        container system "Containers" {
            include *
            autoLayout
        }

        theme default
    }
}
`

// InvalidWorkspaceDSL contains syntax errors for testing validation failures.
const InvalidWorkspaceDSL = `workspace "Invalid" {
    model {
        broken syntax here
    }
}
`

// CreateValidWorkspace creates a valid workspace file for a module.
func CreateValidWorkspace(module string) error {
	ctx := GetContext()
	return ctx.WriteWorkspace(module, ValidWorkspaceDSL)
}

// CreateInvalidWorkspace creates an invalid workspace file for a module.
func CreateInvalidWorkspace(module string) error {
	ctx := GetContext()
	return ctx.WriteWorkspace(module, InvalidWorkspaceDSL)
}

// CreateCustomWorkspace creates a workspace with custom content.
func CreateCustomWorkspace(module, content string) error {
	ctx := GetContext()
	return ctx.WriteWorkspace(module, content)
}

// AssertWorkspaceExists verifies that a workspace file exists.
func AssertWorkspaceExists(module string) error {
	ctx := GetContext()
	if !ctx.WorkspaceExists(module) {
		return fmt.Errorf("workspace does not exist for module %s", module)
	}
	return nil
}

// AssertWorkspaceNotExists verifies that a workspace file does not exist.
func AssertWorkspaceNotExists(module string) error {
	ctx := GetContext()
	if ctx.WorkspaceExists(module) {
		return fmt.Errorf("workspace exists for module %s (should not exist)", module)
	}
	return nil
}

// AssertWorkspaceContains verifies that a workspace contains specific text.
func AssertWorkspaceContains(module, expectedText string) error {
	ctx := GetContext()
	content, err := ctx.ReadWorkspace(module)
	if err != nil {
		return fmt.Errorf("failed to read workspace: %w", err)
	}

	if !strings.Contains(content, expectedText) {
		return fmt.Errorf("workspace does not contain expected text: %s", expectedText)
	}

	return nil
}

// AssertWorkspaceNotContains verifies that a workspace does not contain specific text.
func AssertWorkspaceNotContains(module, unexpectedText string) error {
	ctx := GetContext()
	content, err := ctx.ReadWorkspace(module)
	if err != nil {
		return fmt.Errorf("failed to read workspace: %w", err)
	}

	if strings.Contains(content, unexpectedText) {
		return fmt.Errorf("workspace contains unexpected text: %s", unexpectedText)
	}

	return nil
}

// CreateModuleSource creates a minimal source directory for a module.
// This is needed for create/update commands which analyze source code.
func CreateModuleSource(module string) error {
	ctx := GetContext()
	repoRoot := ctx.GetRepositoryRoot()
	if repoRoot == "" {
		return fmt.Errorf("no repository root available")
	}

	sourcePath := filepath.Join(repoRoot, "src", module)
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Create a minimal Go file
	mainFile := filepath.Join(sourcePath, "main.go")
	content := fmt.Sprintf(`package %s

func main() {
	// Main entry point
}
`, module)

	if err := os.WriteFile(mainFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}

	return nil
}

// DeleteWorkspace removes a workspace file.
func DeleteWorkspace(module string) error {
	ctx := GetContext()
	path := ctx.GetWorkspacePath(module)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// GetDebugLogPath returns the path to debug log files.
func GetDebugLogPath() string {
	ctx := GetContext()
	repoRoot := ctx.GetRepositoryRoot()
	if repoRoot == "" {
		return ""
	}
	return filepath.Join(repoRoot, "out", "logs", "design")
}

// AssertDebugFileExists verifies that a debug file was created.
func AssertDebugFileExists(filename string) error {
	logPath := GetDebugLogPath()
	if logPath == "" {
		return fmt.Errorf("no test environment available")
	}

	filePath := filepath.Join(logPath, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("debug file does not exist: %s", filePath)
	}

	return nil
}

// AssertDebugLogContains verifies that the design.log contains specific text.
func AssertDebugLogContains(expectedText string) error {
	logPath := GetDebugLogPath()
	if logPath == "" {
		return fmt.Errorf("no test environment available")
	}

	logFile := filepath.Join(logPath, "design.log")
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("failed to read log file: %w", err)
	}

	if !strings.Contains(string(content), expectedText) {
		return fmt.Errorf("log file does not contain expected text: %s", expectedText)
	}

	return nil
}

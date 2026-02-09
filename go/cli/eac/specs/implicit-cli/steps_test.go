// Package implicitcli contains godog step implementations for specs/eac/implicit-cli.
//
// Features:
// - specs/eac/implicit-cli/importer/
//
// These tests invoke the importer scripts and verify they work correctly.
// Platform-specific scenarios use runtime detection to skip on non-matching platforms.
package implicitcli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
)

// importerContext holds state between steps for importer tests.
type importerContext struct {
	sharedCtx    *eacgodog.TestContext
	repoRoot     string
	moduleExists bool
	modulePath   string
	tempDir      string
}

var impCtx *importerContext

// RegisterSteps registers all implicit-cli specific step definitions.
func RegisterSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	impCtx = &importerContext{sharedCtx: ctx}

	// Initialize repo root
	initializeImporterContext()

	// Cleanup after each scenario
	sc.After(func(goctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		cleanupImporterContext()
		return goctx, nil
	})

	// Platform detection
	sc.Step(`^I am on Windows with PowerShell$`, iAmOnWindowsWithPowerShell)
	sc.Step(`^I am on Linux with bash$`, iAmOnLinuxWithBash)

	// Module existence
	sc.Step(`^the go-invoker module exists at "([^"]*)"$`, theGoInvokerModuleExistsAt)
	sc.Step(`^the go-invoker module does not exist$`, theGoInvokerModuleDoesNotExist)

	// Actions - use unique patterns to avoid conflict with common steps
	sc.Step(`^I run the PowerShell script "([^"]*)"$`, iRunScript)
	sc.Step(`^I source the bash script "([^"]*)"$`, iSourceScript)

	// Assertions - use shared context for exit code (common steps handle "the exit code is X")
	sc.Step(`^the output contains "([^"]*)"$`, theOutputContains)
	sc.Step(`^the exit code is non-zero$`, theExitCodeIsNonZero)
}

func initializeImporterContext() {
	// Use OriginalRepoRoot because we need to test the actual scripts
	impCtx.repoRoot = impCtx.sharedCtx.OriginalRepoRoot
	impCtx.moduleExists = true
}

func cleanupImporterContext() {
	if impCtx != nil && impCtx.tempDir != "" {
		os.RemoveAll(impCtx.tempDir)
		impCtx.tempDir = ""
	}
}

// ============================================================================
// Platform Detection Steps
// ============================================================================

func iAmOnWindowsWithPowerShell() error {
	if runtime.GOOS != "windows" {
		return godog.ErrPending
	}
	// Check if PowerShell is available
	if _, err := exec.LookPath("powershell"); err != nil {
		return fmt.Errorf("powershell not found: %w", err)
	}
	return nil
}

func iAmOnLinuxWithBash() error {
	if runtime.GOOS != "linux" {
		return godog.ErrPending
	}
	// Check if bash is available
	if _, err := exec.LookPath("bash"); err != nil {
		return fmt.Errorf("bash not found: %w", err)
	}
	return nil
}

// ============================================================================
// Module Existence Steps
// ============================================================================

func theGoInvokerModuleExistsAt(path string) error {
	fullPath := filepath.Join(impCtx.repoRoot, path)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("module not found at %s", fullPath)
	}
	impCtx.moduleExists = true
	impCtx.modulePath = fullPath
	return nil
}

func theGoInvokerModuleDoesNotExist() error {
	// Create a temp directory to run the script from where module doesn't exist
	tempDir, err := os.MkdirTemp("", "importer-test-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	impCtx.tempDir = tempDir
	impCtx.moduleExists = false

	// Copy the importer script to temp dir (without the module)
	if runtime.GOOS == "windows" {
		srcScript := filepath.Join(impCtx.repoRoot, "importer.ps1")
		dstScript := filepath.Join(tempDir, "importer.ps1")
		content, err := os.ReadFile(srcScript)
		if err != nil {
			return fmt.Errorf("failed to read importer.ps1: %w", err)
		}
		if err := os.WriteFile(dstScript, content, 0o644); err != nil {
			return fmt.Errorf("failed to write importer.ps1: %w", err)
		}
	} else {
		srcScript := filepath.Join(impCtx.repoRoot, "importer.sh")
		dstScript := filepath.Join(tempDir, "importer.sh")
		content, err := os.ReadFile(srcScript)
		if err != nil {
			return fmt.Errorf("failed to read importer.sh: %w", err)
		}
		if err := os.WriteFile(dstScript, content, 0o755); err != nil {
			return fmt.Errorf("failed to write importer.sh: %w", err)
		}
	}

	return nil
}

// ============================================================================
// Action Steps
// ============================================================================

func iRunScript(scriptWithArgs string) error {
	var cmd *exec.Cmd
	var workDir string

	if impCtx.moduleExists {
		workDir = impCtx.repoRoot
	} else {
		workDir = impCtx.tempDir
	}

	if runtime.GOOS == "windows" {
		// Parse script and arguments
		parts := strings.Fields(scriptWithArgs)
		scriptPath := filepath.Join(workDir, parts[0])

		// Build command with -File and any additional arguments
		cmdArgs := []string{"-ExecutionPolicy", "Bypass", "-File", scriptPath}
		if len(parts) > 1 {
			cmdArgs = append(cmdArgs, parts[1:]...)
		}
		cmd = exec.Command("powershell", cmdArgs...)
	} else {
		// For non-Windows, use bash to run .ps1 is not applicable
		return fmt.Errorf("iRunScript is for Windows PowerShell scripts only")
	}

	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	impCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			impCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			impCtx.sharedCtx.ExitCode = 1
		}
	} else {
		impCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

func iSourceScript(script string) error {
	var cmd *exec.Cmd
	var workDir string

	if impCtx.moduleExists {
		workDir = impCtx.repoRoot
	} else {
		workDir = impCtx.tempDir
	}

	if runtime.GOOS != "windows" {
		// Parse script and args
		parts := strings.Fields(script)
		scriptPath := filepath.Join(workDir, parts[0])

		// Build bash command to source the script
		bashCmd := fmt.Sprintf("source %s", scriptPath)
		if len(parts) > 1 {
			bashCmd = fmt.Sprintf("source %s %s", scriptPath, strings.Join(parts[1:], " "))
		}

		cmd = exec.Command("bash", "-c", bashCmd)
	} else {
		// For Windows, sourcing doesn't apply the same way
		return fmt.Errorf("iSourceScript is for Linux bash scripts only")
	}

	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	impCtx.sharedCtx.CommandOutput = string(output)

	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			impCtx.sharedCtx.ExitCode = exitErr.ExitCode()
		} else {
			impCtx.sharedCtx.ExitCode = 1
		}
	} else {
		impCtx.sharedCtx.ExitCode = 0
	}

	return nil
}

// ============================================================================
// Assertion Steps
// ============================================================================

func theOutputContains(expected string) error {
	if !strings.Contains(impCtx.sharedCtx.CommandOutput, expected) {
		return fmt.Errorf("expected output to contain %q, got:\n%s", expected, impCtx.sharedCtx.CommandOutput)
	}
	return nil
}

func theExitCodeIsNonZero() error {
	if impCtx.sharedCtx.ExitCode == 0 {
		return fmt.Errorf("expected non-zero exit code, got 0\nOutput:\n%s", impCtx.sharedCtx.CommandOutput)
	}
	return nil
}

// Package tests provides BDD step definitions for the docs command.
//
// This file contains verification steps (Then/And) for docs command scenarios.
package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
)

// registerVerifySteps registers all Then/And step definitions
func registerVerifySteps(sc *godog.ScenarioContext) {
	// Container state verification
	sc.Step(`^MkDocs container should start successfully$`, mkdocsContainerShouldStartSuccessfully)
	sc.Step(`^MkDocs container should be stopped$`, mkdocsContainerShouldBeStopped)

	// HTTP accessibility verification
	sc.Step(`^documentation should be accessible$`, documentationShouldBeAccessible)
	sc.Step(`^documentation should be accessible at "([^"]*)"$`, documentationShouldBeAccessibleAt)

	// Debug file verification
	sc.Step(`^debug log file should exist at "([^"]*)"$`, debugLogFileShouldExistAt)
	sc.Step(`^debug log file should not exist at "([^"]*)"$`, debugLogFileShouldNotExistAt)

	// Note: Exit code and output verification steps are registered in the main test runner
}

// ============================================================================
// Container State Verification Steps
// ============================================================================

func mkdocsContainerShouldStartSuccessfully() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	if !Ctx.DockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container was created and is running
	// Use retry logic to handle transient states like "restarting"
	containerName := "cli-mkdocs"
	maxRetries := 10
	retryDelay := 1 * time.Second

	var lastState string
	for attempt := 0; attempt < maxRetries; attempt++ {
		containers, err := Ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
		if err != nil {
			return fmt.Errorf("failed to list containers: %w", err)
		}

		for _, c := range containers {
			for _, name := range c.Names {
				if strings.TrimPrefix(name, "/") == containerName || strings.Contains(name, containerName) {
					lastState = c.State
					if c.State == "running" {
						Ctx.ContainerStarted = true
						return nil
					}
					// Container exists but not running yet - wait and retry
					break
				}
			}
		}

		// Wait before retry (except on last iteration)
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if lastState != "" {
		return fmt.Errorf("container %s exists but is not running after %d seconds (state: %s)", containerName, maxRetries, lastState)
	}

	return fmt.Errorf("MkDocs container was not created")
}

func mkdocsContainerShouldBeStopped() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	if !Ctx.DockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check that container no longer exists or is not running
	containerName := "cli-mkdocs"
	containers, err := Ctx.DockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName || strings.Contains(name, containerName) {
				return fmt.Errorf("container %s still exists (state: %s)", containerName, c.State)
			}
		}
	}

	return nil
}

// ============================================================================
// HTTP Accessibility Verification Steps
// ============================================================================

func documentationShouldBeAccessible() error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Extract URL from command output
	// Look for pattern like "Documentation: http://localhost:8000"
	lines := strings.Split(Ctx.CommandOutput, "\n")
	var url string
	for _, line := range lines {
		if strings.Contains(line, "Documentation:") {
			// Find the URL after "Documentation: "
			idx := strings.Index(line, "Documentation:")
			if idx != -1 {
				urlPart := strings.TrimSpace(line[idx+len("Documentation:"):])
				if urlPart != "" {
					url = urlPart
					break
				}
			}
		}
	}

	if url == "" {
		// In isolated mode, this is acceptable (mocked)
		if IsolatedTestProjectDir != "" {
			return nil
		}
		return fmt.Errorf("no documentation URL found in output:\n%s", Ctx.CommandOutput)
	}

	// In isolated mode, skip actual HTTP check
	if IsolatedTestProjectDir != "" {
		return nil
	}

	// In integration mode, verify HTTP accessibility
	return verifyHTTPAccessible(url)
}

func documentationShouldBeAccessibleAt(url string) error {
	// In isolated mode, skip actual HTTP check
	if IsolatedTestProjectDir != "" {
		return nil
	}

	// In integration mode, verify HTTP accessibility
	return verifyHTTPAccessible(url)
}

func verifyHTTPAccessible(url string) error {
	// Wait up to 10 seconds for documentation to become accessible
	maxRetries := 10
	retryDelay := 1 * time.Second

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url)
		if err == nil {
			defer resp.Body.Close()

			if resp.StatusCode == 200 {
				// Success - documentation is accessible
				return nil
			}

			lastErr = fmt.Errorf("HTTP %d: expected 200 OK", resp.StatusCode)
		} else {
			lastErr = err
		}

		// Wait before retry (except on last iteration)
		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("documentation not accessible at %s after %d seconds: %v",
		url, maxRetries, lastErr)
}

// ============================================================================
// Debug File Verification Steps
// ============================================================================

func debugLogFileShouldExistAt(relativePath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Get workspace root from OriginalRepoRoot or current directory
	workspaceRoot := OriginalRepoRoot
	if workspaceRoot == "" {
		var err error
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	fullPath := filepath.Join(workspaceRoot, relativePath)

	// Check if file exists
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("debug log file does not exist: %s", fullPath)
		}
		return fmt.Errorf("failed to check debug log file: %w", err)
	}

	return nil
}

func debugLogFileShouldNotExistAt(relativePath string) error {
	if Ctx == nil {
		return fmt.Errorf("test context not initialized")
	}

	// Get workspace root from OriginalRepoRoot or current directory
	workspaceRoot := OriginalRepoRoot
	if workspaceRoot == "" {
		var err error
		workspaceRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	fullPath := filepath.Join(workspaceRoot, relativePath)

	// Check if file does NOT exist
	_, err := os.Stat(fullPath)
	if err == nil {
		return fmt.Errorf("debug log file should not exist but does: %s", fullPath)
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check debug log file: %w", err)
	}

	return nil
}

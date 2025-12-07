package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cucumber/godog"
	"gopkg.in/yaml.v3"
)

// registerNodeVersionSteps registers steps for Node.js version consistency validation
func registerNodeVersionSteps(sc *godog.ScenarioContext, ctx *nodeVersionContext, repoCtx *repositoryContext) {
	// Given steps
	// Note: "I load the system dependencies configuration" and "I discover all GitHub Action workflow files"
	// are already registered by Go version steps, so we reuse them by calling the same methods

	// When steps
	sc.Step(`^I extract the Node\.js version from system-dependencies\.yml$`, func() error {
		return ctx.extractNodeVersionFromSystemDeps()
	})
	sc.Step(`^I extract node-version defaults from all GitHub Actions$`, func() error {
		return ctx.extractNodeVersionFromGitHubActions()
	})

	// Then steps
	sc.Step(`^all GitHub Action node-version defaults should match system-dependencies\.yml$`, func() error {
		return ctx.githubActionsShouldMatchSystemDeps()
	})
	// Note: "if versions differ, I should see the mismatched actions and versions" is already
	// registered by Go version steps as a shared passive assertion

	// Hardcoded action file validation steps
	sc.Step(`^\.github/actions/setup-module-deps/action\.yaml should have matching node-version default$`, func() error {
		return ctx.validateActionNodeVersion(".github/actions/setup-module-deps/action.yaml")
	})
}

// nodeVersionContext holds state for Node.js version consistency validation
type nodeVersionContext struct {
	repoRoot string

	// System dependencies
	sysDepsPath    string
	sysDepsVersion string

	// GitHub Actions
	actionFiles       []string
	actionVersions    map[string]string // file -> node-version default
	mismatchedActions []string
}

func (c *nodeVersionContext) ensureRepoRoot(repoCtx *repositoryContext) error {
	if c.repoRoot == "" {
		if err := repoCtx.ensureRepoRoot(); err != nil {
			return err
		}
		c.repoRoot = repoCtx.repoRoot
	}
	return nil
}

// ============================================================================
// Given Steps
// ============================================================================

func (c *nodeVersionContext) loadSystemDependencies(repoCtx *repositoryContext) error {
	if err := c.ensureRepoRoot(repoCtx); err != nil {
		return err
	}
	c.sysDepsPath = filepath.Join(c.repoRoot, ".r2r", "eac", "system-dependencies.yml")
	if _, err := os.Stat(c.sysDepsPath); os.IsNotExist(err) {
		return fmt.Errorf("system-dependencies.yml not found at %s", c.sysDepsPath)
	}
	return nil
}

func (c *nodeVersionContext) discoverGitHubActionFiles(repoCtx *repositoryContext) error {
	if err := c.ensureRepoRoot(repoCtx); err != nil {
		return err
	}
	actionsDir := filepath.Join(c.repoRoot, ".github", "actions")

	entries, err := os.ReadDir(actionsDir)
	if err != nil {
		return fmt.Errorf("failed to read .github/actions: %w", err)
	}

	c.actionFiles = []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			actionFile := filepath.Join(actionsDir, entry.Name(), "action.yaml")
			if _, err := os.Stat(actionFile); err == nil {
				c.actionFiles = append(c.actionFiles, actionFile)
			}
		}
	}

	if len(c.actionFiles) == 0 {
		return fmt.Errorf("no GitHub Action files found in %s", actionsDir)
	}
	return nil
}

// ============================================================================
// When Steps
// ============================================================================

func (c *nodeVersionContext) extractNodeVersionFromSystemDeps() error {
	data, err := os.ReadFile(c.sysDepsPath)
	if err != nil {
		return fmt.Errorf("failed to read system-dependencies.yml: %w", err)
	}

	var sysDeps struct {
		Dependencies []struct {
			Moniker string `yaml:"moniker"`
			Version string `yaml:"version"`
		} `yaml:"dependencies"`
	}

	if err := yaml.Unmarshal(data, &sysDeps); err != nil {
		return fmt.Errorf("failed to parse system-dependencies.yml: %w", err)
	}

	for _, dep := range sysDeps.Dependencies {
		if dep.Moniker == "npm" {
			// Extract major version from "20.0" -> "20"
			parts := strings.Split(dep.Version, ".")
			if len(parts) > 0 {
				c.sysDepsVersion = parts[0]
			} else {
				c.sysDepsVersion = dep.Version
			}
			return nil
		}
	}

	return fmt.Errorf("npm dependency not found in system-dependencies.yml")
}

func (c *nodeVersionContext) extractNodeVersionFromGitHubActions() error {
	// Initialize map
	c.actionVersions = make(map[string]string)

	// Pattern to match node-version input with default value
	re := regexp.MustCompile(`(?m)^\s+node-version:\s*\n\s+.*\n(?:\s+.*\n)*?\s+default:\s+'?([0-9.]+)'?`)

	for _, actionFile := range c.actionFiles {
		data, err := os.ReadFile(actionFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", actionFile, err)
		}

		matches := re.FindAllStringSubmatch(string(data), -1)
		for _, match := range matches {
			if len(match) >= 2 {
				relPath, _ := filepath.Rel(c.repoRoot, actionFile)
				c.actionVersions[relPath] = match[1]
			}
		}
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

func (c *nodeVersionContext) githubActionsShouldMatchSystemDeps() error {
	c.mismatchedActions = []string{}

	for actionFile, version := range c.actionVersions {
		if version != c.sysDepsVersion {
			c.mismatchedActions = append(c.mismatchedActions,
				fmt.Sprintf("  %s: %s (expected %s)", actionFile, version, c.sysDepsVersion))
		}
	}

	if len(c.mismatchedActions) > 0 {
		return fmt.Errorf(
			"GitHub Action node-version defaults do not match system-dependencies.yml (%s):\n%s\n\n"+
				"Update these GitHub Actions to use node-version default: '%s'",
			c.sysDepsVersion,
			strings.Join(c.mismatchedActions, "\n"),
			c.sysDepsVersion,
		)
	}
	return nil
}

// validateActionNodeVersion validates a specific action file has the correct node-version default
func (c *nodeVersionContext) validateActionNodeVersion(relPath string) error {
	actionPath := filepath.Join(c.repoRoot, relPath)

	data, err := os.ReadFile(actionPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", relPath, err)
	}

	// Pattern to match node-version input with default value
	re := regexp.MustCompile(`(?m)^\s+node-version:\s*\n\s+.*\n(?:\s+.*\n)*?\s+default:\s+'?([0-9.]+)'?`)
	matches := re.FindStringSubmatch(string(data))

	if len(matches) < 2 {
		return fmt.Errorf("%s does not have a node-version default", relPath)
	}

	actualVersion := matches[1]
	if actualVersion != c.sysDepsVersion {
		return fmt.Errorf(
			"%s has node-version default '%s' but system-dependencies.yml specifies '%s'",
			relPath,
			actualVersion,
			c.sysDepsVersion,
		)
	}

	return nil
}

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

// registerGoVersionSteps registers steps for Go version consistency validation
func registerGoVersionSteps(sc *godog.ScenarioContext, ctx *goVersionContext, nodeCtx *nodeVersionContext, repoCtx *repositoryContext) {
	// Given steps
	sc.Step(`^I load the system dependencies configuration$`, func() error {
		// Initialize both Go and Node contexts for shared step
		if err := ctx.loadSystemDependencies(repoCtx); err != nil {
			return err
		}
		return nodeCtx.loadSystemDependencies(repoCtx)
	})
	sc.Step(`^I load the go\.work file$`, func() error {
		return ctx.loadGoWorkFile(repoCtx)
	})
	sc.Step(`^I discover all GitHub Action workflow files$`, func() error {
		// Initialize both Go and Node contexts for shared step
		if err := ctx.discoverGitHubActionFiles(repoCtx); err != nil {
			return err
		}
		return nodeCtx.discoverGitHubActionFiles(repoCtx)
	})

	// When steps
	sc.Step(`^I extract the Go version from system-dependencies\.yml$`, func() error {
		return ctx.extractGoVersionFromSystemDeps()
	})
	sc.Step(`^I extract the Go version from go\.work$`, func() error {
		return ctx.extractGoVersionFromGoWork()
	})
	sc.Step(`^I extract go-version defaults from all GitHub Actions$`, func() error {
		return ctx.extractGoVersionFromGitHubActions()
	})

	// Then steps
	sc.Step(`^the Go versions should match exactly$`, func() error {
		return ctx.goVersionsShouldMatch()
	})
	sc.Step(`^if versions differ, I should see both versions and their locations$`, func() error {
		// Passive assertion - error messages provide details
		return nil
	})
	sc.Step(`^all GitHub Action go-version defaults should match system-dependencies\.yml$`, func() error {
		return ctx.githubActionsShouldMatchSystemDeps()
	})
	sc.Step(`^if versions differ, I should see the mismatched actions and versions$`, func() error {
		// Passive assertion - error messages provide details
		return nil
	})

	// Hardcoded action file validation steps
	sc.Step(`^\.github/actions/setup-commands/action\.yaml should have matching go-version default$`, func() error {
		return ctx.validateActionGoVersion(".github/actions/setup-commands/action.yaml")
	})
	sc.Step(`^\.github/actions/setup-module-deps/action\.yaml should have matching go-version default$`, func() error {
		return ctx.validateActionGoVersion(".github/actions/setup-module-deps/action.yaml")
	})
}

// goVersionContext holds state for Go version consistency validation
type goVersionContext struct {
	repoRoot string

	// System dependencies
	sysDepsPath    string
	sysDepsVersion string

	// go.work
	goWorkPath    string
	goWorkVersion string

	// GitHub Actions
	actionFiles        []string
	actionVersions     map[string]string // file -> go-version default
	mismatchedActions  []string
}

func (c *goVersionContext) ensureRepoRoot(repoCtx *repositoryContext) error {
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

func (c *goVersionContext) loadSystemDependencies(repoCtx *repositoryContext) error {
	if err := c.ensureRepoRoot(repoCtx); err != nil {
		return err
	}
	c.sysDepsPath = filepath.Join(c.repoRoot, ".r2r", "eac", "system-dependencies.yml")
	if _, err := os.Stat(c.sysDepsPath); os.IsNotExist(err) {
		return fmt.Errorf("system-dependencies.yml not found at %s", c.sysDepsPath)
	}
	return nil
}

func (c *goVersionContext) loadGoWorkFile(repoCtx *repositoryContext) error {
	if err := c.ensureRepoRoot(repoCtx); err != nil {
		return err
	}
	c.goWorkPath = filepath.Join(c.repoRoot, "go.work")
	if _, err := os.Stat(c.goWorkPath); os.IsNotExist(err) {
		return fmt.Errorf("go.work not found at %s", c.goWorkPath)
	}
	return nil
}

func (c *goVersionContext) discoverGitHubActionFiles(repoCtx *repositoryContext) error {
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

func (c *goVersionContext) extractGoVersionFromSystemDeps() error {
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
		if dep.Moniker == "go" {
			c.sysDepsVersion = dep.Version
			return nil
		}
	}

	return fmt.Errorf("Go dependency not found in system-dependencies.yml")
}

func (c *goVersionContext) extractGoVersionFromGoWork() error {
	data, err := os.ReadFile(c.goWorkPath)
	if err != nil {
		return fmt.Errorf("failed to read go.work: %w", err)
	}

	// Parse "go 1.24.4" line
	re := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return fmt.Errorf("go version directive not found in go.work")
	}

	c.goWorkVersion = matches[1]
	return nil
}

func (c *goVersionContext) extractGoVersionFromGitHubActions() error {
	// Pattern to match go-version input with default value
	re := regexp.MustCompile(`(?m)^\s+go-version:\s*\n\s+.*\n(?:\s+.*\n)*?\s+default:\s+'?([0-9.]+)'?`)

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

func (c *goVersionContext) goVersionsShouldMatch() error {
	if c.sysDepsVersion != c.goWorkVersion {
		return fmt.Errorf(
			"Go version mismatch:\n"+
				"  system-dependencies.yml: %s\n"+
				"  go.work:                 %s\n\n"+
				"These versions must match to ensure consistent module resolution between local and CI environments.",
			c.sysDepsVersion,
			c.goWorkVersion,
		)
	}
	return nil
}

func (c *goVersionContext) githubActionsShouldMatchSystemDeps() error {
	c.mismatchedActions = []string{}

	for actionFile, version := range c.actionVersions {
		if version != c.sysDepsVersion {
			c.mismatchedActions = append(c.mismatchedActions,
				fmt.Sprintf("  %s: %s (expected %s)", actionFile, version, c.sysDepsVersion))
		}
	}

	if len(c.mismatchedActions) > 0 {
		return fmt.Errorf(
			"GitHub Action go-version defaults do not match system-dependencies.yml (%s):\n%s\n\n"+
				"Update these GitHub Actions to use go-version default: '%s'",
			c.sysDepsVersion,
			strings.Join(c.mismatchedActions, "\n"),
			c.sysDepsVersion,
		)
	}
	return nil
}

// validateActionGoVersion validates a specific action file has the correct go-version default
func (c *goVersionContext) validateActionGoVersion(relPath string) error {
	actionPath := filepath.Join(c.repoRoot, relPath)

	data, err := os.ReadFile(actionPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", relPath, err)
	}

	// Pattern to match go-version input with default value
	re := regexp.MustCompile(`(?m)^\s+go-version:\s*\n\s+.*\n(?:\s+.*\n)*?\s+default:\s+'?([0-9.]+)'?`)
	matches := re.FindStringSubmatch(string(data))

	if len(matches) < 2 {
		return fmt.Errorf("%s does not have a go-version default", relPath)
	}

	actualVersion := matches[1]
	if actualVersion != c.sysDepsVersion {
		return fmt.Errorf(
			"%s has go-version default '%s' but system-dependencies.yml specifies '%s'",
			relPath,
			actualVersion,
			c.sysDepsVersion,
		)
	}

	return nil
}

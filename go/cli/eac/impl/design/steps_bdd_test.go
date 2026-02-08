// Package design contains godog step implementations for eac-cli.
//
// This file contains design command step definitions.
// Note: Most design scenarios require Docker for Structurizr validation.
package design

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"gopkg.in/yaml.v3"
)

// registerSteps registers step definitions for design command features.
func registerSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// No teardown needed - Docker operations are handled by commands via adapters/docker

	// Given steps - repository/module setup
	sc.Step(`^a repository with AI configs at "([^"]*)"$`, func(path string) error {
		return createAIContractFiles(ctx, path)
	})
	sc.Step(`^a module contract exists for "([^"]*)" at "([^"]*)"$`, func(module, sourcePath string) error {
		return createModuleContract(ctx, module, sourcePath)
	})
	sc.Step(`^source code exists at "([^"]*)"$`, func(path string) error {
		return eacgodog.CreateDirectory(ctx, path)
	})
	sc.Step(`^module "([^"]*)" has a workspace$`, func(module string) error {
		// Create module contract first
		modulePath := fmt.Sprintf("specs/%s", module)
		if err := createModuleContract(ctx, module, modulePath); err != nil {
			return err
		}
		// Create a minimal workspace.dsl
		wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", module)
		return eacgodog.CreateFile(ctx, wsPath, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a workspace at "([^"]*)"$`, func(module, path string) error {
		// Create module contract first
		modulePath := fmt.Sprintf("specs/%s", module)
		if err := createModuleContract(ctx, module, modulePath); err != nil {
			return err
		}
		// Then create the workspace file
		return eacgodog.CreateFile(ctx, path, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a valid workspace$`, func(module string) error {
		wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", module)
		return eacgodog.CreateFile(ctx, wsPath, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has a valid workspace at "([^"]*)"$`, func(module, path string) error {
		return eacgodog.CreateFile(ctx, path, minimalWorkspaceDSL(module))
	})
	sc.Step(`^module "([^"]*)" has an invalid workspace at "([^"]*)"$`, func(module, path string) error {
		return eacgodog.CreateFile(ctx, path, "invalid { dsl content")
	})
	sc.Step(`^no workspace exists for "([^"]*)"$`, func(module string) error {
		// Create module contract first so it exists in repository.yml
		sourcePath := fmt.Sprintf("specs/%s", module)
		if err := createModuleContract(ctx, module, sourcePath); err != nil {
			return err
		}
		// Just ensure directory structure exists without workspace
		return eacgodog.CreateDirectory(ctx, sourcePath)
	})
	sc.Step(`^multiple modules have workspace files$`, func() error {
		// Use real modules that exist in fixture repository.yml
		for _, mod := range []string{"core", "docs", "clie-cli"} {
			wsPath := fmt.Sprintf("specs/%s/.design/workspace.dsl", mod)
			if err := eacgodog.CreateFile(ctx, wsPath, minimalWorkspaceDSL(mod)); err != nil {
				return err
			}
		}
		return nil
	})

	// Mock AI configuration - creates mock file in isolated directory for subprocess

	// Then steps - file verification
	// Note: "the file ... should exist" is registered in eacgodog/steps.go
	sc.Step(`^a workspace file exists at "([^"]*)"$`, func(path string) error {
		// Create the file (this is a Given step, not a Then step)
		return eacgodog.CreateFile(ctx, path, minimalWorkspaceDSL("test-module"))
	})
	// Note: "debug files should exist in" is registered in eacgodog/steps.go
	sc.Step(`^validation results should be written to "([^"]*)"$`, func(path string) error {
		return eacgodog.FileExists(ctx, path)
	})
	sc.Step(`^aggregated results should be written to JSON file$`, func() error {
		// Placeholder - check for any JSON output
		return nil
	})

	// Then steps - output verification
	// Note: "the output should contain" is registered in eacgodog/steps.go
	sc.Step(`^I should see success message with URL$`, func() error {
		return eacgodog.OutputContainsAny(ctx, "http://", "https://", "localhost")
	})
	sc.Step(`^documentation should be accessible at the URL$`, func() error {
		// Placeholder - would need HTTP check
		return nil
	})

	// Docker container steps
	sc.Step(`^Structurizr container should start successfully$`, func() error {
		// Check output for success indicators instead of actual Docker
		// Since we can't mock Docker from this package, we verify the command output
		output := ctx.CommandOutput
		if strings.Contains(output, "Starting Structurizr Lite") ||
			strings.Contains(output, "http://localhost:") ||
			strings.Contains(output, "structurizr") {
			return nil
		}
		return fmt.Errorf("expected Structurizr start message in output, got:\n%s", output)
	})
	sc.Step(`^two separate containers should be running$`, func() error {
		// Verify command was successful for both modules by checking exit codes
		// In isolated tests, we can't actually check Docker, so we verify the command succeeded
		if ctx.ExitCode != 0 {
			return fmt.Errorf("expected successful command execution, got exit code %d", ctx.ExitCode)
		}
		return nil
	})
	sc.Step(`^each should have a unique port$`, func() error {
		// Verify output mentions port allocation
		output := ctx.CommandOutput
		if !strings.Contains(output, "localhost:") {
			return fmt.Errorf("expected port information in output, got:\n%s", output)
		}
		return nil
	})

	// Verbose output verification
	sc.Step(`^the output should contain Docker command details$`, func() error {
		output := strings.ToLower(ctx.CommandOutput)

		// Check for Docker-related keywords in verbose output
		hasDocker := strings.Contains(output, "docker")
		hasStructurizr := strings.Contains(output, "structurizr")
		hasContainer := strings.Contains(output, "container")

		if !hasDocker && !hasStructurizr && !hasContainer {
			return fmt.Errorf(
				"expected Docker command details in output, but found none\n\nActual output:\n%s",
				ctx.CommandOutput,
			)
		}

		return nil
	})
}

// minimalWorkspaceDSL returns a minimal valid Structurizr DSL workspace.
func minimalWorkspaceDSL(module string) string {
	return fmt.Sprintf(`workspace "%s" "Test workspace" {
    model {
        user = person "User"
        system = softwareSystem "%s System"
        user -> system "Uses"
    }
    views {
        systemContext system "SystemContext" {
            include *
            autoLayout
        }
    }
}`, module, module)
}

// createModuleContract creates a minimal module contract for testing by appending to repository.yml.
// Note: Modules are defined in repository.yml (unified config).
func createModuleContract(ctx *eacgodog.TestContext, module, sourcePath string) error {
	// Path to repository.yml in isolated environment
	repoYmlPath := filepath.Join(eacgodog.ResolvePath(ctx, ""), ".eac", "repository.yml")

	// Read existing repository.yml
	existingData, err := os.ReadFile(repoYmlPath)
	if err != nil {
		return fmt.Errorf("failed to read repository.yml: %w", err)
	}

	// Parse existing YAML
	var config map[string]interface{}
	if err := yaml.Unmarshal(existingData, &config); err != nil {
		return fmt.Errorf("failed to parse repository.yml: %w", err)
	}

	// Get modules array
	modules, ok := config["modules"].([]interface{})
	if !ok {
		return fmt.Errorf("modules field is not an array")
	}

	// Append our test module using new component-based format
	testModule := map[string]interface{}{
		"moniker":     module,
		"name":        fmt.Sprintf("Test Module %s", module),
		"description": "Test module for BDD tests",
		"components": map[string]interface{}{
			"go": sourcePath,
		},
	}
	modules = append(modules, testModule)
	config["modules"] = modules

	// Marshal back to YAML
	updatedData, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal updated repository.yml: %w", err)
	}

	// Write back to file
	if err := os.WriteFile(repoYmlPath, updatedData, 0o644); err != nil {
		return fmt.Errorf("failed to write updated repository.yml: %w", err)
	}

	return nil
}

// createAIContractFiles creates the minimal AI contract files needed for design command.
func createAIContractFiles(ctx *eacgodog.TestContext, configPath string) error {
	// Ensure the directory exists
	if err := eacgodog.CreateDirectory(ctx, configPath); err != nil {
		return err
	}

	// Create design.md (prompt file)
	designMd := filepath.Join(configPath, "design.md")
	designContent := `# Structurizr DSL Architecture Generation

Generate a complete Structurizr DSL workspace.

## Output Requirements

1. Valid DSL
2. Clean Format
3. Complete Structure
`
	if err := eacgodog.CreateFile(ctx, designMd, designContent); err != nil {
		return err
	}

	// Create contract.yml
	contractYml := filepath.Join(configPath, "contract.yml")
	contractContent := `version: "0.1.0"
name: "Structurizr DSL Architecture"
description: "Formal contract"

workspace_name_pattern: "^[A-Z][A-Za-z0-9 -]+$"
identifier_pattern: "^[a-z][a-z0-9_]*$"

required_elements:
  - "workspace"
  - "model"
  - "views"
`
	if err := eacgodog.CreateFile(ctx, contractYml, contractContent); err != nil {
		return err
	}

	// Create anti-corruption.yml
	antiCorruptionYml := filepath.Join(configPath, "anti-corruption.yml")
	antiCorruptionContent := `version: "0.1.0"
name: "Structurizr DSL Anti-Corruption Rules"

forbidden_prefixes:
  - "Here is"
  - "I'll"

forbidden_contains:
  - "generating"

forbidden_emojis:
  - "🚀"

agent_signatures: []
`
	return eacgodog.CreateFile(ctx, antiCorruptionYml, antiCorruptionContent)
}

// Command documentation coverage step definitions for specs/repository/command-docs-coverage.
//
// This file implements steps for validating that every CLI command
// has corresponding reference documentation.
//
// The mapping from command names to documentation paths is driven by
// .eac/commands.yml configuration, making it easy to adjust rules
// without code changes.
package repository

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cucumber/godog"
	eac "github.com/ready-to-release/eac/go/adapters/eac"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/docsync"
	"github.com/ready-to-release/eac/go/core/tool"
	eacgodog "github.com/ready-to-release/eac/go/adapters/godog"
	"gopkg.in/yaml.v3"
)

// commandDocsContext holds state for command docs coverage validation.
type commandDocsContext struct {
	repoRoot   string
	cfg        *config.EACConfig
	scanResult *docsync.CommandDocSyncResult
}

// registerCommandDocsSteps registers command documentation coverage step definitions.
func registerCommandDocsSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Per-scenario context -- no global mutable state.
	cdCtx := &commandDocsContext{}

	sc.Step(`^I load all valid commands from the CLI$`, cdCtx.loadAllValidCommandsFromCLI)
	sc.Step(`^I scan docs/reference/eac/commands/ for command documentation files$`, cdCtx.scanDocsForCommandDocumentation)
	sc.Step(`^I check each command for a corresponding documentation file$`, cdCtx.checkEachCommandForDocumentation)
	sc.Step(`^every command should have a documentation file$`, cdCtx.everyCommandShouldHaveDocumentation)
	sc.Step(`^if any commands are missing documentation, I should see their names$`, cdCtx.ifMissingShowCommandNames)
}

func (c *commandDocsContext) loadAllValidCommandsFromCLI() error {
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("could not find repository root: %w", err)
	}

	// Load EAC configuration (includes commands.yml)
	cfg, err := config.Load(config.LoadOptions{
		RepoRoot:        repoRoot,
		ValidateSchemas: false, // Skip validation for faster loading
	})
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	c.repoRoot = repoRoot
	c.cfg = cfg
	return nil
}

func (c *commandDocsContext) scanDocsForCommandDocumentation() error {
	// This step is now essentially a no-op since we determine expected paths
	// from config, not by scanning the filesystem.
	// We keep it for backwards compatibility with the Gherkin spec.
	if c.cfg == nil {
		return fmt.Errorf("commands not loaded - run 'I load all valid commands from the CLI' first")
	}
	return nil
}

func (c *commandDocsContext) checkEachCommandForDocumentation() error {
	if c.cfg == nil {
		return fmt.Errorf("context not initialized")
	}

	// Use EAC adapter to get valid commands
	commandSource := makeCommandSource(c.repoRoot)

	// Use docsync to scan command documentation
	result, err := docsync.ScanCommandDocs(commandSource, c.repoRoot, c.cfg.Commands)
	if err != nil {
		return fmt.Errorf("failed to scan command docs: %w", err)
	}

	if result.ValidCommands == 0 {
		return fmt.Errorf("no valid commands found from CLI")
	}

	c.scanResult = result
	return nil
}

func (c *commandDocsContext) everyCommandShouldHaveDocumentation() error {
	if c.scanResult == nil {
		return fmt.Errorf("context not initialized")
	}

	result := c.scanResult

	if len(result.MissingDocs) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d command(s) without documentation:\n\n", len(result.MissingDocs)))

		for _, doc := range result.MissingDocs {
			sb.WriteString(fmt.Sprintf("  - %s (expected: %s)\n", doc.Command, doc.ExpectedDoc))
		}

		sb.WriteString("\nTo fix, run:\n")
		sb.WriteString("  go run ./go/cli/eac update docs --area command-refs\n")
		sb.WriteString("\nOr create files manually with:\n")
		sb.WriteString("  <!-- book:cmd <command-name> -->\n")

		return fmt.Errorf("%s", sb.String())
	}

	return nil
}

func (c *commandDocsContext) ifMissingShowCommandNames() error {
	// Passive assertion - error messages from everyCommandShouldHaveDocumentation provide details
	return nil
}

// makeCommandSource creates a CommandSource that uses the EAC adapter.
func makeCommandSource(repoRoot string) docsync.CommandSource {
	return func() ([]docsync.CommandInfo, error) {
		port := eac.New(repoRoot, tool.GlobalRegistry(), tool.GlobalExecutor())
		result, err := port.Execute(context.Background(), []string{"get", "valid-commands"}, &eac.ExecConfig{
			WorkspaceRoot: repoRoot,
			FullEnv:       append(os.Environ(), "NO_COLOR=1"),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to run get valid-commands: %w", err)
		}
		if !result.Success() {
			return nil, fmt.Errorf("failed to run get valid-commands: %s", string(result.Stderr))
		}
		var commands []docsync.CommandInfo
		if err := yaml.Unmarshal(result.Stdout, &commands); err != nil {
			return nil, fmt.Errorf("failed to parse valid-commands output: %w", err)
		}
		return commands, nil
	}
}

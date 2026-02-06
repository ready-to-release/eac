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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/docsync"
	"github.com/ready-to-release/eac/go/core/environments"
	eacgodog "github.com/ready-to-release/eac/go/godog"
)

// commandDocsContext holds state for command docs coverage validation.
type commandDocsContext struct {
	repoRoot   string
	cfg        *config.EACConfig
	scanResult *docsync.CommandDocSyncResult
}

var cmdDocsCtx *commandDocsContext

// registerCommandDocsSteps registers command documentation coverage step definitions.
func registerCommandDocsSteps(sc *godog.ScenarioContext, ctx *eacgodog.TestContext) {
	// Given steps
	sc.Step(`^I load all valid commands from the CLI$`, loadAllValidCommandsFromCLI)
	sc.Step(`^I scan docs/reference/eac/commands/ for command documentation files$`, scanDocsForCommandDocumentation)

	// When steps
	sc.Step(`^I check each command for a corresponding documentation file$`, checkEachCommandForDocumentation)

	// Then steps
	sc.Step(`^every command should have a documentation file$`, everyCommandShouldHaveDocumentation)
	sc.Step(`^if any commands are missing documentation, I should see their names$`, ifMissingShowCommandNames)
}

func loadAllValidCommandsFromCLI() error {
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

	cmdDocsCtx = &commandDocsContext{
		repoRoot: repoRoot,
		cfg:      cfg,
	}

	return nil
}

func scanDocsForCommandDocumentation() error {
	// This step is now essentially a no-op since we determine expected paths
	// from config, not by scanning the filesystem.
	// We keep it for backwards compatibility with the Gherkin spec.
	if cmdDocsCtx == nil {
		return fmt.Errorf("commands not loaded - run 'I load all valid commands from the CLI' first")
	}
	return nil
}

func checkEachCommandForDocumentation() error {
	if cmdDocsCtx == nil {
		return fmt.Errorf("context not initialized")
	}

	// Find the commands binary
	cmdBinary := findCommandsBinary(cmdDocsCtx.repoRoot)
	if cmdBinary == "" {
		return fmt.Errorf("commands binary not found - run 'go build' or set COMMANDS_PATH")
	}

	// Use docsync to scan command documentation
	result, err := docsync.ScanCommandDocs(cmdBinary, cmdDocsCtx.repoRoot, cmdDocsCtx.cfg.Commands)
	if err != nil {
		return fmt.Errorf("failed to scan command docs: %w", err)
	}

	if result.ValidCommands == 0 {
		return fmt.Errorf("no valid commands found from CLI")
	}

	cmdDocsCtx.scanResult = result
	return nil
}

func everyCommandShouldHaveDocumentation() error {
	if cmdDocsCtx == nil || cmdDocsCtx.scanResult == nil {
		return fmt.Errorf("context not initialized")
	}

	result := cmdDocsCtx.scanResult

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

func ifMissingShowCommandNames() error {
	// Passive assertion - error messages from everyCommandShouldHaveDocumentation provide details
	return nil
}

// findCommandsBinary locates the commands binary.
// Checks multiple locations in order of preference.
func findCommandsBinary(repoRoot string) string {
	// Check environment variable first (set by CI)
	if envPath := os.Getenv(environments.EnvCommandsPath); envPath != "" {
		return envPath
	}

	var candidates []string

	// When running inside a container, prefer the container's built-in binary
	// R2R_CONTAINER_ROOT is always set by the container image (ENV in Dockerfile)
	if containerRoot := os.Getenv(environments.EnvR2RContainerRoot); containerRoot != "" {
		candidates = append(candidates, filepath.Join(containerRoot, "out", "tools", "eac"))
	}

	// Add standard locations from the repo
	candidates = append(candidates,
		filepath.Join(repoRoot, "out", "tools", "eac"),
	)

	// Only add .exe candidates on Windows
	// On Linux, .exe files from a mounted Windows host cannot be executed
	if runtime.GOOS == "windows" {
		candidates = append(candidates,
			filepath.Join(repoRoot, "out", "tools", "eac.exe"),
		)
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

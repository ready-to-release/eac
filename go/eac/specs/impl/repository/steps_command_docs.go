// Command documentation coverage step definitions for specs/repository/command-docs-coverage.
//
// This file implements steps for validating that every CLI command
// has corresponding reference documentation.
//
// The mapping from command names to documentation paths is driven by
// .r2r/eac/commands.yml configuration, making it easy to adjust rules
// without code changes.
package repository

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cucumber/godog"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/specs/internal"
	"gopkg.in/yaml.v3"
)

// commandDocsContext holds state for command docs coverage validation.
type commandDocsContext struct {
	repoRoot         string
	cfg              *config.EACConfig
	validCommands    []string          // All valid CLI commands
	missingCommands  []string          // Commands without documentation
	commandToDocFile map[string]string // Expected doc file for each command
}

var cmdDocsCtx *commandDocsContext

// registerCommandDocsSteps registers command documentation coverage step definitions.
func registerCommandDocsSteps(sc *godog.ScenarioContext, ctx *internal.TestContext) {
	// Given steps
	sc.Step(`^I load all valid commands from the CLI$`, loadAllValidCommandsFromCLI)
	sc.Step(`^I scan docs/reference/commands/ for command documentation files$`, scanDocsForCommandDocumentation)

	// When steps
	sc.Step(`^I check each command for a corresponding documentation file$`, checkEachCommandForDocumentation)

	// Then steps
	sc.Step(`^every command should have a documentation file$`, everyCommandShouldHaveDocumentation)
	sc.Step(`^if any commands are missing documentation, I should see their names$`, ifMissingShowCommandNames)
}

// commandInfo matches the YAML structure from 'get valid-commands'.
type commandInfo struct {
	Command     string `yaml:"command"`
	Description string `yaml:"description"`
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
		repoRoot:         repoRoot,
		cfg:              cfg,
		validCommands:    []string{},
		commandToDocFile: make(map[string]string),
	}

	// Run the CLI to get valid commands
	// Check multiple locations in order of preference:
	// 1. COMMANDS_PATH environment variable (set by CI)
	// 2. Container built-in binary (when running inside Docker via r2r)
	// 3. out/tools/commands (CI artifact location)
	// 4. go/eac/commands/build/commands (local dev build location)
	cmdBinary := os.Getenv("COMMANDS_PATH")
	if cmdBinary == "" {
		var candidates []string

		// When running inside a container, prefer the container's built-in binary
		// R2R_CONTAINER_ROOT is always set by the container image (ENV in Dockerfile)
		containerRoot := os.Getenv("R2R_CONTAINER_ROOT")
		if containerRoot != "" {
			candidates = append(candidates, filepath.Join(containerRoot, "out", "tools", "commands"))
		}

		// Add standard locations from the repo
		candidates = append(candidates,
			filepath.Join(repoRoot, "out", "tools", "commands"),
			filepath.Join(repoRoot, "go", "eac", "commands", "build", "commands"),
		)

		// Only add .exe candidates on Windows
		// On Linux, .exe files from a mounted Windows host cannot be executed
		if runtime.GOOS == "windows" {
			candidates = append(candidates,
				filepath.Join(repoRoot, "out", "tools", "commands.exe"),
				filepath.Join(repoRoot, "go", "eac", "commands", "build", "commands.exe"),
			)
		}

		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				cmdBinary = candidate
				break
			}
		}
		if cmdBinary == "" {
			return fmt.Errorf("commands binary not found - run 'go build' or set COMMANDS_PATH")
		}
	}

	cmd := exec.Command(cmdBinary, "get", "valid-commands")
	cmd.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run 'get valid-commands': %w\nstderr: %s", err, stderr.String())
	}

	// Parse YAML output
	var commands []commandInfo
	if err := yaml.Unmarshal(stdout.Bytes(), &commands); err != nil {
		return fmt.Errorf("failed to parse valid-commands output: %w\noutput: %s", err, stdout.String())
	}

	// Get commands config (guaranteed non-nil after Load)
	cmdsCfg := cfg.Commands

	for _, cmdInfo := range commands {
		// Skip commands that are configured to not require docs
		if cmdsCfg.ShouldSkipDocs(cmdInfo.Command) {
			continue
		}

		cmdDocsCtx.validCommands = append(cmdDocsCtx.validCommands, cmdInfo.Command)

		// Use config to determine expected doc file path
		// The config handles:
		// - Internal commands -> other/internal/{command}.md
		// - Category root commands -> {category}/{category}.md
		// - Subcommands -> {category}/{subcommand}.md
		// - Uncategorized -> other/{command}.md
		docPath := cmdsCfg.GetDocPath(cmdInfo.Command, "", repoRoot)
		cmdDocsCtx.commandToDocFile[cmdInfo.Command] = docPath
	}

	if len(cmdDocsCtx.validCommands) == 0 {
		return fmt.Errorf("no valid commands found from CLI")
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

	cmdDocsCtx.missingCommands = []string{}

	for _, cmd := range cmdDocsCtx.validCommands {
		expectedDoc := cmdDocsCtx.commandToDocFile[cmd]
		fullPath := filepath.Join(cmdDocsCtx.repoRoot, expectedDoc)

		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			cmdDocsCtx.missingCommands = append(cmdDocsCtx.missingCommands, cmd)
		}
	}

	return nil
}

func everyCommandShouldHaveDocumentation() error {
	if cmdDocsCtx == nil {
		return fmt.Errorf("context not initialized")
	}

	if len(cmdDocsCtx.missingCommands) > 0 {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Found %d command(s) without documentation:\n\n", len(cmdDocsCtx.missingCommands)))

		for _, cmd := range cmdDocsCtx.missingCommands {
			expectedPath := cmdDocsCtx.commandToDocFile[cmd]
			sb.WriteString(fmt.Sprintf("  - %s (expected: %s)\n", cmd, expectedPath))
		}

		sb.WriteString("\nTo fix, create documentation files for each missing command using:\n")
		sb.WriteString("  <!-- book:cmd <command-name> -->\n")

		return fmt.Errorf("%s", sb.String())
	}

	return nil
}

func ifMissingShowCommandNames() error {
	// Passive assertion - error messages from everyCommandShouldHaveDocumentation provide details
	return nil
}

package drawio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type drawioInfoCommand struct{}

var _ core.SimpleCommandPort = (*drawioInfoCommand)(nil)

func (c *drawioInfoCommand) Name() string { return "drawio info" }

func (c *drawioInfoCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "drawio-info",
		Short:         "Show information about a DrawIO file",
		Long:          "Displays metadata about a DrawIO file including diagram count,\npage names, cell count, and tool information.\n\nExample:\n  drawio info diagram.drawio.png\n  drawio info --json diagram.drawio.png",
		Flags: []core.FlagSpec{
			{Name: "json", Type: "bool", Usage: "Output as JSON"},
		},
		Args: "drawio-file",
	}
}

func (c *drawioInfoCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return DrawioInfo()
}

// DrawioInfo shows information about a DrawIO file.
func DrawioInfo() int {
	args := os.Args[3:] // Skip program, "drawio", "info"
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags and find input file
	var inputFile string
	jsonOutput := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOutput = true
		default:
			if !strings.HasPrefix(args[i], "-") && inputFile == "" {
				inputFile = args[i]
			}
		}
	}

	if inputFile == "" {
		log.Errorf("Error: input file is required")
		return 1
	}

	// Get repo root
	repoRoot, err := GetRepoRoot()
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Check Docker
	if err := CheckDockerAvailable(repoRoot, nil); err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Make input path absolute
	if !filepath.IsAbs(inputFile) {
		inputFile = filepath.Join(repoRoot, inputFile)
	}

	// Verify input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Errorf("Error: File not found: %s", inputFile)
		return 1
	}

	containerInput, err := TranslateToContainerPath(inputFile, repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build command args
	cmdArgs := []string{"info", containerInput}
	if jsonOutput {
		cmdArgs = append(cmdArgs, "--json")
	}

	// Run command
	output, err := RunDrawioCommandWithOutput(repoRoot, cmdArgs, nil)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Print(output)

	return 0
}

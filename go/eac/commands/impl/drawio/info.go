// Command: drawio info
// Short: Show information about a DrawIO file
// Long: Displays metadata about a DrawIO file including diagram count,
// Long: page names, cell count, and tool information.
// Long:
// Long: Example:
// Long:   drawio info diagram.drawio.png
// Long:   drawio info --json diagram.drawio.png
// Flag.json: type=bool, usage=Output as JSON
// Args: drawio-file
package drawio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/eac/commands/internal/flags"
	"github.com/ready-to-release/eac/go/eac/commands/registry"
)

func init() {
	registry.Register(DrawioInfo)
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
	if err := CheckDockerAvailable(repoRoot); err != nil {
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
	output, err := RunDrawioCommandWithOutput(repoRoot, cmdArgs)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Print(output)

	return 0
}

// Command: drawio render
// Short: Render diagram to actual PNG image
// Long: Renders a DrawIO diagram to an actual PNG image that can be viewed.
// Long:
// Long: This converts the vector diagram into a rasterized PNG image,
// Long: allowing you to see what the diagram actually looks like.
// Long:
// Long: Example:
// Long:   drawio render -i diagram.drawio.png -o rendered.png
// Flag.input: type=string, shorthand=i, usage=Input .drawio.png or decoded XML file, required=true
// Flag.output: type=string, shorthand=o, usage=Output PNG file (rendered image), required=true
package drawio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/clibase/flags"
)

// DrawioRender renders a DrawIO diagram to an actual PNG image.
func DrawioRender() int {
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	var inputFile, outputFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-i", "--input":
			if i+1 < len(args) {
				inputFile = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	if inputFile == "" {
		log.Errorf("Error: --input is required")
		return 1
	}
	if outputFile == "" {
		log.Errorf("Error: --output is required")
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

	// Make paths absolute
	if !filepath.IsAbs(inputFile) {
		inputFile = filepath.Join(repoRoot, inputFile)
	}
	if !filepath.IsAbs(outputFile) {
		outputFile = filepath.Join(repoRoot, outputFile)
	}

	// Verify input file exists
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		log.Errorf("Error: Input file not found: %s", inputFile)
		return 1
	}

	// Ensure output directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
		log.Errorf("Error creating output directory: %v", err)
		return 1
	}

	// Build container paths
	containerInput, err := TranslateToContainerPath(inputFile, repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}
	containerOutput, err := TranslateToContainerPath(outputFile, repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build command args
	cmdArgs := []string{"render", "-i", containerInput, "-o", containerOutput}

	// Run command
	_, err = RunDrawioCommandWithOutput(repoRoot, cmdArgs)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	fmt.Fprintf(os.Stderr, "Rendered to %s\n", outputFile)
	return 0
}

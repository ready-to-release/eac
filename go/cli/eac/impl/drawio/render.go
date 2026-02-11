package drawio

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
)

type drawioRenderCommand struct{}

var _ core.SimpleCommandPort = (*drawioRenderCommand)(nil)

func (c *drawioRenderCommand) Name() string { return "drawio render" }

func (c *drawioRenderCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "drawio-render",
		Short:         "Render diagram to actual PNG image",
		Long:          "Renders a DrawIO diagram to an actual PNG image that can be viewed.\n\nThis converts the vector diagram into a rasterized PNG image,\nallowing you to see what the diagram actually looks like.\n\nExample:\n  drawio render -i diagram.drawio.png -o rendered.png",
		Flags: []core.FlagSpec{
			{Name: "input", Type: "string", Shorthand: "i", Required: true, Usage: "Input .drawio.png or decoded XML file"},
			{Name: "output", Type: "string", Shorthand: "o", Required: true, Usage: "Output PNG file (rendered image)"},
		},
	}
}

func (c *drawioRenderCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return DrawioRender()
}

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

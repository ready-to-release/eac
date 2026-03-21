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

type drawioEncodeCommand struct{}

var _ core.SimpleCommandPort = (*drawioEncodeCommand)(nil)

func (c *drawioEncodeCommand) Name() string { return "drawio encode" }

func (c *drawioEncodeCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "drawio-encode",
		Short:         "Encode human-readable XML back to DrawIO format",
		Long:          "Encodes decoded/human-readable DrawIO XML back to the compressed\nformat used by DrawIO PNG files.\n\nUse this after editing decoded XML, then use 'drawio embed' to\nwrite the encoded XML back into a .drawio.png file.\n\nExample:\n  drawio encode -i decoded.xml -o encoded.xml\n  drawio encode -i decoded.xml | drawio embed --png diagram.drawio.png",
		Flags: []core.FlagSpec{
			{Name: "input", Type: "string", Shorthand: "i", Usage: "Input decoded XML file (or stdin)"},
			{Name: "output", Type: "string", Shorthand: "o", Usage: "Output file (default: stdout)"},
		},
	}
}

func (c *drawioEncodeCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return DrawioEncode()
}

// DrawioEncode encodes human-readable XML back to DrawIO format.
func DrawioEncode() int {
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

	// Require explicit input — stdin piping is not supported
	if inputFile == "" {
		log.Errorf("Error: --input is required. Provide an XML file to encode.")
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

	// Build command args
	cmdArgs := []string{"encode"}

	if inputFile != "" {
		// Make input path absolute
		if !filepath.IsAbs(inputFile) {
			inputFile = filepath.Join(repoRoot, inputFile)
		}

		// Verify input file exists
		if _, err := os.Stat(inputFile); os.IsNotExist(err) {
			log.Errorf("Error: Input file not found: %s", inputFile)
			return 1
		}

		containerInput, err := TranslateToContainerPath(inputFile, repoRoot)
		if err != nil {
			log.Errorf("Error: %v", err)
			return 1
		}
		cmdArgs = append(cmdArgs, "-i", containerInput)
	}

	if outputFile != "" {
		// Make output path absolute
		if !filepath.IsAbs(outputFile) {
			outputFile = filepath.Join(repoRoot, outputFile)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outputFile), 0o755); err != nil {
			log.Errorf("Error creating output directory: %v", err)
			return 1
		}

		containerOutput, err := TranslateToContainerPath(outputFile, repoRoot)
		if err != nil {
			log.Errorf("Error: %v", err)
			return 1
		}
		cmdArgs = append(cmdArgs, "-o", containerOutput)
	}

	// Run command
	output, err := RunDrawioCommandWithOutput(repoRoot, cmdArgs, nil)
	if err != nil {
		if strings.Contains(err.Error(), "Encoded to") {
			// Success message in stderr
		} else {
			log.Errorf("Error: %v", err)
			return 1
		}
	}

	// If no output file, print to stdout
	if outputFile == "" {
		fmt.Print(output)
	} else {
		fmt.Fprintf(os.Stderr, "Encoded to %s\n", outputFile)
	}

	return 0
}

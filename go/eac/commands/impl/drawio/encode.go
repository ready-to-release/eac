// Command: drawio encode
// Short: Encode human-readable XML back to DrawIO format
// Long: Encodes decoded/human-readable DrawIO XML back to the compressed
// Long: format used by DrawIO PNG files.
// Long:
// Long: Use this after editing decoded XML, then use 'drawio embed' to
// Long: write the encoded XML back into a .drawio.png file.
// Long:
// Long: Example:
// Long:   drawio encode -i decoded.xml -o encoded.xml
// Long:   drawio encode -i decoded.xml | drawio embed --png diagram.drawio.png
// Flag.input: type=string, shorthand=i, usage=Input decoded XML file (or stdin)
// Flag.output: type=string, shorthand=o, usage=Output file (default: stdout)
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
	registry.Register(DrawioEncode)
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
	output, err := RunDrawioCommandWithOutput(repoRoot, cmdArgs)
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

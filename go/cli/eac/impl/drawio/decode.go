// Command: drawio decode
// Short: Decode DrawIO content to human-readable XML
// Long: Decodes compressed/encoded DrawIO content to readable XML.
// Long: Accepts .drawio.png files or raw XML input.
// Long: Output format is optimized for LLM understanding and editing.
// Long:
// Long: The decoded XML shows the full mxGraphModel structure with all
// Long: shapes (mxCell elements), their positions, styles, and connections.
// Long:
// Long: Example:
// Long:   drawio decode -i diagram.drawio.png -o decoded.xml
// Long:   drawio decode -i diagram.drawio.png  # outputs to stdout
// Flag.input: type=string, shorthand=i, usage=Input file (.drawio.png or .xml)
// Flag.output: type=string, shorthand=o, usage=Output file (default: stdout)
package drawio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
)

// DrawioDecode decodes DrawIO content to human-readable XML.
func DrawioDecode() int {
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
		log.Errorf("Error: Input file not found: %s", inputFile)
		return 1
	}

	// Translate to container path
	containerInput, err := TranslateToContainerPath(inputFile, repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build command args
	cmdArgs := []string{"decode", "-i", containerInput}

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
		// Check if it's just stderr messages (status messages go to stderr)
		if strings.Contains(err.Error(), "Decoded to") {
			// Success message in stderr, output is fine
		} else {
			log.Errorf("Error: %v", err)
			return 1
		}
	}

	// If no output file, print to stdout
	if outputFile == "" {
		fmt.Print(output)
	} else {
		fmt.Fprintf(os.Stderr, "Decoded to %s\n", outputFile)
	}

	return 0
}

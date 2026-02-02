// Command: drawio create
// Short: Create a new .drawio.png file
// Long: Creates a new DrawIO diagram file with a blank canvas or from
// Long: provided XML content.
// Long:
// Long: The created file uses a minimal 1x1 transparent PNG as the image.
// Long: When opened in DrawIO, the diagram will render from the embedded XML.
// Long:
// Long: Example:
// Long:   drawio create -o diagram.drawio.png
// Long:   drawio create -o diagram.drawio.png --name "Architecture"
// Long:   drawio create -o diagram.drawio.png --xml content.xml
// Flag.output: type=string, shorthand=o, required=true, usage=Output .drawio.png file
// Flag.xml: type=string, usage=XML file to use (default: blank diagram)
// Flag.name: type=string, default=Page-1, usage=Diagram page name
package drawio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/clibase/registry"
)

func init() {
	registry.Register(DrawioCreate)
}

// DrawioCreate creates a new .drawio.png file.
func DrawioCreate() int {
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	var outputFile, xmlFile, name string
	name = "Page-1" // default
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		case "--xml":
			if i+1 < len(args) {
				xmlFile = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				name = args[i+1]
				i++
			}
		}
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

	// Build command args
	cmdArgs := []string{"create", "-o", containerOutput, "--name", name}

	if xmlFile != "" {
		// Make XML path absolute
		if !filepath.IsAbs(xmlFile) {
			xmlFile = filepath.Join(repoRoot, xmlFile)
		}

		// Verify XML file exists
		if _, err := os.Stat(xmlFile); os.IsNotExist(err) {
			log.Errorf("Error: XML file not found: %s", xmlFile)
			return 1
		}

		containerXml, err := TranslateToContainerPath(xmlFile, repoRoot)
		if err != nil {
			log.Errorf("Error: %v", err)
			return 1
		}
		cmdArgs = append(cmdArgs, "--xml", containerXml)
	}

	// Run command
	_, err = RunDrawioCommandWithOutput(repoRoot, cmdArgs)
	if err != nil {
		if strings.Contains(err.Error(), "Created") {
			// Success message in stderr
		} else {
			log.Errorf("Error: %v", err)
			return 1
		}
	}

	fmt.Fprintf(os.Stderr, "Created %s\n", outputFile)

	return 0
}

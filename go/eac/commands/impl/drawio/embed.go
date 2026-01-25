// Command: drawio embed
// Short: Embed XML into a .drawio.png file
// Long: Embeds DrawIO XML data into a PNG file, creating or updating
// Long: the mxfile metadata chunk.
// Long:
// Long: The PNG image portion is not modified - only the embedded XML metadata.
// Long: Open the file in DrawIO to update the visual representation.
// Long:
// Long: If the PNG file doesn't exist, a new file is created with a minimal
// Long: 1x1 transparent PNG (DrawIO will render from XML when opened).
// Long:
// Long: Example:
// Long:   drawio embed --png diagram.drawio.png --xml encoded.xml
// Long:   cat encoded.xml | drawio embed --png diagram.drawio.png
// Flag.png: type=string, required=true, usage=Target PNG file (created if not exists)
// Flag.xml: type=string, usage=XML file to embed (default: stdin)
// Flag.output: type=string, shorthand=o, usage=Output file (default: update --png in place)
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
	registry.Register(DrawioEmbed)
}

// DrawioEmbed embeds XML into a .drawio.png file.
func DrawioEmbed() int {
	args := os.Args[2:]
	if err := flags.ValidateFlagsFromRegistry(args); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	// Parse flags
	var pngFile, xmlFile, outputFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--png":
			if i+1 < len(args) {
				pngFile = args[i+1]
				i++
			}
		case "--xml":
			if i+1 < len(args) {
				xmlFile = args[i+1]
				i++
			}
		case "-o", "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		}
	}

	if pngFile == "" {
		log.Errorf("Error: --png is required")
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

	// Make PNG path absolute
	if !filepath.IsAbs(pngFile) {
		pngFile = filepath.Join(repoRoot, pngFile)
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(pngFile), 0o755); err != nil {
		log.Errorf("Error creating directory: %v", err)
		return 1
	}

	containerPng, err := TranslateToContainerPath(pngFile, repoRoot)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	// Build command args
	cmdArgs := []string{"embed", "--png", containerPng}

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

	if outputFile != "" {
		// Make output path absolute
		if !filepath.IsAbs(outputFile) {
			outputFile = filepath.Join(repoRoot, outputFile)
		}

		containerOutput, err := TranslateToContainerPath(outputFile, repoRoot)
		if err != nil {
			log.Errorf("Error: %v", err)
			return 1
		}
		cmdArgs = append(cmdArgs, "-o", containerOutput)
	}

	// Run command
	_, err = RunDrawioCommandWithOutput(repoRoot, cmdArgs)
	if err != nil {
		if strings.Contains(err.Error(), "Embedded XML into") {
			// Success message in stderr
		} else {
			log.Errorf("Error: %v", err)
			return 1
		}
	}

	target := pngFile
	if outputFile != "" {
		target = outputFile
	}
	fmt.Fprintf(os.Stderr, "Embedded XML into %s\n", target)

	return 0
}

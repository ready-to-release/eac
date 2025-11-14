// Command: templates apply compliance
// Description: Apply compliance templates with optional value replacements
// Usage: go run . templates apply compliance [--source <git-repo-url>] [--destination <path>] [--input-json <file>]
// Flags:
//   --source <git-repo-url>: Git repository URL (default: https://github.com/ready-to-release/eac)
//   --destination <path>: Destination path (default: .docs/references/compliance)
//   --input-json <file>: JSON file with replacement values (optional)
// HasSideEffects: true
package apply

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	templates "github.com/ready-to-release/eac/src/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

const (
	defaultComplianceSource     = "https://github.com/ready-to-release/eac"
	defaultComplianceSourcePath = "templates/compliance"
	defaultComplianceDest       = ".docs/references/compliance"
)

func init() {
	registry.Register(TemplatesApplyCompliance)
}

// complianceConfig holds configuration for the compliance template apply command
type complianceConfig struct {
	Source      string
	SourcePath  string
	Destination string
	ValuesFile  string
}

// TemplatesApplyCompliance applies compliance templates
func TemplatesApplyCompliance() int {
	// Parse command-line flags
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip binary, "templates", "apply", "compliance"
	}

	config, err := parseComplianceFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Resolve relative destination path against initial working directory
	if !filepath.IsAbs(config.Destination) {
		config.Destination = filepath.Join(registry.InitialWorkingDir, config.Destination)
	}

	// Validate source is a Git repository
	if !templates.IsGitRepository(config.Source) {
		fmt.Fprintf(os.Stderr, "Error: --source must be a Git repository URL\n")
		fmt.Fprintf(os.Stderr, "       Provided: %s\n", config.Source)
		return 1
	}

	// Clone repository
	fmt.Printf("Cloning templates from %s...\n", config.Source)
	cloner := templates.NewGitCloner(config.Source)
	clonedDir, err := cloner.CloneToTemp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
		return 1
	}
	defer cloner.Cleanup()

	// Point to the compliance templates subdirectory
	templateDir := filepath.Join(clonedDir, config.SourcePath)

	// Verify template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: template directory does not exist: %s\n", config.SourcePath)
		fmt.Fprintf(os.Stderr, "       In repository: %s\n", config.Source)
		return 1
	}

	fmt.Printf("✓ Templates cloned successfully\n")

	// Load values if provided
	var values templates.TemplateValues
	if config.ValuesFile != "" {
		values, err = templates.LoadValuesFromJSON(config.ValuesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load values: %v\n", err)
			return 1
		}
		fmt.Printf("Loaded %d replacement values\n", len(values))
	}

	// Create renderer
	renderer := templates.NewRenderer(templateDir, config.Destination, values)

	// Render templates
	fmt.Printf("Applying templates to %s...\n", config.Destination)
	if err := renderer.RenderTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to apply templates: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Compliance templates applied successfully to %s\n", config.Destination)
	return 0
}

// parseComplianceFlags parses command-line flags for the compliance apply command
func parseComplianceFlags(args []string) (*complianceConfig, error) {
	fs := flag.NewFlagSet("templates apply compliance", flag.ContinueOnError)

	config := &complianceConfig{
		Source:      defaultComplianceSource,
		SourcePath:  defaultComplianceSourcePath,
		Destination: defaultComplianceDest,
	}

	fs.StringVar(&config.Source, "source", defaultComplianceSource, "Git repository URL")
	fs.StringVar(&config.Destination, "destination", defaultComplianceDest, "Destination path")
	fs.StringVar(&config.ValuesFile, "input-json", "", "JSON file with replacement values (optional)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Validate values file exists if provided
	if config.ValuesFile != "" {
		// Resolve relative path against initial working directory
		if !filepath.IsAbs(config.ValuesFile) {
			config.ValuesFile = filepath.Join(registry.InitialWorkingDir, config.ValuesFile)
		}

		if _, err := os.Stat(config.ValuesFile); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("values file does not exist: %s", config.ValuesFile)
			}
			return nil, fmt.Errorf("cannot access values file: %w", err)
		}
	}

	return config, nil
}

// Command: templates install specs
// Description: Install specification templates to local directory
// Usage: go run . templates install specs [--source <git-repo-url>] [--destination <path>]
// Flags:
//   --source <git-repo-url>: Git repository URL (default: https://github.com/ready-to-release/eac)
//   --destination <path>: Destination path (default: .r2r/templates/specs)
// HasSideEffects: true
package install

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	templates "github.com/ready-to-release/eac/src/commands/impl/templates/internal"
	"github.com/ready-to-release/eac/src/commands/internal/registry"
)

const (
	defaultSpecsSource     = "https://github.com/ready-to-release/eac"
	defaultSpecsSourcePath = "templates/specs"
	defaultSpecsDest       = ".r2r/templates/specs"
)

func init() {
	registry.Register(TemplatesInstallSpecs)
}

// specsConfig holds configuration for the specs template install command
type specsConfig struct {
	Source      string
	SourcePath  string
	Destination string
}

// TemplatesInstallSpecs installs specification templates
func TemplatesInstallSpecs() int {
	// Parse command-line flags
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip binary, "templates", "install", "specs"
	}

	config, err := parseSpecsFlags(args)
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

	// Point to the specs templates subdirectory
	templateDir := filepath.Join(clonedDir, config.SourcePath)

	// Verify template directory exists
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: template directory does not exist: %s\n", config.SourcePath)
		fmt.Fprintf(os.Stderr, "       In repository: %s\n", config.Source)
		return 1
	}

	fmt.Printf("✓ Templates cloned successfully\n")

	// Copy templates without value replacement (install doesn't do replacements)
	renderer := templates.NewRenderer(templateDir, config.Destination, nil)

	// Render templates (no values = simple copy)
	fmt.Printf("Installing templates to %s...\n", config.Destination)
	if err := renderer.RenderTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install templates: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Spec templates installed successfully to %s\n", config.Destination)
	return 0
}

// parseSpecsFlags parses command-line flags for the specs install command
func parseSpecsFlags(args []string) (*specsConfig, error) {
	fs := flag.NewFlagSet("templates install specs", flag.ContinueOnError)

	config := &specsConfig{
		Source:      defaultSpecsSource,
		SourcePath:  defaultSpecsSourcePath,
		Destination: defaultSpecsDest,
	}

	fs.StringVar(&config.Source, "source", defaultSpecsSource, "Git repository URL")
	fs.StringVar(&config.Destination, "destination", defaultSpecsDest, "Destination path")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return config, nil
}

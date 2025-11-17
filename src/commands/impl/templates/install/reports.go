// Command: templates install reports
// Description: Install report templates to local directory
// Usage: go run . templates install reports [--source <local-path>] [--destination <path>]
// Flags:
//   --source <local-path>: Local template directory path (default: clones from https://github.com/ready-to-release/eac, uses templates/reports subdirectory)
//   --destination <path>: Destination path (default: .r2r/templates/reports)
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
	defaultReportsRepoURL    = "https://github.com/ready-to-release/eac"
	defaultReportsSourcePath = "templates/reports"
	defaultReportsDest       = ".r2r/templates/reports"
)

func init() {
	registry.Register(TemplatesInstallReports)
}

// reportsConfig holds configuration for the reports template install command
type reportsConfig struct {
	Source      string // Either GitHub URL (default) or local path (custom)
	Destination string
	isLocalSrc  bool // true if source is a local path
}

// TemplatesInstallReports installs report templates
func TemplatesInstallReports() int {
	// Parse command-line flags
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip binary, "templates", "install", "reports"
	}

	config, err := parseReportsFlags(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Resolve relative destination path against initial working directory
	if !filepath.IsAbs(config.Destination) {
		config.Destination = filepath.Join(registry.InitialWorkingDir, config.Destination)
	}

	var templateDir string
	var cleanup func()

	if config.isLocalSrc {
		// Local source: use path directly
		templateDir = config.Source

		// Verify local template directory exists
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: local template directory does not exist: %s\n", config.Source)
			return 1
		}

		fmt.Printf("Using local templates from %s\n", templateDir)
	} else {
		// Default GitHub URL: clone repository and access subdirectory
		fmt.Printf("Cloning templates from %s...\n", defaultReportsRepoURL)
		cloner := templates.NewGitCloner(defaultReportsRepoURL)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
			return 1
		}
		cleanup = func() {
			cloner.Cleanup()
		}

		// Point to the reports templates subdirectory
		templateDir = filepath.Join(clonedDir, defaultReportsSourcePath)

		// Verify template directory exists in cloned repo
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: template directory does not exist: %s\n", defaultReportsSourcePath)
			fmt.Fprintf(os.Stderr, "       In repository: %s\n", defaultReportsRepoURL)
			cleanup()
			return 1
		}

		fmt.Printf("✓ Templates cloned successfully\n")
	}

	// Ensure cleanup happens if we cloned
	if cleanup != nil {
		defer cleanup()
	}

	// Copy templates without value replacement (install doesn't do replacements)
	renderer := templates.NewRenderer(templateDir, config.Destination, nil)

	// Render templates (no values = simple copy)
	fmt.Printf("Installing templates to %s...\n", config.Destination)
	if err := renderer.RenderTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install templates: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Report templates installed successfully to %s\n", config.Destination)
	return 0
}

// parseReportsFlags parses command-line flags for the reports install command
func parseReportsFlags(args []string) (*reportsConfig, error) {
	fs := flag.NewFlagSet("templates install reports", flag.ContinueOnError)

	config := &reportsConfig{
		Source:      "", // Empty means use default GitHub repo
		Destination: defaultReportsDest,
		isLocalSrc:  false, // Default is GitHub URL
	}

	var customSource string
	fs.StringVar(&customSource, "source", "", "Local template directory path (optional)")
	fs.StringVar(&config.Destination, "destination", defaultReportsDest, "Destination path")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Handle custom source
	if customSource != "" {
		// Custom source must be a local path
		config.isLocalSrc = true

		// Resolve relative path against initial working directory
		if !filepath.IsAbs(customSource) {
			config.Source = filepath.Join(registry.InitialWorkingDir, customSource)
		} else {
			config.Source = customSource
		}
	}

	return config, nil
}

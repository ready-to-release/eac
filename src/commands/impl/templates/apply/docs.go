// Command: templates apply docs
// Description: Apply documentation templates with optional value replacements
// Usage: go run . templates apply docs [--source <local-path>] [--destination <path>] [--input-json <file>]
// Flags:
//   --source <local-path>: Local template directory path (default: clones from https://github.com/ready-to-release/eac, uses templates/docs subdirectory)
//   --destination <path>: Destination path (default: .docs/references/docs)
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
	defaultDocsRepoURL    = "https://github.com/ready-to-release/eac"
	defaultDocsSourcePath = "templates/docs"
	defaultDocsDest       = ".docs/references/docs"
)

func init() {
	registry.Register(TemplatesApplyDocs)
}

// docsConfig holds configuration for the docs template apply command
type docsConfig struct {
	Source      string // Either GitHub URL (default) or local path (custom)
	Destination string
	ValuesFile  string
	isLocalSrc  bool // true if source is a local path
}

// TemplatesApplyDocs applies documentation templates
func TemplatesApplyDocs() int {
	// Parse command-line flags
	args := []string{}
	if len(os.Args) > 4 {
		args = os.Args[4:] // Skip binary, "templates", "apply", "docs"
	}

	config, err := parseDocsFlags(args)
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
		fmt.Printf("Cloning templates from %s...\n", defaultDocsRepoURL)
		cloner := templates.NewGitCloner(defaultDocsRepoURL)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
			return 1
		}
		cleanup = func() {
			cloner.Cleanup()
		}

		// Point to the docs templates subdirectory
		templateDir = filepath.Join(clonedDir, defaultDocsSourcePath)

		// Verify template directory exists in cloned repo
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: template directory does not exist: %s\n", defaultDocsSourcePath)
			fmt.Fprintf(os.Stderr, "       In repository: %s\n", defaultDocsRepoURL)
			cleanup()
			return 1
		}

		fmt.Printf("✓ Templates cloned successfully\n")
	}

	// Ensure cleanup happens if we cloned
	if cleanup != nil {
		defer cleanup()
	}

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

	fmt.Printf("✓ Documentation templates applied successfully to %s\n", config.Destination)
	return 0
}

// parseDocsFlags parses command-line flags for the docs apply command
func parseDocsFlags(args []string) (*docsConfig, error) {
	fs := flag.NewFlagSet("templates apply docs", flag.ContinueOnError)

	config := &docsConfig{
		Source:      "", // Empty means use default GitHub repo
		Destination: defaultDocsDest,
		isLocalSrc:  false, // Default is GitHub URL
	}

	var customSource string
	fs.StringVar(&customSource, "source", "", "Local template directory path (optional)")
	fs.StringVar(&config.Destination, "destination", defaultDocsDest, "Destination path")
	fs.StringVar(&config.ValuesFile, "input-json", "", "JSON file with replacement values (optional)")

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

// Command: templates install
// Description: Install templates from a Git repository with value replacements
// Usage: go run . templates install [--template <git-repo-url>] --values <json-file> --location <output-path>
// Flags:
//   --template <git-repo-url>: Git repository URL to clone templates from (default: https://github.com/ready-to-release/eac)
//   --values <json-file>: JSON file with replacement values (required)
//   --location <output-path>: Output location for rendered templates (required)
// HasSideEffects: true
package templates

import (
	"github.com/ready-to-release/eac/src/commands/internal/registry"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/src/commands/impl/templates/internal"
)

const defaultTemplateRepo = "https://github.com/ready-to-release/eac"

func init() {
	registry.Register(TemplatesInstall)
}

// TemplatesInstall installs templates from a Git repository with value replacements
func TemplatesInstall() int {
	// Create flagset for this command
	fs := flag.NewFlagSet("templates install", flag.ExitOnError)

	repoURL := fs.String("template", defaultTemplateRepo, "Git repository URL to clone templates from (default: https://github.com/ready-to-release/eac)")
	valuesFile := fs.String("values", "", "JSON file with replacement values (required)")
	location := fs.String("location", "", "Output location for rendered templates (required)")

	// Parse flags - skip binary path, "templates" and "install" from os.Args
	args := []string{}
	if len(os.Args) > 3 {
		args = os.Args[3:]
	}
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	// Validate required flags (template now has default)
	if *valuesFile == "" {
		fmt.Fprintln(os.Stderr, "Error: --values flag is required")
		fs.PrintDefaults()
		return 1
	}
	if *location == "" {
		fmt.Fprintln(os.Stderr, "Error: --location flag is required")
		fs.PrintDefaults()
		return 1
	}

	// Resolve relative paths against the initial working directory
	if !filepath.IsAbs(*location) {
		*location = filepath.Join(registry.InitialWorkingDir, *location)
	}
	if !filepath.IsAbs(*valuesFile) {
		*valuesFile = filepath.Join(registry.InitialWorkingDir, *valuesFile)
	}

	// Validate that template is a Git URL
	if !templates.IsGitRepository(*repoURL) {
		fmt.Fprintf(os.Stderr, "Error: --template must be a Git repository URL (e.g., https://github.com/user/repo)\n")
		fmt.Fprintf(os.Stderr, "       Provided: %s\n", *repoURL)
		return 1
	}

	// Validate values file exists
	if _, err := os.Stat(*valuesFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: values file does not exist: %s\n", *valuesFile)
		} else {
			fmt.Fprintf(os.Stderr, "Error: cannot access values file: %v\n", err)
		}
		return 1
	}

	// Clone repository to temp directory
	fmt.Printf("Cloning templates from %s...\n", *repoURL)
	cloner := templates.NewGitCloner(*repoURL)
	clonedDir, err := cloner.CloneToTemp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
		return 1
	}
	defer cloner.Cleanup() // Clean up temp directory when done

	// Point to the templates subdirectory within the cloned repository
	templateDir := filepath.Join(clonedDir, "templates")

	fmt.Printf("✓ Templates cloned successfully\n")

	// Load values from JSON
	values, err := templates.LoadValuesFromJSON(*valuesFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load values: %v\n", err)
		return 1
	}

	// Validate that ProjectName is provided (required)
	if err := templates.ValidateValues(values, []string{"ProjectName"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Warn if location exists (will overwrite)
	if _, err := os.Stat(*location); err == nil {
		fmt.Printf("Warning: Location exists and will be overwritten: %s\n", *location)
		fmt.Println("         (You may need to resolve merge conflicts manually)")
	}

	// Create renderer
	renderer := templates.NewRenderer(templateDir, *location, values)

	// Render templates (will overwrite existing files)
	fmt.Printf("Rendering templates to %s...\n", *location)
	if err := renderer.RenderTemplates(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render templates: %v\n", err)
		return 1
	}

	fmt.Printf("✓ Templates installed successfully to %s\n", *location)
	fmt.Println("\nNote: If files already existed, they have been overwritten.")
	fmt.Println("      Review changes with 'git diff' and resolve any conflicts.")
	return 0
}

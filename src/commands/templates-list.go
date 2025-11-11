// Command: templates list
// Description: List all placeholder variables found in template files
// Usage: go run . templates list [--template <git-repo-url|local-path>]
// Flags:
//   --template <git-repo-url|local-path>: Git repository URL or local directory to scan (default: https://github.com/ready-to-release/eac)
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/templates"
)

const defaultTemplateListRepo = "https://github.com/ready-to-release/eac"

func init() {
	Register("templates list", TemplatesList)
}

// TemplatesList scans templates and lists all placeholder variables
func TemplatesList() int {
	// Create flagset for this command
	fs := flag.NewFlagSet("templates list", flag.ExitOnError)

	template := fs.String("template", defaultTemplateListRepo, "Git repository URL or local directory to scan (default: https://github.com/ready-to-release/eac)")

	// Parse flags - skip "templates" and "list" from os.Args
	args := os.Args[3:]
	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 1
	}

	var templateDir string
	var cleanup func()

	// Check if template is a Git repository URL or local path
	if templates.IsGitRepository(*template) {
		// Clone repository to temp directory
		fmt.Printf("Cloning templates from %s...\n", *template)
		cloner := templates.NewGitCloner(*template)
		clonedDir, err := cloner.CloneToTemp()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clone repository: %v\n", err)
			return 1
		}
		templateDir = clonedDir
		cleanup = func() {
			if err := cloner.Cleanup(); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to cleanup temp directory: %v\n", err)
			}
		}
		fmt.Printf("✓ Templates cloned successfully\n\n")
	} else {
		// Use local directory
		templateDir = *template
		cleanup = func() {} // No cleanup needed for local paths
	}

	// Ensure cleanup happens
	defer cleanup()

	// Create scanner
	scanner := templates.NewPlaceholderScanner(templateDir)

	// Scan templates with location information
	placeholderInfos, err := scanner.ScanWithLocations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to scan templates: %v\n", err)
		return 1
	}

	// Display placeholders
	displayPlaceholders(*template, placeholderInfos)
	return 0
}

// displayPlaceholders prints the found placeholder variables with their locations
func displayPlaceholders(templateDir string, placeholderInfos []templates.PlaceholderInfo) {
	fmt.Printf("Template Placeholders in '%s':\n", templateDir)
	fmt.Println("----------------------------")

	if len(placeholderInfos) == 0 {
		fmt.Println("No placeholders found.")
		return
	}

	// Display each placeholder with its file locations
	for _, info := range placeholderInfos {
		fmt.Printf("  {{ .%s }}\n", info.Name)
		for _, file := range info.Files {
			fmt.Printf("    - %s\n", file)
		}
	}

	fmt.Printf("\nTotal: %d placeholders\n", len(placeholderInfos))
	fmt.Println("\nTo use these templates, provide a values.json file with these keys:")
	fmt.Println("{")
	for i, info := range placeholderInfos {
		if i == len(placeholderInfos)-1 {
			fmt.Printf("  \"%s\": \"value\"\n", info.Name)
		} else {
			fmt.Printf("  \"%s\": \"value\",\n", info.Name)
		}
	}
	fmt.Println("}")
}

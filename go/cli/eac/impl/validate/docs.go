// Command: validate docs
// Short: Validate documentation for obsolete references
// Long: Validates documentation files for references to obsolete or deleted configuration files.
// Long:
// Long: This command helps maintain documentation accuracy by detecting references to files
// Long: that no longer exist in the codebase, such as module-types.yml and system-dependencies.yml.
// Long:
// Long: Expected Output:
// Long:   Lists files containing obsolete references.
// Long:   Shows line numbers and the obsolete reference found.
// Long:   Exit code 0 if no issues, 1 if validation errors found.
package validate

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/repository"
)

// DocsValidationResult represents a single validation issue found in documentation.
type DocsValidationResult struct {
	File        string
	Line        int
	Message     string
	ObsoleteRef string
}

// DocsValidatorOptions configures the documentation validator.
type DocsValidatorOptions struct {
	// ObsoleteFiles is a list of file names that should not be referenced
	ObsoleteFiles []string
	// ExcludeDirs is a list of directory names to skip (e.g., "assets/cache")
	ExcludeDirs []string
	// ExcludeFiles is a list of file path patterns to skip (e.g., deprecation notices)
	ExcludeFiles []string
}

// DefaultDocsValidatorOptions returns the default validator options with known obsolete files.
func DefaultDocsValidatorOptions() DocsValidatorOptions {
	return DocsValidatorOptions{
		ObsoleteFiles: []string{
			"module-types.yml",
			"system-dependencies.yml",
		},
		ExcludeDirs: []string{
			"assets/cache",
			".git",
			"node_modules",
		},
		ExcludeFiles: []string{
			// Deprecation notices that intentionally reference old file names
			"module-types.md",
			"creating-module-types.md",
		},
	}
}

// DocsValidator validates documentation files for obsolete references.
type DocsValidator struct {
	opts   DocsValidatorOptions
	output io.Writer
}

// NewDocsValidator creates a new documentation validator.
func NewDocsValidator(opts DocsValidatorOptions) *DocsValidator {
	return &DocsValidator{
		opts:   opts,
		output: os.Stdout,
	}
}

// ValidateFile validates a single markdown file for obsolete references.
func (v *DocsValidator) ValidateFile(filePath string) ([]DocsValidationResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var results []DocsValidationResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, obsoleteFile := range v.opts.ObsoleteFiles {
			if strings.Contains(line, obsoleteFile) {
				results = append(results, DocsValidationResult{
					File:        filePath,
					Line:        lineNum,
					Message:     fmt.Sprintf("references obsolete file '%s'", obsoleteFile),
					ObsoleteRef: obsoleteFile,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return results, nil
}

// ValidateDirectory validates all markdown files in a directory recursively.
func (v *DocsValidator) ValidateDirectory(dirPath string) (map[string][]DocsValidationResult, error) {
	results := make(map[string][]DocsValidationResult)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip excluded directories
		if info.IsDir() {
			for _, excludeDir := range v.opts.ExcludeDirs {
				if strings.Contains(path, excludeDir) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip excluded files (e.g., deprecation notices)
		fileName := filepath.Base(path)
		for _, excludeFile := range v.opts.ExcludeFiles {
			if fileName == excludeFile {
				return nil
			}
		}

		fileResults, err := v.ValidateFile(path)
		if err != nil {
			return err
		}

		if len(fileResults) > 0 {
			results[path] = fileResults
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	return results, nil
}

// PrintResults prints validation results and returns exit code.
func (v *DocsValidator) PrintResults(results map[string][]DocsValidationResult, repoRoot string) int {
	if len(results) == 0 {
		fmt.Fprintln(v.output, "✓ No obsolete references found in documentation")
		return 0
	}

	totalIssues := 0
	for filePath, fileResults := range results {
		relPath, _ := filepath.Rel(repoRoot, filePath)
		if relPath == "" {
			relPath = filePath
		}

		for _, result := range fileResults {
			totalIssues++
			fmt.Fprintf(v.output, "%s:%d: %s\n", relPath, result.Line, result.Message)
		}
	}

	fmt.Fprintf(v.output, "\n✗ Found %d obsolete reference(s) in %d file(s)\n", totalIssues, len(results))
	return 1
}

// ValidateDocs validates all documentation files for obsolete references.
func ValidateDocs() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	args := os.Args[3:] // Skip program name, "validate", and "docs"

	// Check for help flag
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		printDocsUsage()
		return 0
	}

	// Get repository root
	repoRoot, err := repository.GetRepositoryRoot("")
	if err != nil {
		log.Errorf("Error: failed to get repository root: %v", err)
		return 1
	}

	// Default docs directory
	docsDir := filepath.Join(repoRoot, "docs")

	// Allow overriding docs directory
	if len(args) > 0 {
		docsDir = args[0]
		if !filepath.IsAbs(docsDir) {
			docsDir = filepath.Join(repoRoot, docsDir)
		}
	}

	// Check docs directory exists
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		log.Errorf("Error: documentation directory not found: %s", docsDir)
		return 1
	}

	// Create validator with default options
	opts := DefaultDocsValidatorOptions()
	validator := NewDocsValidator(opts)

	// Validate all documentation files
	results, err := validator.ValidateDirectory(docsDir)
	if err != nil {
		log.Errorf("Error: failed to validate documentation: %v", err)
		return 1
	}

	// Print results
	return validator.PrintResults(results, repoRoot)
}

func printDocsUsage() {
	log.Info("Validate documentation for obsolete references")
	log.Info("")
	log.Info("Usage: clie validate docs [docs-directory]")
	log.Info("")
	log.Info("Checks for:")
	log.Info("  - References to deleted files (module-types.yml)")
	log.Info("  - References to obsolete files (system-dependencies.yml)")
	log.Info("")
	log.Info("Examples:")
	log.Info("  # Validate all documentation files")
	log.Info("  clie validate docs")
	log.Info("")
	log.Info("  # Validate specific directory")
	log.Info("  clie validate docs docs/reference")
}

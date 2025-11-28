package assessment

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AnalysisInput contains the data needed for risk analysis
type AnalysisInput struct {
	Files         []FileInfo
	Specifications []SpecInfo
	Scope         string
	WorkspaceRoot string
}

// SpecInfo represents a specification file
type SpecInfo struct {
	Path    string
	Content string
}

// prepareAnalysisInput prepares all input data for AI analysis
func prepareAnalysisInput(config *Config, files []FileInfo) (*AnalysisInput, error) {
	input := &AnalysisInput{
		Files:         files,
		Scope:         config.Scope,
		WorkspaceRoot: config.WorkspaceRoot,
	}

	// Load specifications
	specs, err := loadSpecifications(config.WorkspaceRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load specifications: %w", err)
	}
	input.Specifications = specs

	return input, nil
}

// loadSpecifications loads all specification files from the specs directory
func loadSpecifications(workspaceRoot string) ([]SpecInfo, error) {
	specsDir := filepath.Join(workspaceRoot, "specs")

	// Check if specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		// Specs directory doesn't exist, return empty list
		return []SpecInfo{}, nil
	}

	var specs []SpecInfo

	err := filepath.Walk(specsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .feature files
		if !info.IsDir() && strings.HasSuffix(path, ".feature") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read spec %s: %w", path, err)
			}

			// Make path relative to workspace root
			relPath, err := filepath.Rel(workspaceRoot, path)
			if err != nil {
				relPath = path
			}

			specs = append(specs, SpecInfo{
				Path:    relPath,
				Content: string(content),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return specs, nil
}

// formatChangedFiles formats the file list for AI prompt
func formatChangedFiles(files []FileInfo, workspaceRoot string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Total files: %d\n\n", len(files)))

	for _, file := range files {
		builder.WriteString(fmt.Sprintf("**File:** %s\n", file.Path))
		builder.WriteString(fmt.Sprintf("**Status:** %s\n", file.Status))

		// Try to read file content (skip deleted files)
		if file.Status != "deleted" {
			content, err := readFileContent(file.Path, workspaceRoot)
			if err == nil && len(content) > 0 {
				// Limit content to first 1000 characters for performance
				if len(content) > 1000 {
					content = content[:1000] + "\n... (truncated)"
				}
				builder.WriteString(fmt.Sprintf("**Content:**\n```\n%s\n```\n", content))
			}
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

// formatSpecifications formats specifications for AI prompt
func formatSpecifications(specs []SpecInfo) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Total specifications: %d\n\n", len(specs)))

	for _, spec := range specs {
		builder.WriteString(fmt.Sprintf("**Spec:** %s\n", spec.Path))

		// Limit content for performance
		content := spec.Content
		if len(content) > 2000 {
			content = content[:2000] + "\n... (truncated)"
		}
		builder.WriteString(fmt.Sprintf("**Content:**\n```gherkin\n%s\n```\n\n", content))
	}

	return builder.String()
}

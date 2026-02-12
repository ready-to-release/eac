package toolhandler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// loadInput loads the input content for analysis based on analysis type.
func (h *AIToolHandler) loadInput(module core.ModuleContractPort, workspaceRoot, component string) (string, error) {
	moniker := module.GetMoniker()

	switch h.analysisType {
	case AIAnalysisTypeDSL:
		return h.loadDSLContent(module, workspaceRoot)
	case AIAnalysisTypeSpecs:
		return h.loadSpecsContent(module, workspaceRoot)
	case AIAnalysisTypeSource:
		return h.loadSourceContent(module, workspaceRoot)
	case AIAnalysisTypeDocs:
		return h.loadDocsContent(workspaceRoot, moniker)
	default:
		return "", fmt.Errorf("unknown analysis type: %s", h.analysisType)
	}
}

// loadDSLContent loads Structurizr DSL files for a module.
// Uses the module's structurizr component root from config-driven discovery.
func (h *AIToolHandler) loadDSLContent(module core.ModuleContractPort, workspaceRoot string) (string, error) {
	root := module.GetComponentRoot("structurizr")
	if root == "" {
		// Fallback for modules without structurizr component
		return "", fmt.Errorf("no structurizr component found for module %s", module.GetMoniker())
	}
	designDir := filepath.Join(workspaceRoot, root)
	return h.loadFilesWithExtension(designDir, ".dsl")
}

// loadSpecsContent loads Gherkin specification files for a module.
// Uses the module's gherkin component root from config-driven discovery.
func (h *AIToolHandler) loadSpecsContent(module core.ModuleContractPort, workspaceRoot string) (string, error) {
	root := module.GetComponentRoot("gherkin")
	if root == "" {
		// Fallback for modules without gherkin component
		return "", fmt.Errorf("no gherkin component found for module %s", module.GetMoniker())
	}
	specsDir := filepath.Join(workspaceRoot, root)
	return h.loadFilesWithExtension(specsDir, ".feature")
}

// loadSourceContent loads source code from the module's component roots.
func (h *AIToolHandler) loadSourceContent(module core.ModuleContractPort, workspaceRoot string) (string, error) {
	var builder strings.Builder

	// Get all component roots
	roots := module.GetComponentRoots()
	for compType, root := range roots {
		fullPath := filepath.Join(workspaceRoot, root)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		builder.WriteString(fmt.Sprintf("## Component: %s (path: %s)\n\n", compType, root))

		// Load source files based on component type
		var content string
		var err error
		switch compType {
		case "go":
			content, err = h.loadFilesWithExtension(fullPath, ".go")
		case "typescript", "javascript":
			content, err = h.loadFilesWithExtensions(fullPath, []string{".ts", ".tsx", ".js", ".jsx"})
		case "python":
			content, err = h.loadFilesWithExtension(fullPath, ".py")
		default:
			// Try common source extensions
			content, err = h.loadFilesWithExtensions(fullPath, []string{".go", ".ts", ".js", ".py", ".rs", ".cs"})
		}

		if err != nil {
			continue
		}
		builder.WriteString(content)
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// loadDocsContent loads markdown documentation files.
func (h *AIToolHandler) loadDocsContent(workspaceRoot, moniker string) (string, error) {
	docsDir := filepath.Join(workspaceRoot, "docs")
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		return "", nil
	}
	return h.loadFilesWithExtension(docsDir, ".md")
}

// loadFilesWithExtension loads all files with a specific extension from a directory.
func (h *AIToolHandler) loadFilesWithExtension(dir, ext string) (string, error) {
	return h.loadFilesWithExtensions(dir, []string{ext})
}

// loadFilesWithExtensions loads all files with any of the specified extensions.
func (h *AIToolHandler) loadFilesWithExtensions(dir string, extensions []string) (string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", nil
	}

	var builder strings.Builder
	maxFiles := 50 // Limit to prevent huge prompts

	fileCount := 0
	limitReached := false
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Debugf("toolhandler: skipping path %s due to walk error: %v", path, err)
			return nil
		}
		if info.IsDir() {
			// Skip vendor, node_modules, hidden directories
			name := info.Name()
			if name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Check extension
		ext := filepath.Ext(path)
		isMatch := false
		for _, e := range extensions {
			if ext == e {
				isMatch = true
				break
			}
		}
		if !isMatch {
			return nil
		}

		// Limit number of files
		if fileCount >= maxFiles {
			if !limitReached {
				limitReached = true
				log.Warnf("toolhandler: file limit reached (%d files) while loading from %s; remaining files skipped", maxFiles, dir)
			}
			return nil
		}
		fileCount++

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			log.Debugf("toolhandler: failed to read file %s: %v", path, err)
			return nil
		}

		relPath, _ := filepath.Rel(dir, path)
		builder.WriteString(fmt.Sprintf("### File: %s\n```\n%s\n```\n\n", relPath, string(content)))

		return nil
	})

	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

// loadPriorResults loads results from previous analysis steps that this analysis depends on.
func (h *AIToolHandler) loadPriorResults(outputDir string) (string, error) {
	if h.analysisType != AIAnalysisTypeSource {
		return "", nil // Only source analysis depends on prior results
	}

	var builder strings.Builder

	// Load DSL analysis results
	dslResults := filepath.Join(outputDir, "dsl-status.md")
	if content, err := os.ReadFile(dslResults); err == nil {
		builder.WriteString("## Prior Analysis: Architecture (DSL)\n\n")
		builder.WriteString(string(content))
		builder.WriteString("\n\n")
	}

	// Load specs analysis results
	specsResults := filepath.Join(outputDir, "specs-status.md")
	if content, err := os.ReadFile(specsResults); err == nil {
		builder.WriteString("## Prior Analysis: Specifications (Gherkin)\n\n")
		builder.WriteString(string(content))
		builder.WriteString("\n\n")
	}

	return builder.String(), nil
}

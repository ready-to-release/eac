// Package toolhandler provides an AI-specific build handler that bridges
// the tool system to the AI adapter for automated analysis.
package toolhandler

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/ai"
	"github.com/ready-to-release/eac/go/adapters/ai/providers"
	coreai "github.com/ready-to-release/eac/go/core/ai/generation"
	"github.com/ready-to-release/eac/go/core/tool"
)

// AIAnalysisType represents the type of AI analysis to perform.
type AIAnalysisType string

const (
	// AIAnalysisTypeDSL analyzes Structurizr DSL files.
	AIAnalysisTypeDSL AIAnalysisType = "dsl"

	// AIAnalysisTypeSpecs analyzes Gherkin specification files.
	AIAnalysisTypeSpecs AIAnalysisType = "specs"

	// AIAnalysisTypeSource analyzes source code.
	AIAnalysisTypeSource AIAnalysisType = "source"

	// AIAnalysisTypeDocs analyzes documentation files.
	AIAnalysisTypeDocs AIAnalysisType = "docs"
)

// AIToolHandler implements BuildHandler for AI analysis tools.
// It bridges the tool system to the AI adapter for automated analysis.
type AIToolHandler struct {
	toolDef      *tool.ToolDefinition
	analysisType AIAnalysisType
}

// NewAIToolHandler creates a new AI tool handler for the given tool definition.
// The analysis type is inferred from the tool ID (e.g., "ai-dsl-analyzer" -> "dsl").
func NewAIToolHandler(toolDef *tool.ToolDefinition) *AIToolHandler {
	analysisType := inferAnalysisType(toolDef.ID)
	return &AIToolHandler{
		toolDef:      toolDef,
		analysisType: analysisType,
	}
}

// NewAIToolHandlerWithType creates a new AI tool handler with an explicit analysis type.
func NewAIToolHandlerWithType(toolDef *tool.ToolDefinition, analysisType AIAnalysisType) *AIToolHandler {
	return &AIToolHandler{
		toolDef:      toolDef,
		analysisType: analysisType,
	}
}

// inferAnalysisType extracts the analysis type from a tool ID.
// Examples: "ai-dsl-analyzer" -> "dsl", "ai-specs-analyzer" -> "specs"
func inferAnalysisType(toolID string) AIAnalysisType {
	// Remove "ai-" prefix and "-analyzer" suffix
	id := strings.TrimPrefix(toolID, "ai-")
	id = strings.TrimSuffix(id, "-analyzer")

	switch id {
	case "dsl":
		return AIAnalysisTypeDSL
	case "specs":
		return AIAnalysisTypeSpecs
	case "source":
		return AIAnalysisTypeSource
	case "docs":
		return AIAnalysisTypeDocs
	default:
		return AIAnalysisType(id)
	}
}

// Name returns the handler identifier.
func (h *AIToolHandler) Name() string {
	return h.toolDef.ID
}

// Build executes the AI analysis for a module.
// Returns exit code (0 = success, non-zero = failure).
func (h *AIToolHandler) Build(module core.ModuleContractPort, workspaceRoot, outputDir string,
	logWriter io.Writer, opts tool.BuildOptions) int {

	// Create AI executor
	executor := ai.NewExecutor(workspaceRoot)
	providers.RegisterBuiltIn(executor)
	executorAdapter := ai.NewExecutorAdapter(executor)

	// Load input content based on analysis type
	content, err := h.loadInput(module, workspaceRoot, opts.Component)
	if err != nil {
		fmt.Fprintf(logWriter, "Error loading input: %v\n", err)
		return 1
	}

	if content == "" {
		fmt.Fprintf(logWriter, "No content found for %s analysis\n", h.analysisType)
		return 0 // No content is not an error
	}

	// Load prior results if this component depends on others
	priorResults, err := h.loadPriorResults(outputDir)
	if err != nil {
		fmt.Fprintf(logWriter, "Warning: failed to load prior results: %v\n", err)
		// Continue without prior results
	}

	// Build prompt
	prompt := h.buildPrompt(module.GetMoniker(), content, priorResults)

	// Build retry configuration
	retryConfig, err := coreai.BuildRetryConfig(
		"ai-summary-"+string(h.analysisType),
		coreai.FormatPlainText,
		executorAdapter,
		nil, // No validator for plain text
		workspaceRoot,
		nil, // Use default retry config
		coreai.WithDefaultMaxAttempts(2),
	)
	if err != nil {
		fmt.Fprintf(logWriter, "Error building retry config: %v\n", err)
		return 1
	}

	// Generate with retry
	fmt.Fprintf(logWriter, "Analyzing %s...\n", h.analysisType)
	result, err := coreai.GenerateWithRetry(context.Background(), retryConfig, prompt)
	if err != nil {
		fmt.Fprintf(logWriter, "Error: AI generation failed: %v\n", err)
		return 1
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(logWriter, "Error creating output directory: %v\n", err)
		return 1
	}

	// Write output
	outputFile := filepath.Join(outputDir, string(h.analysisType)+"-status.md")
	if err := os.WriteFile(outputFile, []byte(result.Output), 0o644); err != nil {
		fmt.Fprintf(logWriter, "Error writing output: %v\n", err)
		return 1
	}

	fmt.Fprintf(logWriter, "Analysis complete: %s\n", outputFile)
	return 0
}

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
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
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
			return nil
		}
		fileCount++

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
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

// buildPrompt constructs the AI prompt for the analysis.
func (h *AIToolHandler) buildPrompt(moniker, content, priorResults string) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("# AI Analysis: %s for module '%s'\n\n", h.analysisType, moniker))

	switch h.analysisType {
	case AIAnalysisTypeDSL:
		builder.WriteString(h.dslPromptInstructions())
	case AIAnalysisTypeSpecs:
		builder.WriteString(h.specsPromptInstructions())
	case AIAnalysisTypeSource:
		builder.WriteString(h.sourcePromptInstructions())
	case AIAnalysisTypeDocs:
		builder.WriteString(h.docsPromptInstructions())
	}

	if priorResults != "" {
		builder.WriteString("\n## Context from Prior Analysis\n\n")
		builder.WriteString(priorResults)
	}

	builder.WriteString("\n## Content to Analyze\n\n")
	builder.WriteString(content)

	builder.WriteString("\n\n## Generate Analysis\n\n")
	builder.WriteString("Generate a comprehensive analysis in markdown format. Include:\n")
	builder.WriteString("- Summary of findings\n")
	builder.WriteString("- Key observations\n")
	builder.WriteString("- Potential issues or areas for improvement\n")
	builder.WriteString("- Recommendations\n")

	return builder.String()
}

func (h *AIToolHandler) dslPromptInstructions() string {
	return `You are analyzing Structurizr DSL architecture documentation.

Focus on:
- Architecture completeness and clarity
- Component relationships and dependencies
- Alignment with best practices
- Missing or unclear elements
- Suggestions for improvement
`
}

func (h *AIToolHandler) specsPromptInstructions() string {
	return `You are analyzing Gherkin BDD specifications.

Focus on:
- Specification completeness and coverage
- Clarity and readability of scenarios
- Missing edge cases or scenarios
- Consistency in language and structure
- Alignment with business requirements
`
}

func (h *AIToolHandler) sourcePromptInstructions() string {
	return `You are analyzing source code, taking into account the architecture and specifications from prior analysis.

Focus on:
- Code quality and maintainability
- Alignment with architecture documentation
- Implementation of specified behaviors
- Potential bugs or issues
- Code organization and patterns
`
}

func (h *AIToolHandler) docsPromptInstructions() string {
	return `You are analyzing documentation files.

Focus on:
- Documentation completeness
- Clarity and accuracy
- Consistency with code and architecture
- Missing or outdated sections
- Suggestions for improvement
`
}

// ListArtifacts returns artifact paths that would be produced.
func (h *AIToolHandler) ListArtifacts(module core.ModuleContractPort, workspaceRoot string) []string {
	return []string{string(h.analysisType) + "-status.md"}
}

// Requirements returns system dependencies required by this handler.
func (h *AIToolHandler) Requirements() []string {
	return h.toolDef.Requirements
}

// ValidateModule checks if the module configuration is valid.
func (h *AIToolHandler) ValidateModule(module core.ModuleContractPort, workspaceRoot, component string) error {
	return nil // AI tools have minimal validation
}

// IsContainer returns true if this handler runs in a Docker container.
func (h *AIToolHandler) IsContainer() bool {
	return false
}

// IsHostInstalled returns true if this handler runs using host-installed tools.
func (h *AIToolHandler) IsHostInstalled() bool {
	return true
}

// IsAITool returns true - identifies this as an AI analysis tool.
func (h *AIToolHandler) IsAITool() bool {
	return true
}

// GetWeight returns the scheduling weight for this tool.
func (h *AIToolHandler) GetWeight() int {
	if h.toolDef.Resources != nil {
		return h.toolDef.Resources.Weight()
	}
	return 3 // Default weight for AI tools
}

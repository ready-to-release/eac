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
	build "github.com/ready-to-release/eac/contracts/runner/0.1.0/build"
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

// AIToolHandler implements build.BuilderPort for AI analysis tools.
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
	logWriter io.Writer, rawOpts any) int {
	opts, _ := rawOpts.(tool.BuildOptions)

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

var _ build.BuilderPort = (*AIToolHandler)(nil)

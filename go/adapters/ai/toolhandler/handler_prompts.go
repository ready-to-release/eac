package toolhandler

import (
	"fmt"
	"strings"
)

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

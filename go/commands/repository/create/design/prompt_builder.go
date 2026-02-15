package design

import (
	"fmt"
	"strings"

	design "github.com/ready-to-release/eac/go/commands/repository/design"
	coreai "github.com/ready-to-release/eac/go/core/ai"
	"github.com/ready-to-release/eac/go/core/paths"
)

// loadAndBuildPrompt loads the contract and builds the AI prompt.
func loadAndBuildPrompt(config *DesignConfig, out *design.Output) (string, error) {
	out.Progress("📋 Loading design contract...")

	fullPrompt, err := buildContractBasedPrompt(config)
	if err != nil {
		return "", err
	}

	if config.Debug {
		log.Debugf("Full AI prompt built: promptLength=%d", len(fullPrompt))
	}

	return fullPrompt, nil
}

// buildContractBasedPrompt builds the AI prompt with contract context.
func buildContractBasedPrompt(config *DesignConfig) (string, error) {
	// Load contract using generalized loader
	loader := coreai.NewContractLoader(config.TemplateRoot, coreai.TypeDesign, paths.DefaultsVersion)

	contractData, err := loader.LoadContract()
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %w", err)
	}

	// Load prompt (default or custom)
	promptContent, err := loadPrompt(config)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Build prompt template
	promptTemplate, err := coreai.BuildPromptWithTemplate(
		promptContent,
		contractData,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to build prompt template: %w", err)
	}

	// Build final prompt with module context
	var prompt strings.Builder
	prompt.WriteString(promptTemplate)
	prompt.WriteString("\n\n>>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<\n\n")

	// Module context
	prompt.WriteString("## Target Module\n\n")
	prompt.WriteString("Module: ")
	prompt.WriteString(config.Module)
	prompt.WriteString("\n\n")
	prompt.WriteString("Analyze the source code in the following directory and generate architecture documentation:\n\n")
	prompt.WriteString("Source Path: ")
	prompt.WriteString(config.SourcePath)
	prompt.WriteString("\n\n")
	prompt.WriteString("Use the naming format: ")
	prompt.WriteString(config.Module)
	prompt.WriteString(" Architecture\n\n")

	// Analysis instructions
	prompt.WriteString("## Analysis Instructions\n\n")
	prompt.WriteString("Examine the module's source code to understand:\n")
	prompt.WriteString("- Main components and their responsibilities\n")
	prompt.WriteString("- Dependencies between components\n")
	prompt.WriteString("- External systems or services the module interacts with\n")
	prompt.WriteString("- Data flow and relationships\n\n")

	// Generation requirements
	prompt.WriteString("## Generation Requirements\n\n")
	prompt.WriteString("Generate a COMPLETE architecture including:\n")
	prompt.WriteString("- System Context view: Show the main system with external actors and systems\n")
	prompt.WriteString("- Container view: Break down the main system into major containers\n")
	prompt.WriteString("- Component views: For each significant container, show internal components\n")
	prompt.WriteString("- Relationships: Connect all elements with descriptive relationships\n\n")

	// Final instruction
	prompt.WriteString("## Generate Now\n\n")
	prompt.WriteString("Generate the complete Structurizr DSL workspace now.\n")
	prompt.WriteString("Return ONLY the DSL content starting with 'workspace' - no markdown fences, no explanations.\n")

	return prompt.String(), nil
}

// loadPrompt loads the AI prompt for design generation with three-tier priority:
// 1. Command flag (--prompt)
// 2. Team override (.eac/templates/coreai.TypeDesign/design.md)
// 3. System default (templates/coreai.TypeDesign/design.md)
// Convention: Empty string uses type name (design.md).
func loadPrompt(config *DesignConfig) (string, error) {
	// Load prompt with three-tier priority system
	loader := coreai.NewContractLoader(config.TemplateRoot, coreai.TypeDesign, "")
	prompt, source, err := loader.LoadPromptWithPriority("", config.PromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to load prompt: %w", err)
	}

	// Log source if not default
	if source != "embedded fallback" && config.Debug {
		log.Errorf("ℹ️  Using %s prompt", source)
	}

	return prompt, nil
}

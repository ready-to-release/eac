// File: go/cli/eac/impl/init/ai_generation.go
package init

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/ready-to-release/eac/go/core/environments"
)

// GenerateConfig uses AI to generate repository configuration from scan results.
// It loads the AI prompt template, executes it with scan results as custom data,
// and returns the generated YAML configuration.
//
// Parameters:
//   - repoRoot: Absolute path to repository root
//   - scanResult: Results from ScanRepository containing detected modules
//   - aiProvider: AI provider to use (e.g., "claude-api", "openai")
//
// Returns:
//   - Generated YAML configuration as string
//   - Error if template loading, AI execution, or parsing fails
func GenerateConfig(repoRoot string, scanResult *ScanResult, aiProvider string) (string, error) {
	return generateConfig(defaultDeps(), repoRoot, scanResult, aiProvider)
}

// generateConfig is the internal implementation of GenerateConfig with injectable dependencies.
func generateConfig(deps *Deps, repoRoot string, scanResult *ScanResult, aiProvider string) (string, error) {
	// Validate inputs
	if repoRoot == "" {
		return "", fmt.Errorf("repository root cannot be empty")
	}
	if scanResult == nil {
		return "", fmt.Errorf("scan result cannot be nil")
	}

	// Load AI prompt template
	promptTemplate, err := loadPromptTemplate(repoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to load AI prompt template: %w", err)
	}

	// Prepare custom data for template
	customData, err := prepareScanData(scanResult, repoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to prepare scan data: %w", err)
	}

	// Execute template to generate final prompt
	promptInput, err := executeTemplate(promptTemplate, customData)
	if err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}

	// Get AI executor
	executor := deps.GetAIExecutor(repoRoot)

	// Execute AI prompt
	ctx := context.Background()
	response, err := executor.Execute(ctx, promptInput)
	if err != nil {
		return "", fmt.Errorf("AI provider execution failed: %w", err)
	}

	// Return the raw YAML response
	return response, nil
}

// loadPromptTemplate loads the AI prompt template from the file system.
// It tries the local development path first, then the system container path.
func loadPromptTemplate(repoRoot string) (string, error) {
	const templateRelPath = "templates/ai/init/scan-repository.md"

	localPath := filepath.Join(repoRoot, templateRelPath)

	// Try local path first (development mode: running eac from within the EAC repo)
	if content, err := os.ReadFile(localPath); err == nil {
		return string(content), nil
	}

	// Try system path (container/installed mode: CLIE_CONTAINER_ROOT points to EAC install)
	systemRoot := os.Getenv(environments.EnvCLIEContainerRoot)
	if systemRoot != "" {
		systemPath := filepath.Join(systemRoot, templateRelPath)
		if content, err := os.ReadFile(systemPath); err == nil {
			return string(content), nil
		}
	}

	// Neither path found — emit a diagnostic that names both paths tried
	msg := fmt.Sprintf(
		"AI prompt template not found.\nSearched:\n  1. %s\n", localPath,
	)
	if systemRoot != "" {
		msg += fmt.Sprintf("  2. %s\n", filepath.Join(systemRoot, templateRelPath))
	} else {
		msg += "  2. (skipped: CLIE_CONTAINER_ROOT environment variable not set)\n"
	}
	msg += "\nTo enable AI generation for standalone binary use, set CLIE_CONTAINER_ROOT to the EAC installation directory."
	return "", fmt.Errorf("%s", msg)
}

// prepareScanData prepares scan results and context for the AI prompt.
func prepareScanData(scanResult *ScanResult, repoRoot string) (map[string]interface{}, error) {
	// Serialize scan results to JSON for inclusion in prompt
	scanJSON, err := json.MarshalIndent(scanResult, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize scan results: %w", err)
	}

	// Load repository context (README, directory structure, etc.)
	repoContext := loadRepositoryContext(repoRoot)

	// Prepare custom data structure for template
	customData := map[string]interface{}{
		"Custom": map[string]interface{}{
			"ScanResults":       string(scanJSON),
			"RepositoryContext": repoContext,
			"ExampleConfigs":    getExampleConfigs(),
		},
	}

	return customData, nil
}

// loadRepositoryContext loads contextual information about the repository.
func loadRepositoryContext(repoRoot string) string {
	var context bytes.Buffer

	// Try to read README
	readmeFiles := []string{"README.md", "README.txt", "README"}
	for _, readmeFile := range readmeFiles {
		readmePath := filepath.Join(repoRoot, readmeFile)
		if content, err := os.ReadFile(readmePath); err == nil {
			context.WriteString("## README\n\n")
			// Limit README to first 2000 characters to avoid token limits
			readmeContent := string(content)
			if len(readmeContent) > 2000 {
				readmeContent = readmeContent[:2000] + "...\n(truncated)"
			}
			context.WriteString(readmeContent)
			context.WriteString("\n\n")
			break
		}
	}

	// If no README found, add placeholder
	if context.Len() == 0 {
		context.WriteString("(No README found)\n")
	}

	return context.String()
}

// getExampleConfigs returns example repository configurations for AI reference.
func getExampleConfigs() string {
	return `Example 1 - Simple Go Service:
---
repository:
  name: my-service
  description: REST API service

modules:
  - moniker: api
    description: Backend REST API service
    components:
      - type: go
        root: .

Example 2 - Full-Stack Application:
---
repository:
  name: full-stack-app
  description: Full-stack web application

modules:
  - moniker: backend
    description: Backend API providing REST endpoints
    components:
      - type: python
        root: backend

  - moniker: frontend
    description: Frontend web application built with React
    depends_on:
      - backend
    components:
      - type: typescript
        root: frontend

Example 3 - Microservices:
---
repository:
  name: microservices
  description: Microservices architecture

modules:
  - moniker: shared-lib
    description: Shared utilities and common code
    components:
      - type: go
        root: lib

  - moniker: auth-service
    description: Authentication and authorization service
    depends_on:
      - shared-lib
    components:
      - type: go
        root: services/auth

  - moniker: api-gateway
    description: API gateway routing requests
    depends_on:
      - auth-service
    components:
      - type: go
        root: services/gateway
`
}

// executeTemplate executes the Go template with the provided data.
func executeTemplate(templateContent string, data map[string]interface{}) (string, error) {
	// Parse template
	tmpl, err := template.New("ai-prompt").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template
	var output bytes.Buffer
	if err := tmpl.Execute(&output, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return output.String(), nil
}

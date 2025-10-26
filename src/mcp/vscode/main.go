package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// MCP Server for VSCode integration

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolResult struct {
	Content []Content `json:"content"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Global encoder for stdout - shared across all responses and progress
var stdoutEncoder *json.Encoder

// Global start time for tracking elapsed time in progress messages
var progressStartTime time.Time
var stageStartTime time.Time
var lastStageGlobalTime float64
var lastStageLocalTime float64
var lastStageName string

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	stdoutEncoder = json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(stdoutEncoder, nil, -32700, "Parse error")
			continue
		}

		handleRequest(stdoutEncoder, &req)
	}
}

func handleRequest(encoder *json.Encoder, req *MCPRequest) {
	switch req.Method {
	case "initialize":
		sendResponse(encoder, req.ID, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"serverInfo": map[string]string{
				"name":    "mcp-server-vscode",
				"version": "0.1.0",
			},
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{},
			},
		})

	case "tools/list":
		tools := []Tool{
			{
				Name:        "vscode-action",
				Description: "Execute a VSCode action",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"action": {
							Type:        "string",
							Description: "Action to execute (e.g., 'git-commit', 'git-push', 'git-pull')",
						},
						"message": {
							Type:        "string",
							Description: "Optional message for the action",
						},
					},
					Required: []string{"action"},
				},
			},
			{
				Name:        "execute-agent",
				Description: "Generate a semantic commit message based on git changes",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"agentFile": {
							Type:        "string",
							Description: "Path to the agent file (for reference)",
						},
					},
					Required: []string{"agentFile"},
				},
			},
			{
				Name:        "quick-commit",
				Description: "Generate a quick commit message using only the generator (no review/approval pipeline)",
				InputSchema: InputSchema{
					Type:       "object",
					Properties: map[string]Property{},
				},
			},
		}
		sendResponse(encoder, req.ID, map[string]interface{}{
			"tools": tools,
		})

	case "tools/call":
		var params CallToolParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			sendError(encoder, req.ID, -32602, "Invalid params")
			return
		}

		result := callTool(&params)
		sendResponse(encoder, req.ID, result)

	default:
		sendError(encoder, req.ID, -32601, "Method not found")
	}
}

func callTool(params *CallToolParams) ToolResult {
	switch params.Name {
	case "vscode-action":
		action, ok := params.Arguments["action"].(string)
		if !ok {
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: "Error: action must be a string",
				}},
			}
		}

		message := ""
		if msg, ok := params.Arguments["message"].(string); ok {
			message = msg
		}

		output := executeAction(action, message)
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: output,
			}},
		}

	case "execute-agent":
		agentFile, ok := params.Arguments["agentFile"].(string)
		if !ok {
			// Return error prefix so it can be detected
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: "ERROR: agentFile must be a string",
				}},
			}
		}

		commitMessage, err := generateSemanticCommitMessage(agentFile)
		if err != nil {
			// Return error prefix so it can be detected
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: fmt.Sprintf("ERROR: %v", err),
				}},
			}
		}

		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: commitMessage,
			}},
		}

	case "quick-commit":
		commitMessage, err := generateQuickCommitMessage()
		if err != nil {
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: fmt.Sprintf("ERROR: %v", err),
				}},
			}
		}

		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: commitMessage,
			}},
		}

	default:
		return ToolResult{
			Content: []Content{{
				Type: "text",
				Text: fmt.Sprintf("Unknown tool: %s", params.Name),
			}},
		}
	}
}

func executeAction(action string, message string) string {
	// This is where the VSCode extension will handle actions
	return fmt.Sprintf("Executing action: %s with message: %s", action, message)
}

func generateSemanticCommitMessage(agentFile string) (string, error) {
	// Initialize progress timer
	progressStartTime = time.Now()

	// Get workspace root from agent file path
	// Agent file is at .claude/agents/vscode-extension-commit-button.md
	workspaceRoot := filepath.Dir(filepath.Dir(filepath.Dir(agentFile)))

	// Step 1: Read the generator agent instructions
	sendProgress("init", "Loading generator agent...")
	generatorInstructions, err := ioutil.ReadFile(agentFile)
	if err != nil {
		return "", fmt.Errorf("failed to read generator agent file: %w", err)
	}
	generatorModel := extractModelFromAgent(string(generatorInstructions))

	// Step 2: Gather git context
	sendProgress("git", "Gathering git context...")
	gitContext, err := gatherGitContext(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to gather git context: %w", err)
	}

	// Step 3: Read documentation files
	sendProgress("docs", "Reading documentation...")
	docs, err := readDocumentationFiles(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to read documentation: %w", err)
	}

	// Step 4: Build prompt for generator
	sendProgress("gen-prompt", "Building generator prompt...")
	generatorPrompt := buildClaudePrompt(string(generatorInstructions), gitContext, docs, workspaceRoot)

	// Step 5: Call generator agent
	sendProgress("gen-claude", "Generating initial commit message...")
	initialCommit, err := callClaude(generatorPrompt, generatorModel)
	if err != nil {
		return "", fmt.Errorf("failed to generate initial commit: %w", err)
	}

	// Step 6: Load reviewer agent
	sendProgress("rev-init", "Loading reviewer agent...")
	reviewerPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-message-reviewer.md")
	reviewerInstructions, err := ioutil.ReadFile(reviewerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read reviewer agent file: %w", err)
	}
	reviewerModel := extractModelFromAgent(string(reviewerInstructions))

	// Step 7: Call reviewer agent
	sendProgress("rev-claude", "Reviewing commit message...")
	reviewerPrompt := string(reviewerInstructions) + "\n\n---\n\n## Commit Message to Review\n\n```\n" + initialCommit + "\n```"
	review, err := callClaude(reviewerPrompt, reviewerModel)
	if err != nil {
		return "", fmt.Errorf("failed to review commit: %w", err)
	}

	// Step 8: Load approver agent
	sendProgress("app-init", "Loading approver agent...")
	approverPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-message-approver.md")
	approverInstructions, err := ioutil.ReadFile(approverPath)
	if err != nil {
		return "", fmt.Errorf("failed to read approver agent file: %w", err)
	}
	approverModel := extractModelFromAgent(string(approverInstructions))

	// Step 9: Call approver agent with commit + review
	sendProgress("app-claude", "Final approval...")
	approverPrompt := string(approverInstructions) + "\n\n---\n\n" + initialCommit + "\n\n## Review\n\n" + review
	approvedSection, err := callClaude(approverPrompt, approverModel)
	if err != nil {
		return "", fmt.Errorf("failed to approve commit: %w", err)
	}

	// Step 10: Check if there are concerns
	var finalCommit string
	var concernsResult string

	if strings.Contains(approvedSection, "Approved (with concerns)") {
		// Load concerns handler agent
		sendProgress("concerns-init", "Loading concerns handler...")
		concernsPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-message-concerns-handler.md")
		concernsInstructions, err := ioutil.ReadFile(concernsPath)
		if err != nil {
			return "", fmt.Errorf("failed to read concerns handler agent file: %w", err)
		}
		concernsModel := extractModelFromAgent(string(concernsInstructions))

		// Step 11: Call concerns handler to fix issues
		sendProgress("concerns-claude", "Fixing concerns...")
		commitWithApproval := initialCommit + "\n\n" + approvedSection
		concernsPrompt := string(concernsInstructions) + "\n\n---\n\n" + commitWithApproval
		correctedCommit, err := callClaude(concernsPrompt, concernsModel)
		if err != nil {
			return "", fmt.Errorf("failed to fix concerns: %w", err)
		}
		concernsResult = correctedCommit
	}

	// Step 12: Determine base commit (with or without concerns fixed)
	var baseCommit string
	if concernsResult != "" {
		baseCommit = concernsResult
	} else {
		baseCommit = initialCommit
	}
	baseCommit = strings.TrimSpace(baseCommit)

	// Step 13: Generate commit title (5th agent)
	sendProgress("title-init", "Loading title generator agent...")
	titleGenPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-title-generator.md")
	titleGenInstructions, err := ioutil.ReadFile(titleGenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read title generator agent file: %w", err)
	}
	titleGenModel := extractModelFromAgent(string(titleGenInstructions))

	sendProgress("title-claude", "Generating commit title...")
	titleGenPrompt := string(titleGenInstructions) + "\n\n---\n\n" + baseCommit
	commitTitle, err := callClaude(titleGenPrompt, titleGenModel)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit title: %w", err)
	}
	// Clean the title (remove any extra whitespace/newlines)
	commitTitle = strings.TrimSpace(commitTitle)

	// Step 14: Stitch together final commit message from all agents
	sendProgress("stitch", "Stitching agent outputs...")

	// Build the composite commit message
	var composite strings.Builder

	// 1. Add commit title as top-level heading (MD041)
	composite.WriteString("# ")
	composite.WriteString(commitTitle)
	composite.WriteString("\n\n")

	// 2. Add the base commit message (remove any heading if present)
	cleanCommit := baseCommit
	if strings.HasPrefix(baseCommit, "# ") {
		// Skip the first heading line
		lines := strings.SplitN(baseCommit, "\n", 2)
		if len(lines) > 1 {
			cleanCommit = strings.TrimSpace(lines[1])
		} else {
			cleanCommit = ""
		}
	}
	composite.WriteString(cleanCommit)
	composite.WriteString("\n\n")

	// 3. Add simplified agent status (single factual statement)
	composite.WriteString("Agent: ")

	// Determine final status
	// If concerns were handled, it's just "Approved" (concerns were fixed)
	// If concerns exist but not handled, show "Approved (with concerns)"
	// Otherwise, just "Approved"
	if strings.Contains(approvedSection, "Approved (with concerns)") && concernsResult == "" {
		composite.WriteString("Approved (with concerns)")
	} else {
		composite.WriteString("Approved")
	}
	composite.WriteString("\n")

	finalCommit = composite.String()

	// Step 15: Validate the final commit message
	sendProgress("validate", "Validating commit message structure...")
	validationErrors := validateCommitMessage(finalCommit, workspaceRoot)
	if len(validationErrors) > 0 {
		// Auto-fix validation errors with fixer agent
		sendProgress("fixer-init", "Loading commit message fixer...")
		fixerPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-message-fixer.md")
		fixerInstructions, err := ioutil.ReadFile(fixerPath)
		if err != nil {
			// If fixer agent doesn't exist, return validation errors
			var errorMessages []string
			errorMessages = append(errorMessages, "❌ Commit message validation failed:")
			for _, verr := range validationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("  • %s", verr.Message))
			}
			return "", fmt.Errorf("%s", strings.Join(errorMessages, "\n"))
		}
		fixerModel := extractModelFromAgent(string(fixerInstructions))

		sendProgress("fixer-claude", "Auto-fixing validation errors...")

		// Build fixer prompt with original message and errors
		var fixerPrompt strings.Builder
		fixerPrompt.WriteString(string(fixerInstructions))
		fixerPrompt.WriteString("\n\n---\n\n## Original Commit Message\n\n```\n")
		fixerPrompt.WriteString(finalCommit)
		fixerPrompt.WriteString("\n```\n\n## Validation Errors\n\n")
		for _, verr := range validationErrors {
			fixerPrompt.WriteString(fmt.Sprintf("• %s\n", verr.Message))
		}

		fixedCommit, err := callClaude(fixerPrompt.String(), fixerModel)
		if err != nil {
			return "", fmt.Errorf("failed to auto-fix commit message: %w", err)
		}

		// Re-validate the fixed commit
		sendProgress("revalidate", "Re-validating fixed commit...")
		revalidationErrors := validateCommitMessage(fixedCommit, workspaceRoot)
		if len(revalidationErrors) > 0 {
			// Fixer didn't work, return both original and new errors
			var errorMessages []string
			errorMessages = append(errorMessages, "❌ Commit message validation failed even after auto-fix:")
			for _, verr := range revalidationErrors {
				errorMessages = append(errorMessages, fmt.Sprintf("  • %s", verr.Message))
			}
			return "", fmt.Errorf("%s", strings.Join(errorMessages, "\n"))
		}

		// Fixed successfully!
		finalCommit = fixedCommit
	}

	sendProgress("complete", "Complete!")
	return finalCommit, nil
}

// generateQuickCommitMessage generates a simple commit message using only the generator agent
func generateQuickCommitMessage() (string, error) {
	workspaceRoot := os.Getenv("WORKSPACE_ROOT")
	if workspaceRoot == "" {
		return "", fmt.Errorf("WORKSPACE_ROOT environment variable not set")
	}

	// Step 1: Gather git context
	sendProgress("git-context", "Gathering git context...")
	gitContext, err := gatherGitContext(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to gather git context: %w", err)
	}

	// Step 2: Read documentation files
	sendProgress("docs", "Reading documentation...")
	docs, err := readDocumentationFiles(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to read documentation: %w", err)
	}

	// Step 3: Load generator agent
	sendProgress("generator", "Loading generator agent...")
	generatorPath := filepath.Join(workspaceRoot, ".claude", "agents", "vscode-extension-commit-button.md")
	generatorInstructions, err := ioutil.ReadFile(generatorPath)
	if err != nil {
		return "", fmt.Errorf("failed to read generator agent file: %w", err)
	}
	generatorModel := extractModelFromAgent(string(generatorInstructions))

	// Step 4: Build prompt and generate
	sendProgress("generating", "Generating quick commit...")
	generatorPrompt := buildClaudePrompt(string(generatorInstructions), gitContext, docs, workspaceRoot)
	commitMessage, err := callClaude(generatorPrompt, generatorModel)
	if err != nil {
		return "", fmt.Errorf("failed to generate commit: %w", err)
	}

	sendProgress("complete", "Complete!")
	return commitMessage, nil
}

type GitContext struct {
	Status      string       // git status --porcelain output
	Diff        string       // git diff --staged output
	HeadSHA     string       // current commit SHA
	FileChanges []FileChange // parsed from Status
}

// ModuleContract represents a deployable unit contract
type ModuleContract struct {
	Moniker     string       `yaml:"moniker"`
	Name        string       `yaml:"name"`
	Type        string       `yaml:"type"`
	Description string       `yaml:"description"`
	Root        string       `yaml:"root"`
	Source      SourceConfig `yaml:"source"`
}

// SourceConfig represents the source configuration in a contract
type SourceConfig struct {
	Includes []string `yaml:"includes"`
}

// CommitValidationError represents a validation error
type CommitValidationError struct {
	Field   string
	Message string
}

func (e CommitValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Field, e.Message)
}

// loadModuleContracts loads all module contracts from contracts/deployable-units/0.1.0/*.yml
func loadModuleContracts(workspaceRoot string) (map[string]ModuleContract, error) {
	contractsDir := filepath.Join(workspaceRoot, "contracts", "deployable-units", "0.1.0")
	contracts := make(map[string]ModuleContract)

	files, err := filepath.Glob(filepath.Join(contractsDir, "*.yml"))
	if err != nil {
		return nil, fmt.Errorf("failed to glob contract files: %w", err)
	}

	for _, file := range files {
		// Skip definitions.yml as it's not a module contract
		if filepath.Base(file) == "definitions.yml" {
			continue
		}

		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read contract file %s: %w", file, err)
		}

		var contract ModuleContract
		if err := yaml.Unmarshal(data, &contract); err != nil {
			return nil, fmt.Errorf("failed to parse contract file %s: %w", file, err)
		}

		if contract.Moniker != "" {
			contracts[contract.Moniker] = contract
		}
	}

	return contracts, nil
}

// validateCommitMessage validates the structure and content of a commit message
func validateCommitMessage(commitMessage string, workspaceRoot string) []CommitValidationError {
	var errors []CommitValidationError

	// Load module contracts
	moduleContracts, err := loadModuleContracts(workspaceRoot)
	if err != nil {
		errors = append(errors, CommitValidationError{
			Field:   "contracts",
			Message: fmt.Sprintf("Failed to load module contracts: %v", err),
		})
		// Continue with validation even if contracts failed to load
	}

	lines := strings.Split(commitMessage, "\n")
	if len(lines) == 0 {
		errors = append(errors, CommitValidationError{
			Field:   "structure",
			Message: "Commit message is empty",
		})
		return errors
	}

	// 1. Validate unique top-level heading (MD041)
	var topLevelHeadings []string
	boldColonRegex := regexp.MustCompile(`^\*\*[^*]+:\*\*`)
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			topLevelHeadings = append(topLevelHeadings, fmt.Sprintf("Line %d: %s", i+1, line))
		}

		// Check for **Bold:** pattern (forbidden)
		trimmed := strings.TrimSpace(line)
		if boldColonRegex.MatchString(trimmed) {
			errors = append(errors, CommitValidationError{
				Field:   "bold-colon",
				Message: fmt.Sprintf("Line %d: Forbidden **Bold:** pattern detected. Use proper ### headers instead: %s", i+1, trimmed),
			})
		}
	}

	if len(topLevelHeadings) == 0 {
		errors = append(errors, CommitValidationError{
			Field:   "heading",
			Message: "Missing top-level heading (# title)",
		})
	} else if len(topLevelHeadings) > 1 {
		errors = append(errors, CommitValidationError{
			Field:   "heading",
			Message: fmt.Sprintf("Multiple top-level headings found (must have exactly 1):\n%s", strings.Join(topLevelHeadings, "\n")),
		})
	}

	// 2. Validate module sections have semantic format: <module>: <type>: <description>
	validTypes := map[string]bool{
		"feat":     true,
		"fix":      true,
		"refactor": true,
		"docs":     true,
		"chore":    true,
		"test":     true,
		"perf":     true,
		"style":    true,
	}

	// Regex to match module section subject lines
	// Format: <module-name>: <semantic-type>: <description>
	semanticFormatRegex := regexp.MustCompile(`^([a-z0-9\-]+):\s*(feat|fix|refactor|docs|chore|test|perf|style):\s*(.+)$`)

	// Track sections for 72-char validation
	inSummarySection := false
	inModuleBodySection := false

	// Find all module sections (lines that look like module subject lines)
	inModuleSection := false
	currentModule := ""
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track Summary section for line length validation
		if strings.HasPrefix(trimmed, "## Summary") {
			inSummarySection = true
			inModuleBodySection = false
			continue
		}

		// End Summary section when we hit Files affected
		if strings.HasPrefix(trimmed, "## Files affected") {
			inSummarySection = false
			inModuleBodySection = false
			continue
		}

		// Check if this is a module section header
		if strings.HasPrefix(trimmed, "## ") && !strings.HasPrefix(trimmed, "## Summary") && !strings.HasPrefix(trimmed, "## Files affected") {
			inModuleSection = true
			inModuleBodySection = false
			currentModule = strings.TrimPrefix(trimmed, "## ")

			// Validate blank line after header (MD022)
			if i+1 < len(lines) && strings.TrimSpace(lines[i+1]) != "" {
				errors = append(errors, CommitValidationError{
					Field:   "header-spacing",
					Message: fmt.Sprintf("Line %d: Missing blank line after header '## %s' (MD022)", i+1, currentModule),
				})
			}
			continue
		}

		// If we're in a module section, the next non-empty line should be the semantic subject line
		if inModuleSection && trimmed != "" && !strings.HasPrefix(trimmed, "```") && !strings.HasPrefix(trimmed, "|") {
			matches := semanticFormatRegex.FindStringSubmatch(trimmed)
			if matches == nil {
				errors = append(errors, CommitValidationError{
					Field:   "semantic-format",
					Message: fmt.Sprintf("Line %d: Module '%s' subject line does not follow semantic format '<module>: <type>: <description>'\nFound: %s\nExpected format: %s: <feat|fix|refactor|docs|chore|test|perf|style>: <description>", i+1, currentModule, trimmed, currentModule),
				})
			} else {
				moduleName := matches[1]
				semanticType := matches[2]
				description := matches[3]

				// Validate module name matches section header
				if moduleName != currentModule {
					errors = append(errors, CommitValidationError{
						Field:   "module-mismatch",
						Message: fmt.Sprintf("Line %d: Module name in subject line '%s' does not match section header '%s'", i+1, moduleName, currentModule),
					})
				}

				// Validate semantic type
				if !validTypes[semanticType] {
					errors = append(errors, CommitValidationError{
						Field:   "semantic-type",
						Message: fmt.Sprintf("Line %d: Invalid semantic type '%s'. Must be one of: feat, fix, refactor, docs, chore, test, perf, style", i+1, semanticType),
					})
				}

				// Validate description is not empty
				if strings.TrimSpace(description) == "" {
					errors = append(errors, CommitValidationError{
						Field:   "description",
						Message: fmt.Sprintf("Line %d: Description cannot be empty", i+1),
					})
				}

				// Validate module exists in contracts (if contracts loaded successfully)
				if len(moduleContracts) > 0 {
					if _, exists := moduleContracts[moduleName]; !exists {
						errors = append(errors, CommitValidationError{
							Field:   "module-unknown",
							Message: fmt.Sprintf("Line %d: Module '%s' not found in contracts/deployable-units/0.1.0/", i+1, moduleName),
						})
					}
				}

				// Subject line length (≤72 characters - GitHub hard limit)
				// 50 chars is the soft limit for readability, 72 is the hard truncation limit
				if len(trimmed) > 72 {
					errors = append(errors, CommitValidationError{
						Field:   "subject-length",
						Message: fmt.Sprintf("Line %d: Subject line exceeds 72 characters (%d chars): %s", i+1, len(trimmed), trimmed),
					})
				}
			}

			inModuleSection = false
			inModuleBodySection = true // Now we're in the body text after subject line
			currentModule = ""
		}

		// Validate 72-character line limit for Summary and module body text
		// Skip empty lines, headers, horizontal rules, table lines, and yaml blocks
		if (inSummarySection || inModuleBodySection) &&
		   trimmed != "" &&
		   !strings.HasPrefix(trimmed, "#") &&
		   !strings.HasPrefix(trimmed, "---") &&
		   !strings.HasPrefix(trimmed, "|") &&
		   !strings.HasPrefix(trimmed, "```") {
			if len(trimmed) > 72 {
				sectionName := "module body"
				if inSummarySection {
					sectionName = "Summary"
				}
				errors = append(errors, CommitValidationError{
					Field:   "line-length",
					Message: fmt.Sprintf("Line %d: %s text exceeds 72 characters (%d chars): %s", i+1, sectionName, len(trimmed), trimmed),
				})
			}
		}

		// Exit module body section when we hit yaml block or next module section
		if inModuleBodySection && strings.HasPrefix(trimmed, "```yaml") {
			inModuleBodySection = false
		}
	}

	// Validate all yaml blocks are properly closed
	yamlBlockOpen := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```yaml") {
			yamlBlockOpen = true
		} else if yamlBlockOpen && strings.HasPrefix(trimmed, "```") {
			yamlBlockOpen = false
		}
	}
	if yamlBlockOpen {
		errors = append(errors, CommitValidationError{
			Field:   "yaml-block",
			Message: "Unclosed yaml code block - missing closing ```",
		})
	}

	return errors
}

func gatherGitContext(workspaceRoot string) (*GitContext, error) {
	ctx := &GitContext{}

	// Get current HEAD SHA
	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = workspaceRoot
	headOutput, err := headCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git rev-parse HEAD failed: %w", err)
	}
	ctx.HeadSHA = strings.TrimSpace(string(headOutput))

	// Get git status
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = workspaceRoot
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}
	ctx.Status = string(statusOutput)

	// Get git diff for STAGED changes only
	diffCmd := exec.Command("git", "diff", "--staged")
	diffCmd.Dir = workspaceRoot
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	ctx.Diff = string(diffOutput)

	// Parse file changes from status (no need for git log - status shows everything)
	ctx.FileChanges = parseFileChanges(ctx.Status)

	return ctx, nil
}

func readDocumentationFiles(workspaceRoot string) (map[string]string, error) {
	docs := make(map[string]string)

	// List of documentation files from the agent instructions
	docPatterns := []string{
		"docs/reference/trunk/revisionable-timeline.md",
		"docs/reference/trunk/repository-layout.md",
		"docs/reference/trunk/versioning.md",
		"docs/reference/trunk/semantic-commits.md",
		"contracts/deployable-units/0.1.0/*.yml",
	}

	for _, docPattern := range docPatterns {
		fullPattern := filepath.Join(workspaceRoot, docPattern)

		// Check if pattern contains wildcards
		if strings.Contains(docPattern, "*") {
			// Expand glob pattern
			matches, err := filepath.Glob(fullPattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Glob pattern error for %s: %v\n", docPattern, err)
				continue
			}

			// Read each matched file
			for _, fullPath := range matches {
				content, err := ioutil.ReadFile(fullPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Could not read %s: %v\n", fullPath, err)
					continue
				}
				// Use relative path as key
				relPath, _ := filepath.Rel(workspaceRoot, fullPath)
				docs[relPath] = string(content)
			}
		} else {
			// Read single file directly
			content, err := ioutil.ReadFile(fullPattern)
			if err != nil {
				// Log but don't fail - some docs might be optional
				fmt.Fprintf(os.Stderr, "Warning: Could not read %s: %v\n", docPattern, err)
				continue
			}
			docs[docPattern] = string(content)
		}
	}

	return docs, nil
}

func buildClaudePrompt(agentInstructions string, gitCtx *GitContext, docs map[string]string, workspaceRoot string) string {
	var prompt bytes.Buffer

	// Start with the agent instructions (the main prompt)
	prompt.WriteString(agentInstructions)
	prompt.WriteString("\n\n---\n\n")

	prompt.WriteString("# PRE-FETCHED GIT DATA (DO NOT RUN GIT COMMANDS)\n\n")

	// Add simple context
	prompt.WriteString(fmt.Sprintf("**Current HEAD SHA:** %s\n\n", gitCtx.HeadSHA))

	// Add file changes table with proper column spacing
	prompt.WriteString("## File Changes (Normalized - Already Parsed)\n\n")
	prompt.WriteString("The following table shows ALL changed files with normalized status and auto-detected modules:\n\n")
	prompt.WriteString(formatFileTable(gitCtx.FileChanges))
	prompt.WriteString("\n")

	prompt.WriteString("## Git Status (Raw Porcelain Output)\n\n")
	prompt.WriteString("```\n")
	prompt.WriteString(gitCtx.Status)
	prompt.WriteString("```\n\n")

	prompt.WriteString("## Git Diff (Complete Changes)\n\n")
	prompt.WriteString("```diff\n")
	prompt.WriteString(gitCtx.Diff)
	prompt.WriteString("```\n\n")

	// Add documentation content
	prompt.WriteString("---\n\n")
	prompt.WriteString("# PRE-FETCHED DOCUMENTATION (DO NOT READ FILES)\n\n")
	prompt.WriteString("The following documentation files have been pre-loaded for you:\n\n")

	for path, content := range docs {
		prompt.WriteString(fmt.Sprintf("## Content of `%s`\n\n", path))
		prompt.WriteString("```markdown\n")
		prompt.WriteString(content)
		prompt.WriteString("\n```\n\n")
	}

	// Add module metadata with glob patterns
	prompt.WriteString("\n---\n\n")
	prompt.WriteString("# MODULE METADATA (GitHub Actions Path Filters)\n\n")
	prompt.WriteString("The following modules have been detected in this commit. Each module section\n")
	prompt.WriteString("MUST include its GitHub Actions compatible glob pattern(s).\n\n")

	// Load module contracts for glob patterns
	moduleContracts, err := loadModuleContracts(workspaceRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load module contracts: %v\n", err)
		moduleContracts = make(map[string]ModuleContract) // Empty map for fallback
	}

	// Group files by module
	moduleMap := make(map[string]bool)
	for _, change := range gitCtx.FileChanges {
		moduleMap[change.Module] = true
	}

	// Sort modules for consistent output
	var modules []string
	for module := range moduleMap {
		modules = append(modules, module)
	}
	// Simple sort
	for i := 0; i < len(modules); i++ {
		for j := i + 1; j < len(modules); j++ {
			if modules[i] > modules[j] {
				modules[i], modules[j] = modules[j], modules[i]
			}
		}
	}

	// List each module with its glob patterns
	for _, module := range modules {
		globs := getModuleGlobPattern(module, moduleContracts)
		prompt.WriteString(fmt.Sprintf("**%s:**\n", module))
		prompt.WriteString("```yaml\n")
		prompt.WriteString("paths:\n")
		for _, glob := range globs {
			prompt.WriteString(fmt.Sprintf("  - '%s'\n", glob))
		}
		prompt.WriteString("```\n\n")
	}

	// Add explicit instruction for output format
	prompt.WriteString("\n---\n\n")
	prompt.WriteString("CRITICAL INSTRUCTIONS - OUTPUT FORMAT:\n\n")
	prompt.WriteString("Required Structure:\n\n")
	prompt.WriteString(fmt.Sprintf("# Revision %s\n\n", gitCtx.HeadSHA))
	prompt.WriteString("SUMMARY (2-4 sentences):\n")
	prompt.WriteString("Write a human-readable executive summary explaining:\n")
	prompt.WriteString("- WHAT was accomplished in this commit\n")
	prompt.WriteString("- WHY it matters for the system\n")
	prompt.WriteString("- Downstream/production IMPACT\n")
	prompt.WriteString("Be generous with detail - this helps stakeholders understand significance.\n\n")
	prompt.WriteString("FILE TABLE:\n")
	prompt.WriteString("Copy the file table EXACTLY as shown above with proper spacing:\n")
	prompt.WriteString("| Status   | File                        | Module     |\n")
	prompt.WriteString("| -------- | --------------------------- | ---------- |\n")
	prompt.WriteString("| added    | path/to/file.go             | module     |\n")
	prompt.WriteString("| modified | path/to/another/file.ts     | module     |\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("MODULE SECTION STRUCTURE:\n\n")
	prompt.WriteString("## <module-name>\n\n")
	prompt.WriteString("<module>: <type>: <description> (≤50 chars)\n\n")
	prompt.WriteString("[Body text explaining WHY - wrap at 72 chars]\n\n")
	prompt.WriteString("```yaml\n")
	prompt.WriteString("paths:\n")
	prompt.WriteString("  - 'pattern/from/metadata/**'\n")
	prompt.WriteString("```\n\n")
	prompt.WriteString("---\n\n")
	prompt.WriteString("[Repeat for each affected module]\n\n")
	prompt.WriteString("Example module section:\n\n")
	prompt.WriteString("## src-mcp-vscode\n\n")
	prompt.WriteString("src-mcp-vscode: feat: add commit generation\n\n")
	prompt.WriteString("Implements tool for generating structured\n")
	prompt.WriteString("commit messages from git context.\n\n")
	prompt.WriteString("```yaml\n")
	prompt.WriteString("paths:\n")
	prompt.WriteString("  - 'src/mcp/vscode/**'\n")
	prompt.WriteString("```\n\n")
	prompt.WriteString("IMPORTANT RULES:\n")
	prompt.WriteString("1. Copy the file table from 'File Changes (Normalized)' section EXACTLY - preserve all spaces\n")
	prompt.WriteString("2. Do NOT reformat or adjust the table spacing - copy character-for-character\n")
	prompt.WriteString("3. Table appears ONCE at the top only\n")
	prompt.WriteString("4. NO file lists in module sections - table shows everything\n")
	prompt.WriteString("5. Recent commits are for context only - not shown in output\n")
	prompt.WriteString(fmt.Sprintf("6. Revision header MUST be: # Revision %s\n", gitCtx.HeadSHA))
	prompt.WriteString("7. CRITICAL: Each module section MUST start with ## <module-name> header\n")
	prompt.WriteString("8. CRITICAL: After body text, include ```yaml paths: block with glob patterns (NO heading)\n")
	prompt.WriteString("9. Copy the glob pattern EXACTLY as shown in the MODULE METADATA section for that module\n\n")
	prompt.WriteString("50/72 RULE - CRITICAL:\n")
	prompt.WriteString("10. Each module subject line: <module>: <type>: <description> MUST be ≤ 50 characters\n")
	prompt.WriteString("11. Body text lines MUST be ≤ 72 characters - wrap longer lines\n")
	prompt.WriteString("12. Blank line between subject and body\n")
	prompt.WriteString("13. Be concise - focus on WHY, not WHAT (diffs show what)\n\n")
	prompt.WriteString("MARKDOWN LINT COMPLIANCE:\n")
	prompt.WriteString("14. CRITICAL: File MUST end with a newline character (MD047 compliance)\n")
	prompt.WriteString("15. Ensure there is exactly one blank line at the end of the commit message\n\n")
	prompt.WriteString("RETURN ONLY THE COMMIT MESSAGE. NO ANALYSIS. NO MARKDOWN CODE BLOCKS.\n")

	return prompt.String()
}

func callClaude(prompt string, model string) (string, error) {
	// Use Claude Code CLI instead of direct API calls
	// This leverages the user's existing Claude Code subscription

	// Build command with optional model flag
	// IMPORTANT: Use empty --setting-sources to bypass hooks and CLAUDE.md
	// AND disable co-author attribution via settings JSON
	// NOTE: We don't disable prompt caching as it significantly improves performance
	// and Claude Code's API-level prompt caching is smart enough to invalidate
	// when the actual git diff changes
	args := []string{
		"--print",
		"--setting-sources", "",
		"--settings", `{"includeCoAuthoredBy":false}`,
	}

	// ENFORCE: Model must be provided (from agent frontmatter)
	if model == "" {
		return "", fmt.Errorf("model not specified in agent frontmatter - all agents must define 'model:' field")
	}
	args = append(args, "--model", model)

	cmd := exec.Command("claude", args...)

	// Pipe the prompt via stdin
	cmd.Stdin = strings.NewReader(prompt)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("claude CLI failed: %w\nStderr: %s", err, stderr.String())
	}

	// The output should be the structured commit message
	commitMessage := strings.TrimSpace(stdout.String())

	// Clean up any markdown code blocks if present
	commitMessage = strings.TrimPrefix(commitMessage, "```")
	commitMessage = strings.TrimSuffix(commitMessage, "```")
	commitMessage = strings.TrimSpace(commitMessage)

	// Remove Claude Code footer if present
	claudeFooter := "🤖 Generated with [Claude Code](https://claude.com/claude-code)\n\nCo-Authored-By: Claude <noreply@anthropic.com>"
	commitMessage = strings.Replace(commitMessage, claudeFooter, "", -1)
	// Also try without leading newlines
	commitMessage = strings.Replace(commitMessage, "\n\n"+claudeFooter, "", -1)
	commitMessage = strings.Replace(commitMessage, "\n"+claudeFooter, "", -1)
	commitMessage = strings.TrimSpace(commitMessage)

	// Fix bold headers to proper markdown headers
	// Convert **Agent Pipeline Architecture** to ### Agent Pipeline Architecture
	lines := strings.Split(commitMessage, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check if line is bold text (starts and ends with **)
		if strings.HasPrefix(trimmed, "**") && strings.HasSuffix(trimmed, "**") && len(trimmed) > 4 {
			// Remove ** and convert to ### header
			headerText := strings.TrimPrefix(trimmed, "**")
			headerText = strings.TrimSuffix(headerText, "**")
			lines[i] = "### " + headerText
		}
	}
	commitMessage = strings.Join(lines, "\n")

	// Ensure commit message ends with exactly one newline (MD047 compliance)
	commitMessage = strings.TrimRight(commitMessage, "\n")
	commitMessage = commitMessage + "\n"

	return commitMessage, nil
}

// FileChange represents a normalized file change
type FileChange struct {
	Status   string // normalized: added, modified, deleted, renamed
	FilePath string
	Module   string
}

// normalizeGitStatus converts git XY status codes to simple categories
func normalizeGitStatus(statusCode string) string {
	if len(statusCode) < 2 {
		return "modified"
	}

	// Get first two characters (XY format)
	x := statusCode[0]
	y := statusCode[1]

	// Check for specific statuses
	switch {
	case x == 'A' || statusCode == "??":
		return "added"
	case x == 'D' || y == 'D':
		return "deleted"
	case x == 'R' || y == 'R':
		return "renamed"
	case x == 'M' || y == 'M', x == ' ' && y == 'M':
		return "modified"
	default:
		return "modified"
	}
}

// determineFileModule intelligently extracts module name from file path structure
// It reasons the module from folder nesting and naming conventions rather than hardcoding mappings
func determineFileModule(filePath string) string {
	// Normalize path separators for consistent matching
	filePath = strings.ReplaceAll(filePath, "\\", "/")

	// Extract base filename for special checks
	fileName := filepath.Base(filePath)

	// Special case: .gitkeep files follow their parent directory's module
	if fileName == ".gitkeep" {
		// Remove the filename and re-evaluate with parent directory
		parentPath := filepath.Dir(filePath)
		// Normalize for consistency
		parentPath = strings.ReplaceAll(parentPath, "\\", "/")

		if parentPath != "." && parentPath != "/" {
			// Add a dummy file to the parent path so patterns match correctly
			// e.g., "src/cli" becomes "src/cli/dummy.file"
			return determineFileModule(parentPath + "/dummy.file")
		}
		return "repo-config"
	}

	// Special case: README.md and documentation files in src/ directories
	// These are docs, not part of the code module
	if fileName == "README.md" || fileName == "CONTRIBUTING.md" {
		// If in src/mcp/ or src/cli/, it's documentation, not the module itself
		if strings.HasPrefix(filePath, "src/") {
			return "docs"
		}
		// Root README files are also docs
		if !strings.Contains(filePath, "/") {
			return "docs"
		}
	}

	// Special case: ALL definitions.yml files anywhere in repo → repo-config
	// These are repository-level metadata files regardless of location
	// Examples: /definitions.yml, contracts/deployable-units/0.1.0/definitions.yml, etc.
	if fileName == "definitions.yml" {
		return "repo-config"
	}

	// Pattern 1: automation/<module-name>/... → extract module name
	// Examples: automation/sh-vscode/ → "sh-vscode", automation/pwsh-build/ → "pwsh-build"
	if strings.HasPrefix(filePath, "automation/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 2 && parts[1] != "" {
			return parts[1] // Module type detected from prefix (sh-, pwsh-, etc.)
		}
		return "automation" // fallback for files directly in automation/
	}

	// Pattern 2: containers/<module-name>/... → extract module name
	// Examples: containers/mkdocs/ → "mkdocs", containers/nginx-proxy/ → "nginx-proxy"
	if strings.HasPrefix(filePath, "containers/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 2 && parts[1] != "" {
			return parts[1]
		}
		return "containers" // fallback
	}

	// Pattern 3: src/mcp/<service>/... → src-mcp-<service>
	// Examples: src/mcp/pwsh/ → "src-mcp-pwsh", src/mcp/vscode/ → "src-mcp-vscode"
	// Note: README.md files are handled above and won't reach here
	if strings.HasPrefix(filePath, "src/mcp/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[2] != "" {
			return "src-mcp-" + parts[2]
		}
		return "src-mcp" // fallback
	}

	// Pattern 3b: src/cli/... → cli module
	if strings.HasPrefix(filePath, "src/cli/") {
		return "cli"
	}

	// Pattern 4: .vscode/extensions/<name>/... → use exact extension moniker
	// Examples: .vscode/extensions/claude-mcp-vscode/ → "claude-mcp-vscode"
	if strings.HasPrefix(filePath, ".vscode/extensions/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[2] != "" {
			// Return the exact extension folder name (matches contract moniker)
			return parts[2]
		}
		return "claude-mcp-vscode" // fallback
	}

	// Pattern 5: contracts/deployable-units/<version>/*.yml → actual module from yml filename
	// Examples: contracts/deployable-units/0.1.0/docs.yml → "docs"
	//           contracts/deployable-units/0.1.0/src-mcp-vscode.yml → "src-mcp-vscode"
	if strings.HasPrefix(filePath, "contracts/deployable-units/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 4 && strings.HasSuffix(parts[3], ".yml") {
			// Extract module name from filename (e.g., "docs.yml" → "docs")
			moduleName := strings.TrimSuffix(parts[3], ".yml")
			if moduleName != "definitions" {
				return moduleName
			}
		}
		return "contracts-deployable-units" // fallback for definitions.yml or unknown
	}

	// Pattern 5b: other contracts/<name>/... → contracts-<name>
	if strings.HasPrefix(filePath, "contracts/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 2 && parts[1] != "" {
			return "contracts-" + parts[1]
		}
		return "contracts" // fallback
	}

	// Pattern 6: docs/... → docs
	if strings.HasPrefix(filePath, "docs/") {
		return "docs"
	}

	// Pattern 7: .claude/... → claude-config
	if strings.HasPrefix(filePath, ".claude/") {
		return "claude-config"
	}

	// Pattern 8: .vscode/ (non-extension files) → vscode-config
	if strings.HasPrefix(filePath, ".vscode/") {
		return "vscode-config"
	}

	// Pattern 9: Root markdown files → docs
	if !strings.Contains(filePath, "/") && strings.HasSuffix(filePath, ".md") {
		return "docs"
	}

	// Pattern 10: Root config files → repo-config
	if !strings.Contains(filePath, "/") {
		if strings.HasPrefix(filePath, ".") ||
			filePath == "package.json" ||
			filePath == "mkdocs.yml" ||
			filePath == "LICENSE" ||
			strings.HasSuffix(filePath, ".json") ||
			strings.HasSuffix(filePath, ".yml") ||
			strings.HasSuffix(filePath, ".yaml") ||
			strings.HasSuffix(filePath, ".lock") {
			return "repo-config"
		}
	}

	// Default: unknown (for truly unrecognized paths)
	return "unknown"
}

// getModuleGlobPattern returns the GitHub Actions compatible glob pattern for a module
// These patterns can be used in GitHub Actions workflow path filters
// If a contract exists for the module, it uses the contract's source.includes patterns
// Otherwise, it falls back to hardcoded patterns for known modules
func getModuleGlobPattern(module string, contracts map[string]ModuleContract) []string {
	// First, check if we have a contract for this module
	if contract, exists := contracts[module]; exists {
		// Use patterns from contract
		var patterns []string
		for _, include := range contract.Source.Includes {
			// If pattern starts with root, it's already absolute
			// Otherwise, combine root + pattern
			if strings.HasPrefix(include, contract.Root) {
				patterns = append(patterns, include)
			} else if contract.Root != "" {
				// Combine root with pattern, handling ** glob patterns
				patterns = append(patterns, filepath.Join(contract.Root, include))
			} else {
				patterns = append(patterns, include)
			}
		}
		// Normalize path separators for GitHub Actions (always use forward slash)
		for i, p := range patterns {
			patterns[i] = strings.ReplaceAll(p, "\\", "/")
		}
		return patterns
	}

	// Fallback to hardcoded patterns for modules without contracts
	// Normalize path separators for cross-platform compatibility
	// GitHub Actions uses forward slashes even on Windows

	// Handle MCP servers
	if strings.HasPrefix(module, "mcp-") {
		service := strings.TrimPrefix(module, "mcp-")
		return []string{fmt.Sprintf("src/mcp/%s/**", service)}
	}

	// Handle automation modules (sh-, pwsh-, py-, etc.)
	if strings.HasPrefix(module, "sh-") || strings.HasPrefix(module, "pwsh-") || strings.HasPrefix(module, "py-") {
		return []string{fmt.Sprintf("automation/%s/**", module)}
	}

	// Handle contract modules
	if strings.HasPrefix(module, "contracts-") {
		contractType := strings.TrimPrefix(module, "contracts-")
		return []string{fmt.Sprintf("contracts/%s/**", contractType)}
	}

	// Handle specific known modules
	switch module {
	case "vscode-ext-claude-commit":
		return []string{".vscode/extensions/claude-mcp-vscode/**"}

	case "cli":
		return []string{"src/cli/**"}

	case "docs":
		// Docs can be in multiple places
		return []string{"docs/**", "*.md"}

	case "claude-config":
		return []string{".claude/**"}

	case "vscode-config":
		// VSCode config but NOT extensions
		return []string{
			".vscode/*.json",
			".vscode/*.md",
			".vscode/settings.*.json",
		}

	case "repo-config":
		// Root level config files - list common patterns
		return []string{
			"*.json",
			"*.yml",
			"*.yaml",
			".gitignore",
			".gitattributes",
			"LICENSE",
			"*.lock",
		}

	// Container modules (without prefix)
	case "mkdocs":
		return []string{"containers/mkdocs/**"}

	case "nginx-proxy":
		return []string{"containers/nginx-proxy/**"}

	case "postgres":
		return []string{"containers/postgres/**"}

	default:
		// For container modules without explicit cases
		if !strings.Contains(module, "-") && module != "unknown" {
			// Likely a container module, try that first
			return []string{fmt.Sprintf("containers/%s/**", module)}
		}

		// Unknown module - return a safe default
		return []string{fmt.Sprintf("**/%s/**", module)}
	}
}

// parseFileChanges parses git status and returns normalized file changes
func parseFileChanges(status string) []FileChange {
	var changes []FileChange
	lines := strings.Split(status, "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse git status format: "XY filename"
		if len(line) > 3 {
			statusCode := line[0:2]
			filename := strings.TrimSpace(line[3:])

			changes = append(changes, FileChange{
				Status:   normalizeGitStatus(statusCode),
				FilePath: filename,
				Module:   determineFileModule(filename),
			})
		}
	}

	return changes
}

// formatFileTable formats file changes as a properly aligned markdown table
func formatFileTable(changes []FileChange) string {
	if len(changes) == 0 {
		return "No files changed.\n"
	}

	// Calculate column widths
	statusWidth := len("Status")
	fileWidth := len("File")
	moduleWidth := len("Module")

	for _, change := range changes {
		if len(change.Status) > statusWidth {
			statusWidth = len(change.Status)
		}
		if len(change.FilePath) > fileWidth {
			fileWidth = len(change.FilePath)
		}
		if len(change.Module) > moduleWidth {
			moduleWidth = len(change.Module)
		}
	}

	var table bytes.Buffer

	// Header row
	table.WriteString(fmt.Sprintf("| %-*s | %-*s | %-*s |\n",
		statusWidth, "Status",
		fileWidth, "File",
		moduleWidth, "Module"))

	// Separator row
	table.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
		strings.Repeat("-", statusWidth),
		strings.Repeat("-", fileWidth),
		strings.Repeat("-", moduleWidth)))

	// Data rows
	for _, change := range changes {
		table.WriteString(fmt.Sprintf("| %-*s | %-*s | %-*s |\n",
			statusWidth, change.Status,
			fileWidth, change.FilePath,
			moduleWidth, change.Module))
	}

	return table.String()
}

// extractModelFromAgent parses the agent file frontmatter and extracts the model field
func extractModelFromAgent(agentContent string) string {
	// Look for "model: <name>" in the frontmatter
	lines := strings.Split(agentContent, "\n")
	inFrontmatter := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			if inFrontmatter {
				// End of frontmatter, didn't find model
				break
			}
			inFrontmatter = true
			continue
		}

		if inFrontmatter && strings.HasPrefix(trimmed, "model:") {
			// Extract model name
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				model := strings.TrimSpace(parts[1])
				fmt.Fprintf(os.Stderr, "[DEBUG] Extracted model from agent: %s\n", model)
				return model
			}
		}
	}

	// No model specified, return empty string (will use default)
	fmt.Fprintf(os.Stderr, "[DEBUG] No model found in agent frontmatter, using default\n")
	return ""
}

func sendResponse(encoder *json.Encoder, id interface{}, result interface{}) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	encoder.Encode(resp)
}

func sendError(encoder *json.Encoder, id interface{}, code int, message string) {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    code,
			Message: message,
		},
	}
	encoder.Encode(resp)
}

// Progress notification structure
type ProgressNotification struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// formatDuration formats duration as 00m00s, or just 00s if less than 1 minute
func formatDuration(seconds float64) string {
	totalSecs := int(seconds)
	mins := totalSecs / 60
	secs := totalSecs % 60

	if mins == 0 {
		if seconds == 0 {
			return fmt.Sprintf("nil", secs)
		}
		return fmt.Sprintf("%02ds", secs)
	}

	return fmt.Sprintf("%02dm%02ds", mins, secs)
}

// sendProgress sends a progress notification to stdout (for the extension to display)
func sendProgress(stage string, message string) {
	// Calculate current elapsed times
	globalElapsed := time.Since(progressStartTime).Seconds()

	// Initialize stage timer if this is the first call
	if stageStartTime.IsZero() {
		stageStartTime = time.Now()
	}
	localElapsed := time.Since(stageStartTime).Seconds()

	// Format stage/header with LAST stage's completion times
	// This shows how long the previous stage took
	var stageWithTime string
	if lastStageGlobalTime > 0 {
		// Show previous stage's times in the header
		stageWithTime = fmt.Sprintf("%s (%s:%s)", stage,
			formatDuration(lastStageGlobalTime),
			formatDuration(lastStageLocalTime))
	} else {
		// First stage - no previous times to show
		stageWithTime = fmt.Sprintf("%s (00m00s:00m00s)", stage)
	}

	// Debug: Log to stderr so we can see if this is being called
	fmt.Fprintf(os.Stderr, "[DEBUG] Sending progress: %s - %s\n", stageWithTime, message)

	notification := ProgressNotification{
		JSONRPC: "2.0",
		Method:  "$/progress",
		Params: map[string]interface{}{
			"stage":   stageWithTime,
			"message": message,
		},
	}

	// Save current times as "last stage" for next progress call
	lastStageGlobalTime = globalElapsed
	lastStageLocalTime = localElapsed

	// Reset stage timer for next stage
	stageStartTime = time.Now()

	// Use the shared encoder to ensure consistent JSON output
	if stdoutEncoder != nil {
		if err := stdoutEncoder.Encode(notification); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to encode progress notification: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] Progress notification sent successfully\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] ERROR: stdoutEncoder is nil!\n")
	}
}

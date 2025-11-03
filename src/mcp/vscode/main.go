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

	"github.com/ready-to-release/eac/src/contracts/modules"
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
var stdoutWriter *bufio.Writer

// Global start time for tracking elapsed time in progress messages
var progressStartTime time.Time
var stageStartTime time.Time
var lastStageGlobalTime float64
var lastStageLocalTime float64
var lastStageName string

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Create buffered writer for stdout so we can explicitly flush
	stdoutWriter = bufio.NewWriter(os.Stdout)
	stdoutEncoder = json.NewEncoder(stdoutWriter)

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

		// Anti-corruption layer architecture:
		// - Agents are instructed to output ONLY pure content (no meta-commentary)
		// - callClaude() applies extractContentBlock() to strip any wrapper text
		// - This function returns exactly ONE Content block to VSCode extension
		// This ensures clean separation: agents produce content, Go ensures purity
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

	// Step 5.5: Early validation for file/module completeness
	// This provides ONE-TIME feedback to generator before continuing the pipeline
	sendProgress("gen-validate", "Validating file completeness...")
	earlyErrors := validateFileCompleteness(initialCommit, gitContext)
	if len(earlyErrors) > 0 {
		sendProgress("gen-feedback", "Providing feedback to generator...")
		fmt.Fprintf(os.Stderr, "[DEBUG] Early validation found %d completeness issues, regenerating with feedback\n", len(earlyErrors))

		// Build feedback prompt with specific issues
		feedbackPrompt := buildFeedbackPrompt(generatorPrompt, initialCommit, earlyErrors)

		// ONE-TIME regeneration with feedback
		sendProgress("gen-regen", "Regenerating with feedback...")
		regenerated, err := callClaude(feedbackPrompt, generatorModel)
		if err != nil {
			// If regeneration fails, continue with original (validator will catch issues later)
			fmt.Fprintf(os.Stderr, "[WARN] Regeneration failed: %v, continuing with original\n", err)
		} else {
			// Use regenerated version
			initialCommit = regenerated
			fmt.Fprintf(os.Stderr, "[DEBUG] Regeneration successful, using improved version\n")
		}
	}

	// Step 6: Call reviewer and title generator in parallel using goroutines
	sendProgress("parallel-init", "Loading reviewer and title generator agents...")

	reviewerPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-message-reviewer.md")
	titleGenPath := filepath.Join(workspaceRoot, ".claude", "agents", "commit-title-generator.md")

	reviewerInstructions, err := ioutil.ReadFile(reviewerPath)
	if err != nil {
		return "", fmt.Errorf("failed to read reviewer agent file: %w", err)
	}

	titleGenInstructions, err := ioutil.ReadFile(titleGenPath)
	if err != nil {
		return "", fmt.Errorf("failed to read title generator agent file: %w", err)
	}

	// Force both agents to use haiku for speed
	reviewerModel := "haiku"
	titleGenModel := "haiku"

	sendProgress("parallel-exec", "Running reviewer and title generator in parallel...")

	// Channel for review result
	type reviewResult struct {
		review string
		err    error
	}
	reviewChan := make(chan reviewResult, 1)

	// Channel for title result
	type titleResult struct {
		title string
		err   error
	}
	titleChan := make(chan titleResult, 1)

	// Launch reviewer agent in goroutine
	go func() {
		reviewerPrompt := string(reviewerInstructions) + "\n\n---\n\n## Commit Message to Review\n\n```\n" + initialCommit + "\n```"
		review, err := callClaude(reviewerPrompt, reviewerModel)
		reviewChan <- reviewResult{review: review, err: err}
	}()

	// Launch title generator agent in goroutine
	go func() {
		titleGenPrompt := string(titleGenInstructions) + "\n\n---\n\n" + initialCommit
		title, err := callClaude(titleGenPrompt, titleGenModel)
		titleChan <- titleResult{title: title, err: err}
	}()

	// Wait for both to complete
	reviewRes := <-reviewChan
	titleRes := <-titleChan

	if reviewRes.err != nil {
		return "", fmt.Errorf("failed to review commit: %w", reviewRes.err)
	}

	if titleRes.err != nil {
		return "", fmt.Errorf("failed to generate commit title: %w", titleRes.err)
	}

	review := reviewRes.review
	commitTitle := strings.TrimSpace(titleRes.title)
	baseCommit := strings.TrimSpace(initialCommit)

	// Debug: Log the commit title
	fmt.Fprintf(os.Stderr, "[DEBUG] Commit title: '%s' (length: %d)\n", commitTitle, len(commitTitle))

	// Step 7: Stitch together final commit message from all agents
	sendProgress("stitch", "Stitching agent outputs...")

	// Build the composite commit message
	var composite strings.Builder

	// 1. Add commit title as top-level heading (MD041)
	composite.WriteString("# ")
	composite.WriteString(commitTitle)
	composite.WriteString("\n\n")

	fmt.Fprintf(os.Stderr, "[DEBUG] After adding title, composite starts with: '%s'\n", composite.String()[:min(100, composite.Len())])

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

	// 3. Add review feedback as a section
	composite.WriteString("## Review\n\n")
	composite.WriteString(review)
	composite.WriteString("\n")

	finalCommit := composite.String()

	// Step 8: Apply pre-validation auto-corrections
	sendProgress("auto-correct", "Applying formatting corrections...")
	fmt.Fprintf(os.Stderr, "[DEBUG] Before auto-correct, first 300 chars:\n%s\n", finalCommit[:min(300, len(finalCommit))])
	finalCommit = autoCorrectCommitMessage(finalCommit)
	fmt.Fprintf(os.Stderr, "[DEBUG] After auto-correct, first 300 chars:\n%s\n", finalCommit[:min(300, len(finalCommit))])

	// Step 9: Validate the final commit message
	sendProgress("validate", "Validating commit message structure...")
	validationErrors := validateCommitMessage(finalCommit, workspaceRoot, gitContext)
	if len(validationErrors) > 0 {
		// Append validation errors to commit message for user awareness
		var errorSection strings.Builder
		errorSection.WriteString("\n\n---\n\n")
		errorSection.WriteString("⚠️ **VALIDATION WARNINGS** - Review before committing:\n\n")
		for _, verr := range validationErrors {
			errorSection.WriteString(fmt.Sprintf("- %s\n", verr.Message))
		}
		finalCommit = finalCommit + errorSection.String()
		sendProgress("complete", "Complete (with validation warnings)")
		return finalCommit, nil
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

	// Step 5: Apply auto-corrections
	sendProgress("auto-correct", "Applying formatting corrections...")
	commitMessage = autoCorrectCommitMessage(commitMessage)

	sendProgress("complete", "Complete!")
	return commitMessage, nil
}

type GitContext struct {
	Status      string       // git status --porcelain output
	Diff        string       // git diff --staged output
	HeadSHA     string       // current commit SHA
	FileChanges []FileChange // parsed from Status
}

// CommitValidationError represents a validation error
type CommitValidationError struct {
	Field   string
	Message string
}

func (e CommitValidationError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Field, e.Message)
}

// autoCorrectCommitMessage applies simple auto-corrections before validation
// This fixes common formatting issues without needing the full fixer agent
func autoCorrectCommitMessage(commitMessage string) string {
	lines := strings.Split(commitMessage, "\n")
	var corrected []string

	yamlBlockOpen := false
	hasTopLevelHeading := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track top-level heading
		if strings.HasPrefix(line, "# ") {
			hasTopLevelHeading = true
		}

		// Track YAML blocks
		if strings.HasPrefix(trimmed, "```yaml") {
			yamlBlockOpen = true
			corrected = append(corrected, line)
			continue
		}

		// Check if we're closing a YAML block
		if yamlBlockOpen && strings.HasPrefix(trimmed, "```") && trimmed == "```" {
			yamlBlockOpen = false
			corrected = append(corrected, line)
			continue
		}

		// If we hit next module section (##) while yaml block is open, close it first
		if yamlBlockOpen && strings.HasPrefix(trimmed, "## ") {
			corrected = append(corrected, "```")
			yamlBlockOpen = false
		}

		// If we hit "Agent: " while yaml block is open, close it first
		if yamlBlockOpen && strings.HasPrefix(trimmed, "Agent:") {
			corrected = append(corrected, "```")
			yamlBlockOpen = false
		}

		corrected = append(corrected, line)
	}

	// If yaml block is still open at end, close it
	if yamlBlockOpen {
		corrected = append(corrected, "```")
	}

	result := strings.Join(corrected, "\n")

	// Debug: Log if heading is missing
	if !hasTopLevelHeading {
		fmt.Fprintf(os.Stderr, "[WARNING] Auto-correct: No top-level heading found in commit message!\n")
		fmt.Fprintf(os.Stderr, "[DEBUG] First 200 chars: %s\n", result[:min(200, len(result))])
	}

	// Ensure file ends with newline
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// validateCommitMessage validates the structure and content of a commit message
func validateCommitMessage(commitMessage string, workspaceRoot string, gitContext *GitContext) []CommitValidationError {
	var errors []CommitValidationError

	// Load module contracts using the contracts library
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot, "0.1.0")
	if err != nil {
		errors = append(errors, CommitValidationError{
			Field:   "contracts",
			Message: fmt.Sprintf("Failed to load module contracts: %v", err),
		})
		// Continue with validation even if contracts failed to load
		moduleRegistry = modules.NewRegistry("0.1.0", workspaceRoot) // Empty registry for fallback
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
				if moduleRegistry.Count() > 0 {
					if !moduleRegistry.Has(moduleName) {
						errors = append(errors, CommitValidationError{
							Field:   "module-unknown",
							Message: fmt.Sprintf("Line %d: Module '%s' not found in contracts/modules/0.1.0/", i+1, moduleName),
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

	// Validate 1-1 file match: every staged file MUST appear in Files affected table
	if gitContext != nil {
		// Extract files from the "## Files affected" table in the commit message
		filesInTable := extractFilesFromTable(commitMessage)

		// Create sets for comparison
		stagedFiles := make(map[string]bool)
		for _, change := range gitContext.FileChanges {
			stagedFiles[change.FilePath] = true
		}

		tableFiles := make(map[string]bool)
		for _, file := range filesInTable {
			tableFiles[file] = true
		}

		// Check for files in staged but NOT in table
		var missingFiles []string
		for file := range stagedFiles {
			if !tableFiles[file] {
				missingFiles = append(missingFiles, file)
			}
		}

		// Check for files in table but NOT in staged
		var extraFiles []string
		for file := range tableFiles {
			if !stagedFiles[file] {
				extraFiles = append(extraFiles, file)
			}
		}

		// Report missing files
		if len(missingFiles) > 0 {
			errors = append(errors, CommitValidationError{
				Field: "file-completeness",
				Message: fmt.Sprintf("Files affected table is missing %d staged file(s):\n  - %s",
					len(missingFiles), strings.Join(missingFiles, "\n  - ")),
			})
		}

		// Report extra files
		if len(extraFiles) > 0 {
			errors = append(errors, CommitValidationError{
				Field: "file-accuracy",
				Message: fmt.Sprintf("Files affected table contains %d file(s) that are NOT staged:\n  - %s",
					len(extraFiles), strings.Join(extraFiles, "\n  - ")),
			})
		}
	}

	return errors
}

// validateFileCompleteness performs early validation for file/module completeness
// This is used to provide ONE-TIME feedback to the generator before the full pipeline
func validateFileCompleteness(commitMessage string, gitContext *GitContext) []CommitValidationError {
	var errors []CommitValidationError

	if gitContext == nil {
		return errors
	}

	// Extract files from the "## Files affected" table
	filesInTable := extractFilesFromTable(commitMessage)

	// Create sets for comparison
	stagedFiles := make(map[string]bool)
	for _, change := range gitContext.FileChanges {
		stagedFiles[change.FilePath] = true
	}

	tableFiles := make(map[string]bool)
	for _, file := range filesInTable {
		tableFiles[file] = true
	}

	// Check for files in staged but NOT in table
	var missingFiles []string
	for file := range stagedFiles {
		if !tableFiles[file] {
			missingFiles = append(missingFiles, file)
		}
	}

	// Check for files in table but NOT in staged
	var extraFiles []string
	for file := range tableFiles {
		if !stagedFiles[file] {
			extraFiles = append(extraFiles, file)
		}
	}

	// Report missing files (completeness issue)
	if len(missingFiles) > 0 {
		errors = append(errors, CommitValidationError{
			Field: "file-completeness",
			Message: fmt.Sprintf("Files affected table is missing %d staged file(s):\n  - %s",
				len(missingFiles), strings.Join(missingFiles, "\n  - ")),
		})
	}

	// Report extra files (accuracy issue)
	if len(extraFiles) > 0 {
		errors = append(errors, CommitValidationError{
			Field: "file-accuracy",
			Message: fmt.Sprintf("Files affected table contains %d file(s) that are NOT staged:\n  - %s",
				len(extraFiles), strings.Join(extraFiles, "\n  - ")),
		})
	}

	// Check module completeness: all modules with changes should have ## sections
	modulesInGit := make(map[string]bool)
	for _, change := range gitContext.FileChanges {
		if change.Module != "" && change.Module != "unknown" {
			modulesInGit[change.Module] = true
		}
	}

	modulesInCommit := extractModuleSections(commitMessage)
	modulesInCommitMap := make(map[string]bool)
	for _, module := range modulesInCommit {
		modulesInCommitMap[module] = true
	}

	// Find modules that have file changes but no section
	var missingModules []string
	for module := range modulesInGit {
		if !modulesInCommitMap[module] {
			missingModules = append(missingModules, module)
		}
	}

	if len(missingModules) > 0 {
		errors = append(errors, CommitValidationError{
			Field: "module-completeness",
			Message: fmt.Sprintf("Commit message is missing sections for %d module(s) that have changes:\n  - %s",
				len(missingModules), strings.Join(missingModules, "\n  - ")),
		})
	}

	return errors
}

// buildFeedbackPrompt creates a prompt for the generator to regenerate with specific feedback
func buildFeedbackPrompt(originalPrompt string, firstAttempt string, validationErrors []CommitValidationError) string {
	var feedback strings.Builder

	feedback.WriteString(originalPrompt)
	feedback.WriteString("\n\n---\n\n")
	feedback.WriteString("# VALIDATOR FEEDBACK - ONE-TIME ADJUSTMENT\n\n")
	feedback.WriteString("Your first attempt had the following completeness issues:\n\n")

	for _, err := range validationErrors {
		feedback.WriteString(fmt.Sprintf("## %s\n\n", err.Field))
		feedback.WriteString(fmt.Sprintf("%s\n\n", err.Message))
	}

	feedback.WriteString("## Your First Attempt\n\n")
	feedback.WriteString("```\n")
	feedback.WriteString(firstAttempt)
	feedback.WriteString("\n```\n\n")

	feedback.WriteString("## Instructions\n\n")
	feedback.WriteString("Please regenerate the commit message, fixing ALL the issues listed above:\n\n")
	feedback.WriteString("1. Add ALL missing files to the Files affected table\n")
	feedback.WriteString("2. Remove any files that are not actually staged\n")
	feedback.WriteString("3. Ensure EVERY module with changes has a corresponding ## module section\n")
	feedback.WriteString("4. Follow the exact same format and structure as before\n\n")
	feedback.WriteString("Output the COMPLETE corrected commit message (starting with ## Summary).\n")

	return feedback.String()
}

// extractModuleSections parses the commit message and returns list of module section headers
func extractModuleSections(commitMessage string) []string {
	var modules []string
	lines := strings.Split(commitMessage, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for ## headers that are NOT Summary or Files affected
		if strings.HasPrefix(trimmed, "## ") {
			header := strings.TrimPrefix(trimmed, "## ")
			header = strings.TrimSpace(header)

			// Skip special sections
			if header != "Summary" && header != "Files affected" && header != "Approved" {
				modules = append(modules, header)
			}
		}
	}

	return modules
}

// extractFilesFromTable parses the "## Files affected" table and returns the list of file paths
func extractFilesFromTable(commitMessage string) []string {
	var files []string
	lines := strings.Split(commitMessage, "\n")
	inFilesTable := false
	tableHeaderSeen := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect start of Files affected section
		if strings.HasPrefix(trimmed, "## Files affected") {
			inFilesTable = true
			continue
		}

		// Exit table when we hit next section or horizontal rule
		if inFilesTable && (strings.HasPrefix(trimmed, "##") || trimmed == "---") {
			break
		}

		// Parse table rows
		if inFilesTable && strings.HasPrefix(trimmed, "|") {
			// Skip header row (| Status | File | Module |)
			if strings.Contains(trimmed, "Status") && strings.Contains(trimmed, "File") {
				tableHeaderSeen = false
				continue
			}

			// Skip separator row (| -------- | --- | --- |)
			if strings.Contains(trimmed, "---") {
				tableHeaderSeen = true
				continue
			}

			// Parse data rows (after header)
			if tableHeaderSeen {
				// Split by | and extract the File column (index 2)
				parts := strings.Split(trimmed, "|")
				if len(parts) >= 4 { // Expected: empty, Status, File, Module, empty
					filePath := strings.TrimSpace(parts[2])
					// Preserve the full path as-is (including "old -> new" for renames)
					// This matches what git status shows and what's in FileChange.FilePath
					if filePath != "" {
						files = append(files, filePath)
					}
				}
			}
		}
	}

	return files
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
		"docs/explanation/continuous-delivery/trunk-based-development.md",
		"docs/reference/continuous-delivery/repository-layout.md",
		"docs/reference/continuous-delivery/versioning.md",
		"docs/reference/continuous-delivery/semantic-commits.md",
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
	moduleRegistry, err := modules.LoadFromWorkspace(workspaceRoot, "0.1.0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load module contracts: %v\n", err)
		moduleRegistry = modules.NewRegistry("0.1.0", workspaceRoot) // Empty registry for fallback
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
	for _, moduleName := range modules {
		globs := getModuleGlobPattern(moduleName, moduleRegistry)
		prompt.WriteString(fmt.Sprintf("**%s:**\n", moduleName))
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

// extractContentBlock implements an anti-corruption layer that extracts
// pure content from agent output, removing any meta-commentary or instructions
func extractContentBlock(agentOutput string) string {
	// Anti-corruption layer: agents should only output content
	// This function ensures we only extract the actual content block,
	// stripping any conversational wrapper or meta-commentary

	// Common patterns to strip:
	// - "Let me output...", "Here is...", "I will provide..."
	// - "Agent: Approved", "Agent: ..." at the end (but KEEP if in commit message body!)
	// - Markdown code fences around the content

	// Debug log
	fmt.Fprintf(os.Stderr, "[DEBUG] extractContentBlock input length: %d bytes\n", len(agentOutput))

	lines := strings.Split(agentOutput, "\n")
	var contentLines []string
	inContent := false
	skipNextEmpty := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip meta-commentary at the start
		if !inContent {
			// Skip lines that look like conversational wrapper
			if strings.HasPrefix(trimmed, "Let me ") ||
				strings.HasPrefix(trimmed, "Here is ") ||
				strings.HasPrefix(trimmed, "I will ") ||
				strings.HasPrefix(trimmed, "I'll ") ||
				strings.HasPrefix(trimmed, "I'm ") ||
				strings.HasPrefix(trimmed, "The corrected ") ||
				strings.HasPrefix(trimmed, "The generated ") ||
				strings.HasPrefix(trimmed, "After reviewing") ||
				(strings.HasPrefix(trimmed, "The ") && strings.Contains(trimmed, "message")) {
				continue
			}

			// Skip opening markdown fence
			if trimmed == "```" || trimmed == "```markdown" || trimmed == "```yaml" {
				skipNextEmpty = true
				continue
			}

			// Skip horizontal rules at start (but allow up to 5 lines in case of frontmatter)
			if trimmed == "---" && i < 5 {
				continue
			}

			// Skip empty lines before content starts
			if trimmed == "" {
				if skipNextEmpty {
					skipNextEmpty = false
				}
				continue
			}

			// Content detection: starts with # (heading), ## (section), or any text
			// This catches both "# title" and "## Summary" formats
			inContent = true
		}

		// Strip "Agent: ..." suffix lines
		if strings.HasPrefix(trimmed, "Agent:") {
			break
		}

		// Strip closing markdown fence
		if trimmed == "```" && inContent {
			// Check if this is a closing fence for a code block vs the wrapper
			// If we see ``` and the previous non-empty line was a yaml block, it's a valid closing
			// Otherwise it might be the wrapper fence
			if len(contentLines) > 0 {
				// Look back for yaml block
				foundYaml := false
				for j := len(contentLines) - 1; j >= 0 && j >= len(contentLines)-10; j-- {
					if strings.Contains(contentLines[j], "```yaml") {
						foundYaml = true
						break
					}
				}
				if !foundYaml {
					// This is the wrapper closing fence, stop here
					break
				}
			}
		}

		// Collect content line
		if inContent {
			contentLines = append(contentLines, line)
		}
	}

	result := strings.Join(contentLines, "\n")
	result = strings.TrimSpace(result)

	// Debug log
	fmt.Fprintf(os.Stderr, "[DEBUG] extractContentBlock output length: %d bytes\n", len(result))
	if len(result) > 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG] extractContentBlock output preview (first 150 chars): %s\n", result[:min(150, len(result))])
	} else {
		fmt.Fprintf(os.Stderr, "[WARN] extractContentBlock returned EMPTY result!\n")
	}

	return result
}

func callClaude(prompt string, model string) (string, error) {
	// Use Claude Code CLI instead of direct API calls
	// This leverages the user's existing Claude Code subscription

	// Build command with optional model flag
	// IMPORTANT: Disable co-author attribution via settings JSON
	// NOTE: We don't disable prompt caching as it significantly improves performance
	// and Claude Code's API-level prompt caching is smart enough to invalidate
	// when the actual git diff changes
	// NOTE: We don't use --setting-sources "" because it bypasses authentication
	args := []string{
		"--print",
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

	// Apply anti-corruption layer: extract only the content block
	rawOutput := stdout.String()
	fmt.Fprintf(os.Stderr, "[DEBUG] Raw agent output length: %d bytes\n", len(rawOutput))

	// The output should be the structured commit message
	commitMessage := extractContentBlock(rawOutput)
	fmt.Fprintf(os.Stderr, "[DEBUG] After extractContentBlock: %d bytes\n", len(commitMessage))

	// Additional cleanup layers (for backward compatibility)
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

	fmt.Fprintf(os.Stderr, "[DEBUG] Final cleaned output: %d bytes, starts with: %s\n",
		len(commitMessage), commitMessage[:min(50, len(commitMessage))])

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

	// Special case: README.md and documentation files
	// These belong to their parent module, NOT a separate "README.md" module
	if fileName == "README.md" || fileName == "CONTRIBUTING.md" {
		// Root README files are docs
		if !strings.Contains(filePath, "/") {
			return "docs"
		}
		// For module READMEs (e.g., automation/sh/vscode/README.md),
		// remove the filename and detect the parent directory's module
		parentDir := filepath.Dir(filePath)
		if parentDir != "." && parentDir != "/" {
			// Recursively detect module from parent directory
			// Add a dummy file to make pattern matching work
			return determineFileModule(parentDir + "/dummy.file")
		}
		// Fallback: if in src/ or docs/, it's documentation
		if strings.HasPrefix(filePath, "src/") || strings.HasPrefix(filePath, "docs/") {
			return "docs"
		}
	}

	// Special case: ALL definitions.yml files anywhere in repo → repo-config
	// These are repository-level metadata files regardless of location
	// Examples: /definitions.yml, contracts/deployable-units/0.1.0/definitions.yml, etc.
	if fileName == "definitions.yml" {
		return "repo-config"
	}

	// Pattern 1: automation/<type>/<name>/... → extract module name as "<type>-<name>"
	// Examples: automation/sh/vscode/ → "sh-vscode", automation/pwsh/build/ → "pwsh-build"
	// Special case: automation/<module-name>/ (flat) → "<module-name>"
	if strings.HasPrefix(filePath, "automation/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[1] != "" && parts[2] != "" {
			// Multi-level: automation/<type>/<name>/ → "type-name"
			return parts[1] + "-" + parts[2]
		} else if len(parts) >= 2 && parts[1] != "" {
			// Flat: automation/<module>/ → "module"
			return parts[1]
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
	// Examples: .vscode/extensions/vscode-ext-commit/ → "vscode-ext-commit"
	if strings.HasPrefix(filePath, ".vscode/extensions/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[2] != "" {
			// Return the exact extension folder name (matches contract moniker)
			return parts[2]
		}
		return "vscode-ext-commit" // fallback
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
func getModuleGlobPattern(moduleName string, registry *modules.Registry) []string {
	// First, check if we have a contract for this module
	if moduleContract, exists := registry.Get(moduleName); exists {
		// Use the GetGlobPatterns method from the module contract
		return moduleContract.GetGlobPatterns()
	}

	// Fallback to hardcoded patterns for modules without contracts
	// Normalize path separators for cross-platform compatibility
	// GitHub Actions uses forward slashes even on Windows

	// Handle MCP servers
	if strings.HasPrefix(moduleName, "mcp-") {
		service := strings.TrimPrefix(moduleName, "mcp-")
		return []string{fmt.Sprintf("src/mcp/%s/**", service)}
	}

	// Handle automation modules (sh-, pwsh-, py-, etc.)
	if strings.HasPrefix(moduleName, "sh-") || strings.HasPrefix(moduleName, "pwsh-") || strings.HasPrefix(moduleName, "py-") {
		return []string{fmt.Sprintf("automation/%s/**", moduleName)}
	}

	// Handle contract modules
	if strings.HasPrefix(moduleName, "contracts-") {
		contractType := strings.TrimPrefix(moduleName, "contracts-")
		return []string{fmt.Sprintf("contracts/%s/**", contractType)}
	}

	// Handle specific known modules
	switch moduleName {
	case "vscode-ext-claude-commit":
		return []string{".vscode/extensions/vscode-ext-commit/**"}

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
		if !strings.Contains(moduleName, "-") && moduleName != "unknown" {
			// Likely a container module, try that first
			return []string{fmt.Sprintf("containers/%s/**", moduleName)}
		}

		// Unknown module - return a safe default
		return []string{fmt.Sprintf("**/%s/**", moduleName)}
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

			// Handle renamed files: "old -> new" format
			// Extract destination for module detection, keep full path for FilePath
			moduleDetectionPath := filename
			if strings.Contains(filename, " -> ") {
				renameParts := strings.SplitN(filename, " -> ", 2)
				if len(renameParts) == 2 {
					moduleDetectionPath = strings.TrimSpace(renameParts[1])
				}
			}

			changes = append(changes, FileChange{
				Status:   normalizeGitStatus(statusCode),
				FilePath: filename, // Preserves "old -> new" format for renames
				Module:   determineFileModule(moduleDetectionPath),
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
	// Flush after sending response to ensure immediate delivery
	if stdoutWriter != nil {
		stdoutWriter.Flush()
	}
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
	// Flush after sending error to ensure immediate delivery
	if stdoutWriter != nil {
		stdoutWriter.Flush()
	}
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
			return "00s"
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
			// CRITICAL: Flush immediately so progress appears without buffering delay
			if stdoutWriter != nil {
				if err := stdoutWriter.Flush(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to flush stdout: %v\n", err)
				}
			}
			fmt.Fprintf(os.Stderr, "[DEBUG] Progress notification sent successfully\n")
		}
	} else {
		fmt.Fprintf(os.Stderr, "[DEBUG] ERROR: stdoutEncoder is nil!\n")
	}
}

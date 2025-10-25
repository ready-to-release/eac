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
	"strings"
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

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req MCPRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(encoder, nil, -32700, "Parse error")
			continue
		}

		handleRequest(encoder, &req)
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
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: "Error: agentFile must be a string",
				}},
			}
		}

		commitMessage, err := generateSemanticCommitMessage(agentFile)
		if err != nil {
			return ToolResult{
				Content: []Content{{
					Type: "text",
					Text: fmt.Sprintf("Error generating commit message: %v", err),
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
	// Get workspace root from agent file path
	// Agent file is at .claude/agents/vscode-ext-commit-button.md
	workspaceRoot := filepath.Dir(filepath.Dir(filepath.Dir(agentFile)))

	// Step 1: Read the agent instructions
	sendProgress("init", "Loading agent configuration...")
	agentInstructions, err := ioutil.ReadFile(agentFile)
	if err != nil {
		return "", fmt.Errorf("failed to read agent file: %w", err)
	}

	// Step 2: Gather git context as specified in the agent file
	sendProgress("git", "Gathering git context...")
	gitContext, err := gatherGitContext(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to gather git context: %w", err)
	}

	// Step 3: Read documentation files referenced in the agent
	sendProgress("docs", "Reading documentation...")
	docs, err := readDocumentationFiles(workspaceRoot)
	if err != nil {
		return "", fmt.Errorf("failed to read documentation: %w", err)
	}

	// Step 4: Build the prompt for Claude
	sendProgress("prompt", "Building prompt...")
	prompt := buildClaudePrompt(string(agentInstructions), gitContext, docs)

	// Step 5: Call Claude to generate the semantic commit message
	sendProgress("claude", "Generating commit message...")
	commitMessage, err := callClaude(prompt)
	if err != nil {
		return "", fmt.Errorf("failed to call Claude: %w", err)
	}

	sendProgress("complete", "Complete!")
	return commitMessage, nil
}

type GitContext struct {
	Status      string
	Diff        string
	RecentLog   string
	HeadSHA     string
	FileChanges []FileChange
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

	// Get git diff for all changes
	diffCmd := exec.Command("git", "diff", "HEAD")
	diffCmd.Dir = workspaceRoot
	diffOutput, err := diffCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff failed: %w", err)
	}
	ctx.Diff = string(diffOutput)

	// Get last 50 commits as specified in agent file
	logCmd := exec.Command("git", "log", "-50", "--oneline", "--no-decorate")
	logCmd.Dir = workspaceRoot
	logOutput, err := logCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}
	ctx.RecentLog = string(logOutput)

	// Parse and normalize file changes
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

func buildClaudePrompt(agentInstructions string, gitCtx *GitContext, docs map[string]string) string {
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

	prompt.WriteString("## Recent Commits (Last 50 - For Style Reference Only)\n\n")
	prompt.WriteString("Use these to match the commit message style, but DO NOT include them in your output:\n\n")
	prompt.WriteString("```\n")
	prompt.WriteString(gitCtx.RecentLog)
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
		globs := getModuleGlobPattern(module)
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
	prompt.WriteString("## mcp-vscode\n\n")
	prompt.WriteString("mcp-vscode: feat: add commit generation\n\n")
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

func callClaude(prompt string) (string, error) {
	// Use Claude Code CLI instead of direct API calls
	// This leverages the user's existing Claude Code subscription

	// Build command with optional model flag
	// Check for CLAUDE_MODEL environment variable (can be set in .mcp.json or system env)
	args := []string{"--print"}
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

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

	// Remove any leading/trailing newlines but preserve structure
	commitMessage = strings.TrimSpace(commitMessage)

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

	// Pattern 3: src/mcp/<service>/... → mcp-<service>
	// Examples: src/mcp/pwsh/ → "mcp-pwsh", src/mcp/vscode/ → "mcp-vscode"
	// Note: README.md files are handled above and won't reach here
	if strings.HasPrefix(filePath, "src/mcp/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[2] != "" {
			return "mcp-" + parts[2]
		}
		return "mcp" // fallback
	}

	// Pattern 3b: src/cli/... → cli module
	if strings.HasPrefix(filePath, "src/cli/") {
		return "cli"
	}

	// Pattern 4: .vscode/extensions/<name>/... → extract extension name or use "vscode-ext"
	// Examples: .vscode/extensions/claude-mcp-vscode/ → "vscode-ext"
	if strings.HasPrefix(filePath, ".vscode/extensions/") {
		parts := strings.Split(filePath, "/")
		if len(parts) >= 3 && parts[2] != "" {
			// Use short form for known extensions
			extName := parts[2]
			if strings.Contains(extName, "claude-mcp") {
				return "vscode-ext"
			}
			return extName
		}
		return "vscode-ext" // fallback
	}

	// Pattern 5: contracts/<name>/<version>/... → contracts-<name>
	// Examples: contracts/repository/0.1.0/ → "contracts-repository"
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
func getModuleGlobPattern(module string) []string {
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
	case "vscode-ext":
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

// sendProgress sends a progress notification to stdout (for the extension to display)
func sendProgress(stage string, message string) {
	notification := ProgressNotification{
		JSONRPC: "2.0",
		Method:  "$/progress",
		Params: map[string]interface{}{
			"stage":   stage,
			"message": message,
		},
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(notification)
}

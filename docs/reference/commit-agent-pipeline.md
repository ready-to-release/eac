# Commit Agent Pipeline Reference

Technical specifications for the 5-agent commit message generation system.

---

## Component Specifications

### VSCode Extension

**Location**: `.vscode/extensions/claude-mcp-vscode/`

**Files**:

- `src/extension.ts` - Main extension logic
- `package.json` - Extension manifest and contributions

**Command**: `claude-mcp-vscode.callMCP`

**UI Elements**:

1. **Status Bar Item**:
   - Idle: `$(robot) Claude Commit` (normal background)
   - Active: `$(sync~spin) Claude Commit` (rainbow cycling background)
   - Location: Right side of status bar

2. **Git SCM Button**:
   - Idle: `$(robot)` icon
   - Active: `$(sync~spin)` icon (animated)
   - Location: Git SCM title bar (navigation group)

3. **Output Channel**: "Claude Commit Agent"

**Git State Validation**:

- Requires: `indexChanges.length > 0`
- Rejects: `workingTreeChanges.length > 0`

---

### MCP Server

**Location**: `src/mcp/vscode/`
**Language**: Go
**Main File**: `main.go`

**Responsibilities**:

1. Implement MCP JSON-RPC protocol
2. Gather git context
3. Load documentation files
4. Detect modules from paths
5. Orchestrate 5 agents
6. Validate final commit message

**Git Context Structure**:

```go
type GitContext struct {
    Status      string        // Raw porcelain output
    Diff        string        // Complete staged diff
    RecentLog   string        // Last 50 commits
    HeadSHA     string        // Current HEAD SHA
    FileChanges []FileChange  // Normalized changes
}

type FileChange struct {
    Status   string  // added, modified, deleted, renamed
    FilePath string  // Relative path
    Module   string  // Detected module name
}
```

**Module Detection Patterns**:

```text
src/mcp/<name>             → src-mcp-<name>
.vscode/extensions/<name>  → vscode-ext-<name>
automation/<name>          → automation-<name>
docs/                      → docs
contracts/                 → contracts
requirements/              → requirements
```

**Documentation Files Loaded**:

```text
docs/explanation/continuous-delivery/trunk-based-development.md
docs/reference/continuous-delivery/repository-layout.md
docs/reference/continuous-delivery/versioning.md
docs/reference/continuous-delivery/semantic-commits.md
contracts/repository/0.1.0/definitions.yml
contracts/deployable-units/0.1.0/*.yml (all modules)
```

**Claude CLI Invocation**:

```go
args := []string{
    "--print",
    "--setting-sources", "",  // Bypass hooks and CLAUDE.md
    "--settings", `{"includeCoAuthoredBy":false}`,  // No footer
    "--model", model,  // From agent frontmatter
}
```

---

## Agent Specifications

### Agent 1: Generator

**File**: `.claude/agents/vscode-extension-commit-button.md`
**Model**: `haiku`

**Input**:

- Git context (diff, status, logs, HEAD SHA)
- Documentation files (inline)
- Module metadata with glob patterns
- File changes table (pre-formatted)

**Output Format**:

````markdown
# Revision <sha>

[2-4 sentence summary of changes and impact]

## Files affected

| Status   | File                        | Module             |
| -------- | --------------------------- | ------------------ |
| added    | src/mcp/vscode/main.go      | src-mcp-vscode     |
| modified | .vscode/.../extension.ts    | vscode-ext-claude  |

## Summary

| Module           | Globs                        |
| ---------------- | ---------------------------- |
| src-mcp-vscode   | `src/mcp/vscode/**`          |
| vscode-ext-claude| `.vscode/extensions/.../**`  |

---

## src-mcp-vscode

src-mcp-vscode: feat: add commit generation

[Body text explaining WHY, wrapped at 72 chars]

```yaml
paths:
  - 'src/mcp/vscode/**'
```

---

## vscode-ext-claude

[Same format for each module]
````

**Rules**:

- One section per module (no duplicates)
- Subject line ≤ 50 characters
- Body text wrapped at 72 characters
- Format: `<module>: <type>: <description>`
- YAML paths block after body (no heading)
- File table appears ONCE at top
- NO Review section

---

### Agent 2: Reviewer

**File**: `.claude/agents/commit-message-reviewer.md`
**Model**: `haiku`

**Input**: Complete commit message from generator

**Output Format**:

```markdown
## Commit Message Review

### Overall Assessment

[Excellent/Good/Needs Improvement/Poor]

### Issues

1. [Most critical issue]

   - Problem: [Explain]
   - Suggestion: [Fix]

2. [Next issue]
   ...

### Recommended Rewrite

[Improved version if needed]

**Additional Notes**: [Best practices advice]
```

**Critical Rule**: Strip ALL positive affirmations ("Good job", "Excellent", "Well done")

---

### Agent 3: Approver

**File**: `.claude/agents/commit-message-approver.md`
**Model**: `haiku`

**Input**:

- Original commit message
- Review section from agent 2

**Output Format (No Concerns)**:

```markdown
## Approved

Approved
```

**Output Format (With Concerns)**:

```markdown
## Approved

Approved (with concerns)

- [Factual issue 1]
- [Factual issue 2]
```

**Rules**:

- Output ONLY the `## Approved` section
- Extract ONLY factual/actionable issues
- Remove all praise and fluff
- Two outcomes: "Approved" or "Approved (with concerns)"

---

### Agent 4: Concerns Handler

**File**: `.claude/agents/commit-message-concerns-handler.md`
**Model**: `haiku`

**Conditional**: Only runs if approver output contains "Approved (with concerns)"

**Input**:

- Original commit message
- `## Approved` section with concerns list

**Output Format**:

```markdown
[Corrected commit message with concerns fixed]
```

**Rules**:

- Apply ALL fixes from concerns list
- Remove the `## Approved` section from output
- Maintain all original formatting and structure
- If no concerns, output original unchanged

---

### Agent 5: Title Generator

**File**: `.claude/agents/commit-title-generator.md`
**Model**: `haiku`

**Input**: Final commit message (corrected or original)

**Output Format**:

```text
feat(src-mcp-vscode): add 5-agent commit workflow
```

**Format**: `<type>(<scope>): <description>`

**Types**:

- `feat` - New feature
- `fix` - Bug fix
- `refactor` - Code restructuring
- `docs` - Documentation
- `chore` - Maintenance, configs, build
- `test` - Tests
- `perf` - Performance
- `style` - Formatting

**Scope Rules**:

- Single module: Use module name
- 2-3 modules: Primary module or "multi-module"
- 4+ modules: "multi-module" or category
- Cross-cutting: Component name

**Rules**:

- Output ONLY the title text (no `#` prefix)
- Maximum 72 characters
- Imperative mood
- No period at end

---

## Final Output Format

### Complete Structure

```markdown
# feat(scope): description

[2-4 sentence summary]

## Files affected

[Table]

## Summary

[Module glob table]

---

## module-name

module: type: subject

Body text explaining WHY.

```yaml
paths:
  - 'glob/pattern/**'
```

---

Agent: Approved

```text

### With Concerns

```markdown
Agent: Approved (with concerns)
```

---

## Progress Notifications

The MCP server sends JSON-RPC notifications:

```json
{
  "jsonrpc": "2.0",
  "method": "$/progress",
  "params": {
    "stage": "gen-claude",
    "message": "Generating initial commit message..."
  }
}
```

**Stages**:

| Stage | Description |
|-------|-------------|
| `git-init` | Gathering git context |
| `docs-init` | Loading documentation |
| `gen-init` | Loading generator agent |
| `gen-prompt` | Building generator prompt |
| `gen-claude` | Generating initial commit |
| `rev-init` | Loading reviewer agent |
| `rev-claude` | Reviewing commit message |
| `app-init` | Loading approver agent |
| `app-claude` | Final approval |
| `concerns-init` | Loading concerns handler |
| `concerns-claude` | Fixing concerns |
| `title-init` | Loading title generator |
| `title-claude` | Generating commit title |
| `stitch` | Stitching outputs |
| `validate` | Validating commit message structure |
| `complete` | Done |

---

## Validation Rules

The MCP server validates final commit messages before returning to extension.

### Validation Checks

1. **Unique Top-Level Heading**:
   - Exactly one `#` heading (MD041 compliance)
   - Error: "commit message must have exactly one top-level heading"

2. **Semantic Format**:
   - All module sections: `<module>: <type>: <description>`
   - Error: "subject line does not follow semantic format"

3. **Valid Semantic Type**:
   - Must be: `feat`, `fix`, `refactor`, `docs`, `chore`, `test`, `perf`, `style`
   - Error: "invalid semantic commit type '<type>'"

4. **Module Name Match**:
   - Subject line module must match section header
   - Error: "subject line does not match module name in section header"

5. **Module Exists**:
   - Module must be in `contracts/deployable-units/0.1.0/<module>.yml`
   - Error: "module '<module>' not found in contracts"

6. **Subject Length**:
   - Must be ≤50 characters
   - Error: "subject line exceeds 50 characters (<N> chars)"

7. **Non-Empty Description**:
   - Cannot be empty string
   - Error: "description cannot be empty"

### Module Contract Format

**Location**: `contracts/deployable-units/0.1.0/<module>.yml`

```yaml
moniker: "src-mcp-vscode"
name: "Go VSCode MCP Server"
type: "mcp-server"
description: "Model Context Protocol server..."
```

**Loading**:

```go
func loadModuleContracts(workspaceRoot string) (map[string]ModuleContract, error)
```

---

## Error Messages

### Extension Errors

| Error | Condition | Message |
|-------|-----------|---------|
| No Git Extension | Git extension not available | "Git extension not found" |
| No Repository | Not in git repository | "No Git repository found" |
| No Staged Changes | `indexChanges.length === 0` | "No staged changes found. Stage your changes before generating a commit message." |
| Unstaged Changes | `workingTreeChanges.length > 0` | "You have unstaged changes. Please stage or stash them before generating a commit message." |
| No Workspace | No workspace folder | "No workspace folder found" |
| Agent Error | MCP server returns error | Shows error from MCP server |

### MCP Server Errors

| Error | Condition | Message |
|-------|-----------|---------|
| Model Not Specified | Agent missing `model:` field | "model not specified in agent frontmatter - all agents must define 'model:' field" |
| Agent File Not Found | Cannot read agent file | "failed to read [agent] agent file: [error]" |
| Git Command Failed | Git command returns error | "git [command] failed: [error]" |
| Documentation Warning | Doc file not found | Warning logged, continues |
| Claude CLI Failed | CLI returns non-zero | "claude CLI failed: [error]\nStderr: [output]" |
| Validation Failed | Commit validation fails | "❌ Commit message validation failed:\n  • [issues]" |

---

## Configuration

### Agent Frontmatter

All agents require YAML frontmatter:

```yaml
---
name: agent-identifier
description: What the agent does
model: haiku
color: blue
---
```

**Required Fields**:

- `name` - Agent identifier
- `description` - Purpose
- `model` - Which Claude model to use
- `color` - UI color hint

### Extension Configuration

No user-facing settings. Behavior is controlled by:

- Git state validation rules (hardcoded)
- Agent file paths (hardcoded relative to workspace)

### MCP Server Configuration

Configured via code (no config file):

- Agent file paths in `.claude/agents/`
- Documentation patterns in `readDocumentationFiles()`
- Module detection patterns in `determineFileModule()`
- Glob patterns in `getModuleGlobPattern()`

---

## Performance Metrics

**Typical Execution**: 30-90 seconds

**Time Breakdown**:

| Stage | Time | Model Calls |
|-------|------|-------------|
| Git context gathering | 1-2s | 0 |
| Documentation loading | 1-2s | 0 |
| Generator agent | 10-30s | 1 (Haiku) |
| Reviewer agent | 5-15s | 1 (Haiku) |
| Approver agent | 3-8s | 1 (Haiku) |
| Concerns handler | 10-30s | 1 (Haiku, conditional) |
| Title generator | 3-8s | 1 (Haiku) |
| Stitching & validation | <1s | 0 |
| **Total** | **30-90s** | **4-5 calls** |

---

## Markdown Compliance

**MD041**: First line must be top-level heading

- Enforced by adding `# <title>` from agent 5

**MD047**: File must end with newline

- Enforced in `callClaude()` function

**Bold Headers**: Converted to markdown headers

- `**Text**` → `### Text`
- Applied in `callClaude()` post-processing

---

## Related Documentation

- [Commit Agent Pipeline Explanation](../explanation/commit-agent-pipeline.md) - Architecture and design
- [Customize Commit Agents](../how-to-guides/vscode-extension/customize-commit-agents.md) - Modification guide
- [VS Code Extension Reference](vscode-extension.md) - Extension API reference

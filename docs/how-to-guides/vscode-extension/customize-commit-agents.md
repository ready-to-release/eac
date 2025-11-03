# Customize Commit Agents

Learn how to modify and extend the commit agent pipeline.

---

## Overview

**What you'll learn:**

- Test individual agents
- Modify agent prompts
- Add a new agent to the pipeline
- Change agent models
- Debug agent issues

**Time Required:** 15-30 minutes

---

## Prerequisites

- VSCode extension running in development mode
- Go installed for MCP server
- Claude Code CLI installed

---

## Test an Agent Individually

Testing agents outside the pipeline helps isolate issues.

### Step 1: Prepare Test Input

Create a test file with sample commit message:

```bash
cat > /tmp/test-commit.md << 'EOF'
# Revision abc123

Added new feature to process data.

## Files affected

| Status   | File               | Module       |
| -------- | ------------------ | ------------ |
| modified | src/main.go        | src-cli      |

## Summary

| Module   | Globs            |
| -------- | ---------------- |
| src-cli  | `src/**`         |

---

## src-cli

src-cli: feat: process data automatically

This enables automatic data processing without manual intervention.

```yaml
paths:
  - 'src/**'
```

EOF

```text

### Step 2: Test Agent

```bash
# Test the reviewer agent
cat /tmp/test-commit.md | \
  claude --print \
  --model haiku \
  --setting-sources "" \
  --file .claude/agents/commit-message-reviewer.md

# Test the title generator
cat /tmp/test-commit.md | \
  claude --print \
  --model haiku \
  --setting-sources "" \
  --file .claude/agents/commit-title-generator.md
```

### Step 3: Verify Output

Check that the agent:

- Follows expected output format
- Produces actionable feedback (reviewer)
- Generates correct title format (title generator)

---

## Modify an Agent Prompt

### Step 1: Locate Agent File

```bash
# All agents are in .claude/agents/
ls .claude/agents/

# Example: modify the reviewer
code .claude/agents/commit-message-reviewer.md
```

### Step 2: Edit Agent Prompt

**File:** `.claude/agents/commit-message-reviewer.md`

```markdown
---
name: commit-message-reviewer
description: Reviews commit messages for quality
model: haiku
color: yellow
---

# Your Modifications Here

For example, add a new check:

## Additional Checks

- [ ] Verify all module descriptions explain WHY, not WHAT
- [ ] Check for security implications mentioned if applicable
- [ ] Ensure performance considerations noted for large changes

[... rest of agent prompt ...]
```

### Step 3: Test Modified Agent

```bash
# Test with your changes
cat /tmp/test-commit.md | \
  claude --print \
  --model haiku \
  --setting-sources "" \
  --file .claude/agents/commit-message-reviewer.md
```

### Step 4: Use in Extension

The extension automatically picks up changes:

1. Click robot button in VSCode
2. Agent will use modified prompt
3. Check Output Channel for agent responses

---

## Change Agent Model

You can use different Claude models for different agents.

### Step 1: Edit Agent Frontmatter

**File:** `.claude/agents/commit-message-concerns-handler.md`

```markdown
---
name: commit-message-concerns-handler
description: Fixes concerns in commit messages
model: sonnet  # Changed from haiku to sonnet
color: green
---
```

### Step 2: Understand Trade-offs

**Haiku**:

- Fast (10-30s)
- Cost-effective
- Good for structured tasks

**Sonnet**:

- Slower (30-90s)
- Higher cost
- Better reasoning
- Use for complex corrections

**Recommendation**: Keep generator as Haiku, consider Sonnet for concerns handler if you need better fixes.

### Step 3: Test Performance

Click robot button and check elapsed time in progress notification.

---

## Add a New Agent to Pipeline

Let's add a "Security Checker" agent between reviewer and approver.

### Step 1: Create Agent File

**File:** `.claude/agents/commit-message-security-checker.md`

```markdown
---
name: commit-message-security-checker
description: Checks for security-related concerns in commits
model: haiku
color: red
---

# Security Checker Agent

You are a security-focused reviewer. Analyze the commit message and identify any security implications.

## Input

You will receive a complete commit message.

## Your Task

1. Check for security-related changes
2. Identify potential security risks
3. Suggest security considerations for the commit message

## Output Format

```markdown
## Security Review

### Security-Related Changes

[List any security-related modifications]

### Potential Risks

[List any security risks or concerns]

### Recommendations

[Suggest security considerations to add to commit body]
```

If no security concerns:

```markdown
## Security Review

No security concerns identified.
```

## Guidelines

- Focus on authentication, authorization, data protection
- Check for exposed credentials or secrets
- Note breaking security changes
- Flag removal of security features

```text

### Step 2: Modify MCP Server

**File:** `src/mcp/vscode/main.go`

Find the agent orchestration section and add your agent:

```go
// After reviewer agent
reviewOutput := callAgent("reviewer", generatorOutput)

// NEW: Security checker
securityOutput := callAgent("security-checker", generatorOutput)

// Before approver (now gets both review and security output)
approverInput := generatorOutput + "\n\n" + reviewOutput + "\n\n" + securityOutput
approverOutput := callAgent("approver", approverInput)
```

### Step 3: Update Progress Stages

Add progress notification for new agent:

```go
case "sec-init":
    return "Loading security checker agent"
case "sec-claude":
    return "Checking for security concerns..."
```

### Step 4: Test

1. Rebuild MCP server: `cd src/mcp/vscode && go build`
2. Make a security-related change (auth, encryption, etc.)
3. Stage changes and click robot button
4. Check Output Channel for security review

---

## Debug Agent Issues

### Problem: Agent Returns Unexpected Format

**Check**:

1. Test agent individually with known input
2. Verify output format matches specification
3. Check for stray text before/after expected output

**Fix**:

```bash
# Add explicit formatting instructions to agent
echo "IMPORTANT: Output ONLY the following format, no additional text:" \
  >> .claude/agents/your-agent.md
```

### Problem: Agent Takes Too Long

**Check**:

1. Which agent is slow? (check Output Channel)
2. Is model appropriate? (Haiku vs Sonnet)
3. Is prompt too complex?

**Fix**:

```markdown
# In agent frontmatter, change model
model: haiku  # Faster model
```

### Problem: Agent Doesn't Follow Instructions

**Diagnosis**:

```bash
# Test with minimal input
echo "test" | claude --print --model haiku --file .claude/agents/your-agent.md
```

**Common Issues**:

- Instructions buried in long prompt
- Conflicting instructions
- Example format doesn't match specification

**Fix**:

1. Move critical instructions to top of prompt
2. Use numbered lists for clarity
3. Provide clear examples
4. Test incrementally

### Problem: MCP Server Can't Find Agent

**Error**: `failed to read [agent] agent file: no such file`

**Check**:

```bash
# Verify file exists
ls .claude/agents/your-agent.md

# Check file path in Go code
grep "your-agent" src/mcp/vscode/main.go
```

**Fix**: Ensure file name matches the name used in Go code.

---

## Development Workflow

### Iterative Agent Development

```bash
# 1. Create agent file
code .claude/agents/new-agent.md

# 2. Test with simple input
echo "test input" | \
  claude --print --model haiku --file .claude/agents/new-agent.md

# 3. Refine prompt based on output

# 4. Test with realistic input
cat /tmp/test-commit.md | \
  claude --print --model haiku --file .claude/agents/new-agent.md

# 5. Integrate into pipeline (modify Go code)

# 6. Rebuild MCP server
cd src/mcp/vscode && go build

# 7. Test in VSCode extension
```

### Testing MCP Server Changes

```bash
# Build server
cd src/mcp/vscode
go build

# The extension will use the rebuilt binary on next invocation
# No need to restart VSCode
```

### View Agent Execution

Check the "Claude Commit Agent" output channel:

```text
══════════════════════════════════════════════
🤖 AGENT 1: GENERATOR
══════════════════════════════════════════════
Input length: 2543 characters
[agent output here]

══════════════════════════════════════════════
📝 AGENT 2: REVIEWER
══════════════════════════════════════════════
Input length: 1872 characters
[agent output here]
```

---

## Best Practices

### Agent Prompts

1. **Be Explicit**: Don't assume the model knows your format
2. **Use Examples**: Show exact output format
3. **Single Responsibility**: One agent, one clear task
4. **Test Independently**: Verify agent works alone first
5. **Keep It Simple**: Complex prompts = unreliable outputs

### Model Selection

1. **Start with Haiku**: Fast enough for most tasks
2. **Use Sonnet for**: Complex reasoning, ambiguous cases
3. **Never use Opus**: Too slow for commit messages

### Pipeline Design

1. **Sequential Dependencies**: Each agent builds on previous
2. **Conditional Execution**: Skip agents when possible
3. **Validate Early**: Catch format errors before final output
4. **Progress Feedback**: Keep user informed

---

## Common Modifications

### Change Commit Message Format

Modify generator agent to change structure:

**File:** `.claude/agents/vscode-extension-commit-button.md`

```markdown
# Add a new section after Summary

## Impact Assessment

| Category | Impact | Description |
| -------- | ------ | ----------- |
| Users    | High   | ... |
| Performance | Low | ... |
```

### Add Custom Validation

Modify MCP server `validateCommitMessage()`:

**File:** `src/mcp/vscode/main.go`

```go
// Add custom check
if strings.Contains(commitMessage, "TODO") {
    errors = append(errors, CommitValidationError{
        Field:   "body",
        Message: "commit message contains TODO - resolve before committing",
    })
}
```

### Customize Documentation Loaded

Modify MCP server `readDocumentationFiles()`:

**File:** `src/mcp/vscode/main.go`

```go
// Add your custom docs
docPatterns := []string{
    "docs/reference/continuous-delivery/...",
    "docs/custom/my-guidelines.md",  // Add this
}
```

---

## Related Documentation

- [Commit Agent Pipeline Explanation](../../explanation/commit-agent-pipeline.md) - Architecture
- [Commit Agent Pipeline Reference](../../reference/commit-agent-pipeline.md) - Technical specs
- [MCP Server Development](work-with-mcp-servers.md) - MCP server development

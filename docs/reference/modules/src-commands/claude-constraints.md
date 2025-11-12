# Claude Integration Constraints

## Context

The `commit-ai` command needs to invoke Claude programmatically to generate commit messages following the contract specification at `contracts/commit-message/0.1.0/structure.yml`.

## Constraints

### Authentication Limitation

**Constraint**: No Anthropic API key available

- Using Claude Pro/Max subscription (no API access)
- Cannot use Anthropic API directly
- Must use alternative method for programmatic access

### Requirements

1. **Single prompt** → Send context to Claude once
2. **Structured output** → Receive formatted commit message back
3. **Programmatic invocation** → From Go binary (`commit-ai.go`)
4. **Contract compliance** → Output must follow `structure.yml` specification
5. **No API costs** → Must use subscription credits, not API credits

## Solution Architecture

### Approach: Claude CLI with Subscription Auth

Use the `claude` CLI tool with subscription authentication by removing API key from environment:

```go
cmd := exec.Command("claude", args...)
cmd.Stdin = strings.NewReader(fullPrompt)

// CRITICAL: Remove ANTHROPIC_API_KEY to force subscription auth
cmd.Env = removeAPIKeyFromEnv(os.Environ())
```

**Why this works:**

- Claude CLI supports both API key AND subscription auth
- By removing `ANTHROPIC_API_KEY` from environment, it falls back to subscription
- Provides programmatic access without API credits

### Agent Architecture

**Single unified agent approach:**

1. **Agent**: `.claude/agents/commit-message-generator.md`

   - Single agent generates complete commit message
   - Uses `commit-message-format` skill for output compliance
   - Receives unified context (all modules, files, git diff)

2. **Skill**: `.claude/skills/commit-message-format.md`

   - Encapsulates `contracts/commit-message/0.1.0/structure.yml` rules
   - Enforces output format, line wrapping, anti-corruption layer
   - Reusable across multiple agents if needed

3. **Flow**:

  ```text
  commit-ai.go
    ↓ gathers context
    ↓ builds unified prompt
    ↓ shells out to `claude` CLI
  commit-message-generator agent
    ↓ activates commit-message-format skill
    ↓ analyzes git diff and modules
    ↓ generates complete message
  commit-ai.go
    ↓ receives output via stdout
    ↓ strips noise (anti-corruption)
    ↓ validates against contract
    ↓ outputs final message
  ```

### Key Implementation Details

**Simplified from multi-agent (before):**

- Old: `commit-message-top-level` agent + N × `commit-message-module` agents
- Old: Complex orchestration, combining sections, multiple Claude invocations
- Old: ~400 lines of orchestration code

**To single-agent (current):**

- New: Single `commit-message-generator` agent
- New: Agent uses skill for contract compliance
- New: One Claude invocation per commit
- New: ~250 lines (removed 150 lines of complexity)

**Trade-offs accepted:**

- ✅ Works with subscription (no API key needed)
- ✅ Programmatic from Go binary
- ✅ Single prompt → structured output
- ❌ Requires noise filtering (`stripAgentNoise`, `extractContentBlock`)
- ❌ Separate Claude session per invocation (no caching between commits)
- ❌ More complex than direct API usage would be

## Alternative Approaches Considered

### Option 1: Claude Code Task Tool

**Rejected**: Requires active Claude Code session

- User would need to be IN a `claude` session
- Would invoke via Task tool to subagent
- Not suitable for standalone CLI tool usage

### Option 2: MCP Server

**Rejected**: Still requires Claude Code session

- Would expose `generate_commit_message` tool
- Claude Code could call during interactive session
- Not suitable for programmatic invocation from Go

### Option 3: Direct API

**Blocked**: No API key available

- Would be ideal: simple HTTP requests from Go
- Much simpler code (~50 lines vs ~250 lines)
- Better rate limits and control
- **Requires API access** - not available with subscription-only

## Future Improvements

### When API Access Available

If/when Anthropic API key becomes available:

```go
// Replace Claude CLI invocation with direct API call
client := anthropic.NewClient(apiKey)
response, err := client.Messages.Create(ctx, &anthropic.MessageCreateParams{
    Model: "claude-sonnet-4",
    Messages: []anthropic.Message{
        {Role: "user", Content: fullPrompt},
    },
})
```

**Benefits:**

- Remove ~150 lines of CLI orchestration
- No noise filtering needed
- Better error handling
- Rate limit control
- Prompt caching between calls

### Current Optimization Opportunities

Within subscription-only constraint:

1. **Strengthen skill output contract** - Further reduce noise in responses
2. **Optimize context size** - Reduce token usage per invocation
3. **Better error recovery** - Handle Claude CLI failures gracefully

## Related Files

- Implementation: `src/commands/commit-ai.go`
- Agent: `.claude/agents/commit-message-generator.md`
- Skill: `.claude/skills/commit-message-format.md`
- Contract: `contracts/commit-message/0.1.0/structure.yml`
- Example: `contracts/commit-message/0.1.0/example.md`

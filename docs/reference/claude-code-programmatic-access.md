# Claude Code Programmatic Access

{{ page_breadcrumb() }}

How to use a Claude CLI/Max subscription to make API-like calls to Claude programmatically.

---

## Overview

With a Claude Max subscription, you can make programmatic API-like calls without needing a separate API key. The Claude Code CLI and SDKs authenticate using your existing subscription credentials.

---

## Methods

### 1. CLI Print Mode (Simplest)

Use `claude -p` for non-interactive, scriptable output:

```bash
# Basic query
claude -p "What is 2 + 2?"

# JSON output for parsing
claude -p --output-format json "Explain this code"

# Streaming JSON for real-time processing
claude -p --output-format stream-json "Generate a function"

# Pipe into your pipeline
claude -p "analyze this" --output-format json | jq '.result'
```

**Output format options** (only work with `--print`):

- `text` - Human readable (default)
- `json` - Single JSON result with metadata
- `stream-json` - Streaming JSON messages as they arrive

### 2. Python SDK

Install:

```bash
pip install claude-agent-sdk
```

Usage:

```python
import anyio
from claude_agent_sdk import query, ClaudeAgentOptions

async def main():
    # Simple query
    async for message in query(prompt="What is 2 + 2?"):
        print(message)

    # With options
    options = ClaudeAgentOptions(
        system_prompt="You are a helpful assistant",
        max_turns=1,
        allowed_tools=["Read", "Write", "Bash"],
        cwd="/path/to/project"
    )
    async for message in query(prompt="Your task", options=options):
        print(message)

anyio.run(main)
```

**Requirements:** Python 3.10 or later

### 3. TypeScript/Node.js SDK

Install:

```bash
npm install @anthropic-ai/claude-agent-sdk
```

### 4. Vercel AI SDK Provider

For web applications using the Vercel AI SDK:

```bash
# For v5 (recommended)
npm install ai-sdk-provider-claude-code ai

# For v4 (legacy)
npm install ai-sdk-provider-claude-code@ai-sdk-v4 ai@^4.3.16
```

This allows using your Max subscription through the Claude Code CLI in web apps.

---

## Important Notes

### Authentication

- **No separate API key needed** - Uses your existing Max subscription auth
- **Remove `ANTHROPIC_API_KEY`** from your environment if set, otherwise Claude Code will charge the API instead of using your subscription

### Usage Limits

Usage counts against your Max subscription limits:

| Plan | Messages per 5 hours | Claude Code prompts per 5 hours |
|------|---------------------|--------------------------------|
| Max 5x ($100/mo) | ~225 | ~50-200 |
| Max 20x ($200/mo) | ~900 | ~200-800 |

### Headless Mode

Claude Code includes headless mode for non-interactive contexts:

- CI/CD pipelines
- Pre-commit hooks
- Build scripts
- Automation workflows

Use `-p` flag with a prompt to enable headless mode.

---

## Sources

- [Claude Code CLI Reference](https://code.claude.com/docs/en/cli-reference)
- [Agent SDK Overview](https://platform.claude.com/docs/en/agent-sdk/overview)
- [Claude Agent SDK Python](https://github.com/anthropics/claude-agent-sdk-python)
- [Using Claude Code with Pro/Max](https://support.claude.com/en/articles/11145838-using-claude-code-with-your-pro-or-max-plan)
- [AI SDK Provider for Claude Code](https://github.com/ben-vargas/ai-sdk-provider-claude-code)
- [Output Format in Claude Code](https://claudelog.com/faqs/what-is-output-format-in-claude-code/)

{{ diataxis_footer() }}

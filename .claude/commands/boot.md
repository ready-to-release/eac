# Boot

```text
description: "Initialize session with project context from agent.md"
```

Initialize the development session by reading and following the **Session Initialization** instructions in `/agent.md`.

## Instructions

1. Read `/agent.md`
2. Follow the **Session Initialization** section exactly as specified
3. Provide the required initialization report
4. **Cache session context** for agent reuse (see below)

All initialization steps, requirements, and success criteria are defined in `/agent.md`.

## Session Context Caching

After completing initialization, create a shared session context file to avoid redundant MCP discovery by agents.

**What to Cache**: Information determined during initialization (from `<env>` and `gitStatus`):
- Current branch (from environment)
- Working directory (from environment)
- MCP server availability (from tool list inspection)
- Execution mode (MCP-First or CLI Fallback)

**What NOT to Call**: Do NOT call data-heavy MCP commands during initialization:
- ❌ `get-modules` (returns ~50+ modules, heavy data)
- ❌ `get-dependencies` (returns full dependency graph)
- ❌ `get-files` (returns ~2690 files)

**Create** `out/claude/session-context.json` with:

```json
{
  "timestamp": "ISO-8601 timestamp",
  "branch": "current branch name",
  "workdir": "current working directory",
  "project": {
    "name": "eac",
    "type": "go-modular-monorepo",
    "go_version": ">=1.21"
  },
  "available_tools": {
    "commands_mcp": true/false,
    "github_mcp": true/false
  },
  "execution_mode": "MCP-First|CLI Fallback",
  "context_expires_at": "timestamp + 5 minutes"
}
```

**Purpose**: Agents check this file to determine:
- Which MCP server to use (commands or GitHub)
- Whether to use MCP tools or CLI fallbacks
- Current branch context for their work

**Cache Lifetime**: 5 minutes (agents regenerate if stale or missing).

**Performance Impact**:
- Saves 5-10 seconds per agent startup (avoids redundant tool availability checks)
- Provides consistent execution mode across all agents in session

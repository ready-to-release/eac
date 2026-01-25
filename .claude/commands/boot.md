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

After completing initialization, create a shared session context file to avoid redundant discovery by agents.

**Create** `out/session-context.json` with:

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

**Purpose**: This context is consumed by all agents to skip redundant MCP discovery calls. Agents check this file first before calling expensive operations like `get-modules` or `get-dependencies`.

**Cache Lifetime**: 5 minutes (agents regenerate if stale or missing).

**Benefits**:
- Reduces agent startup time by 5-10 seconds
- Consistent view of project state across all agents
- Foundation for future multi-agent coordination

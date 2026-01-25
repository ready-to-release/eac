# Boot

```text
description: "Initialize session with project context"
```

Initialize the development session with project context and create a shared session cache for agent reuse.

## Instructions

1. Read `/agent.md` (if exists) for project-specific initialization
2. Follow any **Session Initialization** instructions in that file
3. Provide initialization report
4. **Cache session context** for agent reuse (see below)

## Session Context Caching

After completing initialization, create a shared session context file to avoid redundant discovery by agents.

**Create** `out/session-context.json` with:

```json
{
  "timestamp": "ISO-8601 timestamp",
  "branch": "current branch name",
  "workdir": "current working directory",
  "project": {
    "name": "project-name",
    "type": "detected project type"
  },
  "available_tools": {
    "commands_mcp": true/false,
    "github_mcp": true/false
  },
  "execution_mode": "MCP-First|CLI Fallback",
  "context_expires_at": "timestamp + 5 minutes"
}
```

**Purpose**: This context is consumed by all agents to skip redundant MCP discovery calls. Agents check this file first before calling expensive operations.

**Cache Lifetime**: 5 minutes (agents regenerate if stale or missing).

**Benefits**:
- Reduces agent startup time by 5-10 seconds
- Consistent view of project state across all agents
- Foundation for multi-agent coordination

## Example Initialization Report

```text
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⚡ SYSTEM INITIALIZED ⚡                                      ┃
┃  Project context loaded                                       ┃
┃  MCP servers: [✅ CONNECTED / ⚠️ NOT CONNECTED]               ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

Workspace Context:
- Current branch: [branch name]
- Working directory: [current path]

Project Context Loaded:
- Project type: [detected type]
- Available MCP tools: [list]
- Execution Mode: [MCP-First / CLI Fallback]

Session context cached to: out/session-context.json
Cache expires: [timestamp + 5 minutes]

Ready to assist with project tasks.
```

## Notes

- Customize the initialization report based on your project's needs
- The `out/session-context.json` format can be extended with project-specific metadata
- Agents should always check cache validity (timestamp) before using cached data
- If no `/agent.md` file exists, perform basic initialization and caching

# Status

```text
description: "Probe MCP servers, check worktree, load prior session summary"
```

Run at the start of a session to confirm environment state. Project context is
already loaded automatically from `CLAUDE.md` — this command handles dynamic
checks only.

## Process

1. **Probe MCP availability** (run both in parallel):
   - `ToolSearch` query: `mcp__commands__show-valid-commands`
   - `ToolSearch` query: `mcp__github__get_me`
   - Report: CONNECTED or NOT CONNECTED for each server
   - Set execution mode: MCP-First if commands connected, CLI Fallback otherwise

2. **Check git worktree**:
   - Current branch and working directory from environment context
   - Flag any mismatch between directory path and expected branch name

3. **Load prior session summary** (if exists):
   - List files matching `out/session-summary-*.md`
   - Read the most recent one (from last 7 days)
   - Summarize: what was in progress, any known TODOs

4. **Report status**:

```text
Session Status
==============
MCP commands: [CONNECTED | NOT CONNECTED — fallback: go run ./go/cli/eac]
MCP github:   [CONNECTED | NOT CONNECTED — fallback: gh CLI]
Branch:       <branch-name>
Worktree:     <path>  [✓ Match / ⚠ MISMATCH]
Prior session: <one-line summary or "none found">
```

## What NOT to call

Do NOT call data-heavy MCP commands during status:
- `get-modules` (~50+ modules)
- `get-dependencies` (full dependency graph)
- `get-files` (~2690 files)

These are loaded on-demand by agents when actually needed for a task.

For MCP server troubleshooting details, see `agent.md` → MCP Server Configuration.

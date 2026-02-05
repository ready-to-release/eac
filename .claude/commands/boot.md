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

## Performance Note

**What NOT to Call**: Do NOT call data-heavy MCP commands during initialization:
- ❌ `get-modules` (returns ~50+ modules, heavy data)
- ❌ `get-dependencies` (returns full dependency graph)
- ❌ `get-files` (returns ~2690 files)

These commands should be called on-demand by agents when they actually need the information.

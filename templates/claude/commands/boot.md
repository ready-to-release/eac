# Boot

```text
description: "Initialize session with project context"
```

Initialize the development session with project context and create a shared session cache for agent reuse.

## Instructions

1. Read `/agent.md` (if exists) for project-specific initialization
2. Follow any **Session Initialization** instructions in that file
3. Provide initialization report
4. **Complete initialization report**

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

Ready to assist with project tasks.
```

## Notes

- Customize the initialization report based on your project's needs
- If no `/agent.md` file exists, perform basic initialization

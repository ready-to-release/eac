# Boot

```text
description: "Initialize session with project context from CLAUDE.md"
```

Initialize the development session by loading project context and MCP server capabilities.

## Initialization Steps

1. **Read `/CLAUDE.md`** - Load the main project instructions
2. **Verify MCP servers** - Confirm available `mcp__commands__*` tools
3. **Internalize constraints and guidelines** - Apply all rules to the session
4. **Provide initialization report** - Confirm readiness

## Required Actions

### 1. Load Project Context

Read and internalize the `/CLAUDE.md` file, which includes:

- Project constraints (git operations, file organization)
- Development workflow (Specs → TDD → Validation)
- The Three Rules of Vibe Coding
- Go-specific coding rules
- Required output format
- MCP commands usage guidelines

### 2. Verify MCP Server Capabilities

Confirm that the following MCP command categories are available:

- Module Discovery (`get-modules`, `show-modules`, `show-moduletypes`, etc.)
- Dependency Management (`get-dependencies`, `show-dependencies`, etc.)
- Architecture Documentation (`design-*`)
- Build & Test Operations (`build-module`, `test-module`, etc.)
- Git Operations (`commit-ai`, `show-files-changed`, etc.)
- Templates (`templates-list`, `templates-install`)

### 3. Provide Initialization Report

After loading all context, provide a structured report:

```text
┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⚡ SYSTEM INITIALIZED ⚡                                      ┃
┃  Project context loaded from CLAUDE.md                        ┃
┃  MCP servers: [Status]                                        ┃
┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛

Project Context Loaded:

Active Constraints:
- [Key git constraints]
- [Key file organization constraints]
- [Key MCP usage requirements]

Development Workflow:
- [Specs → TDD → Validation workflow summary]

Vibe Coding Rules:
- [Three Rules summary]

MCP Server Status:
- Module Discovery: [✓/✗]
- Dependency Management: [✓/✗]
- Architecture Docs: [✓/✗]
- Build & Test: [✓/✗]
- Git Operations: [✓/✗]

Ready to assist with project tasks.
```

## Success Criteria

Initialization is complete when:

- ✓ CLAUDE.md has been read and internalized
- ✓ MCP command tools are recognized and available
- ✓ All constraints and guidelines are active
- ✓ Initialization report has been provided
- ✓ Flashy initialization message displayed (per CLAUDE.md requirement)

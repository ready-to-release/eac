# templates install claude

<!-- book:cmd templates install claude -->

## Details

- **Source**: `templates/claude/`
- **Destination**: `.claude/`

## Use Case

Install Claude Code configuration templates for setting up agents, commands, and skills in your repository. These templates demonstrate MCP command usage and provide a starting point for customizing Claude Code workflows.

**What Gets Installed**:

**Agents** (`.claude/agents/`):

- `architect.md` - Generic architecture agent using MCP commands
- `debugger.md` - Generic debugging agent using MCP commands
- `test-engineer.md` - Generic testing agent using MCP commands

**Commands** (`.claude/commands/`):

- `plan.md` - Planning workflow with MCP discovery
- `implement.md` - Implementation workflow with MCP verification
- `test.md` - Testing workflow with MCP commands
- `review.md` - Review workflow with MCP validation

**Skills** (`.claude/skills/`):

- `feature-workflow.md` - End-to-end feature development with MCP
- `refactor-safe.md` - Safe refactoring workflow with MCP validation

**Setup** (`.claude/setup/`):

- `mcp-setup.md` - MCP server configuration guide
- `.mcp.json.template` - MCP configuration template

## Notes

- Templates are language-agnostic and demonstrate MCP command usage
- Existing EAC-specific templates (go-architect, go-plan, etc.) are preserved
- Templates include placeholder values for customization

## See Also

- [templates Commands](./index.md)
- [templates install-ai](./install-ai.md) - AI prompt templates

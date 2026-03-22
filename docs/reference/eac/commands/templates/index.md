# templates Commands

Install project templates for documentation, AI prompts, reports, specifications, and Claude Code configurations.

## Commands in this Category

| Command                                           | Purpose                                  |
| ------------------------------------------------- | ---------------------------------------- |
| [templates install-docs](./install-docs.md)       | Install documentation templates          |
| [templates install-ai](./install-ai.md)           | Install AI prompt templates              |
| [templates install-reports](./install-reports.md) | Install report templates                 |
| [templates install-specs](./install-specs.md)     | Install specification templates          |
| [templates install-claude](./install-claude.md)   | Install Claude Code workflow templates   |

## Quick Examples

```bash
# Install documentation templates
eac templates install docs

# Install to custom location
eac templates install docs --destination ./custom-docs

# Install AI templates
eac templates install ai

# Install Claude Code templates
eac templates install claude

# Install with debug logging
eac templates install reports --debug
```

## Available Template Types

### docs

Documentation templates for project documentation.

### ai

AI prompt templates for structured AI interactions.

### reports

Report templates for test results and build outputs.

### specs

Specification templates for Gherkin/BDD tests.

### claude

Claude Code configuration templates (agents, commands, skills).

**What Gets Installed** (to `.claude/` directory):

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

**Note**: Existing EAC-specific templates (go-architect, go-plan, etc.) are preserved. Generic templates are language-agnostic and demonstrate MCP command usage.

## See Also

- [Category Overview](../../categories/templates.md)
- [create spec](../create/spec.md)

# Init Command

{{ page_breadcrumb() }}

## Purpose

The `init` command initializes AI provider configuration for the EAC project. It creates the necessary configuration file to enable AI-powered features such as automated commit messages, specification generation, architecture design, and pull request descriptions.

## Quick Reference

```bash
# Initialize with Claude API (Anthropic)
eac init --ai claude-api

# Initialize with Claude CLI (Claude Pro subscription)
eac init --ai claude-cli

# Initialize with OpenAI
eac init --ai openai

# Initialize with Google Gemini
eac init --ai gemini

# Initialize with debug output
eac init --ai claude-api --debug
```

## Command Syntax

```text
init --ai <provider> [flags]
```

### Flags

| Flag      | Shorthand | Type   | Required | Description                                                           |
| --------- | --------- | ------ | -------- | --------------------------------------------------------------------- |
| `--ai`    | `-a`      | string | **Yes**  | AI provider to use: `claude-api`, `claude-cli`, `openai`, or `gemini` |
| `--debug` | `-d`      | bool   | No       | Enable debug mode to show detailed configuration output               |

## Usage Examples

### Example 1: Initialize with Claude API

```bash
# Set up environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Initialize
eac init --ai claude-api
```

**Output:**

```text
Initializing AI provider configuration...
✓ Created configuration file: .r2r/eac/eac-config.yml
✓ Provider: claude-api
✓ API key: Found in environment variable ANTHROPIC_API_KEY

Next steps:
  1. Verify API key is set: echo $ANTHROPIC_API_KEY
  2. Test with: eac work commit --all
  3. Generate specs: eac create spec "your feature description"
```

### Example 2: Initialize with Claude CLI

```bash
# No API key needed - uses Claude Pro subscription
eac init --ai claude-cli
```

**Output:**

```text
Initializing AI provider configuration...
✓ Created configuration file: .r2r/eac/eac-config.yml
✓ Provider: claude-cli
✓ Using Claude Pro subscription (no API key required)

Next steps:
  1. Ensure you're logged into Claude CLI
  2. Test with: eac work commit --all
  3. Generate specs: eac create spec "your feature description"
```

### Example 3: Initialize with OpenAI

```bash
# Set up environment variable
export OPENAI_API_KEY="sk-proj-..."

# Initialize
eac init --ai openai
```

**Output:**

```text
Initializing AI provider configuration...
✓ Created configuration file: .r2r/eac/eac-config.yml
✓ Provider: openai
✓ API key: Found in environment variable OPENAI_API_KEY

Next steps:
  1. Verify API key is set: echo $OPENAI_API_KEY
  2. Test with: eac work commit --all
  3. Generate specs: eac create spec "your feature description"
```

## Next Steps After Initialization

Once you've initialized your AI provider configuration, you can use AI-powered features:

### 1. AI-Powered Commits

```bash
# Stage your changes
git add .

# Generate commit message with AI
eac work commit --all
```

The AI analyzes your changes and generates semantic commit messages following project conventions.

### 2. Specification Generation

```bash
# Generate Gherkin specifications from natural language
eac create spec "User authentication with JWT tokens"
```

The AI creates properly formatted Gherkin specifications with Feature, Rule, and Scenario blocks.

### 3. Architecture Design

```bash
# Generate architecture diagrams
eac create design src-auth
```

The AI creates Structurizr DSL diagrams for your modules.

### 4. Pull Request Descriptions

```bash
# Create PR with AI-generated description
eac work pr
```

The AI generates comprehensive PR titles, summaries, and test plans.

## Common Use Cases

### First-Time Setup

```bash
# 1. Get your API key from provider website
# 2. Set environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# 3. Initialize EAC
eac init --ai claude-api

# 4. Test the configuration
eac work commit --all
```

### Switching Providers

```bash
# Switch from OpenAI to Claude
export ANTHROPIC_API_KEY="sk-ant-api03-..."
eac init --ai claude-api

# Configuration file is updated automatically
```

### Team Setup

```bash
# Each team member initializes their own configuration
# 1. Share provider choice (e.g., "we use claude-api")
# 2. Each member gets their own API key
# 3. Each member runs init command

# Developer A
export ANTHROPIC_API_KEY="sk-ant-api03-devA-key..."
eac init --ai claude-api

# Developer B
export ANTHROPIC_API_KEY="sk-ant-api03-devB-key..."
eac init --ai claude-api
```

## Related Resources

- [Init Command Reference](../../../reference/commands/init-reference.md) - Complete technical details, provider comparison, and configuration options
- [Init Security Guide](init-security.md) - API key management and security best practices
- [Workspace Commands](areas/workspace-overview.md) - Workspace management with AI commits
- [Specifications Commands](areas/specifications-overview.md) - AI-powered specification generation
- [Design Commands](areas/design-overview.md) - Architecture diagram generation
- [Tutorials](../../../tutorials/index.md) - Introduction to EAC
- [Claude MCP Setup](../configuration/claude-mcp-setup.md) - MCP server configuration

{{ diataxis_footer() }}

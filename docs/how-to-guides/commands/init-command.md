# Init Command

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

## AI Provider Comparison

| Provider       | API Key Required               | Cost                       | Best For                     | Model             |
| -------------- | ------------------------------ | -------------------------- | ---------------------------- | ----------------- |
| **claude-api** | `ANTHROPIC_API_KEY`            | Pay-per-use                | Production use, high volume  | Claude Sonnet 4.5 |
| **claude-cli** | None (Claude Pro subscription) | Included with subscription | Development, lower volume    | Claude Sonnet 4.5 |
| **openai**     | `OPENAI_API_KEY`               | Pay-per-use                | OpenAI ecosystem integration | GPT-4             |
| **gemini**     | `GOOGLE_API_KEY`               | Pay-per-use                | Google Cloud integration     | Gemini Pro        |

### Provider Selection Guide

**Choose `claude-api` when:**

- You need production-grade reliability
- You have high-volume AI operations
- You want the latest Claude models
- You're willing to pay per API call

**Choose `claude-cli` when:**

- You have a Claude Pro subscription
- You're doing development work
- You have moderate AI usage needs
- You want to avoid API costs

**Choose `openai` when:**

- You're already using OpenAI services
- You prefer GPT-4 models
- You have existing OpenAI infrastructure

**Choose `gemini` when:**

- You're using Google Cloud Platform
- You prefer Google's AI models
- You have Google Cloud credits

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

### Example 4: Initialize with Gemini

```bash
# Set up environment variable
export GOOGLE_API_KEY="AIza..."

# Initialize
eac init --ai gemini
```

**Output:**

```text
Initializing AI provider configuration...
✓ Created configuration file: .r2r/eac/eac-config.yml
✓ Provider: gemini
✓ API key: Found in environment variable GOOGLE_API_KEY

Next steps:
  1. Verify API key is set: echo $GOOGLE_API_KEY
  2. Test with: eac work commit --all
  3. Generate specs: eac create spec "your feature description"
```

### Example 5: Debug Mode

```bash
eac init --ai claude-api --debug
```

**Output:**

```text
[DEBUG] Initializing AI provider configuration...
[DEBUG] Provider: claude-api
[DEBUG] Config path: C:\projects\eac\.r2r\eac\eac-config.yml
[DEBUG] Checking for existing configuration...
[DEBUG] No existing configuration found
[DEBUG] Creating new configuration file...
[DEBUG] Environment variable ANTHROPIC_API_KEY: Found
[DEBUG] Writing configuration to file...
✓ Created configuration file: .r2r/eac/eac-config.yml
✓ Provider: claude-api
✓ API key: Found in environment variable ANTHROPIC_API_KEY

Configuration contents:
---
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
---
```

## Configuration File Format

The init command creates `.r2r/eac/eac-config.yml` with the following structure:

### Claude API Configuration

```yaml
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
max_tokens: 4096
temperature: 0.7
```

### Claude CLI Configuration

```yaml
provider: claude-cli
model: claude-sonnet-4-5
```

### OpenAI Configuration

```yaml
provider: openai
api_key_env: OPENAI_API_KEY
model: gpt-4
max_tokens: 4096
temperature: 0.7
```

### Gemini Configuration

```yaml
provider: gemini
api_key_env: GOOGLE_API_KEY
model: gemini-pro
max_tokens: 4096
temperature: 0.7
```

## Environment Variable Setup

### Windows (PowerShell)

```powershell
# Temporary (current session only)
$env:ANTHROPIC_API_KEY = "sk-ant-api03-..."
$env:OPENAI_API_KEY = "sk-proj-..."
$env:GOOGLE_API_KEY = "AIza..."

# Permanent (user environment variable)
[System.Environment]::SetEnvironmentVariable('ANTHROPIC_API_KEY', 'sk-ant-api03-...', 'User')
[System.Environment]::SetEnvironmentVariable('OPENAI_API_KEY', 'sk-proj-...', 'User')
[System.Environment]::SetEnvironmentVariable('GOOGLE_API_KEY', 'AIza...', 'User')
```

### Windows (Command Prompt)

```cmd
# Temporary (current session only)
set ANTHROPIC_API_KEY=sk-ant-api03-...
set OPENAI_API_KEY=sk-proj-...
set GOOGLE_API_KEY=AIza...

# Permanent (requires admin)
setx ANTHROPIC_API_KEY "sk-ant-api03-..."
setx OPENAI_API_KEY "sk-proj-..."
setx GOOGLE_API_KEY "AIza..."
```

### Linux/macOS (Bash/Zsh)

```bash
# Temporary (current session only)
export ANTHROPIC_API_KEY="sk-ant-api03-..."
export OPENAI_API_KEY="sk-proj-..."
export GOOGLE_API_KEY="AIza..."

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export ANTHROPIC_API_KEY="sk-ant-api03-..."' >> ~/.bashrc
echo 'export OPENAI_API_KEY="sk-proj-..."' >> ~/.bashrc
echo 'export GOOGLE_API_KEY="AIza..."' >> ~/.bashrc

# Reload configuration
source ~/.bashrc
```

### Verify Environment Variables

```bash
# Check if variable is set
echo $ANTHROPIC_API_KEY
echo $OPENAI_API_KEY
echo $GOOGLE_API_KEY

# Should display your API key
# If empty, the variable is not set
```

## Getting API Keys

### Anthropic API Key (claude-api)

1. Visit [console.anthropic.com](https://console.anthropic.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Click "Create Key"
5. Copy the key (starts with `sk-ant-api03-`)
6. Set environment variable: `ANTHROPIC_API_KEY`

**Pricing:** Pay-per-use based on tokens consumed. See [Anthropic Pricing](https://www.anthropic.com/pricing).

### Claude Pro Subscription (claude-cli)

1. Visit [claude.ai](https://claude.ai/)
2. Subscribe to Claude Pro
3. Install Claude CLI (if needed)
4. Log in to Claude CLI
5. No API key required

**Pricing:** Fixed monthly subscription fee.

### OpenAI API Key (openai)

1. Visit [platform.openai.com](https://platform.openai.com/)
2. Sign up or log in
3. Navigate to API Keys section
4. Click "Create new secret key"
5. Copy the key (starts with `sk-proj-` or `sk-`)
6. Set environment variable: `OPENAI_API_KEY`

**Pricing:** Pay-per-use based on tokens consumed. See [OpenAI Pricing](https://openai.com/pricing).

### Google API Key (gemini)

1. Visit [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Sign in with Google account
3. Click "Create API Key"
4. Copy the key (starts with `AIza`)
5. Set environment variable: `GOOGLE_API_KEY`

**Pricing:** Free tier available, then pay-per-use. See [Gemini Pricing](https://ai.google.dev/pricing).

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

### CI/CD Setup

```bash
# Store API key in CI/CD secrets
# GitHub Actions example:
# Add ANTHROPIC_API_KEY to repository secrets

# In your workflow:
- name: Initialize AI provider
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: eac init --ai claude-api
```

## Best Practices

### Security

- **Never commit API keys** to version control
- **Add `.r2r/` to `.gitignore`** to exclude configuration files
- **Use environment variables** instead of hardcoding keys
- **Rotate keys regularly** for security
- **Use separate keys** for development and production

### API Key Management

```bash
# Add to .gitignore
echo ".r2r/eac/eac-config.yml" >> .gitignore

# Verify it's not tracked
git status .r2r/

# Use different keys for different environments
export ANTHROPIC_API_KEY_DEV="sk-ant-api03-dev-..."
export ANTHROPIC_API_KEY_PROD="sk-ant-api03-prod-..."
```

### Cost Optimization

**For claude-cli:**

- Free usage included with Claude Pro subscription
- Rate limits apply
- Best for development work

**For claude-api, openai, gemini:**

- Monitor usage through provider dashboard
- Set spending limits
- Use caching when available
- Start with smaller models for testing

### Provider Selection

- **Development:** Use `claude-cli` if you have Claude Pro
- **Production/CI/CD:** Use `claude-api`, `openai`, or `gemini`
- **Team consistency:** Agree on one provider for the team
- **Experimentation:** Try different providers with `--debug` flag

## Troubleshooting

### Problem: API Key Not Found

**Error:**

```text
Error: ANTHROPIC_API_KEY environment variable not set
```

**Solution:**

```bash
# Verify variable is set
echo $ANTHROPIC_API_KEY

# If empty, set it
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# Try again
eac init --ai claude-api
```

### Problem: Invalid API Key

**Error:**

```text
Error: Authentication failed - invalid API key
```

**Solution:**

1. Verify you copied the complete key (no spaces)
2. Check key hasn't expired
3. Confirm key is for the correct provider
4. Generate a new key from provider dashboard

### Problem: Configuration File Already Exists

**Error:**

```text
Warning: Configuration file already exists at .r2r/eac/eac-config.yml
```

**Solution:**

```bash
# The command will overwrite the existing file
# Your choice:
# 1. Proceed (overwrites old config)
eac init --ai claude-api

# 2. Back up old config first
cp .r2r/eac/eac-config.yml .r2r/eac/eac-config.yml.backup
eac init --ai claude-api
```

### Problem: Permission Denied

**Error:**

```text
Error: Permission denied: .r2r/eac/eac-config.yml
```

**Solution:**

```bash
# Check directory permissions
ls -la .r2r/eac/

# Fix permissions
chmod 755 .r2r/eac
chmod 644 .r2r/eac/eac-config.yml

# Try again
eac init --ai claude-api
```

### Problem: Directory Not Found

**Error:**

```text
Error: Directory not found: .r2r/eac/
```

**Solution:**

```bash
# Create directory
mkdir -p .r2r/eac

# Try again
eac init --ai claude-api
```

### Problem: Provider Not Recognized

**Error:**

```text
Error: Unknown provider: claude
```

**Solution:**

```bash
# Use exact provider name
# Valid options: claude-api, claude-cli, openai, gemini
eac init --ai claude-api  # ✓ Correct
eac init --ai claude      # ✗ Invalid
```

### Problem: AI Features Not Working After Init

**Symptoms:**

- `eac work commit` doesn't generate messages
- `eac create spec` fails
- No AI responses

**Solution:**

```bash
# 1. Verify configuration exists
cat .r2r/eac/eac-config.yml

# 2. Check environment variable
echo $ANTHROPIC_API_KEY  # or appropriate variable

# 3. Test with debug mode
eac work commit --all --debug

# 4. Re-initialize if needed
eac init --ai claude-api --debug
```

## Related Commands

- `work commit` - Generate AI-powered commit messages
- `create spec` - Generate Gherkin specifications with AI
- `create design` - Generate architecture diagrams with AI
- `work pr` - Create pull requests with AI-generated descriptions
- `show config` - Display current configuration

## Configuration File Location

The configuration file is stored at:

```text
.r2r/eac/eac-config.yml
```

**Important:** This file should **NOT** be committed to version control as it may contain sensitive information or preferences specific to your environment.

### Recommended .gitignore Entry

```gitignore
# AI provider configuration (personal settings)
.r2r/eac/eac-config.yml
```

## Notes

- The init command can be run multiple times to change providers
- Each run overwrites the previous configuration
- Environment variables are checked during initialization
- API keys are never stored in the configuration file (only the environment variable name)
- The command validates the provider name but does not validate API keys
- Configuration is local to each repository
- Team members can use different providers if desired

## See Also

- [Workspace Commands](areas/workspace-overview.md) - Workspace management with AI commits
- [Specifications Commands](areas/specifications-overview.md) - AI-powered specification generation
- [Design Commands](areas/design-overview.md) - Architecture diagram generation
- [Tutorials](../../tutorials/index.md) - Introduction to EAC
- [Claude MCP Setup](../configuration/claude-mcp-setup.md) - MCP server configuration

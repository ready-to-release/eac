<!-- EDITOR
# Editor: reference/commands/init-reference.md

## Soul

Complete technical reference for the init command including provider comparison, all configuration options, environment variables, and technical implementation details.

## Sections

1. Overview
2. AI Provider Comparison
3. Provider Selection Guide
4. Configuration File Format
5. Environment Variable Setup
6. Configuration File Location
7. Troubleshooting
8. Technical Notes

-->

# Init Command Reference

## Overview

The `init` command creates the AI provider configuration file at `.r2r/eac/eac-config.yml`. This configuration enables AI-powered features throughout EAC including commit message generation, specification creation, architecture design, and pull request descriptions.

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

### Cost Optimization

**For claude-cli:**

- Free usage included with Claude Pro subscription
- Rate limits apply
- Best for development work

**For claude-api, openai, gemini:**

- Monitor usage through provider dashboard
- Set spending limits in provider console
- Use caching when available
- Start with smaller models for testing

## Configuration File Format

The init command creates `.r2r/eac/eac-config.yml` with provider-specific configuration.

### Claude API Configuration

```yaml
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
max_tokens: 4096
temperature: 0.7
```

**Configuration Options:**

- `provider`: Must be `claude-api`
- `api_key_env`: Environment variable name containing the API key
- `model`: Claude model to use (default: `claude-sonnet-4-5`)
- `max_tokens`: Maximum tokens in response (default: 4096)
- `temperature`: Randomness in responses, 0.0-1.0 (default: 0.7)

### Claude CLI Configuration

```yaml
provider: claude-cli
model: claude-sonnet-4-5
```

**Configuration Options:**

- `provider`: Must be `claude-cli`
- `model`: Claude model to use (default: `claude-sonnet-4-5`)

Note: No API key required. Uses Claude Pro subscription authentication.

### OpenAI Configuration

```yaml
provider: openai
api_key_env: OPENAI_API_KEY
model: gpt-4
max_tokens: 4096
temperature: 0.7
```

**Configuration Options:**

- `provider`: Must be `openai`
- `api_key_env`: Environment variable name containing the API key
- `model`: OpenAI model to use (default: `gpt-4`)
- `max_tokens`: Maximum tokens in response (default: 4096)
- `temperature`: Randomness in responses, 0.0-1.0 (default: 0.7)

### Gemini Configuration

```yaml
provider: gemini
api_key_env: GOOGLE_API_KEY
model: gemini-pro
max_tokens: 4096
temperature: 0.7
```

**Configuration Options:**

- `provider`: Must be `gemini`
- `api_key_env`: Environment variable name containing the API key
- `model`: Gemini model to use (default: `gemini-pro`)
- `max_tokens`: Maximum tokens in response (default: 4096)
- `temperature`: Randomness in responses, 0.0-1.0 (default: 0.7)

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

## Technical Notes

- The init command can be run multiple times to change providers
- Each run overwrites the previous configuration
- Environment variables are checked during initialization
- API keys are never stored in the configuration file (only the environment variable name)
- The command validates the provider name but does not validate API keys
- Configuration is local to each repository
- Team members can use different providers if desired
- The `.r2r/eac/` directory is created automatically if it doesn't exist

## Related Commands

- `work commit` - Generate AI-powered commit messages
- `create spec` - Generate Gherkin specifications with AI
- `create design` - Generate architecture diagrams with AI
- `work pr` - Create pull requests with AI-generated descriptions
- `show config` - Display current configuration

## See Also

- [Init Command Guide](../../how-to-guides/eac/commands/init-command.md) - Quick start and basic usage
- [Init Security Guide](../../how-to-guides/eac/commands/init-security.md) - Security best practices

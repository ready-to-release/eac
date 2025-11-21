# Init Command

**Problem**: Configure AI provider for use with AI-powered commands without exposing API keys in your codebase.

**Solution**: Use `init` to create safe configuration files that reference environment variables instead of storing secrets.

## Key Benefits

- Quick AI provider setup (Claude, OpenAI, Gemini)
- Safe configuration (no secrets in files)
- Supports both API and CLI-based providers
- Auto-copies AI contracts for customization

## Quick Start

```bash
# Claude Pro subscription (no API costs)
r2r eac init --ai claude-cli

# Claude API
r2r eac init --ai claude-api

# OpenAI
r2r eac init --ai openai

# Google Gemini
r2r eac init --ai gemini
```

Creates `.r2r/agent-config.yml` and `.r2r/contracts/` directory structure.

## What It Creates

### Directory Structure

```
.r2r/
├── agent-config.yml          # AI provider configuration
└── contracts/
    └── ai/                   # Customizable AI contracts
        ├── commit-ai/
        ├── specs-create/
        └── design-create/
```

### Configuration File

`.r2r/agent-config.yml` example:

```yaml
# SAFE TO COMMIT: Contains only environment variable references
provider:
  name: claude-api
  model: claude-3-5-haiku-20241022
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}
```

## Command Reference

```bash
r2r eac init --ai <provider>

# Required flag:
--ai, -a <provider>    # AI provider: claude-api, claude-cli, openai, gemini
```

### Provider Options

| Provider | Environment Variable | Notes |
|----------|---------------------|-------|
| `claude-cli` | None | Uses Claude Pro subscription, requires `claude` CLI |
| `claude-api` | `ANTHROPIC_API_KEY` | API access, get key at claude.ai/settings/api |
| `openai` | `OPENAI_API_KEY` | Get key at platform.openai.com/api-keys |
| `gemini` | `GOOGLE_API_KEY` | Get key at makersuite.google.com/app/apikey |

## Typical Workflow

### Initial Setup

```bash
# 1. Initialize with your preferred provider
r2r eac init --ai claude-cli

# 2. Set environment variable (if required)
export ANTHROPIC_API_KEY="your-api-key-here"
# Windows: setx ANTHROPIC_API_KEY "your-api-key-here"

# 3. Commit the config file
git add .r2r/agent-config.yml
git commit -m "chore: configure AI provider"

# 4. Use AI-powered commands
r2r eac specs create "feature description"
r2r eac work commit --all
r2r eac design create src-modulename
```

### Reconfiguring

```bash
# Switch providers (updates existing config)
r2r eac init --ai openai
```

Output:
```
⚠️  Project already initialized
   Config exists: .r2r/agent-config.yml

🔄 Reconfiguring agent configuration...
```

## Provider Details

### Claude CLI (Recommended for Pro Subscribers)

```bash
r2r eac init --ai claude-cli
```

**Benefits:**
- No API costs (uses subscription credits)
- No API key management
- Same quality as API access

**Requirements:**
- Claude Pro subscription
- `claude` CLI installed

**Setup:**
```bash
# Install Claude CLI
npm install -g @anthropic-ai/claude-cli
# or: pipx install claude-cli

# Authenticate
claude auth login
```

### Claude API

```bash
r2r eac init --ai claude-api
```

**Benefits:**
- Pay-as-you-go pricing
- No subscription required
- Works with workspace or personal API keys

**Setup:**
1. Get API key: https://claude.ai/settings/api
2. Set environment variable:
   ```bash
   export ANTHROPIC_API_KEY="sk-ant-..."
   ```

### OpenAI

```bash
r2r eac init --ai openai
```

**Setup:**
1. Get API key: https://platform.openai.com/api-keys
2. Set environment variable:
   ```bash
   export OPENAI_API_KEY="sk-..."
   ```

### Gemini

```bash
r2r eac init --ai gemini
```

**Setup:**
1. Get API key: https://makersuite.google.com/app/apikey
2. Set environment variable:
   ```bash
   export GOOGLE_API_KEY="..."
   ```

## AI Contracts

The init command copies AI contracts to `.r2r/contracts/ai/` for customization.

### Available Contracts

```
.r2r/contracts/ai/
├── commit-ai/              # AI commit message generation
│   ├── system-prompt.md
│   └── user-prompt.md
├── specs-create/           # Specification generation
│   ├── system-prompt.md
│   └── user-prompt.md
└── design-create/          # Architecture diagram generation
    ├── system-prompt.md
    └── user-prompt.md
```

### Customizing Contracts

Edit contract files to customize AI behavior:

```bash
# Customize commit message style
nano .r2r/contracts/ai/commit-ai/system-prompt.md

# Customize spec generation
nano .r2r/contracts/ai/specs-create/system-prompt.md
```

Changes take effect immediately for subsequent AI commands.

## Environment Variables

### Setting Variables

**Linux/macOS:**
```bash
# Temporary (current session)
export ANTHROPIC_API_KEY="sk-ant-..."

# Permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export ANTHROPIC_API_KEY="sk-ant-..."' >> ~/.bashrc
source ~/.bashrc
```

**Windows PowerShell:**
```powershell
# Temporary (current session)
$env:ANTHROPIC_API_KEY = "sk-ant-..."

# Permanent (user-level)
[Environment]::SetEnvironmentVariable("ANTHROPIC_API_KEY", "sk-ant-...", "User")
```

**Windows CMD:**
```cmd
# Permanent
setx ANTHROPIC_API_KEY "sk-ant-..."
```

### Verifying Setup

```bash
# Check environment variable is set
echo $ANTHROPIC_API_KEY        # Linux/macOS
echo %ANTHROPIC_API_KEY%       # Windows CMD
$env:ANTHROPIC_API_KEY         # Windows PowerShell

# Test with AI command
r2r eac specs create "simple test feature"
```

## Best Practices

- **Commit config files**: `.r2r/agent-config.yml` is safe to commit (no secrets)
- **Never commit API keys**: Always use environment variables
- **Use claude-cli for Pro users**: Avoid API costs
- **Customize contracts**: Tailor AI behavior to your project conventions
- **Version control contracts**: Commit customized contracts for team consistency

## Troubleshooting

| Problem | Solution |
|---------|----------|
| `--ai flag is required` | Provide provider: `init --ai claude-api` |
| `unsupported provider` | Use: `claude-api`, `claude-cli`, `openai`, or `gemini` |
| API commands fail | Check environment variable is set correctly |
| `claude: command not found` | Install Claude CLI: `npm install -g @anthropic-ai/claude-cli` |
| Permission denied creating `.r2r/` | Check directory permissions, may need sudo |

## Advanced Usage

### Multiple Projects

Each project has its own configuration:

```bash
cd ~/project-a
r2r init --ai claude-cli

cd ~/project-b
r2r init --ai openai
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Setup AI provider
  run: r2r init --ai claude-api
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

### Team Sharing

Share contract customizations with team:

```bash
git add .r2r/contracts/
git commit -m "docs: customize AI contract prompts"
git push
```

Team members pull and get same AI behavior.

## Summary

1. **Initialize**: `r2r eac init --ai <provider>`
2. **Set API key**: `export ANTHROPIC_API_KEY="..."`  (if needed)
3. **Commit config**: `git add .r2r/` and commit
4. **Use AI commands**: `specs create`, `work commit`, etc.
5. **Customize** (optional): Edit contracts in `.r2r/contracts/ai/`

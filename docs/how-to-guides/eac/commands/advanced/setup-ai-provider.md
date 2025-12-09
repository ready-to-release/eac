# Setup AI Provider

{{ page_breadcrumb() }}

## What You'll Accomplish

Configure AI provider (Anthropic Claude or OpenAI) to enable AI-powered features like commit message generation and PR descriptions.

## Prerequisites

- API key from Anthropic or OpenAI
- Environment variable support (or config file)

## Steps

### 1. Run Init Command

```bash
r2r eac init
```

**What happens**: Interactive wizard guides you through AI provider setup

### 2. Choose AI Provider

Select from:

- **Anthropic Claude** (recommended)
- **OpenAI**

### 3. Enter API Key

Provide your API key when prompted.

**What happens**: Key is stored in environment configuration

### 4. Test Configuration

```bash
r2r eac create commit-message --help
```

**What happens**: If configured correctly, AI features are available

## Configuration Options

### Using Environment Variables

```bash
# Anthropic
export ANTHROPIC_API_KEY="sk-ant-..."

# OpenAI
export OPENAI_API_KEY="sk-..."
```

### Using Config File

Create `.eac/config.yml`:

```yaml
ai:
  provider: anthropic
  api_key: ${ANTHROPIC_API_KEY}
```

## Example Scenario

Setting up Anthropic Claude:

```bash
# Run init
r2r eac init

# Choose provider
? Select AI provider: Anthropic Claude

# Enter key
? API Key: sk-ant-api03-...

# ✓ Configuration saved

# Test it
echo "test" > test.txt
git add test.txt
r2r eac create commit-message
# AI generates commit message successfully!
```

## Common Issues

| Problem | Solution |
|---------|----------|
| "API key invalid" | Verify key is correct and active |
| "Provider not configured" | Run `r2r eac init` again |
| Rate limits | Check API usage/billing |

## Next Steps

- [Make Commits with AI](../development-workflow/make-commits-with-ai.md) → Use AI commits
- [Understand AI Process](./understand-ai-commit-process.md) → How AI works

## Related Commands

- [`init`](../../../reference/commands/other/init.md) - Initialize AI configuration

{{ diataxis_footer() }}

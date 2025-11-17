# AI Provider Configuration Guide

## Overview

The CLI supports multiple AI providers for commands that require AI capabilities (like `commit-ai`). You can configure which provider to use, and the system automatically falls back to Claude CLI if the configured provider fails.

## Supported Providers

| Provider | Description | Requires API Key | Cost |
|----------|-------------|------------------|------|
| **claude-cli** | Uses Claude Pro subscription via CLI | No | Free with Claude Pro |
| **claude-api** | Direct Anthropic API calls | Yes (`ANTHROPIC_API_KEY`) | Pay per use |
| **openai** | OpenAI GPT models | Yes (`OPENAI_API_KEY`) | Pay per use |
| **gemini** | Google Gemini API | Yes (`GOOGLE_API_KEY`) | Pay per use |

## Quick Start

### Zero Configuration (Recommended for most users)

If you have a Claude Pro subscription, **no configuration is needed**. The CLI automatically uses `claude-cli` provider:

```bash
# Just run AI commands - they work immediately
r2r commit-ai

# Or use the ai command directly
r2r ai ask "Explain this error message"
```

### Configure a Different Provider

Create `.r2r/ai-config.yml` in your repository root:

```yaml
# Example: Use Claude API
provider: claude-api
apiKey: ${ANTHROPIC_API_KEY}
model: claude-3-haiku-20240307

# Optional settings
temperature: 0.3
maxTokens: 4000
```

## Configuration Examples

### Claude API (Anthropic)

```yaml
provider: claude-api
apiKey: ${ANTHROPIC_API_KEY}
model: claude-3-5-sonnet-20250514  # Or claude-3-haiku-20240307
temperature: 0.3
maxTokens: 4000
```

Set your API key:
```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

**Available models:**
- `claude-3-5-sonnet-20250514` - Most capable (recommended)
- `claude-3-haiku-20240307` - Fast and cost-effective
- `claude-3-opus-20240229` - Maximum capability

### OpenAI (GPT)

```yaml
provider: openai
apiKey: ${OPENAI_API_KEY}
model: gpt-4
temperature: 0.3
maxTokens: 4000
```

Set your API key:
```bash
export OPENAI_API_KEY="sk-..."
```

**Available models:**
- `gpt-4` - Most capable
- `gpt-4-turbo` - Fast and cost-effective
- `gpt-3.5-turbo` - Cheapest option

### Google Gemini

```yaml
provider: gemini
apiKey: ${GOOGLE_API_KEY}
model: gemini-pro
temperature: 0.3
maxTokens: 4000
```

Set your API key:
```bash
export GOOGLE_API_KEY="..."
```

**Available models:**
- `gemini-1.5-pro` - Most capable
- `gemini-pro` - Standard model

### Claude CLI (Default)

```yaml
provider: claude-cli
model: sonnet  # or haiku, opus
```

**No API key needed** - uses your Claude Pro subscription.

**Available models:**
- `sonnet` - Recommended (default)
- `haiku` - Faster, cheaper
- `opus` - Most capable

## Configuration Options

### Provider Settings

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `provider` | string | Yes | Provider name: `claude-cli`, `claude-api`, `openai`, or `gemini` |
| `apiKey` | string | Conditional | API key (use `${ENV_VAR}` syntax). Required for `claude-api`, `openai`, `gemini` |
| `model` | string | Yes | Model identifier (provider-specific) |
| `temperature` | float | No | Randomness (0.0-1.0). Lower = more deterministic. Default: 0.3 |
| `maxTokens` | integer | No | Maximum response length. Default: 4000 |

### Environment Variable Substitution

Use `${VARIABLE_NAME}` in your config to reference environment variables:

```yaml
provider: openai
apiKey: ${OPENAI_API_KEY}  # Replaced with env var value at runtime
model: gpt-4
```

**Security note:** Never commit API keys directly to your config file. Always use environment variables.

## Automatic Fallback

If the configured provider fails for any reason, the system automatically falls back to `claude-cli`:

```
Warning: Failed to create openai provider (OPENAI_API_KEY is required), using fallback
Falling back to claude-cli provider
```

**Fallback triggers:**
- Missing API key
- Invalid configuration
- Provider API error
- Network issues

This ensures AI commands always work if you have Claude Pro subscription.

## Usage Examples

### Using Different Providers

```bash
# Use default provider (from config or claude-cli)
r2r commit-ai

# Use specific model for one command (overrides config)
r2r ai ask "Summarize this code" --model gpt-4

# With custom temperature
r2r ai ask "Be creative!" --temperature 0.9
```

### Checking Active Provider

The CLI logs which provider is being used:

```bash
r2r commit-ai
# 🤖 Generating commit message using claude-api provider...
```

### Execution Logs

All AI executions are logged to `.r2r/logs/ai-executions.jsonl`:

```jsonl
{"timestamp":"2025-01-15T10:30:00Z","provider":"claude-api","model":"claude-3-haiku-20240307","success":true,"didFallback":false,"duration_ms":1234}
{"timestamp":"2025-01-15T10:35:00Z","provider":"claude-cli","model":"sonnet","success":true,"didFallback":true,"duration_ms":2345}
```

This helps track:
- Which provider was actually used
- Whether fallback occurred
- Performance metrics
- Error tracking

## Troubleshooting

### "API key is required" Error

```
Error: OPENAI_API_KEY is required for openai provider
```

**Solution:** Set the environment variable or update your config:

```bash
export OPENAI_API_KEY="sk-..."
```

### Provider Falls Back to claude-cli

Check:
1. Is your API key set? `echo $ANTHROPIC_API_KEY`
2. Is the key valid? Test with the provider's API
3. Is your config file valid? Check `.r2r/ai-config.yml` syntax
4. Check the logs in `.r2r/logs/ai-executions.jsonl`

### Config File Not Found

The config file is optional. Without it, the CLI uses `claude-cli` (requires Claude Pro).

To create a config:

```bash
mkdir -p .r2r
cat > .r2r/ai-config.yml << EOF
provider: claude-cli
model: sonnet
EOF
```

### "claude: command not found"

You need to install the Claude CLI:

```bash
# Install Claude CLI (requires Claude Pro subscription)
# See: https://claude.ai/download
```

Alternatively, configure a different provider that doesn't require the CLI:

```yaml
provider: openai
apiKey: ${OPENAI_API_KEY}
model: gpt-4
```

## Best Practices

### For Development

Use `claude-cli` (default) - zero configuration, no API costs:

```yaml
provider: claude-cli
model: haiku  # Fast and free with Pro subscription
```

### For CI/CD

Use API-based providers for reliability:

```yaml
provider: claude-api
apiKey: ${ANTHROPIC_API_KEY}
model: claude-3-haiku-20240307
temperature: 0.1  # More deterministic for CI
```

### For Cost Optimization

1. **Start with haiku models** - much cheaper, often sufficient
2. **Use Claude CLI** when possible - free with Pro subscription
3. **Set lower maxTokens** - only pay for what you need
4. **Monitor logs** - track actual usage in `ai-executions.jsonl`

### For Quality

1. **Use sonnet/opus models** - better reasoning and accuracy
2. **Lower temperature** (0.1-0.3) - more consistent results
3. **Increase maxTokens** if responses are cut off

## Security

### API Key Storage

**Never commit API keys to your repository:**

```bash
# Add to .gitignore
echo ".r2r/ai-config.yml" >> .gitignore
```

**Use environment variables:**

```yaml
# Good - uses environment variable
apiKey: ${ANTHROPIC_API_KEY}

# Bad - hardcoded key (DO NOT DO THIS!)
apiKey: sk-ant-api03-hardcoded-key
```

### API Key Permissions

Follow the principle of least privilege:
- Create separate API keys for different environments
- Use read-only keys when possible
- Rotate keys regularly
- Monitor usage for anomalies

### Local vs CI Configuration

**Local development:**
```bash
# .r2r/ai-config.yml (gitignored)
provider: claude-cli
model: haiku
```

**CI/CD:**
```bash
# .r2r/ai-config.ci.yml (committed, no secrets)
provider: claude-api
apiKey: ${CI_ANTHROPIC_API_KEY}
model: claude-3-haiku-20240307
```

## Advanced Usage

### Per-Command Provider Override

```bash
# Use environment variables to override config for one command
ANTHROPIC_API_KEY=sk-... r2r commit-ai

# Or create a temporary config
echo "provider: openai" > /tmp/ai-config.yml
r2r commit-ai --config /tmp/ai-config.yml
```

### Multiple Configurations

```bash
# Development config
cp .r2r/ai-config.yml .r2r/ai-config.dev.yml

# Production config
cp .r2r/ai-config.yml .r2r/ai-config.prod.yml

# Use specific config
export R2R_AI_CONFIG=.r2r/ai-config.prod.yml
r2r commit-ai
```

## Getting API Keys

### Anthropic (Claude)
1. Visit https://console.anthropic.com/
2. Create an account or sign in
3. Navigate to API Keys
4. Create a new key
5. Set `ANTHROPIC_API_KEY` environment variable

### OpenAI (GPT)
1. Visit https://platform.openai.com/
2. Sign in or create an account
3. Navigate to API Keys
4. Create a new secret key
5. Set `OPENAI_API_KEY` environment variable

### Google (Gemini)
1. Visit https://makersuite.google.com/
2. Sign in with Google account
3. Get your API key
4. Set `GOOGLE_API_KEY` environment variable

## Support

For issues or questions:
- Check logs: `.r2r/logs/ai-executions.jsonl`
- GitHub Issues: https://github.com/ready-to-release/eac/issues
- Documentation: See `docs/` directory

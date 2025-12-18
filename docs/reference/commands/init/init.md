# init

<!-- book:cmd init -->

## Configuration File Location

The configuration file is stored at:

```text
.r2r/eac/ai-provider.yml
```

This file should **NOT** be committed to version control.

```gitignore
# Add to .gitignore
.r2r/eac/ai-provider.yml
```

## Troubleshooting

### API Key Not Found

```text
Error: ANTHROPIC_API_KEY environment variable not set
```

Set the environment variable for your provider:

```bash
export ANTHROPIC_API_KEY="sk-ant-api03-..."  # Linux/macOS
$env:ANTHROPIC_API_KEY = "sk-ant-api03-..."  # PowerShell
```

### Invalid API Key

```text
Error: Authentication failed - invalid API key
```

1. Verify you copied the complete key (no spaces)
2. Check key hasn't expired
3. Generate a new key from provider dashboard

### Provider Not Recognized

```text
Error: Unknown provider: claude
```

Use exact provider name from `init --help`:

```bash
r2r eac init --ai claude-api  # Correct
r2r eac init --ai claude      # Invalid
```

### AI Features Not Working

```bash
# 1. Verify configuration exists
cat .r2r/eac/ai-provider.yml

# 2. Check environment variable
echo $ANTHROPIC_API_KEY

# 3. Re-initialize if needed
r2r eac init --ai claude-api
```

## See Also

- [How-to Guide](../../../how-to-guides/eac/commands/getting-started/setup-ai-provider.md) - Setup walkthrough
- [create commit-message](../create/commit-message.md) - AI commit messages
- [create spec](../create/spec.md) - AI specification generation
- [init Commands](../categories/init.md)

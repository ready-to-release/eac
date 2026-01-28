# init

<!-- book:cmd init -->

## Configuration File Location

The init command creates configuration files at:

**Team/Environment Config (safe to commit):**

```text
.r2r/eac/ai-provider.yml
```

Contains environment variable references (e.g., `${ANTHROPIC_API_KEY}`).
**Safe to commit** - no secrets stored directly.

**Personal Config (never commit):**

```text
.r2r/eac/ai-provider.personal.yml
```

Created when using `--ai-token` flag. Contains actual API keys.
**NEVER commit** - automatically gitignored.

## Advanced: Copying System Templates

By default, EAC uses built-in system defaults from `contracts/eac-core/0.1.0/defaults/`. These files don't need to be copied to your repository.

However, if you need to customize these configurations:

```bash
# Copy system default files to your repository
r2r eac init --copy-templates
```

This copies these files to `.r2r/eac/`:

- `ai-config.yml` - AI type definitions (specs, commit-message)
- `component-types.yml` - Component type definitions (go, typescript, etc.)
- `tool-config.yml` - Tool definitions and resource configuration
- `security-tools.yml` - Security tool configurations
- `logging.yml` - Logging configuration
- `environments.yml` - Test environment definitions

Once copied, you can edit these files and commit them. Your versions will take precedence over system defaults.

**When to use `--copy-templates`:**

- ✅ You need custom AI type definitions
- ✅ You need custom component types
- ✅ You want to version control these configurations

**When NOT to use `--copy-templates`:**

- ❌ Default configurations work for you (most users)
- ❌ You want automatic upgrades (system defaults update with EAC version)

### Recommended .gitignore Entry

```gitignore
# AI provider personal configuration (contains actual API keys)
.r2r/eac/*.personal.yml
.r2r/eac/*.local.yml
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

- [How-to Guide](../../../../how-to-guides/local-setup/configure-ai.md) - Setup walkthrough
- [create commit-message](../create/commit-message.md) - AI commit messages
- [create spec](../create/spec.md) - AI specification generation
- [init Commands](../categories/init.md)

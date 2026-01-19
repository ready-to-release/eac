# Setup AI Provider

## What You'll Accomplish

Configure AI provider (Anthropic Claude, OpenAI, or Google Gemini) to enable AI-powered features like commit message generation, PR descriptions, and specification creation.

!!! tip "Start Here"
    This is one of the first commands you should run when setting up EAC. The `init` command creates the configuration file that enables AI features throughout your development workflow.

## Prerequisites

### Required Knowledge

**New to r2r?** Learn these concepts first:

- [Quick Start Guide](../../../../tutorials/getting-started/quick-start.md) - Initialize r2r and understand AI provider configuration

### Required Setup

You need an Anthropic API key:

| Provider       | API Key Required    | Best For                                               |
| -------------- | ------------------- | ------------------------------------------------------ |
| **claude-api** | `ANTHROPIC_API_KEY` | AI-powered commit messages, specs, and PR descriptions |

**Note**: Currently only Anthropic Claude is supported via the `claude-api` provider.

## Configuration Approaches

There are two ways to configure your API key:

### Approach 1: Environment Variables (Recommended for Teams/CI)

The config file references environment variables, keeping keys out of version control:

```bash
# 1. Set environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# 2. Run init
r2r eac init --ai claude-api
```

**What happens**: Creates `.r2r/eac/ai-provider.yml` that references `ANTHROPIC_API_KEY` environment variable.

**Advantages**:

- Config file can be safely committed to git
- Easy to use different keys per environment (dev/staging/prod)
- Works well in CI/CD pipelines

### Approach 2: Personal Config with Token (For Personal Use)

Store the actual API token directly in a personal config file:

```bash
# Run init with token
r2r eac init --ai claude-api --ai-token sk-ant-api03-...
```

**What happens**: Creates `.r2r/eac/ai-provider.personal.yml` with the actual API key stored directly.

**Advantages**:

- No need to manage environment variables
- Simpler for personal development
- Automatically gitignored

**Important**: This file should NEVER be committed to version control.

## Steps

### Option A: Using Environment Variables

#### 1. Set Environment Variable

```bash
# Windows PowerShell
$env:ANTHROPIC_API_KEY = "sk-ant-api03-..."

# Linux/macOS
export ANTHROPIC_API_KEY="sk-ant-api03-..."
```

#### 2. Run Init Command

```bash
# Anthropic Claude API
r2r eac init --ai claude-api
```

### Option B: Using Personal Config

**Run init with your token directly:**

```bash
# Anthropic Claude API
r2r eac init --ai claude-api --ai-token sk-ant-api03-...
```

### Verify Configuration

Check that the config file was created:

```bash
# For environment variables approach
cat .r2r/eac/ai-provider.yml

# For personal config approach
cat .r2r/eac/ai-provider.personal.yml
```

### Test AI Features

```bash
# Make a change
echo "test" > test.txt
git add test.txt

# Generate AI commit message
r2r eac create commit-message

# Or commit directly with AI message
r2r eac work commit
```

## Configuration Files

The init command creates one of two config files depending on the approach used:

### Environment Variables Config (`.r2r/eac/ai-provider.yml`)

Created when using `--ai-provider` without `--ai-token`. References environment variables:

```yaml
provider: claude-api
api_key_env: ANTHROPIC_API_KEY
model: claude-sonnet-4-5
max_tokens: 4096
temperature: 0.7
```

**Safe to commit**: This file only contains the environment variable name, not the actual key.

### Personal Config (`.r2r/eac/ai-provider.personal.yml`)

Created when using `--ai-token`. Stores the actual API key:

```yaml
provider: claude-api
api_key: sk-ant-api03-...  # Actual token stored here
model: claude-sonnet-4-5
max_tokens: 4096
temperature: 0.7
```

**NEVER commit this file**: Contains your actual API key. Should be in `.gitignore`.

## Example Scenarios

### Example 1: Environment Variables (Team Setup)

```bash
# 1. Set API key environment variable
export ANTHROPIC_API_KEY="sk-ant-api03-..."

# 2. Initialize configuration
r2r eac init --ai claude-api

# Output:
# ✓ Created configuration at .r2r/eac/ai-provider.yml
# ✓ Provider: claude-api
# ✓ Model: claude-sonnet-4-5

# 3. Test it
echo "test change" > test.txt
git add test.txt
r2r eac create commit-message

# Output:
# Analyzing staged changes...
#
# docs: add test file
#
# Add test.txt to demonstrate AI commit message generation.
```

### Example 2: Personal Config (Individual Setup)

```bash
# 1. Initialize with token directly
r2r eac init --ai claude-api --ai-token sk-ant-api03-...

# Output:
# ✓ Created personal configuration at .r2r/eac/ai-provider.personal.yml
# ✓ Provider: claude-api
# ✓ Model: claude-sonnet-4-5
# ⚠ This file is gitignored - do not commit it

# 2. Test it (no environment variable needed!)
echo "test change" > test.txt
git add test.txt
r2r eac create commit-message

# Works immediately - token is in the config file
```

## Common Issues

| Problem                            | Solution                                                                      |
| ---------------------------------- | ----------------------------------------------------------------------------- |
| "API key not found" (env vars)     | Set `ANTHROPIC_API_KEY` environment variable before running init              |
| "API key not found" (personal)     | Check `.r2r/eac/ai-provider.personal.yml` exists and contains `api_key` field |
| "Invalid provider"                 | Only `claude-api` is currently supported                                      |
| "Permission denied"                | Check directory permissions for `.r2r/eac/`                                   |
| AI features not working (env vars) | Verify `ANTHROPIC_API_KEY` is still set in current session                    |
| Personal config committed to git   | Add `.r2r/eac/ai-provider.personal.yml` to `.gitignore` immediately           |

## Making Environment Variables Permanent (Option A Only)

If you chose Option A (environment variables), you may want to make them permanent:

### Windows (PowerShell)

```powershell
# Set permanent user environment variable
[System.Environment]::SetEnvironmentVariable('ANTHROPIC_API_KEY', 'sk-ant-api03-...', 'User')

# Restart PowerShell for changes to take effect
```

### Linux/macOS

```bash
# Add to ~/.bashrc or ~/.zshrc
echo 'export ANTHROPIC_API_KEY="sk-ant-api03-..."' >> ~/.bashrc

# Reload configuration
source ~/.bashrc
```

## Protecting Personal Config (Option B Only)

If you chose Option B (personal config with `--ai-token`), ensure it's in `.gitignore`:

### Add to .gitignore

```gitignore
# AI provider personal configuration (contains actual API keys)
.r2r/eac/ai-provider.personal.yml
```

### Verify It's Ignored

```bash
# Should show nothing (file is ignored)
git status | grep ai-provider.personal

# If it shows up, you need to add it to .gitignore
```

**Critical**: If you accidentally committed this file, you must:

1. Remove it from git history
2. Rotate (regenerate) your API key immediately
3. Add it to `.gitignore`

## Next Steps

- [Make Commits with AI](../development-workflow/make-commits-with-ai.md) → Use AI commits

## Related Commands

- [`init`](../../../../reference/commands/init/init.md) - Full init command reference
- [`work commit`](../../../../reference/commands/work/commit.md) - Commit with AI messages
- [`create commit-message`](../../../../reference/commands/create/commit-message.md) - Generate commit messages

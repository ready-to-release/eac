# Quick Start Guide
Get up and running with the r2r CLI. This tutorial walks you through installation, initialization, and running your first commands.

**Prerequisites:** Command-line access, internet connection

## What You'll Learn

By the end of this tutorial, you'll be able to:

- Install the r2r CLI on your platform (Linux, macOS, or Windows)
- Initialize r2r configuration in a project
- Run basic commands to explore your repository
- Execute tests using the r2r CLI
- Navigate to the next steps in your learning journey

## Step 1: Install the CLI

The R2R CLI is distributed as a pre-built binary for Linux, macOS, and Windows.

### Linux and macOS

Run the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/cli/install.sh | bash
```

The script will:

- Detect your platform (OS and architecture)
- Download the latest r2r-cli release
- Install to `~/.local/bin/r2r` (or use `--system` for system-wide installation)
- Verify the installation

If `~/.local/bin` is not in your PATH, add it to your shell profile:

```bash
export PATH="$PATH:$HOME/.local/bin"
```

Then restart your terminal or run `source ~/.bashrc` (or `~/.zshrc`).

### Windows

Run the installation script in PowerShell:

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/cli/install.ps1 | iex
```

The script will:

- Download the latest r2r-cli release for Windows
- Install to `%LOCALAPPDATA%\r2r` (or use `-System` for Program Files)
- Add the installation directory to your PATH
- Verify the installation

You may need to restart your terminal for PATH changes to take effect.

## Step 2: Verify Installation

Check that the CLI is installed correctly:

```bash
r2r --version
```

You should see version information displayed.

## Step 3: Initialize Your Project

Navigate to your project directory and initialize the R2R configuration:

```bash
cd /path/to/your/project
r2r init --ai claude-api
```

This command:

- Creates the `.r2r/eac/` directory structure
- Generates `ai-provider.yml` with AI provider settings
- Uses environment variable placeholders for API keys (safe to commit)

Available AI providers:

- `claude-api` - Anthropic Claude (requires `ANTHROPIC_API_KEY`)
- `openai` - OpenAI GPT (requires `OPENAI_API_KEY`)
- `gemini` - Google Gemini (requires `GOOGLE_API_KEY`)

To use a personal configuration with actual API tokens (gitignored):

```bash
r2r init --ai claude-api --ai-token sk-ant-your-key-here
```

## Step 4: Explore Available Commands

List all available commands:

```bash
r2r show help
```

Get help for a specific command:

```bash
r2r show help show modules
```

## Step 5: View Your Project Structure

Show all modules in your repository:

```bash
r2r show modules
```

This displays a table of all modules with their type and root path.

Show the project configuration:

```bash
r2r show config
```

## Step 6: Run Tests

To run tests for your project:

```bash
r2r test
```

This runs all modules with the default test suites (L0-L2 fast tests).

To test a specific module:

```bash
r2r test eac-commands # default fast suites
```

To run a different test suite:

```bash
r2r test --suite acceptance
```

Available test suites:

- `unit` - L0-L1 tests (fast unit tests, <5 min)
- `integration` - L2 tests (Docker-based emulated tests, <15 min)
- `acceptance` - L3 tests (production-like tests in PLTE, 1-2 hours)
- `production-verification` - L4+PIV tests (production smoke tests)

## What You Learned

Congratulations! You've successfully:

- ✅ Installed the r2r CLI on your system
- ✅ Initialized r2r configuration with AI provider settings
- ✅ Explored available commands with `r2r show help`
- ✅ Viewed repository structure with `r2r show modules`
- ✅ Ran tests with different test suites

## Key Concepts Covered

- **r2r CLI installation** - Binary distribution for multiple platforms
- **Project initialization** - Setting up `.r2r/eac/` configuration
- **AI provider configuration** - Claude, OpenAI, or Gemini integration
- **Repository exploration** - Using `show` commands to understand structure
- **Test suites** - Different test levels (unit, integration, acceptance, production-verification)

## Next Steps

### Continue Learning

- **Next tutorial:** [Your First Feature Specification](./first-specification.md) - Learn to write Gherkin specifications

### Apply What You Learned

Now that you know the basics of r2r, you can accomplish these tasks:

- **[Get Help with Commands](../../how-to-guides/eac/commands/getting-started/get-help-with-commands.md)** - Find and understand any r2r command
- **[Explore Your Repository](../../how-to-guides/eac/commands/getting-started/explore-your-repository.md)** - Discover modules, files, and structure
- **[Setup AI Provider](../../how-to-guides/eac/commands/getting-started/setup-ai-provider.md)** - Configure Claude, OpenAI, or Gemini

### Dive Deeper

- [Everything as Code Paradigm](../../explanation/everything-as-code/paradigm.md) - Understand the philosophy
- [Command Reference](../../reference/commands/) - Complete command documentation

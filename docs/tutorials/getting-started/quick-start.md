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

## Step 3: Initialize R2R Configuration

Navigate to your project directory and create the R2R CLI configuration:

```bash
cd /path/to/your/project
r2r init
```

This command:

- Creates the `.r2r/` directory
- Generates `r2r-cli.yml` for extension management
- Sets up the extension registry configuration

## Step 4: Install EAC Extension

Install the Everything-as-Code (EAC) extension:

```bash
r2r install eac
```

This command:

- Pulls the EAC extension Docker image from the registry
- Registers the `eac` extension in your configuration
- Makes `r2r eac` commands available

## Step 5: Initialize EAC in Your Project

Configure the EAC extension for your project:

```bash
r2r eac init --ai-provider claude-api
```

This command:

- Creates the `.r2r/eac/` directory structure
- Generates `ai-provider.yml` with AI provider settings
- Uses environment variable placeholders for API keys (safe to commit)
- Other configuration files use system defaults automatically (no copying needed)

Available AI providers:

- `claude-api` - Anthropic Claude (requires `ANTHROPIC_API_KEY`)
- `openai` - OpenAI GPT (requires `OPENAI_API_KEY`)
- `gemini` - Google Gemini (requires `GOOGLE_API_KEY`)

To use a personal configuration with actual API tokens (gitignored):

```bash
r2r eac init --ai-provider claude-api --ai-token sk-ant-your-key-here
```

!!! tip "Configuration Files"
    The init command only creates user-specific files (`ai-provider.yml`). Other configuration files like `ai-config.yml` and templates are automatically loaded from built-in system defaults. See [Understanding Configuration Files](./configuration-files.md) to learn more.

## Step 6: Set Your API Key

Before running commands that use AI, set your API key as an environment variable:

**Linux/macOS:**

```bash
export ANTHROPIC_API_KEY=sk-ant-your-key-here
```

**Windows (PowerShell):**

```powershell
$env:ANTHROPIC_API_KEY = "sk-ant-your-key-here"
```

## Step 7: Explore Available Commands

List all available commands:

```bash
r2r eac help
```

Get help for a specific command:

```bash
r2r eac help show
```

!!! tip "Command Discovery"
    EAC provides 147 commands organized into 10 categories (show, get, build, test, create, validate, release, pipeline, work, and more). See [Discovering Available Commands](../../how-to-guides/eac/commands/getting-started/discover-commands.md) for a complete guide to finding and using all commands.

## Step 8: Analyze Your Repository

Before using other commands, analyze your repository structure:

```bash
r2r eac analyze modules
```

This command:

- Scans your repository to discover modules
- Generates `.r2r/eac/repository.yml` with module metadata
- Creates `.r2r/eac/books.yml` with architecture patterns

## Step 9: View Your Project Structure

Show all modules discovered in your repository:

```bash
r2r eac show modules
```

This displays a table of all modules with their type and root path.

Show the project configuration:

```bash
r2r eac show config
```

## Step 10: Run Tests

To run tests for your project:

```bash
r2r eac test
```

This runs all modules with the default test suites (L0-L2 fast tests).

To test a specific module:

```bash
r2r eac test eac-commands # default fast suites
```

To run a different test suite:

```bash
r2r eac test --suite acceptance
```

Available test suites:

- `unit` - L0-L1 tests (fast unit tests, <5 min)
- `integration` - L2 tests (Docker-based emulated tests, <15 min)
- `acceptance` - L3 tests (production-like tests in PLTE, 1-2 hours)
- `production-verification` - L4+PIV tests (production smoke tests)

## What You Learned

Congratulations! You've successfully:

- ✅ Installed the r2r CLI on your system
- ✅ Initialized R2R CLI configuration with `r2r init`
- ✅ Installed the EAC extension with `r2r install eac`
- ✅ Configured EAC with AI provider settings
- ✅ Set up your API key for AI-powered commands
- ✅ Analyzed your repository to discover modules
- ✅ Explored available commands with `r2r eac help`
- ✅ Viewed repository structure with `r2r eac show modules`
- ✅ Ran tests with different test suites

## Key Concepts Covered

- **r2r CLI installation** - Binary distribution for multiple platforms
- **R2R CLI initialization** - Creating `.r2r/r2r-cli.yml` for extension management
- **Extension installation** - Installing containerized extensions like EAC
- **EAC configuration** - Setting up `.r2r/eac/` with AI provider settings
- **Configuration layering** - System defaults vs. user overrides
- **AI provider configuration** - Claude, OpenAI, or Gemini integration
- **Repository analysis** - Discovering modules and architecture patterns
- **Repository exploration** - Using `show` commands to understand structure
- **Test suites** - Different test levels (unit, integration, acceptance, production-verification)

## Next Steps

### Continue Learning

- **Next tutorial:** [Understanding Configuration Files](./configuration-files.md) - Learn about `.r2r/` and `.r2r/eac/` files
- **Then:** [Creating Your First Extension](./creating-your-first-extension.md) - Build a custom r2r extension

### Try Common Tasks

Now that you know the basics of r2r, try these common tasks:

- **[Discover Available Commands](../../how-to-guides/eac/commands/getting-started/discover-commands.md)** - Explore all 147 commands organized by category
- **[Get Help with Commands](../../how-to-guides/eac/commands/getting-started/get-help-with-commands.md)** - Find and understand any r2r command
- **[Explore Your Repository](../../how-to-guides/eac/commands/getting-started/explore-your-repository.md)** - Discover modules, files, and structure
- **[Setup AI Provider](../../how-to-guides/eac/commands/getting-started/setup-ai-provider.md)** - Configure Claude, OpenAI, or Gemini

### Dive Deeper

- **[Everything as Code Paradigm](../../explanation/everything-as-code/paradigm.md)** - Understand the philosophy
- **[Command Reference](../../reference/commands/)** - Complete command documentation
- **[Creating Extensions Guide](../../how-to-guides/r2r/creating-extensions.md)** - Build custom extensions

# Quick Start Guide

Get up and running with the EAC CLI. This tutorial walks you through installation, initialization, and running your first commands.

**Prerequisites:** Command-line access, internet connection

## What You'll Learn

By the end of this tutorial, you'll be able to:

- Install the EAC CLI on your platform (Linux, macOS, or Windows)
- Initialize EAC configuration in a project
- Run basic commands to explore your repository
- Execute tests using the EAC CLI
- Navigate to the next steps in your learning journey

## Step 1: Install the CLI

The EAC CLI is distributed as a pre-built binary for Linux, macOS, and Windows.

### Linux and macOS

Run the installation script:

```bash
curl -fsSL https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/sh/eac/install.sh | bash
```

The script will:

- Detect your platform (OS and architecture)
- Download the latest EAC release
- Install to `~/.local/bin/eac` (or use `--system` for system-wide installation)
- Verify the installation

If `~/.local/bin` is not in your PATH, add it to your shell profile:

```bash
export PATH="$PATH:$HOME/.local/bin"
```

Then restart your terminal or run `source ~/.bashrc` (or `~/.zshrc`).

### Windows

Run the installation script in PowerShell:

```powershell
irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
```

The script will:

- Download the latest EAC release for Windows
- Install to `%LOCALAPPDATA%\eac` (or use `-System` for Program Files)
- Add the installation directory to your PATH
- Verify the installation

You may need to restart your terminal for PATH changes to take effect.

## Step 2: Verify Installation

Check that the CLI is installed correctly:

```bash
eac version
```

You should see version information displayed.

## Step 3: Initialize EAC in Your Project

Navigate to your project directory and initialize EAC:

```bash
cd /path/to/your/project
eac init --ai-provider claude-api
```

This command:

- Creates the `.eac/` directory structure
- Generates `ai-provider.yml` with AI provider settings
- Uses environment variable placeholders for API keys (safe to commit)
- Other configuration files use system defaults automatically (no copying needed)

Available AI providers:

- `claude-api` - Anthropic Claude (requires `ANTHROPIC_API_KEY`)
- `openai` - OpenAI GPT (requires `OPENAI_API_KEY`)
- `gemini` - Google Gemini (requires `GOOGLE_API_KEY`)

To use a personal configuration with actual API tokens (gitignored):

```bash
eac init --ai-provider claude-api --ai-token sk-ant-your-key-here
```

!!! tip "Configuration Files"

    The init command only creates user-specific files (`ai-provider.yml`).
    Other configuration files like `ai-config.yml` and templates are automatically loaded from built-in system defaults.
    See [Understanding Configuration Files](./configuration-files.md) to learn more.

## Step 4: Set Your API Key

Before running commands that use AI, set your API key as an environment variable:

**Linux/macOS:**

```bash
export ANTHROPIC_API_KEY=sk-ant-your-key-here
```

**Windows (PowerShell):**

```powershell
$env:ANTHROPIC_API_KEY = "sk-ant-your-key-here"
```

## Step 5: Explore Available Commands

List all available commands:

```bash
eac help
```

Get help for a specific command:

```bash
eac help show
```

!!! tip "Command Discovery"

    EAC provides hundreds of commands organized into categories (show, get, build, test, create, validate, release, pipeline, work, and more).
    See [Discovering Available Commands](../../how-to-guides/eac/commands/getting-started/discover-commands.md) for a complete guide to finding and using all commands.

## Step 6: Initialize Your Repository

Before using other commands, initialize your repository structure:

```bash
eac init
```

This command:

- Scans your repository to discover modules
- Generates `.eac/repository.yml` with module metadata
- Creates `.eac/books.yml` with architecture patterns

## Step 7: View Your Project Structure

Show all modules discovered in your repository:

```bash
eac show modules
```

This displays a table of all modules with their type and root path.

Show the project configuration:

```bash
eac show config
```

## Step 8: Run Tests

To run tests for your project:

```bash
eac test
```

This runs all modules with the default test suites (L0-L2 fast tests).

To test a specific module:

```bash
eac test eac-commands # default fast suites
```

To run a different test suite:

```bash
eac test --suite acceptance
```

Available test suites:

- `unit` - L0-L1 tests (fast unit tests, <5 min)
- `integration` - L2 tests (Docker-based emulated tests, <15 min)
- `acceptance` - L3 tests (production-like tests in PLTE, 1-2 hours)
- `production-verification` - L4+PIV tests (production smoke tests)

## What You Learned

Congratulations! You've successfully:

- ✅ Installed the EAC CLI on your system
- ✅ Initialized EAC configuration with `eac init`
- ✅ Configured AI provider settings
- ✅ Set up your API key for AI-powered commands
- ✅ Analyzed your repository to discover modules
- ✅ Explored available commands with `eac help`
- ✅ Viewed repository structure with `eac show modules`
- ✅ Ran tests with different test suites

## Key Concepts Covered

- **EAC CLI installation** - Binary distribution for multiple platforms
- **EAC initialization** - Creating `.eac/` with configuration
- **Configuration layering** - System defaults vs. user overrides
- **AI provider configuration** - Claude, OpenAI, or Gemini integration
- **Repository analysis** - Discovering modules and architecture patterns
- **Repository exploration** - Using `show` commands to understand structure
- **Test suites** - Different test levels (unit, integration, acceptance, production-verification)

## Next Steps

### Continue Learning

- **Next tutorial:** [Understanding Configuration Files](./configuration-files.md) - Learn about `.eac/` configuration files

### Try Common Tasks

Now that you know the basics of EAC, try these common tasks:

- **[Discover Available Commands](../../how-to-guides/eac/commands/getting-started/discover-commands.md)**
  Explore all commands organized by category
- **[Get Help with Commands](../../how-to-guides/eac/commands/getting-started/get-help-with-commands.md)**
  Find and understand any EAC command
- **[Explore Your Repository](../../how-to-guides/eac/commands/getting-started/explore-your-repository.md)**
  Discover modules, files, and structure

### Dive Deeper

- **[Everything as Code Paradigm](../../explanation/everything-as-code/paradigm.md)** - Understand the philosophy
- **[Command Reference](../../reference/eac/commands/index.md)** - Complete command documentation
- **[Creating CLIE Extensions](../../how-to-guides/clie/creating-extensions.md)** - Build containerized extensions (optional)

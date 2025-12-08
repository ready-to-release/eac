<!-- EDITOR
# Editor: tutorials/quick-start.md

## Soul

Hands-on walkthrough to install the R2R CLI and run your first commands in minutes.

## Sections

1. Step 1: Install the CLI
2. Step 2: Verify Installation
3. Step 3: Initialize Your Project
4. Step 4: Explore Available Commands
5. Step 5: View Your Project Structure
6. Step 6: Run Tests
7. Next Steps
-->

# Quick Start Guide

This tutorial will help you install the R2R CLI and run your first commands.

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
- Generates `eac-config.yml` with AI provider settings
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

This runs all modules with the default "commit" test suite (L0-L2 fast tests).

To test a specific module:

```bash
r2r test eac-commands
```

To run a different test suite:

```bash
r2r test --suite acceptance
```

Available test suites:

- `commit` - L0-L2 tests (fast, for pre-commit)
- `acceptance` - IV/OV/PV tests (PLTE acceptance)
- `production-verification` - L4+PIV tests (production smoke tests)

## Next Steps

Congratulations! You now have the R2R CLI installed and working. Continue to [Your First Feature Specification](./first-specification.md) to learn how to write Gherkin specifications.

{{ diataxis_footer() }}
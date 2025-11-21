#!/usr/bin/env bash
# Claude Code Environment Setup for Bash/Zsh
#
# This script configures environment variables needed for Claude Code MCP servers.
# It should be sourced from your shell profile.
#
# Installation:
# 1. Add this line to your ~/.bashrc, ~/.zshrc, or ~/.profile:
#    source "/path/to/this/script/scripts/sh/claude-env-setup.sh"
# 2. Set your GITHUB_TOKEN in your shell profile
# 3. Reload your shell or run: source ~/.bashrc (or ~/.zshrc)

echo "Setting up Claude Code environment..."

# Auto-detect project root directory
# This script is located at: scripts/sh/claude-env-setup.sh
# Project root is two directories up from this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

if [ -d "$PROJECT_ROOT" ]; then
    cd "$PROJECT_ROOT" || {
        echo "Warning: Failed to change to project directory: $PROJECT_ROOT"
    }
    echo "Changed to project directory: $PROJECT_ROOT"
else
    echo "Warning: Project directory not found: $PROJECT_ROOT"
    echo "Script location detection failed. Please verify script is in scripts/sh/"
fi

# Configure GitHub Token for the GitHub MCP server
# IMPORTANT: Set your actual GitHub token in your shell profile
# Generate a token at: https://github.com/settings/tokens
#
# Add this to your ~/.bashrc or ~/.zshrc (BEFORE sourcing this script):
# export GITHUB_TOKEN="your-github-token-here"
if [ -z "$GITHUB_TOKEN" ]; then
    echo "Warning: GITHUB_TOKEN not set. GitHub MCP server will not function."
    echo "Set it in your profile: export GITHUB_TOKEN='your-token-here'"
fi

# Configure Anthropic API Key (if needed for other tools)
# IMPORTANT: Set your actual API key in your shell profile
#
# Add this to your ~/.bashrc or ~/.zshrc (BEFORE sourcing this script):
# export ANTHROPIC_API_KEY="your-api-key-here"
if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo "Info: ANTHROPIC_API_KEY not set (optional for MCP servers)"
fi

# Display MCP server configuration status
echo "Claude Code MCP servers will use:"
echo "  Commands Server: go run ./src/mcp/commands/main.go (local)"
echo "  GitHub Server:   Official GitHub MCP server"
echo ""
echo "MCP servers are configured in .claude/settings.json"
echo "Environment setup complete."

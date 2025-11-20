# Claude Code Environment Setup for PowerShell
#
# This script configures environment variables needed for Claude Code MCP servers.
# It should be sourced from your PowerShell profile.
#
# Installation:
# 1. Copy the configuration block below to your PowerShell profile
# 2. Update the paths and tokens as needed
# 3. Reload your profile or restart PowerShell
#
# Quick install: Add this line to your PowerShell profile ($PROFILE):

Set-StrictMode -Version 'Latest'

$VerbosePreference = "SilentlyContinue"
$InformationPreference = "Continue"

Write-Information "Setting up Claude Code environment..."

# Remove OneDrive from PSModulePath if present
# This prevents module conflicts when using OneDrive folder sync
$psModulePattern = "[C]:\\Users\\[a-z\.]*\\OneDrive -[a-zA-Z ]*\\Documents\\PowerShell\\Modules;"
if ($env:PSModulePath -match $psModulePattern) {
    Write-Information "Removing OneDrive from PSModulePath"
    $env:PSModulePath = $env:PSModulePath -replace $psModulePattern
}

# Enable predictive IntelliSense for PowerShell Core
if ($PSVersionTable.PSEdition -eq "Core") {
    Set-PSReadLineOption -PredictionSource History
}

# Auto-detect project root directory
# This script is located at: scripts/pwsh/claude-env-setup.ps1
# Project root is two directories up from this script
$projectRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)

if (Test-Path $projectRoot -PathType Container) {
    Set-Location $projectRoot
    Write-Information "Changed to project directory: $projectRoot"
}
else {
    Write-Warning "Project directory not found: $projectRoot"
    Write-Warning "Script location detection failed. Please verify script is in scripts/pwsh/"
}

# Configure GitHub Token for the GitHub MCP server
# IMPORTANT: Replace with your actual GitHub token
# Generate a token at: https://github.com/settings/tokens
if (-not $env:GITHUB_TOKEN) {
    Write-Warning "GITHUB_TOKEN not set. GitHub MCP server will not function."
    Write-Warning "Set it in your profile: `$env:GITHUB_TOKEN = 'your-token-here'"
}

# Configure Anthropic API Key (if needed for other tools)
# IMPORTANT: Replace with your actual API key
if (-not $env:ANTHROPIC_API_KEY) {
    Write-Information "ANTHROPIC_API_KEY not set (optional for MCP servers)"
}

# Display MCP server configuration status
Write-Information "Claude Code MCP servers will use:"
Write-Information "  Commands Server: go run ./src/mcp/commands/main.go"
Write-Information "  GitHub Server:   go run ./src/mcp/github/main.go"
Write-Information ""
Write-Information "MCP servers are configured in .claude/settings.json"
Write-Information "Environment setup complete."

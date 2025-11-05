#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Import the CommandRunner module and create top-level command aliases

.DESCRIPTION
    This script imports the go.psm1 module with -Force and creates session-wide aliases
    for all top-level commands (show, list, describe, commit-ai, etc.).
    The aliases are temporary and only last for the current PowerShell session.

    Welcome messages are shown only once per session (tracked via $env:CMD_ALIASES_EMITTED).

.PARAMETER NoAlias
    Skip creating command aliases. Use this if you only want the module functions.

.EXAMPLE
    .\importer.ps1
    show files
    list commands

.EXAMPLE
    .\importer.ps1 -NoAlias
    Invoke-GoSrcCommand show files

.NOTES
    Can be executed directly (.\importer.ps1) or dot-sourced (. .\importer.ps1).
    Global scope functions are created, so aliases persist in your session either way.
#>

[CmdletBinding()]
param(
    [switch]$NoAlias
)

$ErrorActionPreference = "Stop"

# Clear cache to force refresh
$env:SRC_COMMANDS_DESCRIBE = $null

# Get the module path
$ModulePath = Join-Path $PSScriptRoot "scripts\pwsh\go-invoker\go.psm1"

# Check if module exists
if (-not (Test-Path $ModulePath)) {
    Write-Error "Module not found at: $ModulePath"
    exit 1
}

# Import the module with -Force to reload if already loaded
Import-Module $ModulePath -Force

Write-Host "✅ CommandRunner module imported successfully!" -ForegroundColor Green

# Create top-level command aliases unless -NoAlias specified
if (-not $NoAlias) {
    New-TopLevelAliases
} else {
    Write-Host ""
    Write-Host "You can now use:" -ForegroundColor Cyan
    Write-Host "  Invoke-GoSrcCommand <command-name> [args...]" -ForegroundColor White
    Write-Host ""
    Write-Host "To create command aliases, call:" -ForegroundColor Cyan
    Write-Host "  New-TopLevelAliases" -ForegroundColor Gray
    Write-Host ""
}

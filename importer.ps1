#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Import the CommandRunner module and optionally create 'run' alias

.DESCRIPTION
    This script imports the runner.psm1 module with -Force and creates a session-wide 'run' alias.
    The alias is temporary and only lasts for the current PowerShell session.

    Welcome messages are shown only once per session (tracked via $env:RUN_WELCOME_EMITTED).

.PARAMETER NoAlias
    Skip creating the 'run' alias. Use this if you only want the module functions.

.EXAMPLE
    . .\importer.ps1
    run example-show-modules

.EXAMPLE
    . .\importer.ps1 -NoAlias
    Invoke-GoSrcCommand example-show-modules

.NOTES
    Must be dot-sourced ('. .\importer.ps1') for the alias to work in your session.
#>

[CmdletBinding()]
param(
    [switch]$NoAlias
)

$ErrorActionPreference = "Stop"

# Check if welcome already shown this session
$ShowWelcome = -not $env:RUN_WELCOME_EMITTED

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

# Create alias unless -NoAlias specified
if (-not $NoAlias) {
    New-RunAlias
} else {
    if ($ShowWelcome) {
        Write-Host ""
        Write-Host "You can now use:" -ForegroundColor Cyan
        Write-Host "  Invoke-GoSrcCommand <command-name> [args...]" -ForegroundColor White
        Write-Host ""
        Write-Host "To create the 'run' alias, call:" -ForegroundColor Cyan
        Write-Host "  New-RunAlias" -ForegroundColor Gray
        Write-Host ""
    }
}

# Mark welcome as shown for this session
$env:RUN_WELCOME_EMITTED = "1"

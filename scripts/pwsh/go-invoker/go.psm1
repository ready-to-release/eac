<#
.SYNOPSIS
    PowerShell module for running commands from src/commands/

.DESCRIPTION
    This module provides the 'Invoke-GoSrcCommand' function (exported as 'run')
    with intelligent tab completion for nested commands.

.NOTES
    Module Name: CommandRunner
    Author: CLI Tools
    Version: 2.0.0
#>

# Get the repository root (three levels up from this module)
$Script:RepoRoot = Split-Path (Split-Path (Split-Path $PSScriptRoot -Parent) -Parent) -Parent

# Cache for command structure
$Script:CommandStructureCache = $null
$Script:CommandStructureCacheTime = $null

<#
.SYNOPSIS
    Get structured command information from Go

.DESCRIPTION
    Calls 'go run . describe-commands' to get JSON structure of all commands
    with caching to avoid repeated calls
#>
function Get-CommandStructure {
    [CmdletBinding()]
    param()

    # Cache for 60 seconds
    if ($Script:CommandStructureCache -and $Script:CommandStructureCacheTime) {
        $age = (Get-Date) - $Script:CommandStructureCacheTime
        if ($age.TotalSeconds -lt 60) {
            return $Script:CommandStructureCache
        }
    }

    $commandsPath = Join-Path $Script:RepoRoot "src/commands"

    Push-Location $commandsPath
    try {
        $jsonOutput = & go run . describe-commands 2>$null
        if ($LASTEXITCODE -eq 0) {
            $structure = $jsonOutput | ConvertFrom-Json -AsHashtable
            $Script:CommandStructureCache = $structure
            $Script:CommandStructureCacheTime = Get-Date
            return $structure
        }
    } catch {
        Write-Verbose "Failed to get command structure: $_"
    } finally {
        Pop-Location
    }

    return $null
}

<#
.SYNOPSIS
    Run a command from src/commands/

.DESCRIPTION
    Executes a Go command with intelligent tab completion for nested commands.

.PARAMETER First
    First part of the command

.PARAMETER Second
    Second part of the command (optional)

.PARAMETER Third
    Third part of the command (optional)

.PARAMETER Fourth
    Fourth part of the command (optional)

.EXAMPLE
    Invoke-GoSrcCommand show modules

.EXAMPLE
    Invoke-GoSrcCommand describe-commands
#>
function Invoke-GoSrcCommand {
    [CmdletBinding()]
    param(
        [Parameter(Position=0)]
        [ArgumentCompleter({ Get-GoSrcCommandPart })]
        [string]$First,

        [Parameter(Position=1)]
        [ArgumentCompleter({ param($cmd, $param, $word, $ast, $bound)
            Get-GoSrcCommandPart -First $bound.First
        })]
        [string]$Second,

        [Parameter(Position=2)]
        [ArgumentCompleter({ param($cmd, $param, $word, $ast, $bound)
            Get-GoSrcCommandPart -First $bound.First -Second $bound.Second
        })]
        [string]$Third,

        [Parameter(Position=3)]
        [ArgumentCompleter({ param($cmd, $param, $word, $ast, $bound)
            Get-GoSrcCommandPart -First $bound.First -Second $bound.Second -Third $bound.Third
        })]
        [string]$Fourth
    )

    $ErrorActionPreference = "Stop"

    # Build command parts array
    $CommandParts = @()
    if ($First) { $CommandParts += $First }
    if ($Second) { $CommandParts += $Second }
    if ($Third) { $CommandParts += $Third }
    if ($Fourth) { $CommandParts += $Fourth }

    # Check if command parts provided
    if ($CommandParts.Count -eq 0) {
        Write-Host "Usage: run <command> [subcommand] [args...]" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Available commands:"

        $structure = Get-CommandStructure
        if ($structure) {
            $structure.commands | Sort-Object name | ForEach-Object {
                Write-Host "  $($_.name)" -ForegroundColor Cyan
                if ($_.description) {
                    Write-Host "    $($_.description)" -ForegroundColor Gray
                }
            }
        }

        return
    }

    $commandsPath = Join-Path $Script:RepoRoot "src/commands"

    # Run the command via dispatcher
    Push-Location $commandsPath
    try {
        & go run . @CommandParts

        # Propagate exit code
        if ($LASTEXITCODE -ne 0) {
            throw "Command failed with exit code $LASTEXITCODE"
        }
    } finally {
        Pop-Location
    }
}

<#
.SYNOPSIS
    Get available Go commands from src/commands/

.DESCRIPTION
    Returns command information using the structured describe-commands output

.PARAMETER Simple
    Return only command names as strings instead of full objects

.EXAMPLE
    Get-GoSrcCommands
    # Returns: Command objects with name, description, etc.

.EXAMPLE
    Get-GoSrcCommands -Simple
    # Returns: @("list-commands", "show files", "show modules")
#>
function Get-GoSrcCommands {
    [CmdletBinding()]
    param(
        [switch]$Simple
    )

    $structure = Get-CommandStructure
    if (-not $structure) {
        return @()
    }

    if ($Simple) {
        return $structure.commands | ForEach-Object { $_.name } | Sort-Object
    } else {
        return $structure.commands | Sort-Object { $_.name }
    }
}

<#
.SYNOPSIS
    Get next command part for auto-completion

.DESCRIPTION
    Returns available command parts based on what's already been typed

.PARAMETER First
    First part already entered

.PARAMETER Second
    Second part already entered

.PARAMETER Third
    Third part already entered

.EXAMPLE
    Get-GoSrcCommandPart
    # Returns: @("describe-commands", "list-commands", "show")

.EXAMPLE
    Get-GoSrcCommandPart -First "show"
    # Returns: @("files", "modules")
#>
function Get-GoSrcCommandPart {
    [CmdletBinding()]
    param(
        [string]$First,
        [string]$Second,
        [string]$Third
    )

    # Inline Get-CommandStructure logic to avoid scoping issues
    $commandsPath = Join-Path $Script:RepoRoot "src/commands"

    Push-Location $commandsPath
    try {
        $jsonOutput = & go run . describe commands 2>$null
        if ($LASTEXITCODE -ne 0) {
            return @()
        }
        $structure = $jsonOutput | ConvertFrom-Json -AsHashtable
    } catch {
        return @()
    } finally {
        Pop-Location
    }

    if (-not $structure) {
        return @()
    }

    # Build the path from provided parts
    $parts = @()
    if ($First) { $parts += $First }
    if ($Second) { $parts += $Second }
    if ($Third) { $parts += $Third }

    $path = $parts -join " "

    # Get suggestions from tree (it's a hashtable now)
    $suggestions = @()
    if ($structure.tree.ContainsKey($path)) {
        $suggestions += $structure.tree[$path]
    }

    # If we're at root level, add all parent prefixes (first parts of commands)
    if ($path -eq "") {
        foreach ($cmd in $structure.commands) {
            $firstPart = $cmd.parts[0]
            if ($suggestions -notcontains $firstPart) {
                $suggestions += $firstPart
            }
        }
    }

    return $suggestions
}

<#
.SYNOPSIS
    Creates a global 'run' alias for the current session

.DESCRIPTION
    Creates a global alias 'run' that points to Invoke-GoSrcCommand
    with intelligent tab completion for nested commands.

.EXAMPLE
    New-RunAlias
    run show <TAB>  # Suggests: files, modules
#>
function New-RunAlias {
    [CmdletBinding()]
    param()

    # Check if alias already exists
    if (Get-Alias -Name run -ErrorAction SilentlyContinue) {
        if (-not $env:RUN_WELCOME_EMITTED) {
            Write-Host "Alias 'run' already exists" -ForegroundColor Yellow
        }
        return
    }

    # Create the alias
    Set-Alias -Name run -Value Invoke-GoSrcCommand -Scope Global -Force

    # Show welcome only once per session
    if (-not $env:RUN_WELCOME_EMITTED) {
        Write-Host "✅ Created global alias 'run' for this session" -ForegroundColor Green
        Write-Host ""
        Write-Host "You can now use:" -ForegroundColor Cyan
        Write-Host "  run <command> [subcommand] [args...]" -ForegroundColor White
        Write-Host ""
        Write-Host "Intelligent tab completion enabled!" -ForegroundColor Gray
        Write-Host "  run show <TAB>    # Suggests: files, modules" -ForegroundColor Gray
        Write-Host ""
        Write-Host "To make this permanent, add this to your " -NoNewline -ForegroundColor Gray
        Write-Host "`$PROFILE" -NoNewline -ForegroundColor White
        Write-Host ":" -ForegroundColor Gray
        Write-Host "  Import-Module ""$($Script:RepoRoot)\scripts\pwsh\go-invoker\go.psm1"" -Force" -ForegroundColor Gray
        Write-Host "  New-RunAlias" -ForegroundColor Gray
        Write-Host ""

        $env:RUN_WELCOME_EMITTED = "1"
    }
}

# Export all functions (including Get-CommandStructure for ArgumentCompleters)
Export-ModuleMember -Function Invoke-GoSrcCommand, New-RunAlias, Get-GoSrcCommands, Get-GoSrcCommandPart, Get-CommandStructure

<#
.SYNOPSIS
    PowerShell module for running commands from go/cli/eac/

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
    Calls 'go run . get commands' to get JSON structure of all commands
    and caches result in $env:SRC_COMMANDS_DESCRIBE for session persistence
#>
function Get-CommandStructure {
    [CmdletBinding()]
    param()

    # Check environment variable cache first (persists across module reloads)
    if ($env:SRC_COMMANDS_DESCRIBE) {
        try {
            $structure = $env:SRC_COMMANDS_DESCRIBE | ConvertFrom-Json -AsHashtable
            return $structure
        } catch {
            Write-Verbose "Failed to parse cached command structure: $_"
            # Clear invalid cache
            $env:SRC_COMMANDS_DESCRIBE = $null
        }
    }

    # Cache miss - fetch from commands binary (or build it)
    $commandsPath = Join-Path $Script:RepoRoot "go/cli/eac"
    $commandsBinary = Join-Path $Script:RepoRoot "out/tools/eac.exe"

    # Build eac.exe if missing
    if (-not (Test-Path $commandsBinary)) {
        Write-Host "   Building commands executable..." -ForegroundColor DarkGray
        # Ensure output directory exists
        $toolsDir = Split-Path $commandsBinary -Parent
        if (-not (Test-Path $toolsDir)) {
            New-Item -Path $toolsDir -ItemType Directory -Force | Out-Null
        }
        Push-Location $commandsPath
        try {
            & go build -o $commandsBinary .
            if ($LASTEXITCODE -ne 0) {
                Write-Verbose "Failed to build commands binary"
                Pop-Location
                return $null
            }
        } finally {
            Pop-Location
        }
    }

    try {
        $jsonOutput = & $commandsBinary get commands 2>$null

        if ($LASTEXITCODE -eq 0) {
            # Store in environment variable as JSON string
            $env:SRC_COMMANDS_DESCRIBE = $jsonOutput

            # Return as hashtable
            $structure = $jsonOutput | ConvertFrom-Json -AsHashtable
            return $structure
        }
    } catch {
        Write-Verbose "Failed to get command structure: $_"
    }

    return $null
}

<#
.SYNOPSIS
    Run a command from go/cli/eac/

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
    Invoke-GoSrcCommand get commands
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
        [string]$Fourth,

        [Parameter(ValueFromRemainingArguments=$true)]
        [string[]]$RemainingArgs
    )

    $ErrorActionPreference = "Stop"

    # Build command parts array
    $CommandParts = @()
    if ($First) { $CommandParts += $First }
    if ($Second) { $CommandParts += $Second }
    if ($Third) { $CommandParts += $Third }
    if ($Fourth) { $CommandParts += $Fourth }
    if ($RemainingArgs) { $CommandParts += $RemainingArgs }

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

    $commandsPath = Join-Path $Script:RepoRoot "go/cli/eac"
    $commandsBinary = Join-Path $Script:RepoRoot "out/tools/eac.exe"

    # Build the binary to a stable location instead of using go run.
    # go run compiles to a temp dir every invocation and can touch go.sum/go.work.sum
    # timestamps, which invalidates the mtime-based hash cache and forces downstream
    # modules to rebuild. Building to out/tools/ is cache-neutral: Go's build cache
    # makes it near-instant when sources are unchanged, and the output location is
    # outside any module's glob patterns.
    $toolsDir = Split-Path $commandsBinary -Parent
    if (-not (Test-Path $toolsDir)) {
        New-Item -Path $toolsDir -ItemType Directory -Force | Out-Null
    }
    Push-Location $commandsPath
    try {
        & go build -o $commandsBinary .
        if ($LASTEXITCODE -ne 0) {
            throw "Failed to build eac binary (exit code $LASTEXITCODE)"
        }
    } finally {
        Pop-Location
    }

    # Save the original working directory as an environment variable
    # so Go commands can resolve relative paths correctly
    $env:CLIE_PWD = (Get-Location).Path

    # Run the pre-built binary
    try {
        & $commandsBinary @CommandParts

        # Propagate exit code
        if ($LASTEXITCODE -ne 0) {
            throw "Command failed with exit code $LASTEXITCODE"
        }
    } finally {
        # Clean up
        Remove-Item Env:\CLIE_PWD -ErrorAction SilentlyContinue
    }
}

<#
.SYNOPSIS
    Get available Go commands from go/cli/eac/

.DESCRIPTION
    Returns command information using the structured get commands output

.PARAMETER Simple
    Return only command names as strings instead of full objects

.EXAMPLE
    Get-GoSrcCommands
    # Returns: Command objects with name, description, etc.

.EXAMPLE
    Get-GoSrcCommands -Simple
    # Returns: @("show help", "show files", "show modules")
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
    Returns available command parts based on what's already been typed.
    For commands with Args: modules, returns module monikers.

.PARAMETER First
    First part already entered

.PARAMETER Second
    Second part already entered

.PARAMETER Third
    Third part already entered

.EXAMPLE
    Get-GoSrcCommandPart
    # Returns: @("get", "show", "build")

.EXAMPLE
    Get-GoSrcCommandPart -First "show"
    # Returns: @("files", "modules")

.EXAMPLE
    Get-GoSrcCommandPart -First "build"
    # Returns: module monikers for completion (since build accepts modules as args)
#>
function Get-GoSrcCommandPart {
    [CmdletBinding()]
    param(
        [string]$First,
        [string]$Second,
        [string]$Third
    )

    # Only read from environment variable cache - never make describe call
    # Cache is populated by importer.ps1 via New-TopLevelAliases
    if (-not $env:SRC_COMMANDS_DESCRIBE) {
        # Cache not available - return placeholder
        return @("no-auto-complete-cache-found")
    }

    try {
        $structure = $env:SRC_COMMANDS_DESCRIBE | ConvertFrom-Json -AsHashtable
    } catch {
        # Invalid cache - return placeholder
        return @("no-auto-complete-cache-found")
    }

    # Build the path from provided parts
    $parts = @()
    if ($First) { $parts += $First }
    if ($Second) { $parts += $Second }
    if ($Third) { $parts += $Third }

    $path = $parts -join " "

    # Check if the current command accepts modules as arguments
    $currentCmd = $structure.commands | Where-Object { $_.name -eq $path -or ($_.parts.Count -eq 1 -and $_.parts[0] -eq $First -and -not $Second) }
    if ($currentCmd -and $currentCmd.args -eq "modules" -and $structure.modules) {
        # Return module monikers for completion
        return $structure.modules
    }

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

    # If no subcommands found, check if first-level command accepts modules
    if ($suggestions.Count -eq 0 -and $First -and -not $Second) {
        $firstLevelCmd = $structure.commands | Where-Object { $_.parts.Count -eq 1 -and $_.parts[0] -eq $First }
        if ($firstLevelCmd -and $firstLevelCmd.args -eq "modules" -and $structure.modules) {
            return $structure.modules
        }
    }

    return $suggestions
}

<#
.SYNOPSIS
    Get unique location_0 (top-level) commands

.DESCRIPTION
    Extracts all unique first-level commands from the command structure
    (e.g., "show", "list", "commit", "describe")

.EXAMPLE
    Get-TopLevelCommands
    # Returns: @("show", "list", "commit", "describe")
#>
function Get-TopLevelCommands {
    [CmdletBinding()]
    param()

    $structure = Get-CommandStructure
    if (-not $structure) {
        return @()
    }

    $topLevel = @{}
    foreach ($cmd in $structure.commands) {
        if ($cmd.parts.Count -gt 0) {
            $first = $cmd.parts[0]
            $topLevel[$first] = $true
        }
    }

    return $topLevel.Keys | Sort-Object
}

<#
.SYNOPSIS
    Creates global aliases for all top-level commands

.DESCRIPTION
    Dynamically creates aliases for each location_0 command (show, list, etc.)
    Each alias invokes Invoke-GoSrcCommand with the first argument pre-filled.
    Enables direct usage like: show files, show modules, etc.

.EXAMPLE
    New-TopLevelAliases
    show files      # Instead of: run show files
    show modules    # Instead of: run show modules
#>
function New-TopLevelAliases {
    [CmdletBinding()]
    param()

    # This function populates cache via Get-TopLevelCommands -> Get-CommandStructure
    $topLevelCommands = Get-TopLevelCommands

    if ($topLevelCommands.Count -eq 0) {
        Write-Host "⚠️  No commands found" -ForegroundColor Yellow
        return
    }

    # Verify cache was populated in environment variable
    if (-not $env:SRC_COMMANDS_DESCRIBE) {
        Write-Host "⚠️  Warning: Command structure cache not populated - autocomplete may not work" -ForegroundColor Yellow
    }

    $createdAliases = @()
    $deletedAliases = @()
    $commandExamples = @{}

    # Get command structure to find examples for each top-level command
    $structure = $null
    if ($env:SRC_COMMANDS_DESCRIBE) {
        try {
            $structure = $env:SRC_COMMANDS_DESCRIBE | ConvertFrom-Json -AsHashtable
        } catch {
            # Ignore parse errors
        }
    }

    # Build examples map: find first subcommand for each top-level command
    if ($structure) {
        foreach ($cmd in $structure.commands) {
            if ($cmd.parts.Count -gt 0) {
                $firstPart = $cmd.parts[0]
                if (-not $commandExamples.ContainsKey($firstPart)) {
                    # Use the full command as example
                    $commandExamples[$firstPart] = $cmd.name
                }
            }
        }
    }

    foreach ($cmdName in $topLevelCommands) {
        # IDEMPOTENT: Always delete existing function/alias before creating new one
        # Use Test-Path instead of slow Get-Command
        $existingFunction = Test-Path "function:Global:$cmdName"
        $existingAlias = Test-Path "alias:$cmdName"

        if ($existingFunction) {
            Remove-Item "function:Global:$cmdName" -Force -ErrorAction SilentlyContinue
            $deletedAliases += $cmdName
        }
        if ($existingAlias) {
            Remove-Item "alias:$cmdName" -Force -ErrorAction SilentlyContinue
        }

        # Create a script block that calls Invoke-GoSrcCommand with first arg pre-filled
        # Use $args to capture all arguments without PowerShell parameter binding interfering with flags
        $scriptBlock = [scriptblock]::Create(@"
            # Extract Second, Third, Fourth from args
            `$Second = if (`$args.Count -gt 0) { `$args[0] } else { `$null }
            `$Third = if (`$args.Count -gt 1) { `$args[1] } else { `$null }
            `$Fourth = if (`$args.Count -gt 2) { `$args[2] } else { `$null }
            `$Rest = if (`$args.Count -gt 3) { `$args[3..(`$args.Count-1)] } else { @() }
            Invoke-GoSrcCommand -First '$cmdName' -Second `$Second -Third `$Third -Fourth `$Fourth -RemainingArgs `$Rest
"@)

        # Create the function
        Set-Item -Path "function:Global:$cmdName" -Value $scriptBlock -Force

        # Register tab completion for this function (only for Second parameter now)
        Register-ArgumentCompleter -CommandName $cmdName -ParameterName Second -ScriptBlock {
            param($commandName, $parameterName, $wordToComplete, $commandAst, $fakeBoundParameters)
            Get-GoSrcCommandPart -First $commandName
        }

        $createdAliases += $cmdName
    }

    Write-Host ""
    if ($deletedAliases.Count -gt 0) {
        Write-Host "🔄 Refreshed $($deletedAliases.Count) existing command(s)" -ForegroundColor Yellow
    }
    Write-Host "✅ Created $($createdAliases.Count) command aliases" -ForegroundColor Green
    Write-Host ""

    if ($createdAliases.Count -gt 0) {
        Write-Host "Registered commands and examples:" -ForegroundColor Cyan
        foreach ($alias in $createdAliases | Sort-Object) {
            $example = $commandExamples[$alias]
            if ($example) {
                Write-Host "  $alias" -NoNewline -ForegroundColor White
                Write-Host " → " -NoNewline -ForegroundColor DarkGray
                Write-Host "$example" -ForegroundColor Gray
            } else {
                Write-Host "  $alias" -ForegroundColor White
            }
        }
    }

    Write-Host ""
    Write-Host "Tab completion enabled:" -ForegroundColor Cyan
    Write-Host "  show <TAB>    # Suggests: files, modules, etc." -ForegroundColor Gray
    Write-Host ""
}

# Export all functions (including Get-CommandStructure for ArgumentCompleters)
Export-ModuleMember -Function Invoke-GoSrcCommand, New-TopLevelAliases, Get-TopLevelCommands, Get-GoSrcCommands, Get-GoSrcCommandPart, Get-CommandStructure

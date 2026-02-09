#Requires -Version 5.1
<#
.SYNOPSIS
    Setup external repositories to use local CLIE/EAC Docker images

.DESCRIPTION
    Configures external repositories to use locally built EAC Docker images for testing.

    NOTE: For development in the EAC repository itself, use importer.ps1 instead.
    This script is for testing the EAC extension in OTHER repositories.

    Performs: Docker image build, clie binary installation, and local configuration.

.PARAMETER TargetRepo
    Target repository path for configuration. Must be a different repository than EAC.

.PARAMETER SkipBuild
    Skip Docker image build (use if image already exists).

.PARAMETER SkipInstall
    Skip clie binary installation (use if already installed).

.PARAMETER SkipTest
    Skip validation tests after setup.

.PARAMETER ImageTag
    Docker image tag to use. Default is 'eac-ext:dev'.

.PARAMETER Force
    Force overwrite of existing configurations.

.EXAMPLE
    # Setup for external repository from within that repo
    cd C:\path\to\external-repo
    C:\path\to\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo .

.EXAMPLE
    # Setup for external repository from EAC repo
    C:\path\to\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo C:\path\to\external-repo

.EXAMPLE
    # Skip build if image already exists
    cd C:\path\to\external-repo
    C:\path\to\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo . -SkipBuild

.EXAMPLE
    # Setup multiple external repos (reuse image)
    C:\path\to\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo C:\repos\project-a
    C:\path\to\eac\scripts\pwsh\local-dev\setup.ps1 -TargetRepo C:\repos\project-b -SkipBuild
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory=$false)]
    [string]$TargetRepo = ".",

    [Parameter(Mandatory=$false)]
    [switch]$SkipBuild,

    [Parameter(Mandatory=$false)]
    [switch]$SkipInstall,

    [Parameter(Mandatory=$false)]
    [switch]$SkipTest,

    [Parameter(Mandatory=$false)]
    [string]$ImageTag = "eac-ext:dev",

    [Parameter(Mandatory=$false)]
    [switch]$Force
)

$ErrorActionPreference = "Stop"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Step {
    param(
        [string]$Message
    )
    Write-ColorOutput ""
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" "Cyan"
    Write-ColorOutput "  $Message" "Cyan"
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" "Cyan"
    Write-ColorOutput ""
}

function Test-Command {
    param([string]$Command)
    try {
        $null = Get-Command $Command -ErrorAction Stop
        return $true
    }
    catch {
        return $false
    }
}

function Get-RepositoryRoot {
    param([string]$StartPath)

    Push-Location $StartPath
    try {
        $root = git rev-parse --show-toplevel 2>&1
        if ($LASTEXITCODE -ne 0) {
            return $null
        }
        # Convert to Windows path if needed
        if ($root -match "^/[a-z]/") {
            $root = $root -replace "^/([a-z])/", '$1:/'
            $root = $root -replace "/", "\"
        }
        return $root
    }
    catch {
        return $null
    }
    finally {
        Pop-Location
    }
}

function Get-EacRepositoryPath {
    # Try to find EAC repository from script location
    $scriptDir = Split-Path -Parent $PSCommandPath
    # Script is at: eac/scripts/pwsh/local-dev/setup.ps1
    # So go up 3 levels: scripts/pwsh/local-dev -> scripts/pwsh -> scripts -> eac (root)
    $possibleRoot = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $scriptDir))

    if (Test-Path (Join-Path $possibleRoot "containers\eac-ext\Dockerfile")) {
        return $possibleRoot
    }

    # Try current directory
    $currentRoot = Get-RepositoryRoot -StartPath "."
    if ($currentRoot -and (Test-Path (Join-Path $currentRoot "containers\eac-ext\Dockerfile"))) {
        return $currentRoot
    }

    return $null
}

# Main script
Write-ColorOutput "═══════════════════════════════════════════════════════" "Green"
Write-ColorOutput "  CLIE External Repository Setup" "Green"
Write-ColorOutput "  Configure external repos to use local EAC Docker images" "Green"
Write-ColorOutput "═══════════════════════════════════════════════════════" "Green"
Write-ColorOutput ""
Write-ColorOutput "NOTE: For EAC repository development, use importer.ps1 instead" "Yellow"
Write-ColorOutput ""

# Phase 1: Prerequisites Check
Write-Step "Phase 1: Checking Prerequisites"

$prerequisites = @{
    "Docker" = Test-Command "docker"
    "Git" = Test-Command "git"
    "CLIE Binary" = Test-Command "clie"
}

$allPrerequisitesMet = $true
foreach ($prereq in $prerequisites.GetEnumerator()) {
    if ($prereq.Value) {
        Write-ColorOutput "  ✓ $($prereq.Key)" "Green"
    }
    else {
        if ($prereq.Key -eq "CLIE Binary" -and -not $SkipInstall) {
            Write-ColorOutput "  ○ $($prereq.Key) (will be installed)" "Yellow"
        }
        else {
            Write-ColorOutput "  ✗ $($prereq.Key)" "Red"
            $allPrerequisitesMet = $false
        }
    }
}

if (-not $allPrerequisitesMet) {
    Write-ColorOutput ""
    Write-ColorOutput "Error: Missing required prerequisites" "Red"
    Write-ColorOutput "Please install missing tools and try again" "Yellow"
    exit 1
}

# Find EAC repository
$eacRepo = Get-EacRepositoryPath
if (-not $eacRepo) {
    Write-ColorOutput ""
    Write-ColorOutput "Error: Cannot find EAC repository" "Red"
    Write-ColorOutput "This script must be run from the EAC repository or have access to it" "Yellow"
    exit 1
}

Write-ColorOutput ""
Write-ColorOutput "  EAC Repository: $eacRepo" "White"

# Resolve target repository
$targetRepoPath = Resolve-Path $TargetRepo -ErrorAction SilentlyContinue
if (-not $targetRepoPath) {
    Write-ColorOutput ""
    Write-ColorOutput "Error: Target repository not found: $TargetRepo" "Red"
    exit 1
}

$targetRepoRoot = Get-RepositoryRoot -StartPath $targetRepoPath
if (-not $targetRepoRoot) {
    Write-ColorOutput ""
    Write-ColorOutput "Error: Target path is not a git repository: $targetRepoPath" "Red"
    Write-ColorOutput "Initialize git first: git init" "Yellow"
    exit 1
}

# Check if target is the EAC repository itself
$eacRepoNormalized = [System.IO.Path]::GetFullPath($eacRepo).TrimEnd('\')
$targetRepoNormalized = [System.IO.Path]::GetFullPath($targetRepoRoot).TrimEnd('\')

if ($eacRepoNormalized -eq $targetRepoNormalized) {
    Write-ColorOutput ""
    Write-ColorOutput "Error: Target repository is the EAC repository itself" "Red"
    Write-ColorOutput ""
    Write-ColorOutput "This script is for setting up EXTERNAL repositories to test EAC." "Yellow"
    Write-ColorOutput ""
    Write-ColorOutput "For development in the EAC repository, use:" "Yellow"
    Write-ColorOutput "  .\scripts\pwsh\importer.ps1" "White"
    Write-ColorOutput ""
    Write-ColorOutput "Then load commands with:" "Yellow"
    Write-ColorOutput "  `$env:CLIE_COMMANDS_PATH = '.\go\cli\eac'" "White"
    Write-ColorOutput "  clie load-commands" "White"
    Write-ColorOutput ""
    exit 1
}

Write-ColorOutput "  Target Repository: $targetRepoRoot" "White"

# Phase 2: Build Docker Image
if (-not $SkipBuild) {
    Write-Step "Phase 2: Building Docker Image"

    $buildScript = Join-Path $eacRepo "scripts\pwsh\local-dev\build-local.ps1"
    if (-not (Test-Path $buildScript)) {
        Write-ColorOutput "Error: Build script not found: $buildScript" "Red"
        exit 1
    }

    Push-Location $eacRepo
    try {
        & $buildScript -Tag $ImageTag
        if ($LASTEXITCODE -ne 0) {
            Write-ColorOutput ""
            Write-ColorOutput "Error: Docker build failed" "Red"
            exit $LASTEXITCODE
        }
    }
    finally {
        Pop-Location
    }
}
else {
    Write-Step "Phase 2: Skipping Docker Build"
    Write-ColorOutput "  Using existing image: $ImageTag" "Yellow"

    # Verify image exists
    $imageExists = docker images --format "{{.Repository}}:{{.Tag}}" | Select-String -Pattern "^$($ImageTag.Replace(':','\:'))$"
    if (-not $imageExists) {
        Write-ColorOutput ""
        Write-ColorOutput "Warning: Image '$ImageTag' not found in local Docker" "Yellow"
        Write-ColorOutput "You may need to build it first" "Yellow"
    }
    else {
        Write-ColorOutput "  ✓ Image exists locally" "Green"
    }
}

# Phase 3: Install CLIE Binary
if (-not $SkipInstall) {
    Write-Step "Phase 3: Installing CLIE Binary"

    $installScript = Join-Path $eacRepo "scripts\pwsh\cli\install.ps1"
    if (-not (Test-Path $installScript)) {
        Write-ColorOutput "Error: Install script not found: $installScript" "Red"
        exit 1
    }

    & $installScript
    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput ""
        Write-ColorOutput "Error: CLIE installation failed" "Red"
        exit $LASTEXITCODE
    }
}
else {
    Write-Step "Phase 3: Skipping CLIE Installation"

    if (Test-Command "clie") {
        $clieVersion = clie --version 2>&1 | Select-String "Version:" | ForEach-Object { $_.Line }
        Write-ColorOutput "  ✓ CLIE is installed: $clieVersion" "Green"
    }
    else {
        Write-ColorOutput "  ✗ CLIE is not installed" "Red"
        Write-ColorOutput "  Remove -SkipInstall flag to install" "Yellow"
    }
}

# Phase 4: Create Local Configuration
Write-Step "Phase 4: Creating Local Configuration"

$initScript = Join-Path $eacRepo "scripts\pwsh\cli\init-local.ps1"
if (-not (Test-Path $initScript)) {
    Write-ColorOutput "Error: Init script not found: $initScript" "Red"
    exit 1
}

Push-Location $targetRepoRoot
try {
    if ($Force) {
        & $initScript -ImageTag $ImageTag -Force
    }
    else {
        & $initScript -ImageTag $ImageTag
    }

    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput ""
        Write-ColorOutput "Error: Configuration creation failed" "Red"
        exit $LASTEXITCODE
    }
}
finally {
    Pop-Location
}

# Phase 5: Validation
if (-not $SkipTest) {
    Write-Step "Phase 5: Validating Setup"

    # Set environment variable for test
    $env:CLIE_REPO_ROOT = $targetRepoRoot

    # Test clie command
    Write-ColorOutput "  Testing clie eac help..." "White"
    $testOutput = clie eac help 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "  ✓ Test successful" "Green"
    }
    else {
        Write-ColorOutput "  ✗ Test failed" "Red"
        Write-ColorOutput ""
        Write-ColorOutput "Test output:" "Yellow"
        Write-ColorOutput $testOutput "Gray"
    }
}
else {
    Write-Step "Phase 5: Skipping Validation"
}

# Summary
Write-ColorOutput ""
Write-ColorOutput "═══════════════════════════════════════════════════════" "Green"
Write-ColorOutput "  ✅ External Repository Setup Complete!" "Green"
Write-ColorOutput "═══════════════════════════════════════════════════════" "Green"
Write-ColorOutput ""

Write-ColorOutput "Configuration Summary:" "Cyan"
Write-ColorOutput "  Docker Image: $ImageTag" "White"
Write-ColorOutput "  EAC Repository: $eacRepo" "White"
Write-ColorOutput "  Target Repository: $targetRepoRoot" "White"
Write-ColorOutput "  Config File: $targetRepoRoot\.clie\clie.local.yml" "White"
Write-ColorOutput ""

Write-ColorOutput "Usage Instructions:" "Cyan"
Write-ColorOutput "  1. Navigate to your repository:" "White"
Write-ColorOutput "     cd $targetRepoRoot" "Gray"
Write-ColorOutput ""
Write-ColorOutput "  2. Set environment variable (required for each session):" "White"
Write-ColorOutput "     `$env:CLIE_REPO_ROOT = `"$targetRepoRoot`"" "Gray"
Write-ColorOutput ""
Write-ColorOutput "  3. Run EAC commands:" "White"
Write-ColorOutput "     clie eac help" "Gray"
Write-ColorOutput "     clie eac init" "Gray"
Write-ColorOutput "     clie eac show modules" "Gray"
Write-ColorOutput ""

Write-ColorOutput "Tips:" "Cyan"
Write-ColorOutput "  • The local config (.clie/clie.local.yml) is gitignored" "White"
Write-ColorOutput "  • For EAC repository development, use importer.ps1 instead of Docker" "White"
Write-ColorOutput "  • Rebuild the Docker image when you make changes in EAC repo:" "White"
Write-ColorOutput "    $eacRepo\scripts\pwsh\local-dev\build-local.ps1" "Gray"
Write-ColorOutput "  • Use -Force to overwrite existing configuration" "White"
Write-ColorOutput ""

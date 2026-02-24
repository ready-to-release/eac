#Requires -Version 5.1
<#
.SYNOPSIS
    Build local Docker image for EAC extension development

.DESCRIPTION
    Builds the eac-ext Docker image locally with a development tag.
    Supports single-platform (fast) and multi-platform builds.

.PARAMETER Tag
    Docker image tag. Default is 'eac-ext:dev'.

.PARAMETER Platform
    Target platform (linux/amd64 or linux/arm64). Default is linux/amd64.

.PARAMETER MultiPlatform
    Build for both linux/amd64 and linux/arm64 platforms.

.PARAMETER Push
    Push image to registry after build (requires buildx).

.PARAMETER NoBuildKit
    Disable Docker BuildKit (use legacy builder).

.EXAMPLE
    # Quick single-platform build
    .\build-local.ps1

.EXAMPLE
    # Multi-platform build
    .\build-local.ps1 -MultiPlatform

.EXAMPLE
    # Custom tag
    .\build-local.ps1 -Tag "eac-ext:my-feature"

.EXAMPLE
    # Build for ARM64
    .\build-local.ps1 -Platform "linux/arm64"
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory=$false)]
    [string]$Tag = "eac-ext:dev",

    [Parameter(Mandatory=$false)]
    [ValidateSet("linux/amd64", "linux/arm64", "")]
    [string]$Platform = "",

    [Parameter(Mandatory=$false)]
    [switch]$MultiPlatform,

    [Parameter(Mandatory=$false)]
    [switch]$Push,

    [Parameter(Mandatory=$false)]
    [switch]$NoBuildKit
)

$ErrorActionPreference = "Stop"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Test-Docker {
    try {
        $null = docker version 2>&1
        return $true
    }
    catch {
        return $false
    }
}

function Test-DockerBuildx {
    try {
        $null = docker buildx version 2>&1
        return $true
    }
    catch {
        return $false
    }
}

function Get-RepositoryRoot {
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
}

# Main script
Write-ColorOutput "EAC Extension - Local Docker Image Builder" "Cyan"
Write-ColorOutput ""

# Check Docker availability
if (-not (Test-Docker)) {
    Write-ColorOutput "Error: Docker is not available" "Red"
    Write-ColorOutput "Please install Docker Desktop from https://www.docker.com/products/docker-desktop" "Yellow"
    exit 1
}

Write-ColorOutput "✓ Docker is available" "Green"

# Find repository root
$repoRoot = Get-RepositoryRoot
if (-not $repoRoot) {
    Write-ColorOutput "Error: Not in a git repository" "Red"
    Write-ColorOutput "This script must be run from within the EAC repository" "Yellow"
    exit 1
}

Write-ColorOutput "✓ Repository root: $repoRoot" "Green"

# Change to repository root
Push-Location $repoRoot

try {
    # Verify Dockerfile exists
    $dockerfilePath = Join-Path $repoRoot "containers\eac-ext\Dockerfile"
    if (-not (Test-Path $dockerfilePath)) {
        Write-ColorOutput "Error: Dockerfile not found at $dockerfilePath" "Red"
        exit 1
    }

    Write-ColorOutput "✓ Dockerfile found" "Green"

    # Verify build context directory with pre-built binary exists
    $buildContextDir = Join-Path $repoRoot "out\build\eac\eac_go"
    if (-not (Test-Path $buildContextDir)) {
        Write-ColorOutput ""
        Write-ColorOutput "Error: Build context not found: $buildContextDir" "Red"
        Write-ColorOutput "The Dockerfile expects a pre-compiled Linux binary." "Yellow"
        Write-ColorOutput "Run the EAC pipeline build first to produce it:" "Yellow"
        Write-ColorOutput "  go run ./go/cli/eac build go/cli/eac" "Gray"
        exit 1
    }

    Write-ColorOutput "✓ Build context found: $buildContextDir" "Green"
    Write-ColorOutput ""

    # Determine build strategy
    Write-ColorOutput "Build Configuration:" "Cyan"
    Write-ColorOutput "  Tag: $Tag" "White"

    if ($MultiPlatform) {
        Write-ColorOutput "  Platforms: linux/amd64, linux/arm64 (multi-platform)" "White"

        # Check if buildx is available
        if (-not (Test-DockerBuildx)) {
            Write-ColorOutput ""
            Write-ColorOutput "Error: docker buildx is not available" "Red"
            Write-ColorOutput "Multi-platform builds require Docker Buildx" "Yellow"
            Write-ColorOutput "Install buildx or use single-platform build without -MultiPlatform flag" "Yellow"
            exit 1
        }

        Write-ColorOutput "  Builder: buildx (multi-platform)" "White"
    }
    else {
        if (-not $Platform) {
            $Platform = "linux/amd64"  # Default platform for Windows hosts
        }
        Write-ColorOutput "  Platform: $Platform (single-platform)" "White"

        if ($NoBuildKit) {
            Write-ColorOutput "  Builder: legacy (BuildKit disabled)" "White"
        }
        else {
            Write-ColorOutput "  Builder: BuildKit" "White"
        }
    }

    if ($Push) {
        Write-ColorOutput "  Push: Yes (will push to registry)" "Yellow"
    }
    else {
        Write-ColorOutput "  Push: No (local only)" "White"
    }

    Write-ColorOutput ""
    Write-ColorOutput "Building Docker image..." "Cyan"
    Write-ColorOutput ""

    # Build Docker command
    if ($MultiPlatform) {
        # Multi-platform build with buildx
        $buildCmd = "docker buildx build"
        $buildCmd += " --platform linux/amd64,linux/arm64"
        $buildCmd += " -t $Tag"
        $buildCmd += " -f containers/eac-ext/Dockerfile"

        if ($Push) {
            $buildCmd += " --push"
        }
        else {
            $buildCmd += " --load"  # Load into local Docker (only works for single platform in load mode)
            Write-ColorOutput "Warning: Multi-platform images cannot be loaded directly into Docker" "Yellow"
            Write-ColorOutput "The image will be built but not available in 'docker images' output" "Yellow"
            Write-ColorOutput "To use the image locally, build for a single platform instead" "Yellow"
            Write-ColorOutput ""
        }

        $buildCmd += " `"$buildContextDir`""
    }
    else {
        # Single-platform build
        if ($NoBuildKit) {
            # Legacy builder
            $env:DOCKER_BUILDKIT = "0"
            $buildCmd = "docker build"
        }
        else {
            # BuildKit enabled
            $env:DOCKER_BUILDKIT = "1"
            $buildCmd = "docker build"
        }

        $buildCmd += " --platform $Platform"
        $buildCmd += " -t $Tag"
        $buildCmd += " -f containers/eac-ext/Dockerfile"
        $buildCmd += " `"$buildContextDir`""
    }

    Write-ColorOutput "Command: $buildCmd" "Gray"
    Write-ColorOutput ""

    # Execute build
    Invoke-Expression $buildCmd

    if ($LASTEXITCODE -ne 0) {
        Write-ColorOutput ""
        Write-ColorOutput "Error: Docker build failed with exit code $LASTEXITCODE" "Red"
        exit $LASTEXITCODE
    }

    Write-ColorOutput ""
    Write-ColorOutput "✅ Docker image built successfully!" "Green"
    Write-ColorOutput ""

    # Verify image exists (only for single-platform local builds)
    if (-not $MultiPlatform -and -not $Push) {
        Write-ColorOutput "Verifying image..." "Cyan"
        $imageCheck = docker images --format "{{.Repository}}:{{.Tag}}" | Select-String -Pattern "^$($Tag.Replace(':','\:'))$"

        if ($imageCheck) {
            Write-ColorOutput "✓ Image available in local Docker" "Green"

            # Get image details
            $imageInfo = docker inspect $Tag | ConvertFrom-Json
            if ($imageInfo) {
                $imageSize = [math]::Round($imageInfo[0].Size / 1MB, 2)
                $imageCreated = [datetime]$imageInfo[0].Created

                Write-ColorOutput ""
                Write-ColorOutput "Image Details:" "Cyan"
                Write-ColorOutput "  Name: $Tag" "White"
                Write-ColorOutput "  Size: $imageSize MB" "White"
                Write-ColorOutput "  Created: $($imageCreated.ToString('yyyy-MM-dd HH:mm:ss'))" "White"
                Write-ColorOutput "  Platform: $Platform" "White"
            }
        }
        else {
            Write-ColorOutput "Warning: Image built but not found in 'docker images'" "Yellow"
        }
    }
    elseif ($Push) {
        Write-ColorOutput "✓ Image pushed to registry" "Green"
    }

    Write-ColorOutput ""
    Write-ColorOutput "Next Steps:" "Cyan"
    Write-ColorOutput "  1. Configure clie to use this local image:" "White"
    Write-ColorOutput "     Create .clie/clie.local.yml with:" "White"
    Write-ColorOutput ""
    Write-ColorOutput "     load_local: true" "Gray"
    Write-ColorOutput "     extensions:" "Gray"
    Write-ColorOutput "       - name: 'eac'" "Gray"
    Write-ColorOutput "         image: '$Tag'" "Gray"
    Write-ColorOutput "         load_local: true" "Gray"
    Write-ColorOutput "         image_pull_policy: 'Never'" "Gray"
    Write-ColorOutput ""
    Write-ColorOutput "  2. Test the image:" "White"
    Write-ColorOutput "     clie eac help" "Gray"
    Write-ColorOutput ""

}
finally {
    # Restore original directory
    Pop-Location
}

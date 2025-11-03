#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Build script for R2R CLI

.DESCRIPTION
    This script handles the build process for the R2R CLI, including:
    - Copying contract schemas to their embed locations
    - Running go generate
    - Building the CLI binary

.PARAMETER Clean
    Clean build artifacts before building

.PARAMETER Test
    Run tests after building

.PARAMETER OutputPath
    Path where the built binary should be placed (default: ./out/r2r.exe)

.EXAMPLE
    .\build-cli.ps1

.EXAMPLE
    .\build-cli.ps1 -Clean -Test
#>

[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$Test,
    [string]$OutputPath = "out/r2r.exe"
)

$ErrorActionPreference = "Stop"

Write-Host "🚀 R2R CLI Build Script" -ForegroundColor Cyan
Write-Host "=" * 60

# Get script directory (repo root)
$RepoRoot = $PSScriptRoot

# Define paths
$CliModulePath = Join-Path $RepoRoot "src/cli"
$ContractsPath = Join-Path $RepoRoot "contracts"
$ValidatorPath = Join-Path $CliModulePath "internal/validator"
$SchemaSourcePath = Join-Path $ContractsPath "cli/0.1.0/config.json"
$SchemaDestPath = Join-Path $ValidatorPath "config/schema.json"

# Step 1: Clean if requested
if ($Clean) {
    Write-Host "`n🧹 Cleaning build artifacts..." -ForegroundColor Yellow

    # Remove old binary
    if (Test-Path $OutputPath) {
        Remove-Item $OutputPath -Force
        Write-Host "  ✓ Removed old binary: $OutputPath"
    }

    # Remove validator schema copy
    if (Test-Path $SchemaDestPath) {
        Remove-Item $SchemaDestPath -Force
        Write-Host "  ✓ Removed old schema copy: $SchemaDestPath"
    }
}

# Step 2: Copy contract schema to validator embed location
Write-Host "`n📋 Copying contract schemas..." -ForegroundColor Yellow

# Ensure validator config directory exists
$ValidatorConfigDir = Join-Path $ValidatorPath "config"
if (-not (Test-Path $ValidatorConfigDir)) {
    New-Item -ItemType Directory -Path $ValidatorConfigDir -Force | Out-Null
    Write-Host "  ✓ Created directory: $ValidatorConfigDir"
}

# Copy schema
if (-not (Test-Path $SchemaSourcePath)) {
    Write-Error "Schema source not found: $SchemaSourcePath"
    exit 1
}

Copy-Item $SchemaSourcePath $SchemaDestPath -Force
Write-Host "  ✓ Copied schema: config.json -> validator/config/schema.json"

# Verify schema was copied
if (-not (Test-Path $SchemaDestPath)) {
    Write-Error "Failed to copy schema to: $SchemaDestPath"
    exit 1
}

$schemaSize = (Get-Item $SchemaDestPath).Length
Write-Host "  ✓ Schema file size: $schemaSize bytes"

# Step 3: Run go generate (if needed in the future)
Write-Host "`n🔧 Running go generate..." -ForegroundColor Yellow
Push-Location $CliModulePath
try {
    go generate ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Warning "go generate returned non-zero exit code (may be expected if no generators defined)"
    } else {
        Write-Host "  ✓ go generate completed"
    }
} finally {
    Pop-Location
}

# Step 4: Build the CLI
Write-Host "`n🔨 Building CLI binary..." -ForegroundColor Yellow
Push-Location $CliModulePath
try {
    # Ensure output directory exists
    $outputDir = Split-Path (Join-Path $RepoRoot $OutputPath) -Parent
    if (-not (Test-Path $outputDir)) {
        New-Item -ItemType Directory -Path $outputDir -Force | Out-Null
    }

    # Build with version info
    $outputFile = Join-Path $RepoRoot $OutputPath
    go build -o $outputFile .

    if ($LASTEXITCODE -ne 0) {
        Write-Error "Build failed with exit code $LASTEXITCODE"
        exit $LASTEXITCODE
    }

    Write-Host "  ✓ Built binary: $OutputPath" -ForegroundColor Green

    # Show binary info
    if (Test-Path $outputFile) {
        $binarySize = (Get-Item $outputFile).Length
        $binarySizeMB = [math]::Round($binarySize / 1MB, 2)
        Write-Host "  ✓ Binary size: $binarySizeMB MB"
    }
} finally {
    Pop-Location
}

# Step 5: Run tests if requested
if ($Test) {
    Write-Host "`n🧪 Running tests..." -ForegroundColor Yellow
    Push-Location $CliModulePath
    try {
        # Run tests with all build tags
        go test -tags="L0 L1 L2" ./... -v

        if ($LASTEXITCODE -ne 0) {
            Write-Error "Tests failed with exit code $LASTEXITCODE"
            exit $LASTEXITCODE
        }

        Write-Host "  ✓ All tests passed" -ForegroundColor Green
    } finally {
        Pop-Location
    }
}

# Step 6: Summary
Write-Host "`n" + ("=" * 60)
Write-Host "✅ Build completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Binary location: $OutputPath"
Write-Host ""
Write-Host "To run the CLI:"
Write-Host "  .\$OutputPath --help"
Write-Host ""

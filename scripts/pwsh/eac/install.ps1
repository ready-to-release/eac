#Requires -Version 5.1
<#
.SYNOPSIS
    EAC CLI Installer for Windows

.DESCRIPTION
    Downloads and installs the EAC CLI from GitHub releases.

.PARAMETER System
    Install to Program Files (requires Administrator). Default installs to user's local app data.

.PARAMETER Upx
    Install UPX-compressed binary (smaller download, slightly slower startup).

.PARAMETER Version
    Specific version to install. Default is 'latest'.

.PARAMETER InstallDir
    Custom installation directory. Overrides -System flag if specified.

.EXAMPLE
    # Install latest version for current user
    .\install.ps1

.EXAMPLE
    # Install specific version system-wide
    .\install.ps1 -System -Version "v1.0.0"

.EXAMPLE
    # One-liner from PowerShell
    irm https://raw.githubusercontent.com/ready-to-release/eac/main/scripts/pwsh/eac/install.ps1 | iex
#>

[CmdletBinding()]
param(
    [switch]$System,
    [switch]$Upx,
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "ready-to-release/eac"
$BinaryName = "eac.exe"
$UserInstallDir = Join-Path $env:LOCALAPPDATA "eac"
$SystemInstallDir = Join-Path $env:ProgramFiles "eac"

# Determine install directory (custom > system > user)
if (-not $InstallDir) {
    $InstallDir = if ($System) { $SystemInstallDir } else { $UserInstallDir }
}

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Get-LatestVersion {
    # Skip API call in test mode
    if ($env:__EAC_TEST_MOCK -eq "1") {
        Write-ColorOutput "Test mode: Using mock version eac/v0.0.0-test" "Gray"
        return "eac/v0.0.0-test"
    }

    try {
        # Fetch releases and find the latest eac/* release (monorepo has multiple release tags)
        $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases?per_page=20" -UseBasicParsing
        foreach ($release in $releases) {
            if ($release.tag_name -like "eac/*") {
                return $release.tag_name
            }
        }
        Write-ColorOutput "No eac release found" "Red"
        exit 1
    }
    catch {
        Write-ColorOutput "Failed to fetch latest version: $_" "Red"
        exit 1
    }
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Install-Binary {
    param(
        [string]$InstallVersion
    )

    # Binary filename for Windows (use UPX version if requested)
    if ($Upx) {
        $binaryFilename = "eac-windows-amd64-upx.exe"
        Write-ColorOutput "Binary type: UPX-compressed (smaller, slightly slower startup)" "Cyan"
    } else {
        $binaryFilename = "eac-windows-amd64.exe"
        Write-ColorOutput "Binary type: Standard" "Cyan"
    }
    $downloadUrl = "https://github.com/$Repo/releases/download/$InstallVersion/$binaryFilename"

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "eac-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

    try {
        # Download or create mock binary
        $tempFile = Join-Path $tempDir $BinaryName

        # Test mode: Use pre-built binary from out/build instead of downloading
        if ($env:__EAC_TEST_MOCK -eq "1") {
            # In test mode, still validate that version looks realistic
            # Reject obviously invalid versions like v999.999.999
            if ($InstallVersion -match "v999\.") {
                Write-ColorOutput "Test mode: Simulating download failure for invalid version" "Gray"
                Write-ColorOutput "Failed to download binary" "Red"
                Write-ColorOutput "Please check if the release exists: https://github.com/$Repo/releases/tag/$InstallVersion" "Yellow"
                exit 1
            }

            Write-ColorOutput "Test mode: Using pre-built binary from out/build (skipping download)" "Gray"

            # Use the actual built eac binary from the build output
            # This is available because eac-installer depends on eac module
            # Script lives at scripts/pwsh/eac/ — go up 3 levels to repo root,
            # then into out/build/eac/eac_go/ where the Go build places binaries
            $builtBinary = Join-Path $PSScriptRoot "..\..\..\out\build\eac\eac_go\eac-windows-amd64.exe"

            if (-not (Test-Path $builtBinary)) {
                Write-ColorOutput "Test mode: Pre-built binary not found at $builtBinary" "Red"
                Write-ColorOutput "Ensure eac module is built before running installer tests" "Yellow"
                exit 1
            }

            # Copy the actual binary
            Copy-Item -Path $builtBinary -Destination $tempFile -Force
        }
        else {
            # Real download in production mode
            Write-ColorOutput "Downloading EAC CLI $InstallVersion..." "Cyan"
            Write-ColorOutput "URL: $downloadUrl" "Gray"

            try {
                # Suppress progress bar for cleaner output and better performance
                $previousProgressPreference = $ProgressPreference
                $ProgressPreference = 'SilentlyContinue'
                try {
                    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile -UseBasicParsing
                }
                finally {
                    $ProgressPreference = $previousProgressPreference
                }
            }
            catch {
                Write-ColorOutput "Failed to download binary" "Red"
                Write-ColorOutput "Please check if the release exists: https://github.com/$Repo/releases/tag/$InstallVersion" "Yellow"
                exit 1
            }
        }

        # Create install directory if needed
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }

        # Install binary
        $targetPath = Join-Path $InstallDir $BinaryName
        Write-ColorOutput "Installing to $targetPath..." "Cyan"

        # Stop any running instances (skip in test mode to avoid killing the test orchestrator)
        if ($env:__EAC_TEST_MOCK -ne "1") {
            Get-Process -Name "eac" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
        }

        Move-Item -Path $tempFile -Destination $targetPath -Force

        Write-ColorOutput "Successfully installed EAC CLI $InstallVersion" "Green"
    }
    finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Add-ToPath {
    param(
        [string]$Directory
    )

    # Skip PATH modification during tests to avoid polluting system PATH
    if ($env:__EAC_TEST_NO_PATH_UPDATE -eq "1") {
        Write-ColorOutput "Skipping PATH modification (test mode)" "Gray"
        return
    }

    $scope = if ($System) { "Machine" } else { "User" }
    $currentPath = [Environment]::GetEnvironmentVariable("Path", $scope)

    if ($currentPath -notlike "*$Directory*") {
        Write-ColorOutput "" "White"
        Write-ColorOutput "Adding $Directory to PATH..." "Cyan"

        $newPath = "$currentPath;$Directory"

        try {
            [Environment]::SetEnvironmentVariable("Path", $newPath, $scope)
            $env:Path = "$env:Path;$Directory"
            Write-ColorOutput "Added to PATH successfully" "Green"
        }
        catch {
            Write-ColorOutput "Failed to add to PATH automatically" "Yellow"
            Write-ColorOutput "" "White"
            Write-ColorOutput "Please add the following to your PATH manually:" "Yellow"
            Write-ColorOutput "  $Directory" "White"
        }
    }
}

function Test-Installation {
    $binaryPath = Join-Path $InstallDir $BinaryName

    if (Test-Path $binaryPath) {
        Write-ColorOutput "" "White"
        Write-ColorOutput "Installation verified!" "Green"

        try {
            & $binaryPath --version
        }
        catch {
            Write-ColorOutput "Binary installed but version check failed" "Yellow"
        }
        return $true
    }
    return $false
}

# Main installation flow
function Main {
    Write-ColorOutput "EAC CLI Installer for Windows" "Green"
    Write-ColorOutput "" "White"

    # Check for admin if system install requested
    if ($System -and -not (Test-Administrator)) {
        Write-ColorOutput "System-wide installation requires Administrator privileges" "Red"
        Write-ColorOutput "Please run this script as Administrator or omit the -System flag" "Yellow"
        exit 1
    }

    Write-ColorOutput "Platform: Windows x64" "Cyan"
    Write-ColorOutput "Install location: $InstallDir" "Cyan"

    # Resolve version
    if ($Version -eq "latest") {
        Write-ColorOutput "Fetching latest version..." "Cyan"
        $Version = Get-LatestVersion
    }
    Write-ColorOutput "Version: $Version" "Cyan"

    # Install
    Install-Binary -InstallVersion $Version

    # Add to PATH
    Add-ToPath -Directory $InstallDir

    # Verify
    if (Test-Installation) {
        Write-ColorOutput "" "White"
        Write-ColorOutput "Done! Run 'eac --help' to get started." "Green"

        if (-not $System) {
            Write-ColorOutput "" "White"
            Write-ColorOutput "Note: You may need to restart your terminal for PATH changes to take effect." "Yellow"
        }
    }
}

Main

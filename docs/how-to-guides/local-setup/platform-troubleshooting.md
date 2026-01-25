# Platform Troubleshooting

Common platform-specific issues and their solutions for R2R and EAC development.

## Windows

### Windows Defender False Positives

Windows Defender's machine learning-based detection may flag Go test binaries as malware, typically with detection names like:

- `Trojan:Win32/Bearfoos.B!ml`
- `Trojan:Win32/Wacatac.B!ml`

The `!ml` suffix indicates this is a heuristic/ML detection, not a signature-based match. This is a **false positive** that occurs because:

1. **Temp directory location** - Go compiles test binaries to `%TEMP%\go-build*` directories
2. **Unsigned executables** - Test binaries aren't code-signed
3. **Behavioral patterns** - Go binaries have characteristics that trigger ML heuristics

#### Solution: Add Exclusions

**Via PowerShell (Recommended)**:

```powershell
# Run as Administrator

# Exclude Go build temp directory
Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp\go-build*"

# Exclude Go module cache (optional)
Add-MpPreference -ExclusionPath "$env:GOPATH\pkg"

# Exclude your project directory (optional)
Add-MpPreference -ExclusionPath "C:\projects\eac"
```

**Via Windows Security UI**:

1. Open **Windows Security** (search in Start menu)
2. Click **Virus & threat protection**
3. Under "Virus & threat protection settings", click **Manage settings**
4. Scroll to **Exclusions** and click **Add or remove exclusions**
5. Click **Add an exclusion** > **Folder**
6. Enter: `C:\Users\<username>\AppData\Local\Temp`

#### Verify Exclusions

```powershell
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

#### Restore Quarantined Files

If Defender already quarantined your test binary:

1. Open **Windows Security** > **Virus & threat protection**
2. Click **Protection history**
3. Find the quarantined item
4. Click **Actions** > **Restore**

### Path Length Limitations

Windows has a 260-character path limit by default. Deep directory structures can cause issues.

**Solution**: Enable long paths in Windows:

```powershell
# Run as Administrator
New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" `
  -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
```

Or via Group Policy:

1. Run `gpedit.msc`
2. Navigate to: Computer Configuration > Administrative Templates > System > Filesystem
3. Enable "Enable Win32 long paths"

### PowerShell Execution Policy

If PowerShell scripts fail to run:

```powershell
# Check current policy
Get-ExecutionPolicy

# Allow local scripts (recommended)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Docker Desktop Issues

**WSL 2 Backend Not Working**:

```powershell
# Enable WSL 2
wsl --install

# Set WSL 2 as default
wsl --set-default-version 2

# Restart Docker Desktop
```

**Insufficient Memory**:

Edit `.wslconfig` in your home directory:

```ini
[wsl2]
memory=8GB
processors=4
```

## macOS

### Gatekeeper Blocking Binaries

If macOS blocks the R2R binary:

```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/r2r
```

Or right-click the binary > Open > Open Anyway.

### Docker Desktop Memory

Increase Docker memory allocation:

1. Docker Desktop > Settings > Resources
2. Increase Memory (recommended: 8GB+)
3. Apply & Restart

### Certificate Issues

If you encounter SSL certificate errors:

```bash
# Update CA certificates
brew install ca-certificates
```

## Linux

### Docker Permission Denied

If Docker commands fail with permission errors:

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, or:
newgrp docker
```

### File Watcher Limits

If you hit inotify limits during development:

```bash
# Check current limit
cat /proc/sys/fs/inotify/max_user_watches

# Increase limit (temporary)
sudo sysctl fs.inotify.max_user_watches=524288

# Make permanent
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
```

### Missing Dependencies

Common missing packages:

```bash
# Debian/Ubuntu
sudo apt install build-essential git curl

# Fedora/RHEL
sudo dnf install @development-tools git curl

# Alpine
apk add build-base git curl
```

## Docker Issues (All Platforms)

### Image Pull Failures

```bash
# Check Docker is running
docker info

# Clear Docker cache
docker system prune -a

# Re-install extension
r2r cleanup --all
r2r install eac
```

### Container Startup Failures

```bash
# Check container logs
docker logs $(docker ps -lq)

# Run interactively for debugging
r2r interactive eac
```

### Disk Space Issues

```bash
# Check Docker disk usage
docker system df

# Clean up unused resources
docker system prune -a --volumes
```

## Network Issues

### Proxy Configuration

If behind a corporate proxy:

```bash
# Docker proxy settings
# Edit ~/.docker/config.json
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy:8080",
      "httpsProxy": "http://proxy:8080",
      "noProxy": "localhost,127.0.0.1"
    }
  }
}
```

### Firewall Blocking Docker

Ensure these ports are allowed:

- 2375/2376 (Docker daemon)
- 443 (HTTPS for image pulls)

## Reporting Issues

If you encounter issues not covered here:

1. Check [GitHub Issues](https://github.com/ready-to-release/eac/issues)
2. Include: OS version, Docker version, error message, steps to reproduce
3. Run diagnostics: `r2r verify` and include output

## Related Resources

- [Go Issue #37800](https://github.com/golang/go/issues/37800) - Windows Defender false positives
- [Docker Documentation](https://docs.docker.com/) - Official Docker docs
- [WSL Documentation](https://docs.microsoft.com/en-us/windows/wsl/) - Windows Subsystem for Linux

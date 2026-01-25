# Fix Windows Defender False Positives

## What You'll Accomplish

Resolve Windows Defender blocking Go test executables with false positive malware detections.

## Prerequisites

- Windows 10/11 with Windows Defender enabled
- Go development environment
- Administrator access (for adding exclusions)

## The Problem

Windows Defender's machine learning-based detection may flag Go test binaries as malware, typically with detection names like:

- `Trojan:Win32/Bearfoos.B!ml`
- `Trojan:Win32/Wacatac.B!ml`

The `!ml` suffix indicates this is a heuristic/ML detection, not a signature-based match. This is a **false positive** that occurs because:

1. **Temp directory location** - Go compiles test binaries to `%TEMP%\go-build*` directories
2. **Unsigned executables** - Test binaries aren't code-signed
3. **Behavioral patterns** - Go binaries have characteristics that trigger ML heuristics

This is a well-documented issue in the Go community and does not indicate actual malware.

## Steps

### 1. Verify It's a False Positive

Check the affected file path. If it matches this pattern, it's your Go test binary:

```
C:\Users\<username>\AppData\Local\Temp\go-build*\**\test.test.exe
```

### 2. Add Exclusion via Windows Security UI

1. Open **Windows Security** (search in Start menu)
2. Click **Virus & threat protection**
3. Under "Virus & threat protection settings", click **Manage settings**
4. Scroll to **Exclusions** and click **Add or remove exclusions**
5. Click **Add an exclusion** → **Folder**
6. Enter: `C:\Users\<username>\AppData\Local\Temp`

### 3. Add Exclusion via PowerShell (Recommended)

Open PowerShell as Administrator and run:

```powershell
# Exclude Go build temp directory
Add-MpPreference -ExclusionPath "$env:LOCALAPPDATA\Temp\go-build*"

# Exclude Go module cache (optional, if issues occur there)
Add-MpPreference -ExclusionPath "$env:GOPATH\pkg"

# Exclude your project directory (optional)
Add-MpPreference -ExclusionPath "C:\projects\eac"
```

### 4. Verify Exclusions

```powershell
Get-MpPreference | Select-Object -ExpandProperty ExclusionPath
```

### 5. Restore Quarantined Files (If Needed)

If Defender already quarantined your test binary:

1. Open **Windows Security** → **Virus & threat protection**
2. Click **Protection history**
3. Find the quarantined item
4. Click **Actions** → **Restore**

## Alternative: Exclude by Process

If you prefer more targeted exclusions:

```powershell
# Exclude Go compiler and test runner
Add-MpPreference -ExclusionProcess "go.exe"
Add-MpPreference -ExclusionProcess "go-build*.exe"
```

## Common Detection Names

| Detection Name | Type | Cause |
| --- | --- | --- |
| `Trojan:Win32/Bearfoos.B!ml` | ML heuristic | Go binary patterns |
| `Trojan:Win32/Wacatac.B!ml` | ML heuristic | Unsigned temp executables |
| `HackTool:Win32/AutoKMS` | False positive | Rarely, on certain Go tools |

## Security Considerations

While these are false positives, maintain good security hygiene:

- **Verify your dependencies** - Run `go mod verify` to check module checksums
- **Use trusted sources** - Only import from known repositories
- **Review new dependencies** - Check before adding new modules
- **Keep Go updated** - Use latest stable Go version

## Reporting False Positives to Microsoft

Help improve detection accuracy:

1. Go to [Microsoft Security Intelligence](https://www.microsoft.com/en-us/wdsi/filesubmission)
2. Submit the detected file as "Incorrectly detected as malware"
3. Include context that it's a Go test binary

## Next Steps

- [Run Tests for Module](./run-tests-for-module.md) - Continue testing after fixing
- [Debug Test Failures](./debug-test-failures.md) - Troubleshoot actual test issues

## Related Resources

- [Go Issue #37800](https://github.com/golang/go/issues/37800) - Windows Defender false positives
- [Microsoft Defender Exclusions](https://docs.microsoft.com/en-us/microsoft-365/security/defender-endpoint/configure-exclusions-microsoft-defender-antivirus)

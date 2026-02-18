# clie update self

Update the CLIE CLI binary to the latest released version.

## Syntax

```bash
clie update self [flags]
```

## Description

The `update self` command replaces the running `clie` binary with the latest version
downloaded from GitHub releases.

**What it does:**

1. Queries the GitHub releases API for all releases in the `ready-to-release/clie` repository
2. Filters releases with a `clie/` tag prefix and selects the highest semantic version
3. Identifies the correct binary asset for the current operating system and architecture
4. Downloads the binary using GitHub API authentication
5. Validates the downloaded file (checks size and executable header)
6. Replaces the current `clie` binary in-place

**Supported platforms** (binary name pattern: `clie-{os}-{arch}`):

| Platform      | Binary name              |
| ------------- | ------------------------ |
| Linux amd64   | `clie-linux-amd64`       |
| Linux arm64   | `clie-linux-arm64`       |
| macOS amd64   | `clie-darwin-amd64`      |
| macOS arm64   | `clie-darwin-arm64`      |
| Windows amd64 | `clie-windows-amd64.exe` |

## Prerequisites

GitHub authentication is required:

```bash
export GITHUB_USERNAME="your-username"
export GITHUB_TOKEN="ghp_..."
```

The command exits with an error if either variable is not set.

## Flags

| Flag      | Default | Description                                           |
| --------- | ------- | ----------------------------------------------------- |
| `--force` | false   | Update even if the current version matches the latest |

## Examples

### Update to Latest Version

```bash
clie update self
```

### Force Reinstall Current Version

```bash
clie update self --force
```

### Check Current Version First

```bash
clie version
clie update self
```

## Platform Notes

### Windows

On Windows, a running executable cannot be overwritten directly. The update command
renames the current binary to `clie.exe.bak`, copies the new binary into place, and
removes the backup. If the copy fails, the backup is restored.

### Linux / macOS

The new binary is copied directly over the current executable path (resolved via
symlink evaluation). File permissions are set to `0755`.

## Troubleshooting

**Error: GitHub authentication required**:

Set `GITHUB_USERNAME` and `GITHUB_TOKEN` before running the command.

**Error: No binary found matching `clie-{os}-{arch}`**

The release may not include a binary for your platform. Check the available assets
listed in the error output.

**Error: Downloaded file is not a valid executable**:

The download may have been corrupted or the GitHub API returned an error page instead
of the binary. Try again or check network connectivity.

## See Also

- [version command](version.md) — Display current CLIE CLI version
- [CLIE CLI Overview](index.md) — Command overview

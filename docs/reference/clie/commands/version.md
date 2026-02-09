# clie version

Display the version number of the CLIE CLI.

## Syntax

```bash
clie version
```

## Description

The `version` command displays the current CLIE CLI version.

Useful for:

- Verifying your installation
- Checking if updates are available
- Reporting issues with version information
- Confirming team members use compatible versions

## Output

```bash
clie version
```

**Example output:**

```text
clie version 1.2.3
```

## Examples

### Check Installed Version

```bash
clie version
```

## Version Format

The version number follows semantic versioning (semver):

```text
MAJOR.MINOR.PATCH
```

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## See Also

- [clie init](./init.md)
- [clie validate](./validate.md)
- [CLIE CLI Overview](./index.md)
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Installation guide

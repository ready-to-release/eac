# r2r version

Display the version number of the R2R CLI.

## Syntax

```bash
r2r version
```

## Description

The `version` command displays the current R2R CLI version. Useful for:

- Verifying your installation
- Checking if updates are available
- Reporting issues with version information
- Confirming team members use compatible versions

## Output

```bash
r2r version
```

**Example output:**

```text
r2r-cli version 1.2.3
```

## Examples

### Check Installed Version

```bash
r2r version
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

- [r2r init](./init.md)
- [r2r validate](./validate.md)
- [R2R CLI Overview](./index.md)
- [Quick Start Tutorial](../../../tutorials/getting-started/quick-start.md) - Installation guide

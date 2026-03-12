# Get modules

<!-- book:cmd get modules -->

Returns a YAML/JSON list of all module contracts in the repository, with optional filtering by versioning scheme and workflow presence.

## Usage

```bash
eac get modules [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--calver` | bool | Filter to only CalVer versioned modules |
| `--semver` | bool | Filter to only SemVer versioned modules |
| `--with-ci` | bool | Filter to modules that have a CI workflow |
| `--with-release` | bool | Filter to modules that have a release workflow |
| `--bundle` | bool | Filter to bundle modules (CalVer with release but no CI) |

When no flags are provided, all modules are returned. Flags can be combined for intersection filtering.

## Output Structure

Each module entry contains:

- `moniker` - Unique module identifier
- `type` - Module type (e.g., go, container, typescript, static)
- `root` - Root path relative to repository
- `depends_on` - List of dependency module monikers
- `versioning` - Versioning scheme (calver or semver)
- Additional metadata (books, files, workflows)

## Examples

```bash
# All modules
eac get modules

# CalVer modules that auto-release on CI pass
eac get modules --calver --with-ci

# CalVer bundle modules (release when deps change)
eac get modules --calver --bundle

# Traditional versioned modules
eac get modules --semver
```

## See Also

- [show modules](../show/modules.md) - Formatted table

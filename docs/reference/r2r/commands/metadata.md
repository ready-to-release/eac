# r2r metadata

Retrieve metadata information from an extension container.

## Syntax

```bash
r2r metadata <extension-name> [flags]
```

## Description

The `metadata` command retrieves metadata about an extension, including version information,
available commands, configuration schema, and other details.

**What it retrieves:**

- Extension version
- Available commands
- Configuration schema
- Dependencies
- Description
- Author information
- Build information

## Examples

### Basic Metadata

```bash
r2r metadata eac
```

**Output:**

```json
{
  "name": "eac",
  "version": "1.2.3",
  "description": "Everything-as-Code automation framework",
  "commands": [
    "build",
    "test",
    "validate",
    "scan",
    "release"
  ],
  "config_schema": ".eac/",
  "build_date": "2024-01-15T10:30:00Z",
  "commit_sha": "abc123def456"
}
```

### Check Extension Version

```bash
# Get version from metadata
r2r metadata eac | jq '.version'
```

**Output:**

```json
"1.2.3"
```

### List Available Commands

```bash
# Extract commands list
r2r metadata eac | jq '.commands[]'
```

**Output:**

```json
"build"
"test"
"validate"
"scan"
"release"
```

## Metadata Fields

### Standard Fields

| Field         | Type   | Description           |
| ------------- | ------ | --------------------- |
| `name`        | string | Extension name        |
| `version`     | string | Version number or SHA |
| `description` | string | Extension purpose     |
| `commands`    | array  | Available commands    |

### Optional Fields

| Field           | Type   | Description         |
| --------------- | ------ | ------------------- |
| `author`        | string | Maintainer name/org |
| `homepage`      | string | Documentation URL   |
| `license`       | string | Software license    |
| `build_date`    | string | ISO 8601 timestamp  |
| `commit_sha`    | string | Git commit SHA      |
| `dependencies`  | array  | Required tools      |
| `config_schema` | string | Configuration path  |

## See Also

- [R2R CLI Overview](index.md) - Command overview
- [install command](install.md) - Install extensions
- [interactive command](interactive.md) - Explore extensions interactively
- [list command](list.md) - List available extensions
- [version command](version.md) - Check R2R CLI version

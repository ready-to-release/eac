# Get valid-commands

<!-- book:cmd get commands -->

Returns all registered commands with their descriptions in structured format, sorted alphabetically. Includes both leaf commands and inferred parent commands.

## Usage

```bash
eac get valid-commands [--as-yaml|--as-json|--as-toml]
```

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output Fields

Each entry contains:

| Field | Description |
|-------|-------------|
| `command` | Full command name (e.g. `get modules`) |
| `description` | Short description of the command |

Parent commands (e.g. `get`, `show`) are automatically inferred from subcommand registrations.

## Examples

```bash
# List all commands as YAML
eac get valid-commands

# As JSON for scripting
eac get valid-commands --as-json
```

## See Also

- [show valid-commands](../show/valid-commands.md) - Human-readable output
- [help](../help/help.md) - Command help
- [get Commands](../categories/get.md)

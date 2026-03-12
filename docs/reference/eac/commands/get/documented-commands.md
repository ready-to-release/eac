# Get documented-commands

<!-- book:cmd get documented-commands -->

Scans all markdown files in the `docs/` folder for EAC commands appearing in bash, powershell, and pwsh code blocks. Returns a mapping of commands to their documentation locations.

## Usage

```bash
eac get documented-commands [--as-yaml|--as-json|--as-toml]
```

## Flags

| Flag | Description |
|------|-------------|
| `--as-yaml` | Output as YAML (default) |
| `--as-json` | Output as JSON |
| `--as-toml` | Output as TOML |

## Output Fields

| Field | Description |
|-------|-------------|
| `commands` | List of documented commands |
| `commands[].command` | Command name (e.g. `build`, `get modules`) |
| `commands[].occurrences` | List of locations where the command appears |
| `commands[].occurrences[].file` | Relative path to the markdown file |
| `commands[].occurrences[].line` | Line number |
| `commands[].occurrences[].language` | Code block language (`bash`, `powershell`, `pwsh`) |
| `commands[].occurrences[].snippet` | The actual command line |
| `summary.total_commands` | Number of unique commands found |
| `summary.total_occurrences` | Total number of command references |
| `summary.total_files` | Number of files containing commands |

## Examples

```bash
# Find all documented commands
eac get documented-commands

# As JSON for analysis
eac get documented-commands --as-json
```

## See Also

- [get valid-commands](./valid-commands.md)
- [show help](../show/help.md)
- [get Commands](../categories/get.md)

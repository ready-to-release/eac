# Show Valid Commands

<!-- book:cmd show valid-commands -->

Displays all registered commands in the EAC CLI in a sorted table with descriptions.

## Usage

```bash
eac show valid-commands
```

## Output Sections

A table with columns:

| Column | Description |
|--------|-------------|
| Command | Full command name (e.g., `show modules`, `build`) |
| Description | Short description of the command |

A footer row shows the total number of registered commands.

## Examples

```bash
# List all valid commands
eac show valid-commands
```

## See Also

- [get valid-commands](../get/valid-commands.md) - JSON output
- [get commands](../get/commands.md) - With tree structure
- [show help](./help.md)

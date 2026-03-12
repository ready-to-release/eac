# Get valid-commands

<!-- book:cmd get valid-commands -->

Returns all registered commands with their descriptions, sorted alphabetically. Includes both leaf commands and inferred parent commands (e.g., `pipeline` is inferred from `pipeline await-ci`).

## Usage

```bash
eac get valid-commands [flags]
```

## Output Structure

```yaml
- command: build
  description: Build modules
- command: get artifacts
  description: Get resolved artifacts for a module
- command: pipeline
  description: Parent command for pipeline subcommands
```

## Examples

```bash
eac get valid-commands
eac get valid-commands --as-json
```

## See Also

- [show valid-commands](../show/valid-commands.md) - Formatted table
- [get commands](./commands.md)

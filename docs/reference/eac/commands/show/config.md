# Show Config

<!-- book:cmd show config -->

Displays all loaded EAC repository configurations with defaults applied, including modules, component types, environments, and testing configuration.

## Usage

```bash
eac show config [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--verbose`, `-v` | bool | Show all config source files with layers and value counts |

## Output Sections

1. **Configuration Summary** - Table with columns: Config, Status, Items. Shows load status for modules, component_types, environments, and testing.
2. **Repository Settings** - Key-value table with type, trunk_branch, max_branch_age_days, parallelism settings.
3. **Modules** - Table with columns: Moniker, Type, Root.
4. **Package Types** - Table with columns: Type, Pool.
5. **Environments** - Table with columns: Name, Type, Description.
6. **Testing Tags** - Table with columns: Tag, Type, Description.
7. **Test Suites** - Table with columns: Moniker, Name, Description.

With `--verbose`, an additional **Configuration Sources** section appears first, showing layered config files (contract, user, personal) with their load status and value counts.

## Examples

```bash
# Show config summary
eac show config

# Show config with source file details
eac show config --verbose
```

## See Also

- [get config](../get/config.md) - JSON output
- [init](../init/init.md) - Configure AI provider

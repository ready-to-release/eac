# Show Components

<!-- book:cmd show components -->

Displays all components in a human-readable markdown report, grouped by module with phase support indicators and dependency information.

## Usage

```bash
eac show components [flags]
```

## Flags

| Flag | Type | Description |
|------|------|-------------|
| `--module <name>` | string | Filter to a specific module |
| `--type <type>` | string | Filter by component type (go, typescript, book, etc.) |
| `--buildable` | bool | Only components with build phase |
| `--lintable` | bool | Only components with lint phase |
| `--testable` | bool | Only components with test phase |
| `--scannable` | bool | Only components with scan phase |

## Output Sections

The report includes:

1. **Summary** - Statistics table with total components, module count, component type count, and counts of buildable/lintable/testable/scannable components.
2. **Components by Type** - Table showing each component type with count and which modules contain them.
3. **Components by Module** - Components grouped by module (in display order), with columns: Component, Type, Root, Build, Lint, Test, Scan (phase columns show checkmarks or dashes).
4. **Dependency Graph** - Two tables showing forward dependencies ("what each component needs") and reverse dependencies ("what depends on each component"). Only shown if dependencies exist.

## Examples

```bash
# Full component report
eac show components

# Components in a specific module
eac show components --module eac

# Only Go components
eac show components --type go

# Only buildable components
eac show components --buildable
```

## See Also

- [get components](../get/components.md) - Structured data output
- [show modules](./modules.md) - Module-level overview

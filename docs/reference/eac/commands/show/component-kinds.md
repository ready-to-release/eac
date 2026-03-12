# Show Component Kinds

<!-- book:cmd show component-kinds -->

Displays all component kinds (types) used across the repository, with a count of how many components use each kind.

## Usage

```bash
eac show component-kinds
```

## Output Sections

A single table with columns:

| Column | Description |
|--------|-------------|
| Component Type | The kind identifier (e.g., `go`, `typescript`, `book`, `dockerfile`) |
| Count | Number of components of this type across all modules |

A footer row shows the total number of distinct component types.

## Examples

```bash
# Show all component types with counts
eac show component-kinds
```

## See Also

- [show modules](./modules.md) - List all modules
- [get modules](../get/modules.md) - JSON output

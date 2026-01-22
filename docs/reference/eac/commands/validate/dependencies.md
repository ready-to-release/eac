# Validate dependencies

<!-- book:cmd validate dependencies -->

## How It Works

Validates Go module dependencies against module contracts:

- **go.mod Parsing**: Extracts dependencies from go.mod files
- **Contract Matching**: Verifies dependencies match repository contracts
- **Version Consistency**: Ensures consistent versions across modules
- **Circular Detection**: Identifies circular dependencies
- **Undeclared Dependencies**: Finds missing contract declarations

Prevents dependency mismatches and ensures consistent module graph.

## See Also

- [validate](./validate.md)
- [get dependencies](../get/dependencies.md)
- [validate Commands](../categories/validate.md)

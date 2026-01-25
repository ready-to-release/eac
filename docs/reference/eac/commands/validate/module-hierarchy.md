# Validate module-hierarchy

<!-- book:cmd validate module-hierarchy -->

## How It Works

Validates module dependency graph structure for correctness:

- **Acyclic Check**: Ensures no circular dependencies exist
- **Layer Validation**: Verifies modules respect architectural layers
- **Dependency Direction**: Checks dependencies flow in correct direction
- **Orphan Detection**: Identifies unreachable or isolated modules
- **Contract Consistency**: Validates dependency declarations match actual usage

Prevents architectural violations and ensures buildable dependency graph.

## See Also

- [validate](./validate.md)
- [get dependencies](../get/dependencies.md)
- [validate Commands](../categories/validate.md)

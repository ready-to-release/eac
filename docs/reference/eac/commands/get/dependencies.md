# Get dependencies

<!-- book:cmd get dependencies -->

## How It Works

Returns the module dependency graph in JSON format:

- **Graph Structure**: Adjacency list with module names as keys
- **Dependency Types**: Build-time dependencies from artifact consumption
- **Ordering**: Use with `get execution-order` to determine build sequence
- **Traversal**: Supports topological sorting for parallel build execution

The dependency graph is extracted from module contracts in `.eac/repository.yml`.

## See Also

- [show dependencies](../show/dependencies.md) - Formatted table
- [get execution-order](./execution-order.md)

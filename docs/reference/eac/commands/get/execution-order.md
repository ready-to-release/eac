# Get execution-order

<!-- book:cmd get execution-order -->

## How It Works

Calculates the optimal build order for modules based on dependencies:

- **Topological Sort**: Orders modules so dependencies build before dependents
- **Parallel Layers**: Groups independent modules that can build concurrently
- **Cycle Detection**: Fails if circular dependencies are detected
- **Build Optimization**: Minimizes total build time through parallelization

Returns ordered list of modules suitable for sequential or parallel execution.

## See Also

- [get dependencies](./dependencies.md)
- [get build-deps](./build-deps.md)
- [build](../build/build.md)

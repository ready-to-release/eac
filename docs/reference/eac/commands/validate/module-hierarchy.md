# Validate module-hierarchy

<!-- book:cmd validate module-hierarchy -->

Validates the module dependency graph for structural integrity, ensuring there are no broken references or cycles.

## Usage

```bash
eac validate module-hierarchy
```

## What It Checks

- **Non-existent module references** -- A module's `depends_on` lists a module that does not exist.
- **Bidirectional consistency** -- The `depends_on` and computed `used_by` relationships are consistent.
- **Circular dependencies** -- No cycles in the dependency graph (e.g., A -> B -> A).
- **Unreachable modules** -- All modules are reachable within the graph.

## Examples

```bash
eac validate module-hierarchy
```

## Common Errors

- **Module 'X' depends on 'Y', but 'Y' does not exist** -- A `depends_on` entry references an undefined module. Check for typos or add the missing contract.
- **Circular dependency: A -> B -> A** -- Two or more modules form a cycle. Refactor to break it.

## See Also

- [get modules](../get/modules.md)
- [show modules](../show/modules.md)
- [validate](./validate.md)

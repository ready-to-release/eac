# container

OCI container runtime port interfaces abstracting Docker, Podman, and
other container runtimes behind a single execution contract.

## Key Types

- **`ContainerPort`** -- Execute, build, pull, and inspect container images
- **`ContainerConfig`** -- Configuration for running a container
- **`BuildConfig`** -- Configuration for building a container image
- **`ContainerResult`** -- Execution result with exit code and output
- **`MountConfig`** -- Volume mount definition (source, target, readonly)
- **`ResourceConfig`** -- CPU, memory, and shared memory limits

## Patterns

- Hexagonal ports: `ContainerPort` is the single runtime-agnostic port
- Config structs: all parameters passed as plain structs, not interfaces

## Internal Structure

| File | Responsibility |
| --- | --- |
| ports.go | `ContainerPort` interface and all config/result structs |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `container` package (moniker: contracts-runtime) decouples tool
execution from any specific container runtime. The Docker adapter is the
primary implementation; commands and the scheduler interact only through
`ContainerPort`.

## Code Health

### Tech Debt
- No test files; add compile-time `var _ ContainerPort = ...` checks to catch interface drift

### Pain Points
- `ContainerConfig` has 15 fields with no builder or validation; callers must manually set correct combinations (e.g., `StdoutWriter` vs `LogWriter` precedence rules documented only in comments)
- `MountConfig` duplicates `core.MountDef` -- keeping two parallel mount types forces adapters to convert between them

### Optimization Opportunities
- Add a `ContainerConfigBuilder` with validation to enforce invariants (e.g., Image required, Timeout >= 0) -- moderate effort, prevents runtime surprises
- Unify `MountConfig` with `core.MountDef` or add a conversion helper -- low effort, reduces adapter boilerplate

# defaults/cmd/serialize

Standalone CLI tool that loads all module contracts from `repository.yml`, resolves their defaults, and serializes the fully-resolved configuration to a YAML output file. Used to snapshot module state for before/after comparison when changing defaults.

## Key Functions

| Function | Purpose |
|----------|---------|
| `main` | Entry point: loads modules from workspace, sorts by moniker, serializes resolved config to output file |
| `findRepoRoot` | Walks up directory tree looking for `.git` to locate repository root |

## Patterns

- **Snapshot tool**: Captures resolved module state at a point in time for diffing
- **Git-based root detection**: Uses `.git` directory presence to find workspace root
- **Component-oriented output**: Serializes components with roots and patterns rather than raw file lists

## Internal Structure

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point, module loading, YAML serialization |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/domain/modules` | `LoadFromWorkspace` to load and resolve all module contracts |

## Role in System

Development utility for validating that changes to module type defaults produce the expected resolved configuration. Run before and after a defaults change, then diff the output files to verify correctness.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- main.go is a simple, focused CLI tool under 100 lines
- No test files present; consider adding integration tests for snapshot generation
- Consider moving to top-level `tools/` directory to clarify it's not importable library code

# utils

Common utility functions shared across CLI packages.

## Key Functions

- `Contains` -- checks if a string exists in a string slice

## Patterns

- **Simple helpers**: provides basic utility functions that avoid importing larger packages for trivial operations

## Internal Structure

| File | Purpose |
|---|---|
| `slices.go` | `Contains` function for string slice membership check |

## Dependencies

None (leaf package).

## Role in System

Provides shared utility functions used across multiple CLI packages. Avoids duplication of common operations like slice membership checks.

## Code Health

### Tech Debt
- `Contains` duplicates functionality available in Go 1.21+ `slices.Contains`; could be replaced with the stdlib version

### Pain Points
- None identified; single-function package

### Optimization Opportunities
- Replace `Contains` with `slices.Contains` from the Go standard library and remove this package if no other functions are added (low effort)

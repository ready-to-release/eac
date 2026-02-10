# envconsts

Isolated environment variable constant definitions for the clie CLI module.

## Key Types

_No exported types. This package defines only constants._

## Patterns

- Module isolation: Constants intentionally duplicate values from `go/core/environments/constants.go` so that clie has no dependency on go/core
- Grouped by purpose: Constants organized into application config, debug/logging, testing, and Docker/container groups

## Internal Structure

| File         | Responsibility                                          |
| ------------ | ------------------------------------------------------- |
| constants.go | All CLIE_* environment variable constant definitions    |

## Dependencies

_None. This is a leaf package with zero imports._

## Role in System

The envconsts package exists to maintain clie's architectural isolation requirement. The clie module must remain fully isolated with no local dependencies on go/core or other modules, allowing it to be lightweight and independently distributable. These constants are used throughout clie's internal packages to reference environment variable names consistently without string literals.

## Code Health

### Tech Debt

_None identified. The intentional duplication from go/core is an architectural decision, not technical debt._

### Pain Points

_None identified._

### Optimization Opportunities

_None identified._

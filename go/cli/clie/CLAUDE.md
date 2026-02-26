# clie Module

Binary: `clie` — extension lifecycle manager.
Module: `github.com/ready-to-release/eac/go/cli/clie`

## ISOLATION CONSTRAINT (hard rule)

Production code in this module MUST NOT import any other workspace module:
- NO `go/core`
- NO `go/cli/eac`
- NO `go/mcp`
- NO `go/clibase`
- NO `go/adapters/*`

Permitted imports: `contracts/*` modules only (they have no workspace dependencies).

Test code MAY import local modules for test infrastructure.

## Environment Constants

Do NOT import `go/core/environments` for env var names. Use the local copies:
`go/cli/clie/internal/envconsts/constants.go`

Variables follow `CLIE_` prefix convention. Add new constants there if needed.
Duplication with `go/core/environments` is intentional and required.

## Import Cycle Resolution

If you encounter an import cycle involving clie, resolve by:
1. Parameter injection (preferred): pass values as function parameters
2. Adding to `internal/envconsts` if it is an environment variable constant
3. Never: creating a shared package that imports other workspace modules

## go.mod Rule

`go/cli/clie/go.mod` must not contain `replace` directives pointing at other
workspace modules.

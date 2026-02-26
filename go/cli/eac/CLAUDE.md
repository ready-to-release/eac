# eac CLI Module

Binary: `eac` — repository automation CLI.
Module: `github.com/ready-to-release/eac/go/cli/eac`

## Role

The eac binary is the primary human-facing CLI for repository operations.
It wraps the full command surface defined in `go/commands/*`.

## Key Entrypoints

- `main.go` — binary entrypoint
- `allcmds/` — command registration
- `imports_*.go` — command group imports (build, lint, scan, test, update, repository)

## MCP Server Relationship

The `go/mcp/commands` module also exposes these commands as MCP tools.
When adding a command to eac, consider whether it also needs MCP exposure.

## Dependency Chain

eac → commands → core → contracts
eac → adapters → core → contracts

No circular dependencies permitted. Run `go vet ./...` after any import change.

## Import Policy

May import from any non-isolated workspace module.
Do NOT import from `go/cli/clie` — that module is isolated.

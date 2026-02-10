# zap

Provides the `scan zap` command for running OWASP ZAP DAST (Dynamic Application Security Testing) scans via Docker containers.

## Key Types

None (command-only package).

## Key Functions

- **`ZAP()`** -- Entry point for the `scan zap` command; runs OWASP ZAP baseline scan against a target URL via Docker
- **`printZAPUsage()`** -- Display command usage and flag documentation

## Patterns

- `init()` registration: registers `ZAP` command function with the global registry
- Docker-based tool execution: uses Docker to run the ZAP scanner container with volume mounts for output

## Internal Structure

| File | Responsibility |
| --- | --- |
| zap.go | OWASP ZAP DAST scanning command via Docker container execution |

## Dependencies

- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `zap` sub-package adds DAST scanning capability to the `scan` command group. It runs the OWASP ZAP baseline scanner against configurable target URLs, producing security evidence files that feed into the risk assessment pipeline.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.

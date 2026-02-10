# evidence

Implements the `update evidence` command that collects and updates audit evidence files for compliance workflows.

## Key Types

None (command-only package).

## Key Functions

- **`UpdateEvidence()`** -- Entry point for the `update evidence` command; collects test and security evidence and updates OSCAL assessment results

## Patterns

- `init()` registration: registers command function with the global registry
- Evidence collection pipeline: gathers test results, security scans, and generates OSCAL-formatted evidence documents

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | Evidence collection and OSCAL assessment result generation |

## Dependencies

- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `evidence` sub-package automates the collection and formatting of audit evidence. It aggregates test results and security scan outputs into OSCAL assessment result documents, supporting compliance workflows and risk assessment.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.

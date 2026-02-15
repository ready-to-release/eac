# create/squash-message

Generates a comprehensive squash commit message from all branch commits compared to a base branch, using AI to synthesize the commit history and cumulative diff into a cohesive message suitable for GitHub PR squash merges.

## Key Types

- **`squashConfig`** -- Parsed command configuration with base branch name and debug flag
- **`SquashJSON`** -- Structured AI output matching `squash-message.schema.json` with type, scope, subject, body, changes, and modules
- **`ChangeItem`** -- Individual change entry with conventional-commit type, optional scope, and description
- **`SquashMessageValidator`** -- Validates formatted squash message header format and length constraints

## Patterns

- Two-phase AI generation: Phase 1 generates structured JSON validated against a JSON schema, then `FormatSquashMessage` converts JSON to plaintext commit format (no AI in Phase 2)
- Retry-based generation: uses `coreai.GenerateWithRetry` with schema validation for reliable structured output
- Three-tier prompt loading: command flag, team override (`.eac/templates/ai/`), system default (`templates/ai/`)
- Branch analysis pipeline: current branch, commit history, cumulative diff, diff stats, changed files with module enrichment
- Module-aware context: enriches changed files with module ownership for multi-module commit messages
- Dependency injection: `Deps` struct holds injectable git repo, exec command, and AI response

## Internal Structure

| File | Responsibility |
| --- | --- |
| command.go | Command entry point, configuration parsing, and orchestration pipeline |
| generator.go | AI prompt assembly, context building, message generation, and final assembly |
| formatter.go | `SquashJSON` type, `FormatSquashMessage` JSON-to-text conversion, auditor summary extraction |
| validator.go | `SquashMessageValidator` for header format and length validation |
| deps.go | Injectable dependencies for testing |

## Dependencies

- `adapters/ai` -- AI executor and provider registration
- `adapters/ai/providers` -- built-in AI provider registration
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/ai` -- prompt loading, retry generation, mock response support, and AI config
- `core/domain` -- JSON schema validator and validation error types
- `core/git` -- git repository interface for branch commits, diffs, and file lists
- `core/logging` -- structured logging
- `core/paths` -- contract schema paths and EAC directory constants
- `core/repository` -- repository root discovery and file-to-module enrichment

## Role in System

The `create squash-message` command generates publication-ready squash commit messages for `eac`, analyzing the full branch history and diff to produce a conventional-commit-formatted message with auditor summary, change breakdown, and module attribution. It complements `create commit-message` (for individual commits) by synthesizing an entire branch's work into a single cohesive message for PR squash merges.

## Code Health

### Tech Debt

- None identified

### Pain Points

- No test coverage for `command.go`, `deps.go`, `generator.go`

### Optimization Opportunities

- None identified

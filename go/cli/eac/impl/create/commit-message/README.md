# commit-message

Generates AI-powered conventional commit messages from staged git changes. Analyzes diffs, maps files to modules, produces a top-level summary plus per-module sections, and validates the result against a JSON schema contract.

## Key Types

- **`executionConfig`** -- Holds workspace root, debug mode, staged files, and diff
- **`CommitJSON`** -- Matches the commit-message JSON schema contract
- **`ModuleSectionJSON`** -- Matches the commit-message-module JSON schema contract
- **`ModuleChange`** -- Module-specific changes in multi-module commits
- **`GenerationResult`** -- AI generation output with provider metadata
- **`ValidationError`** -- Alias for internal contract validation error

## Patterns

- Phased pipeline: parse config, build context, generate summary, generate modules, assemble, validate
- Parallel module generation: concurrent goroutines with indexed results for order preservation
- Two-phase AI output: JSON generation with schema validation, then deterministic formatting
- Retry loop: regenerates on validation failure up to a configurable maximum
- Dependency injection: `Deps` struct holds injectable git repo, exec command, and AI response

## Internal Structure

| File | Responsibility |
| --- | --- |
| deps.go | Injectable dependencies struct with production defaults |
| commit-message.go | Command entry point and phased execution pipeline |
| context.go | Build prompt context from staged files, diffs, and modules |
| generation.go | AI prompt loading, execution, and retry with schema validation |
| formatter.go | Convert JSON schema output to conventional commit text format |
| assembly.go | Combine top-level and module sections, deduplicate, add stubs |
| parallel.go | Concurrent module section generation with order preservation |
| internal/ | Constants, contract validation, auto-cleanup, progress display |

## Dependencies

- `cli/eac/impl/create/commit-message/internal` -- validation rules and constants
- `clibase/registry` -- command registration
- `clibase/flags` -- flag validation and parsing
- `clibase/render` -- table builder for staged file display
- `adapters/ai` -- AI executor and provider registration
- `core/ai` -- AI config, contract loader, prompt loading, retry framework
- `core/domain` -- JSON schema validator
- `core/git` -- staged diff and commit operations
- `core/repository` -- repository root and file-module reports
- `core/logging` -- structured logging
- `core/paths` -- contract schema paths

## Role in System

The `commit-message` package is the AI-driven commit workflow in `eac`, bridging staged git changes with conventional commit format through structured AI generation. It ensures every commit message conforms to the project schema contract, includes module attribution, and can optionally auto-commit the result.

## Code Health

### Tech Debt

- `context.go` (319 lines) exceeds 300 lines

### Pain Points

- No test coverage for `command.go`, `config.go`, `context.go`, `formatter.go`, `generation.go`, `git.go`, `orchestrator.go`

### Optimization Opportunities

- None identified

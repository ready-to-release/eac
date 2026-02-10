# ai-summary

Implements the `update ai-summary` command that generates AI-powered analysis summaries for modules. Supports multiple analysis types (DSL, specs, source, docs) and integrates with the command framework for multi-module execution.

## Key Types

- **`Config`** -- Command configuration with analysis type, module filter, AI provider, model, and flags

## Key Functions

- **`UpdateAISummary()`** -- Entry point for the `update ai-summary` command
- **`parseConfig()`** -- Parse command-line flags into `Config` struct
- **`isValidAnalysisType()`** -- Validate analysis type against allowed values (dsl, specs, source, docs)
- **`printUsage()`** -- Display command usage and available analysis types
- **`RunAISummaryWithFramework()`** -- Execute AI summary generation using the command framework with hooks
- **`ResolveUnitSpecs()`** -- Resolve which units of work need AI summary generation
- **`aiSummaryUnitWorker()`** -- Worker function that generates AI summary for a single unit of work
- **`extractAnalysisType()`** -- Extract analysis type from command arguments

## Patterns

- Command framework integration: uses `cmdframework` with `AfterInit` and `AfterResolve` hooks for multi-module execution
- Multiple analysis types: supports DSL, specs, source, and docs analysis with type-specific prompts
- AI provider abstraction: configurable AI provider and model via flags
- Unit-of-work pattern: each module's analysis runs as an individual unit of work

## Internal Structure

| File | Responsibility |
| --- | --- |
| command.go | Command entry point, config parsing, and usage display |
| framework.go | Framework-based execution with unit resolution and AI summary worker |

## Dependencies

- `clibase/cmdframework` -- multi-module command execution framework
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `ai-summary` sub-package automates the generation of AI-powered analysis documents for modules. It integrates with the broader update command group to produce design documentation, specification summaries, and code analysis reports using configurable AI providers.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.

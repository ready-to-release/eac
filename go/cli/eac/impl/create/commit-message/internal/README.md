# commitmessage (internal)

Commit message validation, cleanup, and verification against the commit-message contract, supporting both single-module and multi-module commit formats.

## Key Types

- **`CommitMessageValidator`** -- Implements `domain.Validator` for commit message validation, delegating to `VerifyCommitMessageContract` with affected module context
- **`TopLevelValidator`** -- Validates top-level commit message structure (header, auditor-summary, body) without module sections
- **`ModuleSectionValidator`** -- Validates individual module sections (module name, dashes, subject line, body text)
- **`CommitMessageContract`** -- Represents the structure.yml contract with version, semantic types, subject line format, constraints, and markdown rules
- **`ValidationError`** -- Alias to `domain.ValidationError`
- **`contentFixContext`** -- Internal state holder for the multi-pass content fixing pipeline

## Key Functions

- `VerifyCommitMessageContract` -- Main validation entry point; validates header, top-level body, module sections, subject lines, code blocks, and line lengths
- `AutoCleanup` -- Performs automatic fixes on commit messages before validation (normalizes spacing, fixes content, closes code blocks, ensures trailing blank line)
- `GetCleanupStats` -- Returns a list of what was cleaned (e.g., "Removed trailing period from title")
- `LoadContractFromConfig` -- Loads commit-message contract from ai-config.yml, returning minimal contract info since actual validation uses JSON schema
- `WithProgress` -- Wraps a function with whimsical progress updates shown during AI generation
- `WithAngryProgress` -- Wraps a function with frustrated progress updates during auto-fix phases

## Patterns

- **Multi-pass cleanup pipeline**: `AutoCleanup` uses three phases -- normalize spacing, fix content, final cleanup -- to create a stable foundation before applying content-level fixes
- **State machine for content fixing**: `contentFixContext` tracks code block state, body section state, module headers, and special fields, processing each line through a chain of handler methods
- **Lazy-compiled regexes**: Conventional commit and module subject line regexes use `sync.Once` for thread-safe lazy initialization
- **Validator composition**: Different aspects of validation (header, body, modules, subject lines, code blocks, line length) are split into separate `validate*` functions and composed in `VerifyCommitMessageContract`

## Internal Structure

| File | Responsibility |
| --- | --- |
| verifier.go | Core verification logic: CommitMessageContract type, LoadContractFromConfig, VerifyCommitMessageContract, and all validate* sub-functions (header, body, modules, subject lines, code blocks, line length), plus helper functions (isModuleName, isDashesLine, isSectionSeparator) |
| validator.go | CommitMessageValidator adapter implementing domain.Validator interface |
| toplevel_validator.go | TopLevelValidator for validating top-level commit output (header, auditor-summary, body) |
| module_validator.go | ModuleSectionValidator for validating individual module sections |
| cleanup.go | AutoCleanup pipeline with contentFixContext state machine, line wrapping, and code block closing |
| constants.go | Line length, formatting, and contract version constants; standard semantic commit types list |
| progress.go | Progress display utilities with whimsical and angry status lines for long-running AI operations |

## Dependencies

- `go/core/ai` -- AI config loader for loading commit-message type configuration
- `go/core/domain` -- ValidationError type and error code constants
- `go/core/logging` -- Component logger for progress messages

## Role in System

This package is the internal engine for the `create commit-message` command. It validates generated commit messages against the commit-message contract, ensures multi-module commits have proper module sections with subject lines, and automatically cleans up common formatting issues. The validators are used both during AI generation (to retry on validation failures) and for final output verification.

## Code Health

### Tech Debt
- constants.go:37 StandardCommitTypes is a hardcoded list with a comment noting it should be loaded from the contract at runtime

### Pain Points
- None identified

### Optimization Opportunities
- None identified

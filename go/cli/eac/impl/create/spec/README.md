# spec

Generates Gherkin specifications from natural language descriptions using AI, with contract-based validation and OSCAL control tag context.

## Key Types

- **`SpecsConfig`** -- Holds command configuration: description, debug/force flags, module, output path, prompt path, template root, and loaded EAC config
- **`Deps`** -- Injectable dependencies for testing; supports injecting AI responses and a custom git repository factory

## Key Functions

- `CreateSpec` -- Entry point for the `create spec` command; orchestrates prompt building, AI generation, validation, path determination, and file writing
- `loadAndBuildPrompt` -- Loads specification contract from YAML and builds the full AI prompt with contract context
- `generateAndClean` -- Generates AI output with retry, applies anti-corruption filtering, and validates Gherkin structure
- `determineAndValidateOutputPath` -- Determines output file path from user input or feature name extraction, with path traversal security validation
- `loadPromptWithFallback` -- Implements three-tier prompt loading: custom path (--prompt flag), team override, and system default
- `buildContractBasedPrompt` -- Loads contract files and builds comprehensive AI prompt combining template, contract, tags config, and OSCAL controls
- `loadModuleControlsContext` -- Loads OSCAL profile for a module and formats available controls as a markdown table for the AI prompt

## Patterns

- **Command registration via init()**: Uses `registry.Register(CreateSpec)` for automatic command discovery
- **Dependency injection for testing**: The `Deps` struct provides mock AI responses and git repo access for test isolation
- **Contract-driven generation**: The AI prompt includes the specification contract, anti-corruption rules, and tags configuration to guide output format
- **Three-tier prompt loading**: Custom path > team override > system default, with clear logging of which source was used
- **Security-validated output paths**: `ValidateOutputPath` prevents path traversal by ensuring output stays within the repository
- **OSCAL-aware context**: When a module has a risk profile, available controls are formatted as a markdown table with IDs, titles, and descriptions for the AI to use when generating control tags

## Internal Structure

| File | Responsibility |
| --- | --- |
| create.go | Command entry point, configuration parsing, prompt building, AI generation orchestration, output path determination, file writing, and OSCAL control context loading |
| deps.go | Injectable Deps struct with AI response and git repo factory for test dependency injection |

## Dependencies

- `go/cli/eac/impl/specs` -- Feature name extraction and output path determination
- `go/adapters/ai` -- AI executor creation and adapter wrapping
- `go/adapters/ai/providers` -- Built-in AI provider registration
- `go/cli/eac/internal/risk/oscal` -- OSCAL profile/catalog loading for control context
- `go/clibase/flags` -- Shared flag parsing and validation
- `go/clibase/registry` -- Command registration and workspace root discovery
- `go/core/ai` -- Retry framework, AI config loading, contract loader, mock response support, prompt template building
- `go/core/config` -- EAC config loading for path resolution and tags configuration
- `go/core/domain` -- Validation error formatting and critical error counting
- `go/core/domain/reports` -- Module contract reports for module validation
- `go/core/git` -- LazyRepo for git repository access
- `go/core/logging` -- Component logger for info, warning, and error output
- `go/core/paths` -- Contract and default version path resolution
- `go/core/repository` -- Repository root discovery
- `go/core/validation/formats/gherkin` -- Gherkin format validator for AI output

## Role in System

This package implements the `create spec` command, which transforms natural language descriptions into properly formatted Gherkin specifications. It uses AI with contract-based validation to ensure generated specs follow the project's specification standards, include appropriate tags, and optionally reference OSCAL controls when a risk profile exists. The generated `.feature` files are saved to the `specs/` directory organized by module.

## Code Health

### Tech Debt

- `create.go` (701 lines) exceeds 300 lines

### Pain Points

- No test coverage for `create.go`, `deps.go`

### Optimization Opportunities

- None identified

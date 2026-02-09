# create/design

Generates Structurizr DSL workspace files for a module by analyzing its source code using AI. Produces system context, container, and component views, then validates the generated DSL against Structurizr CLI via Docker before writing to the specs directory.

## Key Types

- **`DesignConfig`** -- Command configuration with module name, source path, output path, prompt path, and flags for debug, force, and skip-validation
- **`createFlags`** -- Parsed command-line flags for debug, force, output path, prompt path, and skip-validation

## Patterns

- AI-powered code analysis: reads module source code and generates architecture documentation via AI with retry and validation
- Composite validation: quick regex-based validation followed by full Docker-based Structurizr CLI validation
- Three-tier prompt loading: command flag (`--prompt`), team override (`.eac/templates/`), system default (`templates/`)
- Contract-based prompt building: loads AI contract and combines with module context and generation instructions
- Module validation via contracts: validates module moniker against module registry before generation
- Docker pre-flight check: verifies Docker availability before expensive AI generation
- Mock injection for testing: package-level mock AI response and git repository providers

## Internal Structure

| File | Responsibility |
| --- | --- |
| create.go | Command entry point, config parsing, prompt building, AI generation with validation, output writing |
| mocks.go | Test infrastructure: mock AI response and git repository injection |

## Dependencies

- `adapters/ai` -- AI executor and provider registration
- `adapters/ai/providers` -- built-in AI provider registration
- `cli/eac/impl/design` -- output handler for console messages
- `cli/eac/impl/design/helper` -- module name validation
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/ai` -- contract loader, prompt building, retry generation, and AI config
- `core/config` -- EAC configuration for specs path resolution
- `core/domain/reports` -- module contract loading and registry access
- `core/logging` -- structured logging
- `core/paths` -- defaults version and specs directory constants
- `core/repository` -- repository root discovery
- `core/validation/formats/structurizr` -- composite Structurizr DSL validator (quick + Docker)

## Role in System

The `create design` command provides AI-powered architecture documentation generation for `eac`, enabling teams to bootstrap Structurizr DSL workspaces from source code analysis. It integrates with the design validation pipeline to ensure generated workspaces are syntactically valid, and outputs to the standard specs directory structure for use with `serve design` and `validate design` commands.

## Code Health

### Tech Debt
- All production logic concentrated in create.go (530 lines); prompt building, config parsing, and validation could be split
- Global mutable `mockAIResponse` and `gitRepoProvider` in mocks.go for test injection follow the same pattern seen across create/ subcommands
- `buildContractBasedPrompt` (create.go:352-424, ~72 lines) mixes file I/O, string assembly, and error handling

### Pain Points
- Only two production files (create.go + mocks.go) makes the package feel monolithic despite well-decomposed functions
- Docker pre-flight check in `checkDockerAvailability` silently skips validation when Docker is unavailable, which may mask real failures

### Optimization Opportunities
- Extract prompt-loading and contract-building logic into a separate file to keep create.go focused on orchestration (high feasibility, natural boundary)
- Consolidate mock injection pattern (mockAIResponse, gitRepoProvider) with commit-message and squash-message into a shared test-support package (moderate feasibility, cross-package change)

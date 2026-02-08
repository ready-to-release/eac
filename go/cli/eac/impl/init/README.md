# init

Initializes EAC project configuration by creating the `.eac` directory structure, scanning for modules, and optionally configuring AI providers. Supports intelligent re-initialization that preserves user customizations while updating auto-generated content.

## Key Types

- **`agentConfig`** -- Holds AI provider settings: provider name, env var, model, and endpoint
- **`tokenConfig`** -- Holds actual AI and Git API token values for personal config
- **`ScanResult`** -- Repository scan output containing detected `ModuleInfo` entries
- **`ModuleInfo`** -- Detected module with name, root path, language, build tool, and key files
- **`ConfigGenerator`** -- Interface for generating repository.yml from scan results
- **`RuleBasedGenerator`** -- Generates config using predefined heuristics per language
- **`AIGenerator`** -- Generates config using AI-powered analysis with template prompts
- **`ExistingConfig`** -- Detected existing configuration state for re-initialization decisions
- **`MergeResult`** -- Tracks modules added, updated, and removed during re-initialization
- **`APIKeyResolution`** -- Result of multi-source API key lookup with source tracking

## Patterns

- Two-mode execution: first-time initialization vs. intelligent re-initialization
- Strategy pattern: `ConfigGenerator` interface with rule-based and AI-backed implementations
- AI-with-fallback: attempts AI generation first, falls back to rule-based on failure
- Multi-language scanner: detects Go, Python, Rust, TypeScript, .NET, and Java modules via manifest files
- Config merge: preserves user edits (names, versioning, dependencies) while updating AI-generated content
- API key resolution: prioritized lookup from environment variable, personal config, then error with guidance
- Dependency injection: `Deps` struct with `defaultDeps()` for production, test-specific `Deps` for mocking

## Internal Structure

| File | Responsibility |
| --- | --- |
| init.go | Command entry point, flag parsing, directory creation, config file generation |
| deps.go | `Deps` struct and `defaultDeps()` for injectable dependencies (git repo, AI executor) |
| scanner.go | Filesystem walk detecting modules by package manager files (go.mod, Cargo.toml, etc.) |
| config_generator.go | `ConfigGenerator` interface with `RuleBasedGenerator` and `AIGenerator` implementations |
| ai_generation.go | AI prompt template loading, scan data preparation, and AI executor integration |
| reinit.go | Re-initialization: detect existing config, re-scan, merge, and write updated files |
| merge.go | Config merging logic preserving user customizations across modules and components |
| api_key.go | Multi-source API key resolution with prioritized lookup |

## Dependencies

- `adapters/ai` -- AI executor for config generation prompts
- `adapters/ai/providers` -- Built-in AI provider registration and default models
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/config` -- configuration loading and `RepositoryConfig` types
- `core/environments` -- environment variable constants
- `core/git` -- lazy git repository provider with mock injection
- `core/logging` -- structured logging
- `core/paths` -- EAC config directory and file path resolution
- `core/repository` -- repository root discovery

## Role in System

The `init` package is the onboarding entry point for `eac-cli`, bootstrapping a project's `.eac` configuration from scratch or updating it as the repository evolves. Its scanner detects modules across multiple languages, and its strategy-based config generation produces repository.yml either through heuristics or AI-enhanced analysis, making it the foundation that all other commands depend on for configuration.

## Code Health

### Tech Debt
- init.go is 757 lines with `Init()` spanning ~233 lines; it handles flag parsing, directory creation, config writing, template copying, and YAML generation
- ~~Global mutable `var aiExecutor` and `gitRepoProvider` for test injection~~ (resolved: replaced with `Deps` struct)
- `generateWithScan` (init.go:669) mixes AI generation, fallback logic, and file writing in ~84 lines

### Pain Points
- init.go concentrates too many responsibilities: agent config, directory setup, file generation, and template copying could each be separate files
- Six different YAML generation functions in init.go (repository, books, environments) make the file hard to navigate

### Optimization Opportunities
- Extract YAML generation functions into a dedicated `generators.go` file to keep init.go under 400 lines (high feasibility, no cross-cutting concerns)
- Good test coverage exists (scanner, merge, reinit, api_key all have dedicated tests); focus new tests on init.go orchestration edge cases (moderate feasibility)

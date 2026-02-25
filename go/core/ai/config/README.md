# ai/config

Loads and manages unified AI configuration from `ai-config.yml`. Provides a three-tier configuration loading strategy (user override, system default, contracts fallback) and a type-specific `ContractLoader` that resolves prompts with a four-tier priority: custom path, team override, system default, embedded fallback.

## Key Types

| Type | Purpose |
|------|---------|
| `AIConfigLoader` | Loads unified AI configuration with three-tier fallback (user repo, system defaults, contracts defaults) |
| `ContractLoader` | Type-specific API wrapping `AIConfigLoader` with prompt loading, contract conversion, and priority resolution |
| `AIConfig` | Root configuration structure from `ai-config.yml` containing version, defaults, and per-type configs |
| `AITypeConfig` | Per-type config specifying output format, data references, and retry strategy |
| `RetryStrategyConfig` | Retry strategy configuration (type, focus categories, max attempts) |
| `AIDefaults` | Reserved defaults inherited by all AI types (currently unused) |

## Key Functions

| Function | Purpose |
|----------|---------|
| `LoadAIConfig` | Convenience function to load AI config from workspace root |
| `ExtractStringList` | Extracts a string array from a `map[string]interface{}` |
| `ExtractString` | Extracts a string value from a `map[string]interface{}` |
| `ExtractInt` | Extracts an int value from a `map[string]interface{}` |
| `ExtractBool` | Extracts a bool value from a `map[string]interface{}` |
| `ExtractMap` | Extracts a nested map from a `map[string]interface{}` |

## Patterns

- **Three-tier fallback**: `AIConfigLoader.Load()` tries user repo, system default, then contracts default
- **Four-tier prompt priority**: `ContractLoader.LoadPrompt()` tries custom path, team override, system default, embedded fallback
- **Container awareness**: Uses `CLIE_CONTAINER_ROOT` env var to resolve system paths in container vs dev mode
- **Lazy loading with caching**: `AIConfigLoader` loads config once and caches the result
- **Type-safe extraction helpers**: `ExtractString`, `ExtractInt`, etc. for safely pulling values from untyped maps

## Internal Structure

| File | Purpose |
|------|---------|
| `loader.go` | `AIConfigLoader` and `ContractLoader` implementations, extraction helpers |
| `types.go` | Configuration type definitions (`AIConfig`, `AITypeConfig`, `RetryStrategyConfig`) |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/domain` | `Contract` type for backward-compatible contract loading |
| `core/paths` | Path constants (`CLIEDir`, `EACDir`, `AIConfigFilename`, `ContainerRootEnv`) |

## Role in System

Central configuration provider for all AI generation commands. Commands like `create-spec`, `create-design`, and `get-commit-message` use `ContractLoader` to load their type-specific config, prompts, and validation schemas. The three-tier config loading ensures AI works in container deployments, local development, and user-customized setups.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- loader.go (391 lines): Extraction helpers (`ExtractString`, `ExtractInt`, `ExtractBool`, `ExtractMap`, `ExtractStringList`) at lines 326-384 could be replaced with a generic function using Go generics, reducing 60 lines to ~20
- types.go is concise at 40 lines

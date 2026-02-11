# conf

Layered YAML configuration loading, merging, validation, and repository root discovery.

## Key Types

- **`Config`** -- Top-level configuration: registry, defaults, environment, extensions, load_local
- **`Extension`** -- Single extension definition with image, env, volumes, ports, resources
- **`Registry`** -- Registry settings: default host, authentication, timeout, cache TTL
- **`Defaults`** -- Default values for pull policy, resource limits, environment
- **`Environment`** -- Global environment variables and secrets
- **`EnvVar`** -- Name/value pair with optional required flag for passthrough vars
- **`ValidationError`** -- Aggregated validation error with list of messages
- **`TestConfig`** -- Isolated test configuration helper

## Key Functions

- **`InitConfig`** -- Finds and loads base config, merges override files, stores RootDir
- **`LoadConfig`** -- Loads a single YAML file via Viper, unmarshals to Global, validates
- **`MergeConfigFile`** -- Merges an override file into the existing Global config
- **`FindRepositoryRoot`** -- Walks up directory tree looking for `.git` folder
- **`ValidatePinnedExtensions`** -- Checks extensions use SHA-pinned tags (CI enforcement)

## Patterns

- Layered config: Base `clie.yml` merged with `.local.yml`, `.personal.yml`, `.dev.yml` overrides in priority order
- Global singleton: `conf.Global` holds the merged configuration, `conf.RootDir` holds repository root
- Test isolation: `InitConfig()` panics in test environments; tests use `TestConfig` and `NewTestConfig()`
- Regex validation: Docker image refs, env var names, memory limits validated with compiled patterns
- Config file candidates: Priority-ordered discovery including `CLIE_CONFIG_PATH` env var and username-specific files

## Internal Structure

| File                  | Responsibility                                                     |
| --------------------- | ------------------------------------------------------------------ |
| config.go             | Config/Extension types, Global singleton, LoadConfig, mergeConfigs |
| config-validation.go  | validateConfig with regex patterns for images, env, ports, resources |
| config-extensions.go  | ValidatePinnedExtensions, checkLatestTags, GHCR tag fetching       |
| init.go               | InitConfig, FindRepositoryRoot, findConfigFile, config candidates  |
| errors.go             | ConfigError type hierarchy with structured errors and suggestions  |
| testing.go            | TestConfig helper, NewTestConfig, WriteConfigFile, ResetGlobalConfig |

## Dependencies

- `internal/logging` -- Debug and warning log output
- `internal/envconsts` -- Environment variable constant names

## Role in System

The conf package is the central configuration hub for the entire clie CLI. Every command calls `conf.InitConfig()` to populate the `conf.Global` singleton before accessing extension definitions, registry settings, or environment variables. The layered merge system allows team-wide base configs to be overridden by individual developers without modifying committed files.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- `config-extensions.go` -- `ValidatePinnedExtensions` mixes CI validation logic with cache loading and registry HTTP calls, making it hard to test in isolation.

### Optimization Opportunities

- The `mergeConfigs` function uses field-by-field merge logic that must be updated whenever new fields are added to Config or Extension types. A reflection-based or code-generated merge would reduce maintenance burden.

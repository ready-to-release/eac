# validation/formats/structurizr

Validates Structurizr DSL syntax using three modes: quick (fast regex-based syntax checks),
full (Docker-based validation using the official Structurizr CLI), and composite (quick first, then full if quick passes).

## Key Types

| Type                 | Purpose                                                                           |
| -------------------- | --------------------------------------------------------------------------------- |
| `Validator`          | Base interface extending `validation.Validator`                                   |
| `ValidatorMode`      | Enum for validation modes (`ModeQuick`, `ModeFull`, `ModeComposite`)              |
| `QuickValidator`     | Fast regex-based syntax validation (~100ms, no Docker) for AI generation feedback |
| `DockerValidator`    | Full validation using Structurizr CLI in Docker (~6-7 seconds)                    |
| `CompositeValidator` | Runs quick validation first, then Docker validation if quick passes               |
| `ContainerProvider`  | Function type returning a `ContainerPort` for Docker execution                    |

## Key Functions

| Function                        | Purpose                                                             |
| ------------------------------- | ------------------------------------------------------------------- |
| `NewQuickValidator`             | Creates a quick validator with default identifier patterns          |
| `NewQuickValidatorWithContract` | Creates a quick validator with contract-derived identifier patterns |
| `NewDockerValidator`            | Creates a Docker-based validator for a specific module              |
| `NewCompositeValidator`         | Creates a composite validator combining quick and full validation   |
| `SetContainerProvider`          | Sets the global container provider for Docker validation            |

## Patterns

- **Tiered validation**: Quick mode for AI generation (fast feedback), full mode for `validate` command (comprehensive), composite for `create` command (fast + comprehensive)
- **Contract-aware patterns**: `QuickValidator` can load identifier regex from contract data
- **Docker isolation**: `DockerValidator` mounts workspace content into Structurizr CLI container
- **Skip-on-failure optimization**: `CompositeValidator` can skip expensive Docker validation when quick syntax checks fail

## Internal Structure

| File           | Purpose                                                                                                   |
| -------------- | --------------------------------------------------------------------------------------------------------- |
| `validator.go` | `Validator` interface, `ValidatorMode` enum                                                               |
| `quick.go`     | `QuickValidator` with regex-based syntax checks (workspace structure, braces, identifiers, relationships) |
| `docker.go`    | `DockerValidator` using Structurizr CLI container, `ContainerProvider`, `SetContainerProvider`            |
| `composite.go` | `CompositeValidator` combining quick and Docker validation with configurable skip behavior                |

## Dependencies

| Package                       | Purpose                                               |
| ----------------------------- | ----------------------------------------------------- |
| `core/domain`                 | `Contract` for contract-driven identifier patterns    |
| `core/validation`             | `ValidationError`, error codes, `ErrorFormatter`      |
| `contracts/container-runtime` | `ContainerPort` for Docker execution (in `docker.go`) |

## Role in System

Used by the AI generation pipeline (`ai/generation`) for quick validation during `create-design`, and by the `validate-design` command for full Docker-based validation. The composite mode serves the `create` command workflow where fast feedback is needed during generation but comprehensive validation is desired for the final output.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: `DockerValidator` depends on a global `defaultContainerProvider` variable (`docker.go` line 29) which requires initialization via `SetContainerProvider` before use.
- **Optimization Opportunities**: None identified.

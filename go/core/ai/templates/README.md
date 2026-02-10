# ai/templates

Provides Go `text/template` based prompt rendering for AI generation commands. Converts a contract and custom data into a fully rendered prompt string using standard Go template syntax.

## Key Types

| Type | Purpose |
|------|---------|
| `PromptData` | Template data holder with `Contract` (YAML string), `ContractRaw` (raw object), and `Custom` (key-value map) |

## Key Functions

| Function | Purpose |
|----------|---------|
| `BuildPromptWithTemplate` | Parses and executes a Go template with contract YAML and custom data, returning the rendered prompt |

## Patterns

- **Template rendering**: Uses `text/template` with `{{.Contract}}`, `{{.ContractRaw}}`, and `{{.Custom.Key}}` placeholders
- **Contract serialization**: Marshals `domain.Contract.RawData` to YAML for template injection
- **Helpful error messages**: Template execution failures include hints about common issues (missing fields, syntax errors)

## Internal Structure

| File | Purpose |
|------|---------|
| `builder.go` | `PromptData` type and `BuildPromptWithTemplate` function |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/domain` | `Contract` type for accessing raw contract data |

## Role in System

Used by AI generation commands that need template-based prompt construction. Commands load a prompt template (via `ai/config.ContractLoader`), then call `BuildPromptWithTemplate` to inject contract data and command-specific custom values before passing the rendered prompt to the generation pipeline.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.

# validation/formats/gherkin

Validates AI-generated Gherkin specifications against structure, tag, naming, and GxP compliance rules. Performs comprehensive checks on feature declarations, rule blocks, scenario nesting, indentation, file size constraints, verification tags, and tag format patterns.

## Key Types

| Type        | Purpose                                                                                           |
| ----------- | ------------------------------------------------------------------------------------------------- |
| `Validator` | Validates Gherkin content against structure and tag rules from a contract and `TestingTagsConfig` |

## Key Functions

| Function       | Purpose                                                                    |
| -------------- | -------------------------------------------------------------------------- |
| `NewValidator` | Creates a Gherkin validator with a contract and testing tags configuration |

## Patterns

- **Contract-driven validation**: Rules and patterns are loaded from the specification contract
- **Tag configuration integration**: Verification tag requirements and format patterns come from `testing-tags.yml`
- **Multi-category checks**: Validates structure (Feature, Rule, Scenario nesting), tags (presence, format), naming (conventions), and size (rule/scenario counts)
- **Enhanced error reporting**: Uses `validation.ErrorFormatter` for structured error messages with expected/actual comparisons

## Internal Structure

| File           | Purpose                                                                                                     |
| -------------- | ----------------------------------------------------------------------------------------------------------- |
| `validator.go` | `Validator` type with all validation logic (~700 lines covering structure, tags, naming, size, indentation) |

## Dependencies

| Package           | Purpose                                                                   |
| ----------------- | ------------------------------------------------------------------------- |
| `core/config`     | `TestingTagsConfig` for verification tag requirements and format patterns |
| `core/domain`     | `Contract` for validation rules                                           |
| `core/paths`      | Path constants for defaults version                                       |
| `core/validation` | `ValidationError`, error codes, `ErrorFormatter`                          |

## Role in System

Used by the AI generation pipeline (`ai/generation`) to validate Gherkin output from `create-spec` and related commands. Ensures generated specifications meet structural requirements before being saved, triggering retries when validation fails.

## Code Health

- **Tech Debt**: The validator is a single ~700-line file. Could benefit from splitting into per-category validator functions in separate files (structure, tags, naming, size).
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.

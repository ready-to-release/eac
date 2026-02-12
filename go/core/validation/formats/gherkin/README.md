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

| File                   | Purpose                                                                      |
| ---------------------- | ---------------------------------------------------------------------------- |
| `validator.go`         | `Validator` type, `NewValidator`, and core validation orchestration           |
| `validator_tags.go`    | Tag validation: presence, format patterns, verification tag requirements     |
| `validator_features.go`| Structure validation: Feature, Rule, Scenario nesting, naming, size checks   |
| `validator_contract.go`| Contract-driven validation rules and pattern loading                          |

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

### Tech Debt
- `validator.go` is 405 lines
- `validator_tags.go` is 306 lines

### Pain Points
- None identified

### Optimization Opportunities
- Consider further splitting validation logic by concern (structure, tags, naming, size)

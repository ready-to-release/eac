# Create spec

<!-- book:cmd create spec -->

## How It Works

The command uses AI generation with Gherkin format support:

- **Format**: Generates .feature files in valid Gherkin syntax (Feature, Scenario, Given/When/Then)
- **Validation**: Multi-layer validation ensures quality:
  - Gherkin syntax validation (proper feature file structure)
  - Step definition validation (steps match implemented step functions)
  - Quality standards validation (follows specification best practices)
- **Retry**: If any validation fails, AI receives error feedback and regenerates improved specifications
- **Customization**: Uses three-tier prompt system for domain-specific specification styles

Supported formats: Gherkin BDD specifications with scenario outlines and examples tables.

## Custom Prompts

The spec generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.eac/templates/ai/specs/specs.md` (team-wide customization)
3. **System Default**: `templates/ai/specs/specs.md` (fallback)

See [commit-message](../get/commit-message.md#custom-prompts) for detailed customization guide.

## See Also

- [validate specs](../validate/specs.md)
- [test](../test/test.md)
- [create Commands](../categories/create.md)

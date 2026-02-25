# Create design

<!-- book:cmd create design -->

## How It Works

The command uses AI generation with Structurizr DSL format support:

- **Format**: Generates workspace.dsl files in valid Structurizr DSL syntax
- **Validation**: Validates DSL syntax using Structurizr CLI (requires Docker)
- **Retry**: If syntax validation fails, AI receives error feedback and regenerates corrected DSL
- **Customization**: Uses three-tier prompt system for team-specific design approaches

Supported formats: Structurizr DSL for C4 model diagrams (system context, container, component).

## Custom Prompts

The design generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.eac/templates/ai/design/design.md` (team-wide customization)
3. **System Default**: `templates/ai/design/design.md` (fallback)

See [commit-message](../get/commit-message.md#custom-prompts) for detailed customization guide.

## See Also

- [update design](../update/design.md)
- [validate design](../validate/design.md)
- [serve design](../serve/design.md)
- [create Commands](../categories/create.md)

# Create design

<!-- book:cmd create design -->

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

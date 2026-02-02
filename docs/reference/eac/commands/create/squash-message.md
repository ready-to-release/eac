# Create squash-message

<!-- book:cmd create squash-message -->

## How It Works

The command uses AI generation for squash merge commit messages:

- **Format**: Generates conventional commit message summarizing all commits in branch
- **Validation**: Validates message format and structure
- **Retry**: If validation fails, AI receives error feedback and regenerates improved message
- **Customization**: Uses three-tier prompt system for team-specific message styles

Supported formats: Conventional commits with comprehensive body summarizing branch changes.

## Custom Prompts

The squash message generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.eac/templates/ai/commit/squash.md` (team-wide customization)
3. **System Default**: `templates/ai/commit/squash.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide.

## See Also

- [work merge](../work/merge.md)
- [create pr](./pr.md)
- [create commit-message](./commit-message.md)
- [create Commands](../categories/create.md)

# Create squash-message

<!-- book:cmd create squash-message -->

## Custom Prompts

The squash message generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/commit/squash.md` (team-wide customization)
3. **System Default**: `templates/ai/commit/squash.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:

```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [work merge](../work/merge.md)
- [create pr](./pr.md)
- [create commit-message](./commit-message.md)
- [create Commands](../categories/create.md)

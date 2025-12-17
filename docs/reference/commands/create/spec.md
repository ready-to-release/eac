# Create spec

<!-- book:cmd create spec -->

## Custom Prompts

The spec generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/specs/specs.md` (team-wide customization)
3. **System Default**: `templates/ai/specs/specs.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:

```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [validate specs](../validate/specs.md)
- [test](../test/test.md)
- [create Commands](../categories/create.md)

# Create risk-profile

<!-- book:cmd create risk-profile -->

## Custom Prompts

The risk profile generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/risk/profile.md` (team-wide customization)
3. **System Default**: `templates/ai/risk/profile.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide or:

```bash
cat .r2r/eac/templates/ai/README.md
```

## See Also

- [create risk-assess](./risk-assess.md)
- [validate risk-profile](../validate/risk-profile.md)
- [scan](../categories/scan.md)
- [create Commands](../categories/create.md)

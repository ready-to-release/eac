# Create risk-profile

<!-- book:cmd create risk-profile -->

## How It Works

The command uses AI generation with OSCAL profile format support:

- **Format**: Generates OSCAL 1.1.3 profile documents defining control baselines
- **Validation**: Validates against OSCAL JSON schema
- **Retry**: If validation fails, AI receives error feedback and regenerates corrected profile
- **Customization**: Uses three-tier prompt system for organization-specific security baselines

Supported formats: OSCAL profile JSON documents for security control selection and tailoring.

## Custom Prompts

The risk profile generation supports **three-tier prompt system** for customization:

1. **Command Flag**: `--prompt /path/to/custom.md` (highest priority)
2. **Team Override**: `.r2r/eac/templates/ai/risk/profile.md` (team-wide customization)
3. **System Default**: `templates/ai/risk/profile.md` (fallback)

See [commit-message](./commit-message.md#custom-prompts) for detailed customization guide.

## See Also

- [create risk-assess](./risk-assess.md)
- [validate risk-profile](../validate/risk-profile.md)
- [scan](../categories/scan.md)
- [create Commands](../categories/create.md)

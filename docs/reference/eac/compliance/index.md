# Compliance Reference

Technical reference for EAC compliance features including risk controls, OSCAL integration, and control tag validation.

## In This Section

| Reference                           | Description                                          |
| ----------------------------------- | ---------------------------------------------------- |
| [Control Tags](./control-tags.md)   | Control tag format, validation, and OSCAL mapping    |
| [Risk Controls](./risk-controls.md) | Risk profile creation and assessment commands        |

## Quick Reference

```bash
# Validate control tags against OSCAL catalog
r2r eac validate control-tags

# Validate OSCAL files
r2r eac validate risk-catalog
r2r eac validate risk-profile

# Create risk artifacts
r2r create risk-profile assessment.md
r2r eac create risk-assess --profile specs/.risk-controls/risk-profile.json
```

## Related Documentation

- [Validate Command Reference](../commands/validate/index.md) - Full validation command options
- [GxP Tagging (Conceptual)](../../../explanation/specifications/compliance/gxp-tagging.md) - Compliance tagging principles

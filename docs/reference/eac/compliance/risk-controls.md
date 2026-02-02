# Risk Controls Reference

> **CLI commands for risk control validation and evidence collection**

This page documents the EAC commands for working with OSCAL-based risk controls.

## Quick Reference

```bash
# Validate control tags against OSCAL catalog
eac validate control-tags

# Validate OSCAL catalog documents
eac validate risk-catalog

# Validate OSCAL profile documents
eac validate risk-profile

# Create risk profile from assessment document
r2r create risk-profile assessment.md

# Generate assessment results with evidence
eac create risk-assess --profile specs/.risk-controls/risk-profile.json
```

---

## Control Tag Validation

### `validate control-tags`

Checks all `@control:` tags in specifications reference valid controls from the OSCAL catalog.

```bash
eac validate control-tags
```

**Output**: Reports invalid control IDs with file locations.

**Example output**:

```text
Validating control tags...
  specs/auth-service/login.feature:15 - invalid control: @control:ac-99 (not in catalog)
  specs/api-gateway/security.feature:42 - invalid control: @control:xyz-1 (unknown family)

Validation failed: 2 invalid control references found
```

---

## OSCAL Document Validation

### `validate risk-catalog`

Validates catalog documents against OSCAL 1.1.3 catalog schema.

```bash
eac validate risk-catalog
```

**Validates**: `templates/specs/risk-catalog/*.catalog.json`

### `validate risk-profile`

Validates profile documents against OSCAL 1.1.2 profile schema.

```bash
eac validate risk-profile
```

**Validates**: `specs/.risk-controls/*.profile.json`

---

## Risk Profile Creation

### `create risk-profile`

AI-powered command to generate an OSCAL profile from a risk assessment document.

```bash
r2r create risk-profile assessment.md
```

**Input**: Markdown risk assessment document describing identified risks

**Output**: `specs/.risk-controls/risk-profile.json`

**Example**:

```json
{
  "profile": {
    "uuid": "...",
    "metadata": { "title": "Repository-Wide Risk Controls" },
    "imports": [{
      "href": "../../../templates/specs/risk-catalog/controls.catalog.json",
      "include-controls": [{
        "with-ids": ["ac-2", "ac-3", "au-2", "ia-5", "ia-5(1)", "sc-7", "sc-8"]
      }]
    }]
  }
}
```

---

## Assessment Results Generation

### `create risk-assess`

Generates OSCAL assessment results linking controls to test evidence.

```bash
eac create risk-assess --profile specs/.risk-controls/risk-profile.json
```

**Output**: `out/risk/assessment-results.json`

**Contains**:

- **Observations**: Test results, security scans, timestamps
- **Findings**: Per-control satisfied/not-satisfied status
- **Evidence Links**: Traces controls → scenarios → test results

---

## File Structure

```text
Project Root/
├── specs/
│   ├── .risk-controls/              # OSCAL profiles
│   │   └── risk-profile.json        # Single profile for entire repository
│   └── <module>/
│       └── <feature>/
│           └── specification.feature  # Scenarios with @control: tags
├── templates/
│   └── specs/
│       └── risk-catalog/
│           └── controls.catalog.json  # Standard control definitions
└── out/
    └── risk/
        └── assessment-results.json    # Generated evidence
```

---

## Control Tag Format

**Pattern**: `@control:<family>-<number>` or `@control:<family>-<number>(<enhancement>)`

| Component     | Description                            | Example       |
| ------------- | -------------------------------------- | ------------- |
| `family`      | 2-4 lowercase letters                  | `ac`, `ia`    |
| `number`      | 1+ digits                              | `2`, `12`     |
| `enhancement` | Optional number in parentheses         | `(1)`, `(10)` |

**Examples**:

- `@control:ac-2` - Account Management
- `@control:ia-5(1)` - Password-Based Authentication
- `@controls:ac-2,au-3` - Multiple controls (comma-separated, no spaces)

---

## Related Documentation

- [Control Tags](./control-tags.md) - Complete control tag usage guide
- [Risk Controls (Conceptual)](../../../explanation/specifications/compliance/risk-controls.md) - What are risk controls and OSCAL
- [Validate Command Reference](../commands/validate/index.md) - Full validation command options
- [Create Command Reference](../commands/create/index.md) - Full create command options

# Risk Assessment to OSCAL Profile Generator

You are a security controls analyst. Generate a complete OSCAL 1.1.2 profile document that maps risks to security controls.

## Output Format

Generate ONLY valid OSCAL profile JSON. No markdown, no explanations, just pure JSON:

```json
{
  "profile": {
    "uuid": "GENERATE-NEW-UUID",
    "metadata": {
      "title": "Solution Risk Profile",
      "last-modified": "CURRENT-TIMESTAMP",
      "version": "1.0.0",
      "oscal-version": "1.1.2"
    },
    "imports": [
      {
        "href": "CATALOG-URL-HERE",
        "include-controls": [
          {
            "with-ids": ["ac-2", "ia-5", "si-10"]
          }
        ]
      }
    ],
    "back-matter": {
      "resources": [
        {
          "uuid": "GENERATE-UUID-1",
          "description": "ac-2: Account Management - Brief description of the control"
        },
        {
          "uuid": "GENERATE-UUID-2",
          "description": "ia-5: Authenticator Management - Brief description of the control"
        }
      ]
    }
  }
}
```

## Required OSCAL Profile Structure

### profile (required)

Root object containing the profile.

### uuid (required)

Generate a new UUID v4 for this profile.

### metadata (required)

- **title**: "Solution Risk Profile" (always use this title)
- **last-modified**: Current timestamp in ISO 8601 format (e.g., "2025-01-13T10:30:00Z")
- **version**: "1.0.0"
- **oscal-version**: "1.1.2"

### imports (required, array with 1 item)

- **href**: Use the catalog URL: `{{.Custom.CatalogURL}}`
- **include-controls**: Array with one object containing:
  - **with-ids**: Array of control IDs selected from the Available Controls list

### back-matter (optional but recommended)

- **resources**: Array of resource objects, one per control
- Each resource has:
  - **uuid**: Generate unique UUID v4
  - **description**: Format as "control-id: Control Title - Brief description"

## Critical Rules

**You MUST follow these rules:**

1. **ONLY use control IDs from the Available Controls list below**
2. Control IDs must be lowercase with hyphen format (e.g., "ac-2")
3. Generate valid UUIDs (use format: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx where x is hex digit, y is 8/9/a/b)
4. Use ISO 8601 timestamp format for last-modified
5. Set href to: `{{.Custom.CatalogURL}}`
6. Generate pure JSON - no markdown fences, no commentary
7. Every HIGH/CRITICAL risk must have at least one control
8. Only include controls that address risks in the assessment

## Available Controls

**Catalog Source**: `{{.Custom.CatalogURL}}`

**You MUST select controls ONLY from this list:**

```text
{{.Custom.AvailableControls}}
```

**Any control IDs not in the list above will be rejected.**

---

## Risk-to-Control Mapping Guide

| Risk Type | Example Controls | When to Use |
|-----------|-----------------|-------------|
| Credential exposure | ia-5, ia-2 | Authentication, password management |
| Unauthorized access | ac-2, ac-3, ac-6 | Account management, access control |
| Supply chain compromise | sr-3, sr-4 | Third-party dependencies |
| Dependency vulnerabilities | ra-5, si-2 | Vulnerability scanning, patching |
| Injection attacks | si-10 | Input validation |
| Data exposure | sc-8, sc-13, sc-28 | Encryption, secure transmission |
| Insufficient logging | au-2, au-3, au-12 | Audit events |
| Configuration issues | cm-2, cm-6 | Configuration management |

**Note**: These are examples. Always use controls from the Available Controls list above.

---

## Assessment Document

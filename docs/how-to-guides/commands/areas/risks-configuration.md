# Risk Management Configuration

This guide covers configuration options for EAC's risk management system, including OSCAL profiles, AI prompts, and assessment settings.

## Configuration Files

| File                                    | Purpose                              |
| --------------------------------------- | ------------------------------------ |
| `.r2r/eac/risk/profile.json`            | OSCAL profile with selected controls |
| `.r2r/eac/risk/assessment-results.json` | Assessment findings and evidence     |
| `.r2r/eac/risk/catalog.json`            | Custom control catalog (optional)    |
| `.r2r/eac/ai/risk/`                     | AI prompt templates                  |

## OSCAL Profile Configuration

### Profile Structure

```json
{
  "profile": {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "metadata": {
      "title": "Project Security Profile",
      "version": "1.0.0",
      "oscal-version": "1.1.2",
      "last-modified": "2024-12-01T00:00:00Z"
    },
    "imports": [
      {
        "href": "https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json",
        "include-controls": [
          {
            "with-ids": ["ac-1", "ac-2", "ac-3", "au-1", "au-2"]
          }
        ]
      }
    ],
    "modify": {
      "set-parameters": [
        {
          "param-id": "ac-1_prm_1",
          "values": ["organization-defined frequency"]
        }
      ]
    }
  }
}
```

### Control Selection

Select controls by ID:

```json
{
  "include-controls": [
    {
      "with-ids": [
        "ac-1",   // Access Control Policy
        "ac-2",   // Account Management
        "ac-3",   // Access Enforcement
        "au-1",   // Audit Policy
        "au-2"    // Audit Events
      ]
    }
  ]
}
```

Or by matching pattern:

```json
{
  "include-controls": [
    {
      "matching": [
        { "pattern": "ac-*" },    // All access control
        { "pattern": "au-*" }     // All audit controls
      ]
    }
  ]
}
```

### Built-in Catalogs

| Catalog           | URI                | Controls               |
| ----------------- | ------------------ | ---------------------- |
| NIST 800-53 Rev 5 | `nist-800-53-rev5` | ~1000 controls         |
| CIS Controls v8   | `cis-v8`           | 18 control families    |
| SOC 2             | `soc2`             | Trust service criteria |

## Assessment Results Configuration

### Results Structure

```json
{
  "assessment-results": {
    "uuid": "550e8400-e29b-41d4-a716-446655440001",
    "metadata": {
      "title": "Security Assessment Results",
      "version": "1.0.0",
      "oscal-version": "1.1.2",
      "last-modified": "2024-12-01T00:00:00Z"
    },
    "import-ap": {
      "href": "#profile"
    },
    "results": [
      {
        "uuid": "result-uuid",
        "title": "Automated Assessment",
        "start": "2024-12-01T00:00:00Z",
        "end": "2024-12-01T00:05:00Z",
        "findings": []
      }
    ]
  }
}
```

### Finding Structure

```json
{
  "uuid": "finding-uuid",
  "title": "Access Control Implementation",
  "description": "Evaluation of AC-2 Account Management",
  "target": {
    "type": "control",
    "target-id": "ac-2",
    "status": {
      "state": "satisfied"
    }
  },
  "implementation-statement-uuid": "impl-uuid",
  "related-observations": [
    { "observation-uuid": "obs-1" }
  ]
}
```

### Status States

| State           | Description                     |
| --------------- | ------------------------------- |
| `satisfied`     | Control fully implemented       |
| `not-satisfied` | Control not implemented         |
| `other`         | Partial or compensating control |

## AI Configuration

### Prompt Templates

Location: `.r2r/eac/ai/risk/`

```text
.r2r/eac/ai/risk/
├── create-profile.md       # Profile generation prompt
├── assess-control.md       # Control assessment prompt
├── generate-finding.md     # Finding generation prompt
└── summarize-results.md    # Report summary prompt
```

### Create Profile Prompt

```markdown
# Risk Profile Generation

## Context
You are generating an OSCAL profile for a software project.

## Input
- Project type: {{.Project.Type}}
- Compliance requirements: {{.Requirements}}
- Existing controls: {{.ExistingControls}}

## Output
Generate an OSCAL profile selecting appropriate controls from the specified catalog.

## Guidelines
1. Select controls relevant to the project type
2. Include all required compliance controls
3. Add recommended security controls
4. Document rationale for each selection
```

### Assessment Prompt

```markdown
# Control Assessment

## Context
Assess the implementation status of a security control.

## Control
- ID: {{.Control.ID}}
- Title: {{.Control.Title}}
- Description: {{.Control.Description}}

## Evidence
- Test Results: {{.Evidence.Tests}}
- Security Scans: {{.Evidence.Scans}}
- Code Analysis: {{.Evidence.Code}}

## Output
Provide assessment finding with:
1. Implementation status (satisfied/not-satisfied/other)
2. Evidence references
3. Gaps identified
4. Recommendations
```

## Risk Domains

### Domain Configuration

```yaml
# .r2r/eac/risk/domains.yml
domains:
  security:
    name: Information Security
    description: Controls for protecting information assets
    control_families:
      - access-control
      - audit
      - identification
      - system-protection

  operational:
    name: Operational Risk
    description: Controls for operational processes
    control_families:
      - contingency
      - maintenance
      - media-protection

  compliance:
    name: Regulatory Compliance
    description: Controls for regulatory requirements
    control_families:
      - privacy
      - audit
      - accountability
```

### Domain Selection

```bash
# Create profile for specific domain
r2r eac create-risk --domain security

# Create profile for multiple domains
r2r eac create-risk --domain security --domain compliance

# Create profile for all domains
r2r eac create-risk --domain all
```

## Evidence Mapping

### Test Evidence

Link tests to controls:

```gherkin
@control:AC-2
Feature: User Account Management

  Scenario: Create new user account
    Given an administrator is logged in
    When they create a new user account
    Then the account should be created with default permissions
```

### Security Scan Evidence

```yaml
# Evidence mapping in assessment
evidence:
  - control: AC-2
    sources:
      - type: test
        path: specs/auth/account-management.feature
        status: passing
      - type: sast
        tool: semgrep
        findings: 0
      - type: vuln
        tool: trivy
        findings: 0
```

## Environment Variables

| Variable             | Description                | Default                                 |
| -------------------- | -------------------------- | --------------------------------------- |
| `OSCAL_CATALOG_PATH` | Path to custom catalogs    | `.r2r/eac/risk/catalogs/`               |
| `RISK_PROFILE_PATH`  | Path to active profile     | `.r2r/eac/risk/profile.json`            |
| `RISK_RESULTS_PATH`  | Path to assessment results | `.r2r/eac/risk/assessment-results.json` |

## Validation Rules

### Profile Validation

```bash
r2r eac validate-risk
```

Validates:

- OSCAL schema compliance
- Control references exist in catalog
- Parameter values are valid
- UUID uniqueness

### Catalog Validation

```bash
r2r eac validate-risk-catalog
```

Validates:

- OSCAL catalog schema
- Control hierarchy
- Parameter definitions
- Required metadata

## Integration Settings

### CI/CD Integration

```yaml
# .github/workflows/risk.yml
risk:
  schedule: "0 0 * * 1"  # Weekly assessment
  on_merge: true          # Assess on merge to main
  fail_on:
    - not-satisfied       # Fail if any control not satisfied
```

### Report Generation

```yaml
# Report settings
reports:
  format: markdown        # markdown, json, html
  output: out/reports/    # Output directory
  include:
    - summary             # Executive summary
    - findings            # Detailed findings
    - evidence            # Evidence references
    - gaps                # Gap analysis
```

## Example Configurations

### Minimal Configuration

```json
{
  "profile": {
    "uuid": "minimal-profile",
    "metadata": {
      "title": "Minimal Security Profile",
      "version": "1.0.0",
      "oscal-version": "1.1.2"
    },
    "imports": [
      {
        "href": "nist-800-53-rev5",
        "include-controls": [
          { "with-ids": ["ac-1", "ac-2", "au-1", "au-2"] }
        ]
      }
    ]
  }
}
```

### SOC 2 Configuration

```json
{
  "profile": {
    "uuid": "soc2-profile",
    "metadata": {
      "title": "SOC 2 Compliance Profile",
      "version": "1.0.0",
      "oscal-version": "1.1.2"
    },
    "imports": [
      {
        "href": "soc2",
        "include-controls": [
          {
            "matching": [
              { "pattern": "cc*" }
            ]
          }
        ]
      }
    ]
  }
}
```

## Troubleshooting

| Issue                    | Cause                   | Solution                     |
| ------------------------ | ----------------------- | ---------------------------- |
| Profile validation fails | Invalid OSCAL structure | Check against schema         |
| Control not found        | ID mismatch             | Verify control ID in catalog |
| Assessment empty         | No evidence mapped      | Add test/scan evidence       |
| AI generation fails      | Missing context         | Check prompt templates       |

## Related Documentation

- [Risk Management Overview](risks-overview.md) - Concepts and workflows
- [Risk Commands](risks-commands.md) - Command reference
- [Security Configuration](security-configuration.md) - Security scan settings

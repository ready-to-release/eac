<!-- EDITOR
# Editor: how-to-guides/commands/areas/risks-commands.md

## Soul

Command reference for OSCAL-based risk management with 5 commands: AI profile generation, evidence-based assessment, report display, profile validation, and catalog validation.

## Sections

1. Quick Reference
2. create risk
3. create risk-assess
4. show risk-report
5. validate risk
6. validate risk-catalog
7. Common Workflows
8. Related Documentation
-->

# Risk Commands

Command reference for EAC's risk management system.

## Quick Reference

| Command                 | Description                                                     |
| ----------------------- | --------------------------------------------------------------- |
| `create risk`           | Create OSCAL profile from risk assessment using AI              |
| `create risk-assess`    | Update OSCAL assessment-results with test and security evidence |
| `show risk-report`      | Display aggregated risk assessment report                       |
| `validate risk`         | Validate OSCAL profiles and assessment-results against schemas  |
| `validate risk-catalog` | Validate OSCAL catalogs against OSCAL 1.1.3 schema              |

---

## create risk

Create an OSCAL profile from risk assessment requirements using AI.

### Synopsis

```bash
r2r eac create risk [options]
```

### Description

Generates an OSCAL profile by analyzing project requirements and selecting appropriate controls from a security catalog. Uses AI to determine relevant controls based on:

- Project type and technology stack
- Compliance requirements (SOC 2, ISO 27001, etc.)
- Risk domains (security, operational, compliance)
- Existing security controls

### Flags

| Flag          | Short | Type   | Default                      | Description                                 |
| ------------- | ----- | ------ | ---------------------------- | ------------------------------------------- |
| `--domain`    | `-d`  | string | `security`                   | Risk domain to assess                       |
| `--framework` | `-f`  | string | -                            | Compliance framework (soc2, iso27001, nist) |
| `--catalog`   | `-c`  | string | `nist-800-53-rev5`           | Control catalog to use                      |
| `--output`    | `-o`  | string | `.r2r/eac/risk/profile.json` | Output file path                            |
| `--debug`     |       | bool   | `false`                      | Enable debug output                         |

### Examples

```bash
# Create security risk profile
r2r eac create risk --domain security

# Create SOC 2 compliance profile
r2r eac create risk --framework soc2

# Create profile with multiple domains
r2r eac create risk --domain security --domain operational

# Use custom catalog
r2r eac create risk --catalog cis-v8

# Custom output location
r2r eac create risk --output custom/profile.json
```

### Output

```text
Creating risk profile...

Analyzing project requirements...
  ✓ Project type: go-monorepo
  ✓ Compliance needs: SOC 2
  ✓ Risk domain: security

Selecting controls from NIST 800-53...
  ✓ Access Control (AC): 12 controls
  ✓ Audit (AU): 8 controls
  ✓ Identification (IA): 6 controls

✓ Profile created: .r2r/eac/risk/profile.json
  Total controls: 26
```

### Exit Codes

| Code | Description                  |
| ---- | ---------------------------- |
| 0    | Profile created successfully |
| 1    | Error creating profile       |
| 2    | Invalid catalog or framework |

---

## create risk-assess

Update OSCAL assessment-results with evidence from tests and security scans.

### Synopsis

```bash
r2r eac create risk-assess [options]
```

### Description

Analyzes test results, security scan findings, and code coverage to create or update OSCAL assessment results. Links evidence to controls and determines implementation status.

Evidence sources:

- Test results (unit, integration, e2e)
- Security scans (SAST, vulnerability, secrets)
- Code coverage reports
- Compliance check results

### Flags

| Flag             | Short | Type   | Default                                 | Description                   |
| ---------------- | ----- | ------ | --------------------------------------- | ----------------------------- |
| `--profile`      | `-p`  | string | `.r2r/eac/risk/profile.json`            | Profile to assess against     |
| `--output`       | `-o`  | string | `.r2r/eac/risk/assessment-results.json` | Output file                   |
| `--evidence-dir` | `-e`  | string | `out/`                                  | Directory containing evidence |
| `--debug`        |       | bool   | `false`                                 | Enable debug output           |

### Examples

```bash
# Run assessment with defaults
r2r eac create risk-assess

# Use custom profile
r2r eac create risk-assess --profile custom/profile.json

# Specify evidence directory
r2r eac create risk-assess --evidence-dir build/reports/

# Debug mode
r2r eac create risk-assess --debug
```

### Output

```text
Running risk assessment...

Loading profile: .r2r/eac/risk/profile.json
  Controls to assess: 26

Gathering evidence...
  ✓ Test results: 47 passing, 2 failing
  ✓ Security scans: 0 critical, 2 medium
  ✓ Coverage: 78%

Assessing controls...
  AC-1 Access Control Policy: satisfied
  AC-2 Account Management: satisfied
  AC-3 Access Enforcement: not-satisfied (missing tests)
  ...

Assessment complete:
  ✓ Satisfied: 22 controls
  ✗ Not satisfied: 3 controls
  ○ Other: 1 control

✓ Results saved: .r2r/eac/risk/assessment-results.json
```

### Exit Codes

| Code | Description                  |
| ---- | ---------------------------- |
| 0    | Assessment completed         |
| 1    | Assessment failed            |
| 2    | Profile not found            |
| 3    | Evidence directory not found |

---

## show risk-report

Display aggregated risk assessment report.

### Synopsis

```bash
r2r eac show risk-report [options]
```

### Description

Generates a human-readable report from OSCAL assessment results showing:

- Overall compliance status
- Control implementation summary
- Gaps and recommendations
- Evidence traceability

### Flags

| Flag          | Short | Type   | Default                                 | Description                           |
| ------------- | ----- | ------ | --------------------------------------- | ------------------------------------- |
| `--format`    | `-f`  | string | `table`                                 | Output format (table, json, markdown) |
| `--results`   | `-r`  | string | `.r2r/eac/risk/assessment-results.json` | Assessment results file               |
| `--detailed`  | `-d`  | bool   | `false`                                 | Show detailed findings                |
| `--gaps-only` |       | bool   | `false`                                 | Show only unsatisfied controls        |

### Examples

```bash
# Display summary report
r2r eac show risk-report

# Detailed report
r2r eac show risk-report --detailed

# Show only gaps
r2r eac show risk-report --gaps-only

# JSON output for processing
r2r eac show risk-report --format json

# Markdown for documentation
r2r eac show risk-report --format markdown > risk-report.md
```

### Output (Table Format)

```text
Risk Assessment Report
═══════════════════════════════════════════════════════

Assessment Date: 2024-12-01
Profile: SOC 2 Compliance
Total Controls: 26

Summary
───────────────────────────────────────────────────────
✓ Satisfied:     22 (85%)
✗ Not Satisfied:  3 (12%)
○ Other:          1 (3%)

Control Status by Family
───────────────────────────────────────────────────────
│ Family              │ Total │ Satisfied │ Gaps │
├─────────────────────┼───────┼───────────┼──────┤
│ Access Control (AC) │    12 │        11 │    1 │
│ Audit (AU)          │     8 │         7 │    1 │
│ Identification (IA) │     6 │         4 │    2 │

Gaps Requiring Attention
───────────────────────────────────────────────────────
1. AC-3 Access Enforcement
   Status: not-satisfied
   Gap: Missing integration tests for access control
   Recommendation: Add tests for role-based access

2. AU-6 Audit Review
   Status: not-satisfied
   Gap: No audit log review process documented
   Recommendation: Implement audit log monitoring
```

### Exit Codes

| Code | Description             |
| ---- | ----------------------- |
| 0    | Report generated        |
| 1    | Error generating report |
| 2    | Results file not found  |

---

## validate risk

Validate OSCAL profiles and assessment-results against schemas.

### Synopsis

```bash
r2r eac validate risk [options]
```

### Description

Validates OSCAL documents against the official OSCAL 1.1.3 JSON schemas:

- Profile schema validation
- Assessment-results schema validation
- Control reference verification
- UUID uniqueness checks

### Flags

| Flag        | Short | Type   | Default                                 | Description              |
| ----------- | ----- | ------ | --------------------------------------- | ------------------------ |
| `--profile` | `-p`  | string | `.r2r/eac/risk/profile.json`            | Profile file to validate |
| `--results` | `-r`  | string | `.r2r/eac/risk/assessment-results.json` | Results file to validate |
| `--all`     | `-a`  | bool   | `false`                                 | Validate all OSCAL files |
| `--strict`  |       | bool   | `false`                                 | Enable strict validation |

### Examples

```bash
# Validate default files
r2r eac validate risk

# Validate specific profile
r2r eac validate risk --profile custom/profile.json

# Validate all OSCAL files
r2r eac validate risk --all

# Strict mode (fail on warnings)
r2r eac validate risk --strict
```

### Output

```text
Validating OSCAL documents...

Profile: .r2r/eac/risk/profile.json
  ✓ Schema validation passed
  ✓ Control references valid
  ✓ UUIDs unique

Assessment Results: .r2r/eac/risk/assessment-results.json
  ✓ Schema validation passed
  ✓ Finding references valid
  ✓ Evidence links valid

✓ All validations passed
```

### Validation Errors

```text
Validating OSCAL documents...

Profile: .r2r/eac/risk/profile.json
  ✗ Schema validation failed:
    - Line 15: Missing required field 'metadata.version'
    - Line 42: Invalid control reference 'ac-99'

Assessment Results: .r2r/eac/risk/assessment-results.json
  ✓ Schema validation passed
  ⚠ Warning: Finding 'f-123' references non-existent observation

✗ Validation failed with 2 errors, 1 warning
```

### Exit Codes

| Code | Description       |
| ---- | ----------------- |
| 0    | Validation passed |
| 1    | Validation failed |
| 2    | File not found    |

---

## validate risk-catalog

Validate OSCAL catalogs against OSCAL 1.1.3 schema.

### Synopsis

```bash
r2r eac validate risk-catalog [catalog-path]
```

### Description

Validates custom OSCAL catalog files against the official schema. Used when creating or modifying custom control catalogs.

### Arguments

| Argument       | Required | Description          |
| -------------- | -------- | -------------------- |
| `catalog-path` | Yes      | Path to catalog file |

### Flags

| Flag       | Short | Type | Default | Description              |
| ---------- | ----- | ---- | ------- | ------------------------ |
| `--strict` |       | bool | `false` | Enable strict validation |

### Examples

```bash
# Validate custom catalog
r2r eac validate risk-catalog .r2r/eac/risk/catalogs/custom.json

# Strict validation
r2r eac validate risk-catalog --strict custom-catalog.json
```

### Output

```text
Validating catalog: .r2r/eac/risk/catalogs/custom.json

  ✓ Schema validation passed
  ✓ Control hierarchy valid
  ✓ Parameter definitions valid
  ✓ 45 controls validated

✓ Catalog validation passed
```

### Exit Codes

| Code | Description            |
| ---- | ---------------------- |
| 0    | Validation passed      |
| 1    | Validation failed      |
| 2    | Catalog file not found |

---

## Common Workflows

### Initial Risk Setup

```bash
# 1. Create risk profile
r2r eac create risk --framework soc2

# 2. Run security scans
r2r eac security

# 3. Run tests
r2r eac test

# 4. Create assessment
r2r eac create risk-assess

# 5. View report
r2r eac show risk-report
```

### Continuous Assessment

```bash
# In CI pipeline
r2r eac security
r2r eac test
r2r eac create risk-assess
r2r eac validate risk
```

### Audit Preparation

```bash
# Generate detailed report
r2r eac show risk-report --detailed --format markdown > audit-report.md

# Export JSON for tools
r2r eac show risk-report --format json > audit-data.json
```

---

## Related Documentation

- [Risk Management Overview](risks-overview.md) - Concepts and workflows
- [Risk Configuration](risks-configuration.md) - Configuration reference
- [Security Commands](security-commands.md) - Security scanning

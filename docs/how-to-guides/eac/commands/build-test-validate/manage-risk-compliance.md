# Manage Risk Compliance

## What You'll Accomplish

Track security compliance using OSCAL (Open Security Controls Assessment Language) with automated evidence collection.

## Prerequisites

- OSCAL catalog defined (NIST, CIS, etc.)
- Risk assessment needs tracking
- Security scans configured

## Steps

### 1. Create Risk Profile

```bash
r2r eac create risk-profile --catalog nist-800-53
```

**What happens**: Generates OSCAL profile from risk assessment

### 2. Run Security Scans

```bash
r2r eac scan --scanner compliance
```

**What happens**: Collects evidence for compliance controls

### 3. Update Assessment Results

```bash
r2r eac create risk-assess
```

**What happens**: Updates OSCAL assessment-results with test and security evidence

### 4. Validate Compliance

```bash
r2r eac validate risk-profile
r2r eac validate risk-catalog
```

**What happens**: Checks OSCAL documents are valid

## OSCAL Components

### Profile

Defines which controls to implement:

```json
{
  "profile": {
    "uuid": "...",
    "metadata": {...},
    "imports": [
      {
        "href": "nist-800-53-catalog.json",
        "include-controls": [
          {"control-id": "ac-1"},
          {"control-id": "au-1"}
        ]
      }
    ]
  }
}
```

### Assessment Results

Evidence of compliance:

```json
{
  "results": [{
    "control-id": "ac-1",
    "status": "satisfied",
    "evidence": [
      {
        "description": "Security scan passed",
        "link": "scan-results.json"
      }
    ]
  }]
}
```

## Example Scenario

Tracking NIST 800-53 compliance:

```bash
# Create profile for required controls
r2r eac create risk-profile --catalog nist-800-53

# Output:
# ✓ Created oscal/profile.json
# Included 25 controls

# Run compliance scans
r2r eac scan --scanner compliance

# Output:
# Checking AC-1 (Access Control Policy)... ✓
# Checking AU-1 (Audit Policy)... ✓
# Checking SC-7 (Boundary Protection)... ✗
#   Missing firewall rules

# Update assessment with evidence
r2r eac create risk-assess

# Output:
# ✓ Updated oscal/assessment-results.json
# 24/25 controls satisfied

# Validate OSCAL documents
r2r eac validate risk-profile
# ✓ Profile valid per OSCAL 1.1.3 schema

r2r eac validate risk-catalog
# ✓ Catalog valid
```

## Control Tagging

Tag tests with control IDs:

```go
// @control: AC-1
func TestAccessControl(t *testing.T) {
    // Test access control policy
}
```

When tests run, evidence is collected for AC-1.

## Common Issues

| Problem             | Solution                         |
| ------------------- | -------------------------------- |
| "Control not found" | Check control ID against catalog |
| Invalid OSCAL       | Validate against schema          |
| Missing evidence    | Run security scans and tests     |

## Customizing Reports

The `create risk-assess` command generates reports using templates that can be customized:

**Custom template**: `.r2r/eac/templates/reports/risk/risk-assess.md`

## Next Steps

- [Scan for Security Issues](./scan-for-security-issues.md) → Collect evidence

## Related Commands

- [`create risk-profile`](../../../../reference/commands/create/risk-profile.md) - Create profile
- [`create risk-assess`](../../../../reference/commands/create/risk-assess.md) - Update assessment
- [`scan`](../../../../reference/commands/scan/scan.md) - Check compliance with --scanner compliance
- [`validate risk-profile`](../../../../reference/commands/validate/risk-profile.md) - Validate profile

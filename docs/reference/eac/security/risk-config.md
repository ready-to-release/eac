# Risk Configuration Reference

Risk scoring and OSCAL profile configuration for security assessments.

---

## Configuration Files

**Contract default**: `contracts/scanner/0.1.0/schemas/defaults/risk-config.yml`

**User override**: `.eac/risk-config.yml` (highest priority)

**Schema**: `contracts/scanner/0.1.0/schemas/risk-config.schema.json`

---

## Configuration Structure

```yaml
# OSCAL profile reference
profile:
  path: risk-profile.json  # Path to OSCAL profile (relative)
  catalog_url: https://...  # NIST SP 800-53 catalog URL

# Risk scoring
scoring:
  # Impact ratings (1-5 scale)
  impact:
    api: 4
    service: 4
    gateway: 4
    library: 3
    core: 3
    cli: 2
    tool: 2
    docs: 1
    _default: 3

  # Criticality levels (high/medium/low)
  criticality:
    api: high
    gateway: high
    service: high
    core: medium
    library: medium
    cli: low
    tool: low
    _default: medium

  # Severity weights for likelihood
  severity_weights:
    critical: 4
    high: 3
    medium: 2
    low: 1

# Module-specific profiles (optional)
module_profiles:
  billing-service:
    path: billing-service.profile.json
```

---

## Impact Ratings

**Scale**: 1 (minimal) to 5 (catastrophic)

| Rating | Meaning                                      | Examples      |
| ------ | -------------------------------------------- | ------------- |
| **5**  | Catastrophic - System-wide failure           | N/A (reserved) |
| **4**  | High - Major service disruption              | api, gateway  |
| **3**  | Moderate - Feature impairment                | library, core |
| **2**  | Low - Minor degradation                      | cli, tool     |
| **1**  | Minimal - Documentation/config only          | docs, config  |

---

## Criticality Levels

**Levels**: high, medium, low

| Level      | Meaning                   | Examples      |
| ---------- | ------------------------- | ------------- |
| **high**   | Mission-critical services | api, gateway  |
| **medium** | Important infrastructure  | core, library |
| **low**    | Developer tools           | cli, tool     |

---

## OSCAL Profile

OSCAL (Open Security Controls Assessment Language) profile defines selected security controls from NIST SP 800-53.

**Location**: Path specified in `profile.path` (e.g., `risk-profile.json`)

**Generate from markdown**:

```bash
eac create risk-profile assessment.md
```

**Validate profile**:

```bash
eac validate risk-profile profile.json
eac validate risk-catalog  # Validate against NIST catalog
```

**Profile structure** (simplified):

```json
{
  "profile": {
    "uuid": "...",
    "metadata": {
      "title": "System Risk Profile",
      "version": "1.0.0"
    },
    "imports": [
      {
        "href": "<nist-catalog-url>",
        "include-controls": [
          { "control-id": "ac-2" },
          { "control-id": "ia-5" }
        ]
      }
    ]
  }
}
```

---

## Risk Score Calculation

**Formula**: `Risk = Impact × Likelihood`

**Likelihood** is calculated from scan findings:

```text
Likelihood = Base + (Critical × 4) + (High × 3) + (Medium × 2) + (Low × 1)
```

Where:
- **Base**: Starting likelihood value
- **Critical/High/Medium/Low**: Count of findings at each severity
- **Weights**: From `severity_weights` configuration

---

## Commands

| Command                                     | Purpose                   |
| ------------------------------------------- | ------------------------- |
| `eac create risk-profile <markdown>`        | Generate OSCAL profile    |
| `eac create risk-assess --profile <json>`   | Create assessment results |
| `eac validate risk-profile <json>`          | Validate OSCAL profile    |
| `eac validate risk-catalog`                 | Validate catalog ref      |

---

## Related Documentation

- **[Security Index](./index.md)** - Security scanning overview
- **[Create Commands](../commands/create/index.md)** - Risk creation commands
- **[Validate Commands](../commands/validate/index.md)** - Validation commands

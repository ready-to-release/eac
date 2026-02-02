# Risk Configuration Reference

Technical reference for EAC risk configuration, including OSCAL profile references and scoring settings.

## Configuration File

Risk configuration is defined in `contracts/eac-security/0.1.0/defaults/risk-config.yml` with user overrides in `.r2r/eac/risk-config.yml`.

### Location Priority

1. `.r2r/eac/risk-config.yml` - User/team overrides (highest priority)
2. `contracts/eac-security/0.1.0/defaults/risk-config.yml` - Contract defaults

## Configuration Structure

```yaml
# Risk and compliance configuration

# OSCAL profile reference
profile:
  path: risk-profile.json          # Path to OSCAL profile (relative to this file)
  catalog_url: https://...         # NIST catalog URL for control validation

# Risk scoring configuration
scoring:
  impact:
    api: 4                         # Impact rating 1-5
    service: 4
    gateway: 4
    library: 3
    core: 3
    cli: 2
    tool: 2
    docs: 1
    config: 1
    _default: 3                    # Fallback for unknown types

  criticality:
    api: high                      # Criticality: high/medium/low
    gateway: high
    service: high
    core: medium
    library: medium
    cli: low
    tool: low
    _default: medium

  severity_weights:
    critical: 4                    # Likelihood increment per severity
    high: 3
    medium: 2
    low: 1

# Module-specific profiles (optional)
module_profiles:
  billing-service:
    path: billing-service.profile.json
```

## Profile Configuration

### Profile Path

The `profile.path` field references an OSCAL profile JSON file:

```yaml
profile:
  path: risk-profile.json  # Relative to config file location
```

For absolute paths, specify the full path:

```yaml
profile:
  path: /etc/oscal/enterprise-profile.json
```

### Catalog URL

The catalog URL is used for control validation:

```yaml
profile:
  catalog_url: https://raw.githubusercontent.com/usnistgov/oscal-content/main/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json
```

## Scoring Configuration

### Impact Ratings

Impact ratings determine potential damage severity (1-5 scale):

| Rating | Meaning | Examples |
|--------|---------|----------|
| 5 | Critical | Payment processing, authentication |
| 4 | High | APIs, gateways, core services |
| 3 | Medium | Libraries, shared components |
| 2 | Low | CLI tools, utilities |
| 1 | Minimal | Documentation, configuration |

### Criticality Levels

Criticality determines operational importance:

| Level | Meaning | Protection Level |
|-------|---------|------------------|
| high | Business critical | Maximum security controls |
| medium | Important | Standard security controls |
| low | Supporting | Basic security controls |

### Severity Weights

Severity weights increment likelihood scores per finding:

```yaml
severity_weights:
  critical: 4  # Each critical finding adds 4 to likelihood
  high: 3      # Each high finding adds 3
  medium: 2    # Each medium finding adds 2
  low: 1       # Each low finding adds 1
```

## Module-Specific Profiles

For modules requiring different control sets:

```yaml
module_profiles:
  billing-service:
    path: billing-service.profile.json

  internal-tools:
    path: minimal-profile.json
```

Module profiles override the solution-wide profile for specific modules.

## User Overrides

Create `.r2r/eac/risk-config.yml` to customize for your organization:

```yaml
# Override default scoring for your organization
scoring:
  impact:
    gateway: 5        # Our gateways are business-critical
    docs: 2           # Our docs contain sensitive info

  criticality:
    gateway: high
    docs: medium

# Custom profile location
profile:
  path: specs/.risk-controls/risk-profile.json
```

## Programmatic Access

### Go API

```go
import "github.com/ready-to-release/eac/go/core/config"

// Load risk configuration
cfg, err := config.LoadRiskConfig(repoRoot, configRoot)
if err != nil {
    return err
}

// Access scoring
scoring := cfg.GetScoring()
impact := scoring.GetImpact("api")           // Returns 4
criticality := scoring.GetCriticality("api") // Returns "high"

// Access profile
profile, err := cfg.GetProfile()
if err == nil {
    controls := profile.ControlIDs()
    hasControl := profile.HasControl("ac-1")
}

// Access catalog URL
catalogURL := cfg.GetCatalogURL()
```

### Interface Definitions

The configuration implements these interfaces from `contracts/eac-security/0.1.0/interfaces`:

- `RiskConfigPort` - Main configuration access
- `ProfilePort` - OSCAL profile access
- `RiskScoringPort` - Scoring configuration access

## Related Commands

| Command | Description |
|---------|-------------|
| `create risk-profile` | Generate OSCAL profile from risk assessment |
| `create risk-assess` | Create assessment results from evidence |
| `validate risk-profile` | Validate OSCAL profile document |

## Related Documentation

- [Scan Command Reference](../commands/scan/index.md) - Security scanning
- [OSCAL Compliance](../compliance/index.md) - OSCAL framework overview
- [Shift-Left Security](../../../explanation/continuous-delivery/security/shift-left.md) - Security integration

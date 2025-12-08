<!-- EDITOR
# Editor: how-to-guides/commands/areas/risks-overview.md

## Soul

OSCAL-based risk management provides AI-powered compliance tracking, control assessment, and traceability from code to compliance controls with evidence from tests and security scans.

## Sections

1. What is Risk Management?
2. When to Use Risk Management
3. Key Concepts
4. Workflow Overview
5. Integration Points
6. Next Steps
7. Related Areas
-->

# Risk Management

Risk management in EAC provides OSCAL-based compliance tracking, AI-powered risk assessment, and traceability from code to compliance controls.

## What is Risk Management?

EAC's risk management system enables you to:

- **Define risk profiles** using OSCAL (Open Security Controls Assessment Language)
- **Create control assessments** with AI-generated analysis
- **Link controls to evidence** (tests, security scans, code)
- **Generate compliance reports** for audits

The system uses AI to analyze your codebase, specifications, and security scan results to create comprehensive risk assessments that map to industry-standard compliance frameworks.

## When to Use Risk Management

Use risk management commands when you need:

| Scenario                              | Commands                                 |
| ------------------------------------- | ---------------------------------------- |
| Starting compliance for a new project | `create risk`                            |
| Updating assessments after changes    | `create risk-assess`                     |
| Preparing for audits                  | `show risk-report`                       |
| Validating OSCAL files                | `validate risk`, `validate risk-catalog` |

### Common Use Cases

- **SOC 2 compliance** - Map controls to code and test evidence
- **ISO 27001 certification** - Track security control implementation
- **Internal audits** - Generate traceability reports
- **Security reviews** - Document control status and gaps

## Key Concepts

### OSCAL (Open Security Controls Assessment Language)

OSCAL is a NIST standard for representing security control information in machine-readable formats. EAC uses:

- **Catalogs** - Define available controls (e.g., NIST 800-53)
- **Profiles** - Select controls applicable to your system
- **Assessment Results** - Document control implementation status

### Risk Domains

Controls are organized into domains:

| Domain        | Description                       |
| ------------- | --------------------------------- |
| `security`    | Information security controls     |
| `operational` | Operational and process controls  |
| `compliance`  | Regulatory and legal compliance   |
| `technical`   | Technical implementation controls |

### Traceability

The system creates links between:

```text
Control → Implementation → Evidence
   │            │              │
   │            │              └─ Test results, security scans
   │            └─ Code, configuration, documentation
   └─ OSCAL profile control selection
```

### AI-Powered Assessment

When you run `create risk-assess`, AI analyzes:

1. **Test results** - Which controls have passing tests
2. **Security scans** - SAST, vulnerability, and compliance findings
3. **Code coverage** - Implementation completeness
4. **Documentation** - Specification coverage

## Workflow Overview

### Initial Setup

```bash
# 1. Create risk profile from compliance requirements
r2r eac create risk --domain security --framework soc2

# 2. Review generated OSCAL profile
cat .r2r/eac/risk/profile.json

# 3. Customize control selections as needed
```

### Ongoing Assessment

```bash
# 1. Run security scans to gather evidence
r2r eac security

# 2. Run tests to generate test evidence
r2r eac test

# 3. Update assessment with latest evidence
r2r eac create risk-assess

# 4. Review assessment results
r2r eac show risk-report
```

### Audit Preparation

```bash
# 1. Validate all OSCAL files
r2r eac validate risk

# 2. Generate comprehensive report
r2r eac show risk-report --format detailed

# 3. Export for auditors
r2r eac show risk-report --format json > audit-report.json
```

## Integration Points

### With Security Scanning

Risk assessment automatically incorporates:

- SAST findings (`scan sast`)
- Vulnerability scan results (`scan vuln`)
- Compliance check results (`scan compliance`)
- SBOM data (`scan sbom`)

### With Testing

Test results provide evidence for control implementation:

- Unit test coverage maps to technical controls
- Integration tests validate operational controls
- Specification tests document behavior

### With CI/CD

Integrate risk assessment into pipelines:

```yaml
- name: Update risk assessment
  run: r2r eac create risk-assess

- name: Validate compliance
  run: r2r eac validate risk
```

## Next Steps

- [Risk Configuration](risks-configuration.md) - Configure AI prompts, domains, and OSCAL settings
- [Risk Commands](risks-commands.md) - Full command reference

## Related Areas

- [Security](security-overview.md) - Security scanning that feeds into risk assessment
- [Specifications](specifications-overview.md) - BDD specs that provide control evidence
- [Pipeline](pipeline-overview.md) - CI/CD integration for automated assessment

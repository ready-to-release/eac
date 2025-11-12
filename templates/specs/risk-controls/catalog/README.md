# Risk Controls Catalog

This catalog contains 299 predefined risk control specifications organized across 30 security domains.

## DISCLAIMER

**IMPORTANT**: This catalog provides **example risk controls** with **indicative regulatory mappings** only. These mappings are provided to illustrate potential applicability and should **NOT be used as-is for compliance purposes**.

**Proper usage requires:**

- Conducting a thorough risk assessment for your specific context
- Analyzing applicable threats and vulnerabilities in your environment
- Engaging qualified security and compliance personnel
- Selecting and tailoring controls based on your organizational requirements
- Validating regulatory applicability with legal and compliance teams

The regulatory tags (`@iso27001`, `@hipaa`, etc.) indicate potential relevance but do not constitute compliance guidance.

## Regulatory Mappings

Controls are tagged with relevant regulatory standards to help identify potential applicability:

**Information Security & Privacy:**

- `@iso27001` - ISO/IEC 27001:2022 (Information Security Management)
- `@gdpr` - EU General Data Protection Regulation
- `@ccpa` - California Consumer Privacy Act
- `@lgpd` - Brazil LGPD (Data Protection)
- `@nist-800-53` - NIST 800-53 (Federal Security Controls)

**Healthcare & Pharmaceutical:**

- `@hipaa` - HIPAA Security Rule
- `@fda-21cfr11` - FDA 21 CFR Part 11 (Electronic Records/Signatures)
- `@gamp5` - GAMP 5 (Pharmaceutical IT Validation)
- `@alcoa-plus` - ALCOA+ Principles (Data Integrity)
- `@gxp` - Good Practice Regulations

**Medical Devices:**

- `@iso13485` - ISO 13485 (Medical Device Quality Management)
- `@iec62304` - IEC 62304 (Medical Device Software Lifecycle)
- `@mdr` - EU Medical Device Regulation
- `@ivdr` - EU In Vitro Diagnostic Regulation
- `@fda-premarket` - FDA Premarket Cybersecurity Guidance
- `@imdrf` - IMDRF Cybersecurity Guidance

**Financial Services:**

- `@pci-dss` - PCI-DSS v4.0 (Payment Card Industry)
- `@sox` - Sarbanes-Oxley Act (Financial Controls)

**AI/ML Governance:**

- `@eu-ai-act` - EU AI Act (2024)
- `@nist-ai-rmf` - NIST AI Risk Management Framework
- `@iso42001` - ISO/IEC 42001 (AI Management System)

**Cloud & Infrastructure:**

- `@csa-ccm` - CSA Cloud Controls Matrix
- `@iso27017` - ISO/IEC 27017 (Cloud Security)
- `@iso27018` - ISO/IEC 27018 (Cloud Privacy)

**Application Security:**

- `@owasp` - OWASP Guidelines
- `@owasp-masvs` - OWASP Mobile App Security Verification Standard
- `@nist-ssdf` - NIST Secure Software Development Framework

**Cryptography:**

- `@fips-140-2` - FIPS 140-2 (Cryptographic Modules)
- `@fips-140-3` - FIPS 140-3 (Updated Standard)

**Risk Management:**

- `@iso14971` - ISO 14971 (Medical Device Risk Management)
- `@iso31000` - ISO 31000 (Risk Management Principles)

**Digital Identity (EU):**

- `@esign-act` - E-SIGN Act (Electronic Signatures)
- `@eidas` - eIDAS (EU Electronic Identification)

**Supply Chain:**

- `@slsa` - SLSA Framework (Supply Chain Levels for Software Artifacts)

See `catalog.yml` for the complete list of mappings and version information.

---

## Organization

Controls are organized by domain:

| Domain | Example Controls | Count | Folder |
|--------|-----------------|-------|--------|
| Authentication | `@risk-control:auth-mfa` | 9 | `authentication/` |
| Data Protection | `@risk-control:encrypt-rest` | 10 | `data-protection/` |
| Audit & Logging | `@risk-control:audit-trail` | 10 | `audit-logging/` |
| Access Control | `@risk-control:access-rbac` | 10 | `access-control/` |
| Privacy | `@risk-control:privacy-consent` | 10 | `privacy/` |
| AI/ML Governance | `@risk-control:ai-bias` | 10 | `ai-ml/` |
| ... | ... | ... | ... |

See `catalog.yml` for complete list.

## Control Structure

Each control follows this pattern:

```gherkin
# @industry:PHARMA @industry:MEDDEV @industry:HEALTH @industry:FINANCE @industry:GENERAL
# @severity:critical @risk-type:security @control-type:preventive
# @iso27001 @hipaa @pci-dss @fda-21cfr11 @nist-800-53
# @implementation:required @automation:full

@risk-control:control-name
Feature: Control Name

  As a [role]
  I want [objective]
  So that [outcome]

  Rule: Basic Requirement

    @risk-control:control-name-01
    Scenario: Simple test
      Given [precondition]
      When [action]
      Then [expected result]
```

**Metadata Elements:**

- **Industries**: `@industry:PHARMA`, `@industry:MEDDEV`, `@industry:HEALTH`, `@industry:FINANCE`, `@industry:GENERAL`
- **Severity**: `@severity:critical`, `@severity:high`, `@severity:medium`, `@severity:low`
- **Risk Type**: `@risk-type:security`, `@risk-type:privacy`, `@risk-type:compliance`
- **Control Type**: `@control-type:preventive`, `@control-type:detective`, `@control-type:corrective`
- **Implementation**: `@implementation:required`, `@implementation:recommended`
- **Automation**: `@automation:full`, `@automation:partial`, `@automation:manual`

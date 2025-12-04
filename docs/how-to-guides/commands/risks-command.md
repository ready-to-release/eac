# Risks Command

**Problem**: Regulated industries require systematic risk assessment, risk control specifications, and traceability between risks and implementations.

**Solution**: Use `risks` to generate AI-powered risk assessments, create executable risk control specifications, and track linkages between controls and implementations.

## Key Benefits

- AI-powered risk assessment from code changes
- Automated risk control specification generation
- Traceability between risks, controls, and implementations
- Compliance-ready documentation for audits
- Path traversal protection for security
- Integration with existing specification workflow

## Quick Start

```bash
# Generate risk assessment from staged changes
r2r eac risks assessment --scope staged

# Create risk control specifications from assessment
r2r eac risks create reports/risk-assessment-2025-01-15.md

# List all risk controls and their linkages
r2r eac risks list --filter all
```

## Typical Workflow

### Complete Risk Management Cycle

```bash
# 1. Make code changes
# ... edit files ...

# 2. Stage changes
git add src/auth/

# 3. Generate risk assessment
r2r eac risks assessment --scope staged --destination reports/

# 4. Review assessment
cat reports/risk-assessment-2025-01-15-143022.md

# 5. Create risk control specifications
r2r eac risks create reports/risk-assessment-2025-01-15-143022.md

# 6. Review generated controls
ls specs/risk-controls/

# 7. List all controls and their linkages
r2r eac risks list --filter all

# 8. Check for orphaned controls
r2r eac risks list --filter orphaned
```

## Command Reference

### risks assessment

Generate AI-powered risk assessment from code changes.

```bash
r2r eac risks assessment [options]

# Options:
--scope, -s <type>        # Scope: staged, changed, or all (default: staged)
--destination, -d <path>  # Output directory (default: reports/)
--prompt, -p <file>       # Custom AI prompt file
--debug, -D               # Enable debug mode

# Examples:
r2r eac risks assessment --scope staged
r2r eac risks assessment --scope changed --destination assessments/
r2r eac risks assessment --scope all --debug
r2r eac risks assessment --prompt custom-risk-prompt.md
```

**What it does:**

1. Analyzes code changes (staged, changed, or all files)
2. Reviews existing specifications for context
3. Uses AI to identify security, compliance, and operational risks
4. Generates comprehensive risk assessment report
5. Saves to `reports/risk-assessment-<date>-<time>.md`
6. Suggests risk controls with appropriate domains

**Scope options:**

- **staged**: Only git-staged changes (default, safe for commit-time assessment)
- **changed**: All modified but uncommitted changes
- **all**: All repository files (comprehensive assessment)

**Assessment includes:**

- Executive summary with overall risk level
- Risks organized by severity (Critical, High, Medium, Low, Info)
- Affected files and related specifications
- Impact and likelihood analysis
- Specific recommendations
- Suggested risk controls with domains

### risks create

Create risk control specifications from assessment report.

```bash
r2r eac risks create <file-or-folder> [options]

# Options:
--force, -f            # Overwrite existing control files
--allow-orphans        # Create controls without spec linkages
--output, -o <dir>     # Custom output directory
--prompt, -p <file>    # Custom AI prompt file
--debug, -D            # Enable debug mode

# Examples:
r2r eac risks create reports/risk-assessment-2025-01-15.md
r2r eac risks create reports/ --force
r2r eac risks create assessment.md --output custom-controls/
r2r eac risks create assessment.md --allow-orphans --debug
```

**What it does:**

1. Parses risk assessment report
2. Extracts risks requiring controls
3. Uses AI to generate Gherkin risk control specifications
4. Organizes by domain (authentication, api-security, etc.)
5. Saves to `specs/risk-controls/<domain>/<control-name>.feature`
6. Links controls to affected files and specifications

**Positional argument:**

- `<file-or-folder>`: Path to assessment file (.md) or directory containing assessments
  - Single file: Process one assessment
  - Directory: Process all .md files in directory

**Flags:**

- `--force, -f`: Overwrite existing control files (default: skip existing)
- `--allow-orphans`: Create controls even if no related specs exist (default: warn)
- `--output, -o`: Custom output directory (default: `specs/risk-controls/`)
- `--prompt, -p`: Custom AI prompt for control generation
- `--debug, -D`: Save intermediate outputs to `out/` directory

**Path traversal protection:**

The command validates all paths to prevent directory traversal attacks:

- Rejects paths containing `..`
- Ensures paths are within repository root
- Validates output directory is safe

### risks list

List risk controls and their linkages to specifications.

```bash
r2r eac risks list [options]

# Options:
--filter, -f <type>    # Filter: all, orphaned, linked, missing-links (default: all)
--json, -j             # Output in JSON format
--debug, -D            # Enable debug mode

# Examples:
r2r eac risks list
r2r eac risks list --filter orphaned
r2r eac risks list --filter linked --json
r2r eac risks list --filter missing-links
```

**What it does:**

1. Scans `specs/risk-controls/` for control features
2. Searches `specs/` for implementation linkages
3. Identifies controls with/without implementations
4. Displays traceability information

**Filter options:**

- **all**: Show all controls (default)
- **orphaned**: Controls with no implementation linkages
- **linked**: Controls with at least one implementation
- **missing-links**: Controls that should have more linkages

**Output modes:**

- **Text** (default): Human-readable table format
- **JSON** (`--json`): Machine-readable format for automation

**Example text output:**

```text
Risk Controls Summary
====================

Control: auth-mfa
Domain: authentication
File: specs/risk-controls/authentication/multi-factor-authentication-control.feature
Scenarios: 2
  - @risk-control:auth-mfa-01
  - @risk-control:auth-mfa-02
Implementations: 3
  - specs/cli/login/specification.feature:45 (@ov @risk-control:auth-mfa-01)
  - specs/api/auth/specification.feature:12 (@risk-control:auth-mfa-01)
  - specs/api/auth/specification.feature:34 (@risk-control:auth-mfa-02)
Status: Linked

Control: api-rate-limiting-control
Domain: api-security
File: specs/risk-controls/api-security/api-rate-limiting-control.feature
Scenarios: 3
  - @risk-control:api-rate-limiting-control-01
  - @risk-control:api-rate-limiting-control-02
  - @risk-control:api-rate-limiting-control-03
Implementations: 0
Status: Orphaned
```

**Example JSON output:**

```json
{
  "controls": [
    {
      "name": "auth-mfa",
      "domain": "authentication",
      "file": "specs/risk-controls/authentication/multi-factor-authentication-control.feature",
      "scenarios": [
        "@risk-control:auth-mfa-01",
        "@risk-control:auth-mfa-02"
      ],
      "implementations": [
        {
          "file": "specs/cli/login/specification.feature",
          "line": 45,
          "tag": "@risk-control:auth-mfa-01"
        }
      ],
      "status": "linked"
    }
  ],
  "summary": {
    "total": 15,
    "linked": 12,
    "orphaned": 3
  }
}
```

## Risk Control Domains

Risk controls are organized by domain for better management:

| Domain              | Purpose                                             | Example Controls                                                  |
| ------------------- | --------------------------------------------------- | ----------------------------------------------------------------- |
| **authentication**  | Login, MFA, session management, identity            | `auth-mfa`, `session-timeout`, `credential-storage`               |
| **authorization**   | Access control, permissions, RBAC                   | `rbac`, `least-privilege`, `access-review`                        |
| **api-security**    | API authentication, rate limiting, input validation | `api-rate-limiting`, `api-authentication`, `api-input-validation` |
| **data-protection** | Encryption, data handling, privacy                  | `encrypt-at-rest`, `encrypt-in-transit`, `data-masking`           |
| **compliance**      | Regulatory requirements, audit trails               | `audit-trail`, `regulatory-reporting`, `compliance-monitoring`    |
| **infrastructure**  | Network security, deployment, configuration         | `network-segmentation`, `secure-config`, `patch-management`       |
| **application**     | Code security, dependencies, secure coding          | `secure-coding`, `dependency-scan`, `vulnerability-management`    |
| **monitoring**      | Logging, alerting, incident response                | `security-events`, `incident-response`, `log-integrity`           |

## Generated Risk Control Format

### Control Specification Structure

```gherkin
# ========================================
# Risk Control: Multi-Factor Authentication Control
# ========================================
#
# Source: Risk Assessment RISK-001, Generated: 2025-01-15
# Severity: high | Domain: authentication
#
# This control addresses:
# - Multi-factor authentication not enforced for privileged accounts
# - Impact: Unauthorized access to privileged operations
# - Affected files: src/auth/handler.go, src/auth/middleware.go
#
# Implementation specifications must tag scenarios with:
# @risk-control:multi-factor-authentication-control-[id]
#

@risk-control:multi-factor-authentication-control
Feature: Multi-Factor Authentication Control

  Risk ID: RISK-001
  Severity: high
  Domain: authentication

  As a security officer
  I want multi-factor authentication required for all privileged accounts
  So that unauthorized access to privileged operations is prevented

  Background:
    Given the system is deployed in production
    And security controls are active

  Rule: Multi-factor authentication required for privileged access

    @risk-control:multi-factor-authentication-control-01
    Scenario: MFA enforced for privileged accounts
      Given a user attempts to access privileged operations
      When authentication is required
      Then the system MUST require at least two authentication factors
      And the system MUST validate both factors before granting access

    @risk-control:multi-factor-authentication-control-02
    Scenario: Single factor authentication rejected for privileged access
      Given a user provides only password credentials
      When attempting to access privileged operations
      Then the system MUST reject the authentication attempt
      And the system MUST log the failed authentication
      And no privileged access is granted
```

### Linking Implementations to Controls

Tag your implementation scenarios with the control ID:

```gherkin
# In specs/cli/login/specification.feature

@ov @risk-control:multi-factor-authentication-control-01
Scenario: Login with MFA for admin user
  Given I am an admin user
  When I run "r2r login --mfa"
  And I provide valid credentials
  And I provide valid MFA token
  Then I should be authenticated with privileged access
  # @ov = operational verification (functional test)
```

## Integration with Compliance Processes

### Assessment-First Process (REQUIRED)

**You MUST conduct risk assessment BEFORE creating controls.**

1. **Conduct Risk Assessment** - Identify YOUR specific threats and risks
2. **Generate Assessment** - Use `risks assessment` to document findings
3. **Review Assessment** - Validate with qualified personnel
4. **Create Controls** - Use `risks create` to generate specifications
5. **Link Implementations** - Tag scenarios with `@risk-control` tags
6. **Verify Linkages** - Use `risks list` to check traceability

### Regulated Industry Examples

| Framework                 | Use Case                            | Workflow                                                   |
| ------------------------- | ----------------------------------- | ---------------------------------------------------------- |
| **FDA 21 CFR Part 11**    | Electronic signatures, audit trails | `risks assessment --scope all` before validation           |
| **ISO 13485 / IEC 62304** | Medical device software             | `risks assessment` for each release, traceability via tags |
| **PCI-DSS**               | Payment card data protection        | `risks assessment --scope changed` pre-commit              |
| **GDPR**                  | Data privacy compliance             | `risks assessment` when handling personal data             |
| **EU AI Act**             | AI system risk management           | `risks assessment` for model changes                       |

### Compliance Workflow Example

```bash
# Development phase
git add src/payment/

# Pre-commit risk assessment (PCI-DSS requirement)
r2r eac risks assessment --scope staged

# Review assessment for critical/high risks
cat reports/risk-assessment-*.md

# If risks identified, create controls
r2r eac risks create reports/risk-assessment-*.md

# Implement controls in code
# ... develop features ...

# Link implementations
# Add @risk-control tags to specs

# Verify traceability before merge
r2r eac risks list --filter missing-links

# Generate compliance report
r2r eac risks list --json > compliance/risk-controls-$(date +%Y-%m-%d).json
```

## Debug Mode

Enable debug mode to inspect AI generation process:

```bash
r2r eac risks assessment --debug
r2r eac risks create assessment.md --debug
```

Creates debug files in `out/`:

```text
out/
├── debug-risk-assessment-context.md     # Full context sent to AI
├── debug-risk-assessment-prompt.md      # Complete AI prompt
├── debug-risk-assessment-response.md    # Raw AI response
├── debug-risk-profile-parsed.json       # Parsed risk data
├── debug-risk-profile-controls.json     # Generated control structures
└── debug-risk-profile-prompts/          # Individual control prompts
    ├── risk-001-prompt.md
    └── risk-002-prompt.md
```

**Use debug mode when:**

- AI generates unexpected assessments
- Risk severity seems incorrect
- Controls are not created as expected
- Customizing prompts or templates
- Troubleshooting path issues

## Best Practices

### Assessment

- **Assess early**: Run `risks assessment` before committing sensitive changes
- **Right scope**: Use `--scope staged` for incremental, `--scope all` for comprehensive
- **Review critically**: AI identifies potential risks, qualified personnel validate
- **Document decisions**: Save assessments in version control (`reports/` directory)

### Control Creation

- **Assessment-first**: Never create controls without assessment
- **Validate controls**: Review generated controls before committing
- **Use force carefully**: `--force` overwrites existing controls, use with caution
- **Organize by domain**: Keep domain organization consistent
- **Version control**: Commit controls with implementation changes

### Linkage Management

- **Tag consistently**: Use exact tag format `@risk-control:<name>-<id>`
- **Link implementations**: Always link scenarios to controls
- **Check regularly**: Use `risks list --filter orphaned` to find unlinked controls
- **Update proactively**: When controls change, update linked scenarios

### Compliance

- **Qualified review**: Always engage qualified personnel for regulatory controls
- **Regular cadence**: Quarterly reviews + event-driven updates
- **Audit trail**: Keep all assessments and controls in version control
- **Traceability**: Maintain clear links from risks → controls → implementations

## Common Compliance Frameworks

| Framework                 | Key Requirements                                                               | Example Control Tags                                                                                   |
| ------------------------- | ------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| **FDA 21 CFR Part 11**    | Electronic signatures, audit trails, validation, access controls               | `@risk-control:auth-mfa`, `@risk-control:audit-trail`, `@risk-control:doc-esignatures`                 |
| **ISO 13485 / IEC 62304** | Safety classification, risk management, V&V, traceability                      | `@risk-control:validation-csv`, `@risk-control:risk-assessment`                                        |
| **PCI-DSS**               | Protect cardholder data, secure systems, access control, monitoring            | `@risk-control:encrypt-rest`, `@risk-control:encrypt-transit`, `@risk-control:auth-mfa`                |
| **GDPR**                  | Data protection by design, right to be forgotten, consent, breach notification | `@risk-control:privacy-consent`, `@risk-control:privacy-erasure`, `@risk-control:incident-breach`      |
| **EU AI Act**             | Risk classification, transparency, human oversight, data governance            | `@risk-control:ai-bias`, `@risk-control:ai-explainability`, `@risk-control:ai-monitoring`              |
| **SOX**                   | Financial reporting controls, change management, access controls               | `@risk-control:access-review`, `@risk-control:change-control`, `@risk-control:segregation-duties`      |
| **ISO 27001**             | Information security management, risk assessment, controls                     | `@risk-control:asset-inventory`, `@risk-control:vulnerability-mgmt`, `@risk-control:incident-response` |

## Customization

### Custom Assessment Prompts

Override default AI behavior for risk assessment:

```bash
# Create custom prompt
cat > custom-risk-prompt.md << 'EOF'
Focus on these specific risks:
- PCI-DSS compliance gaps
- Payment data exposure
- Cardholder data handling
- Secure transmission requirements
EOF

# Use custom prompt
r2r eac risks assessment --prompt custom-risk-prompt.md
```

### Custom Control Generation Prompts

Customize control specification generation:

```bash
# Create custom control prompt
cat > custom-control-prompt.md << 'EOF'
Generate risk controls following PCI-DSS structure:
- Use MUST/SHALL language
- Reference PCI-DSS requirements explicitly
- Include testing requirements
EOF

# Use custom prompt
r2r eac risks create assessment.md --prompt custom-control-prompt.md
```

### AI Config Customization

Edit AI configurations in `.r2r/eac/ai/`:

```bash
# Modify assessment prompt
nano .r2r/eac/ai/risks-assessment/assessment.md

# Modify control creation prompt
nano .r2r/eac/ai/risks-create/create.md

# Changes apply to all future risk commands
r2r eac risks assessment
```

## Troubleshooting

| Problem                              | Solution                                                                   |
| ------------------------------------ | -------------------------------------------------------------------------- |
| AI generates too many risks          | Use `--scope staged` instead of `--scope all`, be more specific            |
| Missing specifications in assessment | Ensure specs exist in `specs/` directory, check file patterns              |
| Controls not created                 | Check assessment has "Risk Controls Needed" section, use `--debug`         |
| Path traversal error                 | Remove `..` from paths, use absolute paths or paths within repo            |
| Orphaned controls                    | Add `@risk-control` tags to implementation specs, verify with `risks list` |
| API key error                        | Run `r2r eac init` to configure AI provider                                |
| Controls overwritten                 | Avoid using `--force` unless intentional, back up existing controls        |
| Wrong domain                         | Manually move control file to correct domain directory                     |

## Advanced Usage

### Batch Assessment Processing

```bash
# Assess multiple scopes
for scope in staged changed all; do
  r2r eac risks assessment --scope $scope --destination reports/$scope/
done

# Process all assessments
r2r eac risks create reports/ --force
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Risk Assessment
  run: |
    r2r eac risks assessment --scope staged --json > risk-assessment.json
    if grep -q '"severity": "critical"' risk-assessment.json; then
      echo "Critical risks identified - blocking merge"
      exit 1
    fi

- name: Generate Controls
  run: |
    r2r eac risks create reports/ --allow-orphans

- name: Verify Linkages
  run: |
    r2r eac risks list --filter orphaned --json > orphaned.json
    if [ $(jq '.summary.orphaned' orphaned.json) -gt 0 ]; then
      echo "Warning: Orphaned controls found"
      cat orphaned.json
    fi
```

### Pre-commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash
if git diff --cached --name-only | grep -qE '\.(go|py|js|ts)$'; then
  echo "Running risk assessment on staged changes..."
  r2r eac risks assessment --scope staged

  # Review assessment
  latest_assessment=$(ls -t reports/risk-assessment-*.md | head -1)

  # Check for critical risks
  if grep -q "### Critical Risks" "$latest_assessment"; then
    echo "ERROR: Critical risks identified"
    echo "Review: $latest_assessment"
    exit 1
  fi
fi
```

### Compliance Reporting

```bash
# Generate comprehensive compliance report
r2r eac risks list --json > compliance-report.json

# Extract metrics
total_controls=$(jq '.summary.total' compliance-report.json)
orphaned=$(jq '.summary.orphaned' compliance-report.json)
coverage=$((($total_controls - $orphaned) * 100 / $total_controls))

echo "Risk Control Coverage: $coverage%"
```

## Integration with Other Commands

### With Specs Command

```bash
# Generate specification
r2r eac create spec "Payment processing feature"

# Assess risks in new spec
r2r eac risks assessment --scope changed

# Create controls
r2r eac risks create reports/risk-assessment-*.md

# Link implementation to control
# Edit specs/payment/payment.feature
# Add @risk-control:payment-validation-control-01 tag
```

### With Work Command

```bash
# Create workspace
r2r eac work create feature/payment-integration

# Develop feature
# ...

# Assess risks before commit
r2r eac risks assessment --scope staged

# Create controls if needed
r2r eac risks create reports/risk-assessment-*.md

# Commit with controls
r2r eac work commit --all

# Verify linkages before PR
r2r eac risks list --filter orphaned
r2r eac work pr
```

### With Test Command

```bash
# Create risk controls
r2r eac risks create assessment.md

# Implement control scenarios
# Add step definitions for @risk-control tags

# Run risk control test suite
r2r eac test suite risk-controls

# Verify all controls pass
r2r eac test --tag @risk-control
```

## Summary

**Complete workflow:**

1. **Assess**: `r2r eac risks assessment --scope staged`
2. **Review**: Examine generated assessment report
3. **Create**: `r2r eac risks create reports/risk-assessment-*.md`
4. **Implement**: Add `@risk-control` tags to implementation scenarios
5. **Verify**: `r2r eac risks list --filter orphaned`
6. **Test**: Run test suite for risk controls
7. **Document**: Commit assessments and controls to version control

**Remember:**

- Risk assessment is MANDATORY before creating controls
- Controls are living documents that evolve with threats
- Traceability is essential for compliance
- Always engage qualified personnel for regulatory controls
- Use version control for audit trail

Risk controls bridge the gap between risk identification and implementation verification, enabling continuous compliance and audit-ready traceability.

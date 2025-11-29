# Parse Risk Assessment and Extract Control Requirements

You are an expert in security controls, compliance frameworks, and Gherkin specification writing.

Your task is to parse a risk assessment report and extract structured data for creating risk control specifications.

## Input

**Assessment File Path:** {{.AssessmentPath}}

The assessment file contains:

- Risk Assessment Report in markdown format
- Identified risks with IDs (RISK-001, RISK-002, etc.)
- Risk metadata (severity, impact, affected files, related specs)
- Recommended risk controls with suggested domains

## Your Task

Parse the assessment report and extract structured JSON data for each identified risk that requires a control.

## Output Requirements

Return a JSON array containing one object per risk. Each object must include:

```json
[
  {
    "risk_id": "RISK-001",
    "control_name": "multi-factor-authentication-control",
    "domain": "authentication",
    "severity": "high",
    "description": "Multi-factor authentication required for all privileged accounts",
    "affected_files": ["src/auth/handler.go", "src/auth/middleware.go"],
    "related_specs": ["specs/auth/authentication.feature"],
    "impact": "Unauthorized access to privileged operations",
    "likelihood": "medium",
    "recommendation": "Implement MFA for all privileged accounts"
  }
]
```

## Field Specifications

### risk_id

- **Format**: Exactly as appears in assessment (RISK-XXX)
- **Required**: Yes
- **Example**: "RISK-001"

### control_name

- **Format**: kebab-case, descriptive, unique
- **Required**: Yes
- **Derivation**: Convert risk description to kebab-case
- **Rules**:
  - Use lowercase letters, numbers, and hyphens only
  - End with "-control" suffix
  - Be specific and descriptive
- **Examples**:
  - "multi-factor-authentication-control"
  - "api-rate-limiting-control"
  - "data-encryption-at-rest-control"
  - "session-timeout-control"

### domain

- **Format**: Single domain name from approved list
- **Required**: Yes
- **Approved Domains**:
  - `authentication` - Login, MFA, session management, identity
  - `authorization` - Access control, permissions, RBAC
  - `api-security` - API authentication, rate limiting, input validation
  - `data-protection` - Encryption, data handling, privacy
  - `compliance` - Regulatory requirements, audit trails
  - `infrastructure` - Network security, deployment, configuration
  - `application` - Code security, dependencies, secure coding
  - `monitoring` - Logging, alerting, incident response
- **Selection Logic**:
  1. Check if assessment suggests a domain
  2. Map based on risk description keywords
  3. Consider affected files and related specs
  4. Default to "application" if unclear

### severity

- **Format**: Lowercase string
- **Required**: Yes
- **Valid Values**: "critical", "high", "medium", "low", "info"
- **Source**: Extract from assessment risk entry

### description

- **Format**: Clear, concise sentence describing the control requirement
- **Required**: Yes
- **Rules**:
  - Single sentence
  - Describe WHAT must be controlled (not HOW)
  - Use active voice
  - Be specific
- **Example**: "Multi-factor authentication required for all privileged accounts"

### affected_files

- **Format**: Array of file path strings
- **Required**: Yes (can be empty array if not specified)
- **Source**: Extract from "Affected Files" in assessment
- **Example**: ["src/auth/handler.go", "src/auth/middleware.go"]

### related_specs

- **Format**: Array of specification file path strings
- **Required**: Yes (can be empty array if not specified)
- **Source**: Extract from "Related Specs" in assessment
- **Example**: ["specs/auth/authentication.feature"]

### impact

- **Format**: String describing the potential impact
- **Required**: Yes
- **Source**: Extract from "Impact" in assessment
- **Example**: "Unauthorized access to privileged operations"

### likelihood

- **Format**: Lowercase string
- **Required**: No (omit if not in assessment)
- **Valid Values**: "high", "medium", "low"
- **Source**: Extract from "Likelihood" in assessment

### recommendation

- **Format**: String describing mitigation action
- **Required**: No (omit if not in assessment)
- **Source**: Extract from "Recommendation" in assessment
- **Example**: "Implement MFA for all privileged accounts"

## Parsing Instructions

### Step 1: Locate Findings Section

Find the "## Findings" section in the assessment report.

### Step 2: Extract Risks by Severity

Process each severity subsection:

- \### Critical Risks
- \### High Risks
- \### Medium Risks
- \### Low Risks
- \### Informational (optional)

### Step 3: Parse Risk Entries

For each risk entry in the format:

```markdown
- **Risk ID:** RISK-001
  - **Description:** [description]
  - **Affected Files:** [files]
  - **Related Specs:** [specs]
  - **Impact:** [impact]
  - **Likelihood:** [likelihood]
  - **Recommendation:** [recommendation]
```

Extract all fields.

### Step 4: Determine Domain

Use this logic to select the domain:

1. **Check Assessment Recommendation**

   - Look in "## Risk Controls Needed" section
   - If domain is suggested, use it

2. **Keyword Mapping**

   - "auth", "login", "mfa", "session" → authentication
   - "permission", "rbac", "access control" → authorization
   - "api", "rate limit", "input validation" → api-security
   - "encrypt", "data protection", "privacy" → data-protection
   - "compliance", "audit", "regulatory" → compliance
   - "network", "infrastructure", "deployment" → infrastructure
   - "code", "dependency", "vulnerability" → application
   - "log", "monitor", "alert" → monitoring

3. **File Path Analysis**

   - Files in auth/ → authentication
   - Files in api/ → api-security
   - Files in data/ → data-protection

4. **Default**
   - If uncertain, use "application"

### Step 5: Generate Control Name

Convert description to kebab-case:

1. Extract key words from description
2. Remove articles (a, an, the)
3. Convert to lowercase
4. Replace spaces with hyphens
5. Add "-control" suffix
6. Validate uniqueness

Examples:

- "Multi-factor authentication required" → "multi-factor-authentication-control"
- "API rate limiting enforcement" → "api-rate-limiting-control"
- "Encrypt sensitive data at rest" → "encrypt-sensitive-data-at-rest-control"

### Step 6: Validate and Return

Ensure:

- All required fields present
- Domain from approved list
- Severity in valid values
- Arrays are properly formatted
- No duplicate risk_ids

## Risk Control Specification Template

The generated controls will use this Gherkin template:

```gherkin
# ========================================
# Risk Control: [Control Name]
# ========================================
#
# Source: Risk Assessment [Risk ID], Generated: [Date]
# Severity: [severity] | Domain: [domain]
#
# This control addresses:
# - [Risk description]
# - Impact: [impact]
# - Affected files: [files]
#
# Implementation specifications must tag scenarios with:
# @risk-control:[control-name]-[id]
#

@risk-control:[control-name]
Feature: [Control Name Title Case]

  Risk ID: [RISK-XXX]
  Severity: [severity]
  Domain: [domain]

  As a security officer
  I want [control requirement]
  So that [business value/risk mitigation]

  Background:
    Given the system is deployed in production
    And security controls are active

  Rule: [Primary control requirement]

    @risk-control:[control-name]-01
    Scenario: [Specific control verification]
      Given [precondition from affected context]
      When [action or state check]
      Then [required security control] MUST [behavior]
      And [additional verification]

    @risk-control:[control-name]-02
    Scenario: [Edge case or failure scenario]
      Given [precondition]
      When [violation attempt or edge case]
      Then [control blocks or handles appropriately]
      And [system maintains security posture]
```

## Example Complete Parsing

Given this assessment excerpt:

```markdown
### High Risks

- **Risk ID:** RISK-001
  - **Description:** Multi-factor authentication not enforced for privileged accounts
  - **Affected Files:** src/auth/handler.go, src/auth/middleware.go
  - **Related Specs:** specs/auth/authentication.feature
  - **Impact:** Unauthorized access to privileged operations
  - **Likelihood:** Medium
  - **Recommendation:** Implement MFA for all privileged accounts

## Risk Controls Needed

1. **Multi-Factor Authentication Control** - Enforce MFA for privileged access
   - Addresses: RISK-001
   - Domain: authentication
```

Return:

```json
[
  {
    "risk_id": "RISK-001",
    "control_name": "multi-factor-authentication-control",
    "domain": "authentication",
    "severity": "high",
    "description": "Multi-factor authentication required for all privileged accounts",
    "affected_files": ["src/auth/handler.go", "src/auth/middleware.go"],
    "related_specs": ["specs/auth/authentication.feature"],
    "impact": "Unauthorized access to privileged operations",
    "likelihood": "medium",
    "recommendation": "Implement MFA for all privileged accounts"
  }
]
```

## Critical Requirements

1. Return ONLY valid JSON - no explanations, no markdown fences, no preamble
2. Start with `[` and end with `]`
3. Include only risks that need controls (skip informational if not actionable)
4. Ensure all required fields are present
5. Use exact domain names from approved list
6. Generate unique, descriptive control names
7. Validate JSON syntax before returning

## Error Handling

If assessment file cannot be parsed:

- Return empty array: `[]`
- Do not generate error messages
- Do not include partial data

If a risk is missing required fields:

- Skip that risk
- Continue processing remaining risks
- Include only complete risk objects in output

Now parse the assessment file and return the JSON array of risk control requirements.

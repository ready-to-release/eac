# Control Tags
>
> **Risk and compliance control linkage**

Link scenarios to standardized security and compliance requirements using [NIST OSCAL](https://pages.nist.gov/OSCAL/) format.

---

## Control Tag Formats

**Single Control**: `@control:<control-id>`
**Multiple Controls**: `@controls:<id1>,<id2>`

---

## Control ID Format

**Pattern**: `<family>-<number>` or `<family>-<number>(<enhancement>)`

**Parts**:

- `<family>` - Control family (2-4 lowercase letters: `ac`, `au`, `ia`, `sc`, etc.)
- `<number>` - Control number (1+ digits: `2`, `12`, etc.)
- `<enhancement>` - Optional enhancement number in parentheses: `(1)`, `(10)`

**Examples**:

- `ac-2` - Account Management (NIST 800-53)
- `au-3` - Audit Record Content
- `ia-5(1)` - Password-Based Authentication (enhancement)

---

## Purpose

- Links test scenarios to OSCAL catalog controls
- Enables automated compliance evidence collection
- Provides standardized control traceability
- Supports audit and assessment reporting

---

## Examples

### Single Control

```gherkin
@ov @control:ac-2
Scenario: Account creation requires approval
  Given a user registration request
  When an administrator reviews the request
  Then the account should require approval
  And the approval should be logged
```

### Control with Enhancement

```gherkin
@ov @control:ia-5(1)
Scenario: Password authentication enforces complexity
  Given a user creating a password
  When the password is validated
  Then it must meet complexity requirements
  And weak passwords must be rejected
```

### Multiple Controls

```gherkin
@ov @controls:ac-2,au-3
Scenario: Account creation is audited
  Given an account creation request
  When the account is created
  Then an audit record must be created
  And the record must include timestamp, user ID, and admin approver
```

---

## Traceability and Evidence Collection

### Find scenarios for a control

```bash
# Single control
grep -r "@control:ac-2" specs/

# All account management controls
grep -r "@control:ac-" specs/
```

### Validate control tags

```bash
# Check all @control: tags reference valid catalog controls
r2r eac validate control-tags

# Output: Reports invalid control IDs with file locations
```

### Collect evidence

```bash
# Run tests
r2r eac test <module> --suite acceptance

# Collect compliance evidence
r2r create risk-assess <module> --profile specs/.risk-controls/<module>.profile.json

# Output: out/risk/<module>/assessment-results.json
# Contains: Controls + Test Evidence + Satisfied/Not-Satisfied Status
```

---

## OSCAL Profile Integration

Control tags work with OSCAL profiles to provide complete traceability:

```text
OSCAL Catalog          → Standard control definitions (NIST 800-53)
     ↓
OSCAL Profile          → Selected controls for your system
     ↓
@control: tags         → BDD scenarios verifying controls
     ↓
Test Results           → Cucumber JSON with pass/fail status
     ↓
Assessment Results     → OSCAL evidence linking controls to tests
```

### Profile Example

`specs/.risk-controls/auth-service.profile.json`:

```json
{
  "profile": {
    "metadata": { "title": "Authentication Service Controls" },
    "imports": [{
      "href": "../../../templates/specs/risk-catalog/controls.catalog.json",
      "include-controls": [{
        "with-ids": ["ac-2", "ac-3", "au-2", "ia-5", "ia-5(1)"]
      }]
    }]
  }
}
```

When you run `create-spec` with this profile, the AI automatically suggests applicable `@control:` tags.

---

## Migration from Old Format

### Deprecated (old format)

```gherkin
@risk-control:auth-mfa-01
Scenario: MFA required
```

### Current (OSCAL format)

```gherkin
@control:ia-2(1)
Scenario: Multi-factor authentication required
```

### Migration steps

1. Map old controls to OSCAL equivalents (e.g., `auth-mfa` → `ia-2(1)`)
2. Create OSCAL profile with selected controls
3. Replace `@risk-control:` tags with `@control:` tags
4. Run `validate control-tags` to verify
5. Collect evidence with `create risk-assess`

---

## Common NIST 800-53 Control Families

| Family | Description                          | Example Tags                        |
| ------ | ------------------------------------ | ----------------------------------- |
| **AC** | Access Control                       | `@control:ac-2`, `@control:ac-3`    |
| **AU** | Audit and Accountability             | `@control:au-2`, `@control:au-3`    |
| **IA** | Identification and Authentication    | `@control:ia-2`, `@control:ia-5(1)` |
| **SC** | System and Communications Protection | `@control:sc-7`, `@control:sc-8(1)` |
| **SI** | System and Information Integrity     | `@control:si-2`, `@control:si-10`   |
| **CM** | Configuration Management             | `@control:cm-2`, `@control:cm-6`    |
| **IR** | Incident Response                    | `@control:ir-4`, `@control:ir-6`    |

---

## When to Use Control Tags

### Always Use When

- Testing security requirements
- Validating access controls
- Audit and compliance scenarios
- Authentication and authorization
- Data protection and encryption
- Incident response procedures

### Example Mappings

**Access Control**:

- Account management → `@control:ac-2`
- Least privilege → `@control:ac-6`
- Session termination → `@control:ac-12`

**Audit**:

- Audit events → `@control:au-2`
- Audit record content → `@control:au-3`
- Audit review → `@control:au-6`

**Authentication**:

- Identification and authentication → `@control:ia-2`
- MFA → `@control:ia-2(1)`
- Authenticator management → `@control:ia-5`
- Password-based auth → `@control:ia-5(1)`

---

## Best Practices

### DO

- ✅ Link security and compliance scenarios to controls
- ✅ Use standardized OSCAL control IDs
- ✅ Validate control tags against catalog
- ✅ Collect automated evidence
- ✅ Create OSCAL profiles for your system
- ✅ Review control coverage regularly

### DON'T

- ❌ Use custom control naming schemes
- ❌ Skip control tag validation
- ❌ Store compliance evidence separately from tests
- ❌ Forget to update controls when requirements change
- ❌ Over-tag scenarios with too many controls

---

## Complete Example

```gherkin
@L2 @ov @control:ia-5(1)
Feature: auth_password-authentication
  Password-based authentication with complexity requirements

  Rule: Passwords must meet complexity requirements

    @ov @control:ia-5(1)
    Scenario: Strong password is accepted
      Given a user creating a password
      When the password is "P@ssw0rd123!"
      Then the password should be accepted
      And the password should be hashed
      And the hash should be stored securely

    @ov @control:ia-5(1)
    Scenario: Weak password is rejected
      Given a user creating a password
      When the password is "password"
      Then the password should be rejected
      And an error message should explain requirements
```

---

## Related Documentation

- [Risk Controls](../compliance/risk-controls.md) - Complete risk control documentation
- [GxP Tagging](../compliance/gxp-tagging.md) - Regulatory compliance tagging
- [Verification Tags](./verification-tags.md) - Types of validation

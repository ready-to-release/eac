# Control Tags

> **Risk and compliance control linkage**

Link scenarios to standardized security and compliance requirements using [NIST OSCAL](https://pages.nist.gov/OSCAL/) format.

---

## Overview

Control tags connect BDD specifications to compliance controls from established catalogs (NIST 800-53, ISO 27001, CIS, etc.).

### Purpose

- Links test scenarios to OSCAL catalog controls
- Enables automated compliance evidence collection
- Provides standardized control traceability
- Supports audit and assessment reporting

### Tag Formats

**Single Control**: `@control:<control-id>`
**Multiple Controls**: `@controls:<id1>,<id2>`

### Control ID Format

**Pattern**: `<family>-<number>` or `<family>-<number>(<enhancement>)`

**Examples**:

- `ac-2` - Account Management (NIST 800-53)
- `au-3` - Audit Record Content
- `ia-5(1)` - Password-Based Authentication (enhancement)

---

## When to Use Control Tags

### Always Use When

- Testing security requirements
- Validating access controls
- Audit and compliance scenarios
- Authentication and authorization
- Data protection and encryption
- Incident response procedures

### Example

```gherkin
@ov @control:ac-2
Scenario: Account creation requires approval
  Given a user registration request
  When an administrator reviews the request
  Then the account should require approval
  And the approval should be logged
```

---

## Reference Documentation

For complete CLI commands, OSCAL integration, and validation details, see:

**[Control Tags Reference](../../../reference/eac/compliance/control-tags.md)**

Complete implementation guide including:

- Validation commands (`r2r eac validate control-tags`)
- Evidence collection commands
- OSCAL profile integration
- Migration from old formats
- Common NIST 800-53 control families

---

## Related Documentation

- [Risk Controls](../compliance/risk-controls.md) - Complete risk control documentation
- [GxP Tagging](../compliance/gxp-tagging.md) - Regulatory compliance tagging
- [Verification Tags](./verification-tags.md) - Types of validation

# Specification Validation

CLI commands for validating Gherkin specifications and tag usage.

## Quick Reference

```bash
# Validate all specifications
r2r eac validate specs

# Validate with effective tag output
r2r eac validate specs --show-effective-tags

# Validate specific aspects
r2r eac validate test-tags
r2r eac validate control-tags
```

---

## `validate specs`

Validates Gherkin specifications against quality contracts.

```bash
r2r eac validate specs
```

**Checks**:

- Valid Gherkin syntax
- Required tags present (test level, verification type)
- Tag format correctness
- Feature file naming conventions
- Scenario structure compliance

### Options

```bash
# Show effective tags after inheritance
r2r eac validate specs --show-effective-tags

# Validate specific module
r2r eac validate specs --module eac-commands
```

---

## `validate test-tags`

Validates that all test tags are defined in the tag contract.

```bash
r2r eac validate test-tags
```

**Checks**:

- All `@L0`-`@L4` tags are valid test levels
- Verification tags (`@ov`, `@iv`, `@pv`, etc.) are valid
- Custom tags match defined patterns

---

## `validate control-tags`

Validates that `@control:` tags reference valid OSCAL catalog controls.

```bash
r2r eac validate control-tags
```

**Checks**:

- Control IDs exist in the OSCAL catalog
- Control format is valid (`<family>-<number>` or `<family>-<number>(<enhancement>)`)
- Reports invalid control IDs with file locations

**Example output**:

```text
Validating control tags...
  specs/auth-service/login.feature:15 - invalid: @control:ac-99 (not in catalog)

Validation failed: 1 invalid control reference found
```

---

## Effective Tag Display

When using `--show-effective-tags`, the output shows how tags accumulate through inheritance:

```text
specs/auth-service/login/specification.feature
  Feature: @L2 @ov @control:ia-5
    Scenario: User login with valid credentials
      Effective: @L2, @ov, @control:ia-5, @control:au-2
      (Scenario adds: @control:au-2)
```

This helps debug why tests are or aren't selected for specific suites.

---

## Related Documentation

- [Tag Taxonomy (Conceptual)](../../../explanation/specifications/taxonomy/index.md) - Tag system concepts
- [Tag Inheritance (Conceptual)](../../../explanation/specifications/taxonomy/tag-inheritance.md) - How tags accumulate
- [Validate Command Reference](../commands/validate/index.md) - Full validation command options

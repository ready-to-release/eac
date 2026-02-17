# Specification Validation

Commands for validating Gherkin specifications and tag usage.

---

## validate specs

Validates Gherkin specifications against quality contracts.

```bash
eac validate specs                    # Validate all specs
eac validate specs --show-effective-tags  # Show tag inheritance
eac validate specs --module <module>  # Validate specific module
```

**Validation checks**:

- ✓ Valid Gherkin syntax
- ✓ Required tags present (test level @L0-@L4)
- ✓ Tag format correctness
- ✓ Feature file naming conventions
- ✓ Scenario structure compliance

**Example output with `--show-effective-tags`**:

```text
specs/auth/login/specification.feature
  Feature: @L2 @ov @control:ia-5
    Scenario: User login with valid credentials
      Effective: @L2, @ov, @control:ia-5, @control:au-2
      (Scenario adds: @control:au-2)
```

---

## validate test-tags

Validates that test tags match defined tag taxonomy.

```bash
eac validate test-tags
```

**Validation checks**:

- ✓ Test level tags (@L0-@L4) are valid
- ✓ Verification tags (@ov, @iv, @pv) are defined
- ✓ Custom tags match allowed patterns

**Example failure**:

```text
specs/api/users.feature:25 - invalid tag: @L5 (valid: @L0-@L4)
```

---

## validate control-tags

Validates security control tags reference valid OSCAL catalog controls.

```bash
eac validate control-tags
```

**Validation checks**:

- ✓ Control IDs exist in OSCAL catalog
- ✓ Format is valid: `@control:<family>-<number>` or `@control:<family>-<number>(<enhancement>)`
- ✓ Reports invalid controls with file locations

**Valid formats**:

- `@control:ac-2` - Access Control family, control 2
- `@control:ia-5(1)` - Identification/Authentication family, control 5, enhancement 1
- `@control:au-2` - Audit and Accountability family, control 2

**Example failure**:

```text
specs/auth/login.feature:15 - invalid: @control:ac-99 (not in catalog)
```

---

## Related Documentation

- **[Specifications Index](./index.md)** - Tag system overview
- **[Tag Taxonomy](../../../explanation/specifications/taxonomy/index.md)** - Tag concepts
- **[Test Suites](../testing/test-suites.md)** - How tags select tests
- **[Validate Commands](../commands/validate/index.md)** - Full CLI reference

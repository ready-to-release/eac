# Specifications Reference

Technical reference for Gherkin specification validation and tag system.

## Specification Files

Specifications are written in Gherkin (`.feature` files) and stored in `specs/{module}/`.

**Format**: Gherkin BDD scenarios with tags

**Example location**: `specs/core/cache-invalidation/specification.feature`

---

## Tag System

Specifications use tags to control test selection and categorization:

| Tag Category           | Examples                    | Purpose                        |
| ---------------------- | --------------------------- | ------------------------------ |
| **Test Level**         | @L0, @L1, @L2, @L3, @L4     | Environment requirements       |
| **Verification Type**  | @ov, @iv, @pv               | What's being verified          |
| **Control Mapping**    | @control:ac-2, @control:ia-5 | OSCAL security control mapping |
| **Deployment Module**  | @depm:core, @depm:api       | Module under deployment        |
| **Dependencies**       | @deps:go, @deps:docker      | Required tools/runtimes        |
| **Environment**        | @env:isolated, @env:plte    | Test environment type          |
| **Execution Modality** | @Manual                     | Manual vs. automated           |

**See**: [Tag Taxonomy (Explanation)](../../../explanation/specifications/taxonomy/index.md)

---

## Validation Commands

```bash
# Validate all specifications
eac validate specs

# Validate test tags
eac validate test-tags

# Validate security control tags
eac validate control-tags

# Show effective tags after inheritance
eac validate specs --show-effective-tags
```

**See**: [Validation Details](./validation.md)

---

## Tag Inheritance

Tags inherit from Feature to Scenario:

```gherkin
@L1 @ov
Feature: User Authentication

  @control:ia-5
  Scenario: Login with valid credentials
    # Effective tags: @L1, @ov, @control:ia-5
```

---

## Related Documentation

- **[Validation Commands](./validation.md)** - Validation command reference
- **[Tag Taxonomy](../../../explanation/specifications/taxonomy/index.md)** - Tag system concepts
- **[Test Suites](../testing/test-suites.md)** - How tags select tests for suites
- **[Validate Commands](../commands/validate/index.md)** - CLI command reference

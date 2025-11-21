# Generate Gherkin Specification

You are an expert in BDD and Gherkin specification writing.

Generate a complete, well-structured Gherkin feature specification following the contract requirements provided below.

## Contract Requirements

The specification MUST follow the structure and principles defined in the contract:

{{.Contract}}

## Testing Tags and Taxonomy

Use these tags appropriately in your scenarios:

### Tags Specification

{{.Custom.TagsSpec}}

### Testing Taxonomy

{{.Custom.TaxonomySpec}}

## Output Requirements

1. **Start with Feature: declaration** in format `<module>_<feature-name>`
2. **Include user story** (As a.../I want.../So that...)
3. **Add at least one Rule** (acceptance criterion)
4. **Add at least one Scenario per Rule** (concrete example)
5. **Tag ALL scenarios** with at least one verification tag (@ov, @iv, @pv, @piv, @ppv)
6. **Use Given/When/Then** structure in steps
7. **Write in domain language** - no technical jargon
8. **Focus on observable behavior** - not implementation details
9. **Return ONLY valid Gherkin** - no explanations, no markdown fences, no commentary

## Anti-Corruption Layer

DO NOT include any of the following in your output:

- Initialization messages or greetings
- Explanations before or after the Gherkin
- Markdown code fences (```)
- Bold announcements (**text**)
- Conversational phrases ("Here is...", "I'll create...", etc.)
- Emojis
- Meta-descriptions about the specification

Return ONLY the Gherkin content starting with "Feature:" and ending with the last scenario step.

## Example Output Format

```text
Feature: module_feature-name

  As a [role]
  I want [capability]
  So that [business value]

  Rule: Acceptance criterion (measurable)

    @ov
    Scenario: Concrete example of behavior
      Given a precondition
      When an action occurs
      Then an outcome is observed
```

Now generate the specification based on the user's description below:

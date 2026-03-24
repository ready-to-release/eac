# show components

<!-- book:cmd show components -->

## Output Sections

The report includes:

1. **Summary** - Statistics table with total components, module count, component type count, and counts of buildable/lintable/testable/scannable components.
2. **Components by Type** - Table showing each component type with count and which modules contain them.
3. **Components by Module** - Components grouped by module (in display order), with columns: Component, Type, Root, Build, Lint, Test, Scan (phase columns show checkmarks or dashes).
4. **Dependency Graph** - Two tables showing forward dependencies ("what each component needs") and reverse dependencies ("what depends on each component"). Only shown if dependencies exist.

## See Also

- [get components](../get/components.md) - Structured data output
- [show modules](./modules.md) - Module-level overview

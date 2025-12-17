# Command Overview
Conceptual documentation explaining the design, organization, and usage patterns of EAC commands. These guides help you understand how commands work together and how to use them effectively.

## In This Section

| Guide                                         | Description                                                                         | Lines |
| --------------------------------------------- | ----------------------------------------------------------------------------------- | ----- |
| [Command Taxonomy](./command-taxonomy.md)     | How commands are organized into categories and the principles behind command design | 475   |
| [Naming Conventions](./naming-conventions.md) | Command naming patterns, consistency rules, and how to interpret command names      | 458   |
| [Common Flags](./common-flags.md)             | Global options available across all commands and flag usage patterns                | 603   |
| [Output Formats](./output-formats.md)         | Understanding JSON vs human-readable output, when to use each format                | 678   |

## Key Concepts

### Command Organization

Commands are organized by **verb-noun** patterns into 14 categories:

- **Information Commands**: `get` (JSON) and `show` (formatted)
- **Creation Commands**: `create` (AI-powered generation)
- **Validation Commands**: `validate` (contract checking)
- **Workflow Commands**: `work`, `test`, `build`, `pipeline`, `release`
- **Security Commands**: `scan`
- **Utility Commands**: `serve`, `templates`, `update`

### Design Principles

1. **Consistency**: Similar commands follow similar patterns
2. **Predictability**: Command names clearly indicate their purpose
3. **Duality**: Information available in both JSON and human-readable formats
4. **Composability**: Commands designed to work together in pipelines

### Quick Reference

For immediate command usage, see:

- [All Commands by Category](../categories/index.md)
- [Command Quick Reference](../index.md#command-categories)
- [Common Workflows](../index.md#common-workflows)

# Reference

**Information-oriented**: Technical reference material, specifications, and API documentation.

---

## Available Reference Documentation

### [Continuous Delivery](continuous-delivery/index.md)

Repository conventions, versioning, and commit message format.

**Topics covered:**

- Module structure and deployable units
- Semantic commit message format
- Semantic versioning (SemVer)
- Git tagging conventions
- Version increment rules

---

### [Specifications](specifications/index.md)

Quick reference for specification formats, commands, and syntax.

**Topics covered:**

- Gherkin format (specification.feature with Feature → Rule → Scenario structure)
- TDD format (unit test structure and Go test conventions)
- Godog commands (running specifications)
- Verification tags (IV/OV/PV classification)

---

### [VS Code Extension Reference](vscode-extension.md)

Quick reference for commands, actions, servers, file locations, and packaging.

**Topics covered:**

- Available actions and MCP servers
- Example workflows
- Common commands (setup, development, debugging)
- File locations
- Packaging and distribution
- JSON-RPC message examples
- Extension configuration

---

### [Commit Agent Pipeline Reference](commit-agent-pipeline.md)

Technical specifications for the 5-agent commit message generation system.

**Topics covered:**

- Component specifications (VSCode, MCP server, agents)
- Agent input/output formats
- Progress notification stages
- Validation rules
- Error messages
- Configuration options
- Performance metrics

---

### [Decision Records](decision-records/index.md)

Architectural decisions and their rationale.

**Decisions covered:**

- DR-001: Mono-Repository Layout
- DR-002: Semantic Versioning with Trunk-Based Development
- DR-003: Three-Layer Testing Approach (ATDD/BDD/TDD)
- DR-004: Diataxis Framework for Documentation Organization

---

## What is Reference?

Reference documentation is **information-oriented** content that provides technical descriptions of the system. It's like a dictionary - organized for easy lookup when you know what you're looking for.

**Characteristics:**

- Technical and precise
- Organized for quick lookup
- Consistent structure
- Comprehensive coverage
- Assumes you know what to look for

**Not finding what you need?**

- **Just getting started?** See [Tutorials](../tutorials/index.md)
- **Need to solve a problem?** See [How-to Guides](../how-to-guides/index.md)
- **Want to understand concepts?** See [Explanation](../explanation/index.md)

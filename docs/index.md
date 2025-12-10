# Everything as Code

> **Help turn every commit into deployable, compliant software you can trust**

---

## What is r2r (Ready to Release)?

**r2r** is an extensible CLI that enables Everything-as-Code workflows from your terminal, IDE, or CI/CD pipeline. Built by engineers, for engineers.

The CLI is your primary interface for:

- Writing executable specifications that validate your system
- Running continuous compliance checks on every commit
- Generating audit evidence as a byproduct of your pipeline
- Integrating with MCP servers and VSCode for IDE-native workflows
- Automating delivery flows with containers and GitHub Actions

**This repository is both the tool and a working example** - it demonstrates CI/CD implementation with the same principles and patterns explained in the documentation. Study the `.github/workflows/`, specs, and build processes to see Everything-as-Code in action.

## Why Everything as Code?

Traditional compliance creates friction: manual documentation, periodic audits, late validation. Development teams wait for approvals. Compliance teams scramble during audit prep. Quality suffers.

**The r2r CLI transforms compliance from a bottleneck into automation:**

- **Terminal-First**: Run validation and evidence generation from `r2r` commands
- **Shift-Left Compliance** - Catch issues at commit time (5 minutes) vs. production (days)
- **Executable Specifications** - Requirements and policies as code in version control
- **Continuous Validation** - Compliance checked on every commit, not quarterly
- **Automated Evidence** - Traceability generated automatically by your pipeline
- **Reference Implementation** - This repo's own CI/CD demonstrates the patterns

---

## Documentation Navigation

Documentation is organized using the [Diataxis framework](https://diataxis.fr/) - a systematic approach to technical documentation authoring:

<table>
<tr>
<td width="50%" valign="top" markdown="1">

### [Tutorials](tutorials/)

> **Learning-oriented guides**

Step-by-step lessons that take you through a series of steps to complete a project. Start here if you're new to the CLI and want hands-on guidance through core concepts.

- [Quick Start Guide](tutorials/getting-started/quick-start.md)
- [Your First Feature Specification](tutorials/getting-started/first-specification.md)

</td>
<td width="50%" valign="top" markdown="1">

### [How-to Guides](how-to-guides/)

> **Task-oriented recipes**

Guides that show you how to solve specific problems. Use these when you need to accomplish a particular task.

- [EAC Commands](how-to-guides/eac/commands/)
- [EAC Configuration](how-to-guides/eac/configuration/)
- [EAC Modules](how-to-guides/eac/modules/)

</td>
</tr>
<tr>
<td width="50%" valign="top" markdown="1">

### [Reference](reference/)

> **Information-oriented descriptions**

Technical reference material for looking up details. Check here for system architecture, module contracts, command syntax, configuration options, and specifications.

- [R2R and EAC Architecture](reference/r2r-eac/) - System overview and design
- [Repository Layout](reference/r2r-eac/repository-layout.md) - File structure
- [Command Reference](reference/commands/) - CLI commands
- [Decision Records](reference/decision-records/) - Architectural decisions

</td>
<td width="50%" valign="top" markdown="1">

### [Explanation](explanation/)

> **Understanding-oriented discussion**

Conceptual explanations that clarify and illuminate. Read these to understand the "why" behind the system.

- [Everything as Code](explanation/everything-as-code/)
- [Continuous Delivery Model](explanation/continuous-delivery/)
- [Compliance Transformation](explanation/transformation/)

</td>
</tr>
</table>

**Choose your path:**

- "I'm new and want to learn" → [Tutorials](tutorials/)
- "I need to accomplish a task" → [How-to Guides](how-to-guides/)
- "I need technical details" → [Reference](reference/)
- "I want to understand why" → [Explanation](explanation/)

---

## Working with Documentation

### Directory Structure

```text
docs/
├── index.md                    # This file
├── assets/                     # Binary files ONLY (.gif, .png, .pdf)
├── tutorials/                  # Learning-oriented guides
├── how-to-guides/              # Task-oriented recipes
├── reference/                  # Technical specifications
└── explanation/                # Conceptual discussions
```

# Everything as Code

> **Turn every commit into deployable, compliant software you can trust**

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

Documentation is organized using the [Diataxis framework](https://diataxis.fr/):

### [Tutorials](tutorials/index.md)

**Learning-oriented**: Step-by-step lessons for newcomers

Start here if you're new to the CLI and want hands-on guidance through core concepts and workflows.

### [How-to Guides](how-to-guides/index.md)

**Problem-oriented**: Recipes for specific tasks

Use these when you need to accomplish a specific task like writing specifications, setting up CI validation, or linking risk controls.

### [Reference](reference/index.md)

**Information-oriented**: Technical descriptions and specifications

Look here for command syntax, configuration options, Gherkin format details, and API specifications.

### [Explanation](explanation/index.md)

**Understanding-oriented**: Conceptual discussions and design rationale

Read these to understand the "why" behind continuous delivery models, compliance transformation, testing strategies, and architectural decisions.

**Choose based on what you need:**

- "I'm new and want to learn" → [Tutorials](tutorials/index.md)
- "I need to accomplish a task" → [How-to Guides](how-to-guides/index.md)
- "I need technical details" → [Reference](reference/index.md)
- "I want to understand why" → [Explanation](explanation/index.md)

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

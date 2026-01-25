# Overview

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

<!-- markdownlint-disable MD033 -->
<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 2em; margin: 2em 0;">
<div>
<h3><a href="tutorials/">Tutorials</a></h3>
<p><strong>Learning-oriented guides</strong></p>
<p>Step-by-step lessons that take you through a series of steps to complete a project. Start here if you're new to the CLI and want hands-on guidance through core concepts.</p>
<ul>
<li><a href="tutorials/getting-started/quick-start/">Quick Start Guide</a></li>
<li><a href="tutorials/getting-started/creating-your-first-extension/">Creating Your First Extension</a></li>
</ul>
</div>
<div>
<h3><a href="how-to-guides/">How-to Guides</a></h3>
<p><strong>Task-oriented recipes</strong></p>
<p>Guides that show you how to solve specific problems. Use these when you need to accomplish a particular task.</p>
<ul>
<li><a href="how-to-guides/eac/commands/">EAC Commands</a></li>
<li><a href="how-to-guides/eac/modules/">EAC Modules</a></li>
</ul>
</div>
<div>
<h3><a href="reference/">Reference</a></h3>
<p><strong>Information-oriented descriptions</strong></p>
<p>Technical reference material for looking up details. Check here for system architecture, module contracts, command syntax, configuration options, and specifications.</p>
<ul>
<li><a href="reference/eac/architecture/">EAC Architecture</a> - System overview and design</li>
<li><a href="reference/devex/internal/repository-layout.md">Repository Layout</a> - File structure</li>
<li><a href="reference/eac/commands/">Command Reference</a> - CLI commands</li>
<li><a href="reference/eac/decision-records/">Decision Records</a> - Architectural decisions</li>
</ul>
</div>
<div>
<h3><a href="explanation/">Explanation</a></h3>
<p><strong>Understanding-oriented discussion</strong></p>
<p>Conceptual explanations that clarify and illuminate. Read these to understand the "why" behind the system.</p>
<ul>
<li><a href="explanation/everything-as-code/">Everything as Code</a></li>
<li><a href="explanation/continuous-delivery/">Continuous Delivery Model</a></li>
<li><a href="explanation/transformation/">Compliance Transformation</a></li>
</ul>
</div>
</div>
<!-- markdownlint-enable MD033 -->

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

---

## Repository Modules

This repository is organized into modules - each representing a distinct deliverable with its own versioning, CI/CD pipeline, and release process.

<!-- book:insert modules-overview -->

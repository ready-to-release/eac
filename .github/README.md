# Everything as Code

## Turn every commit into deployable, compliant software you can trust

[![Release Status](https://github.com/ready-to-release/eac/actions/workflows/trigger.yml/badge.svg)](https://github.com/ready-to-release/eac/actions/workflows/trigger.yml)
[![GitHub Stars](https://img.shields.io/github/stars/ready-to-release/eac?style=social)](https://github.com/ready-to-release/eac/stargazers)
[![License](https://img.shields.io/github/license/ready-to-release/eac)](https://github.com/ready-to-release/eac/blob/main/LICENSE)
[![Docs](https://img.shields.io/badge/docs-GitHub_Pages-blue)](https://ready-to-release.github.io/eac/)
[![Latest Release](https://img.shields.io/github/v/release/ready-to-release/eac?include_prereleases)](https://github.com/ready-to-release/eac/releases)

---

## What is Everything as Code?

Everything as Code is an extensible CLI-in-a-box called r2r (Ready to Release) for managing software delivery flows with integrated MCP servers and VSCode extension. It brings continuous delivery, compliance automation, and executable specifications together into a unified workflow that works from your terminal, IDE, or CI/CD pipeline.

## Why Everything as Code?

Traditional compliance creates friction: manual documentation, periodic audits, late validation. Development teams wait for approvals. Compliance teams scramble during audit prep. Quality suffers.

Everything as Code transforms compliance from a bottleneck into an automated capability by:

- **Everything as Code** - Requirements, policies, and evidence in version control
- **Continuous Validation** - Compliance checked on every commit, not quarterly
- **Shift-Left Compliance** - Issues caught at commit time (5 minutes) vs. production (days)
- **Automated Evidence** - Evidence generated as a byproduct of your pipeline
- **Executable Specifications** - Requirements expressed as tests that validate your system

---

## Table of Content

Documentation is organized using the [Diataxis framework](https://diataxis.fr/):

### [Tutorials](docs/tutorials/index.md)

**Learning-oriented**: Step-by-step lessons for newcomers

Start here if you're new to the CLI and want hands-on guidance through core concepts and workflows.

### [How-to Guides](docs/how-to-guides/index.md)

**Problem-oriented**: Recipes for specific tasks

Use these when you need to accomplish a specific task like writing specifications, setting up CI validation, or linking risk controls.

### [Reference](docs/reference/index.md)

**Information-oriented**: Technical descriptions and specifications

Look here for command syntax, configuration options, Gherkin format details, and API specifications.

### [Explanation](docs/explanation/index.md)

**Understanding-oriented**: Conceptual discussions and design rationale

Read these to understand the "why" behind continuous delivery models, compliance transformation, testing strategies, and architectural decisions.

**Choose based on what you need:**

- "I'm new and want to learn" → [Tutorials](docs/tutorials/index.md)
- "I need to accomplish a task" → [How-to Guides](docs/how-to-guides/index.md)
- "I need technical details" → [Reference](docs/reference/index.md)
- "I want to understand why" → [Explanation](docs/explanation/index.md)

---

## Installation

Installation instructions coming soon.

---

## Maintainers

- Casper Leon Nielsen ([@casperease](https://github.com/casperease)
- Mikael Ottesen Hansen ([@miohansen](https://github.com/miohansen))

## Support and Community

Need help getting started or have questions?

- **Documentation**: Browse the [full documentation](https://ready-to-release.github.io/eac/) for guides and references
- **Issues**: Report bugs or request features on [GitHub Issues](https://github.com/ready-to-release/eac/issues/new)

---

## License

This project uses a dual-license structure:

### Software License

The r2r software (all source code) is licensed under the **MIT License**.

- **License**: MIT
- **Details**: See [LICENSE](LICENSE)
- **What it covers**: All source code in `src/`, build scripts, configuration files, and other software components

### Documentation License

The documentation is licensed under **Creative Commons Attribution-ShareAlike 4.0 International (CC BY-SA 4.0)**.

- **License**: CC BY-SA 4.0
- **Details**: See [docs/LICENSE](docs/LICENSE)
- **What it covers**: All documentation in `docs/`, including guides, tutorials, explanations, and reference materials

This dual-license approach allows the software to be freely used and modified under permissive terms, while ensuring documentation improvements are shared back with the community.

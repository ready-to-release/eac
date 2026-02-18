# Developer Experience

Cross-cutting reference documentation for developers working with EAC and CLIE.

This section serves two distinct audiences:

## Two Audiences

### Internal: Monorepo Contributors

**For developers contributing to the EAC repository itself.**

Working on the CLIE CLI, EAC commands, or supporting modules? The [Internal](./internal/index.md) section covers:

- Repository structure and layout
- CLI vs Extensions architecture
- Contract system overview
- Git workflow conventions

### External: EAC Adopters

**For developers using EAC in their own projects.**

Adopting EAC in your repository? The [External](./external/index.md) section covers:

- Getting started with EAC
- Configuration options
- Recommended project structure

## Quick Navigation

| Audience         | Section                         | Key Topics                              |
| ---------------- | ------------------------------- | --------------------------------------- |
| **Contributors** | [Internal](./internal/index.md) | Repository layout, contracts, workflows |
| **Adopters**     | [External](./external/index.md) | Setup, configuration, structure         |

## Which Section Do I Need?

**Use Internal if you are:**

- Adding new EAC commands
- Modifying the CLIE CLI
- Working on eac-core or eac-commands modules
- Contributing to the EAC repository

**Use External if you are:**

- Setting up EAC in a new project
- Configuring EAC for your team
- Integrating EAC into your CI/CD pipeline
- Using EAC commands in your workflow

# Repository Layout and Module Structure

## Overview

This repository is organized as a monorepo with clearly defined module boundaries. Understanding module structure is essential for creating accurate semantic commit messages and version increments.

## Deployable Units vs Supporting Modules

The repository distinguishes between two categories:

### Deployable Units

Independently built, versioned, and deployed artifacts. Each has a detailed contract in `contracts/deployable-units/0.1.0/{moniker}.yml` defining:

- Build and deployment configuration
- Versioning strategy and current version
- Runtime dependencies and environment
- Integration points and APIs

---

## References

- [Trunk-Based Development](../../explanation/continuous-delivery/trunk-based-development.md)

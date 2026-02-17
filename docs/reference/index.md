# Reference

Technical reference for **EAC** (Everything-as-Code) and **CLIE** (CLI Extender).

## Overview

**EAC** automates quality engineering workflows—build, test, scan, validate, document, and release—for modular codebases. All configuration, architecture, and specifications are defined as code using YAML contracts validated against JSON schemas.

**CLIE** is the containerized CLI framework that runs EAC in isolated Docker containers, providing reproducible, platform-independent execution. EAC can also run independently via the MCP server for AI tool integration.

## Quick Access

| Resource                                             | Description                              |
| ---------------------------------------------------- | ---------------------------------------- |
| [Command Cheat Sheet](./eac/commands/cheat-sheet.md) | Quick reference for most common commands |

## Products

| Product Reference               | Description                             |
| ------------------------------- | --------------------------------------- |
| [CLIE CLI](./clie/index.md)     | CLIE CLI framework and commands         |
| [EAC Extension](./eac/index.md) | EAC extension architecture and commands |

## Developer Experience

| Resource                                      | Description                   |
| --------------------------------------------- | ----------------------------- |
| [For Contributors](./devex/internal/index.md) | Repository contributors guide |
| [For Adopters](./devex/external/index.md)     | EAC adoption guide            |

## EAC Extension

| Resource                                                  | Description                     |
| --------------------------------------------------------- | ------------------------------- |
| [Continuous Delivery](./eac/continuous-delivery/index.md) | CI/CD workflows and conventions |
| [Decision Records](./eac/decision-records/index.md)       | Architectural decisions         |
| [Modules](./eac/modules/index.md)                         | Complete module reference       |

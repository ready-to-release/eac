# Reference

Technical reference for **EAC** (Everything-as-Code) and **CLIE** (CLI Extender).

## Overview

**EAC** automates quality engineering workflows - build, test, scan, validate, document, and release - for modular codebases. All configuration, architecture, and specifications are defined as code using YAML contracts validated against JSON schemas.

**CLIE** is an optional containerized CLI framework that can run EAC in isolated Docker containers for reproducible, platform-independent execution. EAC also integrates with LLM tools via its MCP server.

## Quick Access

| Resource                                             | Description                              |
| ---------------------------------------------------- | ---------------------------------------- |
| [Command Cheat Sheet](./eac/cheat-sheet.md) | Quick reference for most common commands |

## Products

| Product Reference               | Description                             |
| ------------------------------- | --------------------------------------- |
| [EAC CLI](./eac/index.md)       | EAC CLI architecture and commands       |
| [CLIE CLI](./clie/index.md)     | CLIE extension host framework           |

## Developer Experience

| Resource                                      | Description                   |
| --------------------------------------------- | ----------------------------- |
| [For Contributors](./devex/internal/index.md) | Repository contributors guide |
| [For Adopters](./devex/external/index.md)     | EAC adoption guide            |

## EAC CLI

| Resource                                                  | Description                     |
| --------------------------------------------------------- | ------------------------------- |
| [Continuous Delivery](./eac/continuous-delivery/index.md) | CI/CD workflows and conventions |
| [Decision Records](./eac/architecture/decisions/index.md)       | Architectural decisions         |
| [Modules](./eac/modules/index.md)                         | Complete module reference       |

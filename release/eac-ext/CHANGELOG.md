# Changelog

All notable changes to **eac-ext** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!-- Release Highlights (since 0.0.1)

  Tool System & Containers:
  - Executor mode: global override (auto/container/system) controls how containerized tools run
  - Container credential forwarding: GITHUB_TOKEN, GOPRIVATE, NPM_TOKEN, SEMGREP_APP_TOKEN via allowlist
  - Subprocess routing: container-based tools use process groups for clean termination
  - tool-config.yml modernization: credentials section, executor-mode, tool-bindings keys

  Container Infrastructure:
  - Container component auto-discovery for -oci suffix containers
  - Standardized container naming conventions

-->

## [0.0.1] - 2026-02-20

### Added

- Initial container-based Docker extension with CI/CD workflows
- Build dependency system for module-to-module build ordering with dry-run support
- Post-build steps and npm testing support
- Unified release and CI orchestration system
- Spec infrastructure and test infrastructure for extension validation

### Changed

- Standardized container image tagging and naming conventions
- Centralized EAC configuration structure with registry consolidation
- Containerized CI workflows with concurrency controls and release gating
- Restructured monorepo from src/ to go/eac and go/r2r layout
- Migrated MkDocs config to template-based architecture
- Enhanced CI workflows with artifact resolution and diagnostic links

[Unreleased]: https://github.com/ready-to-release/eac/compare/eac-ext/0.0.1...HEAD
[0.0.1]: https://github.com/ready-to-release/eac/releases/tag/eac-ext/0.0.1

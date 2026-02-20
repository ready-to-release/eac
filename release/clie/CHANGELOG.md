# Changelog

All notable changes to **clie** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

<!-- Release Highlights (since 0.0.1)

  Internal:
  - Relocated from go/clie to go/cli/clie (no user-facing changes)
  - Optimized startup time
  - Allow unauthenticated public extension discovery

-->

## [0.0.1] - 2026-02-20

### Added

- Initial CLI with core functionality and cross-platform binary releases (linux/amd64, darwin/arm64, windows/amd64)
- Multi-module framework with unified release and CI orchestration system
- Comprehensive build orchestration with build summaries and artifact handling
- Release automation with changelog system, gating, and concurrency controls
- Books support and documentation infrastructure
- AI configuration framework and design command
- Template tagging system and template installation commands
- Risk profile command and architecture validation

### Changed

- Standardized binary naming conventions across platforms
- Centralized contracts, testing, and configuration into unified system
- Restructured testing framework with module isolation and test coverage expansion
- Containerized CI workflows with go-bvt-image
- Migrated workflow tagging logic to Go commands
- Reorganized modules by dependency layer

[Unreleased]: https://github.com/ready-to-release/eac/compare/clie/0.0.1...HEAD
[0.0.1]: https://github.com/ready-to-release/eac/releases/tag/clie/0.0.1

# Changelog

All notable changes to **ext-eac** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.7] - 2025-12-11

### Fixed

- fix(multi-module): enhance release and test infrastructure

## [0.0.6] - 2025-12-10

### Added

- feat(multi-module): enhance build system with dependencies and dry-run (#29)

### Changed

- chore(multi-module): consolidate AI module into eac-commands
- ci: fix
- chore(multi-module): refactor workflows and add build summaries
- refactor(multi-module): migrate mkdocs config to templates (#35)
- chore(multi-module): add books support and documentation infrastructure
- refactor(multi-module): simplify template resolution logic
- chore(deps): bump golang from 1.24-alpine to 1.25-alpine in /containers/ext-eac (#18)
- chore(deps): bump actions/checkout from 4 to 6 (#19)
- chore(multi-module): standardize template tagging system
- chore(multi-module): standardize CI/CD workflows and build system

### Fixed

- fix(multi-module): enhance ci workflows and error handling

## [0.0.5] - 2025-12-02

### Added

- feat(multi-module): add post-build steps and npm testing support

### Changed

- chore(multi-module): unify CI/CD and add repository config
- refactor: restructure monorepo from src/ to go/eac and go/r2r

## [0.0.4] - 2025-12-01

### Added

- feat(multi-module): add ext-eac Docker extension and CI/CD workflows
- feat(multi-module): restructure specs and test infrastructure
- feat: commands for design and docs

### Changed

- refactor(multi-module): reorganize commands with verb-first structure
- refactor(multi-module): simplify module contract structure
- refactor(multi-module): reorganize registry and consolidate dependencies
- refactor(multi-module): standardize binary naming
- refactor(multi-module): standardize platform-specific binary naming
- refactor(multi-module): centralize EAC configuration structure
- chore(multi-module): enhance CI workflows and testing infrastructure
- chore(multi-module): migrate workflow tagging logic to Go commands
- chore(multi-module): enhance CI workflows and update build utilities
- chore(multi-module): simplify CI verification in release workflows
- ci: summaries
- ci: fix
- ci(multi-module): containerize CI workflows with go-bvt-image
- ci: fix
- ci: fix
- chore(multi-module): rename CI workflows and add release automation
- chore(multi-module): reorganize ai module and normalize line endings
- chore(multi-module): update dependencies and module type references
- chore(multi-module): establish module ownership of CI workflows
- ci(github): enable caching and improve release CI verification
- ci: fix
- ci: fix
- ci: fix
- chore(multi-module): enhance test infrastructure
- docs(readme): remove duplicate module-level readmes
- docs: added support for interactive drawio drawing

### Fixed

- fix: release

## [0.0.3] - 2025-12-01

### Changed

- refactor(multi-module): standardize container image tagging

## [0.0.2] - 2025-11-29

### Added

- Initial container-based extension release

[Unreleased]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.7...HEAD
[0.0.7]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.6...ext-eac/0.0.7
[0.0.6]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.5...ext-eac/0.0.6
[0.0.5]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.4...ext-eac/0.0.5
[0.0.4]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.3...ext-eac/0.0.4
[0.0.3]: https://github.com/ready-to-release/eac/compare/ext-eac/0.0.2...ext-eac/0.0.3
[0.0.2]: https://github.com/ready-to-release/eac/releases/tag/ext-eac/0.0.2

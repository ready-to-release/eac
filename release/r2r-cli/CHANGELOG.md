# Changelog

All notable changes to **r2r-cli** will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.21] - 2025-12-10

### Fixed

- fix(eac-commands): improve artifact logging and error message display
- fix(eac-commands): improve mkdocs book build logic and logging

## [0.0.19] - 2025-12-10

### Changed

- chore(deps): bump golang.org/x/sys from 0.38.0 to 0.39.0 in /go/r2r/cli (#47)
- chore(deps): bump github.com/spf13/cobra from 1.10.1 to 1.10.2 in /go/r2r/cli (#48)
- refactor(multi-module): consolidate AI configuration framework
- chore(multi-module): expand documentation and enhance artifact handling
- chore(multi-module): refactor workflows and add build summaries
- refactor(multi-module): unify risk profile command and architecture
- chore(multi-module): refactor CI/CD and improve cross-platform paths
- chore(multi-module): add books support and documentation infrastructure
- refactor(multi-module): implement test isolation for work-pull command
- chore(deps): bump actions/checkout from 4 to 6 (#19)
- chore(deps)!: bump actions/upload-artifact from 4 to 5 (#21)
- chore(multi-module): standardize template tagging system
- chore(deps)!: bump actions/download-artifact from 4 to 6 (#5)
- chore(deps): bump golang.org/x/sys from 0.37.0 to 0.38.0 in /go/r2r/cli (#12)
- chore(deps): bump github.com/docker/docker from 28.0.0+incompatible to 28.5.2+incompatible in /go/r2r/cli (#14)
- ci: add acceptance

### Fixed

- fix(multi-module): enhance ci workflows and error handling

## [0.0.17] - 2025-12-02

### Changed

- chore(multi-module): unify CI/CD and add repository config

## [0.0.16] - 2025-12-01

### Changed

- refactor(multi-module): add release automation and changelog system
- refactor(multi-module): add r2r-config module and update contracts
- refactor(r2r-cli): migrate example configs to contract documentation
- refactor(multi-module): simplify module contract structure
- refactor(multi-module): standardize binary naming
- refactor(multi-module): standardize platform-specific binary naming
- refactor(multi-module): centralize contracts, testing, and configuration
- refactor(multi-module): centralize contracts and testing
- refactor(multi-module): migrate to centralized contracts repository
- refactor(multi-module): migrate to centralized contract loading
- perf(multi-module): optimize parallelization and resource management
- test(multi-module): expand test coverage and validation framework
- test(multi-module): expand test coverage and validation framework
- chore(multi-module): restructure testing framework and add module isolation
- chore(multi-module): migrate workflow tagging logic to Go commands
- ci: removed redundant workflow and made the change trigger more dynamic
- chore(multi-module): enhance CI workflows and update build utilities

## [0.0.15] - 2025-12-01

### Added

- feat(multi-module): add r2r-cli release workflow
- feat(multi-module): restructure docs and update CLI design
- feat(multi-module): restructure specs and test infrastructure
- feat(multi-module): refactor design command with AI
- feat(multi-module): implement comprehensive build orchestration and markdown validation

### Changed

- refactor(multi-module): add release automation and changelog system
- refactor(multi-module): add r2r-config module and update contracts
- refactor(r2r-cli): migrate example configs to contract documentation
- refactor(multi-module): simplify module contract structure
- refactor(multi-module): standardize binary naming
- refactor(multi-module): standardize platform-specific binary naming
- refactor(multi-module): centralize contracts, testing, and configuration
- refactor(multi-module): centralize contracts and testing
- refactor(multi-module): migrate to centralized contracts repository
- refactor(multi-module): migrate to centralized contract loading
- perf(multi-module): optimize parallelization and resource management
- test(multi-module): expand test coverage and validation framework
- test(multi-module): expand test coverage and validation framework
- chore(multi-module): restructure testing framework and add module isolation
- chore(multi-module): migrate workflow tagging logic to Go commands
- ci: removed redundant workflow and made the change trigger more dynamic
- chore(multi-module): enhance CI workflows and update build utilities
- ci: fix
- ci: summaries
- ci: fix
- ci(multi-module): containerize CI workflows with go-bvt-image
- chore(multi-module): rename CI workflows and add release automation
- chore(multi-module): reorganize ai module and normalize line endings
- chore(multi-module): introduce orchestrator for parallel testing
- chore(multi-module): establish module ownership of CI workflows
- ci(github): enable caching and improve release CI verification
- chore(multi-module): standardize build and test command structure
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- ci: fix
- chore(multi-module): enhance test infrastructure
- chore(multi-module): restructure design directories
- chore(multi-module): enhance contracts and add validation
- docs: feat: everything-as-code documentation added

### Fixed

- fix: bad deps and leftover code
- fix: ci
- fix: ci
- fix: release
- fix: manual release
- fix: release
- fix: go mod

## [0.0.14] - 2025-12-01

### Changed

- refactor(multi-module): add handlers configuration system
- refactor(multi-module): simplify module contract structure
- refactor(multi-module): reorganize registry and consolidate dependencies
- refactor(multi-module): standardize binary naming
- refactor(multi-module): standardize platform-specific binary naming

## [0.0.13] - 2025-11-29

### Changed

- refactor(multi-module): standardize binary naming conventions

## [0.0.11] - 2025-11-27

### Added

- Initial cross-platform binary releases

## [0.0.10] - 2025-11-27

### Added

- Initial release with core CLI functionality

[Unreleased]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.21...HEAD
[0.0.21]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.19...r2r-cli/0.0.21
[0.0.19]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.17...r2r-cli/0.0.19
[0.0.17]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.16...r2r-cli/0.0.17
[0.0.16]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.15...r2r-cli/0.0.16
[0.0.15]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.14...r2r-cli/0.0.15
[0.0.14]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.13...r2r-cli/0.0.14
[0.0.13]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.11...r2r-cli/0.0.13
[0.0.11]: https://github.com/ready-to-release/eac/compare/r2r-cli/0.0.10...r2r-cli/0.0.11
[0.0.10]: https://github.com/ready-to-release/eac/releases/tag/r2r-cli/0.0.10

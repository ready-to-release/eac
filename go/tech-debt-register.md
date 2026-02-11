# Tech Debt Register

Extracted from `go/**/README.md` Code Health sections.

| Category                   | Count |
| -------------------------- | ----- |
| Tech Debt items            | 119   |
| Pain Points                | 112   |
| Optimization Opportunities | 95    |

---

## Tech Debt

### adapters

#### TD-001 -- Global mutable schema validator state

- **Package:** ai
- **Source:** `go/adapters/ai/README.md`
- **Detail:** Package-level `schemaValidator` and `schemaValidatorOnce` in config_loader.go are global mutable state coupled to the workspace root of the first caller
- **Severity:** medium

#### TD-002 -- Toolhandler handler.go oversized at 460 lines

- **Package:** toolhandler
- **Source:** `go/adapters/ai/toolhandler/README.md`
- **Detail:** handler.go ~460 lines
- **Severity:** low

#### TD-003 -- Execute function too long and global init registration

- **Package:** behave
- **Source:** `go/adapters/behave/README.md`
- **Detail:** Execute() ~170 lines, init() global registration
- **Severity:** medium

#### TD-004 -- Execute function too long and global init registration

- **Package:** cucumber
- **Source:** `go/adapters/cucumber/README.md`
- **Detail:** Execute() ~130 lines, init() global registration
- **Severity:** medium

#### TD-005 -- Wide interface and oversized function with global vars

- **Package:** docker
- **Source:** `go/adapters/docker/README.md`
- **Detail:** 13-method interface, executeOnce ~175 lines, 4 package-level mutable vars
- **Severity:** medium

#### TD-006 -- Inconsistent execution API and naive duration parsing

- **Package:** dotnet
- **Source:** `go/adapters/dotnet/README.md`
- **Detail:** runner.go:180-202 manually pipes stdout/stderr and calls cmd.Start/cmd.Wait, while the restore step at line 144 uses the higher-level tool.GlobalExecutor().Execute. trx_parser.go:77-83 reformatTRXDuration does naive string splitting without validating numeric parts
- **Severity:** medium

#### TD-007 -- New constructor calls globals directly

- **Package:** eac
- **Source:** `go/adapters/eac/README.md`
- **Detail:** New() calls globals directly
- **Severity:** low

#### TD-008 -- Cache wrapper length and mixed concerns

- **Package:** godog
- **Source:** `go/adapters/godog/README.md`
- **Detail:** cache.go ~200 lines of wrappers, buildMockingEnvironment mixes concerns
- **Severity:** medium

#### TD-009 -- Global init registration and oversized Execute function

- **Package:** gotest
- **Source:** `go/adapters/gotest/README.md`
- **Detail:** init() global registration, Execute() ~120 lines, package-level var
- **Severity:** medium

#### TD-010 -- Oversized Execute function and global init registration

- **Package:** mocha
- **Source:** `go/adapters/mocha/README.md`
- **Detail:** Execute() ~124 lines, init() global
- **Severity:** medium

#### TD-011 -- Exported mutex and no unit tests

- **Package:** npm
- **Source:** `go/adapters/npm/README.md`
- **Detail:** exported NpmInstallMu mutex, no unit tests
- **Severity:** medium

#### TD-012 -- Silently ignored sync errors and shallow change detection

- **Package:** nuget
- **Source:** `go/adapters/nuget/README.md`
- **Detail:** isolation.go:84-88 silently ignores errors from syncDirectory for src/, test/, tests/ directories. isolation.go:157-177 projectFilesChanged only checks top-level glob matches, so nested .csproj files in subdirectories are not detected
- **Severity:** medium

#### TD-013 -- Exported mutex and no unit tests

- **Package:** pip
- **Source:** `go/adapters/pip/README.md`
- **Detail:** exported PipInstallMu mutex, no unit tests
- **Severity:** medium

#### TD-014 -- Oversized Execute function and global init registration

- **Package:** pytest
- **Source:** `go/adapters/pytest/README.md`
- **Detail:** Execute() ~150 lines, init() global
- **Severity:** medium

#### TD-015 -- FindTestRoot returns first component regardless of type

- **Package:** reqnroll
- **Source:** `go/adapters/reqnroll/README.md`
- **Detail:** runner.go:112-117 FindTestRoot returns the first component root it finds regardless of component type, which could be incorrect if a module has multiple components
- **Severity:** medium

#### TD-016 -- Extremely large Update and handleMouse and Model

- **Package:** console
- **Source:** `go/adapters/tui/console/README.md`
- **Detail:** Update() ~530 lines, handleMouse() ~310 lines, Model ~1160 lines
- **Severity:** low

#### TD-017 -- Oversized Start function and verbose struct mapping

- **Package:** tui
- **Source:** `go/adapters/tui/README.md`
- **Detail:** ParallelConsole.Start() ~200 lines, verbose struct mapping, bootstrap init()
- **Severity:** medium

### cli/clie

#### TD-018 -- Missing nil-guard and fragile SHA extraction

- **Package:** cmd
- **Source:** `go/cli/clie/cmd/README.md`
- **Detail:** nil-guard missing on parsedCmd, fragile SHA extraction, test file in production build
- **Severity:** high

#### TD-019 -- Hardcoded grammar TODOs in command parser

- **Package:** command-parser
- **Source:** `go/cli/clie/internal/command-parser/README.md`
- **Detail:** 3 TODOs for hardcoded grammar
- **Severity:** low

#### ~~TD-020~~ -- ~~Unused GetExtensions function~~ (RESOLVED)

- **Package:** conf
- **Source:** `go/cli/clie/internal/conf/README.md`
- **Detail:** ~~GetExtensions() never called~~ Removed unused stub method and its test.

#### ~~TD-021~~ -- ~~Duplicate Ping logic and~~ reimplemented stdlib functions (PARTIALLY RESOLVED)

- **Package:** docker (clie)
- **Source:** `go/cli/clie/internal/docker/README.md`
- **Detail:** duplicate Ping logic remains. ~~reimplements stdlib string functions~~ Replaced custom `contains`, `containsSubstring`, `containsString`, `splitLines`, `lastIndex` with `strings.Contains`, `slices.Contains`, `strings.Split`, `strings.LastIndex`.
- **Severity:** low

#### TD-022 -- pullImage creates its own Docker client

- **Package:** tui (clie)
- **Source:** `go/cli/clie/internal/tui/README.md`
- **Detail:** pullImage creates own Docker client
- **Severity:** low

#### TD-023 -- TODOs for subcommand flags and debug logging

- **Package:** validator
- **Source:** `go/cli/clie/internal/validator/README.md`
- **Detail:** 2 TODOs for subcommand flags and debug logging
- **Severity:** low

### cli/eac

#### TD-024 -- Oversized main and exported mutable global

- **Package:** eac
- **Source:** `go/cli/eac/README.md`
- **Detail:** main() ~147 lines, InitialWorkingDir exported mutable global
- **Severity:** medium

#### TD-025 -- Oversized worker functions and no tests for 4 files

- **Package:** build
- **Source:** `go/cli/eac/impl/build/README.md`
- **Detail:** buildUnitWorker ~285 lines, Build() ~156 lines, package-level mutex, no tests for 4 files
- **Severity:** medium

#### TD-026 -- Oversized go.go and buildx Build function

- **Package:** builders
- **Source:** `go/cli/eac/impl/build/builders/README.md`
- **Detail:** 2 TODOs, go.go 665 lines, buildx Build ~115 lines
- **Severity:** medium

#### TD-027 -- Many format functions and oversized state machine

- **Package:** content
- **Source:** `go/cli/eac/impl/build/docprep/content/README.md`
- **Detail:** ~20 Format\* functions, ParseHelpOutput ~121 lines state machine
- **Severity:** low

#### TD-028 -- No test file for drawio.go

- **Package:** diagrams
- **Source:** `go/cli/eac/impl/build/docprep/diagrams/README.md`
- **Detail:** no test for drawio.go
- **Severity:** low

#### TD-029 -- Oversized GenerateNavForDir and limited testing

- **Package:** navigation
- **Source:** `go/cli/eac/impl/build/docprep/navigation/README.md`
- **Detail:** GenerateNavForDir ~109 lines, single test file
- **Severity:** low

#### TD-030 -- No test file for orphan.go

- **Package:** staging
- **Source:** `go/cli/eac/impl/build/docprep/staging/README.md`
- **Detail:** orphan.go has no test file
- **Severity:** low

#### TD-031 -- Oversized file with no tests for context and formatter

- **Package:** commit-message
- **Source:** `go/cli/eac/impl/create/commit-message/README.md`
- **Detail:** 560+ lines with 20+ functions, no test for context/formatter
- **Severity:** medium

#### TD-032 -- Hardcoded StandardCommitTypes

- **Package:** commitmessage (internal)
- **Source:** `go/cli/eac/impl/create/commit-message/internal/README.md`
- **Detail:** hardcoded StandardCommitTypes
- **Severity:** low

#### TD-033 -- Oversized create.go with global mock state

- **Package:** create/design
- **Source:** `go/cli/eac/impl/create/design/README.md`
- **Detail:** create.go 530 lines, global mock state, buildContractBasedPrompt ~72 lines
- **Severity:** medium

#### TD-034 -- Oversized assess.go with no unit tests for 4 files

- **Package:** create/risk-assess
- **Source:** `go/cli/eac/impl/create/risk-assess/README.md`
- **Detail:** assess.go 731 lines, CreateRiskAssess ~145 lines, no unit tests for 4 files
- **Severity:** medium

#### TD-035 -- Oversized function and no unit tests

- **Package:** create/squash-message
- **Source:** `go/cli/eac/impl/create/squash-message/README.md`
- **Detail:** CreateSquashMessage ~127 lines, no unit tests for formatter/validator
- **Severity:** medium

#### TD-036 -- Oversized validator with no tests for export and mock

- **Package:** design
- **Source:** `go/cli/eac/impl/design/README.md`
- **Detail:** validator.go 631 lines, export.go 446 lines no tests, mock 319 lines
- **Severity:** medium

#### TD-037 -- Fragile string slice check in mock

- **Package:** design (helper)
- **Source:** `go/cli/eac/impl/design/helper/README.md`
- **Detail:** fragile string slice check in mock
- **Severity:** low

#### TD-038 -- Oversized container function with no unit tests

- **Package:** docs/helper
- **Source:** `go/cli/eac/impl/docs/helper/README.md`
- **Detail:** startMkDocsContainer ~104 lines, no unit tests, hardcoded constants
- **Severity:** medium

#### TD-039 -- Global mutable containerProvider and no subcommand tests

- **Package:** drawio
- **Source:** `go/cli/eac/impl/drawio/README.md`
- **Detail:** global mutable containerProvider, no subcommand tests
- **Severity:** medium

#### TD-040 -- Low test coverage and inconsistent helper usage

- **Package:** get
- **Source:** `go/cli/eac/impl/get/README.md`
- **Detail:** only 8 tests for ~38 files, 11 subcommands skip shared helper
- **Severity:** medium

#### TD-041 -- Extremely oversized init.go with too many responsibilities

- **Package:** init
- **Source:** `go/cli/eac/impl/init/README.md`
- **Detail:** init.go 757 lines, Init() ~233 lines, generateWithScan ~84 lines
- **Severity:** medium

#### TD-042 -- Oversized artifact_helpers.go

- **Package:** internal
- **Source:** `go/cli/eac/impl/internal/README.md`
- **Detail:** artifact_helpers.go 698 lines
- **Severity:** low

#### TD-043 -- Oversized aggregation.go

- **Package:** testview
- **Source:** `go/cli/eac/impl/internal/manifests/testview/README.md`
- **Detail:** aggregation.go 400 lines
- **Severity:** low

#### TD-044 -- Oversized lintUnitWorker and framework.go with 1 test

- **Package:** lint
- **Source:** `go/cli/eac/impl/lint/README.md`
- **Detail:** lintUnitWorker ~201 lines, framework.go 556 lines, only 1 test
- **Severity:** medium

#### TD-045 -- Oversized download_evidence_artifacts.go

- **Package:** pipeline
- **Source:** `go/cli/eac/impl/pipeline/README.md`
- **Detail:** download_evidence_artifacts.go 399 lines
- **Severity:** low

#### TD-046 -- Oversized schedule.go with custom error types

- **Package:** ci
- **Source:** `go/cli/eac/impl/pipeline/ci/README.md`
- **Detail:** schedule.go 572 lines, custom error types
- **Severity:** medium

#### TD-047 -- Oversized runner.go and github.go

- **Package:** helper
- **Source:** `go/cli/eac/impl/pipeline/helper/README.md`
- **Detail:** runner.go 363 lines, github.go 283 lines
- **Severity:** low

#### TD-048 -- Four oversized functions and no tests for 4 files

- **Package:** release
- **Source:** `go/cli/eac/impl/release/README.md`
- **Detail:** 4 oversized functions (213-292 lines), no tests for 4 files
- **Severity:** medium

#### TD-049 -- Oversized unit_worker.go and scanUnitWorker

- **Package:** scan
- **Source:** `go/cli/eac/impl/scan/README.md`
- **Detail:** unit_worker.go 537 lines, scanUnitWorker ~129 lines
- **Severity:** low

#### TD-050 -- Extremely oversized serve.go with global singleton

- **Package:** serve
- **Source:** `go/cli/eac/impl/serve/README.md`
- **Detail:** serve.go 683 lines, GlobalServeContext singleton, no tests for 3 files
- **Severity:** medium

#### TD-051 -- Single oversized serve.go file

- **Package:** gource
- **Source:** `go/cli/eac/impl/serve/gource/README.md`
- **Detail:** serve.go 455 lines single file
- **Severity:** low

#### TD-052 -- GlobalServeContext mutable global

- **Package:** servers
- **Source:** `go/cli/eac/impl/serve/servers/README.md`
- **Detail:** GlobalServeContext mutable global
- **Severity:** medium

#### TD-053 -- Oversized test-summary.go with low test coverage

- **Package:** show
- **Source:** `go/cli/eac/impl/show/README.md`
- **Detail:** test-summary.go 655 lines, only 7 tests for ~35 files
- **Severity:** medium

#### TD-054 -- Near-identical install sub-packages with no tests

- **Package:** templates
- **Source:** `go/cli/eac/impl/templates/README.md`
- **Detail:** near-identical install sub-packages, no tests for 4 of 5
- **Severity:** medium

#### TD-055 -- Oversized test.go and testAfterResolve function

- **Package:** test
- **Source:** `go/cli/eac/impl/test/README.md`
- **Detail:** testAfterResolve ~231 lines, test.go 673 lines
- **Severity:** medium

#### TD-056 -- Extremely oversized clear.go with hardcoded semaphore list

- **Package:** update/cache-clear
- **Source:** `go/cli/eac/impl/update/cache-clear/README.md`
- **Detail:** clear.go 497 lines, hardcoded semaphore list
- **Severity:** medium

#### TD-057 -- Global mock state requires explicit reset

- **Package:** update/design
- **Source:** `go/cli/eac/impl/update/design/README.md`
- **Detail:** global mock state
- **Severity:** low

#### TD-058 -- No tests for 4 files

- **Package:** update/docs
- **Source:** `go/cli/eac/impl/update/docs/README.md`
- **Detail:** no tests for 4 files
- **Severity:** medium

#### TD-059 -- Oversized runUpdate with no tests for 3 files

- **Package:** docs-manifest
- **Source:** `go/cli/eac/impl/update/docs-manifest/README.md`
- **Detail:** runUpdate ~191 lines, no tests for 3 files
- **Severity:** medium

#### TD-060 -- Oversized update.go

- **Package:** pdf-screenshots
- **Source:** `go/cli/eac/impl/update/pdf-screenshots/README.md`
- **Detail:** update.go 490 lines
- **Severity:** low

#### TD-061 -- Mutable package-level vars and no tests for 7 files

- **Package:** validate
- **Source:** `go/cli/eac/impl/validate/README.md`
- **Detail:** mutable package-level vars, no tests for 7 files
- **Severity:** medium

#### TD-062 -- Oversized Remove and CreatePR with no test

- **Package:** work
- **Source:** `go/cli/eac/impl/work/README.md`
- **Detail:** Remove() ~247 lines, CreatePR() ~184 lines, no test for work.go
- **Severity:** medium

#### TD-063 -- Oversized git_ops.go

- **Package:** work/internal
- **Source:** `go/cli/eac/impl/work/internal/README.md`
- **Detail:** git_ops.go 382 lines
- **Severity:** low

#### TD-064 -- Package-level AI response state and duplicate extraction

- **Package:** risk
- **Source:** `go/cli/eac/internal/risk/README.md`
- **Detail:** Deps.AIResponse package-level state, duplicate control-ID extraction
- **Severity:** medium

#### TD-065 -- Many types packed in one file

- **Package:** evidence
- **Source:** `go/cli/eac/internal/risk/evidence/README.md`
- **Detail:** many types in one file
- **Severity:** low

#### ~~TD-066~~ -- ~~Duplicate GetProfileControlIDs functions~~ (RESOLVED)

- **Package:** oscal
- **Source:** `go/cli/eac/internal/risk/oscal/README.md`
- **Detail:** ~~duplicate GetProfileControlIDs/GetControlIDsFromProfile~~ Removed duplicate `GetControlIDsFromProfile`, consolidated all callers to `GetProfileControlIDs`.

### clibase

#### TD-067 -- Package-level fixed regex array

- **Package:** ansi
- **Source:** `go/clibase/ansi/README.md`
- **Detail:** package-level fixed regex array (minor)
- **Severity:** low

#### TD-068 -- Package-level log vars and untested incremental.go

- **Package:** caching
- **Source:** `go/clibase/caching/README.md`
- **Detail:** package-level log vars, incremental.go no tests
- **Severity:** medium

#### TD-069 -- Package-level log var and silent manifest errors

- **Package:** itemcache
- **Source:** `go/clibase/caching/itemcache/README.md`
- **Detail:** package-level log var, silent manifest load errors
- **Severity:** medium

#### TD-070 -- Package-level log var and no signal shutdown

- **Package:** capacity
- **Source:** `go/clibase/capacity/README.md`
- **Detail:** package-level log var, signal handler no shutdown
- **Severity:** medium

#### TD-071 -- Oversized Run and summary functions with package-level log

- **Package:** cmdframework
- **Source:** `go/clibase/cmdframework/README.md`
- **Detail:** Run() ~218 lines, summary functions 258+144 lines, package-level log
- **Severity:** medium

#### TD-072 -- Wide Console interface and global bootstrap with no tests

- **Package:** display
- **Source:** `go/clibase/display/README.md`
- **Detail:** Console interface ~22 methods, package-level globalBootstrap, no tests
- **Severity:** medium

#### TD-073 -- Mutable GlobalFlags slice and fragile runtime.Caller

- **Package:** flags
- **Source:** `go/clibase/flags/README.md`
- **Detail:** registry.go:15 mutable package-level GlobalFlags slice; safer as a function returning a copy. registry.go:39 ValidateFlagsFromRegistry uses runtime.Caller() to auto-detect calling command; this is fragile across refactors
- **Severity:** medium

#### TD-074 -- Oversized FormatDetailed and FormatCompact functions

- **Package:** initsummary
- **Source:** `go/clibase/initsummary/README.md`
- **Detail:** formatter.go:139 FormatDetailed is ~240 lines; extracting per-section helpers would improve readability. formatter.go:10 FormatCompact is ~128 lines with similar section-by-section logic
- **Severity:** low

#### TD-075 -- Oversized AcquireWithWait and repetitive config functions

- **Package:** locking
- **Source:** `go/clibase/locking/README.md`
- **Detail:** locking.go:142 AcquireWithWait is ~114 lines with deeply nested select/ticker logic. Nine near-identical \*Config() convenience functions could be collapsed
- **Severity:** low

#### TD-076 -- Oversized RunUnits and mutable defaultDetector

- **Package:** orchestrator
- **Source:** `go/clibase/orchestrator/README.md`
- **Detail:** unit_scheduler_core.go:261 RunUnits is ~180 lines. orchestrator_core.go:314 processWorkItem is ~108 lines. unit_scheduler_capacity.go:61 mutable package-level defaultDetector var
- **Severity:** medium

#### ~~TD-077~~ -- ~~Mutable package-level registry map without mutex~~ (RESOLVED)

- **Package:** render
- **Source:** `go/clibase/render/README.md`
- **Detail:** ~~custom/registry.go:18 mutable package-level registry map; concurrent command registration would race without external synchronization.~~ Added `sync.RWMutex`. Limited test coverage for console_table.go, json.go, and toml.go remains.

#### ~~TD-078~~ -- ~~Mutable registry map without concurrency protection~~ (RESOLVED)

- **Package:** custom
- **Source:** `go/clibase/render/custom/README.md`
- **Detail:** ~~registry.go:18 mutable package-level registry map with no mutex~~ Added `sync.RWMutex` to guard Register, Get, and List.

#### TD-079 -- Oversized services.go with stub methods returning zero values

- **Package:** services
- **Source:** `go/clibase/services/README.md`
- **Detail:** services.go (571 lines) contains ~15 private adapter structs in a single file. Several adapter methods return hardcoded zero values. GetComponentAmp always returns 1.0
- **Severity:** medium

#### TD-080 -- Mutable package-level registry vars

- **Package:** testrunners
- **Source:** `go/clibase/testrunners/README.md`
- **Detail:** registry.go:29-32 four mutable package-level vars (runners, descriptors, fallback, mu); protected by mutex but still global state
- **Severity:** low

#### TD-081 -- Mutable UpdateGolden package-level flag

- **Package:** testutil
- **Source:** `go/clibase/testutil/README.md`
- **Detail:** golden.go:14 package-level mutable UpdateGolden flag; safe in practice since flag.Parse runs once, but not idiomatic for library code
- **Severity:** low

#### ~~TD-082~~ -- ~~Contains duplicates stdlib slices.Contains~~ (RESOLVED)

- **Package:** utils
- **Source:** `go/clibase/utils/README.md`
- **Detail:** ~~Contains duplicates functionality available in Go 1.21+ slices.Contains~~ Replaced with `slices.Contains` and deleted the entire `utils` package.

### core

#### TD-083 -- No behavioral tests for adapters

- **Package:** adapters
- **Source:** `go/core/adapters/README.md`
- **Detail:** No test files exist; adapter correctness relies solely on compile-time interface checks. unit_adapter.go:90-96 PoolAllocationAdapter declares an anonymous interface instead of importing the concrete type
- **Severity:** medium

#### TD-084 -- Mutable BuildPromptWithTemplate and no unit tests

- **Package:** ai
- **Source:** `go/core/ai/README.md`
- **Detail:** ai.go:74: package-level var BuildPromptWithTemplate is mutable; could be replaced with a function call. No unit tests for generation/structured_generator.go (276 lines) or generation/strategies.go (318 lines)
- **Severity:** medium

#### TD-085 -- ShouldForcePull and ShouldForceNoCacheDocker obscure ShouldSkip

- **Package:** cache
- **Source:** `go/core/cache/README.md`
- **Detail:** config.go: ShouldForcePull and ShouldForceNoCacheDocker are thin wrappers that obscure the underlying ShouldSkip call; document or inline them to reduce indirection
- **Severity:** low

#### TD-086 -- Oversized DetectChanges and single-file structure

- **Package:** changedetect
- **Source:** `go/core/changedetect/README.md`
- **Detail:** changedetect.go:152: DetectChanges (~182 lines) combines git-state fast path, parallel hashing, dependency propagation, and result assembly. changedetect.go (421 lines): single file holds all core types, the detector, and state computation
- **Severity:** medium

#### TD-087 -- Oversized function with many parameters and complex regex

- **Package:** changelog
- **Source:** `go/core/changelog/README.md`
- **Detail:** semver.go:163: CalculateNextVersionConstrained takes 7 parameters. conventional.go:38: conventionalCommitRegex is a complex multi-group regex lacking doc comments
- **Severity:** low

#### TD-088 -- Multiple global mutable singletons and oversized functions

- **Package:** config
- **Source:** `go/core/config/README.md`
- **Detail:** Multiple global mutable singletons: global.go (globalConfig), cache.go (globalConfigCache), git_provider.go (gitRemoteProvider) each with separate locking. ModuleComponents.Clone() is a 50+ line manual deep-copy. LoadRepository is a 90-line orchestration method
- **Severity:** medium

#### TD-089 -- No test file for critical CTRF aggregation paths

- **Package:** ctrf
- **Source:** `go/core/ctrf/README.md`
- **Detail:** ctrf.go (167 lines) has no test file; Parse, Merge, and AddTest are critical aggregation paths that should have unit tests
- **Severity:** medium

#### TD-090 -- Oversized ResolveDefaults and implicit path coupling

- **Package:** defaults
- **Source:** `go/core/defaults/README.md`
- **Detail:** type_defaults.go: ResolveDefaults accepts 12 positional parameters. defaults.go: Path-building functions use inline string concatenation rather than referencing paths.SpecsDir
- **Severity:** medium

#### ~~TD-091~~ -- ~~Package-level compiled regex in docsync~~ (RESOLVED)

- **Package:** docsync
- **Source:** `go/core/docsync/README.md`
- **Detail:** ~~commands.go:144: package-level compiled regex cmdMarkerPattern; harmless but couples pattern to package init~~ Already has doc comment; this is idiomatic Go for package-level compiled regex.

#### TD-092 -- Duplicate component types and duplicate config path

- **Package:** domain
- **Source:** `go/core/domain/README.md`
- **Detail:** ModuleComponents and ComponentEntry in components.go are near-duplicates of types in config/modules.go. EACConfigRelPath constant in loader.go:16 is explicitly flagged as a duplication of paths.EACConfigRelPath. globalRegistry in registry.go:14 is package-level mutable state
- **Severity:** medium

#### TD-093 -- Oversized constants block and unsafe platform code

- **Package:** environments
- **Source:** `go/core/environments/README.md`
- **Detail:** constants.go: single const block with 100+ constants spanning test mocks, CI metadata, debug flags, and app config. memory_windows.go:29: uses unsafe.Sizeof with kernel32.GlobalMemoryStatusEx
- **Severity:** low

#### TD-094 -- Oversized 28-method interface and mutable time var

- **Package:** git
- **Source:** `go/core/git/README.md`
- **Detail:** interface.go: GitRepository has 28 methods -- consider splitting into focused interfaces. time.go:7: package-level mutable var timeNow used for test overrides
- **Severity:** medium

#### TD-095 -- Undocumented regex patterns in github package

- **Package:** github
- **Source:** `go/core/github/README.md`
- **Detail:** package_safety.go:213,220: compiled regexps bundleTagPattern and moduleVersionPattern lack doc comments explaining expected formats
- **Severity:** low

#### TD-096 -- ~~Duplicate worker-count logic~~ and swallowed errors (PARTIALLY RESOLVED)

- **Package:** hash
- **Source:** `go/core/hash/README.md`
- **Detail:** ~~parallel.go: DefaultParallelOptions() and normalizeWorkers() duplicate floor/cap worker-count logic.~~ Resolved: extracted `clampWorkers` helper. hash.go: UncommittedState reads files sequentially and silently swallows errors as deleted (remaining)
- **Severity:** medium

#### TD-097 -- Seven mutable package-level vars and duplicate construction

- **Package:** logging
- **Source:** `go/core/logging/README.md`
- **Detail:** debug.go: Seven package-level mutable vars form a large implicit global surface. configure.go and logger.go both construct zap.ErrorOutput(zapcore.AddSync(io.Discard)) independently
- **Severity:** medium

#### TD-098 -- Duplicate AST walk boilerplate in markdown

- **Package:** markdown
- **Source:** `go/core/markdown/README.md`
- **Detail:** extractCodeBlocks and extractSections share near-identical AST walk boilerplate; a generic walk-and-collect helper would reduce duplication
- **Severity:** low

#### TD-099 -- Hardcoded platform and architecture lists

- **Package:** module-deps
- **Source:** `go/core/module-deps/README.md`
- **Detail:** verify.go:229-230: hardcoded platform/arch lists ("linux", "windows", "darwin", "amd64", "arm64") should come from config
- **Severity:** medium

#### TD-100 -- Oversized ports.go dominated by trivial getters

- **Package:** output
- **Source:** `go/core/output/README.md`
- **Detail:** ports.go (319 lines): dominated by trivial getter methods to satisfy port interfaces. reader.go (360 lines): DiskOutputReader handles reading, validation, integrity checks, and build-ID extraction
- **Severity:** low

#### ~~TD-101~~ -- ~~Unused hasMainWorkspace~~ and repeated path joins (PARTIALLY RESOLVED)

- **Package:** paths
- **Source:** `go/core/paths/README.md`
- **Detail:** ~~paths_config.go: hasMainWorkspace is tracked but immediately discarded.~~ Removed dead variable. Many path builder functions in paths_output.go and paths_cache.go repeat filepath.Join(repoRoot, OutDir, ...) inline
- **Severity:** low

#### TD-102 -- No tests and identical pass-through implementations

- **Package:** platform
- **Source:** `go/core/platform/README.md`
- **Detail:** No test files exist; both WrapCommand and LineEnding are untested. WrapCommand has identical pass-through implementations on both platforms
- **Severity:** low

#### ~~TD-103~~ -- ~~Regexes recompiled on every ParseContent call~~ (RESOLVED)

- **Package:** releasenotes
- **Source:** `go/core/releasenotes/README.md`
- **Detail:** ~~Already hoisted to package-level vars (lines 18-20). Local aliases at lines 56-57 just reference the package vars.~~

#### TD-104 -- Repetitive FileCache filter methods

- **Package:** repository
- **Source:** `go/core/repository/README.md`
- **Detail:** file_cache.go: 6 filter methods with near-identical iteration patterns -- could be generalized with a predicate-based Filter method
- **Severity:** low

#### TD-105 -- Four mutable test injection globals in reports

- **Package:** reports
- **Source:** `go/core/domain/reports/README.md`
- **Detail:** Four separate package-level mutable globals for test injection: ghCLI and ghCLIExecutor in approval_comments.go, gitRepoProvider in specs.go, versionResolverRepo in version_resolver.go. resolvePhases hard-codes testable component types. globalComponentCache has no invalidation
- **Severity:** medium

#### TD-106 -- Hand-maintained schema types list and confusing constant name

- **Package:** schema
- **Source:** `go/core/domain/schema/README.md`
- **Detail:** GetSchemaTypes() returns a hand-maintained list that can fall out of sync with schemaFileNames map. SchemaEACConfig constant is named "ai-provider" which is confusing
- **Severity:** low

#### TD-107 -- Oversized ResolveForBuild and dense resolution file

- **Package:** resolver
- **Source:** `go/core/resolver/README.md`
- **Detail:** component_resolver.go:41: ResolveForBuild (~179 lines) handles dependency graph construction, tool-chain expansion, pool allocation, and weight calculation. component_resolver.go (500 lines): high density
- **Severity:** medium

#### TD-108 -- WorkScheduler interface on edge of bloat

- **Package:** scheduling
- **Source:** `go/core/scheduling/README.md`
- **Detail:** scheduler.go: WorkScheduler interface has 8 methods -- on the edge of interface bloat. dependency_tracker.go:92: lazy reverse-map built on first cascade call creates a subtle performance cliff
- **Severity:** low

#### TD-109 -- Oversized steps_cache_test.go and global mutable test state

- **Package:** specs
- **Source:** `go/core/specs/README.md`
- **Detail:** steps_cache_test.go is 1104 lines. Multiple package-level mutable vars for test state: toolState, logCtx, cfgState, cacheCtx
- **Severity:** low

#### TD-110 -- TODO for tag expression parsing and oversized discovery

- **Package:** testing
- **Source:** `go/core/testing/README.md`
- **Detail:** discovery.go:156: TODO comment -- "Handle complex expressions if needed". discovery.go:297: discoverModuleAllTests() is ~110 lines. discovery.go:20: package-level var log global mutable logger
- **Severity:** medium

#### TD-111 -- Oversized executeContainer and multiple mutable singletons

- **Package:** tool
- **Source:** `go/core/tool/README.md`
- **Detail:** executor.go:231: executeContainer (~200 lines) handles image resolution, mount setup, env forwarding, DinD, memory limits, and execution in a single function. global.go: multiple mutable singletons. categories.go: mutable package-level maps
- **Severity:** medium

#### TD-112 -- Oversized Gherkin validator and error code file

- **Package:** validation
- **Source:** `go/core/validation/README.md`
- **Detail:** formats/gherkin/validator.go: Validate() is ~200 lines. error_codes.go: 1000-line file. formats/structurizr/docker.go:29: global mutable defaultContainerProvider
- **Severity:** medium

#### TD-113 -- Oversized DetectTestModuleChanges and mutable default rules

- **Package:** workunit
- **Source:** `go/core/workunit/README.md`
- **Detail:** state_manager.go (557 lines): DetectTestModuleChanges (~195 lines) is the largest function. unit_state.go:33,43: mutable package-level var DefaultRules and IntegrationTestRule maps
- **Severity:** medium

#### TD-114 -- Boilerplate field-by-field mapping in loader

- **Package:** modules
- **Source:** `go/core/domain/modules/README.md`
- **Detail:** loadModules in loader.go:24-173 manually maps config.Module fields to domain.BaseContract field-by-field (~100 lines of boilerplate). MatchesFile hard-codes component names as catch-all cases
- **Severity:** medium

#### TD-115 -- Stub JSON validator not yet migrated

- **Package:** validation/formats/json
- **Source:** `go/core/validation/formats/json/README.md`
- **Detail:** This is a stub. The comment on line 22 says "Implementation will be moved from domain.JSONSchemaValidator" -- that migration has not happened yet
- **Severity:** low

#### TD-116 -- Single 700-line Gherkin validator file

- **Package:** validation/formats/gherkin
- **Source:** `go/core/validation/formats/gherkin/README.md`
- **Detail:** The validator is a single ~700-line file. Could benefit from splitting into per-category validator functions in separate files
- **Severity:** low

### mcp

#### TD-117 -- All production code in single oversized main.go

- **Package:** eac-mcp-server
- **Source:** `go/mcp/commands/README.md`
- **Detail:** All production code in a single main.go (322 lines) with 25+ functions; type definitions, request handling, command discovery, and execution are intermixed. main.go:135 protocol version 2024-11-05 is hardcoded inline
- **Severity:** medium

### specs

#### TD-118 -- Unimplemented TODOs for tag conflict and link checking

- **Package:** repository
- **Source:** `go/specs/repository/README.md`
- **Detail:** TODO at steps_test.go:488 for proper Gherkin parsing and tag conflict detection remains unimplemented. TODO at steps_test.go:550 for link checking remains unimplemented. steps_test.go (~593 lines) is the largest step file
- **Severity:** low

#### TD-119 -- Global mutable context variables in specs tests

- **Package:** repository
- **Source:** `go/specs/repository/README.md`
- **Detail:** Global mutable context variables (cmdDocsCtx in steps_command_docs_test.go:32, modIsoCtx in steps_module_isolation_test.go:30) create implicit shared state. Global mutable discoveryCache in steps_test_sanity_test.go:28
- **Severity:** low

---

## Pain Points

### adapters

#### PP-001 -- Config loaded from disk on every Execute call

- **Package:** ai
- **Source:** `go/adapters/ai/README.md`
- **Detail:** Executor.loadConfig() reads from disk on every Execute() call; caching the config with file-mtime invalidation would reduce I/O for repeated invocations
- **Impact:** medium

#### PP-002 -- Gemini creates new client on every call

- **Package:** providers
- **Source:** `go/adapters/ai/providers/README.md`
- **Detail:** Gemini.Execute creates new genai.Client on every call
- **Impact:** medium

#### PP-003 -- Silent error skipping in file loading

- **Package:** toolhandler
- **Source:** `go/adapters/ai/toolhandler/README.md`
- **Detail:** loadFilesWithExtensions silently skips errors
- **Impact:** medium

#### PP-004 -- Approximate test counts and no unit tests

- **Package:** behave
- **Source:** `go/adapters/behave/README.md`
- **Detail:** approximate test counts, no unit tests
- **Impact:** medium

#### PP-005 -- Approximate test counts and no unit tests

- **Package:** cucumber
- **Source:** `go/adapters/cucumber/README.md`
- **Detail:** approximate counts, no unit tests
- **Impact:** medium

#### PP-006 -- No tests for scan/browser and stdout writes

- **Package:** docker
- **Source:** `go/adapters/docker/README.md`
- **Detail:** no tests for scan/browser, ensureImage writes to os.Stdout
- **Impact:** medium

#### PP-007 -- Redundant IsDinD check in docker util

- **Package:** util
- **Source:** `go/adapters/docker/util/README.md`
- **Detail:** redundant IsDinD check
- **Impact:** low

#### PP-008 -- Duplicate Execute methods in eac adapter

- **Package:** eac
- **Source:** `go/adapters/eac/README.md`
- **Detail:** duplicate Execute methods
- **Impact:** low

#### PP-009 -- Direct stderr writes and package-level var log

- **Package:** godog
- **Source:** `go/adapters/godog/README.md`
- **Detail:** logBinaryNotFoundDiagnostics writes to os.Stderr, package-level var log
- **Impact:** medium

#### PP-010 -- Manual tag parsing and no unit tests

- **Package:** gotest
- **Source:** `go/adapters/gotest/README.md`
- **Detail:** manual tag parsing, no unit tests
- **Impact:** medium

#### PP-011 -- Approximate counts and no unit tests

- **Package:** mocha
- **Source:** `go/adapters/mocha/README.md`
- **Detail:** approximate counts, no unit tests
- **Impact:** medium

#### PP-012 -- Unreliable mtime comparison and swallowed walk errors

- **Package:** npm
- **Source:** `go/adapters/npm/README.md`
- **Detail:** mtime/size comparison unreliable, swallowed walk errors
- **Impact:** medium

#### PP-013 -- Unreliable mtime comparison and swallowed walk errors

- **Package:** pip
- **Source:** `go/adapters/pip/README.md`
- **Detail:** mtime comparison unreliable, swallowed walk errors
- **Impact:** medium

#### PP-014 -- Approximate counts and no unit tests for Execute

- **Package:** pytest
- **Source:** `go/adapters/pytest/README.md`
- **Detail:** approximate counts, no unit tests for Execute
- **Impact:** medium

#### PP-015 -- Oversized NewModel with 8 positional params

- **Package:** console
- **Source:** `go/adapters/tui/console/README.md`
- **Detail:** NewModel 8 positional params, package-level var log
- **Impact:** medium

#### PP-016 -- Model.go mixes concerns at 340 lines

- **Package:** demo
- **Source:** `go/adapters/tui/demo/README.md`
- **Detail:** model.go ~340 lines mixes concerns
- **Impact:** low

#### PP-017 -- Global registry mutable state and duplicate observer mapping

- **Package:** tui
- **Source:** `go/adapters/tui/README.md`
- **Detail:** globalRegistry mutable state, observer duplicates mapping
- **Impact:** medium

#### PP-018 -- Package-level regexps prevent customization

- **Package:** stream
- **Source:** `go/adapters/tui/stream/README.md`
- **Detail:** package-level compiled regexps prevent customization
- **Impact:** low

### cli/clie

#### PP-019 -- Oversized addExtensionToConfig and signal handler races

- **Package:** cmd
- **Source:** `go/cli/clie/cmd/README.md`
- **Detail:** addExtensionToConfig 227 lines, signal handler races
- **Impact:** high

#### PP-020 -- EBNF loaded but unused in command parser

- **Package:** command-parser
- **Source:** `go/cli/clie/internal/command-parser/README.md`
- **Detail:** EBNF loaded but unused
- **Impact:** low

#### PP-021 -- ValidatePinnedExtensions mixes concerns

- **Package:** conf
- **Source:** `go/cli/clie/internal/conf/README.md`
- **Detail:** ValidatePinnedExtensions mixes concerns
- **Impact:** low

#### PP-022 -- Oversized EnsureImageExists and large file count

- **Package:** docker (clie)
- **Source:** `go/cli/clie/internal/docker/README.md`
- **Detail:** EnsureImageExists 162 lines, 13 non-test files
- **Impact:** medium

### cli/eac

#### PP-023 -- Help-flag logic duplicates dispatch

- **Package:** eac
- **Source:** `go/cli/eac/README.md`
- **Detail:** help-flag logic duplicates dispatch, test build tags
- **Impact:** low

#### PP-024 -- FormatDefault map iteration order nondeterministic

- **Package:** help
- **Source:** `go/cli/eac/help/README.md`
- **Detail:** FormatDefault map iteration order
- **Impact:** low

#### PP-025 -- Artifact derivation duplication in build

- **Package:** build
- **Source:** `go/cli/eac/impl/build/README.md`
- **Detail:** artifact derivation duplication
- **Impact:** medium

#### PP-026 -- No tests for 6 files and raw env-var CI detection

- **Package:** builders
- **Source:** `go/cli/eac/impl/build/builders/README.md`
- **Detail:** no tests for 6 files, raw env-var CI detection
- **Impact:** medium

#### PP-027 -- Duplicate ToolCommandExecutor and implicit phase ordering

- **Package:** docprep
- **Source:** `go/cli/eac/impl/build/docprep/README.md`
- **Detail:** duplicate ToolCommandExecutor, implicit phase ordering
- **Impact:** medium

#### PP-028 -- Overlapping regex patterns in content package

- **Package:** content
- **Source:** `go/cli/eac/impl/build/docprep/content/README.md`
- **Detail:** 3 overlapping regex patterns
- **Impact:** low

#### PP-029 -- Five regex patterns across navigation files

- **Package:** navigation
- **Source:** `go/cli/eac/impl/build/docprep/navigation/README.md`
- **Detail:** 5 regex patterns across files
- **Impact:** low

#### PP-030 -- Mixes orchestration/git/formatting in commit-message

- **Package:** commit-message
- **Source:** `go/cli/eac/impl/create/commit-message/README.md`
- **Detail:** mixes orchestration/git/formatting
- **Impact:** medium

#### PP-031 -- Monolithic feel and silent Docker skip

- **Package:** create/design
- **Source:** `go/cli/eac/impl/create/design/README.md`
- **Detail:** monolithic feel, silent Docker skip
- **Impact:** medium

#### PP-032 -- Untested oscal.go and evidence code

- **Package:** create/risk-assess
- **Source:** `go/cli/eac/impl/create/risk-assess/README.md`
- **Detail:** oscal.go 367 lines untested, evidence untested
- **Impact:** high

#### PP-033 -- Single file pipeline in squash-message

- **Package:** create/squash-message
- **Source:** `go/cli/eac/impl/create/squash-message/README.md`
- **Detail:** single file pipeline
- **Impact:** low

#### PP-034 -- No tests for export/formatter and duplicate Docker formatting

- **Package:** design
- **Source:** `go/cli/eac/impl/design/README.md`
- **Detail:** no tests for export/formatter, duplicate Docker volume formatting
- **Impact:** medium

#### PP-035 -- Untestable browser opening and thin pass-through

- **Package:** docs/helper
- **Source:** `go/cli/eac/impl/docs/helper/README.md`
- **Detail:** thin pass-through, untestable browser opening
- **Impact:** low

#### PP-036 -- Untestable without Docker and duplicate limitedBuffer

- **Package:** drawio
- **Source:** `go/cli/eac/impl/drawio/README.md`
- **Detail:** untestable without Docker, duplicate limitedBuffer
- **Impact:** medium

#### PP-037 -- Pattern duplication and oversized files in get

- **Package:** get
- **Source:** `go/cli/eac/impl/get/README.md`
- **Detail:** pattern duplication, oversized files
- **Impact:** medium

#### PP-038 -- Too many responsibilities and YAML generation functions

- **Package:** init
- **Source:** `go/cli/eac/impl/init/README.md`
- **Detail:** too many responsibilities, 6 YAML generation functions
- **Impact:** medium

#### PP-039 -- Fragile platform inference in internal

- **Package:** internal
- **Source:** `go/cli/eac/impl/internal/README.md`
- **Detail:** fragile platform inference
- **Impact:** medium

#### PP-040 -- Mixed concerns and generic utility helper in lint

- **Package:** lint
- **Source:** `go/cli/eac/impl/lint/README.md`
- **Detail:** mixed concerns, generic utility helper
- **Impact:** medium

#### PP-041 -- Package name does not match command

- **Package:** list
- **Source:** `go/cli/eac/impl/list/README.md`
- **Detail:** package name doesn't match command
- **Impact:** low

#### PP-042 -- No unit tests in pipeline

- **Package:** pipeline
- **Source:** `go/cli/eac/impl/pipeline/README.md`
- **Detail:** no unit tests
- **Impact:** medium

#### PP-043 -- Implicit state machine in CI scheduling

- **Package:** ci
- **Source:** `go/cli/eac/impl/pipeline/ci/README.md`
- **Detail:** implicit state machine
- **Impact:** medium

#### PP-044 -- Env-var mock control in pipeline helper

- **Package:** helper
- **Source:** `go/cli/eac/impl/pipeline/helper/README.md`
- **Detail:** env-var mock control
- **Impact:** low

#### PP-045 -- Per-file boilerplate duplication and overlapping CI polling

- **Package:** release
- **Source:** `go/cli/eac/impl/release/README.md`
- **Detail:** per-file boilerplate duplication, overlapping CI-polling logic
- **Impact:** medium

#### PP-046 -- Duplicate per-scanner image helpers and no direct unit tests

- **Package:** scan
- **Source:** `go/cli/eac/impl/scan/README.md`
- **Detail:** duplicate per-scanner image helpers, no direct unit tests
- **Impact:** medium

#### PP-047 -- Serve function with too many phases

- **Package:** serve
- **Source:** `go/cli/eac/impl/serve/README.md`
- **Detail:** Serve() too many phases, gource/serve.go 455 lines untested
- **Impact:** medium

#### PP-048 -- Regex conversion false negatives in unused specs

- **Package:** unused
- **Source:** `go/cli/eac/impl/specs/unused/README.md`
- **Detail:** regex conversion false negatives
- **Impact:** medium

#### PP-049 -- Boilerplate duplication and oversized files in show

- **Package:** show
- **Source:** `go/cli/eac/impl/show/README.md`
- **Detail:** boilerplate duplication, oversized files
- **Impact:** medium

#### PP-050 -- High boilerplate duplication and no test coverage

- **Package:** templates
- **Source:** `go/cli/eac/impl/templates/README.md`
- **Detail:** high boilerplate duplication, no coverage
- **Impact:** medium

#### PP-051 -- No test files for 6 source files

- **Package:** test
- **Source:** `go/cli/eac/impl/test/README.md`
- **Detail:** no test files for 6 source files
- **Impact:** medium

#### PP-052 -- Single file may grow in CTRF package

- **Package:** ctrf
- **Source:** `go/cli/eac/impl/test/internal/ctrf/README.md`
- **Detail:** single file may grow
- **Impact:** low

#### PP-053 -- Fragile Docker output parsing and manual size parsing

- **Package:** update/cache-clear
- **Source:** `go/cli/eac/impl/update/cache-clear/README.md`
- **Detail:** fragile Docker output parsing, manual size parsing
- **Impact:** medium

#### PP-054 -- Explicit reset needed for design update mock

- **Package:** update/design
- **Source:** `go/cli/eac/impl/update/design/README.md`
- **Detail:** explicit reset needed
- **Impact:** low

#### PP-055 -- Untestable shell-out and single function orchestration

- **Package:** update/docs
- **Source:** `go/cli/eac/impl/update/docs/README.md`
- **Detail:** untestable shell-out, orchestration in single function
- **Impact:** medium

#### PP-056 -- Mixed orchestration/output and direct file I/O

- **Package:** docs-manifest
- **Source:** `go/cli/eac/impl/update/docs-manifest/README.md`
- **Detail:** mixed orchestration/output, direct file I/O
- **Impact:** medium

#### PP-057 -- Duplicate output formatters in validate

- **Package:** validate
- **Source:** `go/cli/eac/impl/validate/README.md`
- **Detail:** 3 duplicate output formatters
- **Impact:** medium

#### PP-058 -- Repeated scaffolding and inconsistent error handling

- **Package:** work
- **Source:** `go/cli/eac/impl/work/README.md`
- **Detail:** repeated scaffolding, inconsistent error handling
- **Impact:** medium

#### PP-059 -- Repetitive timing boilerplate in work internal

- **Package:** work/internal
- **Source:** `go/cli/eac/impl/work/internal/README.md`
- **Detail:** repetitive timing boilerplate
- **Impact:** low

#### PP-060 -- Silent fallback to default likelihood in risk

- **Package:** risk
- **Source:** `go/cli/eac/internal/risk/README.md`
- **Detail:** silent fallback to default likelihood
- **Impact:** medium

#### PP-061 -- Import-cycle workaround types in risk evidence

- **Package:** evidence
- **Source:** `go/cli/eac/internal/risk/evidence/README.md`
- **Detail:** import-cycle workaround types
- **Impact:** low

### clibase

#### PP-062 -- Untested incremental.go in caching

- **Package:** caching
- **Source:** `go/clibase/caching/README.md`
- **Detail:** incremental.go untested
- **Impact:** medium

#### PP-063 -- Limited test coverage and no Windows tests

- **Package:** capacity
- **Source:** `go/clibase/capacity/README.md`
- **Detail:** limited test coverage, no Windows tests
- **Impact:** medium

#### PP-064 -- Over 1170 lines of summary code with minimal tests

- **Package:** cmdframework
- **Source:** `go/clibase/cmdframework/README.md`
- **Detail:** summary.go+builder ~1170 lines, minimal test coverage
- **Impact:** medium

#### PP-065 -- High coupling and mixed concerns in display

- **Package:** display
- **Source:** `go/clibase/display/README.md`
- **Detail:** high coupling, mixed concerns
- **Impact:** medium

#### PP-066 -- Large flag package hard to navigate

- **Package:** flags
- **Source:** `go/clibase/flags/README.md`
- **Detail:** 12 source files totaling ~2500 non-test lines; navigation is difficult. docs.go doc-generation logic (271 lines) is tightly coupled to the flag set types
- **Impact:** medium

#### PP-067 -- Compact and detailed formatters share logic independently

- **Package:** initsummary
- **Source:** `go/clibase/initsummary/README.md`
- **Detail:** Compact and detailed formatters share structural logic but are implemented independently; changes to summary fields require parallel updates in both
- **Impact:** medium

#### PP-068 -- Direct fmt.Printf bypassing structured logging

- **Package:** locking
- **Source:** `go/clibase/locking/README.md`
- **Detail:** locking.go:224 prints directly to fmt.Printf for "Waiting for lock" messages, bypassing structured logging and the display layer
- **Impact:** medium

#### PP-069 -- Dual-pool semaphore mixed with cascade-failure bookkeeping

- **Package:** orchestrator
- **Source:** `go/clibase/orchestrator/README.md`
- **Detail:** Dual-pool semaphore logic mixed with cascade-failure bookkeeping makes RunUnits hard to follow. The defensive BUG: warn suggests scheduler drain is not fully trusted
- **Impact:** high

#### PP-070 -- Format.go mixes display-name and result-line concerns

- **Package:** output
- **Source:** `go/clibase/output/README.md`
- **Detail:** format.go (342 lines) mixes display-name extraction, result-line formatting, and list formatting; these are logically separate concerns
- **Impact:** low

#### PP-071 -- Registry.go combines types and dispatch in 544 lines

- **Package:** registry
- **Source:** `go/clibase/registry/README.md`
- **Detail:** registry.go (544 lines) combines type definitions, registration logic, and dispatch in a single file; splitting would improve readability
- **Impact:** low

#### PP-072 -- YAML-first serialization path adds latency

- **Package:** render
- **Source:** `go/clibase/render/README.md`
- **Detail:** YAML-first serialization path (JSON/TOML marshal via YAML intermediate) adds latency and subtle ordering bugs if YAML tags differ from JSON tags
- **Impact:** medium

#### PP-073 -- Large adapter surface in services

- **Package:** services
- **Source:** `go/clibase/services/README.md`
- **Detail:** The large number of adapter types implementing port interfaces makes the file dense; new port methods require updates across many adapters
- **Impact:** medium

#### PP-074 -- No tests for StreamingRunner JSON parsing

- **Package:** testrunners
- **Source:** `go/clibase/testrunners/README.md`
- **Detail:** streaming.go has no dedicated test file; the JSON-streaming parser is a critical path that should have unit tests
- **Impact:** high

### core

#### PP-075 -- Wide ModuleContractPort interface and repeated adapt loops

- **Package:** adapters
- **Source:** `go/core/adapters/README.md`
- **Detail:** ModuleContractAdapter delegates 16 methods, making the underlying ModuleContractPort interface unusually wide. AdaptModules, AdaptUnitSpecs, and AdaptUnitResults repeat the same slice-map-adapt loop
- **Impact:** medium

#### PP-076 -- Retry.go mixes config, orchestration, and validator loading

- **Package:** ai
- **Source:** `go/core/ai/README.md`
- **Detail:** generation/retry.go (426 lines) mixes config building, retry orchestration, and validator loading. templates/builder.go has no direct tests
- **Impact:** medium

#### PP-077 -- Adapter code tested only through integration paths

- **Package:** changedetect
- **Source:** `go/core/changedetect/README.md`
- **Detail:** Only 1 test file for 3 source files; adapter bridge code in adapters.go is tested only through integration paths. Heavy use of goroutines makes debugging difficult
- **Impact:** medium

#### PP-078 -- Fragile regex-based parsing with mutable state

- **Package:** changelog
- **Source:** `go/core/changelog/README.md`
- **Detail:** parser.go uses regex-based line-by-line parsing with mutable state variables -- fragile for edge cases in markdown formatting. No shared test helpers between test files
- **Impact:** medium

#### PP-079 -- Difficult navigation across 56 config files

- **Package:** config
- **Source:** `go/core/config/README.md`
- **Detail:** Package spans ~56 files making navigation difficult. RepositoryConfig carries path-helper methods (30+) that could live in core/paths
- **Impact:** medium

#### PP-080 -- Single CTRF file with all types and operations

- **Package:** ctrf
- **Source:** `go/core/ctrf/README.md`
- **Detail:** All types, constructors, parsing, and merging logic live in a single file; splitting into types.go and operations.go would improve navigability
- **Impact:** low

#### PP-081 -- Mirrored TypeDefaults types require manual sync

- **Package:** defaults
- **Source:** `go/core/defaults/README.md`
- **Detail:** TypeDefaults in this package mirrors config.TypeDefaults to avoid import cycles; changes to one must be manually synchronized with the other
- **Impact:** medium

#### PP-082 -- Oversized BaseContract with 20+ methods

- **Package:** domain
- **Source:** `go/core/domain/README.md`
- **Detail:** BaseContract in types.go has grown to 20+ methods. shared_definitions.go validation maps are returned by reference -- callers could mutate them
- **Impact:** medium

#### PP-083 -- Foundation package triggers wide recompilation

- **Package:** environments
- **Source:** `go/core/environments/README.md`
- **Detail:** As a foundation package imported by nearly everything, adding a new constant triggers recompilation of a large portion of the module graph. Mock-related constants mixed with production constants
- **Impact:** medium

#### PP-084 -- Duplicate evidence write logic

- **Package:** evidence
- **Source:** `go/core/evidence/README.md`
- **Detail:** evidence.go: WriteEvidence and WriteComponentEvidence share ~80% of their logic; duplication could be reduced with a shared internal writer
- **Impact:** low

#### PP-085 -- Oversized git.go and expensive mock maintenance

- **Package:** git
- **Source:** `go/core/git/README.md`
- **Detail:** git.go (601 lines) covers status, diff, staging, and branch operations. mock.go (427 lines) must shadow all 28 interface methods
- **Impact:** medium

#### PP-086 -- Duplicate CLI parsing logic and wide API interface

- **Package:** github
- **Source:** `go/core/github/README.md`
- **Detail:** cli_mock.go (319 lines) duplicates CLI output parsing logic from gh_client.go. No clear boundary between workflow and release concerns within the API interface
- **Impact:** medium

#### PP-087 -- Oversized FilesParallel with multiple context checks

- **Package:** hash
- **Source:** `go/core/hash/README.md`
- **Detail:** parallel.go: FilesParallel is ~108 lines with three separate context-cancellation checks
- **Impact:** low

#### PP-088 -- Oversized ConfigureLogging and unprotected debug outputs

- **Package:** logging
- **Source:** `go/core/logging/README.md`
- **Detail:** configure.go: ConfigureLogging is ~142 lines with inline core construction. SetDebugOutput/SetStdOutput are not protected by a mutex
- **Impact:** medium

#### PP-089 -- Only JSON and YAML code blocks validated

- **Package:** markdown
- **Source:** `go/core/markdown/README.md`
- **Detail:** Only JSON and YAML code blocks are validated; other common languages silently pass. sanitizeMessage uses a loop instead of strings.Join(strings.Fields(...))
- **Impact:** low

#### PP-090 -- Oversized GetVersion mixing concerns

- **Package:** module-deps
- **Source:** `go/core/module-deps/README.md`
- **Detail:** GetVersion() (~70 lines) mixes version lookup with artifact resolution and platform iteration. loadModuleContract() creates hidden filesystem coupling
- **Impact:** medium

#### PP-091 -- Port adapter layer has high coupling surface

- **Package:** output
- **Source:** `go/core/output/README.md`
- **Detail:** Port adapter layer (ports.go) must be updated every time a field is added to UoWManifest, Artifact, or view types. Validation logic in reader.go mixes read and verify concerns
- **Impact:** medium

#### PP-092 -- No validation of repoRoot in path builders

- **Package:** paths
- **Source:** `go/core/paths/README.md`
- **Detail:** No validation that repoRoot is non-empty or absolute in any builder function
- **Impact:** medium

#### PP-093 -- String concatenation in ParseContent

- **Package:** releasenotes
- **Source:** `go/core/releasenotes/README.md`
- **Detail:** ParseContent() uses string concatenation for building content; strings.Builder would be more efficient for large files
- **Impact:** low

#### PP-094 -- Mutable testRepos var and widest package in core

- **Package:** repository
- **Source:** `go/core/repository/README.md`
- **Detail:** repository_test.go:16: mutable package-level var testRepos shared across tests. Sub-packages make this the widest package in core with 21 source files across 4 directories
- **Impact:** medium

#### PP-095 -- Duplicate branch-resolution logic in reports

- **Package:** reports
- **Source:** `go/core/domain/reports/README.md`
- **Detail:** GetApprovalComments and GetSpecs share nearly identical branch-resolution and commit-range logic. Each report function independently calls config.Load()
- **Impact:** medium

#### PP-096 -- Three resolve methods share overlapping logic

- **Package:** resolver
- **Source:** `go/core/resolver/README.md`
- **Detail:** Three resolve methods (ResolveForBuild, ResolveForLint, ResolveForScan) share overlapping logic for tool lookup and weight calculation. Weight calculation spread across three methods with subtly different heuristics
- **Impact:** medium

#### PP-097 -- Limited test coverage for dependency tracker

- **Package:** scheduling
- **Source:** `go/core/scheduling/README.md`
- **Detail:** Only 2 test files for 6 source files; dependency_tracker.go and heap.go lack dedicated unit tests
- **Impact:** medium

#### PP-098 -- Unsafe parallel test execution with global state

- **Package:** specs
- **Source:** `go/core/specs/README.md`
- **Detail:** Test state globals (cacheCtx, cfgState, etc.) make parallel test execution unsafe. loggingTestCounter is a global int64
- **Impact:** medium

#### PP-099 -- Oversized validation.go with deeply nested logic

- **Package:** testing
- **Source:** `go/core/testing/README.md`
- **Detail:** validation.go is 551 lines with deeply nested logic; could benefit from smaller validation step functions
- **Impact:** medium

#### PP-100 -- AssertCallCount uses rune conversion for numbers

- **Package:** testutil
- **Source:** `go/core/testutil/README.md`
- **Detail:** The AssertCallCount function uses rune conversion for number formatting which only works for single-digit counts
- **Impact:** low

#### PP-101 -- char/4 heuristic accuracy varies by language

- **Package:** tokensize
- **Source:** `go/core/tokensize/README.md`
- **Detail:** The char/4 heuristic is a rough approximation; accuracy varies by language and content type
- **Impact:** low

#### PP-102 -- 42 files make tool the largest core package

- **Package:** tool
- **Source:** `go/core/tool/README.md`
- **Detail:** 42 files make this the largest core package; navigating between bridges requires significant context switching. handler_adapter.go defines 4 separate handler interfaces with overlapping signatures
- **Impact:** medium

#### PP-103 -- No unit tests in root validation package

- **Package:** validation
- **Source:** `go/core/validation/README.md`
- **Detail:** No unit tests in the root validation/ package. Four separate interfaces may confuse consumers
- **Impact:** medium

#### PP-104 -- Process-level cache and panics in workspace

- **Package:** workspace
- **Source:** `go/core/workspace/README.md`
- **Detail:** workspace.go uses process-level cache and os.Getenv directly. MustDetect() and RootOrPanic() use panics unsafe in library contexts
- **Impact:** medium

#### PP-105 -- Mixed granularity levels in StateManager

- **Package:** workunit
- **Source:** `go/core/workunit/README.md`
- **Detail:** state_manager.go mixes UoW-level, module-level, and test-module-level change detection in one struct. unit_spec.go var aliases re-export contract constructors
- **Impact:** medium

#### ~~PP-106~~ -- ~~GetUsedBy rebuilds reverse dependency graph every call~~. matchWithFallback overlapping branches remain (PARTIALLY RESOLVED)

- **Package:** modules
- **Source:** `go/core/domain/modules/README.md`
- **Detail:** ~~GetUsedBy in registry.go rebuilds the entire reverse dependency graph on every call.~~ Resolved: `reverseDepGraph` pre-computed in `Add`; `GetUsedBy` is now O(1). matchWithFallback contains multiple overlapping fallback branches (remains)
- **Impact:** medium

#### PP-107 -- Oversized loadValidatorForFormat switch statement

- **Package:** generation
- **Source:** `go/core/ai/generation/README.md`
- **Detail:** The loadValidatorForFormat function in retry.go (line 373) has a large switch statement that grows with each new format
- **Impact:** low

#### PP-108 -- ContractLoader version parameter is ignored

- **Package:** ai/config
- **Source:** `go/core/ai/config/README.md`
- **Detail:** The ContractLoader version parameter is ignored (line 179 of loader.go), which is intentional but may confuse callers expecting version-aware behavior
- **Impact:** low

#### PP-109 -- thin tool only handles specs removal

- **Package:** defaults/cmd/thin
- **Source:** `go/core/defaults/cmd/thin/README.md`
- **Detail:** The tool only handles specs: [] removal. Other redundant fields are detected but not yet removed
- **Impact:** low

#### PP-110 -- Global DockerValidator containerProvider dependency

- **Package:** validation/formats/structurizr
- **Source:** `go/core/validation/formats/structurizr/README.md`
- **Detail:** DockerValidator depends on a global defaultContainerProvider variable which requires initialization via SetContainerProvider before use
- **Impact:** medium

### mcp

#### PP-111 -- Command discovery on every tools/list call

- **Package:** eac-mcp-server
- **Source:** `go/mcp/commands/README.md`
- **Detail:** getCommands() shells out to the EAC adapter on every tools/list call with no caching; repeated tool-list requests re-discover commands each time. Error logging goes to stderr without structured logging
- **Impact:** medium

### specs

#### PP-112 -- Global mutable context variables in specs tests

- **Package:** repository
- **Source:** `go/specs/repository/README.md`
- **Detail:** Global mutable context variables create implicit shared state across scenarios. Global mutable discoveryCache caches test discovery results as package-level state
- **Impact:** low

---

## Optimization Opportunities

### adapters

#### OO-001 -- Cache venv across test runs

- **Package:** behave
- **Source:** `go/adapters/behave/README.md`
- **Detail:** venv caching
- **Effort:** medium

#### OO-002 -- Cache npm ci across test runs

- **Package:** cucumber
- **Source:** `go/adapters/cucumber/README.md`
- **Detail:** npm ci caching
- **Effort:** medium

#### OO-003 -- Extract executeOnce helpers and background port cleanup

- **Package:** docker
- **Source:** `go/adapters/docker/README.md`
- **Detail:** extract executeOnce helpers, background port cleanup
- **Effort:** medium

#### OO-004 -- Pre-build extension-keyed index for godog

- **Package:** godog
- **Source:** `go/adapters/godog/README.md`
- **Detail:** pre-build extension-keyed index
- **Effort:** low

#### ~~OO-005 -- Skip go generate when no directives~~ RESOLVED

- **Package:** gotest
- **Source:** `go/adapters/gotest/README.md`
- **Detail:** ~~skip go generate when no directives~~ `hasGenerateDirectives` scans for `//go:generate` directives before invoking `go generate`; skips when none found
- **Effort:** low

#### OO-006 -- Cache npm packages for mocha

- **Package:** mocha
- **Source:** `go/adapters/mocha/README.md`
- **Detail:** npm caching
- **Effort:** medium

#### OO-007 -- Single-pass syncDirectory for npm

- **Package:** npm
- **Source:** `go/adapters/npm/README.md`
- **Detail:** single-pass syncDirectory
- **Effort:** low

#### OO-008 -- Single-pass syncDirectory for pip

- **Package:** pip
- **Source:** `go/adapters/pip/README.md`
- **Detail:** single-pass syncDirectory
- **Effort:** low

#### OO-009 -- Cache venv across pytest runs

- **Package:** pytest
- **Source:** `go/adapters/pytest/README.md`
- **Detail:** venv caching
- **Effort:** medium

#### OO-010 -- Cache paneHeights and extract mouse hit-testing

- **Package:** console
- **Source:** `go/adapters/tui/console/README.md`
- **Detail:** cache paneHeights, extract mouse hit-testing
- **Effort:** medium

#### OO-011 -- Extract message pump and isolate demo code

- **Package:** tui
- **Source:** `go/adapters/tui/README.md`
- **Detail:** extract message pump, isolate demo code
- **Effort:** medium

### cli/clie

#### OO-012 -- Dedicated GetLatestSHA function

- **Package:** cmd
- **Source:** `go/cli/clie/cmd/README.md`
- **Detail:** dedicated GetLatestSHA function
- **Effort:** low

#### OO-013 -- Dynamic EBNF parsing

- **Package:** command-parser
- **Source:** `go/cli/clie/internal/command-parser/README.md`
- **Detail:** dynamic EBNF parsing
- **Effort:** high

#### OO-014 -- Reflection-based merge for conf

- **Package:** conf
- **Source:** `go/cli/clie/internal/conf/README.md`
- **Detail:** reflection-based merge
- **Effort:** medium

#### ~~OO-015~~ -- ~~Replace custom string helpers with stdlib~~ (RESOLVED)

- **Package:** docker (clie)
- **Source:** `go/cli/clie/internal/docker/README.md`
- **Detail:** ~~replace custom string helpers~~ Replaced with `strings.Contains`, `slices.Contains`, `strings.Split`, `strings.LastIndex`.

#### OO-016 -- Refactor pullImage to accept DockerClient interface

- **Package:** tui (clie)
- **Source:** `go/cli/clie/internal/tui/README.md`
- **Detail:** refactor to accept DockerClient interface
- **Effort:** low

### cli/eac

#### OO-017 -- Split buildUnitWorker into phases

- **Package:** build
- **Source:** `go/cli/eac/impl/build/README.md`
- **Detail:** split buildUnitWorker
- **Effort:** medium

#### OO-018 -- Extract cross-compilation helper and unify CI detection

- **Package:** builders
- **Source:** `go/cli/eac/impl/build/builders/README.md`
- **Detail:** extract cross-compilation helper, unify CI detection
- **Effort:** medium

#### OO-019 -- Generic format helper for content

- **Package:** content
- **Source:** `go/cli/eac/impl/build/docprep/content/README.md`
- **Detail:** generic format helper
- **Effort:** low

#### OO-020 -- Split GenerateNavForDir into smaller functions

- **Package:** navigation
- **Source:** `go/cli/eac/impl/build/docprep/navigation/README.md`
- **Detail:** split GenerateNavForDir
- **Effort:** low

#### OO-021 -- Split commit-message into orchestration and git-helpers

- **Package:** commit-message
- **Source:** `go/cli/eac/impl/create/commit-message/README.md`
- **Detail:** split into orchestration and git-helpers
- **Effort:** medium

#### OO-022 -- Extract prompt loading and consolidate mocks

- **Package:** create/design
- **Source:** `go/cli/eac/impl/create/design/README.md`
- **Detail:** extract prompt loading, consolidate mocks
- **Effort:** medium

#### OO-023 -- Split assess.go and add unit tests

- **Package:** create/risk-assess
- **Source:** `go/cli/eac/impl/create/risk-assess/README.md`
- **Detail:** split assess.go, add unit tests
- **Effort:** medium

#### OO-024 -- Add focused unit tests for squash-message

- **Package:** create/squash-message
- **Source:** `go/cli/eac/impl/create/squash-message/README.md`
- **Detail:** add focused unit tests
- **Effort:** low

#### OO-025 -- Add unit tests and extract Docker helpers for design

- **Package:** design
- **Source:** `go/cli/eac/impl/design/README.md`
- **Detail:** add unit tests, extract Docker helpers
- **Effort:** medium

#### OO-026 -- Extract config generation and add unit tests for docs helper

- **Package:** docs/helper
- **Source:** `go/cli/eac/impl/docs/helper/README.md`
- **Detail:** extract config generation, add unit tests
- **Effort:** medium

#### OO-027 -- Deduplicate limitedBuffer and add mock tests

- **Package:** drawio
- **Source:** `go/cli/eac/impl/drawio/README.md`
- **Detail:** deduplicate limitedBuffer, add mock tests
- **Effort:** low

#### OO-028 -- Migrate to ExecuteGetCommand and add tests

- **Package:** get
- **Source:** `go/cli/eac/impl/get/README.md`
- **Detail:** migrate to ExecuteGetCommand, add tests
- **Effort:** medium

#### OO-029 -- Extract generators.go from init

- **Package:** init
- **Source:** `go/cli/eac/impl/init/README.md`
- **Detail:** extract generators.go
- **Effort:** medium

#### ~~OO-030~~ -- ~~Split artifact_helpers into focused files~~ (RESOLVED)

- **Package:** internal
- **Source:** `go/cli/eac/impl/internal/README.md`
- **Detail:** ~~split into focused files~~
- **Effort:** low

#### ~~OO-031~~ -- ~~Split aggregation builders in testview~~ (RESOLVED)

- **Package:** testview
- **Source:** `go/cli/eac/impl/internal/manifests/testview/README.md`
- **Detail:** ~~split aggregation builders~~
- **Effort:** low

#### OO-032 -- Split lintUnitWorker and add tests

- **Package:** lint
- **Source:** `go/cli/eac/impl/lint/README.md`
- **Detail:** split lintUnitWorker, add tests
- **Effort:** medium

#### OO-033 -- Extract artifact download helper for pipeline

- **Package:** pipeline
- **Source:** `go/cli/eac/impl/pipeline/README.md`
- **Detail:** extract artifact download helper
- **Effort:** low

#### OO-034 -- Separate command from scheduler and remove custom errors

- **Package:** ci
- **Source:** `go/cli/eac/impl/pipeline/ci/README.md`
- **Detail:** separate command from scheduler, remove custom errors
- **Effort:** medium

#### OO-035 -- Extract topologicalSort and constructor injection for mock

- **Package:** helper
- **Source:** `go/cli/eac/impl/pipeline/helper/README.md`
- **Detail:** extract topologicalSort, constructor injection for mock
- **Effort:** low

#### OO-036 -- Extract command scaffold and share CI-query utilities

- **Package:** release
- **Source:** `go/cli/eac/impl/release/README.md`
- **Detail:** extract command scaffold, share CI-query utilities
- **Effort:** medium

#### OO-037 -- Consolidate image helpers for scan

- **Package:** scan
- **Source:** `go/cli/eac/impl/scan/README.md`
- **Detail:** consolidate image helpers
- **Effort:** low

#### OO-038 -- Split serve.go and replace GlobalServeContext

- **Package:** serve
- **Source:** `go/cli/eac/impl/serve/README.md`
- **Detail:** split serve.go, replace GlobalServeContext
- **Effort:** medium

#### OO-039 -- Extract Docker lifecycle helper from gource

- **Package:** gource
- **Source:** `go/cli/eac/impl/serve/gource/README.md`
- **Detail:** extract Docker lifecycle helper
- **Effort:** low

#### OO-040 -- Replace GlobalServeContext with dependency injection

- **Package:** servers
- **Source:** `go/cli/eac/impl/serve/servers/README.md`
- **Detail:** replace with dependency injection
- **Effort:** medium

#### OO-041 -- Introduce ExecuteShowCommand helper and extract table-building

- **Package:** show
- **Source:** `go/cli/eac/impl/show/README.md`
- **Detail:** introduce ExecuteShowCommand helper, extract table-building
- **Effort:** medium

#### OO-042 -- Cache parsed feature steps for unused specs

- **Package:** unused
- **Source:** `go/cli/eac/impl/specs/unused/README.md`
- **Detail:** cache parsed feature steps
- **Effort:** low

#### OO-043 -- Extract generic install handler for templates

- **Package:** templates
- **Source:** `go/cli/eac/impl/templates/README.md`
- **Detail:** extract generic install handler
- **Effort:** medium

#### OO-044 -- Break testAfterResolve into sub-functions

- **Package:** test
- **Source:** `go/cli/eac/impl/test/README.md`
- **Detail:** break testAfterResolve into sub-functions
- **Effort:** low

#### OO-045 -- Split clear.go and make semaphore list discoverable

- **Package:** update/cache-clear
- **Source:** `go/cli/eac/impl/update/cache-clear/README.md`
- **Detail:** split clear.go, discoverable semaphore list
- **Effort:** medium

#### OO-046 -- Constructor-based DI for design update

- **Package:** update/design
- **Source:** `go/cli/eac/impl/update/design/README.md`
- **Detail:** constructor-based DI
- **Effort:** low

#### OO-047 -- Interface wrappers and split runMermaidUpdate

- **Package:** update/docs
- **Source:** `go/cli/eac/impl/update/docs/README.md`
- **Detail:** interface wrappers, split runMermaidUpdate
- **Effort:** medium

#### OO-048 -- Split runUpdate and add serialization tests

- **Package:** docs-manifest
- **Source:** `go/cli/eac/impl/update/docs-manifest/README.md`
- **Detail:** split runUpdate, add serialization tests
- **Effort:** low

#### OO-049 -- Extract scanning and validation from pdf-screenshots

- **Package:** pdf-screenshots
- **Source:** `go/cli/eac/impl/update/pdf-screenshots/README.md`
- **Detail:** extract scanning and validation
- **Effort:** low

#### OO-050 -- Extract shared ValidationOutput formatter

- **Package:** validate
- **Source:** `go/cli/eac/impl/validate/README.md`
- **Detail:** extract shared ValidationOutput formatter
- **Effort:** medium

#### OO-051 -- Extract shared command-lifecycle runner for work

- **Package:** work
- **Source:** `go/cli/eac/impl/work/README.md`
- **Detail:** extract shared command-lifecycle runner
- **Effort:** medium

#### OO-052 -- Extract timing wrapper for work internal

- **Package:** work/internal
- **Source:** `go/cli/eac/impl/work/internal/README.md`
- **Detail:** extract timing wrapper
- **Effort:** low

#### OO-053 -- Consolidate functions and use interface-based AI for risk

- **Package:** risk
- **Source:** `go/cli/eac/internal/risk/README.md`
- **Detail:** consolidate functions, interface-based AI
- **Effort:** medium

#### OO-054 -- Extract shared types to core for risk evidence

- **Package:** evidence
- **Source:** `go/cli/eac/internal/risk/evidence/README.md`
- **Detail:** extract shared types to core
- **Effort:** medium

#### ~~OO-055~~ -- ~~Consolidate duplicate oscal functions~~ (RESOLVED)

- **Package:** oscal
- **Source:** `go/cli/eac/internal/risk/oscal/README.md`
- **Detail:** ~~consolidate~~ Resolved via TD-066: removed duplicate `GetControlIDsFromProfile`, kept `GetProfileControlIDs`
- **Effort:** low

### clibase

#### ~~OO-056 -- Short-circuit with combinedBadAnsi check~~ RESOLVED

- **Package:** ansi
- **Source:** `go/clibase/ansi/README.md`
- **Detail:** ~~short-circuit with combinedBadAnsi check~~ `StripBad` and `FilterBadOnly` write path now short-circuit via `combinedBadAnsi.Match`
- **Effort:** low

#### OO-057 -- Inject logger via constructor for itemcache

- **Package:** itemcache
- **Source:** `go/clibase/caching/itemcache/README.md`
- **Detail:** inject logger via constructor
- **Effort:** low

#### OO-058 -- Extract shared moduleCache iteration for cmdframework

- **Package:** cmdframework
- **Source:** `go/clibase/cmdframework/README.md`
- **Detail:** extract shared moduleCache iteration, add resolve tests
- **Effort:** medium

#### OO-059 -- Consolidate factory functions and replace runtime.Caller

- **Package:** flags
- **Source:** `go/clibase/flags/README.md`
- **Detail:** Consolidate the four nearly identical \*Config() factory functions in commands.go into a data-driven table (low effort). Replace runtime.Caller() detection with explicit command-name parameter (medium effort)
- **Effort:** medium

#### OO-060 -- Extract shared section-building helpers for initsummary

- **Package:** initsummary
- **Source:** `go/clibase/initsummary/README.md`
- **Detail:** Extract shared section-building helpers used by both FormatCompact and FormatDetailed to reduce duplication (medium effort)
- **Effort:** medium

#### ~~OO-061~~ -- ~~Replace fmt.Printf and consolidate config functions~~ (RESOLVED)

- **Package:** locking
- **Source:** `go/clibase/locking/README.md`
- **Detail:** ~~Replace fmt.Printf with a writer parameter or log call. Consolidate \*Config() functions into a single NewConfig~~ Resolved: replaced fmt.Printf with io.Writer field, added NewConfig/NewUnitConfig/NewFileConfig factories with action table
- **Effort:** low

#### OO-062 -- Break RunUnits worker-pool loop into dispatchLoop

- **Package:** orchestrator
- **Source:** `go/clibase/orchestrator/README.md`
- **Detail:** Break RunUnits worker-pool loop into a dispatchLoop method to improve testability (low effort). Replace resultsMu mutex-guarded slice with indexed atomic writes (medium effort)
- **Effort:** medium

#### ~~OO-063~~ -- ~~Split format.go into focused files for output~~ (RESOLVED)

- **Package:** output
- **Source:** `go/clibase/output/README.md`
- **Detail:** ~~Split format.go into display_name.go and result_line.go for easier navigation (low effort)~~
- **Effort:** low

#### ~~OO-064~~ -- ~~Split registry.go into types, register, and dispatch~~ (RESOLVED)

- **Package:** registry
- **Source:** `go/clibase/registry/README.md`
- **Detail:** ~~registry.go was replaced by command_registry.go (137 lines) -- already well-organized, no split needed~~

#### ~~OO-065~~ -- ~~Add unit tests and guard global map for render~~ (PARTIALLY RESOLVED)

- **Package:** render
- **Source:** `go/clibase/render/README.md`
- **Detail:** Add unit tests for json.go and toml.go round-trip fidelity remains. ~~Guard custom/registry.go global map with sync.RWMutex~~ Done.

#### ~~OO-066~~ -- ~~Guard global registry map with mutex~~ (RESOLVED)

- **Package:** custom
- **Source:** `go/clibase/render/custom/README.md`
- **Detail:** ~~Guard the global registry map with sync.RWMutex~~ Done.
- **Effort:** low

#### OO-067 -- Extract adapter structs and audit stub methods

- **Package:** services
- **Source:** `go/clibase/services/README.md`
- **Detail:** Extract adapter structs into per-domain files. Audit stub methods returning zero values
- **Effort:** low

#### OO-068 -- Add unit tests for StreamingRunner JSON parsing

- **Package:** testrunners
- **Source:** `go/clibase/testrunners/README.md`
- **Detail:** Add unit tests for StreamingRunner JSON parsing edge cases (malformed lines, partial output)
- **Effort:** low

#### ~~OO-069~~ -- ~~Replace Contains with stdlib slices.Contains~~ (RESOLVED)

- **Package:** utils
- **Source:** `go/clibase/utils/README.md`
- **Detail:** ~~Replace Contains with slices.Contains from the Go standard library and remove this package~~ Done. Package deleted.

### core

#### OO-070 -- Add table-driven adapter tests

- **Package:** adapters
- **Source:** `go/core/adapters/README.md`
- **Detail:** Adding a table-driven adapter test for each type would be low-cost and catch method signature drift early
- **Effort:** low

#### OO-071 -- Add unit tests for generation and extract validator factory

- **Package:** ai
- **Source:** `go/core/ai/README.md`
- **Detail:** Add unit tests for StructuredGenerator and retry strategies directly (low effort, high value). Extract loadValidatorForFormat from retry.go into a standalone factory (medium effort)
- **Effort:** medium

#### OO-072 -- Extract parallel hashing helper for changedetect

- **Package:** changedetect
- **Source:** `go/core/changedetect/README.md`
- **Detail:** Extract the parallel hashing goroutine pool into a reusable helper to reduce duplication between DetectChanges and ComputeCurrentState (low effort, ~40 lines saved)
- **Effort:** low

#### ~~OO-073 -- Build moniker-to-index map for O(1) lookup~~ RESOLVED

- **Package:** config
- **Source:** `go/core/config/README.md`
- **Detail:** GetModule/GetByMoniker perform linear scans over []Module; a moniker-to-index map built once after load would be O(1). GetByExtension also does a linear scan
- **Effort:** low
- **Resolution:** Added `monikerIndex` map to `RepositoryConfig` (built once via `buildMonikerIndex()` after module loading) and `extensionToType`/`extensionToKind` maps to `ComponentKindsConfig` (built once via `buildExtensionIndex()` after component kinds loading). All lookup functions now use O(1) map access with linear-scan fallback during loading.

#### OO-074 -- Add Merge and Parse unit tests for CTRF

- **Package:** ctrf
- **Source:** `go/core/ctrf/README.md`
- **Detail:** Add unit tests covering Merge edge cases and round-trip Parse/Marshal tests
- **Effort:** low

#### OO-075 -- Extract branch-comparison sub-interface for git

- **Package:** git
- **Source:** `go/core/git/README.md`
- **Detail:** Extract branch-comparison methods into a dedicated sub-interface; most consumers only need read-only operations (low effort, high impact on mock size)
- **Effort:** low

#### ~~OO-076~~ -- ~~Extract common write helper for evidence~~ (RESOLVED)

- **Package:** evidence
- **Source:** `go/core/evidence/README.md`
- **Detail:** ~~Extract a common writeFile(outputDir, module, scanner, findings) helper to eliminate duplication between module-level and component-level writers~~ Resolved: extracted `writeFile` and `resolveOutputDir` helpers
- **Effort:** low

#### ~~OO-077~~ -- ~~Extract outSubPath helper for paths~~ (RESOLVED)

- **Package:** paths
- **Source:** `go/core/paths/README.md`
- **Detail:** ~~Extract a common outSubPath(repoRoot string, segments ...string) string helper to reduce duplication across the ~30 OutDir-based builders in paths_output.go and paths_cache.go~~ Resolved: extracted `outSubPath` helper, converted 33 functions
- **Effort:** low

#### OO-078 -- Hoist releasenotes regexes to package level

- **Package:** releasenotes
- **Source:** `go/core/releasenotes/README.md`
- **Detail:** Hoist versionHeaderRegex and sectionHeaderRegex to package-level compiled vars (low effort, avoids repeated compilation)
- **Effort:** low

#### OO-079 -- Consolidate FileCache filter methods

- **Package:** repository
- **Source:** `go/core/repository/README.md`
- **Detail:** Consolidate FileCache filter methods into a single generic filter to reduce boilerplate (low effort, ~80 lines saved)
- **Effort:** low

#### OO-080 -- Extract shared resolution template method for resolver

- **Package:** resolver
- **Source:** `go/core/resolver/README.md`
- **Detail:** Extract shared resolution logic (component iteration, tool lookup, weight assignment) into a template method to reduce duplication across build/lint/scan
- **Effort:** medium

#### OO-081 -- Add unit tests for DependencyTracker and unitHeap

- **Package:** scheduling
- **Source:** `go/core/scheduling/README.md`
- **Detail:** Add direct unit tests for DependencyTracker and unitHeap to improve fault localization when scheduling bugs arise (low effort)
- **Effort:** low

#### OO-082 -- Split steps_cache_test.go by feature area

- **Package:** specs
- **Source:** `go/core/specs/README.md`
- **Detail:** Split steps_cache_test.go into smaller files by feature area
- **Effort:** medium

#### OO-083 -- Add benchmark tests for fixture pooling

- **Package:** testing
- **Source:** `go/core/testing/README.md`
- **Detail:** Add benchmark tests for fixture pooling to guard against performance regressions (low effort, high value)
- **Effort:** low

#### OO-084 -- Extract executeContainer helpers for tool

- **Package:** tool
- **Source:** `go/core/tool/README.md`
- **Detail:** Extract executeContainer mount-building and env-forwarding into helper functions (medium effort). Consider merging handler interfaces into a single Handler interface (medium effort)
- **Effort:** medium

#### OO-085 -- Generate error code registry and add formatter tests

- **Package:** validation
- **Source:** `go/core/validation/README.md`
- **Detail:** Extract error code registry into a generated file. Add table-driven tests for ErrorFormatter
- **Effort:** medium

#### OO-086 -- Split StateManager into focused managers

- **Package:** workunit
- **Source:** `go/core/workunit/README.md`
- **Detail:** Split StateManager into focused managers per granularity (UoW, module, test-module) to reduce file size and simplify testing
- **Effort:** medium

#### ~~OO-087~~ -- ~~Pre-compute reverse dependency graph for modules~~ (RESOLVED)

- **Package:** modules
- **Source:** `go/core/domain/modules/README.md`
- **Detail:** ~~Pre-compute the reverse dependency graph in Registry at registration time and invalidate on Add; this is low effort and eliminates repeated O(n) scans~~ Added `reverseDepGraph` field to `Registry`, built incrementally in `Add`; `GetUsedBy` is now O(1) lookup.
- **Effort:** low

#### OO-088 -- Replace injectable environment reader for workspace

- **Package:** workspace
- **Source:** `go/core/workspace/README.md`
- **Detail:** Replace direct os.Getenv calls with an injectable environment reader to simplify testing without real env vars (low effort)
- **Effort:** low

#### ~~OO-089~~ -- ~~Split constants into domain-scoped files~~ (RESOLVED)

- **Package:** environments
- **Source:** `go/core/environments/README.md`
- **Detail:** ~~Split 75 constants into domain-scoped files to reduce diff noise and merge conflicts (low effort)~~
- **Effort:** low

#### ~~OO-090~~ -- ~~Extract shared resolution helper for defaults~~ (RESOLVED)

- **Package:** defaults
- **Source:** `go/core/defaults/README.md`
- **Detail:** ~~The 3-tier resolution pattern is repeated 10 times inside ResolveDefaults; extracting a resolveField helper would halve the function length~~ Resolved: extracted `resolveSliceField` and `resolveStringField` helpers
- **Effort:** low

#### OO-091 -- Extract validation methods from DiskOutputReader

- **Package:** output
- **Source:** `go/core/output/README.md`
- **Detail:** Extract validation methods from DiskOutputReader into a dedicated Validator struct to separate read and verify responsibilities
- **Effort:** low

#### ~~OO-092 -- Make pool allocation vars unexported or constant~~ RESOLVED

- **Package:** resource
- **Source:** `go/core/resource/README.md`
- **Detail:** ~~Consider making HostOnlyAllocation and ContainerAllocation constants or unexported to prevent accidental mutation~~
- **Resolution:** Converted `var` function aliases (`HostOnlyAllocation`, `ContainerAllocation`, `AllocationForWeight`) to proper `func` declarations, making them immutable and preventing accidental reassignment
- **Effort:** low

#### OO-093 -- Pre-build ownership path trie for scale

- **Package:** ownership
- **Source:** `go/core/ownership/README.md`
- **Detail:** Pre-build a trie or path-prefix index for findCandidates() to reduce matching cost on large file sets
- **Effort:** medium

### mcp

#### ~~OO-094~~ -- ~~Cache command discovery~~ (RESOLVED)

- **Package:** eac-mcp-server
- **Source:** `go/mcp/commands/README.md`
- **Detail:** ~~Cache command discovery results after first tools/list call since the command set does not change during a server session~~
- **Resolution:** Added `sync.Once` caching in `getCommands()`; commands are discovered once on first `tools/list` call and reused for the session
- **Effort:** low

### specs

#### OO-095 -- Extract steps_test.go into focused step files

- **Package:** repository
- **Source:** `go/specs/repository/README.md`
- **Detail:** Extract steps_test.go into focused step files (hierarchy, ownership, markdown). Implement the TODO stubs for Gherkin tag conflict detection and link checking
- **Effort:** medium

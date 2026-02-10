# repository/gomod

Parses `go.mod` files across the monorepo, builds a dependency graph between modules, maps between Go module paths and module contract monikers, and validates that actual Go dependencies align with the module registry.

## Key Types

| Type | Purpose |
|------|---------|
| `GoModInfo` | Parsed information from a single `go.mod` file (module path, directory, requires, replaces) |
| `Require` | A require statement with path, version, and indirect flag |
| `Replace` | A replace directive mapping old path to new path |
| `DependencyGraph` | Complete module dependency structure with nodes and dependency edges |
| `ModuleNode` | A single module in the graph with moniker, paths, dependencies, and reverse dependencies |
| `GraphBuilder` | Builder that accumulates `GoModInfo` entries and constructs the `DependencyGraph` |
| `Mapper` | Bidirectional mapper between Go module paths and module contract monikers |
| `Validator` | Validates actual Go dependencies against the module registry |
| `ValidationReport` | Results of dependency validation with discrepancies and summary statistics |
| `Discrepancy` | Per-module comparison of contract vs actual dependencies |
| `ValidationSummary` | High-level counts of total, matching, and discrepant modules |

## Key Functions

| Function | Purpose |
|----------|---------|
| `FindGoModFiles` | Finds all `go.mod` files in a directory tree, excluding vendor/git/node_modules |
| `ParseGoMod` | Parses a single `go.mod` file extracting module path, requires, and replaces |
| `ParseAllGoMods` | Finds and parses all `go.mod` files in the repository |
| `FilterInternalDependencies` | Filters requires to only internal (same-repo) modules |
| `BuildFromDirectory` | Scans directory for `go.mod` files and builds the complete dependency graph |
| `ValidateAndReport` | Performs full validation and returns a formatted report |
| `GetModuleNameFromPath` | Extracts a simple module name from a module directory path |

## Patterns

- **Three-pass graph building**: First pass creates nodes, second pass builds dependencies, third pass calculates reverse dependencies
- **Bidirectional mapping**: `Mapper` maintains `pathToMoniker` and `monikerToPath` maps with longest-prefix matching for subdirectory resolution
- **Registry validation**: `Validator` checks that all actual Go dependencies exist as registered modules
- **Cross-platform paths**: Uses `filepath.ToSlash` for consistent path handling

## Internal Structure

| File | Purpose |
|------|---------|
| `types.go` | Core type definitions (`GoModInfo`, `DependencyGraph`, `ModuleNode`, `ValidationReport`, etc.) |
| `finder.go` | `FindGoModFiles`, `ExtractModuleDir`, `IsInternalModule` |
| `parser.go` | `ParseGoMod`, `ParseAllGoMods`, `FilterInternalDependencies`, `GetModuleNameFromPath` |
| `graph.go` | `GraphBuilder`, `BuildFromDirectory`, `DependencyGraph` query methods |
| `validator.go` | `Validator`, `ValidateAndReport`, `FormatReport`, `ValidationReport` query methods |
| `mapper.go` | `Mapper` with bidirectional path/moniker resolution and prefix matching |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/domain/modules` | `Registry`, `ModuleContract` for module lookups and component root resolution |

## Role in System

Powers the `validate-go-tidy` and dependency validation commands. The dependency graph feeds into CI workflows that verify Go module dependencies match the declared module contracts, catching undeclared or invalid cross-module dependencies.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: The regex patterns in `parser.go` (lines 13-19) are compiled at package init; this is fine but they could be documented more clearly as module-level constants.

package gomod

// GoModInfo contains parsed information from a go.mod file.
type GoModInfo struct {
	FilePath   string    // Absolute path to go.mod file
	ModulePath string    // Module path from "module" declaration
	ModuleDir  string    // Relative directory (e.g., "go/cli/clie")
	Requires   []Require // Direct dependencies
	Replaces   []Replace // Replace directives
}

// Require represents a require statement in go.mod.
type Require struct {
	Path     string // Module path (e.g., "github.com/ready-to-release/eac/go/core")
	Version  string // Version string (e.g., "v0.0.0")
	Indirect bool   // True if marked with // indirect
}

// Replace represents a replace directive in go.mod.
type Replace struct {
	OldPath string // Original module path
	NewPath string // Replacement path (can be local or remote)
}

// DependencyGraph represents the module dependency structure.
type DependencyGraph struct {
	Modules      map[string]*ModuleNode // moniker -> node
	Dependencies map[string][]string    // moniker -> []dependent_monikers
}

// ModuleNode represents a single module in the dependency graph.
type ModuleNode struct {
	Moniker    string   // Module moniker (e.g., "clie")
	ModulePath string   // Full module path
	SourceRoot string   // Relative source root (e.g., "go/cli/clie")
	GoModPath  string   // Path to go.mod file
	DependsOn  []string // Monikers of dependencies
	UsedBy     []string // Monikers of dependents (reverse)
}

// ValidationReport contains results of dependency validation.
type ValidationReport struct {
	Discrepancies []Discrepancy
	Summary       ValidationSummary
}

// Discrepancy represents invalid dependencies for a module.
type Discrepancy struct {
	Moniker              string
	ContractDependencies []string
	ActualDependencies   []string
	Invalid              []string // Dependencies that don't exist as valid modules
	Status               string
}

// ValidationSummary provides high-level validation statistics.
type ValidationSummary struct {
	TotalModules        int
	Matching            int
	WithDiscrepancies   int
	ModulesWithoutGoMod int
}

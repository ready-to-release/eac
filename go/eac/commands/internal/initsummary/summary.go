// Package initsummary provides structured initialization summary for build/test commands.
// It captures all context about what was requested, what will be executed, and verification status.
package initsummary

import "time"

// Summary captures all initialization context for display.
// It separates the data from formatting, enabling different output formats (TUI, CI, JSON).
type Summary struct {
	// Command being executed
	Command string // "build" or "test"

	// Execution context
	ExecutionContext string // local/CI/container/test

	// Request Context
	RequestedModules  []string // What user asked for (CLI args)
	CalculatedModules []string // Final list after dependency resolution
	AddedDepm         []string // Module dependencies added (depm)

	// Execution Plan
	ExecutionLayers [][]string // Modules grouped by dependency layer
	LayerCount      int        // Number of layers

	// Flags & Configuration
	Flags Flags // All active flags

	// Dependency Status
	DepmStatus DepmStatus // Module dependency (depm) verification
	DepsStatus DepsStatus // System dependency (deps) verification

	// Incremental Build (if applicable, build only)
	Incremental *IncrementalInfo // nil if not used

	// Test-specific fields (only populated for test command)
	Test *TestInfo // nil for build command

	// Build artifact validation (test only)
	ArtifactValidation *ArtifactValidationInfo // nil if not validated

	// Output directory
	OutputDir string // out/build/ or out/test/<suite>/
}

// Flags captures all command-line flags that affect execution.
type Flags struct {
	// Go module management (build)
	TidyFirst    bool // Run go mod tidy before build
	TidyExplicit bool // Was --tidy-first or --no-tidy explicitly set?

	// Module dependency handling (depm)
	SkipDepm        bool // --skip-depm: don't include transitive module deps
	UseExistingDepm bool // --use-existing-depm: skip building if artifacts exist

	// System dependency handling (deps)
	SkipDepsVerification bool // --skip-deps-verification: skip go, docker, etc. checks

	// Build behavior
	ForceRebuild bool   // --rebuild: force full rebuild, ignore incremental state
	DryRun       bool   // --dry-run: simulate without executing
	BuildAll     bool   // --all: include non-default artifacts
	Version      string // --version: version string to inject

	// Test behavior
	ListOnly bool // --list-only: list tests without running

	// Output/display
	ShowTimings bool // --timings: show timing analysis
	DebugMode   bool // --debug: enable debug logs to console
	UseTUI      bool // --tui/--no-tui: TUI mode
}

// DepmStatus captures module dependency (depm) verification results.
type DepmStatus struct {
	Verified bool     // Were module deps checked? (false if --skip-depm)
	Skipped  bool     // Was --skip-depm used?
	Total    int      // Total dependency modules found
	Resolved []string // Successfully resolved module deps
	Missing  []string // Failed to resolve (would cause build failure)
}

// DepsStatus captures system dependency (deps) verification results.
type DepsStatus struct {
	Verified  bool         // Were system deps checked? (false if --skip-deps-verification)
	Skipped   bool         // Was --skip-deps-verification used?
	Required  []string     // Required system deps (go, docker, npm, etc.)
	Available []DepsResult // Verification results for each
	Missing   []string     // Missing deps (would cause build failure)
}

// DepsResult captures verification result for a single system dependency.
type DepsResult struct {
	Name      string // Dependency name (go, docker, npm)
	Available bool   // Is it available?
	Version   string // Detected version (empty if not available)
}

// IncrementalInfo captures incremental build detection results (local devbox only).
type IncrementalInfo struct {
	Enabled       bool          // Is incremental mode active?
	DetectionTime time.Duration // How long detection took
	Changed       []string      // Modules that will be built (changed + propagated + explicitly requested)
	UpToDate      []string      // Modules that will be skipped (up-to-date and not requested)
	FreshBuild    bool          // True if no prior state (first build)
}

// TestInfo captures test-specific initialization data.
type TestInfo struct {
	// Suite information
	SuiteName        string // Name of the test suite
	SuiteDescription string // Suite description

	// Discovery stats (tests)
	TotalDiscovered  int // Total tests discovered in codebase
	Skipped          int // Tests skipped due to @skip tags
	NotMatchingSuite int // Tests not matching suite criteria
	OSFiltered       int // Tests filtered by OS compatibility
	Selected         int // Final selected tests to run

	// Module stats
	ModulesRequested   []string // Modules requested via CLI (empty = all)
	ModulesInScope     []string // Modules with tests that will run
	ModulesNoTests     []string // Requested modules with no matching tests

	// Tag inference
	InferenceRulesApplied int // Number of inference rules applied
}

// ArtifactValidationInfo captures build artifact validation results.
type ArtifactValidationInfo struct {
	Validated              bool                // Was validation performed?
	ModulesChecked         []string            // Modules whose artifacts were validated
	AllPresent             bool                // Were all required artifacts present?
	MissingFrom            []string            // Modules with missing artifacts
	MissingArtifactDetails map[string][]string // Module -> list of missing artifact IDs

	// Staleness check (source files changed since build)
	AllCurrent   bool              // Are all builds current (source unchanged)?
	StaleModules []string          // Modules where source changed since build
	StaleReasons map[string]string // Module -> reason for staleness
}

// AllValid returns true if all artifacts are present AND all builds are current.
func (a *ArtifactValidationInfo) AllValid() bool {
	return a.AllPresent && a.AllCurrent
}

// New creates a new Summary with command type.
func New(command string) *Summary {
	return &Summary{
		Command:           command,
		RequestedModules:  []string{},
		CalculatedModules: []string{},
		AddedDepm:         []string{},
		ExecutionLayers:   [][]string{},
	}
}

// SetRequest sets the requested and calculated modules.
func (s *Summary) SetRequest(requested, calculated []string) *Summary {
	s.RequestedModules = requested
	s.CalculatedModules = calculated

	// Calculate added depm (modules in calculated but not in requested)
	requestedSet := make(map[string]bool)
	for _, m := range requested {
		requestedSet[m] = true
	}
	s.AddedDepm = []string{}
	for _, m := range calculated {
		if !requestedSet[m] {
			s.AddedDepm = append(s.AddedDepm, m)
		}
	}
	return s
}

// SetExecutionPlan sets the execution layers from the calculated plan.
func (s *Summary) SetExecutionPlan(layers [][]string) *Summary {
	s.ExecutionLayers = layers
	s.LayerCount = len(layers)
	return s
}

// SetFlags sets the flag configuration.
func (s *Summary) SetFlags(flags Flags) *Summary {
	s.Flags = flags
	return s
}

// SetExecutionContext sets the execution context (local, CI, container, test).
func (s *Summary) SetExecutionContext(ctx string) *Summary {
	s.ExecutionContext = ctx
	return s
}

// SetDepmStatus sets the module dependency verification status.
func (s *Summary) SetDepmStatus(status DepmStatus) *Summary {
	s.DepmStatus = status
	return s
}

// SetDepsStatus sets the system dependency verification status.
func (s *Summary) SetDepsStatus(status DepsStatus) *Summary {
	s.DepsStatus = status
	return s
}

// SetIncremental sets the incremental build detection results.
func (s *Summary) SetIncremental(info *IncrementalInfo) *Summary {
	s.Incremental = info
	return s
}

// SetTestInfo sets test-specific information.
func (s *Summary) SetTestInfo(info *TestInfo) *Summary {
	s.Test = info
	return s
}

// SetArtifactValidation sets build artifact validation results.
func (s *Summary) SetArtifactValidation(info *ArtifactValidationInfo) *Summary {
	s.ArtifactValidation = info
	return s
}

// SetOutputDir sets the output directory path.
func (s *Summary) SetOutputDir(dir string) *Summary {
	s.OutputDir = dir
	return s
}

// HasDepmAdded returns true if module dependencies were added to the request.
func (s *Summary) HasDepmAdded() bool {
	return len(s.AddedDepm) > 0
}

// HasIncremental returns true if incremental detection was performed.
func (s *Summary) HasIncremental() bool {
	return s.Incremental != nil
}

// HasTestInfo returns true if this is a test command with test info.
func (s *Summary) HasTestInfo() bool {
	return s.Test != nil
}

// HasArtifactValidation returns true if artifact validation was performed.
func (s *Summary) HasArtifactValidation() bool {
	return s.ArtifactValidation != nil
}

// TotalModules returns the total number of modules to be processed.
func (s *Summary) TotalModules() int {
	return len(s.CalculatedModules)
}

// LayerSizes returns the size of each execution layer.
func (s *Summary) LayerSizes() []int {
	sizes := make([]int, len(s.ExecutionLayers))
	for i, layer := range s.ExecutionLayers {
		sizes[i] = len(layer)
	}
	return sizes
}

// IsBuild returns true if this is a build command.
func (s *Summary) IsBuild() bool {
	return s.Command == "build"
}

// IsTest returns true if this is a test command.
func (s *Summary) IsTest() bool {
	return s.Command == "test"
}

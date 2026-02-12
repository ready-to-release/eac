package display

import "github.com/ready-to-release/eac/go/core/workunit"

// InitSummary holds structured init summary data for the Init pane.
type InitSummary struct {
	Command           string
	ExecutionContext   string
	RequestedModules  int
	CalculatedModules int
	AddedDepm         int
	UoWCount          int
	ExecutionTree     []ExecutionModule
	ParallelismMode   string
	EffectiveWorkers  int
	TurboBoost        int
	WeightedCapacity  int
	Flags             InitSummaryFlags
	DepsVerified      bool
	DepsSkipped       bool
	DepsAvailable     []string
	DepsMissing       []string
	DepmVerified      bool
	DepmSkipped       bool
	DepmResolved      int
	DepmExisting      int
	DepmTotal         int
	DepmMissing       []string
	IncrementalEnabled  bool
	IncrementalChanged  int
	IncrementalUpToDate int
	IncrementalFresh    bool
	TestSuiteName     string
	TestSelected      int
	TestDiscovered    int
	TestOSFiltered    int
	OutputDir         string
	PlannedTools      []PlannedTool
}

// ExecutionModule represents a module and its UoWs.
type ExecutionModule struct {
	Name string
	UoWs []UoWEntry
}

// UoWEntry represents a unit of work with its globally unique ID.
type UoWEntry struct {
	ID            string
	DisplayName   string
	Weight        int
	Tags          workunit.TagSummary
	Module        string
	Component     string
	Tool          string
	ComponentType string
	Container     bool
}

// InitSummaryFlags captures relevant flags for display.
type InitSummaryFlags struct {
	TidyFirst    bool
	ForceRebuild bool
	DryRun       bool
	UseTUI       bool
	SkipDeps     bool
	SkipDepm     bool
}

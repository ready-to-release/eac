package display

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Phase.String()
// ---------------------------------------------------------------------------

func TestPhaseString(t *testing.T) {
	tests := []struct {
		name  string
		phase Phase
		want  string
	}{
		{
			name:  "PhaseInit returns Initialization",
			phase: PhaseInit,
			want:  PhaseNameInitialization,
		},
		{
			name:  "PhaseRun returns Run",
			phase: PhaseRun,
			want:  PhaseNameRun,
		},
		{
			name:  "PhaseSummary returns Summary",
			phase: PhaseSummary,
			want:  PhaseNameSummary,
		},
		{
			name:  "unknown positive phase returns Unknown",
			phase: Phase(99),
			want:  PhaseNameUnknown,
		},
		{
			name:  "unknown negative phase returns Unknown",
			phase: Phase(-1),
			want:  PhaseNameUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phase.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPhaseIotaOrder(t *testing.T) {
	// Verify the iota ordering is Init=0, Run=1, Summary=2.
	// This matters because consumers may persist phase values as integers.
	assert.Equal(t, Phase(0), PhaseInit)
	assert.Equal(t, Phase(1), PhaseRun)
	assert.Equal(t, Phase(2), PhaseSummary)
}

// ---------------------------------------------------------------------------
// Phase display name constants
// ---------------------------------------------------------------------------

func TestPhaseNameConstants(t *testing.T) {
	assert.Equal(t, "Initialization", PhaseNameInitialization)
	assert.Equal(t, "Run", PhaseNameRun)
	assert.Equal(t, "Summary", PhaseNameSummary)
	assert.Equal(t, "Unknown", PhaseNameUnknown)
}

// ---------------------------------------------------------------------------
// Level iota order
// ---------------------------------------------------------------------------

func TestLevelIotaOrder(t *testing.T) {
	// Verify the iota ordering is Info=0, Warn=1, Error=2.
	assert.Equal(t, Level(0), LevelInfo)
	assert.Equal(t, Level(1), LevelWarn)
	assert.Equal(t, Level(2), LevelError)
}

// ---------------------------------------------------------------------------
// CacheHit iota order
// ---------------------------------------------------------------------------

func TestCacheHitIotaOrder(t *testing.T) {
	assert.Equal(t, CacheHit(0), CacheUnknown)
	assert.Equal(t, CacheHit(1), CacheMiss)
	assert.Equal(t, CacheHit(2), CacheHitFresh)
}

// ---------------------------------------------------------------------------
// Default constants
// ---------------------------------------------------------------------------

func TestDefaultConstants(t *testing.T) {
	assert.Equal(t, 40, DefaultHeight)
	assert.Equal(t, 1000, DefaultBufferSize)
}

// ---------------------------------------------------------------------------
// SetTUIBootstrap / GetTUIBootstrap
// ---------------------------------------------------------------------------

func TestTUIBootstrapGetSetRoundTrip(t *testing.T) {
	// Save original and restore after test to avoid polluting global state.
	original := globalBootstrap
	t.Cleanup(func() { globalBootstrap = original })

	// Initially may or may not be nil depending on test order,
	// so explicitly clear it first.
	SetTUIBootstrap(nil)
	assert.Nil(t, GetTUIBootstrap(), "expected nil after setting nil")

	bootstrap := &TUIBootstrap{}
	SetTUIBootstrap(bootstrap)
	got := GetTUIBootstrap()
	require.NotNil(t, got)
	assert.Same(t, bootstrap, got, "should return the exact same pointer")
}

func TestTUIBootstrapOverwrite(t *testing.T) {
	original := globalBootstrap
	t.Cleanup(func() { globalBootstrap = original })

	first := &TUIBootstrap{}
	second := &TUIBootstrap{}

	SetTUIBootstrap(first)
	assert.Same(t, first, GetTUIBootstrap())

	SetTUIBootstrap(second)
	assert.Same(t, second, GetTUIBootstrap(), "second set should overwrite first")
}

func TestTUIBootstrapClearWithNil(t *testing.T) {
	original := globalBootstrap
	t.Cleanup(func() { globalBootstrap = original })

	SetTUIBootstrap(&TUIBootstrap{})
	require.NotNil(t, GetTUIBootstrap())

	SetTUIBootstrap(nil)
	assert.Nil(t, GetTUIBootstrap(), "setting nil should clear the bootstrap")
}

// ---------------------------------------------------------------------------
// Config zero-value
// ---------------------------------------------------------------------------

func TestConfigZeroValue(t *testing.T) {
	var cfg Config

	assert.Equal(t, 0, cfg.Height)
	assert.Equal(t, 0, cfg.BufferSize)
	assert.False(t, cfg.ASCIIMode)
	assert.False(t, cfg.TUI3Demo)
	assert.False(t, cfg.SkipTUIDelay)
	assert.Empty(t, cfg.RunPhaseName)
	assert.Empty(t, cfg.CommandName)
}

// ---------------------------------------------------------------------------
// Line struct
// ---------------------------------------------------------------------------

func TestLineStruct(t *testing.T) {
	now := time.Now()
	line := Line{
		Text:      "build successful",
		Source:    "module-a",
		Level:     LevelWarn,
		Timestamp: now,
	}

	assert.Equal(t, "build successful", line.Text)
	assert.Equal(t, "module-a", line.Source)
	assert.Equal(t, LevelWarn, line.Level)
	assert.Equal(t, now, line.Timestamp)
}

// ---------------------------------------------------------------------------
// Status struct
// ---------------------------------------------------------------------------

func TestStatusStruct(t *testing.T) {
	status := Status{
		Phase:     "Run",
		Running:   []string{"mod-a", "mod-b"},
		Completed: 3,
		Total:     10,

		Roof:           8,
		PressureTarget: 6,

		Locks: []LockStatus{
			{Name: "docker", Type: "semaphore", Capacity: 4, Used: 2, Waiting: 1},
		},

		ActiveContainerTools: []string{"docker"},
		UsedContainerTools:   []string{"docker", "podman"},
		ActiveSystemTools:    []string{"go"},
		UsedSystemTools:      []string{"go", "npm"},

		DockerMemPercent: 75.5,
		DockerAvailable:  true,

		RunningContainerCount: 2,
		TotalContainerCount:   5,
		RunningSystemCount:    3,
		TotalSystemCount:      7,
	}

	assert.Equal(t, "Run", status.Phase)
	assert.Equal(t, []string{"mod-a", "mod-b"}, status.Running)
	assert.Equal(t, 3, status.Completed)
	assert.Equal(t, 10, status.Total)
	assert.Equal(t, 8, status.Roof)
	assert.Equal(t, 6, status.PressureTarget)
	assert.Len(t, status.Locks, 1)
	assert.Equal(t, "docker", status.Locks[0].Name)
	assert.Equal(t, 4, status.Locks[0].Capacity)
	assert.Equal(t, 2, status.Locks[0].Used)
	assert.Equal(t, 1, status.Locks[0].Waiting)
	assert.Equal(t, 75.5, status.DockerMemPercent)
	assert.True(t, status.DockerAvailable)
	assert.Equal(t, 2, status.RunningContainerCount)
	assert.Equal(t, 5, status.TotalContainerCount)
	assert.Equal(t, 3, status.RunningSystemCount)
	assert.Equal(t, 7, status.TotalSystemCount)
}

// ---------------------------------------------------------------------------
// SummaryData struct
// ---------------------------------------------------------------------------

func TestSummaryDataStruct(t *testing.T) {
	data := SummaryData{
		Success:     true,
		TotalTime:   5 * time.Second,
		InitSummary: "initialized 3 modules",
		RunSummary:  "all passed",
		Details:     []string{"detail-1", "detail-2"},
		NextSteps:   "run deploy",
	}

	assert.True(t, data.Success)
	assert.Equal(t, 5*time.Second, data.TotalTime)
	assert.Equal(t, "initialized 3 modules", data.InitSummary)
	assert.Equal(t, "all passed", data.RunSummary)
	assert.Equal(t, []string{"detail-1", "detail-2"}, data.Details)
	assert.Equal(t, "run deploy", data.NextSteps)
}

// ---------------------------------------------------------------------------
// InitSummary struct
// ---------------------------------------------------------------------------

func TestInitSummaryStruct(t *testing.T) {
	summary := InitSummary{
		Command:           "build",
		ExecutionContext:   "ci",
		RequestedModules:  5,
		CalculatedModules: 8,
		AddedDepm:         3,
		UoWCount:          12,
		ParallelismMode:   "auto",
		EffectiveWorkers:  4,
		TurboBoost:        2,
		WeightedCapacity:  16,
		Flags: InitSummaryFlags{
			TidyFirst:    true,
			ForceRebuild: false,
			DryRun:       true,
			UseTUI:       true,
			SkipDeps:     false,
			SkipDepm:     true,
		},
		DepsVerified:        true,
		DepsSkipped:         false,
		DepsAvailable:       []string{"go", "docker"},
		DepsMissing:         []string{},
		DepmVerified:        true,
		DepmSkipped:         false,
		DepmResolved:        5,
		DepmExisting:        3,
		DepmTotal:           8,
		DepmMissing:         []string{"lib-x"},
		IncrementalEnabled:  true,
		IncrementalChanged:  2,
		IncrementalUpToDate: 6,
		IncrementalFresh:    false,
		TestSuiteName:       "unit",
		TestSelected:        10,
		TestDiscovered:      15,
		TestOSFiltered:      2,
		OutputDir:           "/tmp/output",
		PlannedTools: []PlannedTool{
			{Name: "go", IsContainer: false},
			{Name: "docker", IsContainer: true},
		},
	}

	assert.Equal(t, "build", summary.Command)
	assert.Equal(t, "ci", summary.ExecutionContext)
	assert.Equal(t, 5, summary.RequestedModules)
	assert.Equal(t, 8, summary.CalculatedModules)
	assert.Equal(t, 3, summary.AddedDepm)
	assert.Equal(t, 12, summary.UoWCount)
	assert.Equal(t, "auto", summary.ParallelismMode)
	assert.Equal(t, 4, summary.EffectiveWorkers)
	assert.Equal(t, 2, summary.TurboBoost)
	assert.Equal(t, 16, summary.WeightedCapacity)

	assert.True(t, summary.Flags.TidyFirst)
	assert.False(t, summary.Flags.ForceRebuild)
	assert.True(t, summary.Flags.DryRun)
	assert.True(t, summary.Flags.UseTUI)
	assert.False(t, summary.Flags.SkipDeps)
	assert.True(t, summary.Flags.SkipDepm)

	assert.True(t, summary.DepsVerified)
	assert.False(t, summary.DepsSkipped)
	assert.Equal(t, []string{"go", "docker"}, summary.DepsAvailable)
	assert.Empty(t, summary.DepsMissing)

	assert.True(t, summary.DepmVerified)
	assert.Equal(t, 5, summary.DepmResolved)
	assert.Equal(t, 3, summary.DepmExisting)
	assert.Equal(t, 8, summary.DepmTotal)
	assert.Equal(t, []string{"lib-x"}, summary.DepmMissing)

	assert.True(t, summary.IncrementalEnabled)
	assert.Equal(t, 2, summary.IncrementalChanged)
	assert.Equal(t, 6, summary.IncrementalUpToDate)
	assert.False(t, summary.IncrementalFresh)

	assert.Equal(t, "unit", summary.TestSuiteName)
	assert.Equal(t, 10, summary.TestSelected)
	assert.Equal(t, 15, summary.TestDiscovered)
	assert.Equal(t, 2, summary.TestOSFiltered)
	assert.Equal(t, "/tmp/output", summary.OutputDir)

	require.Len(t, summary.PlannedTools, 2)
	assert.Equal(t, "go", summary.PlannedTools[0].Name)
	assert.False(t, summary.PlannedTools[0].IsContainer)
	assert.Equal(t, "docker", summary.PlannedTools[1].Name)
	assert.True(t, summary.PlannedTools[1].IsContainer)
}

// ---------------------------------------------------------------------------
// ExecutionModule and UoWEntry
// ---------------------------------------------------------------------------

func TestExecutionModuleWithUoWEntries(t *testing.T) {
	module := ExecutionModule{
		Name: "core",
		UoWs: []UoWEntry{
			{
				ID:            "core:unit:go",
				DisplayName:   "core unit tests",
				Weight:        3,
				Module:        "core",
				Component:     "unit",
				Tool:          "go",
				ComponentType: "test",
				Container:     false,
			},
			{
				ID:            "core:lint:golangci",
				DisplayName:   "core linting",
				Weight:        1,
				Module:        "core",
				Component:     "lint",
				Tool:          "golangci",
				ComponentType: "lint",
				Container:     true,
			},
		},
	}

	assert.Equal(t, "core", module.Name)
	require.Len(t, module.UoWs, 2)
	assert.Equal(t, "core:unit:go", module.UoWs[0].ID)
	assert.Equal(t, 3, module.UoWs[0].Weight)
	assert.False(t, module.UoWs[0].Container)
	assert.True(t, module.UoWs[1].Container)
}

// ---------------------------------------------------------------------------
// PlannedWorkItem
// ---------------------------------------------------------------------------

func TestPlannedWorkItem(t *testing.T) {
	item := PlannedWorkItem{
		ID:            "mod-a:build:go",
		DisplayName:   "mod-a build",
		Weight:        2,
		Module:        "mod-a",
		Component:     "build",
		ComponentType: "build",
	}

	assert.Equal(t, "mod-a:build:go", item.ID)
	assert.Equal(t, "mod-a build", item.DisplayName)
	assert.Equal(t, 2, item.Weight)
	assert.Equal(t, "mod-a", item.Module)
	assert.Equal(t, "build", item.Component)
	assert.Equal(t, "build", item.ComponentType)
}

// ---------------------------------------------------------------------------
// UoWEnrichment
// ---------------------------------------------------------------------------

func TestUoWEnrichment(t *testing.T) {
	now := time.Now()
	enrichment := UoWEnrichment{
		ID:          "mod-a:test:go",
		Tool:        "go",
		Container:   false,
		Weight:      5,
		CacheStatus: CacheHitFresh,
		CacheTime:   now,
		DependsOn:   []string{"mod-b:build:go"},
	}

	assert.Equal(t, "mod-a:test:go", enrichment.ID)
	assert.Equal(t, "go", enrichment.Tool)
	assert.False(t, enrichment.Container)
	assert.Equal(t, 5, enrichment.Weight)
	assert.Equal(t, CacheHitFresh, enrichment.CacheStatus)
	assert.Equal(t, now, enrichment.CacheTime)
	assert.Equal(t, []string{"mod-b:build:go"}, enrichment.DependsOn)
}

func TestUoWEnrichmentCacheStatuses(t *testing.T) {
	tests := []struct {
		name   string
		status CacheHit
	}{
		{name: "unknown cache", status: CacheUnknown},
		{name: "cache miss", status: CacheMiss},
		{name: "cache hit fresh", status: CacheHitFresh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := UoWEnrichment{CacheStatus: tt.status}
			assert.Equal(t, tt.status, e.CacheStatus)
		})
	}
}

// ---------------------------------------------------------------------------
// LockStatus
// ---------------------------------------------------------------------------

func TestLockStatus(t *testing.T) {
	lock := LockStatus{
		Name:     "docker",
		Type:     "semaphore",
		Capacity: 4,
		Used:     2,
		Waiting:  1,
	}

	assert.Equal(t, "docker", lock.Name)
	assert.Equal(t, "semaphore", lock.Type)
	assert.Equal(t, 4, lock.Capacity)
	assert.Equal(t, 2, lock.Used)
	assert.Equal(t, 1, lock.Waiting)
}

// ---------------------------------------------------------------------------
// TUIBootstrap function fields
// ---------------------------------------------------------------------------

func TestTUIBootstrapWithFunctions(t *testing.T) {
	original := globalBootstrap
	t.Cleanup(func() { globalBootstrap = original })

	var newForCommandCalled bool
	var newObserverCalled bool
	var newHooksCalled bool
	var unwrapCalled bool

	bootstrap := &TUIBootstrap{
		NewForCommand: func(commandPath string, config Config) Console {
			newForCommandCalled = true
			assert.Equal(t, "build", commandPath)
			return nil
		},
		NewObserver: func(console Console) (interface{}, interface{}) {
			newObserverCalled = true
			return nil, nil
		},
		NewHooks: func(console Console) interface{} {
			newHooksCalled = true
			return nil
		},
		UnwrapConsole: func(console Console) Console {
			unwrapCalled = true
			return console
		},
	}

	SetTUIBootstrap(bootstrap)
	got := GetTUIBootstrap()

	// Exercise the function fields
	got.NewForCommand("build", Config{})
	assert.True(t, newForCommandCalled, "NewForCommand should have been called")

	got.NewObserver(nil)
	assert.True(t, newObserverCalled, "NewObserver should have been called")

	got.NewHooks(nil)
	assert.True(t, newHooksCalled, "NewHooks should have been called")

	got.UnwrapConsole(nil)
	assert.True(t, unwrapCalled, "UnwrapConsole should have been called")
}

// ---------------------------------------------------------------------------
// Phase.String() is usable in formatted output
// ---------------------------------------------------------------------------

func TestPhaseStringUsableInFormatting(t *testing.T) {
	tests := []struct {
		name   string
		phase  Phase
		format string
		want   string
	}{
		{
			name:   "Sprintf with Init phase",
			phase:  PhaseInit,
			format: "Initialization",
			want:   "Initialization",
		},
		{
			name:   "Sprintf with Run phase",
			phase:  PhaseRun,
			format: "Run",
			want:   "Run",
		},
		{
			name:   "Sprintf with Summary phase",
			phase:  PhaseSummary,
			format: "Summary",
			want:   "Summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.phase.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// InitSummaryFlags zero value
// ---------------------------------------------------------------------------

func TestInitSummaryFlagsZeroValue(t *testing.T) {
	var flags InitSummaryFlags

	assert.False(t, flags.TidyFirst)
	assert.False(t, flags.ForceRebuild)
	assert.False(t, flags.DryRun)
	assert.False(t, flags.UseTUI)
	assert.False(t, flags.SkipDeps)
	assert.False(t, flags.SkipDepm)
}

// ---------------------------------------------------------------------------
// PlannedTool
// ---------------------------------------------------------------------------

func TestPlannedTool(t *testing.T) {
	tests := []struct {
		name        string
		tool        PlannedTool
		wantName    string
		wantIsCtner bool
	}{
		{
			name:        "system tool",
			tool:        PlannedTool{Name: "go", IsContainer: false},
			wantName:    "go",
			wantIsCtner: false,
		},
		{
			name:        "container tool",
			tool:        PlannedTool{Name: "docker", IsContainer: true},
			wantName:    "docker",
			wantIsCtner: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.tool.Name)
			assert.Equal(t, tt.wantIsCtner, tt.tool.IsContainer)
		})
	}
}

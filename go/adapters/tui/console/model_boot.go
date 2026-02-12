package console

// BootState tracks progressive layout readiness during initialization.
// The state only advances forward, never regresses.
type BootState int

const (
	BootChrome  BootState = iota // Frame only, no data
	BootMetrics                  // System metrics available (CPU, memory)
	BootConfig                   // Command config available (name, workers, mode)
	BootPlanned                  // Skeleton tabs from module discovery (grey, before tool resolution)
	BootTabs                     // Tabs registered, ready for execution
	BootRunning                  // Execution underway
)

// ConfigMeta holds early configuration data for progressive display.
// Populated by ConfigReadyMsg before full InitSummary arrives.
type ConfigMeta struct {
	CommandName      string // "build", "test", "lint", "scan"
	ExecutionContext string // "local", "CI"
	ParallelismMode  string // "devbox", "ci"
	EffectiveWorkers int    // Worker count
	WeightedCapacity int    // Scheduler capacity
	OutputDir        string // "out/build", "out/test"
}

// ModelBootState tracks boot-related fields for progressive layout readiness.
type ModelBootState struct {
	State  BootState   // Progressive layout readiness
	Config *ConfigMeta // Early config data (nil until ConfigReadyMsg)
}

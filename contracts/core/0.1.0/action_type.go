package core

import "sync"

// ActionType represents the type of operation for a work unit.
type ActionType string

const (
	ActionBuild     ActionType = "build"
	ActionTest      ActionType = "test"
	ActionScan      ActionType = "scan"
	ActionLint      ActionType = "lint"
	ActionAISummary ActionType = "ai-summary"
	ActionServe     ActionType = "serve"
)

// ActionDescriptor provides display metadata for an action type.
type ActionDescriptor struct {
	Type      ActionType
	Verb      string // "Building"
	PastVerb  string // "built"
	OutputDir string // "out/build"
	LogFile   string // "build.log"
}

var (
	actionRegistryMu sync.RWMutex
	actionRegistry   = make(map[ActionType]ActionDescriptor)
)

// RegisterActionType registers a descriptor for an action type.
// Call this from an init() function in any package that introduces a new
// ActionType. The six built-in types (build, test, scan, lint, ai-summary,
// serve) are pre-registered by this package's own init().
// Registering an already-registered type silently overwrites the previous entry.
func RegisterActionType(d ActionDescriptor) {
	actionRegistryMu.Lock()
	defer actionRegistryMu.Unlock()
	actionRegistry[d.Type] = d
}

// GetActionDescriptor returns the descriptor for an action type.
// The signature is unchanged; all existing callers continue to compile
// and behave identically.
func GetActionDescriptor(t ActionType) (ActionDescriptor, bool) {
	actionRegistryMu.RLock()
	defer actionRegistryMu.RUnlock()
	d, ok := actionRegistry[t]
	return d, ok
}

func init() {
	RegisterActionType(ActionDescriptor{ActionBuild, "Building", "built", "out/build", "build.log"})
	RegisterActionType(ActionDescriptor{ActionTest, "Testing", "tested", "out/test", "test.log"})
	RegisterActionType(ActionDescriptor{ActionScan, "Scanning", "scanned", "out/security", "scan.log"})
	RegisterActionType(ActionDescriptor{ActionLint, "Linting", "linted", "out/lint", "lint.log"})
	RegisterActionType(ActionDescriptor{ActionAISummary, "Summarizing", "summarized", "out/ai-summary", "ai-summary.log"})
	RegisterActionType(ActionDescriptor{ActionServe, "Serving", "served", "out/serve", "serve.log"})
}

package core

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

var actionRegistry = map[ActionType]ActionDescriptor{
	ActionBuild:     {ActionBuild, "Building", "built", "out/build", "build.log"},
	ActionTest:      {ActionTest, "Testing", "tested", "out/test", "test.log"},
	ActionScan:      {ActionScan, "Scanning", "scanned", "out/security", "scan.log"},
	ActionLint:      {ActionLint, "Linting", "linted", "out/lint", "lint.log"},
	ActionAISummary: {ActionAISummary, "Summarizing", "summarized", "out/ai-summary", "ai-summary.log"},
}

// GetActionDescriptor returns the descriptor for an action type.
func GetActionDescriptor(t ActionType) (ActionDescriptor, bool) {
	d, ok := actionRegistry[t]
	return d, ok
}

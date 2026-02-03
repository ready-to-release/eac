package workunit

// Context represents the operation type for a unit of work.
type Context string

const (
	// ContextBuild represents a build operation.
	ContextBuild Context = "build"

	// ContextTest represents a test operation.
	ContextTest Context = "test"

	// ContextLint represents a lint operation.
	ContextLint Context = "lint"

	// ContextScan represents a scan operation.
	ContextScan Context = "scan"

	// ContextAISummary represents an AI summary operation.
	ContextAISummary Context = "ai-summary"
)

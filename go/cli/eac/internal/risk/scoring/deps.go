package scoring

// Deps holds injectable dependencies for testing.
type Deps struct {
	AIResponse string // Non-empty bypasses AI, returns this response directly
}

func defaultDeps() *Deps {
	return &Deps{}
}

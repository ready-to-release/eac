// Package internal provides security scanner implementations.
package internal

// MockContext centralizes all security tool mocking in one place.
// It provides a single point for setting and resetting all security scanner mocks.
type MockContext struct {
	TrivyOutput   interface{}
	SemgrepOutput interface{}
	ZAPOutput     interface{}
}

// SetAllMocks sets all security tool mocks from a MockContext.
func SetAllMocks(ctx *MockContext) {
	if ctx == nil {
		return
	}

	if ctx.TrivyOutput != nil {
		SetMockTrivyOutput(ctx.TrivyOutput)
	}
	if ctx.SemgrepOutput != nil {
		SetMockSemgrepOutput(ctx.SemgrepOutput)
	}
	if ctx.ZAPOutput != nil {
		SetMockZAPOutput(ctx.ZAPOutput)
	}
}

// ResetAllMocks clears all security tool mocks.
func ResetAllMocks() {
	ResetMockTrivyOutput()
	ResetMockSemgrepOutput()
	ResetMockZAPOutput()
}

// HasMocksSet returns true if any security tool mocks are currently set.
func HasMocksSet() bool {
	return mockTrivyOutput != nil ||
		mockSemgrepOutput != nil ||
		mockZAPOutput != nil
}

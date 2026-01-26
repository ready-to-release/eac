package toolcontainer

import (
	"context"
)

// MockRunner is a mock implementation of Runner for testing.
type MockRunner struct {
	mode       Mode
	runFunc    func(ctx context.Context, cfg *RunConfig) *RunResult
	closeFunc  func() error
	runHistory []*RunConfig
}

// NewMockRunner creates a new mock runner.
func NewMockRunner(mode Mode) *MockRunner {
	return &MockRunner{
		mode:       mode,
		runHistory: make([]*RunConfig, 0),
	}
}

// Mode returns the configured mode.
func (m *MockRunner) Mode() Mode {
	return m.mode
}

// Run executes the mock run function if set, otherwise returns success.
func (m *MockRunner) Run(ctx context.Context, cfg *RunConfig) *RunResult {
	m.runHistory = append(m.runHistory, cfg)

	if m.runFunc != nil {
		return m.runFunc(ctx, cfg)
	}
	return &RunResult{ExitCode: 0}
}

// Close executes the mock close function if set.
func (m *MockRunner) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// WithRunFunc sets a custom run function for the mock.
func (m *MockRunner) WithRunFunc(f func(ctx context.Context, cfg *RunConfig) *RunResult) *MockRunner {
	m.runFunc = f
	return m
}

// WithCloseFunc sets a custom close function for the mock.
func (m *MockRunner) WithCloseFunc(f func() error) *MockRunner {
	m.closeFunc = f
	return m
}

// RunHistory returns the history of run calls.
func (m *MockRunner) RunHistory() []*RunConfig {
	return m.runHistory
}

// LastRun returns the last run configuration, or nil if no runs.
func (m *MockRunner) LastRun() *RunConfig {
	if len(m.runHistory) == 0 {
		return nil
	}
	return m.runHistory[len(m.runHistory)-1]
}

// Ensure MockRunner implements Runner.
var _ Runner = (*MockRunner)(nil)

package eac

import (
	"context"
)

// MockEAC is a simple mock for unit tests that records calls
// and returns predefined results.
type MockEAC struct {
	Calls   []MockCall
	Results map[string]*Result // keyed by first arg (e.g., "build", "show")
	Default *Result
	// ErrorFunc optionally returns an error for specific calls.
	// When set, it takes precedence over Results/Default.
	ErrorFunc func(args []string) error
}

// MockCall records a single call to Execute.
type MockCall struct {
	Args   []string
	Config *ExecConfig
}

func (m *MockEAC) Execute(_ context.Context, args []string, cfg *ExecConfig) (*Result, error) {
	m.Calls = append(m.Calls, MockCall{Args: args, Config: cfg})

	if m.ErrorFunc != nil {
		if err := m.ErrorFunc(args); err != nil {
			return nil, err
		}
	}

	if len(args) > 0 {
		if r, ok := m.Results[args[0]]; ok {
			return r, nil
		}
	}
	if m.Default != nil {
		return m.Default, nil
	}
	return &Result{ExitCode: 0}, nil
}

// NewMock creates a MockEAC with sensible defaults for testing.
func NewMock() *MockEAC {
	return &MockEAC{
		Results: make(map[string]*Result),
		Default: &Result{ExitCode: 0},
	}
}

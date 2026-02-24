package ai

// Option is a functional option for configuring AI provider execution.
type Option func(*ExecuteOptions)

// ExecuteOptions holds optional parameters for AI execution.
type ExecuteOptions struct {
	Model       string  // Model to use (e.g., "haiku", "sonnet", "gpt-4")
	Temperature float64 // Randomness (0.0 - 1.0), default 0.3
	MaxTokens   int     // Max response length, default 4000
	Debug       bool    // Include debug logs in output, default false
}

// DefaultExecuteOptions returns ExecuteOptions with sensible defaults.
func DefaultExecuteOptions() *ExecuteOptions {
	return &ExecuteOptions{Temperature: 0.3, MaxTokens: 4000}
}

// ApplyOptions applies functional options to ExecuteOptions with defaults.
func ApplyOptions(opts ...Option) *ExecuteOptions {
	options := DefaultExecuteOptions()
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// WithModel sets the AI model to use for execution.
func WithModel(model string) Option {
	return func(opts *ExecuteOptions) {
		opts.Model = model
	}
}

// WithTemperature sets the randomness/creativity level (0.0 = deterministic, 1.0 = creative).
func WithTemperature(temp float64) Option {
	return func(opts *ExecuteOptions) {
		opts.Temperature = temp
	}
}

// WithMaxTokens sets the maximum number of tokens in the response.
func WithMaxTokens(max int) Option {
	return func(opts *ExecuteOptions) {
		opts.MaxTokens = max
	}
}

// WithDebug enables debug logging in the output.
// When enabled, execution logs are included in the prompt output.
// When disabled (default), no logs are recorded or stored.
func WithDebug(debug bool) Option {
	return func(opts *ExecuteOptions) {
		opts.Debug = debug
	}
}

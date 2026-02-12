package display

// TUIBootstrap holds the factory functions for creating TUI components.
// This enables dependency injection from adapters/tui into clibase/cmdframework
// without clibase needing to import the adapter package.
type TUIBootstrap struct {
	// NewForCommand creates a Console for the given command path and config.
	NewForCommand func(commandPath string, config Config) Console

	// NewObserver creates an observer that translates execution events to TUI ops.
	// The observer must implement both interfaces.ExecutionObserver and interfaces.WriterFactory.
	NewObserver func(console Console) (observer interface{}, writerFactory interface{})

	// NewHooks creates TUI hooks for controlled interaction.
	// Returns an interfaces.TUIHooks implementation.
	NewHooks func(console Console) interface{}

	// UnwrapConsole extracts the inner Console from a wrapper (e.g., parallel.Console).
	// This is needed because some Console implementations wrap another Console.
	// Returns the input console if no unwrapping is needed.
	UnwrapConsole func(console Console) Console
}

// global TUI bootstrap - set by adapters/tui at init time.
var globalBootstrap *TUIBootstrap

// SetTUIBootstrap registers the TUI factory functions.
// Called by adapters/tui during initialization.
func SetTUIBootstrap(b *TUIBootstrap) {
	globalBootstrap = b
}

// GetTUIBootstrap returns the registered TUI bootstrap, or nil if none set.
func GetTUIBootstrap() *TUIBootstrap {
	return globalBootstrap
}

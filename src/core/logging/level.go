package logging

import "go.uber.org/zap/zapcore"

// consoleEnabler implements zapcore.LevelEnabler for console output.
// In default mode: Info, Warn, Error only (no Debug)
// In debug mode: all levels including Debug
type consoleEnabler struct {
	debugMode bool
}

// Enabled returns true if the level should be logged to console.
func (e *consoleEnabler) Enabled(lvl zapcore.Level) bool {
	if e.debugMode {
		// Debug mode: show all levels on console
		return lvl >= zapcore.DebugLevel
	}
	// Default: show Info and above on console (hide Debug)
	return lvl >= zapcore.InfoLevel
}

// fileEnabler implements zapcore.LevelEnabler for file output.
// Always enables all levels (Debug and above) for file.
type fileEnabler struct{}

// Enabled returns true if the level should be logged to file.
// All levels are logged to file.
func (e *fileEnabler) Enabled(lvl zapcore.Level) bool {
	return lvl >= zapcore.DebugLevel
}

// newConsoleEnabler creates a level enabler for console output.
func newConsoleEnabler(debugMode bool) zapcore.LevelEnabler {
	return &consoleEnabler{debugMode: debugMode}
}

// newFileEnabler creates a level enabler for file output.
func newFileEnabler() zapcore.LevelEnabler {
	return &fileEnabler{}
}

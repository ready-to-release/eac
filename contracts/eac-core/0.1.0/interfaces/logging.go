package interfaces

// LoggerPort defines the logging contract.
// Commands use this interface for all logging operations.
type LoggerPort interface {
	// Basic logging
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)

	// Formatted logging
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})

	// Child logger with suffix (e.g., "git" -> "git.StagedFiles")
	WithSuffix(suffix string) LoggerPort

	// Sync flushes buffered logs
	Sync() error
}

// LoggerFactoryPort creates loggers.
type LoggerFactoryPort interface {
	// NewLogger creates a logger for the given component.
	NewLogger(component string) LoggerPort

	// NewModuleLogger creates a logger for module-specific logging.
	NewModuleLogger(command, workspaceRoot, module string) (LoggerPort, error)
}

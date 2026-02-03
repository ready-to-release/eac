package interfaces

// SimpleServicesPort provides core services for any command type.
// This is the minimal set of dependencies that all commands need.
// Orchestrated commands (build, test, scan, lint) get additional
// services through cmdframework.
type SimpleServicesPort interface {
	// WorkspaceRoot returns the repository root path.
	WorkspaceRoot() string

	// ConfigRoot returns the .eac configuration directory path.
	ConfigRoot() string

	// Config returns the configuration access port.
	Config() ConfigPort

	// Modules returns the module registry port.
	Modules() ModuleRegistryPort

	// Tools returns the tool registry port.
	Tools() ToolRegistryPort

	// Close cleans up any resources held by the services.
	Close() error
}

// SimpleServicesOptions configures which services to initialize.
type SimpleServicesOptions struct {
	// InitTools loads tool-config.yml and initializes tool bridges.
	// Default: false (Go zero-value)
	InitTools bool

	// DebugMode enables debug logging.
	// Default: false (Go zero-value)
	DebugMode bool
}

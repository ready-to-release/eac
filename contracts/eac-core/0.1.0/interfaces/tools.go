package interfaces

// ToolRegistryPort provides access to tool definitions.
type ToolRegistryPort interface {
	// Get retrieves a tool by canonical name.
	Get(canonicalName string) (ToolDefinitionPort, bool)

	// GetForComponent returns tools assigned to a component type for an operation.
	GetForComponent(componentType string, operation OperationType) []ToolDefinitionPort

	// Has checks if a tool exists in the registry.
	Has(canonicalName string) bool

	// AllCanonical returns all canonical tool names.
	AllCanonical() []string
}

// ToolDefinitionPort provides access to a single tool's definition.
type ToolDefinitionPort interface {
	// Identity
	GetID() string
	GetCanonicalID() string
	GetDescription() string
	GetType() ToolType

	// Execution
	GetBinary() string       // For system tools
	GetImage() string        // For container tools
	GetCommand() []string    // Container command
	GetEntrypoint() []string // Container entrypoint
	GetEnv() map[string]string
	GetWorkDir() string

	// Resources
	GetResources() ToolResourcesPort

	// Verification
	GetVerify() ToolVerifyPort

	// Display
	DisplayName() string
	ShortName() string
}

// ToolResourcesPort provides resource requirements for scheduling.
type ToolResourcesPort interface {
	// Weight returns the scheduling weight (typically CPU count).
	Weight() int

	// CPUs returns the CPU allocation.
	GetCPUs() int

	// Memory returns the memory allocation string.
	GetMemory() string
}

// ToolVerifyPort provides verification configuration.
type ToolVerifyPort interface {
	// IsCommandBased returns true if verification uses a command.
	IsCommandBased() bool

	// IsEnvBased returns true if verification uses environment variables.
	IsEnvBased() bool

	// GetCommand returns the verification command.
	GetCommand() string

	// GetPattern returns the regex pattern for version extraction.
	GetPattern() string

	// GetEnvVars returns required environment variables.
	GetEnvVars() []string
}

// ToolAssignmentPort maps operations to tools for a component type.
type ToolAssignmentPort interface {
	// GetToolID returns the primary tool ID for an operation.
	GetToolID(op OperationType) string

	// GetToolIDs returns all tool IDs for an operation.
	GetToolIDs(op OperationType) []string
}

// ToolType represents the execution type of a tool.
type ToolType string

const (
	// ToolTypeSystem represents a tool executed as a local system binary.
	ToolTypeSystem ToolType = "system"

	// ToolTypeContainer represents a tool executed in a Docker container.
	ToolTypeContainer ToolType = "container"
)

// OperationType represents the type of operation.
type OperationType string

const (
	OperationBuild OperationType = "build"
	OperationLint  OperationType = "lint"
	OperationScan  OperationType = "scan"
	OperationTest  OperationType = "test"
	OperationServe OperationType = "serve"
)

// ExecutionContextPort provides execution parameters for tool execution.
type ExecutionContextPort interface {
	// Paths
	GetWorkspaceRoot() string
	GetModuleRoot() string
	GetOutputDir() string

	// Input context
	GetFiles() []string
	GetOperation() OperationType

	// Runtime overrides
	GetEnvOverrides() map[string]string
	GetArgsOverrides() []string

	// Placeholders for mount path resolution
	GetPlaceholders() map[string]string

	// Resource amplifier
	GetAmp() float64
}

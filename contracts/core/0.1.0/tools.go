package core

// ToolRegistryPort provides access to tool definitions.
type ToolRegistryPort interface {
	// Get retrieves a tool by canonical name.
	Get(canonicalName string) (ToolDefinitionPort, bool)

	// GetForComponent returns tools assigned to a component type for an operation.
	GetForComponent(componentType string, operation ActionType) []ToolDefinitionPort

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
	GetToolID(op ActionType) string

	// GetToolIDs returns all tool IDs for an operation.
	GetToolIDs(op ActionType) []string
}

// ToolType represents the execution type of a tool.
type ToolType string

const (
	// ToolTypeSystem represents a tool executed as a local system binary.
	ToolTypeSystem ToolType = "system"

	// ToolTypeContainer represents a tool executed in a Docker container.
	ToolTypeContainer ToolType = "container"
)


// ExecutionContextPort provides execution parameters for tool execution.
type ExecutionContextPort interface {
	// Paths
	GetWorkspaceRoot() string
	GetModuleRoot() string
	GetOutputDir() string

	// Input context
	GetFiles() []string
	GetOperation() ActionType

	// Runtime overrides
	GetEnvOverrides() map[string]string
	GetArgsOverrides() []string

	// Placeholders for mount path resolution
	GetPlaceholders() map[string]string

	// Resource amplifier
	GetAmp() float64
}

// ============================================================================
// Tool config interfaces (absorbed from contracts/tools/0.1.0/interfaces)
// ============================================================================

// ToolConfigPort provides access to tool configuration with namespace-based loading.
// Tools are organized into namespaces (bootstrap, go, docs, node) and loaded on-demand.
type ToolConfigPort interface {
	// GetTool returns a tool definition by ID.
	// Returns false if the tool is not found.
	GetTool(id string) (ToolDefPort, bool)

	// ListTools returns all available tool IDs.
	ListTools() []string

	// ListToolsByNamespace returns tool IDs within a specific namespace.
	ListToolsByNamespace(namespace Namespace) []string

	// GetBinding returns the binding mode for a tool.
	// Returns BindingAuto if not explicitly configured.
	GetBinding(toolID string) Binding

	// GetComponentTools returns tool assignments for a component type.
	GetComponentTools(componentType string) (ToolConfigAssignmentPort, bool)

	// EnsureNamespace loads tools in the given namespace if not already loaded.
	// Returns an error if the namespace cannot be loaded.
	EnsureNamespace(ns Namespace) error

	// IsNamespaceLoaded returns true if the namespace has been loaded.
	IsNamespaceLoaded(ns Namespace) bool
}

// ToolDefPort defines a tool that can be executed.
// Tools can be system binaries or Docker containers.
type ToolDefPort interface {
	// ID returns the unique tool identifier (e.g., "golangci-lint").
	ID() string

	// Description returns a human-readable description.
	Description() string

	// Type returns the tool type (system or container).
	Type() ToolType

	// IsSystem returns true if this is a system binary tool.
	IsSystem() bool

	// IsContainer returns true if this is a Docker container tool.
	IsContainer() bool

	// Binary returns the binary name for system tools.
	Binary() string

	// Args returns default arguments for system tools.
	Args() []string

	// Image returns the Docker image for container tools.
	Image() string

	// Tag returns the Docker image tag for container tools.
	Tag() string

	// FullImage returns the complete image reference (image:tag).
	FullImage() string

	// LocalPath returns the local build context path for local containers.
	LocalPath() string

	// IsLocalContainer returns true if this is a locally-built container.
	IsLocalContainer() bool

	// Command returns the container command.
	Command() []string

	// Entrypoint returns the container entrypoint override.
	Entrypoint() []string

	// Env returns environment variables.
	Env() map[string]string

	// WorkDir returns the working directory.
	WorkDir() string

	// Mounts returns volume mount configurations.
	Mounts() []MountDef

	// Resources returns resource requirements.
	Resources() ResourceDef

	// Requirements returns tool dependencies.
	Requirements() []string

	// Platforms returns supported platforms (empty means all).
	Platforms() []string

	// Verify returns verification configuration.
	Verify() VerifyDef

	// Version returns the required version constraint.
	Version() string

	// Serve returns serve-specific configuration (if servable).
	Serve() ServeDef
}

// ToolConfigAssignmentPort maps operations to tools for a component type.
// This is the tool-config specific assignment port (from tools.yml).
type ToolConfigAssignmentPort interface {
	// Builder returns the builder tool ID.
	Builder() string

	// Linter returns the primary linter tool ID.
	Linter() string

	// Linters returns all linter tool IDs.
	Linters() []string

	// Scanner returns the primary scanner tool ID.
	Scanner() string

	// Scanners returns all scanner tool IDs.
	Scanners() []string

	// Tester returns the tester tool ID.
	Tester() string

	// Server returns the primary server tool ID.
	Server() string

	// Servers returns all server tool IDs.
	Servers() []string
}

// Namespace represents a group of related tools loaded together.
type Namespace string

const (
	// NSBootstrap contains essential tools always loaded at startup (go, docker, git).
	NSBootstrap Namespace = "bootstrap"

	// NSGo contains Go development tools (golangci-lint, gcc, go-build, go-race).
	NSGo Namespace = "go"

	// NSDocs contains documentation tools (mkdocs, drawio, structurizr).
	NSDocs Namespace = "docs"

	// NSNode contains Node.js development tools (npm, tsc, npm-build).
	NSNode Namespace = "node"

	// NSPython contains Python development tools (python, ruff, pytest).
	NSPython Namespace = "python"

	// NSSecurity contains security scanning tools (trivy, semgrep, zap).
	NSSecurity Namespace = "security"
)

// Binding specifies how to resolve a tool (system vs container).
type Binding string

const (
	// BindingAuto tries system first, falls back to container.
	BindingAuto Binding = "auto"

	// BindingSystem forces system tool only.
	BindingSystem Binding = "system"

	// BindingContainer forces container tool only.
	BindingContainer Binding = "container"
)

// MountDef describes a volume mount for container tools.
type MountDef struct {
	// Source is the host path or placeholder (e.g., "{workspace}", "{module}").
	Source string

	// Target is the container path.
	Target string

	// ReadOnly indicates if the mount should be read-only.
	ReadOnly bool

	// Type is the mount type: "bind" (default), "volume", "tmpfs".
	Type string
}

// ResourceDef defines resource requirements for a tool.
type ResourceDef struct {
	// CPUs is the number of CPUs (used as scheduling weight).
	CPUs int

	// Memory is the memory limit (e.g., "4g", "512m").
	Memory string

	// ShmSize is shared memory size for Chromium-based tools (e.g., "2g").
	ShmSize string
}

// VerifyDef defines how to verify a tool is available.
type VerifyDef struct {
	// Command is the command to run for verification (e.g., "go version").
	Command string

	// Pattern is the regex to extract version from output.
	Pattern string

	// EnvVars are required environment variables.
	EnvVars []string

	// Require is "any" or "all" for env var checking.
	Require string
}

// ServeDef defines serve-specific configuration for development servers.
type ServeDef struct {
	// ContainerPort is the port inside the container.
	ContainerPort int

	// HostPortRange is the range of ports to try on host (e.g., "9000-9999").
	HostPortRange string

	// HostPortFixed is a fixed host port (0 = auto-assign).
	HostPortFixed int

	// WatchEnabled enables file watching for live reload.
	WatchEnabled bool

	// WatchPaths are paths to watch for changes.
	WatchPaths []string

	// RestartPolicy is "unless-stopped", "always", or "no".
	RestartPolicy string

	// AutoOpenBrowser opens browser on start.
	AutoOpenBrowser bool

	// BrowserPath is the URL path to open (default: "/").
	BrowserPath string
}

// ComponentToNamespace maps component types to their tool namespaces.
func ComponentToNamespace(componentType string) Namespace {
	switch componentType {
	case "go", "go-feature", "godog":
		return NSGo
	case "node", "typescript", "javascript":
		return NSNode
	case "python":
		return NSPython
	case "book", "base-site", "pdf-render", "structurizr", "design":
		return NSDocs
	default:
		return NSBootstrap
	}
}

package core

// ToolsConfig holds the complete tool configuration from tools.yml.
type ToolsConfig struct {
	// Namespaces maps namespace names to tool IDs.
	Namespaces map[Namespace][]string `yaml:"namespaces"`

	// SystemTools maps tool ID to system tool definition.
	SystemTools map[string]*ToolDefinition `yaml:"system-tools"`

	// ContainerTools maps tool ID to container tool definition.
	ContainerTools map[string]*ToolDefinition `yaml:"container-tools"`

	// Bindings maps tool ID to binding mode.
	Bindings map[string]Binding `yaml:"tool-bindings"`

	// ComponentTools maps component type to tool assignments.
	ComponentTools map[string]*ToolAssignment `yaml:"component-tools"`

	// Environments maps environment name to overrides.
	Environments map[string]*EnvironmentConfig `yaml:"environments"`

	// Caches maps cache name to configuration.
	Caches map[string]*CacheConfig `yaml:"caches"`
}

// ToolDefinition defines a tool that can be executed.
// This is the concrete implementation of ToolDefPort.
type ToolDefinition struct {
	// IDValue is the unique tool identifier.
	IDValue string `yaml:"id,omitempty"`

	// DescriptionValue is a human-readable description.
	DescriptionValue string `yaml:"description,omitempty"`

	// TypeValue is the tool type (system or container).
	TypeValue ToolType `yaml:"type,omitempty"`

	// System tool configuration
	BinaryValue string   `yaml:"binary,omitempty"`
	ArgsValue   []string `yaml:"args,omitempty"`

	// Container tool configuration
	ImageValue     string `yaml:"image,omitempty"`
	TagValue       string `yaml:"tag,omitempty"`
	LocalPathValue string `yaml:"localPath,omitempty"`
	ContainerValue string `yaml:"container,omitempty"` // Legacy reference

	// Container execution
	CommandValue    []string `yaml:"command,omitempty"`
	EntrypointValue []string `yaml:"entrypoint,omitempty"`

	// Execution environment
	EnvValue     map[string]string `yaml:"env,omitempty"`
	WorkDirValue string            `yaml:"workdir,omitempty"`

	// Container settings
	MountsValue     []MountConfig   `yaml:"mounts,omitempty"`
	NetworkValue    string          `yaml:"network,omitempty"`
	UserValue       string          `yaml:"user,omitempty"`
	PrivilegedValue bool            `yaml:"privileged,omitempty"`
	ResourcesValue  *ResourceConfig `yaml:"resources,omitempty"`

	// Tool dependencies and constraints
	RequirementsValue []string `yaml:"requirements,omitempty"`
	PlatformsValue    []string `yaml:"platforms,omitempty"`

	// Verification
	VerifyValue  *VerifyConfig `yaml:"verify,omitempty"`
	VersionValue string        `yaml:"version,omitempty"`

	// Output handling
	OutputFormatValue string `yaml:"output_format,omitempty"`

	// Serve configuration
	ServeValue *ServeConfig `yaml:"serve,omitempty"`
}

// ID returns the tool identifier.
func (t *ToolDefinition) ID() string { return t.IDValue }

// Description returns the tool description.
func (t *ToolDefinition) Description() string { return t.DescriptionValue }

// Type returns the tool type.
func (t *ToolDefinition) Type() ToolType { return t.TypeValue }

// IsSystem returns true if this is a system binary tool.
func (t *ToolDefinition) IsSystem() bool { return t.TypeValue == ToolTypeSystem }

// IsContainer returns true if this is a container tool.
func (t *ToolDefinition) IsContainer() bool { return t.TypeValue == ToolTypeContainer }

// Binary returns the binary name.
func (t *ToolDefinition) Binary() string { return t.BinaryValue }

// Args returns default arguments.
func (t *ToolDefinition) Args() []string { return t.ArgsValue }

// Image returns the Docker image.
func (t *ToolDefinition) Image() string { return t.ImageValue }

// Tag returns the Docker image tag.
func (t *ToolDefinition) Tag() string { return t.TagValue }

// FullImage returns the complete image reference.
func (t *ToolDefinition) FullImage() string {
	if t.ImageValue == "" {
		return ""
	}
	if t.TagValue == "" {
		return t.ImageValue + ":latest"
	}
	return t.ImageValue + ":" + t.TagValue
}

// LocalPath returns the local build context path.
func (t *ToolDefinition) LocalPath() string { return t.LocalPathValue }

// IsLocalContainer returns true if this is a locally-built container.
func (t *ToolDefinition) IsLocalContainer() bool {
	return t.TypeValue == ToolTypeContainer && t.LocalPathValue != ""
}

// Command returns the container command.
func (t *ToolDefinition) Command() []string { return t.CommandValue }

// Entrypoint returns the container entrypoint.
func (t *ToolDefinition) Entrypoint() []string { return t.EntrypointValue }

// Env returns environment variables.
func (t *ToolDefinition) Env() map[string]string { return t.EnvValue }

// WorkDir returns the working directory.
func (t *ToolDefinition) WorkDir() string { return t.WorkDirValue }

// Mounts returns volume mount configurations.
func (t *ToolDefinition) Mounts() []MountDef {
	if t.MountsValue == nil {
		return nil
	}
	result := make([]MountDef, len(t.MountsValue))
	for i, m := range t.MountsValue {
		result[i] = MountDef{
			Source:   m.Source,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
			Type:     m.Type,
		}
	}
	return result
}

// Resources returns resource requirements.
func (t *ToolDefinition) Resources() ResourceDef {
	if t.ResourcesValue == nil {
		return ResourceDef{}
	}
	return ResourceDef{
		CPUs:    t.ResourcesValue.CPUs,
		Memory:  t.ResourcesValue.Memory,
		ShmSize: t.ResourcesValue.ShmSize,
	}
}

// Requirements returns tool dependencies.
func (t *ToolDefinition) Requirements() []string { return t.RequirementsValue }

// Platforms returns supported platforms.
func (t *ToolDefinition) Platforms() []string { return t.PlatformsValue }

// Verify returns verification configuration.
func (t *ToolDefinition) Verify() VerifyDef {
	if t.VerifyValue == nil {
		return VerifyDef{}
	}
	return VerifyDef{
		Command: t.VerifyValue.Command,
		Pattern: t.VerifyValue.Pattern,
		EnvVars: t.VerifyValue.EnvVars,
		Require: t.VerifyValue.Require,
	}
}

// Version returns the required version constraint.
func (t *ToolDefinition) Version() string { return t.VersionValue }

// Serve returns serve configuration.
func (t *ToolDefinition) Serve() ServeDef {
	if t.ServeValue == nil {
		return ServeDef{}
	}
	return ServeDef{
		ContainerPort:   t.ServeValue.ContainerPort,
		HostPortRange:   t.ServeValue.HostPortRange,
		HostPortFixed:   t.ServeValue.HostPortFixed,
		WatchEnabled:    t.ServeValue.WatchEnabled,
		WatchPaths:      t.ServeValue.WatchPaths,
		RestartPolicy:   t.ServeValue.RestartPolicy,
		AutoOpenBrowser: t.ServeValue.AutoOpenBrowser,
		BrowserPath:     t.ServeValue.BrowserPath,
	}
}

// MountConfig describes a volume mount configuration.
type MountConfig struct {
	Source   string `yaml:"source"`
	Target   string `yaml:"target"`
	ReadOnly bool   `yaml:"readonly,omitempty"`
	Type     string `yaml:"type,omitempty"`
}

// ResourceConfig defines resource requirements.
type ResourceConfig struct {
	CPUs    int    `yaml:"cpus,omitempty"`
	Memory  string `yaml:"memory,omitempty"`
	ShmSize string `yaml:"shm_size,omitempty"`
}

// VerifyConfig defines verification configuration.
type VerifyConfig struct {
	Command string   `yaml:"command,omitempty"`
	Pattern string   `yaml:"pattern,omitempty"`
	EnvVars []string `yaml:"env_vars,omitempty"`
	Require string   `yaml:"require,omitempty"`
}

// ServeConfig defines serve-specific configuration.
type ServeConfig struct {
	ContainerPort   int      `yaml:"container_port,omitempty"`
	HostPortRange   string   `yaml:"host_port_range,omitempty"`
	HostPortFixed   int      `yaml:"host_port_fixed,omitempty"`
	WatchEnabled    bool     `yaml:"watch_enabled,omitempty"`
	WatchPaths      []string `yaml:"watch_paths,omitempty"`
	WatchExclude    []string `yaml:"watch_exclude,omitempty"`
	RestartPolicy   string   `yaml:"restart_policy,omitempty"`
	AutoOpenBrowser bool     `yaml:"auto_open_browser,omitempty"`
	BrowserPath     string   `yaml:"browser_path,omitempty"`
}

// ToolAssignment maps operations to tools for a component type.
// This is the concrete implementation of ToolConfigAssignmentPort.
type ToolAssignment struct {
	BuilderValue  string   `yaml:"builder,omitempty"`
	LinterValue   string   `yaml:"linter,omitempty"`
	LintersValue  []string `yaml:"linters,omitempty"`
	ScannerValue  string   `yaml:"scanner,omitempty"`
	ScannersValue []string `yaml:"scanners,omitempty"`
	TesterValue   string   `yaml:"tester,omitempty"`
	ServerValue   string   `yaml:"server,omitempty"`
	ServersValue  []string `yaml:"servers,omitempty"`
}

// Builder returns the builder tool ID.
func (a *ToolAssignment) Builder() string { return a.BuilderValue }

// Linter returns the primary linter tool ID.
func (a *ToolAssignment) Linter() string { return a.LinterValue }

// Linters returns all linter tool IDs.
func (a *ToolAssignment) Linters() []string {
	if len(a.LintersValue) > 0 {
		return a.LintersValue
	}
	if a.LinterValue != "" {
		return []string{a.LinterValue}
	}
	return nil
}

// Scanner returns the primary scanner tool ID.
func (a *ToolAssignment) Scanner() string { return a.ScannerValue }

// Scanners returns all scanner tool IDs.
func (a *ToolAssignment) Scanners() []string {
	if len(a.ScannersValue) > 0 {
		return a.ScannersValue
	}
	if a.ScannerValue != "" {
		return []string{a.ScannerValue}
	}
	return nil
}

// Tester returns the tester tool ID.
func (a *ToolAssignment) Tester() string { return a.TesterValue }

// Server returns the primary server tool ID.
func (a *ToolAssignment) Server() string { return a.ServerValue }

// Servers returns all server tool IDs.
func (a *ToolAssignment) Servers() []string {
	if len(a.ServersValue) > 0 {
		return a.ServersValue
	}
	if a.ServerValue != "" {
		return []string{a.ServerValue}
	}
	return nil
}

// EnvironmentConfig defines environment-specific tool overrides.
type EnvironmentConfig struct {
	Description    string                     `yaml:"description,omitempty"`
	ComponentTools map[string]*ToolAssignment `yaml:"component-tools,omitempty"`
}

// CacheConfig defines a cache directory configuration.
type CacheConfig struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description,omitempty"`
}

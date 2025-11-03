package contracts

// Contract represents a generic contract interface
type Contract interface {
	GetMoniker() string
	GetName() string
	GetType() string
	GetDescription() string
	GetRoot() string
	GetVersion() string
}

// Versioning represents version configuration
type Versioning struct {
	VersionScheme string `yaml:"version_scheme"` // e.g., "MAJOR.MINOR.PATCH"
}

// Source represents source file configuration
type Source struct {
	Root                       string   `yaml:"root"`                          // Root directory path
	Includes                   []string `yaml:"includes"`                      // Glob patterns for included files
	ChangelogPath              string   `yaml:"changelog_path"`                // Path to CHANGELOG.md
	ExcludeChildrenOwnedSource *bool    `yaml:"exclude_children_owned_source"` // Defer ownership to children in same source space (default: true)
}

// BaseContract contains common contract fields
type BaseContract struct {
	Moniker     string     `yaml:"moniker"`      // Unique identifier (e.g., "src-mcp-vscode")
	Name        string     `yaml:"name"`         // Human-readable name
	Type        string     `yaml:"type"`         // Contract type (e.g., "mcp-server", "cli")
	Description string     `yaml:"description"`  // Brief description
	Parent      string     `yaml:"parent"`       // Parent module moniker (defaults to "." for repository)
	Versioning  Versioning `yaml:"versioning"`   // Version configuration
	Source      Source     `yaml:"source"`       // Source file patterns
	DependsOn   []string   `yaml:"depends_on"`   // Dependencies (monikers)
	UsedBy      []string   `yaml:"used_by"`      // Reverse dependencies (monikers)
}

// GetMoniker returns the contract moniker
func (c *BaseContract) GetMoniker() string {
	return c.Moniker
}

// GetName returns the contract name
func (c *BaseContract) GetName() string {
	return c.Name
}

// GetType returns the contract type
func (c *BaseContract) GetType() string {
	return c.Type
}

// GetDescription returns the contract description
func (c *BaseContract) GetDescription() string {
	return c.Description
}

// GetRoot returns the contract root path
func (c *BaseContract) GetRoot() string {
	return c.Source.Root
}

// GetVersion returns the version scheme (placeholder for future use)
func (c *BaseContract) GetVersion() string {
	return c.Versioning.VersionScheme
}

// GetParent returns the parent module moniker
func (c *BaseContract) GetParent() string {
	return c.Parent
}

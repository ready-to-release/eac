package contracts

// BaseContract represents the base structure for module contracts
type BaseContract struct {
	Moniker     string            `yaml:"moniker"`
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	Description string            `yaml:"description"`
	DependsOn   []string          `yaml:"depends_on"`
	Files       Files             `yaml:"files"`
	Flags       Flags             `yaml:"flags"`
	Metadata    map[string]string `yaml:"metadata,omitempty"` // Generic key-value store for module-specific data
}

// Files represents all file ownership patterns for a module
type Files struct {
	Root string `yaml:"root"` // Base directory for this module

	// Patterns relative to Root
	Source    []string `yaml:"source"`    // Primary source code files
	Config    []string `yaml:"config"`    // Configuration/build files
	Assets    []string `yaml:"assets"`    // Static assets, templates, docs
	Tests     []string `yaml:"tests"`     // Test files
	Exclude   []string `yaml:"exclude"`   // Patterns to exclude
	Changelog string   `yaml:"changelog"` // Changelog file path

	// Patterns relative to repo root
	Repo RepoPatterns `yaml:"repo"`
}

// RepoPatterns represents patterns relative to repository root
type RepoPatterns struct {
	Specs    []string `yaml:"specs"`     // Specification files
	TestImpl string   `yaml:"test_impl"` // Test implementation directory path
	Design   string   `yaml:"design"`    // Design workspace directory path
	Other    []string `yaml:"other"`     // Other files outside module root
	Exclude  []string `yaml:"exclude"`   // Exclusions relative to repo root
}

// Flags represents behavioral flags for a module (reserved for future use)
type Flags struct {
}

// Getter methods for BaseContract

func (b *BaseContract) GetMoniker() string {
	return b.Moniker
}

func (b *BaseContract) GetName() string {
	return b.Name
}

func (b *BaseContract) GetType() string {
	return b.Type
}

func (b *BaseContract) GetDescription() string {
	return b.Description
}

func (b *BaseContract) GetRoot() string {
	return b.Files.Root
}

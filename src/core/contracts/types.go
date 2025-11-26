package contracts

// BaseContract represents the base structure for module contracts
type BaseContract struct {
	Moniker     string     `yaml:"moniker"`
	Name        string     `yaml:"name"`
	Type        string     `yaml:"type"`
	Description string     `yaml:"description"`
	Parent      string     `yaml:"parent"`
	Source      Source     `yaml:"source"`
	Tests       *Tests     `yaml:"tests,omitempty"` // Optional: test directory and patterns
	Versioning  Versioning `yaml:"versioning"`
	DependsOn   []string   `yaml:"depends_on"`
	UsedBy      []string   `yaml:"used_by"`
}

// Tests represents the test configuration for a module
type Tests struct {
	Root     string   `yaml:"root"`     // Test directory root
	Includes []string `yaml:"includes"` // Test file patterns
}

// Source represents the source configuration for a module
type Source struct {
	Root                      string   `yaml:"root"`
	ChangelogPath             string   `yaml:"changelog_path"`
	SpecsRoot                 string   `yaml:"specs_root"`                   // Optional: defaults to specs/<module-name>
	Includes                  []string `yaml:"includes"`
	Excludes                  []string `yaml:"excludes"`                     // Optional: patterns to exclude from includes
	IsCatchAllSingleton       *bool    `yaml:"is_catch_all_singleton"`
	ExcludeChildrenOwnedSource *bool   `yaml:"exclude_children_owned_source"`
}

// Versioning represents the versioning configuration for a module
type Versioning struct {
	VersionScheme string `yaml:"version_scheme"`
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

func (b *BaseContract) GetParent() string {
	return b.Parent
}

func (b *BaseContract) GetRoot() string {
	return b.Source.Root
}

func (b *BaseContract) GetVersion() string {
	return b.Versioning.VersionScheme
}

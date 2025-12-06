package contracts

// RepositoryContract represents repository-level configuration
type RepositoryContract struct {
	Repository RepositoryConfig `yaml:"repository"`
}

// RepositoryConfig contains repository-level settings
type RepositoryConfig struct {
	Type             string           `yaml:"type"`                // mono | poly | adjunct
	Schemes          []string         `yaml:"schemes"`             // Valid versioning schemes (SemVer, CalVer)
	TrunkBranch      string           `yaml:"trunk_branch"`        // Main branch name
	MaxBranchAgeDays int              `yaml:"max_branch_age_days"` // Maximum branch age in days
	PR               PRConfig         `yaml:"pr"`                  // Pull request configuration
	Versioning       VersioningConfig `yaml:"versioning"`          // Versioning constraints
}

// PRConfig contains pull request workflow configuration
type PRConfig struct {
	DeleteBranchOnMerge bool   `yaml:"delete_branch_on_merge"` // Auto-delete branch after merge
	MergeStrategy       string `yaml:"merge_strategy"`         // squash | merge | rebase
}

// VersioningConfig contains repository-wide versioning constraints
type VersioningConfig struct {
	Constraint string `yaml:"constraint"` // unrestricted | patch-only | calver-only
}

// ModulesContract represents the modules configuration file
type ModulesContract struct {
	Defaults *ModuleDefaults `yaml:"defaults,omitempty"` // Module-level defaults
	Modules  []BaseContract  `yaml:"modules"`            // Module definitions
}

// ModuleDefaults contains default values inherited by all modules
type ModuleDefaults struct {
	Paths       DefaultPaths       `yaml:"paths"`
	Conventions DefaultConventions `yaml:"conventions"`
}

// DefaultPaths contains default path configuration
type DefaultPaths struct {
	TestImplRoot string   `yaml:"test_impl_root"` // Default test implementation root
	SpecsRoot    string   `yaml:"specs_root"`     // Default specs root
	Templates    string   `yaml:"templates"`      // Default templates directory
	Out          OutPaths `yaml:"out"`            // Output directory structure
}

// OutPaths contains output directory structure
type OutPaths struct {
	Root     string `yaml:"root"`     // Root output directory
	Build    string `yaml:"build"`    // Build output directory
	Test     string `yaml:"test"`     // Test output directory
	Logs     string `yaml:"logs"`     // Logs output directory
	Security string `yaml:"security"` // Security scan output directory
	Tools    string `yaml:"tools"`    // Tools output directory
}

// DefaultConventions contains default filename conventions
type DefaultConventions struct {
	GodogTest   string `yaml:"godog_test"`   // Godog test file name
	PackageJSON string `yaml:"package_json"` // Node.js package file name
	Changelog   string `yaml:"changelog"`    // Changelog file name
}

// ModuleVersioning contains module versioning configuration
type ModuleVersioning struct {
	Scheme  string `yaml:"scheme"`            // SemVer | CalVer
	Current string `yaml:"current,omitempty"` // Current version (optional)
}

// BaseContract represents the base structure for module contracts
type BaseContract struct {
	Moniker     string            `yaml:"moniker"`
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	Description string            `yaml:"description"`
	DependsOn   []string          `yaml:"depends_on"`
	Versioning  *ModuleVersioning `yaml:"versioning,omitempty"` // Module versioning configuration
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

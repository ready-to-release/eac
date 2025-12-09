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
	Moniker     string                 `yaml:"moniker"`
	Name        string                 `yaml:"name"`
	Type        string                 `yaml:"type"`
	Description string                 `yaml:"description"`
	DependsOn   []string               `yaml:"depends_on"`
	Versioning  *ModuleVersioning      `yaml:"versioning,omitempty"`  // Module versioning configuration
	Build       *ModuleBuild           `yaml:"build,omitempty"`       // Per-module build configuration (artifacts, options)
	DockerBuild map[string]interface{} `yaml:"docker_build,omitempty"` // Per-module Docker build configuration
	Files       Files                  `yaml:"files"`
	Flags       Flags                  `yaml:"flags"`
	Metadata    map[string]string      `yaml:"metadata,omitempty"` // Generic key-value store for module-specific data
}

// ModuleBuild contains per-module build configuration
// This allows modules to define their own artifacts instead of relying on type-level defaults
type ModuleBuild struct {
	Handler   string           `yaml:"handler,omitempty"`   // Explicit build handler override (e.g., "mkdocs", "docker")
	Artifacts []ModuleArtifact `yaml:"artifacts,omitempty"` // Artifacts to produce
	Options   *BuildOptions    `yaml:"options,omitempty"`   // Build behavior options
}

// ModuleArtifact defines an artifact to be produced by a module build
type ModuleArtifact struct {
	ID          string `yaml:"id"`                    // Unique artifact identifier
	Type        string `yaml:"type"`                  // executable, file, directory, test
	Pattern     string `yaml:"pattern"`               // Output path pattern with variables: {moniker}, {ext}
	Compression string `yaml:"compression,omitempty"` // none, strip, upx
	DeriveFrom  string `yaml:"derive_from,omitempty"` // Source artifact to derive from (for compressed variants)
}

// BuildOptions contains optional build behavior flags
type BuildOptions struct {
	CommandsBinary bool `yaml:"commands_binary,omitempty"` // Copy to tools directory after build
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

	// Workflow file ownership (paths relative to repo root)
	Workflows Workflows `yaml:"workflows"`

	// Patterns relative to repo root
	Repo RepoPatterns `yaml:"repo"`
}

// Workflows defines GitHub Actions workflow file ownership
type Workflows struct {
	CI      string `yaml:"ci"`      // CI workflow file path
	Release string `yaml:"release"` // Release workflow file path
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

// HasBuildArtifacts returns true if the module has per-module build artifacts defined
func (b *BaseContract) HasBuildArtifacts() bool {
	return b.Build != nil && len(b.Build.Artifacts) > 0
}

// GetBuildArtifacts returns the per-module build artifacts, or nil if none defined
func (b *BaseContract) GetBuildArtifacts() []ModuleArtifact {
	if b.Build == nil {
		return nil
	}
	return b.Build.Artifacts
}

// HasExecutableArtifacts returns true if any artifacts are of type executable
func (b *BaseContract) HasExecutableArtifacts() bool {
	if b.Build == nil {
		return false
	}
	for _, a := range b.Build.Artifacts {
		if a.Type == "executable" {
			return true
		}
	}
	return false
}

// HasTestArtifacts returns true if any artifacts are of type test
func (b *BaseContract) HasTestArtifacts() bool {
	if b.Build == nil {
		return false
	}
	for _, a := range b.Build.Artifacts {
		if a.Type == "test" {
			return true
		}
	}
	return false
}

// IsToolsBinary returns true if this module should copy its binary to tools directory
func (b *BaseContract) IsToolsBinary() bool {
	return b.Build != nil && b.Build.Options != nil && b.Build.Options.CommandsBinary
}

// GetBuildHandler returns the explicit build handler for this module, or empty string if not set
func (b *BaseContract) GetBuildHandler() string {
	if b.Build == nil {
		return ""
	}
	return b.Build.Handler
}

// GetArtifactsByType returns all artifacts of the specified type
func (b *BaseContract) GetArtifactsByType(artifactType string) []ModuleArtifact {
	if b.Build == nil {
		return nil
	}
	var result []ModuleArtifact
	for _, a := range b.Build.Artifacts {
		if a.Type == artifactType {
			result = append(result, a)
		}
	}
	return result
}

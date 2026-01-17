// Package config provides a central configuration loader for all EAC repository configs.
// It consolidates loading of modules, environments, testing-tags, and test-suites
// with integrated JSON Schema validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/schema"
	"github.com/ready-to-release/eac/go/eac/core/paths"
	"gopkg.in/yaml.v3"
)

// EACConfigRelPath is re-exported for backwards compatibility.
const EACConfigRelPath = paths.EACConfigRelPath

// Config file names.
const (
	ModuleTypesFileName        = "module-types.yml"
	EnvironmentsFileName       = "environments.yml"
	TestingTagsFileName        = "testing-tags.yml"
	TestSuitesFileName         = "test-suites.yml"
	SystemDependenciesFileName = "system-dependencies.yml"
)

// EACConfig holds all loaded EAC repository configuration.
// Use Load() to create and populate this struct.
//
// # Field Guarantees (after successful Load)
//
// All config fields are guaranteed non-nil after a successful Load() call.
// The loader uses fail-fast for core configs and empty defaults for optional configs:
//
//   - Repository:         GUARANTEED non-nil (fails if cannot load defaults)
//   - ModuleTypes:        GUARANTEED non-nil (fails if cannot load defaults)
//   - SystemDependencies: GUARANTEED non-nil (loads defaults if file missing)
//   - SecurityTools:      GUARANTEED non-nil (uses defaults if file missing)
//   - Environments:       GUARANTEED non-nil (empty if file missing)
//   - TestingTags:        GUARANTEED non-nil (empty if file missing)
//   - TestSuites:         GUARANTEED non-nil (empty if file missing)
//   - Books:              GUARANTEED non-nil (empty if file missing)
//   - Commands:           GUARANTEED non-nil (uses defaults if file missing)
//
// # Module Access
//
// Modules are now part of Repository config. Access via cfg.Repository.Modules.
//
// # Sub-field Nil Checks Still Required
//
// While top-level config fields are guaranteed non-nil, nested fields may be nil:
//   - moduleType.Build may be nil (module type doesn't define artifacts)
//   - moduleType.DockerBuild may be nil (not a docker-building type)
//   - module.Metadata may be nil (no metadata defined)
type EACConfig struct {
	// Root paths
	RepoRoot   string
	ConfigRoot string

	// Core configs - fail-fast if loading fails (non-nil guaranteed)
	// Repository now includes modules (unified config from repository.yml)
	Repository  *RepositoryConfig
	ModuleTypes *ModuleTypesConfig

	// Optional configs - empty defaults if file missing (non-nil guaranteed)
	Environments       *EnvironmentsConfig
	TestingTags        *TestingTagsConfig
	TestSuites         *TestSuitesConfig
	SystemDependencies *SystemDependenciesConfig
	Books              *BooksConfig
	SecurityTools      *SecurityToolsConfig
	Commands           *CommandsConfig

	// Schema validator (lazy initialized)
	validator     *schema.Validator
	validatorOnce sync.Once
	validatorErr  error
}

// LoadOptions configures the loading behavior.
type LoadOptions struct {
	// RepoRoot overrides automatic repo root detection
	RepoRoot string
	// ValidateSchemas enables JSON Schema validation during load (default: true)
	ValidateSchemas bool
	// LazyLoad defers loading until first access (default: false)
	LazyLoad bool
	// SkipWorkflowValidation skips GitHub workflow file validation (default: false)
	// Useful for test environments or commands that don't need workflow information
	SkipWorkflowValidation bool
}

// DefaultLoadOptions returns the default load options.
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        false,
	}
}

// Load loads all EAC configuration from the repository.
// Uses global in-memory cache to avoid repeated file I/O and validation.
//
// BACKWARD COMPATIBLE: API unchanged, caching is internal optimization.
//
// Thread-safe: Multiple goroutines can call concurrently.
// First caller loads and validates, subsequent callers use cached data.
//
// It validates configs against JSON schemas if ValidateSchemas is true.
func Load(opts LoadOptions) (*EACConfig, error) {
	// Determine repo root
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = findRepositoryRoot("")
		if err != nil {
			return nil, fmt.Errorf("failed to find repository root: %w", err)
		}
	}

	// LazyLoad bypasses cache - return empty config immediately
	if opts.LazyLoad {
		configRoot := paths.EACConfigPath(repoRoot)
		return &EACConfig{
			RepoRoot:   repoRoot,
			ConfigRoot: configRoot,
		}, nil
	}

	// Initialize global cache once per process
	globalConfigCacheOnce.Do(func() {
		globalConfigCache = NewConfigCache()
	})

	// Check cache first
	if cachedCfg, ok := globalConfigCache.Get(repoRoot, opts.ValidateSchemas); ok {
		return cachedCfg, nil
	}

	// Cache miss - load and validate

	configRoot := paths.EACConfigPath(repoRoot)

	cfg := &EACConfig{
		RepoRoot:   repoRoot,
		ConfigRoot: configRoot,
	}

	// Load all configs (validation happens inside LoadAll)
	if err := cfg.LoadAll(opts); err != nil {
		return nil, err
	}

	// Cache the validated config
	globalConfigCache.Set(repoRoot, opts.ValidateSchemas, cfg)

	return cfg, nil
}

// LoadAll loads all configuration files.
// Core configs (Repository, ModuleTypes) must load successfully - fails immediately on error.
// Optional configs (Environments, TestingTags, etc.) collect errors but continue loading.
func (c *EACConfig) LoadAll(opts LoadOptions) error {
	validateSchemas := opts.ValidateSchemas
	// === CORE CONFIGS: Fail fast if any of these fail ===
	// These are required for the system to function correctly

	// Load repository config first (now includes modules)
	if err := c.LoadRepository(validateSchemas); err != nil {
		return fmt.Errorf("core config failed - repository: %w", err)
	}

	if err := c.LoadModuleTypes(validateSchemas); err != nil {
		return fmt.Errorf("core config failed - module-types: %w", err)
	}

	// Apply type-specific defaults after both modules and types are loaded
	// Skip if Repository is nil (can happen in test environments without contract files)
	if c.Repository != nil {
		c.Repository.ApplyTypeDefaults(c.ModuleTypes)

		// Validate defined workflow paths and auto-discover missing ones
		// Skip if explicitly disabled (useful for test environments)
		if !opts.SkipWorkflowValidation {
			if err := c.Repository.ValidateAndDiscoverWorkflows(c.RepoRoot); err != nil {
				return fmt.Errorf("workflow validation failed: %w", err)
			}
		}
	}

	// === OPTIONAL CONFIGS: Continue on error, collect for reporting ===
	var errs []error

	if err := c.LoadEnvironments(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("environments: %w", err))
	}

	if err := c.LoadTestingTags(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("testing-tags: %w", err))
	}

	if err := c.LoadTestSuites(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("test-suites: %w", err))
	}

	if err := c.LoadBooks(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("books: %w", err))
	}

	if err := c.LoadSystemDependencies(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("system-dependencies: %w", err))
	}

	if err := c.LoadSecurityTools(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("security-tools: %w", err))
	}

	if err := c.LoadCommands(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("commands: %w", err))
	}

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}

	return nil
}

// LoadRepository loads the repository-wide configuration.
// Merges contract defaults with type-specific defaults and user config.
//
// Merge order: base defaults → type-specific defaults → user config
// User values always win over all defaults.
func (c *EACConfig) LoadRepository(validateSchema bool) error {
	// Step 1: Load base defaults from contract
	defaults, err := LoadRepositoryDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading repository defaults: %w", err)
	}

	// Step 2: Peek at user config to get repository type
	// This determines which type-specific defaults to load
	repoType, err := peekRepositoryType(c.ConfigRoot)
	if err != nil {
		return fmt.Errorf("detecting repository type: %w", err)
	}

	// If no user-specified type, use type from base defaults
	if repoType == "" && defaults != nil {
		repoType = defaults.Repository.Type
	}

	// Step 3: Load type-specific defaults (if they exist)
	typeDefaults, err := LoadRepositoryTypeDefaults(c.RepoRoot, repoType)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading repository type defaults: %w", err)
	}

	// Step 4: Pre-merge base + type-specific defaults
	mergedDefaults := MergeRepository(defaults, typeDefaults)

	// Step 5: Check if user config exists
	repoPath := filepath.Join(c.ConfigRoot, RepositoryFileName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Use merged defaults only (may be nil in test environments without contract files)
		if mergedDefaults == nil {
			// Create empty config with hardcoded defaults to guarantee non-nil
			mergedDefaults = &RepositoryConfig{
				Paths: PathsConfig{
					SpecsRoot: "specs",
					Out: OutConfig{
						Root:  "out",
						Build: "out/build",
						Test:  "out/test",
						Logs:  "out/logs",
						Scan:  "out/scan",
						Tools: "out/tools",
					},
				},
			}
		}
		c.Repository = mergedDefaults
		return nil
	}

	// Step 6: Validate schema if requested
	if validateSchema {
		data, err := c.readConfigFile(RepositoryFileName)
		if err != nil {
			return err
		}
		if err := c.validateSchema(schema.SchemaRepository, data); err != nil {
			return err
		}
	}

	// Step 7: Load user config (unmerged)
	userCfg, err := loadRepositoryConfigUnmerged(c.RepoRoot)
	if err != nil {
		return err
	}

	// Step 8: Final merge: (base + type-specific) + user
	c.Repository = MergeRepository(mergedDefaults, userCfg)
	return nil
}

// LoadModuleTypes loads the module-types configuration.
// Merges contract defaults with user-defined types.
func (c *EACConfig) LoadModuleTypes(validateSchema bool) error {
	// Load defaults from contract
	defaults, err := LoadModuleTypesDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading module-types defaults: %w", err)
	}

	// Check if user config exists
	typesPath := filepath.Join(c.ConfigRoot, ModuleTypesFileName)
	if _, err := os.Stat(typesPath); os.IsNotExist(err) {
		// Use defaults only (may be nil in test environments without contract files)
		if defaults == nil {
			// Create empty config to guarantee non-nil
			defaults = &ModuleTypesConfig{}
		}
		c.ModuleTypes = defaults
		return nil
	}

	data, err := c.readConfigFile(ModuleTypesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaModuleTypes, data); err != nil {
			return err
		}
	}

	var userCfg ModuleTypesConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", ModuleTypesFileName, err)
	}

	// Merge defaults with user config
	c.ModuleTypes = MergeModuleTypes(defaults, &userCfg)
	return nil
}

// LoadEnvironments loads the environments configuration with fallback to system defaults.
func (c *EACConfig) LoadEnvironments(validateSchema bool) error {
	// Try loading with fallback (user override → system default)
	data, err := c.readConfigFileWithFallback(EnvironmentsFileName)
	if err != nil {
		// If file not found in either location, use empty config
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			c.Environments = &EnvironmentsConfig{}
			return nil
		}
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaEnvironments, data); err != nil {
			return err
		}
	}

	var cfg EnvironmentsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", EnvironmentsFileName, err)
	}

	c.Environments = &cfg
	return nil
}

// LoadTestingTags loads the testing-tags configuration with fallback to system defaults.
func (c *EACConfig) LoadTestingTags(validateSchema bool) error {
	// Try loading with fallback (user override → system default)
	data, err := c.readConfigFileWithFallback(TestingTagsFileName)
	if err != nil {
		// If file not found in either location, initialize empty config
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			c.TestingTags = &TestingTagsConfig{}
			return nil
		}
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaTestingTags, data); err != nil {
			return err
		}
	}

	var cfg TestingTagsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", TestingTagsFileName, err)
	}

	if err := cfg.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize testing-tags: %w", err)
	}

	c.TestingTags = &cfg
	return nil
}

// LoadTestSuites loads the test-suites configuration.
// Merges contract defaults with user config.
func (c *EACConfig) LoadTestSuites(validateSchema bool) error {
	// Load defaults from contract
	defaults, err := LoadTestSuitesDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading test-suites defaults: %w", err)
	}

	// Check if user config exists
	suitesPath := filepath.Join(c.ConfigRoot, TestSuitesFileName)
	if _, err := os.Stat(suitesPath); os.IsNotExist(err) {
		// Use defaults only (may be nil in test environments without contract files)
		if defaults == nil {
			// Create empty config to guarantee non-nil
			defaults = &TestSuitesConfig{}
		}
		defaults.buildSuiteMap()
		c.TestSuites = defaults
		return nil
	}

	data, err := c.readConfigFile(TestSuitesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaTestSuites, data); err != nil {
			return err
		}
	}

	var userCfg TestSuitesConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", TestSuitesFileName, err)
	}

	// Merge defaults with user config
	c.TestSuites = MergeTestSuites(defaults, &userCfg)
	return nil
}

// LoadSystemDependencies loads the system-dependencies configuration.
// Merges contract defaults with user config.
func (c *EACConfig) LoadSystemDependencies(validateSchema bool) error {
	// Load defaults from contract
	defaults, err := LoadSystemDependenciesDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading system-dependencies defaults: %w", err)
	}

	// Check if user config exists
	depsPath := filepath.Join(c.ConfigRoot, SystemDependenciesFileName)
	if _, err := os.Stat(depsPath); os.IsNotExist(err) {
		// Use defaults only (may be nil in test environments without contract files)
		if defaults == nil {
			// Create empty config to guarantee non-nil
			defaults = &SystemDependenciesConfig{}
		}
		defaults.buildDepMap()
		c.SystemDependencies = defaults
		return nil
	}

	data, err := c.readConfigFile(SystemDependenciesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaSystemDependencies, data); err != nil {
			return err
		}
	}

	var userCfg SystemDependenciesConfig
	if err := yaml.Unmarshal(data, &userCfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", SystemDependenciesFileName, err)
	}

	// Merge defaults with user config
	c.SystemDependencies = MergeSystemDependencies(defaults, &userCfg)
	return nil
}

// LoadBooks loads the books configuration (optional - only if file exists).
func (c *EACConfig) LoadBooks(validateSchema bool) error {
	// Check if books file exists - it's optional
	booksPath := filepath.Join(c.ConfigRoot, BooksFileName)
	if _, err := os.Stat(booksPath); os.IsNotExist(err) {
		// Initialize empty config to guarantee non-nil
		c.Books = &BooksConfig{}
		return nil
	}

	data, err := c.readConfigFile(BooksFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaBooks, data); err != nil {
			return err
		}
	}

	var cfg BooksConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", BooksFileName, err)
	}

	c.Books = &cfg
	return nil
}

// LoadSecurityTools loads the security tools configuration with fallback to system defaults.
func (c *EACConfig) LoadSecurityTools(validateSchema bool) error {
	// Try loading with fallback (user override → system default → hardcoded defaults)
	data, err := c.readConfigFileWithFallback(SecurityToolsFileName)
	if err != nil {
		// If file not found in either location, use hardcoded defaults (backwards compatible)
		if os.IsNotExist(err) || strings.Contains(err.Error(), "not found") {
			cfg := DefaultSecurityToolsConfig()
			c.SecurityTools = &cfg
			return nil
		}
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaSecurityTools, data); err != nil {
			return err
		}
	}

	var cfg SecurityToolsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", SecurityToolsFileName, err)
	}

	c.SecurityTools = &cfg
	return nil
}

// LoadCommands loads the commands configuration (optional - uses defaults if file missing).
func (c *EACConfig) LoadCommands(validateSchema bool) error {
	// Check if commands file exists - it's optional
	commandsPath := filepath.Join(c.ConfigRoot, CommandsFileName)
	if _, err := os.Stat(commandsPath); os.IsNotExist(err) {
		// Use defaults if file doesn't exist
		c.Commands = DefaultCommandsConfig()
		if err := c.Commands.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize default commands config: %w", err)
		}
		return nil
	}

	data, err := c.readConfigFile(CommandsFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaCommands, data); err != nil {
			return err
		}
	}

	var cfg CommandsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", CommandsFileName, err)
	}

	// Initialize (compile regex patterns)
	if err := cfg.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize commands config: %w", err)
	}

	c.Commands = &cfg
	return nil
}

// GetBookByName finds a book by its name.
func (c *EACConfig) GetBookByName(name string) *Book {
	if c.Books == nil {
		return nil
	}
	return c.Books.GetBookByName(name)
}

// GetBooksByModule returns all books that belong to a module.
// Uses the module's books list to look up Book configs by name.
func (c *EACConfig) GetBooksByModule(moniker string) []*Book {
	if c.Books == nil || c.Repository == nil {
		return nil
	}
	// Find the module and get its books list
	module := c.Repository.GetByMoniker(moniker)
	if module == nil {
		return nil
	}
	return c.Books.GetBooksByNames(module.Books)
}

// GetDefaultBooksByModule returns only default books for a module.
// The first book in a module's books list is the default; others require --all flag.
func (c *EACConfig) GetDefaultBooksByModule(moniker string) []*Book {
	if c.Books == nil || c.Repository == nil {
		return nil
	}
	// Find the module and get its books list
	module := c.Repository.GetByMoniker(moniker)
	if module == nil {
		return nil
	}
	return c.Books.GetDefaultBooksByNames(module.Books)
}

// GetModuleForBook finds which module owns a book by name.
// Returns the module moniker or empty string if not found.
func (c *EACConfig) GetModuleForBook(bookName string) string {
	if c.Repository == nil {
		return ""
	}
	for _, module := range c.Repository.Modules {
		for _, b := range module.Books {
			if b == bookName {
				return module.Moniker
			}
		}
	}
	return ""
}

// GetEvidenceBooksByModule returns all evidence books for a module.
// Evidence books are built via 'update evidence' command, not the build command.
func (c *EACConfig) GetEvidenceBooksByModule(moniker string) []*Book {
	if c.Books == nil || c.Repository == nil {
		return nil
	}
	module := c.Repository.GetByMoniker(moniker)
	if module == nil {
		return nil
	}
	return c.Books.GetBooksByNames(module.EvidenceBooks)
}

// GetModulesWithEvidenceBooks returns all module monikers that have evidence_books configured.
func (c *EACConfig) GetModulesWithEvidenceBooks() []string {
	if c.Repository == nil {
		return nil
	}
	var modules []string
	for _, module := range c.Repository.Modules {
		if len(module.EvidenceBooks) > 0 {
			modules = append(modules, module.Moniker)
		}
	}
	return modules
}

// readConfigFile reads a config file from the config root.
func (c *EACConfig) readConfigFile(filename string) ([]byte, error) {
	path := filepath.Join(c.ConfigRoot, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return data, nil
}

// readConfigFileWithFallback reads a config file with fallback to system defaults.
// It tries user override first (c.ConfigRoot), then falls back to system defaults
// in the Docker image (R2R_CONTAINER_ROOT/.r2r/eac/).
// Returns error only if neither location has the file.
func (c *EACConfig) readConfigFileWithFallback(filename string) ([]byte, error) {
	// Try user override first
	userPath := filepath.Join(c.ConfigRoot, filename)
	data, err := os.ReadFile(userPath)
	if err == nil {
		return data, nil
	}

	// If not found in user config, try system defaults
	if os.IsNotExist(err) {
		systemRoot := os.Getenv("R2R_CONTAINER_ROOT")
		if systemRoot == "" {
			// Local dev mode - system defaults are same as user config root
			systemRoot = c.RepoRoot
		}
		systemPath := filepath.Join(systemRoot, ".r2r", "eac", filename)
		data, err = os.ReadFile(systemPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("config file not found in user repo (%s) or system defaults (%s)", userPath, systemPath)
			}
			return nil, fmt.Errorf("failed to read %s from system defaults: %w", filename, err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("failed to read %s: %w", userPath, err)
}

// GetBuildArtifacts returns the final list of artifact IDs to build for a module.
// This method encapsulates all artifact merging and filtering logic:
// - Merges module-level artifacts (from repository.yml) with type-level artifacts (from module-types.yml)
// - Module-level artifacts take priority if defined
// - Filters UPX-compressed artifacts: only included when buildAll is true
// - When buildAll is false: returns only current platform artifacts
// - When buildAll is true: returns all platform artifacts plus UPX variants
// - Adds book-derived artifacts if module has books defined
//
// This is the ONLY method non-core code should use to determine build artifacts.
func (c *EACConfig) GetBuildArtifacts(moniker string, buildAll bool) []Artifact {
	module := c.Repository.GetByMoniker(moniker)
	if module == nil {
		return nil
	}

	moduleType := c.ModuleTypes.Get(module.Type)

	// Determine source artifacts: module-level takes priority over type-level
	var artifacts []Artifact
	if module.Build != nil && len(module.Build.Artifacts) > 0 {
		// Use module-level artifacts
		for _, ma := range module.Build.Artifacts {
			artifacts = append(artifacts, Artifact{
				ID:          ma.ID,
				Type:        ma.Type,
				Pattern:     ma.Pattern,
				Compression: ma.Compression,
				DeriveFrom:  ma.DeriveFrom,
			})
		}
	} else if moduleType != nil && moduleType.Build != nil && len(moduleType.Build.Artifacts) > 0 {
		// Use type-level artifacts
		artifacts = moduleType.Build.Artifacts
	}

	// Add book-derived artifacts if module has books
	if len(module.Books) > 0 {
		if err := c.LoadBooks(false); err != nil {
			// Books loading is non-critical for artifact resolution
			// Continue with module-defined artifacts only
		}
		bookArtifacts := c.generateBookArtifacts(module, buildAll)

		// Deduplicate by ID
		existingIDs := make(map[string]bool)
		for _, a := range artifacts {
			existingIDs[a.ID] = true
		}
		for _, ba := range bookArtifacts {
			if !existingIDs[ba.ID] {
				artifacts = append(artifacts, ba)
				existingIDs[ba.ID] = true
			}
		}
	}

	// Filter and expand artifacts based on buildAll flag
	return c.filterArtifacts(artifacts, buildAll)
}

// GetBuildArtifactIDs returns just the artifact IDs to build for a module.
// This is a convenience wrapper around GetBuildArtifacts for callers that only need IDs.
func (c *EACConfig) GetBuildArtifactIDs(moniker string, buildAll bool) []string {
	artifacts := c.GetBuildArtifacts(moniker, buildAll)
	ids := make([]string, len(artifacts))
	for i, a := range artifacts {
		ids[i] = a.ID
	}
	return ids
}

// filterArtifacts filters and expands artifacts based on the buildAll flag.
func (c *EACConfig) filterArtifacts(artifacts []Artifact, buildAll bool) []Artifact {
	var result []Artifact
	seenIDs := make(map[string]bool)

	for _, artifact := range artifacts {
		// Skip UPX-compressed artifacts unless --all flag is used
		if artifact.Compression == CompressionUPX && !buildAll {
			continue
		}

		// For executables, handle platform expansion
		if artifact.Type == ArtifactTypeExecutable && len(artifact.Platforms) > 0 {
			if buildAll {
				// Expand to all platforms
				for _, os := range artifact.Platforms {
					archs := c.getArchitecturesForPlatform(os, artifact.Pattern)
					for _, arch := range archs {
						id := c.deriveArtifactID(artifact, os, arch)
						if !seenIDs[id] {
							seenIDs[id] = true
							result = append(result, Artifact{
								ID:          id,
								Type:        artifact.Type,
								Pattern:     artifact.Pattern,
								Compression: artifact.Compression,
								DeriveFrom:  artifact.DeriveFrom,
								Platforms:   []string{os},
							})
						}
					}
				}
			} else {
				// Current platform only - use original ID
				if !seenIDs[artifact.ID] {
					seenIDs[artifact.ID] = true
					result = append(result, artifact)
				}
			}
		} else {
			// Non-executable artifacts - include as-is
			if !seenIDs[artifact.ID] {
				seenIDs[artifact.ID] = true
				result = append(result, artifact)
			}
		}
	}

	return result
}

// getArchitecturesForPlatform determines supported architectures based on OS and pattern.
func (c *EACConfig) getArchitecturesForPlatform(os, pattern string) []string {
	// Check for architecture-specific patterns
	if containsAny(pattern, "-amd64", "{os}-amd64") {
		return []string{"amd64"}
	}
	if containsAny(pattern, "-arm64", "{os}-arm64") {
		return []string{"arm64"}
	}
	// Windows only supports amd64
	if os == "windows" {
		return []string{"amd64"}
	}
	// Linux and Darwin support both
	return []string{"amd64", "arm64"}
}

// deriveArtifactID creates an artifact ID based on OS, arch, and compression.
func (c *EACConfig) deriveArtifactID(artifact Artifact, os, arch string) string {
	baseID := os + "-" + arch
	if artifact.Compression == CompressionUPX {
		return baseID + "-upx"
	}
	return baseID
}

// generateBookArtifacts creates artifact definitions from books for a module.
func (c *EACConfig) generateBookArtifacts(module *Module, buildAll bool) []Artifact {
	var artifacts []Artifact

	books := c.Books.GetBooksByNames(module.Books)
	for i, book := range books {
		// Skip non-default books (not first) unless --all flag is used
		isDefault := i == 0
		if !buildAll && !isDefault {
			continue
		}

		output := book.GetOutput()
		switch {
		case output == "site":
			artifacts = append(artifacts, Artifact{
				ID:      "site",
				Type:    ArtifactTypeDirectory,
				Pattern: "site",
			})
		case len(output) > 4 && output[:4] == "pdf-":
			theme := output[4:]
			if theme == "all" {
				for _, t := range []string{"dark", "light"} {
					artifacts = append(artifacts, Artifact{
						ID:      book.Name + "-" + t,
						Type:    ArtifactTypeFile,
						Pattern: book.Name + "-" + t + ".pdf",
					})
				}
			} else {
				artifacts = append(artifacts, Artifact{
					ID:      book.Name,
					Type:    ArtifactTypeFile,
					Pattern: book.Name + "-" + theme + ".pdf",
				})
			}
		}
	}

	return artifacts
}

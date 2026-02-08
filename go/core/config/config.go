// Package config provides a central configuration loader for all EAC repository configs.
// It consolidates loading of modules, environments, testing-tags, and test-suites
// with integrated JSON Schema validation.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	scanner "github.com/ready-to-release/eac/contracts/scanner/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/schema"
	"github.com/ready-to-release/eac/go/core/paths"
	"github.com/ready-to-release/eac/go/core/workspace"
	"gopkg.in/yaml.v3"
)

// EACConfigRelPath is re-exported for backwards compatibility.
const EACConfigRelPath = paths.EACConfigRelPath

// Config file names.
const (
	EnvironmentsFileName   = "environments.yml"
	TestingTagsFileName    = "testing-tags.yml"
	TestSuitesFileName     = "test-suites.yml"
	ComponentTypesFileName = "component-types.yml"
)

// EACConfig holds all loaded EAC repository configuration.
// Use Load() to create and populate this struct.
//
// # Field Guarantees (after successful Load)
//
// All config fields are guaranteed non-nil after a successful Load() call.
// The loader uses fail-fast for core configs and empty defaults for optional configs:
//
//   - Repository:     GUARANTEED non-nil (fails if cannot load defaults)
//   - ComponentTypes: GUARANTEED non-nil (loads defaults if file missing)
//   - Security:       GUARANTEED non-nil (loads from security contract)
//   - Environments:   GUARANTEED non-nil (empty if file missing)
//   - TestingTags:    GUARANTEED non-nil (empty if file missing)
//   - TestSuites:     GUARANTEED non-nil (empty if file missing)
//   - Books:          GUARANTEED non-nil (empty if file missing)
//   - Commands:       GUARANTEED non-nil (uses defaults if file missing)
//
// # Module Access
//
// Modules are now part of Repository config. Access via cfg.Repository.Modules.
//
// # Tool Verification
//
// Tool availability is verified via the tool package (tool.GlobalRegistry().VerifyAll).
// Tools are defined in tool-config.yml, not system-dependencies.yml.
//
// # Sub-field Nil Checks Still Required
//
// While top-level config fields are guaranteed non-nil, nested fields may be nil:
//   - component.Build may be nil (component doesn't define artifacts)
//   - module.Metadata may be nil (no metadata defined)
type EACConfig struct {
	// Root paths
	RepoRoot   string
	ConfigRoot string

	// Core configs - fail-fast if loading fails (non-nil guaranteed)
	// Repository now includes modules (unified config from repository.yml)
	Repository *RepositoryConfig

	// Optional configs - empty defaults if file missing (non-nil guaranteed)
	Environments   *EnvironmentsConfig
	TestingTags    *TestingTagsConfig
	TestSuites     *TestSuitesConfig
	Books          *BooksConfig
	Commands       *CommandsConfig
	ComponentTypes *ComponentTypesConfig
	LintProviders  *LintProvidersConfig

	// Blueprints: component blueprints, module templates, artifact matrices
	Blueprints *BlueprintsConfig

	// New contract-based configs (Phase 2.2/2.3)
	// These load from versioned contracts (testing, security)
	Testing  *TestingConfig  // From testing contract
	Security *SecurityConfig // From security contract

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
		repoRoot, err = workspace.Root()
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
// Core configs (Repository, ComponentTypes) must load successfully - fails immediately on error.
// Optional configs (Environments, TestingTags, etc.) collect errors but continue loading.
func (c *EACConfig) LoadAll(opts LoadOptions) error {
	validateSchemas := opts.ValidateSchemas
	// === CORE CONFIGS: Fail fast if any of these fail ===
	// These are required for the system to function correctly

	// Load repository config first (now includes modules)
	if err := c.LoadRepository(validateSchemas); err != nil {
		return fmt.Errorf("core config failed - repository: %w", err)
	}

	// Load component types for component-specific defaults
	if err := c.LoadComponentTypes(validateSchemas); err != nil {
		return fmt.Errorf("core config failed - component-types: %w", err)
	}

	// Apply component-specific defaults after both modules and component types are loaded
	// Skip if Repository is nil (can happen in test environments without contract files)
	if c.Repository != nil {
		c.Repository.ApplyComponentDefaults(c.ComponentTypes, c.RepoRoot)
		c.Repository.computeDisplayOrder(c.ComponentTypes)
	}

	// === OPTIONAL CONFIGS: Continue on error, collect for reporting ===
	var errs []error

	if err := c.LoadEnvironments(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("environments: %w", err))
	}

	if err := c.LoadTestingTags(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("testing-tags: %w", err))
	}

	if err := c.LoadTimeouts(); err != nil {
		errs = append(errs, fmt.Errorf("timeouts: %w", err))
	}

	if err := c.LoadTestSuites(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("test-suites: %w", err))
	}

	if err := c.LoadBooks(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("books: %w", err))
	}

	if err := c.LoadCommands(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("commands: %w", err))
	}

	if err := c.LoadLintProviders(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("lint-providers: %w", err))
	}

	// Load new contract-based configs (Phase 2.2/2.3)
	if err := c.LoadTesting(); err != nil {
		errs = append(errs, fmt.Errorf("testing (testing): %w", err))
	}

	if err := c.LoadSecurity(); err != nil {
		errs = append(errs, fmt.Errorf("security (security): %w", err))
	}

	// Note: ComponentTypes is loaded in core configs section above

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

	// Step 9: Load blueprints and expand module templates
	if err := c.LoadBlueprints(validateSchema); err != nil {
		return fmt.Errorf("loading blueprints: %w", err)
	}

	if err := c.Repository.ExpandModuleTemplates(c.RepoRoot, c.Blueprints); err != nil {
		return fmt.Errorf("expanding module templates: %w", err)
	}

	// Step 10: Expand module groups in depends_on
	// Must happen after template expansion (templates can add depends_on with group names)
	// Also validates and strips the "root" sentinel for baseline tooling modules
	if err := c.Repository.expandModuleGroups(); err != nil {
		return fmt.Errorf("expanding module groups: %w", err)
	}

	return nil
}

// LoadEnvironments loads the environments configuration.
// Loads contract defaults first, then merges user config on top if present.
func (c *EACConfig) LoadEnvironments(validateSchema bool) error {
	// Load defaults from contract
	cfg, err := LoadEnvironmentsDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return err
	}

	// Ensure we have a valid config even if no defaults
	if cfg == nil {
		cfg = &EnvironmentsConfig{}
	}

	// Try loading user override
	userPath := filepath.Join(c.ConfigRoot, EnvironmentsFileName)
	data, err := os.ReadFile(userPath)
	if err == nil {
		if validateSchema {
			if err := c.validateSchema(schema.SchemaEnvironments, data); err != nil {
				return err
			}
		}

		var userCfg EnvironmentsConfig
		if err := yaml.Unmarshal(data, &userCfg); err != nil {
			return fmt.Errorf("failed to parse %s: %w", EnvironmentsFileName, err)
		}

		// Merge user environments into defaults (user overrides)
		for _, userEnv := range userCfg.Environments {
			found := false
			for i, defEnv := range cfg.Environments {
				if defEnv.Moniker == userEnv.Moniker {
					cfg.Environments[i] = userEnv // Replace
					found = true
					break
				}
			}
			if !found {
				cfg.Environments = append(cfg.Environments, userEnv)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", EnvironmentsFileName, err)
	}

	c.Environments = cfg
	return nil
}

// LoadTestingTags loads the testing-tags configuration.
// Loads contract defaults first, then merges user config on top if present.
func (c *EACConfig) LoadTestingTags(validateSchema bool) error {
	// Load defaults from contract
	cfg, err := LoadTestingTagsDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return err
	}

	// Ensure we have a valid config even if no defaults
	if cfg == nil {
		cfg = &TestingTagsConfig{}
	}

	// Load user config if present (optional)
	userPath := filepath.Join(c.ConfigRoot, TestingTagsFileName)
	if data, err := os.ReadFile(userPath); err == nil {
		if validateSchema {
			if err := c.validateSchema(schema.SchemaTestingTags, data); err != nil {
				return err
			}
		}

		var userCfg TestingTagsConfig
		if err := yaml.Unmarshal(data, &userCfg); err != nil {
			return fmt.Errorf("failed to parse %s: %w", TestingTagsFileName, err)
		}

		// Merge user config into defaults (user takes precedence)
		if len(userCfg.Tags) > 0 {
			cfg.Tags = userCfg.Tags
		}
		if len(userCfg.Types) > 0 {
			cfg.Types = userCfg.Types
		}
		if len(userCfg.SkipReasons) > 0 {
			cfg.SkipReasons = userCfg.SkipReasons
		}
	}

	if err := cfg.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize testing-tags: %w", err)
	}

	c.TestingTags = cfg
	return nil
}

// LoadTimeouts loads the timeout configuration.
// Loads contract defaults first, then merges personal overrides from .eac/timeouts.personal.yml.
// Sets the global timeout configuration for use throughout the application.
func (c *EACConfig) LoadTimeouts() error {
	// Load defaults from contract
	cfg, err := LoadTimeoutsDefaults(c.RepoRoot)
	if err != nil && !errors.Is(err, ErrNoDefaults) {
		return fmt.Errorf("loading timeouts defaults: %w", err)
	}

	// Ensure we have a valid config even if no defaults
	if cfg == nil {
		cfg = DefaultTimeoutConfig()
	}

	// Load personal override if it exists (.eac/timeouts.personal.yml)
	personalPath := filepath.Join(c.ConfigRoot, "timeouts.personal.yml")
	if data, err := os.ReadFile(personalPath); err == nil {
		var personalCfg TimeoutConfig
		if err := yaml.Unmarshal(data, &personalCfg); err != nil {
			return fmt.Errorf("failed to parse timeouts.personal.yml: %w", err)
		}

		// Merge personal overrides into defaults
		cfg = MergeTimeoutConfigs(cfg, &personalCfg)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read timeouts.personal.yml: %w", err)
	}

	// Set global timeouts for use throughout the application
	SetGlobalTimeouts(cfg)

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

// LoadBooks loads the books configuration with template expansion.
// Loads both defaults and user config, then expands templates and snippets.
func (c *EACConfig) LoadBooks(validateSchema bool) error {
	// Load defaults from contracts folder
	var defaultsRaw *BooksConfigRaw
	defaultsData, err := loadDefaultFile(c.RepoRoot, BooksFileName)
	if err == nil {
		defaultsRaw, err = LoadBooksConfigRaw(defaultsData)
		if err != nil {
			return fmt.Errorf("failed to parse default %s: %w", BooksFileName, err)
		}
	}

	// Load user config (optional)
	var userRaw *BooksConfigRaw
	booksPath := filepath.Join(c.ConfigRoot, BooksFileName)
	if _, err := os.Stat(booksPath); err == nil {
		data, err := c.readConfigFile(BooksFileName)
		if err != nil {
			return err
		}

		if validateSchema {
			if err := c.validateSchema(schema.SchemaBooks, data); err != nil {
				return err
			}
		}

		userRaw, err = LoadBooksConfigRaw(data)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", BooksFileName, err)
		}
	}

	// Get modules for evidence book generation
	var modules []Module
	if c.Repository != nil {
		modules = c.Repository.Modules
	}

	// Merge and expand
	c.Books = MergeBooksConfigs(defaultsRaw, userRaw, modules)
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

// LoadBlueprints loads the blueprints configuration.
// Loads contract defaults (required), then merges optional .eac/blueprints.yml override.
func (c *EACConfig) LoadBlueprints(validateSchema bool) error {
	// 1. Load defaults (required — errors are fatal)
	cfg, err := LoadBlueprintsDefaults(c.RepoRoot)
	if err != nil {
		return fmt.Errorf("loading blueprints defaults: %w", err)
	}

	// 2. Load .eac/blueprints.yml override (optional)
	overridePath := filepath.Join(c.ConfigRoot, BlueprintsFileName)
	if data, err := os.ReadFile(overridePath); err == nil {
		if validateSchema {
			if err := c.validateSchema(schema.SchemaBlueprints, data); err != nil {
				return err
			}
		}

		var override BlueprintsConfig
		if err := yaml.Unmarshal(data, &override); err != nil {
			return fmt.Errorf("parsing %s: %w", BlueprintsFileName, err)
		}
		cfg = MergeBlueprintsConfig(cfg, &override)
	}

	c.Blueprints = cfg

	// If component-kinds are defined in blueprints, use them as ComponentTypes.
	// This enables the unified config: component kinds live in blueprints.yml
	// alongside templates and artifact matrices.
	if len(cfg.ComponentKinds) > 0 {
		if c.ComponentTypes == nil {
			c.ComponentTypes = &ComponentTypesConfig{
				ComponentTypes: make(map[string]*ComponentType),
			}
		}
		for name, kind := range cfg.ComponentKinds {
			c.ComponentTypes.ComponentTypes[name] = kind
		}
	}

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
	return c.Books.GetBooksByNames(module.GetBooks())
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
	return c.Books.GetDefaultBooksByNames(module.GetBooks())
}

// GetModuleForBook finds which module owns a book by name.
// Returns the module moniker or empty string if not found.
func (c *EACConfig) GetModuleForBook(bookName string) string {
	if c.Repository == nil {
		return ""
	}
	for i := range c.Repository.Modules {
		for _, b := range c.Repository.Modules[i].GetBooks() {
			if b == bookName {
				return c.Repository.Modules[i].Moniker
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
	for i := range c.Repository.Modules {
		if len(c.Repository.Modules[i].EvidenceBooks) > 0 {
			modules = append(modules, c.Repository.Modules[i].Moniker)
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

	// Collect artifacts from all components
	var artifacts []Artifact
	for _, comp := range module.Components {
		if comp != nil && comp.Build != nil {
			for _, ma := range comp.Build.Artifacts {
				artifacts = append(artifacts, Artifact{
					ID:          ma.ID,
					Type:        ma.Type,
					Pattern:     ma.Pattern,
					Compression: ma.Compression,
					DeriveFrom:  ma.DeriveFrom,
				})
			}
		}
	}

	// Add book-derived artifacts if module has books
	if len(module.GetBooks()) > 0 {
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
func (c *EACConfig) getArchitecturesForPlatform(targetOS, pattern string) []string {
	// Check for architecture-specific patterns
	if containsAny(pattern, "-amd64", "{os}-amd64") {
		return []string{"amd64"}
	}
	if containsAny(pattern, "-arm64", "{os}-arm64") {
		return []string{"arm64"}
	}
	// Windows only supports amd64
	if targetOS == "windows" {
		return []string{"amd64"}
	}
	// Linux and Darwin support both
	return []string{"amd64", "arm64"}
}

// deriveArtifactID creates an artifact ID based on OS, arch, and compression.
func (c *EACConfig) deriveArtifactID(artifact Artifact, targetOS, arch string) string {
	baseID := targetOS + "-" + arch
	if artifact.Compression == CompressionUPX {
		return baseID + "-upx"
	}
	return baseID
}

// generateBookArtifacts creates artifact definitions from books for a module.
func (c *EACConfig) generateBookArtifacts(module *Module, buildAll bool) []Artifact {
	var artifacts []Artifact

	books := c.Books.GetBooksByNames(module.GetBooks())
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

// LoadTesting loads the testing configuration from the testing contract.
// This provides TestConfigPort interface access to suites and tags.
func (c *EACConfig) LoadTesting() error {
	cfg, err := LoadTestingConfig(c.RepoRoot, c.ConfigRoot)
	if err != nil {
		// Non-fatal: testing config is optional, provide empty default
		c.Testing = &TestingConfig{
			suites:   make(map[string]*core.SuiteDefinition),
			tags:     make(map[string]*core.TagDefinition),
			tagTypes: make(map[string]core.TagType),
		}
		return nil
	}
	c.Testing = cfg
	return nil
}

// LoadSecurity loads the security configuration from the security contract.
// This provides SecurityConfigPort interface access to scanners and policies.
func (c *EACConfig) LoadSecurity() error {
	cfg, err := LoadSecurityConfig(c.RepoRoot, c.ConfigRoot)
	if err != nil {
		// Non-fatal: security config is optional, provide empty default
		c.Security = &SecurityConfig{
			scanners: make(map[string]*scanner.ScannerDefinition),
			policies: &scanner.PoliciesConfig{
				ComponentScanners: make(map[string][]string),
				Default:           []string{"trivy-sbom", "trivy-vuln"},
			},
		}
		return nil
	}
	c.Security = cfg
	return nil
}

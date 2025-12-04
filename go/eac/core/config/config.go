// Package config provides a central configuration loader for all EAC repository configs.
// It consolidates loading of modules, environments, testing-tags, and test-suites
// with integrated JSON Schema validation.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/ready-to-release/eac/go/eac/core/contracts/schema"
	"gopkg.in/yaml.v3"
)

// EAC configuration path constants (local to avoid import cycle with repository)
const (
	r2rDir = ".r2r"
	eacDir = "eac"

	// EACConfigRelPath is the relative path from repo root to EAC configuration.
	// Note: Duplicated here to avoid import cycle with repository package.
	EACConfigRelPath = r2rDir + "/" + eacDir
)

// Config file names
const (
	ModulesFileName            = "modules.yml"
	ModuleTypesFileName        = "module-types.yml"
	EnvironmentsFileName       = "environments.yml"
	TestingTagsFileName        = "testing-tags.yml"
	TestSuitesFileName         = "test-suites.yml"
	SystemDependenciesFileName = "system-dependencies.yml"
)

// EACConfig holds all loaded EAC repository configuration.
// Use Load() to create and populate this struct.
type EACConfig struct {
	// Root paths
	RepoRoot   string
	ConfigRoot string

	// Repository-wide configuration (loaded first, used by other configs)
	Repository *RepositoryConfig

	// Loaded configurations
	Modules            *ModulesConfig
	ModuleTypes        *ModuleTypesConfig
	Environments       *EnvironmentsConfig
	TestingTags        *TestingTagsConfig
	TestSuites         *TestSuitesConfig
	SystemDependencies *SystemDependenciesConfig
	Handlers           *HandlersConfig
	Books              *BooksConfig

	// Schema validator (lazy initialized)
	validator     *schema.Validator
	validatorOnce sync.Once
	validatorErr  error
}

// LoadOptions configures the loading behavior
type LoadOptions struct {
	// RepoRoot overrides automatic repo root detection
	RepoRoot string
	// ValidateSchemas enables JSON Schema validation during load (default: true)
	ValidateSchemas bool
	// LazyLoad defers loading until first access (default: false)
	LazyLoad bool
}

// DefaultLoadOptions returns the default load options
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        false,
	}
}

// Load loads all EAC configuration from the repository.
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

	configRoot := filepath.Join(repoRoot, r2rDir, eacDir)

	cfg := &EACConfig{
		RepoRoot:   repoRoot,
		ConfigRoot: configRoot,
	}

	if opts.LazyLoad {
		return cfg, nil
	}

	// Load all configs
	if err := cfg.LoadAll(opts.ValidateSchemas); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadAll loads all configuration files
func (c *EACConfig) LoadAll(validateSchemas bool) error {
	var errs []error

	// Load repository config first (provides path variables for other configs)
	if err := c.LoadRepository(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("repository: %w", err))
	}

	if err := c.LoadModules(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("modules: %w", err))
	}

	if err := c.LoadModuleTypes(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("module-types: %w", err))
	}

	// Apply type-specific defaults after both modules and types are loaded
	if c.Modules != nil {
		c.Modules.ApplyTypeDefaults(c.ModuleTypes, c.Repository)
	}

	if err := c.LoadEnvironments(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("environments: %w", err))
	}

	if err := c.LoadTestingTags(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("testing-tags: %w", err))
	}

	if err := c.LoadTestSuites(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("test-suites: %w", err))
	}

	if err := c.LoadSystemDependencies(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("system-dependencies: %w", err))
	}

	if err := c.LoadHandlers(validateSchemas); err != nil {
		errs = append(errs, fmt.Errorf("handlers: %w", err))
	}

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}

	return nil
}

// LoadRepository loads the repository-wide configuration
func (c *EACConfig) LoadRepository(validateSchema bool) error {
	// repository.yml is optional - check if it exists
	repoPath := filepath.Join(c.ConfigRoot, RepositoryFileName)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		// Use defaults if file doesn't exist
		cfg := DefaultRepositoryConfig()
		c.Repository = &cfg
		return nil
	}

	if validateSchema {
		data, err := c.readConfigFile(RepositoryFileName)
		if err != nil {
			return err
		}
		if err := c.validateSchema(schema.SchemaRepository, data); err != nil {
			return err
		}
	}

	cfg, err := LoadRepositoryConfig(c.RepoRoot)
	if err != nil {
		return err
	}
	c.Repository = cfg
	return nil
}

// LoadModules loads the modules configuration
func (c *EACConfig) LoadModules(validateSchema bool) error {
	data, err := c.readConfigFile(ModulesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaModules, data); err != nil {
			return err
		}
	}

	var cfg ModulesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", ModulesFileName, err)
	}

	cfg.applyDefaults()
	c.Modules = &cfg
	return nil
}

// LoadModuleTypes loads the module-types configuration
func (c *EACConfig) LoadModuleTypes(validateSchema bool) error {
	data, err := c.readConfigFile(ModuleTypesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaModuleTypes, data); err != nil {
			return err
		}
	}

	var cfg ModuleTypesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", ModuleTypesFileName, err)
	}

	cfg.buildTypeMap()
	c.ModuleTypes = &cfg
	return nil
}

// LoadEnvironments loads the environments configuration
func (c *EACConfig) LoadEnvironments(validateSchema bool) error {
	data, err := c.readConfigFile(EnvironmentsFileName)
	if err != nil {
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

// LoadTestingTags loads the testing-tags configuration
func (c *EACConfig) LoadTestingTags(validateSchema bool) error {
	data, err := c.readConfigFile(TestingTagsFileName)
	if err != nil {
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

// LoadTestSuites loads the test-suites configuration
func (c *EACConfig) LoadTestSuites(validateSchema bool) error {
	data, err := c.readConfigFile(TestSuitesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaTestSuites, data); err != nil {
			return err
		}
	}

	var cfg TestSuitesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", TestSuitesFileName, err)
	}

	cfg.buildSuiteMap()
	c.TestSuites = &cfg
	return nil
}

// LoadSystemDependencies loads the system-dependencies configuration
func (c *EACConfig) LoadSystemDependencies(validateSchema bool) error {
	data, err := c.readConfigFile(SystemDependenciesFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaSystemDependencies, data); err != nil {
			return err
		}
	}

	var cfg SystemDependenciesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", SystemDependenciesFileName, err)
	}

	cfg.buildDepMap()
	c.SystemDependencies = &cfg
	return nil
}

// LoadHandlers loads the handlers configuration
func (c *EACConfig) LoadHandlers(validateSchema bool) error {
	// Check if handlers file exists - it's optional
	handlersPath := filepath.Join(c.ConfigRoot, HandlersFileName)
	if _, err := os.Stat(handlersPath); os.IsNotExist(err) {
		// Handlers config is optional - use empty config
		c.Handlers = &HandlersConfig{}
		c.Handlers.buildHandlerMap()
		return nil
	}

	data, err := c.readConfigFile(HandlersFileName)
	if err != nil {
		return err
	}

	if validateSchema {
		if err := c.validateSchema(schema.SchemaHandlers, data); err != nil {
			return err
		}
	}

	var cfg HandlersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse %s: %w", HandlersFileName, err)
	}

	cfg.buildHandlerMap()
	c.Handlers = &cfg
	return nil
}

// LoadBooks loads the books configuration (optional - only if file exists)
func (c *EACConfig) LoadBooks(validateSchema bool) error {
	// Check if books file exists - it's optional
	booksPath := filepath.Join(c.ConfigRoot, BooksFileName)
	if _, err := os.Stat(booksPath); os.IsNotExist(err) {
		// Books config is optional - no books defined
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

// GetBookByName finds a book by its name (module moniker)
func (c *EACConfig) GetBookByName(name string) *Book {
	if c.Books == nil {
		return nil
	}
	return c.Books.GetBookByName(name)
}

// readConfigFile reads a config file from the config root
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

// validateSchema validates data against a JSON schema
func (c *EACConfig) validateSchema(schemaType schema.SchemaType, data []byte) error {
	c.validatorOnce.Do(func() {
		c.validator, c.validatorErr = schema.NewValidator(c.RepoRoot)
	})

	if c.validatorErr != nil {
		return fmt.Errorf("failed to initialize schema validator: %w", c.validatorErr)
	}

	return c.validator.ValidateYAML(schemaType, data)
}

// ValidateAll validates all loaded configs against their schemas
func (c *EACConfig) ValidateAll() error {
	var errs []error

	// Repository config is optional
	if c.Repository != nil {
		repoPath := filepath.Join(c.ConfigRoot, RepositoryFileName)
		if _, err := os.Stat(repoPath); err == nil {
			data, _ := c.readConfigFile(RepositoryFileName)
			if err := c.validateSchema(schema.SchemaRepository, data); err != nil {
				errs = append(errs, fmt.Errorf("repository: %w", err))
			}
		}
	}

	if c.Modules != nil {
		data, _ := c.readConfigFile(ModulesFileName)
		if err := c.validateSchema(schema.SchemaModules, data); err != nil {
			errs = append(errs, fmt.Errorf("modules: %w", err))
		}
	}

	if c.ModuleTypes != nil {
		data, _ := c.readConfigFile(ModuleTypesFileName)
		if err := c.validateSchema(schema.SchemaModuleTypes, data); err != nil {
			errs = append(errs, fmt.Errorf("module-types: %w", err))
		}
	}

	if c.Environments != nil {
		data, _ := c.readConfigFile(EnvironmentsFileName)
		if err := c.validateSchema(schema.SchemaEnvironments, data); err != nil {
			errs = append(errs, fmt.Errorf("environments: %w", err))
		}
	}

	if c.TestingTags != nil {
		data, _ := c.readConfigFile(TestingTagsFileName)
		if err := c.validateSchema(schema.SchemaTestingTags, data); err != nil {
			errs = append(errs, fmt.Errorf("testing-tags: %w", err))
		}
	}

	if c.TestSuites != nil {
		data, _ := c.readConfigFile(TestSuitesFileName)
		if err := c.validateSchema(schema.SchemaTestSuites, data); err != nil {
			errs = append(errs, fmt.Errorf("test-suites: %w", err))
		}
	}

	if c.SystemDependencies != nil {
		data, _ := c.readConfigFile(SystemDependenciesFileName)
		if err := c.validateSchema(schema.SchemaSystemDependencies, data); err != nil {
			errs = append(errs, fmt.Errorf("system-dependencies: %w", err))
		}
	}

	if c.Handlers != nil {
		data, err := c.readConfigFile(HandlersFileName)
		// Only validate if file exists (handlers are optional)
		if err == nil {
			if err := c.validateSchema(schema.SchemaHandlers, data); err != nil {
				errs = append(errs, fmt.Errorf("handlers: %w", err))
			}
		}
	}

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}

	return nil
}

// MultiError holds multiple errors
type MultiError struct {
	Errors []error
}

func (e *MultiError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	msg := fmt.Sprintf("%d errors:", len(e.Errors))
	for _, err := range e.Errors {
		msg += "\n  - " + err.Error()
	}
	return msg
}

// findRepositoryRoot finds the git repository root by walking up directories.
// This is a local implementation to avoid import cycles with the repository package.
func findRepositoryRoot(startPath string) (string, error) {
	// Check for explicit repository root override first
	// This takes precedence over R2R_DOCKER_MODE to allow test isolation
	if repoRoot := os.Getenv("R2R_REPO_ROOT"); repoRoot != "" {
		return filepath.Clean(repoRoot), nil
	}

	// Check for Docker R2R mode - repository is mounted at /var/task
	// Only applies when no explicit override is set
	if os.Getenv("R2R_DOCKER_MODE") == "true" {
		return "/var/task", nil
	}

	// Use current directory if no path provided
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Walk up looking for .git directory
	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				return currentPath, nil
			}
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
		}
		currentPath = parentPath
	}
}

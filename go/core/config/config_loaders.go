// File: go/core/config/config_loaders.go
// Purpose: EACConfig Load* methods for optional configuration files.
//
// Extracted from config.go to keep that file focused on the core Load/LoadAll
// entry points and artifact helpers.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	scanner "github.com/ready-to-release/eac/contracts/scanner/0.1.0"
	"github.com/ready-to-release/eac/go/core/domain/schema"
	"gopkg.in/yaml.v3"
)

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
// Three-tier merge: contract defaults → .eac/timeouts.yml (repo) → .eac/timeouts.personal.yml (local).
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

	// Load repository-level override if it exists (.eac/timeouts.yml)
	repoPath := filepath.Join(c.ConfigRoot, "timeouts.yml")
	if data, err := os.ReadFile(repoPath); err == nil {
		var repoCfg TimeoutConfig
		if err := yaml.Unmarshal(data, &repoCfg); err != nil {
			return fmt.Errorf("failed to parse timeouts.yml: %w", err)
		}
		cfg = MergeTimeoutConfigs(cfg, &repoCfg)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read timeouts.yml: %w", err)
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

	// If component-kinds are defined in blueprints, load them into ComponentKinds.
	// This enables the unified config: component kinds live in blueprints.yml
	// alongside templates and artifact matrices.
	if len(cfg.ComponentKinds) > 0 {
		if c.ComponentKinds == nil {
			c.ComponentKinds = &ComponentKindsConfig{
				Kinds: make(map[string]*ComponentType),
			}
		}
		for name, kind := range cfg.ComponentKinds {
			c.ComponentKinds.Kinds[name] = kind
		}
	}

	return nil
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
	cfg, err := LoadSecurityConfig(c.ConfigRoot)
	if err != nil {
		// Non-fatal: security config is optional, provide empty default
		c.Security = &SecurityConfig{
			scanners: make(map[string]*scanner.ScannerDefinition),
			policies: &scanner.PoliciesConfig{
				ComponentScanners: make(map[string][]string),
				Default:           defaultScannerIDs,
			},
		}
		return nil
	}
	c.Security = cfg
	return nil
}

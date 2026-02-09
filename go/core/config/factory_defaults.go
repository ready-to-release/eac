package config

import (
	"fmt"
	"sync"

	clone "github.com/huandu/go-clone/generic"
	"gopkg.in/yaml.v3"
)

// Factory default singletons for all config types.
// Each is parsed once from the embedded contract filesystem via sync.Once.
// The stored values are never mutated after initialization.

var (
	factoryRepositoryDefaults     *RepositoryConfig
	factoryRepositoryDefaultsOnce sync.Once
	factoryRepositoryDefaultsErr  error
)

var (
	factoryEnvironmentsDefaults     *EnvironmentsConfig
	factoryEnvironmentsDefaultsOnce sync.Once
	factoryEnvironmentsDefaultsErr  error
)

var (
	factoryTestingTagsDefaults     *TestingTagsConfig
	factoryTestingTagsDefaultsOnce sync.Once
	factoryTestingTagsDefaultsErr  error
)

var (
	factoryTimeoutsDefaults     *TimeoutConfig
	factoryTimeoutsDefaultsOnce sync.Once
	factoryTimeoutsDefaultsErr  error
)

var (
	factoryTestSuitesDefaults     *TestSuitesConfig
	factoryTestSuitesDefaultsOnce sync.Once
	factoryTestSuitesDefaultsErr  error
)

var (
	factoryRegistriesDefaults     RegistriesConfig
	factoryRegistriesDefaultsOnce sync.Once
	factoryRegistriesDefaultsErr  error
)

var (
	factoryLintProvidersDefaults     *LintProvidersConfig
	factoryLintProvidersDefaultsOnce sync.Once
	factoryLintProvidersDefaultsErr  error
)

var (
	factoryBlueprintsDefaults     *BlueprintsConfig
	factoryBlueprintsDefaultsOnce sync.Once
	factoryBlueprintsDefaultsErr  error
)

// Factory repo-type defaults keyed by repo type string.
var factoryRepoTypeDefaults sync.Map // map[string]*repoTypeEntry

type repoTypeEntry struct {
	cfg *RepositoryConfig
	err error
}

// --- Repository ---

func getFactoryRepositoryDefaults() (*RepositoryConfig, error) {
	factoryRepositoryDefaultsOnce.Do(func() {
		factoryRepositoryDefaults, factoryRepositoryDefaultsErr = parseRepositoryDefaults()
	})
	return factoryRepositoryDefaults, factoryRepositoryDefaultsErr
}

func cloneRepositoryDefaults() (*RepositoryConfig, error) {
	factory, err := getFactoryRepositoryDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseRepositoryDefaults() (*RepositoryConfig, error) {
	data, err := loadDefaultFile("", "repository.yml")
	if err != nil {
		return nil, err
	}
	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory repository defaults: %w", err)
	}
	return &cfg, nil
}

// --- Repository Type ---

func getFactoryRepoTypeDefaults(repoType string) (*RepositoryConfig, error) {
	if repoType == "" {
		return nil, ErrNoDefaults
	}
	if v, ok := factoryRepoTypeDefaults.Load(repoType); ok {
		entry := v.(*repoTypeEntry)
		return entry.cfg, entry.err
	}
	filename := fmt.Sprintf("repository-%s.yml", repoType)
	data, err := loadDefaultFile("", filename)
	if err != nil {
		entry := &repoTypeEntry{err: err}
		factoryRepoTypeDefaults.Store(repoType, entry)
		return nil, err
	}
	var cfg RepositoryConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		entry := &repoTypeEntry{err: fmt.Errorf("parsing repository type defaults (%s): %w", repoType, err)}
		factoryRepoTypeDefaults.Store(repoType, entry)
		return nil, entry.err
	}
	entry := &repoTypeEntry{cfg: &cfg}
	factoryRepoTypeDefaults.Store(repoType, entry)
	return entry.cfg, nil
}

func cloneRepoTypeDefaults(repoType string) (*RepositoryConfig, error) {
	factory, err := getFactoryRepoTypeDefaults(repoType)
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

// --- Registries ---

func getFactoryRegistriesDefaults() (RegistriesConfig, error) {
	factoryRegistriesDefaultsOnce.Do(func() {
		factoryRegistriesDefaults, factoryRegistriesDefaultsErr = parseRegistriesDefaults()
	})
	return factoryRegistriesDefaults, factoryRegistriesDefaultsErr
}

func cloneRegistriesDefaults() (RegistriesConfig, error) {
	factory, err := getFactoryRegistriesDefaults()
	if err != nil {
		return nil, err
	}
	if factory == nil {
		return nil, nil
	}
	return clone.Clone(factory), nil
}

func parseRegistriesDefaults() (RegistriesConfig, error) {
	data, err := loadDefaultFile("", "registries.yml")
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Registries RegistriesConfig `yaml:"registries"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parsing factory registries defaults: %w", err)
	}
	return wrapper.Registries, nil
}

// --- Environments ---

func getFactoryEnvironmentsDefaults() (*EnvironmentsConfig, error) {
	factoryEnvironmentsDefaultsOnce.Do(func() {
		factoryEnvironmentsDefaults, factoryEnvironmentsDefaultsErr = parseEnvironmentsDefaults()
	})
	return factoryEnvironmentsDefaults, factoryEnvironmentsDefaultsErr
}

func cloneEnvironmentsDefaults() (*EnvironmentsConfig, error) {
	factory, err := getFactoryEnvironmentsDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseEnvironmentsDefaults() (*EnvironmentsConfig, error) {
	data, err := loadDefaultFile("", EnvironmentsFileName)
	if err != nil {
		return nil, err
	}
	var cfg EnvironmentsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory environments defaults: %w", err)
	}
	return &cfg, nil
}

// --- Testing Tags ---

func getFactoryTestingTagsDefaults() (*TestingTagsConfig, error) {
	factoryTestingTagsDefaultsOnce.Do(func() {
		factoryTestingTagsDefaults, factoryTestingTagsDefaultsErr = parseTestingTagsDefaults()
	})
	return factoryTestingTagsDefaults, factoryTestingTagsDefaultsErr
}

func cloneTestingTagsDefaults() (*TestingTagsConfig, error) {
	factory, err := getFactoryTestingTagsDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseTestingTagsDefaults() (*TestingTagsConfig, error) {
	data, err := loadDefaultFile("", TestingTagsFileName)
	if err != nil {
		return nil, err
	}
	var cfg TestingTagsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory testing-tags defaults: %w", err)
	}
	return &cfg, nil
}

// --- Timeouts ---

func getFactoryTimeoutsDefaults() (*TimeoutConfig, error) {
	factoryTimeoutsDefaultsOnce.Do(func() {
		factoryTimeoutsDefaults, factoryTimeoutsDefaultsErr = parseTimeoutsDefaults()
	})
	return factoryTimeoutsDefaults, factoryTimeoutsDefaultsErr
}

func cloneTimeoutsDefaults() (*TimeoutConfig, error) {
	factory, err := getFactoryTimeoutsDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseTimeoutsDefaults() (*TimeoutConfig, error) {
	data, err := loadDefaultFile("", "timeouts.yml")
	if err != nil {
		return nil, err
	}
	var cfg TimeoutConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory timeouts defaults: %w", err)
	}
	return &cfg, nil
}

// --- Test Suites ---

func getFactoryTestSuitesDefaults() (*TestSuitesConfig, error) {
	factoryTestSuitesDefaultsOnce.Do(func() {
		factoryTestSuitesDefaults, factoryTestSuitesDefaultsErr = parseTestSuitesDefaults()
	})
	return factoryTestSuitesDefaults, factoryTestSuitesDefaultsErr
}

func cloneTestSuitesDefaults() (*TestSuitesConfig, error) {
	factory, err := getFactoryTestSuitesDefaults()
	if err != nil {
		return nil, err
	}
	cloned := clone.Clone(factory)
	cloned.buildSuiteMap()
	return cloned, nil
}

func parseTestSuitesDefaults() (*TestSuitesConfig, error) {
	data, err := loadDefaultFile("", "test-suites.yml")
	if err != nil {
		return nil, err
	}
	var cfg TestSuitesConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory test-suites defaults: %w", err)
	}
	cfg.buildSuiteMap()
	return &cfg, nil
}

// --- Lint Providers ---

func getFactoryLintProvidersDefaults() (*LintProvidersConfig, error) {
	factoryLintProvidersDefaultsOnce.Do(func() {
		factoryLintProvidersDefaults, factoryLintProvidersDefaultsErr = parseLintProvidersDefaults()
	})
	return factoryLintProvidersDefaults, factoryLintProvidersDefaultsErr
}

func cloneLintProvidersDefaults() (*LintProvidersConfig, error) {
	factory, err := getFactoryLintProvidersDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseLintProvidersDefaults() (*LintProvidersConfig, error) {
	data, err := loadDefaultFile("", LintProvidersFileName)
	if err != nil {
		return nil, err
	}
	var cfg LintProvidersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory lint-providers defaults: %w", err)
	}
	return &cfg, nil
}

// --- Blueprints ---

func getFactoryBlueprintsDefaults() (*BlueprintsConfig, error) {
	factoryBlueprintsDefaultsOnce.Do(func() {
		factoryBlueprintsDefaults, factoryBlueprintsDefaultsErr = parseBlueprintsDefaults()
	})
	return factoryBlueprintsDefaults, factoryBlueprintsDefaultsErr
}

func cloneBlueprintsDefaults() (*BlueprintsConfig, error) {
	factory, err := getFactoryBlueprintsDefaults()
	if err != nil {
		return nil, err
	}
	return clone.Clone(factory), nil
}

func parseBlueprintsDefaults() (*BlueprintsConfig, error) {
	data, err := loadDefaultFile("", BlueprintsFileName)
	if err != nil {
		return nil, err
	}
	var cfg BlueprintsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing factory blueprints defaults: %w", err)
	}
	return &cfg, nil
}

// --- Test Reset ---

// ResetFactoryDefaultsForTesting resets all factory default singletons.
// Use ONLY in tests that need to verify defaults loading behavior.
func ResetFactoryDefaultsForTesting() {
	factoryRepositoryDefaultsOnce = sync.Once{}
	factoryRepositoryDefaults = nil
	factoryRepositoryDefaultsErr = nil

	factoryEnvironmentsDefaultsOnce = sync.Once{}
	factoryEnvironmentsDefaults = nil
	factoryEnvironmentsDefaultsErr = nil

	factoryTestingTagsDefaultsOnce = sync.Once{}
	factoryTestingTagsDefaults = nil
	factoryTestingTagsDefaultsErr = nil

	factoryTimeoutsDefaultsOnce = sync.Once{}
	factoryTimeoutsDefaults = nil
	factoryTimeoutsDefaultsErr = nil

	factoryTestSuitesDefaultsOnce = sync.Once{}
	factoryTestSuitesDefaults = nil
	factoryTestSuitesDefaultsErr = nil

	factoryRegistriesDefaultsOnce = sync.Once{}
	factoryRegistriesDefaults = nil
	factoryRegistriesDefaultsErr = nil

	factoryLintProvidersDefaultsOnce = sync.Once{}
	factoryLintProvidersDefaults = nil
	factoryLintProvidersDefaultsErr = nil

	factoryBlueprintsDefaultsOnce = sync.Once{}
	factoryBlueprintsDefaults = nil
	factoryBlueprintsDefaultsErr = nil

	factoryRepoTypeDefaults = sync.Map{}
}


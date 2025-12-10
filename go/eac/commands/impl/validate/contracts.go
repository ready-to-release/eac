// Command: validate contracts
// Short: Validate repository contracts against JSON schemas
// Long: Validate repository contracts against JSON schemas.
// Long:
// Long: This command validates all EAC repository configuration files against their
// Long: JSON Schema definitions. It checks:
// Long:   - repository.yml (optional)
// Long:   - modules.yml
// Long:   - environments.yml
// Long:   - testing-tags.yml
// Long:   - test-suites.yml
// Long:
// Long: Schema validation ensures configuration files are well-formed and contain
// Long: valid values according to the contract specifications.
// Long:
// Long: Example:
// Long:   validate contracts
package validate

import (
	"github.com/ready-to-release/eac/go/eac/commands/registry"
	"github.com/ready-to-release/eac/go/eac/core/config"
	"github.com/ready-to-release/eac/go/eac/core/logging"
)

var log = logging.C()

func init() {
	registry.Register(ValidateContracts)
}

// ValidateContracts validates all repository contracts against JSON schemas
func ValidateContracts() int {
	log.Info("Validating repository contracts...")
	log.Info("")

	// Load with schema validation enabled
	opts := config.LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        true, // We'll load each one individually for detailed reporting
	}

	cfg, err := config.Load(opts)
	if err != nil {
		log.Errorf("Error: %v", err)
		return 1
	}

	var hasErrors bool
	var validated int
	var total int

	// Validate repository.yml (includes modules)
	total++
	log.Infof("  %-25s ", "repository.yml")
	if err := cfg.LoadRepository(true); err != nil {
		log.Info("FAILED")
		log.Errorf("    %v", err)
		hasErrors = true
	} else if cfg.Repository != nil {
		log.Infof("OK (%d modules)", len(cfg.Repository.Modules))
		validated++
	}

	// Validate environments.yml
	total++
	log.Infof("  %-25s ", "environments.yml")
	if err := cfg.LoadEnvironments(true); err != nil {
		log.Info("FAILED")
		log.Errorf("    %v", err)
		hasErrors = true
	} else {
		log.Infof("OK (%d environments)", len(cfg.Environments.Environments))
		validated++

		// Additional semantic validation
		if err := cfg.Environments.Validate(); err != nil {
			log.Infof("    Warning: semantic validation: %v", err)
		}
	}

	// Validate testing-tags.yml
	total++
	log.Infof("  %-25s ", "testing-tags.yml")
	if err := cfg.LoadTestingTags(true); err != nil {
		log.Info("FAILED")
		log.Errorf("    %v", err)
		hasErrors = true
	} else {
		log.Infof("OK (%d tags, %d skip reasons)",
			len(cfg.TestingTags.Tags),
			len(cfg.TestingTags.SkipReasons))
		validated++
	}

	// Validate test-suites.yml
	total++
	log.Infof("  %-25s ", "test-suites.yml")
	if err := cfg.LoadTestSuites(true); err != nil {
		log.Info("FAILED")
		log.Errorf("    %v", err)
		hasErrors = true
	} else {
		log.Infof("OK (%d suites)", len(cfg.TestSuites.Suites))
		validated++
	}

	log.Info("")

	if hasErrors {
		log.Infof("Validation failed: %d/%d contracts valid", validated, total)
		return 1
	}

	log.Infof("All %d contracts validated successfully", validated)
	return 0
}

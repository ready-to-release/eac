package validate

import (
	"context"
	"os"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/clibase/flags"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/logging"
)

type validateContractsCommand struct{}

var _ core.SimpleCommandPort = (*validateContractsCommand)(nil)

func (c *validateContractsCommand) Name() string { return "validate contracts" }

func (c *validateContractsCommand) Metadata() core.CommandMetadata {
	return core.CommandMetadata{
		CanonicalName: "validate-contracts",
		Short:         "Validate repository contracts against JSON schemas",
		Long: "Validate repository contracts against JSON schemas.\n\nThis command validates all EAC repository configuration files against their\nJSON Schema definitions. It checks:\n  - repository.yml (optional)\n  - repository.yml\n  - environments.yml\n  - testing-tags.yml\n  - test-suites.yml\n\nSchema validation ensures configuration files are well-formed and contain\nvalid values according to the contract specifications.",
		Notes: "Expected Output:\n  Displays schema validation results for each contract file (repository.yml,\n  environments.yml, testing-tags.yml, test-suites.yml). Shows count of validated\n  items per file. Exit code 0 if all contracts valid, 1 if any validation errors.",
		Examples: []string{
			"eac validate contracts",
		},
	}
}

func (c *validateContractsCommand) Execute(_ context.Context, _ *core.CommandRequest) int {
	return ValidateContracts()
}

var log = logging.C()

// ValidateContracts validates all repository contracts against JSON schemas.
func ValidateContracts() int {
	// Validate flags against registry metadata
	if err := flags.ValidateFlagsFromRegistry(os.Args[2:]); err != nil {
		log.Errorf("%v", err)
		return 1
	}

	log.Info("Validating repository domain...")
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

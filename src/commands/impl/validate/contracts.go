// Command: validate contracts
// Short: Validate repository contracts against JSON schemas
// Long: Validate repository contracts against JSON schemas.
// Long:
// Long: This command validates all EAC repository configuration files against their
// Long: JSON Schema definitions. It checks:
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
// HasSideEffects: false
package validate

import (
	"fmt"
	"os"

	"github.com/ready-to-release/eac/src/commands/registry"
	"github.com/ready-to-release/eac/src/core/config"
)

func init() {
	registry.Register(ValidateContracts)
}

// ValidateContracts validates all repository contracts against JSON schemas
func ValidateContracts() int {
	fmt.Println("Validating repository contracts...")
	fmt.Println()

	// Load with schema validation enabled
	opts := config.LoadOptions{
		ValidateSchemas: true,
		LazyLoad:        true, // We'll load each one individually for detailed reporting
	}

	cfg, err := config.Load(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	var hasErrors bool
	var validated int

	// Validate modules.yml
	fmt.Printf("  %-25s ", "modules.yml")
	if err := cfg.LoadModules(true); err != nil {
		fmt.Printf("FAILED\n")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
		hasErrors = true
	} else {
		fmt.Printf("OK (%d modules)\n", len(cfg.Modules.Modules))
		validated++
	}

	// Validate environments.yml
	fmt.Printf("  %-25s ", "environments.yml")
	if err := cfg.LoadEnvironments(true); err != nil {
		fmt.Printf("FAILED\n")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
		hasErrors = true
	} else {
		fmt.Printf("OK (%d environments)\n", len(cfg.Environments.Environments))
		validated++

		// Additional semantic validation
		if err := cfg.Environments.Validate(); err != nil {
			fmt.Printf("    Warning: semantic validation: %v\n", err)
		}
	}

	// Validate testing-tags.yml
	fmt.Printf("  %-25s ", "testing-tags.yml")
	if err := cfg.LoadTestingTags(true); err != nil {
		fmt.Printf("FAILED\n")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
		hasErrors = true
	} else {
		fmt.Printf("OK (%d tags, %d skip reasons)\n",
			len(cfg.TestingTags.Tags),
			len(cfg.TestingTags.SkipReasons))
		validated++
	}

	// Validate test-suites.yml
	fmt.Printf("  %-25s ", "test-suites.yml")
	if err := cfg.LoadTestSuites(true); err != nil {
		fmt.Printf("FAILED\n")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
		hasErrors = true
	} else {
		fmt.Printf("OK (%d suites)\n", len(cfg.TestSuites.Suites))
		validated++
	}

	fmt.Println()

	if hasErrors {
		fmt.Printf("Validation failed: %d/4 contracts valid\n", validated)
		return 1
	}

	fmt.Printf("All %d contracts validated successfully\n", validated)
	return 0
}

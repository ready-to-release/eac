// Package config provides a central configuration loader for all EAC repository configs.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ready-to-release/eac/go/eac/core/contracts/schema"
)

// validateSchema validates data against a JSON schema.
func (c *EACConfig) validateSchema(schemaType schema.SchemaType, data []byte) error {
	c.validatorOnce.Do(func() {
		c.validator, c.validatorErr = schema.NewValidator(c.RepoRoot)
	})

	if c.validatorErr != nil {
		return fmt.Errorf("failed to initialize schema validator: %w", c.validatorErr)
	}

	return c.validator.ValidateYAML(schemaType, data)
}

// ValidateAll validates all loaded configs against their schemas.
func (c *EACConfig) ValidateAll() error {
	var errs []error

	// Repository config includes modules
	if c.Repository != nil {
		repoPath := filepath.Join(c.ConfigRoot, RepositoryFileName)
		if _, err := os.Stat(repoPath); err == nil {
			data, err := c.readConfigFile(RepositoryFileName)
			if err != nil {
				errs = append(errs, fmt.Errorf("repository: failed to read: %w", err))
			} else if err := c.validateSchema(schema.SchemaRepository, data); err != nil {
				errs = append(errs, fmt.Errorf("repository: %w", err))
			}
		}
	}

	if c.ModuleTypes != nil {
		data, err := c.readConfigFile(ModuleTypesFileName)
		if err != nil {
			errs = append(errs, fmt.Errorf("module-types: failed to read: %w", err))
		} else if err := c.validateSchema(schema.SchemaModuleTypes, data); err != nil {
			errs = append(errs, fmt.Errorf("module-types: %w", err))
		}
	}

	if c.Environments != nil {
		data, err := c.readConfigFile(EnvironmentsFileName)
		if err != nil {
			errs = append(errs, fmt.Errorf("environments: failed to read: %w", err))
		} else if err := c.validateSchema(schema.SchemaEnvironments, data); err != nil {
			errs = append(errs, fmt.Errorf("environments: %w", err))
		}
	}

	if c.TestingTags != nil {
		data, err := c.readConfigFile(TestingTagsFileName)
		if err != nil {
			errs = append(errs, fmt.Errorf("testing-tags: failed to read: %w", err))
		} else if err := c.validateSchema(schema.SchemaTestingTags, data); err != nil {
			errs = append(errs, fmt.Errorf("testing-tags: %w", err))
		}
	}

	if c.TestSuites != nil {
		data, err := c.readConfigFile(TestSuitesFileName)
		if err != nil {
			errs = append(errs, fmt.Errorf("test-suites: failed to read: %w", err))
		} else if err := c.validateSchema(schema.SchemaTestSuites, data); err != nil {
			errs = append(errs, fmt.Errorf("test-suites: %w", err))
		}
	}

	if c.SystemDependencies != nil {
		data, err := c.readConfigFile(SystemDependenciesFileName)
		if err != nil {
			errs = append(errs, fmt.Errorf("system-dependencies: failed to read: %w", err))
		} else if err := c.validateSchema(schema.SchemaSystemDependencies, data); err != nil {
			errs = append(errs, fmt.Errorf("system-dependencies: %w", err))
		}
	}

	if len(errs) > 0 {
		return &MultiError{Errors: errs}
	}

	return nil
}

// MultiError holds multiple errors.
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

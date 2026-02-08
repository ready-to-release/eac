package schema

import (
	"encoding/json"
	"fmt"
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// factoryValidator holds the compiled-once schema validator.
// The Validator is read-only after construction, so it can be shared directly
// without cloning.
var (
	factoryValidator     *Validator
	factoryValidatorOnce sync.Once
	factoryValidatorErr  error
)

// FactoryValidator returns the process-wide schema validator singleton.
// Compiled once from embedded contract schemas, safe for concurrent use.
// The returned Validator is read-only and must not be mutated.
func FactoryValidator() (*Validator, error) {
	factoryValidatorOnce.Do(func() {
		factoryValidator, factoryValidatorErr = compileFactoryValidator()
	})
	return factoryValidator, factoryValidatorErr
}

// compileFactoryValidator reads all schemas from core.FS and compiles them.
func compileFactoryValidator() (*Validator, error) {
	c := jsonschema.NewCompiler()

	v := &Validator{
		compiler: c,
		schemas:  make(map[SchemaType]*jsonschema.Schema),
	}

	for schemaType, fileName := range schemaFileNames {
		data, err := core.FS.ReadFile(core.SchemaPath(fileName))
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded schema %s: %w", fileName, err)
		}

		var schemaDoc any
		if err := json.Unmarshal(data, &schemaDoc); err != nil {
			return nil, fmt.Errorf("failed to parse schema %s: %w", fileName, err)
		}

		schemaURL := fmt.Sprintf("file:///%s", fileName)
		if err := c.AddResource(schemaURL, schemaDoc); err != nil {
			return nil, fmt.Errorf("failed to add schema resource %s: %w", fileName, err)
		}

		schema, err := c.Compile(schemaURL)
		if err != nil {
			return nil, fmt.Errorf("failed to compile schema %s: %w", fileName, err)
		}

		v.schemas[schemaType] = schema
	}

	return v, nil
}

// ResetFactoryValidatorForTesting resets the factory validator singleton.
func ResetFactoryValidatorForTesting() {
	factoryValidatorOnce = sync.Once{}
	factoryValidator = nil
	factoryValidatorErr = nil
}

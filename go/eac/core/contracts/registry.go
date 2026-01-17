package contracts

import (
	"fmt"
	"sync"
)

// ValidatorRegistry manages registered validators by name.
type ValidatorRegistry struct {
	mu         sync.RWMutex
	validators map[string]Validator
}

var globalRegistry = &ValidatorRegistry{
	validators: make(map[string]Validator),
}

// Register registers a validator with the given name
// This should be called from init() functions in validator implementations.
func Register(name string, validator Validator) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if _, exists := globalRegistry.validators[name]; exists {
		return fmt.Errorf("validator already registered: %s", name)
	}

	globalRegistry.validators[name] = validator
	return nil
}

// Get retrieves a validator by name.
func Get(name string) (Validator, error) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	validator, exists := globalRegistry.validators[name]
	if !exists {
		return nil, fmt.Errorf("validator not found: %s", name)
	}

	return validator, nil
}

// Has checks if a validator is registered.
func Has(name string) bool {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	_, exists := globalRegistry.validators[name]
	return exists
}

// List returns all registered validator names.
func List() []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	names := make([]string, 0, len(globalRegistry.validators))
	for name := range globalRegistry.validators {
		names = append(names, name)
	}
	return names
}

// Clear removes all registered validators (mainly for testing).
func Clear() {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	globalRegistry.validators = make(map[string]Validator)
}

package domain

import (
	"fmt"
	"sync"
)

// ValidatorRegistry manages registered validators by name.
// Use NewValidatorRegistry() to create isolated instances for testing,
// or use the package-level functions (Register, Get, Has, List, Clear)
// which delegate to the default global registry.
type ValidatorRegistry struct {
	mu         sync.RWMutex
	validators map[string]Validator
}

// NewValidatorRegistry creates a new, empty ValidatorRegistry.
// This is useful for tests that need an isolated registry without
// affecting the global default.
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{
		validators: make(map[string]Validator),
	}
}

// defaultRegistry is the package-level registry used by the package-level
// convenience functions (Register, Get, Has, List, Clear).
var defaultRegistry = NewValidatorRegistry()

// Register adds a validator to the registry under the given name.
// Returns an error if a validator with that name is already registered.
func (r *ValidatorRegistry) Register(name string, validator Validator) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.validators[name]; exists {
		return fmt.Errorf("validator already registered: %s", name)
	}

	r.validators[name] = validator
	return nil
}

// Get retrieves a validator by name.
// Returns an error if no validator is registered with that name.
func (r *ValidatorRegistry) Get(name string) (Validator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	validator, exists := r.validators[name]
	if !exists {
		return nil, fmt.Errorf("validator not found: %s", name)
	}

	return validator, nil
}

// Has checks if a validator is registered with the given name.
func (r *ValidatorRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.validators[name]
	return exists
}

// List returns all registered validator names.
func (r *ValidatorRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.validators))
	for name := range r.validators {
		names = append(names, name)
	}
	return names
}

// Clear removes all registered validators.
func (r *ValidatorRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.validators = make(map[string]Validator)
}

// Package-level convenience functions that delegate to defaultRegistry.

// Register registers a validator with the given name in the default registry.
// This should be called from init() functions in validator implementations.
func Register(name string, validator Validator) error {
	return defaultRegistry.Register(name, validator)
}

// Get retrieves a validator by name from the default registry.
func Get(name string) (Validator, error) {
	return defaultRegistry.Get(name)
}

// Has checks if a validator is registered in the default registry.
func Has(name string) bool {
	return defaultRegistry.Has(name)
}

// List returns all registered validator names from the default registry.
func List() []string {
	return defaultRegistry.List()
}

// Clear removes all registered validators from the default registry (mainly for testing).
func Clear() {
	defaultRegistry.Clear()
}

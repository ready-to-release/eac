package domain

// BaseValidator provides common validation utilities for domain.
type BaseValidator struct {
	contract *Contract
}

// NewBaseValidator creates a new base validator.
func NewBaseValidator(contract *Contract) *BaseValidator {
	return &BaseValidator{
		contract: contract,
	}
}

// GetContract returns the contract.
func (v *BaseValidator) GetContract() *Contract {
	return v.contract
}

// ValidationContext holds context information for validation.
type ValidationContext struct {
	// Custom context data that validators can use
	Data map[string]interface{}
}

// NewValidationContext creates a new validation context.
func NewValidationContext() *ValidationContext {
	return &ValidationContext{
		Data: make(map[string]interface{}),
	}
}

// Set sets a context value.
func (ctx *ValidationContext) Set(key string, value interface{}) {
	ctx.Data[key] = value
}

// Get gets a context value.
func (ctx *ValidationContext) Get(key string) (interface{}, bool) {
	val, ok := ctx.Data[key]
	return val, ok
}

// GetString gets a string context value.
func (ctx *ValidationContext) GetString(key string) string {
	if val, ok := ctx.Data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetStringSlice gets a string slice context value.
func (ctx *ValidationContext) GetStringSlice(key string) []string {
	if val, ok := ctx.Data[key]; ok {
		if slice, ok := val.([]string); ok {
			return slice
		}
	}
	return nil
}

// ToMap converts context to map[string]interface{} for compatibility.
func (ctx *ValidationContext) ToMap() map[string]interface{} {
	return ctx.Data
}

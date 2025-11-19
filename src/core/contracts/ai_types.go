package contracts

import "fmt"

// ContractType represents the category of contract
type ContractType string

const (
	ContractTypeAI     ContractType = "ai"     // AI-generated content (needs anti-corruption, validation)
	ContractTypeDomain ContractType = "domain" // Structure/metadata validation only
)

// Contract represents any validation contract
type Contract struct {
	Version     string                 `yaml:"version"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        ContractType           // Inferred from path, NOT from YAML
	RawData     map[string]interface{} // Full contract for custom access
}

// IsAI returns true if this is an AI contract
func (c *Contract) IsAI() bool {
	return c.Type == ContractTypeAI
}

// IsDomain returns true if this is a domain contract
func (c *Contract) IsDomain() bool {
	return c.Type == ContractTypeDomain
}

// AntiCorruptionRules represents noise filtering rules for AI output
type AntiCorruptionRules struct {
	Version             string                 `yaml:"version"`
	Name                string                 `yaml:"name"`
	Description         string                 `yaml:"description"`
	ForbiddenPrefixes   []string               `yaml:"forbidden_prefixes"`
	ForbiddenContains   []string               `yaml:"forbidden_contains"`
	ForbiddenEmojis     []string               `yaml:"forbidden_emojis"`
	NoiseHeaderKeywords []string               `yaml:"noise_header_keywords"`
	AgentSignatures     []string               `yaml:"agent_signatures"`
	RawData             map[string]interface{} // Full rules for custom access
}

// ValidationError represents a contract violation
type ValidationError struct {
	Code     string
	Message  string
	Line     int
	Severity string // "error" or "warning"
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%s] Line %d: %s", e.Code, e.Line, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Validator interface must be implemented by specific contract validators
type Validator interface {
	// Validate validates output against the contract
	Validate(output string, context map[string]interface{}) []ValidationError

	// VerifyImplementation verifies that the validator implements all contract rules
	VerifyImplementation() []ValidationError
}

// AIExecutor interface abstracts AI execution for contract-based generation
type AIExecutor interface {
	Execute(ctx interface{}, prompt string, opts ...interface{}) (string, error)
}

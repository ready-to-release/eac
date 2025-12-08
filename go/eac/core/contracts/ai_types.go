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

// AIConfig represents the unified AI configuration from ai-config.yml
type AIConfig struct {
	Version     string                   `yaml:"version"`
	Name        string                   `yaml:"name"`
	Description string                   `yaml:"description"`
	Defaults    AIDefaults               `yaml:"defaults"`
	Types       map[string]AITypeConfig  `yaml:"types"`
}

// AIDefaults contains default settings inherited by all AI types
type AIDefaults struct {
	AntiCorruption AntiCorruptionRules `yaml:"anti_corruption"`
}

// AITypeConfig represents configuration for a specific AI generation type
type AITypeConfig struct {
	Name           string                 `yaml:"name"`
	Description    string                 `yaml:"description"`
	Output         AIOutputConfig         `yaml:"output"`
	Validation     map[string]interface{} `yaml:"validation"`
	Data           map[string]string      `yaml:"data"`
	AntiCorruption AntiCorruptionRules    `yaml:"anti_corruption"`
}

// AIOutputConfig contains output parsing configuration
type AIOutputConfig struct {
	StartMarker *string `yaml:"start_marker"` // nil = immediate start
	Format      string  `yaml:"format"`       // text, dsl, gherkin, yaml, json, markdown
}

// AntiCorruptionRules represents noise filtering rules for AI output
type AntiCorruptionRules struct {
	ForbiddenPrefixes []string `yaml:"forbidden_prefixes"`
	ForbiddenContains []string `yaml:"forbidden_contains"`
	ForbiddenEmojis   []string `yaml:"forbidden_emojis"`
	AgentSignatures   []string `yaml:"agent_signatures"`
}

// Merge combines two AntiCorruptionRules, with override taking precedence for non-empty fields
// but appending arrays rather than replacing them
func (r AntiCorruptionRules) Merge(override AntiCorruptionRules) AntiCorruptionRules {
	result := AntiCorruptionRules{
		ForbiddenPrefixes: append([]string{}, r.ForbiddenPrefixes...),
		ForbiddenContains: append([]string{}, r.ForbiddenContains...),
		ForbiddenEmojis:   append([]string{}, r.ForbiddenEmojis...),
		AgentSignatures:   append([]string{}, r.AgentSignatures...),
	}
	result.ForbiddenPrefixes = append(result.ForbiddenPrefixes, override.ForbiddenPrefixes...)
	result.ForbiddenContains = append(result.ForbiddenContains, override.ForbiddenContains...)
	result.ForbiddenEmojis = append(result.ForbiddenEmojis, override.ForbiddenEmojis...)
	result.AgentSignatures = append(result.AgentSignatures, override.AgentSignatures...)
	return result
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

// AIExecutorWithProviderInfo extends AIExecutor with provider information retrieval.
// Implementations can optionally implement this interface to provide metadata about
// the AI provider used for generation.
type AIExecutorWithProviderInfo interface {
	AIExecutor
	// GetProviderName returns the name of the AI provider used for the last execution.
	// Returns empty string if no execution has occurred or provider info is unavailable.
	GetProviderName() string
}

package config

// Retry strategy types (from ai-config.schema.json).
const (
	StrategyStandard          = "standard"
	StrategyFocused           = "focused"
	StrategyEscalating        = "escalating"
	StrategyEscalatingFocused = "escalating-focused"
)

// AIConfig represents the unified AI configuration from ai-config.yml.
type AIConfig struct {
	Version     string                  `yaml:"version"`
	Name        string                  `yaml:"name"`
	Description string                  `yaml:"description"`
	Defaults    AIDefaults              `yaml:"defaults"`
	Types       map[string]AITypeConfig `yaml:"types"`
}

// AIDefaults contains default settings inherited by all AI types
// Currently unused but reserved for future use.
type AIDefaults struct{}

// AITypeConfig represents configuration for a specific AI generation type
// AI generates structured output in OutputFormat, then commands handle formatting.
type AITypeConfig struct {
	Name          string               `yaml:"name"`
	Description   string               `yaml:"description"`
	OutputFormat  string               `yaml:"output_format"`  // Required: json, gherkin, oscal, structurizr, plaintext
	Data          map[string]string    `yaml:"data"`           // Optional: reference data files to inject
	RetryStrategy *RetryStrategyConfig `yaml:"retry_strategy"` // Optional: retry strategy
}

// RetryStrategyConfig represents retry strategy configuration from ai-config.yml.
type RetryStrategyConfig struct {
	Type            string   `yaml:"type"`             // standard, focused, escalating, escalating-focused
	FocusCategories []string `yaml:"focus_categories"` // For focused/escalating-focused strategies
	MaxAttempts     int      `yaml:"max_attempts"`     // Maximum retry attempts (1-10)
}

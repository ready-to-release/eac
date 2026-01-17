package contracts

// ContractType represents the category of contract.
type ContractType string

const (
	ContractTypeAI     ContractType = "ai"     // AI-generated content (validated against JSON schema)
	ContractTypeDomain ContractType = "domain" // Structure/metadata validation only
)

// Contract represents any validation contract.
type Contract struct {
	Version     string                 `yaml:"version"`
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        ContractType           // Inferred from path, NOT from YAML
	RawData     map[string]interface{} // Full contract for custom access
}

// IsAI returns true if this is an AI contract.
func (c *Contract) IsAI() bool {
	return c.Type == ContractTypeAI
}

// IsDomain returns true if this is a domain contract.
func (c *Contract) IsDomain() bool {
	return c.Type == ContractTypeDomain
}

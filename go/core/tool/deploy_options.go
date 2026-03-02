package tool

// DeployOptions holds options for a deploy operation.
type DeployOptions struct {
	// Environment is the target environment moniker (e.g., "development", "production").
	Environment string

	// DryRun when true runs a what-if deployment without applying changes.
	DryRun bool

	// Component optionally restricts deployment to a specific component within the module.
	Component string

	// DeployConfig holds environment-specific deploy configuration resolved from environments.yml.
	DeployConfig *DeployEnvironmentConfig
}

// DeployEnvironmentConfig holds deploy-specific configuration from an environment.
type DeployEnvironmentConfig struct {
	// SubscriptionID is the Azure subscription ID for bicep deployments.
	SubscriptionID string

	// TenantID is the Azure tenant ID.
	TenantID string

	// ResourceGroup is the Azure resource group name.
	ResourceGroup string

	// Location is the Azure region (e.g., "northeurope").
	Location string

	// Env holds additional environment variables to pass to the deployer.
	Env map[string]string
}

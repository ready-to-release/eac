package testing

import (
	"github.com/ready-to-release/eac/go/eac/core/contracts"
	"github.com/ready-to-release/eac/go/eac/core/contracts/modules"
)

// RegistryOption is a function that configures a mock registry.
type RegistryOption func(*modules.Registry)

// ModuleOption is a function that configures a mock module.
type ModuleOption func(*contracts.BaseContract)

// NewMockRegistry creates a new mock module registry for testing.
func NewMockRegistry(opts ...RegistryOption) *modules.Registry {
	registry := modules.NewRegistry("1.0.0", "/mock/workspace")

	for _, opt := range opts {
		opt(registry)
	}

	return registry
}

// WithModule adds a module to the mock registry with optional configuration.
func WithModule(moniker string, opts ...ModuleOption) RegistryOption {
	return func(r *modules.Registry) {
		base := contracts.BaseContract{
			Moniker:    moniker,
			Components: make(contracts.ModuleComponents),
		}

		// Apply module options
		for _, opt := range opts {
			opt(&base)
		}

		module := modules.NewModuleContract(base, r.WorkspaceRoot())
		_ = r.Add(module) // Ignore error in mock context
	}
}

// WithVersioning adds versioning configuration to a module with default CalVer scheme.
func WithVersioning() ModuleOption {
	return func(m *contracts.BaseContract) {
		m.Versioning = &contracts.ModuleVersioning{
			Scheme: "CalVer", // Default to CalVer
		}
	}
}

// WithSemver adds semver versioning configuration to a module.
func WithSemver() ModuleOption {
	return func(m *contracts.BaseContract) {
		m.Versioning = &contracts.ModuleVersioning{
			Scheme: "SemVer",
		}
	}
}

// WithChangelog sets the changelog path for a module.
func WithChangelog(path string) ModuleOption {
	return func(m *contracts.BaseContract) {
		if m.Versioning == nil {
			m.Versioning = &contracts.ModuleVersioning{
				Scheme: "CalVer",
			}
		}
		m.Versioning.Changelog = path
	}
}

// WithDependsOn adds dependencies to a module.
func WithDependsOn(dependencies ...string) ModuleOption {
	return func(m *contracts.BaseContract) {
		m.DependsOn = append(m.DependsOn, dependencies...)
	}
}

// WithComponent adds a component to a module with specified root path.
func WithComponent(componentType, root string) ModuleOption {
	return func(m *contracts.BaseContract) {
		if m.Components == nil {
			m.Components = make(contracts.ModuleComponents)
		}
		m.Components[componentType] = &contracts.ComponentEntry{
			Root: root,
		}
	}
}

// WithGoComponent adds a Go component to a module.
func WithGoComponent(root string) ModuleOption {
	return WithComponent("go", root)
}

// WithDocsComponent adds a docs component to a module.
func WithDocsComponent(root string) ModuleOption {
	return WithComponent("docs", root)
}

// WithSpecsComponent adds a specs component to a module.
func WithSpecsComponent(root string) ModuleOption {
	return WithComponent("specs", root)
}

// WithReleaseType sets the release type for a module (published, internal, bundle, none).
func WithReleaseType(releaseType string) ModuleOption {
	return func(m *contracts.BaseContract) {
		if m.Versioning == nil {
			m.Versioning = &contracts.ModuleVersioning{
				Scheme: "CalVer",
			}
		}
		m.Versioning.ReleaseType = releaseType
	}
}

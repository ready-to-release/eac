package testing

import (
	"github.com/ready-to-release/eac/go/eac/core/domain"
	"github.com/ready-to-release/eac/go/eac/core/domain/modules"
)

// RegistryOption is a function that configures a mock registry.
type RegistryOption func(*modules.Registry)

// ModuleOption is a function that configures a mock module.
type ModuleOption func(*domain.BaseContract)

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
		base := domain.BaseContract{
			Moniker:    moniker,
			Components: make(domain.ModuleComponents),
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
	return func(m *domain.BaseContract) {
		m.Versioning = &domain.ModuleVersioning{
			Scheme: "CalVer", // Default to CalVer
		}
	}
}

// WithSemver adds semver versioning configuration to a module.
func WithSemver() ModuleOption {
	return func(m *domain.BaseContract) {
		m.Versioning = &domain.ModuleVersioning{
			Scheme: "SemVer",
		}
	}
}

// WithChangelog sets the changelog path for a module.
func WithChangelog(path string) ModuleOption {
	return func(m *domain.BaseContract) {
		if m.Versioning == nil {
			m.Versioning = &domain.ModuleVersioning{
				Scheme: "CalVer",
			}
		}
		m.Versioning.Changelog = path
	}
}

// WithDependsOn adds dependencies to a module.
func WithDependsOn(dependencies ...string) ModuleOption {
	return func(m *domain.BaseContract) {
		m.DependsOn = append(m.DependsOn, dependencies...)
	}
}

// WithComponent adds a component to a module with specified root path.
func WithComponent(componentType, root string) ModuleOption {
	return func(m *domain.BaseContract) {
		if m.Components == nil {
			m.Components = make(domain.ModuleComponents)
		}
		m.Components[componentType] = &domain.ComponentEntry{
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
	return func(m *domain.BaseContract) {
		if m.Versioning == nil {
			m.Versioning = &domain.ModuleVersioning{
				Scheme: "CalVer",
			}
		}
		m.Versioning.ReleaseType = releaseType
	}
}

// Package resolver provides unified component-to-tool resolution.
// This file provides the port interface implementation for dependency injection.
package resolver

import (
	"sync"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/adapters"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// Compile-time interface check.
var _ core.UnitResolverPort = (*ResolverPort)(nil)

// ResolverPort adapts ComponentResolver to implement core.UnitResolverPort.
// This is the port interface for dependency injection and clean architecture.
type ResolverPort struct {
	resolver *ComponentResolver
}

// NewResolverPort creates a new port-compatible resolver.
func NewResolverPort() *ResolverPort {
	return &ResolverPort{
		resolver: NewComponentResolver(),
	}
}

// NewResolverPortWithResolver wraps an existing ComponentResolver.
func NewResolverPortWithResolver(r *ComponentResolver) *ResolverPort {
	return &ResolverPort{resolver: r}
}

// ResolveForBuild returns work units for buildable components in a module.
// Implements core.UnitResolverPort.
func (p *ResolverPort) ResolveForBuild(module core.ModuleContractPort, cached map[string]bool) []core.UnitSpecPort {
	// Convert interface to concrete type
	concrete, ok := unwrapModuleContract(module)
	if !ok {
		return nil
	}

	// Resolve using concrete implementation
	specs := p.resolver.ResolveForBuild(concrete, cached)

	// Convert back to port interfaces
	return adapters.AdaptUnitSpecs(specs)
}

// ResolveForLint returns work units for lintable components in a module.
// Implements core.UnitResolverPort.
func (p *ResolverPort) ResolveForLint(module core.ModuleContractPort, cached map[string]bool) []core.UnitSpecPort {
	// Convert interface to concrete type
	concrete, ok := unwrapModuleContract(module)
	if !ok {
		return nil
	}

	// Resolve using concrete implementation
	specs := p.resolver.ResolveForLint(concrete, cached)

	// Convert back to port interfaces
	return adapters.AdaptUnitSpecs(specs)
}

// ResolveForScan returns work units for scannable components in a module.
// Implements core.UnitResolverPort.
func (p *ResolverPort) ResolveForScan(module core.ModuleContractPort, categories []string, cached map[string]bool) []core.UnitSpecPort {
	// Convert interface to concrete type
	concrete, ok := unwrapModuleContract(module)
	if !ok {
		return nil
	}

	// Convert string categories to typed categories
	var scanCats []ScanCategory
	for _, c := range categories {
		scanCats = append(scanCats, ScanCategory(c))
	}

	// Resolve using concrete implementation
	specs := p.resolver.ResolveForScan(concrete, scanCats, cached)

	// Convert back to port interfaces
	return adapters.AdaptUnitSpecs(specs)
}

// unwrapModuleContract extracts the concrete *modules.ModuleContract from an interface.
// Returns nil, false if the module cannot be unwrapped.
func unwrapModuleContract(module core.ModuleContractPort) (*modules.ModuleContract, bool) {
	if module == nil {
		return nil, false
	}

	// Direct type assertion (most common case)
	if m, ok := module.(*modules.ModuleContract); ok {
		return m, true
	}

	// Check for adapter with Unwrap method
	type unwrapper interface {
		Unwrap() *modules.ModuleContract
	}
	if u, ok := module.(unwrapper); ok {
		return u.Unwrap(), true
	}

	return nil, false
}

// ============================================================================
// Global Singleton
// ============================================================================

var (
	globalResolverPort     *ResolverPort
	globalResolverPortOnce sync.Once
	globalResolverPortMu   sync.Mutex
)

// GlobalUnitResolver returns the global unit resolver port instance.
// This is the recommended entry point for dependency injection.
func GlobalUnitResolver() core.UnitResolverPort {
	globalResolverPortOnce.Do(func() {
		globalResolverPort = NewResolverPort()
	})
	return globalResolverPort
}

// ResetGlobalUnitResolver resets the global resolver (for testing only).
func ResetGlobalUnitResolver() {
	globalResolverPortMu.Lock()
	defer globalResolverPortMu.Unlock()
	globalResolverPort = nil
	globalResolverPortOnce = sync.Once{}
}

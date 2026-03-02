package tool

import (
	"sync"

	deploy "github.com/ready-to-release/eac/contracts/runner/0.1.0/deploy"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
)

// DeployBridge provides a unified interface for resolving deploy handlers.
// Mirrors BuildBridge but for deployer tools.
type DeployBridge struct {
	mu sync.RWMutex

	// Native handlers (registered from adapters)
	nativeHandlers map[string]deploy.DeployerPort

	// Tool system integration
	registry Registry
	executor Executor
}

// NewDeployBridge creates a new deploy bridge.
func NewDeployBridge() *DeployBridge {
	return &DeployBridge{
		nativeHandlers: make(map[string]deploy.DeployerPort),
	}
}

// RegisterNativeHandler registers a native deploy handler.
// Native handlers take precedence over tool-config.yml handlers.
func (b *DeployBridge) RegisterNativeHandler(h deploy.DeployerPort) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nativeHandlers[h.Name()] = h
}

// SetToolSystem configures the tool system for tool-config.yml defined tools.
func (b *DeployBridge) SetToolSystem(registry Registry, resolver *DefaultResolver, executor Executor) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.registry = registry
	b.executor = executor
}

// ComponentDeployHandler pairs a component name with its deploy handler.
type ComponentDeployHandler struct {
	Component string
	Handler   deploy.DeployerPort
}

// GetHandlersForModule returns deploy handlers for all deployable components in a module.
func (b *DeployBridge) GetHandlersForModule(module *modules.ModuleContract) []ComponentDeployHandler {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if module == nil {
		return nil
	}

	cfg := config.Global()
	if cfg == nil || cfg.ComponentKinds == nil {
		return nil
	}

	var result []ComponentDeployHandler
	for _, compName := range module.GetEnabledComponents() {
		compTypeName := module.Components.GetComponentType(compName)
		compType := cfg.ComponentKinds.Get(compTypeName)
		if compType == nil || !compType.IsDeployable() {
			continue
		}

		deployerName := compType.GetDeployers()[0]
		if h := b.getHandlerUnlocked(deployerName); h != nil {
			result = append(result, ComponentDeployHandler{
				Component: compName,
				Handler:   h,
			})
		}
	}

	return result
}

// GetHandler returns a deploy handler by name.
// Checks native handlers first, then falls back to tool registry.
func (b *DeployBridge) GetHandler(name string) deploy.DeployerPort {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.getHandlerUnlocked(name)
}

func (b *DeployBridge) getHandlerUnlocked(name string) deploy.DeployerPort {
	if h, ok := b.nativeHandlers[name]; ok {
		return h
	}
	// Tool registry adapter (container-based deployers)
	if b.registry != nil && b.executor != nil {
		if tool, ok := b.registry.Get(name); ok {
			return NewDeployToolHandlerAdapter(tool, b.executor)
		}
	}
	return nil
}

// Global deploy bridge instance.
var (
	globalDeployBridge     *DeployBridge
	globalDeployBridgeOnce sync.Once
)

// GlobalDeployBridge returns the global deploy bridge instance.
func GlobalDeployBridge() *DeployBridge {
	globalDeployBridgeOnce.Do(func() {
		globalDeployBridge = NewDeployBridge()
	})
	return globalDeployBridge
}

package resolver

import (
	"math"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
	"github.com/ready-to-release/eac/go/core/resource"
	"github.com/ready-to-release/eac/go/core/tool"
	"github.com/ready-to-release/eac/go/core/workunit"
)

// ComponentResolver resolves module components to executable work units.
// It is the single source of truth for component-to-tool mapping across
// build, lint, and scan operations.
type ComponentResolver struct {
	cfg         *config.EACConfig
	buildBridge *tool.BuildBridge
}

// NewComponentResolver creates a resolver from global config.
func NewComponentResolver() *ComponentResolver {
	return &ComponentResolver{
		cfg:         config.Global(),
		buildBridge: tool.GlobalBuildBridge(),
	}
}

// NewComponentResolverWithConfig creates a resolver with explicit config.
// Useful for testing.
func NewComponentResolverWithConfig(cfg *config.EACConfig, buildBridge *tool.BuildBridge) *ComponentResolver {
	return &ComponentResolver{
		cfg:         cfg,
		buildBridge: buildBridge,
	}
}

// ResolveForBuild returns UnitSpecs for all buildable components in a module.
// Respects build_after dependencies and module-level overrides.
func (r *ComponentResolver) ResolveForBuild(module *modules.ModuleContract, cachedModules map[string]bool) []workunit.UnitSpec {
	if module == nil || r.cfg == nil {
		return nil
	}

	var specs []workunit.UnitSpec

	// Build enabled components map
	enabledComponents := r.getEnabledComponentsMap(module)

	// Build set of components that will be scheduled (have builders or tool chains)
	// This prevents deadlocks from build_after dependencies on non-buildable components
	scheduledComponents := make(map[string]bool)
	for compName, compType := range enabledComponents {
		typeConfig := r.cfg.ComponentKinds.Get(compType)
		if typeConfig == nil || !typeConfig.IsBuildable() {
			continue
		}

		// For tool chain types, check if the first tool has a handler
		if typeConfig.HasToolChain() {
			builders := typeConfig.GetBuilders()
			if len(builders) > 0 {
				handler := r.buildBridge.GetHandler(builders[0])
				if handler != nil {
					scheduledComponents[compName] = true
				}
			}
			continue
		}

		// For single-builder types, use GetHandlerForComponent
		handler := r.buildBridge.GetHandlerForComponent(compType)
		if handler == nil {
			continue
		}
		scheduledComponents[compName] = true
	}

	// Create dependency graph for build_after ordering
	depGraph := NewDependencyGraph(module.Moniker, enabledComponents, r.getBuildAfterFunc())

	for compName, compType := range enabledComponents {
		if !scheduledComponents[compName] {
			continue
		}

		// Get component type config
		typeConfig := r.cfg.ComponentKinds.Get(compType)

		// Build external dependencies from dependency graph and component depends_on
		// These are added to the FIRST tool in a tool chain, or the single tool
		var externalDeps []toolChainDep
		for _, depComp := range depGraph.DependsOn(compName) {
			if !scheduledComponents[depComp] {
				continue
			}
			externalDeps = append(externalDeps, toolChainDep{
				Module:        module.Moniker,
				ComponentType: enabledComponents[depComp],
				Component:     depComp,
				Tool:          r.resolveActualToolName(enabledComponents[depComp], PhaseBuild),
			})
		}

		// Add intra-module dependencies from component config (depends_on field)
		compDependsOn := module.Components.GetComponentDependsOn(compName)
		for _, depComp := range compDependsOn {
			if !scheduledComponents[depComp] {
				continue
			}
			alreadyAdded := false
			for _, existing := range externalDeps {
				if existing.Component == depComp {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				externalDeps = append(externalDeps, toolChainDep{
					Module:        module.Moniker,
					ComponentType: enabledComponents[depComp],
					Component:     depComp,
					Tool:          r.resolveActualToolName(enabledComponents[depComp], PhaseBuild),
				})
			}
		}

		// Build metadata from component config (book, theme, etc.)
		metadata := r.getComponentMetadata(module, compName)

		// Check if this is a tool chain (multiple tools)
		if typeConfig != nil && typeConfig.HasToolChain() {
			// Expand tool chain into multiple UnitSpecs
			toolChainSpecs := expandToolChain(module.Moniker, compName, compType, typeConfig.GetBuilders(), externalDeps)
			for _, tcSpec := range toolChainSpecs {
				// Get handler for this specific tool
				handler := r.buildBridge.GetHandler(tcSpec.Tool)
				isContainer := handler != nil && handler.IsContainer()
				weight := r.getToolWeight(module.Moniker, compName, compType)
				poolAlloc := resource.AllocationForWeight(weight, isContainer)

				spec := workunit.UnitSpec{
					ID: workunit.UnitID{
						Action:        core.ActionBuild,
						Module:        module.Moniker,
						ComponentType: compType,
						ComponentName: compName,
						Tool:          tcSpec.Tool,
					},
					ComponentType:  compType,
					Weight:         weight,
					PoolAllocation: poolAlloc,
					DependsOn:      tcSpec.DependsOn,
					Cached:         cachedModules != nil && cachedModules[module.Moniker],
					Metadata:       metadata,
				}
				specs = append(specs, spec)
			}
		} else {
			// Single tool - use existing logic
			handler := r.buildBridge.GetHandlerForComponent(compType)
			actualToolName := ""
			if handler != nil {
				actualToolName = handler.Name()
			}

			// Convert externalDeps to []workunit.UnitID
			var dependsOn []workunit.UnitID
			for _, dep := range externalDeps {
				dependsOn = append(dependsOn, workunit.UnitID{
					Action:        core.ActionBuild,
					Module:        dep.Module,
					ComponentType: enabledComponents[dep.Component],
					ComponentName: dep.Component,
					Tool:          dep.Tool,
				})
			}

			isContainer := handler.IsContainer()
			weight := r.getWeight(module.Moniker, compName, compType, core.ActionBuild)
			poolAlloc := resource.AllocationForWeight(weight, isContainer)

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionBuild,
					Module:        module.Moniker,
					ComponentType: compType,
					ComponentName: compName,
					Tool:          actualToolName,
				},
				ComponentType:  compType,
				Weight:         weight,
				PoolAllocation: poolAlloc,
				DependsOn:      dependsOn,
				Cached:         cachedModules != nil && cachedModules[module.Moniker],
				Metadata:       metadata,
			}
			specs = append(specs, spec)
		}
	}

	return specs
}

// ResolveForLint returns UnitSpecs for all lintable components in a module.
func (r *ComponentResolver) ResolveForLint(module *modules.ModuleContract, cachedModules map[string]bool) []workunit.UnitSpec {
	if module == nil || r.cfg == nil || r.cfg.LintProviders == nil {
		return nil
	}

	var specs []workunit.UnitSpec

	// Build enabled components map
	enabledComponents := r.getEnabledComponentsMap(module)

	for compName, compType := range enabledComponents {
		// Get lint provider tool names for this component type
		toolNames := r.cfg.LintProviders.GetProvidersForComponentType(compType)
		if len(toolNames) == 0 {
			continue
		}

		for _, toolName := range toolNames {
			handler := r.buildBridge.GetHandler(toolName)
			isContainer := handler != nil && handler.IsContainer()
			weight := r.getWeight(module.Moniker, compName, compType, core.ActionLint)

			poolAlloc := resource.AllocationForWeight(weight, isContainer)

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionLint,
					Module:        module.Moniker,
					ComponentType: compType,
					ComponentName: compName,
					Tool:          toolName,
				},
				ComponentType:  compType,
				Weight:         weight,
				PoolAllocation: poolAlloc,
				DependsOn:      []workunit.UnitID{},
				Cached:         cachedModules != nil && cachedModules[module.Moniker],
				Metadata:       make(map[string]any),
			}

			specs = append(specs, spec)
		}
	}

	return specs
}

// ResolveForScan returns UnitSpecs for all scannable components in a module.
// Unlike build/lint which produce one UOW per component, scan produces
// multiple UOWs per component (one per scanner category).
func (r *ComponentResolver) ResolveForScan(module *modules.ModuleContract, scanCategories []ScanCategory, cachedModules map[string]bool) []workunit.UnitSpec {
	if module == nil || r.cfg == nil {
		return nil
	}

	var specs []workunit.UnitSpec

	// Build enabled components map
	enabledComponents := r.getEnabledComponentsMap(module)

	for compName, compType := range enabledComponents {
		typeConfig := r.cfg.ComponentKinds.Get(compType)
		if typeConfig == nil || !typeConfig.IsScannable() {
			continue
		}

		// Get scanner categories for this component type
		categories := r.getScannerCategories(typeConfig, scanCategories)

		for _, category := range categories {
			// Resolve specific tool for this scanner category
			toolName := r.resolveScannerTool(compType, category, typeConfig)
			if toolName == "" {
				continue
			}

			weight := r.getScanWeight(module.Moniker, compName, toolName)
			// All scanners run in containers - set PoolAllocation explicitly
			poolAlloc := resource.ContainerAllocation(weight, weight)

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Action:        core.ActionScan,
					Module:        module.Moniker,
					ComponentType: compType,
					ComponentName: compName,
					Tool:          toolName,
					Extra:         map[string]string{"category": string(category)},
				},
				ComponentType:  compType,
				Weight:         weight,
				PoolAllocation: poolAlloc,
				DependsOn:      []workunit.UnitID{},
				Cached:         cachedModules != nil && cachedModules[module.Moniker],
				Metadata: map[string]any{
					"scanner_category": string(category),
				},
			}

			specs = append(specs, spec)
		}
	}

	return specs
}

// getEnabledComponentsMap returns a map of component name -> component type
// for all enabled components in the module.
func (r *ComponentResolver) getEnabledComponentsMap(module *modules.ModuleContract) map[string]string {
	result := make(map[string]string)
	for _, compName := range module.GetEnabledComponents() {
		compType := module.Components.GetComponentType(compName)
		result[compName] = compType
	}
	return result
}

// getBuildAfterFunc returns a function that retrieves build_after for a component type.
func (r *ComponentResolver) getBuildAfterFunc() func(compType string) []string {
	return func(compType string) []string {
		if r.cfg == nil || r.cfg.ComponentKinds == nil {
			return nil
		}
		typeConfig := r.cfg.ComponentKinds.Get(compType)
		if typeConfig == nil {
			return nil
		}
		return typeConfig.GetBuildAfter()
	}
}

// resolveToolForPhase returns the tool name for a component type and phase.
// Priority: component-type tools.phase.default > component-type builder > empty
func (r *ComponentResolver) resolveToolForPhase(compType string, phase Phase, typeConfig *config.ComponentType) string {
	if typeConfig == nil {
		return ""
	}

	switch phase {
	case PhaseBuild:
		if builders := typeConfig.GetBuilders(); len(builders) > 0 {
			return builders[0]
		}
		return ""
	default:
		return ""
	}
}

// resolveToolForComponent is a helper that looks up the tool for a component type and phase.
func (r *ComponentResolver) resolveToolForComponent(compType string, phase Phase) string {
	if r.cfg == nil || r.cfg.ComponentKinds == nil {
		return ""
	}
	typeConfig := r.cfg.ComponentKinds.Get(compType)
	return r.resolveToolForPhase(compType, phase, typeConfig)
}

// resolveActualToolName resolves the actual handler name for a component type and phase.
// This uses the handler's Name() method to get the real tool ID (e.g., "mkdocs-build" instead of "mkdocs").
func (r *ComponentResolver) resolveActualToolName(compType string, phase Phase) string {
	// Use GetHandlerForComponent which uses the resolver with proper bindings
	handler := r.buildBridge.GetHandlerForComponent(compType)
	if handler != nil {
		return handler.Name()
	}
	return ""
}

// getScannerCategories returns the scanner categories to run.
// If requestedCategories is empty, uses defaults from component type config.
func (r *ComponentResolver) getScannerCategories(typeConfig *config.ComponentType, requestedCategories []ScanCategory) []ScanCategory {
	if len(requestedCategories) > 0 {
		// Filter to only categories supported by this component type
		supported := make(map[ScanCategory]bool)
		for _, s := range typeConfig.GetScanners() {
			supported[ScanCategory(s)] = true
		}

		var result []ScanCategory
		for _, cat := range requestedCategories {
			if supported[cat] {
				result = append(result, cat)
			}
		}
		return result
	}

	// Use component type defaults
	scanners := typeConfig.GetScanners()
	if len(scanners) == 0 {
		return nil
	}

	categories := make([]ScanCategory, len(scanners))
	for i, s := range scanners {
		categories[i] = ScanCategory(s)
	}
	return categories
}

// resolveScannerTool returns the specific tool for a scanner category.
// E.g., category "sbom" -> "trivy-sbom", category "sast" -> "semgrep"
func (r *ComponentResolver) resolveScannerTool(compType string, category ScanCategory, typeConfig *config.ComponentType) string {
	// Default scanner tool mappings
	defaultTools := map[ScanCategory]string{
		ScanCategorySBOM:       "trivy-sbom",
		ScanCategoryVuln:       "trivy-vuln",
		ScanCategorySecrets:    "trivy-secrets",
		ScanCategorySAST:       "semgrep",
		ScanCategoryIAC:        "trivy-iac",
		ScanCategoryCompliance: "trivy-compliance",
		ScanCategoryZAP:        "zap",
	}

	// Return default tool for the category
	if toolName, ok := defaultTools[category]; ok {
		return toolName
	}

	return ""
}

// getToolWeight calculates scheduling weight for a specific tool in a tool chain.
// Uses the component type's resources.cpus as the base weight.
func (r *ComponentResolver) getToolWeight(moniker, compName, compType string) int {
	// Use component type's weight (from resources.cpus in component-types.yml)
	// Tool chains share the component type's weight - they run as a single UOW
	return r.getWeight(moniker, compName, compType, core.ActionBuild)
}

// getWeight calculates scheduling weight for a component.
// Weight is determined by: tool.Resources.CPUs * componentType.Amp * module.amp
// The base weight comes from the tool definition, with amplifiers applied from
// component type configuration and module-level overrides.
func (r *ComponentResolver) getWeight(moniker, compName, compType string, op core.ActionType) int {
	// Get base weight from tool resources (default 1)
	baseWeight := 1
	if toolDef := r.buildBridge.GetToolForComponent(compType); toolDef != nil {
		if toolDef.Resources != nil && toolDef.Resources.CPUs > 0 {
			baseWeight = toolDef.Resources.CPUs
		}
	}

	// Fall back to component type's weight if tool didn't provide one
	// This handles native handlers (e.g., pdf, site) that aren't in the tool registry
	var typeConfig *config.ComponentType
	if r.cfg != nil && r.cfg.ComponentKinds != nil {
		typeConfig = r.cfg.ComponentKinds.Get(compType)
		if baseWeight == 1 && typeConfig != nil {
			if typeWeight := typeConfig.GetWeight(); typeWeight > 1 {
				baseWeight = typeWeight
			}
		}
	}

	// Apply component-type amplifier if configured
	compTypeAmp := 1.0
	if typeConfig != nil {
		compTypeAmp = typeConfig.GetAmp()
	}

	// Apply module-level amplifier if configured
	moduleAmp := 1.0
	if r.cfg != nil && r.cfg.Repository != nil {
		if module, ok := r.cfg.Repository.GetModule(moniker); ok {
			moduleAmp = module.GetComponentAmp(compName, string(op))
		}
	}

	// Calculate final weight: baseWeight * compTypeAmp * moduleAmp
	weight := int(math.Ceil(float64(baseWeight) * compTypeAmp * moduleAmp))
	if weight < 1 {
		weight = 1
	}
	return weight
}

// getScanWeight returns the scheduling weight for a scanner tool.
// Different scanners have different resource requirements.
// Weight = base scanner weight × module-level amp.
func (r *ComponentResolver) getScanWeight(moniker, compName, toolName string) int {
	// Scanner weights based on resource requirements
	baseWeight := 1
	switch toolName {
	case "semgrep":
		// SAST can be CPU-intensive on large codebases
		baseWeight = 2
	case "trivy-vuln":
		// Vulnerability scanning involves network requests for DB updates
		baseWeight = 2
	case "trivy-sbom":
		// SBOM generation is relatively lightweight
		baseWeight = 1
	case "trivy-secrets":
		// Secret scanning is relatively fast
		baseWeight = 1
	case "trivy-iac":
		// IaC scanning is moderate
		baseWeight = 1
	case "trivy-compliance":
		// Compliance scanning is moderate
		baseWeight = 1
	case "zap":
		// DAST requires running services, high resource
		baseWeight = 3
	}

	// Apply module-level amplifier if configured
	amp := 1.0
	if r.cfg.Repository != nil {
		if module, ok := r.cfg.Repository.GetModule(moniker); ok {
			amp = module.GetComponentAmp(compName, "scan")
		}
	}

	weight := int(math.Ceil(float64(baseWeight) * amp))
	if weight < 1 {
		weight = 1
	}
	return weight
}

// getComponentMetadata extracts component config values as metadata.
// This passes values like book name, theme, etc. to handlers.
func (r *ComponentResolver) getComponentMetadata(module *modules.ModuleContract, compName string) map[string]any {
	metadata := make(map[string]any)

	// Get component entry from module
	if module.Components == nil {
		return metadata
	}

	// Extract config values that handlers need
	if book := module.Components.GetComponentConfig(compName, "book"); book != "" {
		metadata["book"] = book
	}
	if theme := module.Components.GetComponentConfig(compName, "theme"); theme != "" {
		metadata["theme"] = theme
	}

	return metadata
}

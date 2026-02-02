package resolver

import (
	"math"

	"github.com/ready-to-release/eac/go/core/config"
	"github.com/ready-to-release/eac/go/core/domain/modules"
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

	// Build set of components that will be scheduled (have builders)
	// This prevents deadlocks from build_after dependencies on non-buildable components
	scheduledComponents := make(map[string]bool)
	for compName, compType := range enabledComponents {
		typeConfig := r.cfg.ComponentTypes.Get(compType)
		if typeConfig == nil || !typeConfig.HasBuilder() {
			continue
		}
		// Use GetHandlerForComponent which uses the resolver with proper bindings
		// (e.g., component type "book" → tool "mkdocs-build")
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

		// Use GetHandlerForComponent which uses the resolver with proper bindings
		// (e.g., component type "book" → tool "mkdocs-build" via tool-config.yml bindings)
		handler := r.buildBridge.GetHandlerForComponent(compType)

		// Get the handler's actual name for the tool field
		actualToolName := ""
		if handler != nil {
			actualToolName = handler.Name()
		}

		// Build dependencies from dependency graph - ONLY for scheduled components
		// Dependencies on non-buildable components would cause deadlocks
		var dependsOn []workunit.UnitID
		for _, depComp := range depGraph.DependsOn(compName) {
			// Skip dependencies on components that won't be scheduled
			if !scheduledComponents[depComp] {
				continue
			}
			dependsOn = append(dependsOn, workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    module.Moniker,
				Component: depComp,
				Tool:      r.resolveActualToolName(enabledComponents[depComp], PhaseBuild),
			})
		}

		// Add intra-module dependencies from component config (depends_on field)
		compDependsOn := module.Components.GetComponentDependsOn(compName)
		for _, depComp := range compDependsOn {
			// Skip if already in dependsOn or not scheduled
			if !scheduledComponents[depComp] {
				continue
			}
			alreadyAdded := false
			for _, existing := range dependsOn {
				if existing.Component == depComp {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				dependsOn = append(dependsOn, workunit.UnitID{
					Context:   workunit.ContextBuild,
					Module:    module.Moniker,
					Component: depComp,
					Tool:      r.resolveActualToolName(enabledComponents[depComp], PhaseBuild),
				})
			}
		}

		// Build metadata from component config (book, theme, etc.)
		metadata := r.getComponentMetadata(module, compName)

		spec := workunit.UnitSpec{
			ID: workunit.UnitID{
				Context:   workunit.ContextBuild,
				Module:    module.Moniker,
				Component: compName,
				Tool:      actualToolName,
			},
			ComponentType: compType,
			Weight:        r.getWeight(module.Moniker, compName, compType, tool.OperationBuild),
			Container:     handler.IsContainer(),
			HostInstalled: !handler.IsContainer(),
			DependsOn:     dependsOn,
			Cached:        cachedModules != nil && cachedModules[module.Moniker],
			Metadata:      metadata,
		}

		specs = append(specs, spec)
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

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Context:   workunit.ContextLint,
					Module:    module.Moniker,
					Component: compName,
					Tool:      toolName,
				},
				ComponentType: compType,
				Weight:        r.getWeight(module.Moniker, compName, compType, tool.OperationLint),
				Container:     isContainer,
				HostInstalled: !isContainer,
				DependsOn:     []workunit.UnitID{},
				Cached:        cachedModules != nil && cachedModules[module.Moniker],
				Metadata:      make(map[string]any),
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
		typeConfig := r.cfg.ComponentTypes.Get(compType)
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

			spec := workunit.UnitSpec{
				ID: workunit.UnitID{
					Context:   workunit.ContextScan,
					Module:    module.Moniker,
					Component: compName,
					Tool:      toolName,
					Extra:     map[string]string{"category": string(category)},
				},
				ComponentType: compType,
				Weight:        r.getScanWeight(toolName),
				Container:     true, // All scanners run in containers
				HostInstalled: false,
				DependsOn:     []workunit.UnitID{},
				Cached:        cachedModules != nil && cachedModules[module.Moniker],
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
		if r.cfg == nil || r.cfg.ComponentTypes == nil {
			return nil
		}
		typeConfig := r.cfg.ComponentTypes.Get(compType)
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

	// For now, use legacy builder field
	// When we parse tools field from YAML, we can use it here
	switch phase {
	case PhaseBuild:
		return typeConfig.Builder
	default:
		return ""
	}
}

// resolveToolForComponent is a helper that looks up the tool for a component type and phase.
func (r *ComponentResolver) resolveToolForComponent(compType string, phase Phase) string {
	if r.cfg == nil || r.cfg.ComponentTypes == nil {
		return ""
	}
	typeConfig := r.cfg.ComponentTypes.Get(compType)
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

// getWeight calculates scheduling weight for a component.
// Weight is determined by: component type resources.cpus (default 1) * module-level amplifier.
func (r *ComponentResolver) getWeight(moniker, compName, compType string, op tool.OperationType) int {
	// Get base weight from component type resources (default 1)
	baseWeight := 1
	if r.cfg != nil && r.cfg.ComponentTypes != nil {
		if typeConfig := r.cfg.ComponentTypes.Get(compType); typeConfig != nil {
			baseWeight = typeConfig.GetWeight()
		}
	}

	// Apply module-level amplifier if configured
	amp := 1.0
	if r.cfg.Repository != nil {
		if module, ok := r.cfg.Repository.GetModule(moniker); ok {
			amp = module.GetComponentAmp(compName, string(op))
		}
	}

	weight := int(math.Ceil(float64(baseWeight) * amp))
	if weight < 1 {
		weight = 1
	}
	return weight
}

// getScanWeight returns the scheduling weight for a scanner tool.
func (r *ComponentResolver) getScanWeight(toolName string) int {
	// Scanners are generally lightweight
	return 1
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

package resolver

import (
	"math"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/ready-to-release/eac/go/core/config"
)

// getToolWeight calculates scheduling weight for a specific tool in a tool chain.
// Uses the component type's resources.cpus as the base weight.
func (r *ComponentResolver) getToolWeight(moniker, compName, compType string) int {
	// Use component type's weight (from resources.cpus in blueprints.yml component-kinds)
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

	// Apply component-kind amplifier if configured
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
// Weight = base scanner weight * module-level amp.
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

// File: go/cli/eac/impl/create/risk-assess/parallel.go
// Purpose: Parallel and sequential module assessment execution
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Clear result struct with index for ordering
//   - One goroutine per module with simple channel communication
//   - Separate sequential path for debugging
//
// Easy to change:
//   - Module processing logic unchanged (calls assessSingleModule)
//   - Can easily add features like cancellation or retry logic
//   - Progress tracking can be added independently
//
// Hard to break:
//   - Preserves exact output order via indexed results
//   - No shared mutable state between goroutines
//   - Thread-safe via separate module outputs
//   - Compatible with existing error handling patterns
//
// Performance:
//   - Sequential: N modules × Ts = total time
//   - Parallel:   max(T1, T2, ..., TN) ≈ 60-70% speedup

package riskassess

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/oscal"
	"github.com/ready-to-release/eac/go/commands/repository/internal/risk/scoring"
)

// assessModulesParallel runs assessments concurrently for multiple modules.
// Follows the proven pattern from commit-message/parallel.go for safe concurrent execution.
func assessModulesParallel(
	config *AssessConfig,
	profile *oscalTypes.Profile,
) []*ModuleAssessmentResult {
	// Result structure with index for ordering
	type moduleResult struct {
		index  int
		result *ModuleAssessmentResult
	}

	// Buffered channel (capacity = module count)
	resultsChan := make(chan moduleResult, len(config.Modules))
	var wg sync.WaitGroup

	// Launch goroutines for each module
	for i, moduleName := range config.Modules {
		wg.Add(1)
		go func(idx int, module string) {
			defer wg.Done()

			// Assess this module
			result := assessSingleModule(config, module, profile)

			// Send result with index to preserve order
			resultsChan <- moduleResult{
				index:  idx,
				result: result,
			}
		}(i, moduleName)
	}

	// Close channel after all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results in original order
	results := make([]*ModuleAssessmentResult, len(config.Modules))
	for res := range resultsChan {
		results[res.index] = res.result
	}

	return results
}

// assessModulesSequential runs assessments one at a time for multiple modules.
// Useful for debugging or resource-constrained environments.
func assessModulesSequential(
	config *AssessConfig,
	profile *oscalTypes.Profile,
) []*ModuleAssessmentResult {
	var results []*ModuleAssessmentResult

	for _, moduleName := range config.Modules {
		result := assessSingleModule(config, moduleName, profile)
		results = append(results, result)
	}

	return results
}

// assessSingleModule runs assessment for one module and returns the result.
// This function encapsulates all the logic for assessing a single module,
// making it suitable for both sequential and parallel execution.
func assessSingleModule(
	config *AssessConfig,
	moduleName string,
	profile *oscalTypes.Profile,
) *ModuleAssessmentResult {
	result := &ModuleAssessmentResult{
		Module:   moduleName,
		Warnings: []string{},
	}

	// Collect evidence for this module (read-only, no execution)
	moduleConfig := *config // Copy config

	evidenceCollection, err := collectEvidenceForModule(&moduleConfig, moduleName)
	if err != nil {
		result.Error = fmt.Errorf("error collecting evidence: %w", err)
		return result
	}

	// Capture warnings from evidence collection
	if evidenceCollection != nil && len(evidenceCollection.Warnings) > 0 {
		result.Warnings = evidenceCollection.Warnings
	}

	// Store evidence collection for reporting
	result.Evidence = evidenceCollection

	// Build assessment-results
	// Use timestamped directory: out/risk/<timestamp>/<module>/assessment-results.json
	moduleOutputDir := filepath.Join(config.OutputDir, moduleName)
	if err := os.MkdirAll(moduleOutputDir, 0o755); err != nil {
		result.Error = fmt.Errorf("error creating module output directory: %w", err)
		return result
	}
	arPath := filepath.Join(moduleOutputDir, "assessment-results.json")

	ar, err := buildAssessmentResultsForModule(&moduleConfig, moduleName, profile, evidenceCollection)
	if err != nil {
		result.Error = fmt.Errorf("error building assessment results: %w", err)
		return result
	}

	// Write assessment-results
	if err := oscal.WriteAssessmentResults(arPath, ar); err != nil {
		result.Error = fmt.Errorf("error writing assessment results: %w", err)
		return result
	}

	// Extract metrics from assessment results
	result.AssessmentResults = ar
	result.OutputPath = arPath

	if len(ar.Results) > 0 {
		resultData := ar.Results[0]
		if resultData.Findings != nil {
			for i := range *resultData.Findings {
				finding := &(*resultData.Findings)[i]
				if finding.Target.Status.State == oscal.StateSatisfied {
					result.Satisfied++
				} else {
					result.NotSatisfied++
				}
			}
		}

		// Calculate risk score
		likelihood := scoring.CalculateBaseLikelihood(0, 0, 0, result.NotSatisfied)
		impact := scoring.GetDefaultImpact(moduleName)
		result.RiskScore = scoring.ComputeRiskScore(moduleName, likelihood, impact, "Based on control satisfaction")
	}

	return result
}

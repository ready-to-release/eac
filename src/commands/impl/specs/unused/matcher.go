package unused

// UnusedStep represents a step definition that is not used by any feature file.
type UnusedStep struct {
	StepDefinition
	PairImplDir string // which impl directory this belongs to
}

// AnalysisResult contains the results of unused step analysis.
type AnalysisResult struct {
	Pairs           []PairResult
	SharedSteps     []StepDefinition // steps from internal/steps.go
	UnusedShared    []UnusedStep     // shared steps not used by ANY pair
	TotalUnused     int
	TotalSteps      int
	TotalFeatures   int
}

// PairResult contains unused step analysis for a single impl↔specs pair.
type PairResult struct {
	Pair              ImplSpecsPair
	ModuleSteps       []StepDefinition // steps defined in this module
	FeatureSteps      []FeatureStep    // steps used in feature files
	UnusedModuleSteps []UnusedStep     // module steps not used
	UnusedSharedSteps []UnusedStep     // shared steps not used by THIS pair
}

// FindUnusedSteps analyzes all pairs and finds unused step definitions.
func FindUnusedSteps(pairs []ImplSpecsPair, sharedStepsFile string) (*AnalysisResult, error) {
	result := &AnalysisResult{}

	// Parse shared steps if file exists
	var sharedSteps []StepDefinition
	if sharedStepsFile != "" {
		var err error
		sharedSteps, err = ParseStepDefinitions([]string{sharedStepsFile}, true)
		if err != nil {
			// Not fatal - shared steps file might not exist
			sharedSteps = nil
		}
	}
	result.SharedSteps = sharedSteps

	// Track which shared steps are used by at least one pair
	sharedStepUsage := make(map[string]bool) // pattern -> used

	// Analyze each pair
	for _, pair := range pairs {
		pairResult, err := analyzePair(pair, sharedSteps)
		if err != nil {
			return nil, err
		}

		// Track shared step usage
		for _, step := range sharedSteps {
			if !isStepUnused(step, pairResult.FeatureSteps) {
				sharedStepUsage[step.Pattern] = true
			}
		}

		result.Pairs = append(result.Pairs, pairResult)
		result.TotalUnused += len(pairResult.UnusedModuleSteps)
		result.TotalSteps += len(pairResult.ModuleSteps)
		result.TotalFeatures += len(pair.FeatureFiles)
	}

	// Find globally unused shared steps (not used by ANY pair)
	for _, step := range sharedSteps {
		if !sharedStepUsage[step.Pattern] {
			result.UnusedShared = append(result.UnusedShared, UnusedStep{
				StepDefinition: step,
				PairImplDir:    "shared",
			})
			result.TotalUnused++
		}
	}
	result.TotalSteps += len(sharedSteps)

	return result, nil
}

// analyzePair analyzes a single impl↔specs pair for unused steps.
func analyzePair(pair ImplSpecsPair, sharedSteps []StepDefinition) (PairResult, error) {
	result := PairResult{Pair: pair}

	// Parse module-specific step definitions
	moduleSteps, err := ParseStepDefinitions(pair.StepFiles, false)
	if err != nil {
		return result, err
	}
	result.ModuleSteps = moduleSteps

	// Parse feature steps
	featureSteps, err := ParseFeatureFiles(pair.FeatureFiles)
	if err != nil {
		return result, err
	}
	result.FeatureSteps = featureSteps

	// Find unused module steps
	for _, step := range moduleSteps {
		if isStepUnused(step, featureSteps) {
			result.UnusedModuleSteps = append(result.UnusedModuleSteps, UnusedStep{
				StepDefinition: step,
				PairImplDir:    pair.ImplDir,
			})
		}
	}

	// Find shared steps unused by THIS pair (if pair uses internal)
	if pair.UsesInternal {
		for _, step := range sharedSteps {
			if isStepUnused(step, featureSteps) {
				result.UnusedSharedSteps = append(result.UnusedSharedSteps, UnusedStep{
					StepDefinition: step,
					PairImplDir:    pair.ImplDir,
				})
			}
		}
	}

	return result, nil
}

// isStepUnused checks if a step definition is not matched by any feature step.
func isStepUnused(stepDef StepDefinition, featureSteps []FeatureStep) bool {
	if stepDef.CompileErr != nil {
		// Can't match if regex doesn't compile - consider it unused
		return true
	}

	for _, featureStep := range featureSteps {
		if MatchesStep(stepDef, featureStep) {
			return false
		}
	}

	return true
}

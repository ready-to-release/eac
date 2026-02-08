package docprep

import (
	"fmt"
	"time"
)

// Phase is a single preprocessing step.
// Phases are pure functions of their context: they read inputs,
// transform staging files, and update shared state.
type Phase interface {
	// Name returns a human-readable name for logging.
	Name() string

	// Execute runs the phase. It may read and write files in
	// StagingDir and update shared state in PreprocessContext.
	Execute(pctx *PreprocessContext) error
}

// PhaseFunc adapts a function to the Phase interface.
type PhaseFunc struct {
	PhaseName string
	Fn        func(pctx *PreprocessContext) error
}

// Name returns the phase name.
func (p PhaseFunc) Name() string { return p.PhaseName }

// Execute runs the phase function.
func (p PhaseFunc) Execute(pctx *PreprocessContext) error { return p.Fn(pctx) }

// Pipeline orchestrates sequential phase execution.
type Pipeline struct {
	phases []Phase
}

// NewPipeline creates an empty pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Add appends a phase to the pipeline.
func (p *Pipeline) Add(phase Phase) *Pipeline {
	p.phases = append(p.phases, phase)
	return p
}

// AddFunc appends a function phase to the pipeline.
func (p *Pipeline) AddFunc(name string, fn func(*PreprocessContext) error) *Pipeline {
	p.phases = append(p.phases, PhaseFunc{PhaseName: name, Fn: fn})
	return p
}

// Phases returns the list of phases for inspection.
func (p *Pipeline) Phases() []Phase {
	return p.phases
}

// Execute runs all phases in order, stopping on first error.
func (p *Pipeline) Execute(pctx *PreprocessContext) error {
	pctx.Log.Infof("Book preprocessing: %s (%s mode)", pctx.Book.Name, pctx.Mode.Name())
	start := time.Now()

	for _, phase := range p.phases {
		pctx.Log.Infof("  Phase: %s", phase.Name())

		if err := pctx.Ctx.Err(); err != nil {
			return fmt.Errorf("cancelled before %s: %w", phase.Name(), err)
		}

		if err := phase.Execute(pctx); err != nil {
			return fmt.Errorf("phase %s: %w", phase.Name(), err)
		}
	}

	elapsed := time.Since(start)
	pctx.Log.Infof("Preprocessing complete: %s (took %v)", pctx.Book.Name, elapsed)

	if pctx.WarnAsError && len(pctx.Warnings) > 0 {
		for i, w := range pctx.Warnings {
			pctx.Log.Infof("  Warning %d: %s", i+1, w)
		}
		return fmt.Errorf("preprocessing failed with %d warning(s)", len(pctx.Warnings))
	}

	return nil
}

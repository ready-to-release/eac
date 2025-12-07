// Package testing provides test fixture pooling for performance optimization.
//
// The FixturePool manages reusable test environment templates that can be
// quickly copied instead of created from scratch, dramatically reducing
// test isolation overhead.
//
// Key features:
//   - Creates base template ONCE per feature file
//   - Fast directory copy (~50ms) vs full setup (~300-900ms)
//   - Automatic cleanup of template and temp directories
//   - Thread-safe for parallel test execution
//
// Performance Impact:
//   - Without fixture pool: 0.3-0.9s per scenario (git init + copies)
//   - With fixture pool: 0.05s per scenario (fast copy only)
//   - Savings: 77-90% reduction in isolation overhead
//
// Usage:
//
//	// At feature start
//	pool := testing.NewFixturePool()
//	template, err := pool.CreateTemplate(originalRepoRoot)
//	if err != nil {
//	    return err
//	}
//	defer pool.Cleanup()
//
//	// Per scenario
//	isolation, err := pool.NewIsolationFromTemplate(template)
//	if err != nil {
//	    return err
//	}
//	defer isolation.Cleanup()
package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FixturePool manages reusable test environment templates.
// Templates are created once and can be quickly copied for each scenario.
type FixturePool struct {
	mu           sync.Mutex
	templates    map[string]*FixtureTemplate // Key: originalRepoRoot
	tempDirs     []string                    // Track all temp dirs for cleanup
	cleanedUp    bool
}

// FixtureTemplate represents a reusable test environment template.
// Contains a fully initialized git repository with all static resources.
type FixtureTemplate struct {
	TemplateDir      string // Path to template directory
	OriginalRepoRoot string // Original repo root (for reference)
}

// NewFixturePool creates a new fixture pool.
func NewFixturePool() *FixturePool {
	return &FixturePool{
		templates: make(map[string]*FixtureTemplate),
		tempDirs:  make([]string, 0),
	}
}

// CreateTemplate creates a new fixture template for the given repository root.
// The template includes:
//   - Initialized git repository with initial commit
//   - Copied contracts/ directory
//   - Copied AI configs
//   - Copied mkdocs configs
//   - Test AI configuration
//
// This is expensive (~300-900ms) but only happens ONCE per test suite.
// Template is cached and reused for all scenarios in the suite.
func (p *FixturePool) CreateTemplate(originalRepoRoot string) (*FixtureTemplate, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if template already exists for this repo root
	if template, exists := p.templates[originalRepoRoot]; exists {
		return template, nil
	}

	// Create the template using standard TestIsolation.Setup()
	// This does all the expensive setup work: git init, copy contracts, etc.
	isolation := NewTestIsolation().
		WithOriginalRepoRoot(originalRepoRoot).
		WithCopyContracts(true).
		WithCopyAIContracts(true).
		WithCopyMkdocsConfig(true).
		WithMockAIConfig(true)

	if err := isolation.Setup(); err != nil {
		return nil, fmt.Errorf("failed to create fixture template: %w", err)
	}

	template := &FixtureTemplate{
		TemplateDir:      isolation.IsolatedDir(),
		OriginalRepoRoot: originalRepoRoot,
	}

	p.templates[originalRepoRoot] = template
	p.tempDirs = append(p.tempDirs, template.TemplateDir)

	return template, nil
}

// NewIsolationFromTemplate creates a new TestIsolation by fast-copying a template.
// This is much faster (~50ms) than creating from scratch (~300-900ms).
//
// The returned TestIsolation is fully functional and can be used like any other.
// Call isolation.Cleanup() when done to remove the temporary directory.
func (p *FixturePool) NewIsolationFromTemplate(template *FixtureTemplate) (*TestIsolation, error) {
	if template == nil {
		return nil, fmt.Errorf("template is nil - cannot create isolation")
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	// Create temp directory for this scenario
	tmpDir, err := os.MkdirTemp("", "isolated-test-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Fast copy template to temp directory (~50ms)
	if err := copyDir(template.TemplateDir, tmpDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("failed to copy template: %w", err)
	}

	// Create TestIsolation wrapper with all fields from template
	isolation := &TestIsolation{
		originalRepoRoot:   template.OriginalRepoRoot,
		isolatedDir:        tmpDir,
		cleanedUp:          false,
		copyContracts:      true,
		copyAIContracts:    true,
		copyMkdocsConfig:   true,
		createMockAIConfig: true,
	}

	p.tempDirs = append(p.tempDirs, tmpDir)

	return isolation, nil
}

// GetTemplate returns an existing template for the given repo root.
// Returns nil if no template exists - caller must create template first.
func (p *FixturePool) GetTemplate(originalRepoRoot string) *FixtureTemplate {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.templates[originalRepoRoot]
}

// Cleanup removes all template and temporary directories created by this pool.
// Safe to call multiple times.
func (p *FixturePool) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cleanedUp {
		return
	}

	// Remove all temp directories (templates + scenario copies)
	for _, dir := range p.tempDirs {
		os.RemoveAll(dir)
	}

	p.templates = make(map[string]*FixtureTemplate)
	p.tempDirs = make([]string, 0)
	p.cleanedUp = true
}

// TrackTempDir adds a directory to the cleanup list without creating a template.
// This is used for temp directories created outside the fixture pool.
func (p *FixturePool) TrackTempDir(dir string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tempDirs = append(p.tempDirs, dir)
}

// copyDirFast performs a fast directory copy optimized for fixture templates.
// This is the same as copyDir but could be optimized with hard links or copy-on-write
// in the future for even better performance.
func copyDirFast(src, dst string) error {
	// For now, use the standard copyDir
	// Future optimization: use os.Link() for hard links where supported
	// or ioctl FICLONE for copy-on-write on Linux/macOS
	return copyDir(src, dst)
}

// FixtureStats returns statistics about the fixture pool.
type FixtureStats struct {
	TemplateCount int
	TempDirCount  int
}

// Stats returns statistics about the fixture pool.
func (p *FixturePool) Stats() FixtureStats {
	p.mu.Lock()
	defer p.mu.Unlock()
	return FixtureStats{
		TemplateCount: len(p.templates),
		TempDirCount:  len(p.tempDirs),
	}
}

// EstimatePerformanceGain calculates the estimated time savings from using fixtures.
// Returns savings in seconds for the given number of scenarios.
func EstimatePerformanceGain(scenarioCount int) (withoutFixtures, withFixtures, savings float64) {
	const (
		setupTimeMin    = 0.3  // Minimum setup time per scenario (seconds)
		setupTimeMax    = 0.9  // Maximum setup time per scenario (seconds)
		copyTime        = 0.05 // Fast copy time per scenario (seconds)
		templateCreation = 0.6  // Average template creation time (once)
	)

	// Average setup time without fixtures
	avgSetup := (setupTimeMin + setupTimeMax) / 2.0
	withoutFixtures = avgSetup * float64(scenarioCount)

	// Time with fixtures: template creation + (copies per scenario)
	withFixtures = templateCreation + (copyTime * float64(scenarioCount))

	savings = withoutFixtures - withFixtures
	return
}

// Helper function to calculate size of a directory (for diagnostics)
func calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

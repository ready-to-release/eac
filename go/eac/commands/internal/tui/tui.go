// Package tui provides a text user interface for displaying build and test output.
package tui

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/stream"
)

// Default TUI configuration values
const (
	// DefaultHeight is the default TUI console height in rows (3-20)
	DefaultHeight = 15
)

// Config configures the console window.
type Config struct {
	Height     int  // Total height (default: DefaultHeight)
	ShowHeader bool // Show status header (default: true)
	BufferSize int  // Line buffer size (default: 1000)
}

// Console manages the TUI console window for build/test output.
type Console struct {
	config     Config
	lineChan   chan console.Line
	statusChan chan console.Status
	program    *tea.Program

	mu      sync.Mutex
	started bool
	stopped bool
	ready   chan struct{} // Signals when TUI is ready

	// Track multi-writers for cleanup
	writers []*stream.MultiWriter
}

// New creates a new console with the given configuration.
func New(config Config) *Console {
	if config.Height <= 0 {
		config.Height = 5
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	return &Console{
		config:     config,
		lineChan:   make(chan console.Line, 100),
		statusChan: make(chan console.Status, 10),
		ready:      make(chan struct{}),
	}
}

// Start starts the TUI program. This blocks until the program exits.
// Use StartAsync for non-blocking start.
func (c *Console) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return nil
	}
	c.started = true
	c.mu.Unlock()

	model := console.NewModel(
		c.config.Height,
		c.config.ShowHeader,
		c.lineChan,
		c.statusChan,
	)

	// Prevent lipgloss from querying terminal background color (causes OSC escape leaks)
	lipgloss.SetHasDarkBackground(true)

	// Use inline mode (no alt screen) so output persists after TUI exits
	// Disable bracketed paste to prevent escape sequence leaks
	c.program = tea.NewProgram(model,
		tea.WithoutBracketedPaste(),
	)

	// Signal that TUI is ready
	close(c.ready)

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		c.Stop()
	}()

	_, err := c.program.Run()
	return err
}

// StartAsync starts the TUI program in a goroutine.
// Waits for TUI to be ready before returning. Call Stop to stop the program.
func (c *Console) StartAsync(ctx context.Context) {
	go func() {
		_ = c.Start(ctx)
	}()

	// Wait for TUI to be ready (or timeout after 1 second)
	select {
	case <-c.ready:
	case <-time.After(1 * time.Second):
	}
}

// Stop stops the TUI program.
func (c *Console) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	c.mu.Unlock()

	// Close all multi-writers first
	c.mu.Lock()
	for _, w := range c.writers {
		w.Close()
	}
	c.writers = nil
	c.mu.Unlock()

	// Give TUI a moment to render final state
	time.Sleep(100 * time.Millisecond)

	// Close channels (this will trigger TUI to exit)
	close(c.lineChan)
	close(c.statusChan)

	// Quit the program and wait for it to fully exit
	if c.program != nil {
		c.program.Quit()
		c.program.Wait()
	}

	// Reset terminal state - clear any lingering escape sequences
	// \033[0m = reset attributes, \033[?25h = show cursor
	fmt.Print("\033[0m\033[?25h")
}

// NewWriter creates an io.Writer for a module's output.
// Output written to this writer will appear in the console.
// The returned writer also writes to the provided log writer.
func (c *Console) NewWriter(source string, logWriter io.Writer) io.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()

	mw := stream.NewMultiWriter(c.lineChan, source, logWriter)
	c.writers = append(c.writers, mw)
	return mw
}

// UpdateStatus sends a status update to the console.
func (c *Console) UpdateStatus(status console.Status) {
	c.mu.Lock()
	stopped := c.stopped
	c.mu.Unlock()

	if stopped {
		return
	}

	select {
	case c.statusChan <- status:
	default:
		// Drop if channel full
	}
}

// SendLine directly sends a line to the console.
func (c *Console) SendLine(line console.Line) {
	c.mu.Lock()
	stopped := c.stopped
	c.mu.Unlock()

	if stopped {
		return
	}

	select {
	case c.lineChan <- line:
	default:
		// Drop if channel full
	}
}

// SendInfo sends an info line to the console.
func (c *Console) SendInfo(source, text string) {
	c.SendLine(console.Line{
		Text:   text,
		Source: source,
		Level:  console.LevelInfo,
	})
}

// SendWarn sends a warning line to the console.
func (c *Console) SendWarn(source, text string) {
	c.SendLine(console.Line{
		Text:   text,
		Source: source,
		Level:  console.LevelWarn,
	})
}

// SendError sends an error line to the console.
func (c *Console) SendError(source, text string) {
	c.SendLine(console.Line{
		Text:   text,
		Source: source,
		Level:  console.LevelError,
	})
}

// SetPhase switches to a new phase (Init, Run, End)
func (c *Console) SetPhase(phase Phase) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	// Send phase update directly to Bubbletea program
	program.Send(console.PhaseUpdateMsg{
		Phase:  phase,
		Status: console.PhaseActive,
	})
}

// SetPhaseSummary sets the summary text for a collapsed phase
func (c *Console) SetPhaseSummary(phase Phase, summary string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.PhaseUpdateMsg{
		Phase:   phase,
		Summary: summary,
	})
}

// CompletePhase marks a phase as complete with a summary
func (c *Console) CompletePhase(phase Phase, success bool, summary string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	status := console.PhaseComplete
	if !success {
		status = console.PhaseFailed
	}

	program.Send(console.PhaseUpdateMsg{
		Phase:   phase,
		Status:  status,
		Summary: summary,
	})
}

// WriteToPhase writes a line to a specific phase's buffer
func (c *Console) WriteToPhase(phase Phase, text string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	// Send line to specific phase buffer via Bubbletea
	program.Send(console.PhaseLineMsg{
		Phase: phase,
		Line: console.Line{
			Text:   text,
			Source: phase.String(),
			Level:  console.LevelInfo,
		},
	})
}

// Status is an alias for console.Status for public use.
type Status = console.Status

// Line is an alias for console.Line for public use.
type Line = console.Line

// Level is an alias for console.Level for public use.
type Level = console.Level

// Phase is an alias for console.Phase for public use.
type Phase = console.Phase

// PhaseStatus is an alias for console.PhaseStatus for public use.
type PhaseStatus = console.PhaseStatus

// Level constants for public use.
const (
	LevelInfo  = console.LevelInfo
	LevelWarn  = console.LevelWarn
	LevelError = console.LevelError
)

// Phase constants for public use.
const (
	PhaseInit = console.PhaseInit
	PhaseRun  = console.PhaseRun
	PhaseEnd  = console.PhaseEnd
)

// PhaseStatus constants for public use.
const (
	PhasePending  = console.PhasePending
	PhaseActive   = console.PhaseActive
	PhaseComplete = console.PhaseComplete
	PhaseFailed   = console.PhaseFailed
)

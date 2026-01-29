// Package tui provides a text user interface for displaying build and test output.
package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/console"
	"github.com/ready-to-release/eac/go/eac/commands/internal/tui/stream"
)

// Default TUI configuration values.
const (
	// DefaultHeight is the initial TUI height before WindowSizeMsg arrives.
	// In alt-screen mode, the actual terminal height is used instead.
	DefaultHeight = 40
)

// Config configures the console window.
type Config struct {
	Height       int    // Total height (default: DefaultHeight)
	BufferSize   int    // Line buffer size (default: 1000)
	RunPhaseName string // Custom name for Run phase (e.g., "building", "testing")
	ASCIIMode    bool   // Use ASCII-only characters (default: false, use --ascii for compatibility)
	SkipTUIDelay bool // Skip TUI exit delay (exit immediately when done)
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
	done    chan struct{} // Signals when Start() has fully completed (including printSummary)

	// Track multi-writers for cleanup
	writers []*stream.MultiWriter

	// Store final model state for post-exit summary
	finalModel *console.Model
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
		done:       make(chan struct{}),
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

	// Signal completion when Start() returns (after all cleanup including printSummary)
	defer func() {
		close(c.done)
	}()

	model := console.NewModel(
		c.config.Height,
		c.config.RunPhaseName,
		c.lineChan,
		c.statusChan,
		c.config.ASCIIMode,
		c.config.SkipTUIDelay,
	)

	// Prevent lipgloss from querying terminal background color (causes OSC escape leaks)
	lipgloss.SetHasDarkBackground(true)

	// Use alt screen mode to take over terminal, then restore on exit
	// Disable bracketed paste to prevent escape sequence leaks
	// Disable signal handler so our custom handler can catch Ctrl-C
	// Start with mouse mode enabled for scrolling
	c.program = tea.NewProgram(model,
		tea.WithAltScreen(), // Take over screen, restore on exit
		tea.WithoutBracketedPaste(),
		tea.WithoutSignalHandler(), // Let our custom signal handler catch Ctrl-C
		tea.WithMouseAllMotion(),  // Enable mouse for scrolling and hover
	)

	// Signal that TUI is ready
	close(c.ready)

	// Set up signal handler for Ctrl-C (SIGINT)
	// This ensures single Ctrl-C triggers immediate cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Ensure cleanup always runs on exit (normal, Ctrl-C, or error)
	defer func() {
		c.mu.Lock()
		if !c.stopped {
			c.stopped = true

			// Close all multi-writers first
			for _, w := range c.writers {
				w.Close()
			}
			c.writers = nil

			// Close channels to trigger TUI exit
			close(c.lineChan)
			close(c.statusChan)
		}
		c.mu.Unlock()

		// Reset terminal state - always run this to ensure clean terminal
		// \033[0m = reset attributes, \033[?25h = show cursor
		fmt.Print("\033[0m\033[?25h")
	}()

	go func() {
		select {
		case <-sigChan:
			// On Ctrl-C, quit the TUI immediately
			if c.program != nil {
				c.program.Quit()
			}
		case <-ctx.Done():
			// On context cancellation, quit the TUI
			if c.program != nil {
				c.program.Quit()
			}
		}
	}()

	// Run the TUI and capture final model state
	finalModel, err := c.program.Run()
	if err != nil {
		return err
	}

	// Store final model for post-exit summary
	if m, ok := finalModel.(console.Model); ok {
		c.mu.Lock()
		c.finalModel = &m
		c.mu.Unlock()

		// Print plain-text summary after alt screen is restored
		c.printSummary(&m)
	}

	return nil
}

// StartAsync starts the TUI program in a goroutine.
// Waits for TUI to be ready before returning. Call Stop to stop the program.
func (c *Console) StartAsync(ctx context.Context) {
	go func() {
		//nolint:errcheck // TUI errors are non-fatal in async mode
		_ = c.Start(ctx)
	}()

	// Wait for TUI to be ready
	<-c.ready
}

// Wait waits for the TUI program to fully complete, including printSummary().
// Does not force the program to quit. Use Stop() to force quit.
func (c *Console) Wait() {
	c.mu.Lock()
	started := c.started
	c.mu.Unlock()

	// Wait for Start() to fully complete (including printSummary)
	// The done channel is closed by Start() after all cleanup
	if started {
		<-c.done
	}

	c.mu.Lock()
	c.stopped = true
	c.mu.Unlock()
}

// Stop stops the TUI program.
func (c *Console) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	started := c.started
	c.mu.Unlock()

	// Close all multi-writers first
	c.mu.Lock()
	for _, w := range c.writers {
		w.Close()
	}
	c.writers = nil
	c.mu.Unlock()

	// Close channels (this will trigger TUI to exit)
	close(c.lineChan)
	close(c.statusChan)

	// Quit the program and wait for it to fully exit
	if c.program != nil {
		c.program.Quit()
		c.program.Wait()
	}

	// Wait for Start() to fully complete (including printSummary)
	// This ensures the final output is printed before we reset the terminal
	if started {
		<-c.done
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

// SetPhase switches to a new phase (Init, Run, End).
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

// SetPhaseSummary sets the summary text for a collapsed phase.
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

// CompletePhase marks a phase as complete with a summary.
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

// WriteToPhase writes a line to a specific phase's buffer.
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

// WriteResult writes a line to the results buffer (appears below Run pane).
func (c *Console) WriteResult(text string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	// Send line to results buffer via Bubbletea
	program.Send(console.ResultLineMsg{
		Line: console.Line{
			Text:   text,
			Source: "results",
			Level:  console.LevelInfo,
		},
	})
}

// StartModule notifies the TUI that a module has started with its weight.
// This creates the tab in pending state (scheduled but waiting for slot).
func (c *Console) StartModule(moniker string, weight int) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.ModuleStartMsg{
		Moniker: moniker,
		Weight:  weight,
	})
}

// MarkModuleRunning notifies the TUI that a module has acquired its execution slot.
// This transitions the tab from pending to running state.
func (c *Console) MarkModuleRunning(moniker string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.ModuleRunningMsg{
		Moniker: moniker,
	})
}

// MarkModuleComplete notifies the TUI that a module has finished with the given exit code.
// Exit code 0 = success, non-zero = failure.
func (c *Console) MarkModuleComplete(moniker string, exitCode int) {
	c.MarkModuleCompleteWithCacheInfo(moniker, exitCode, time.Time{}, "")
}

// MarkModuleCompleteWithCacheInfo marks a module as complete with optional cache info.
// For cached modules, cacheTime is when the artifact was built, logPath is the build log location.
func (c *Console) MarkModuleCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.ModuleCompleteMsg{
		Moniker:   moniker,
		ExitCode:  exitCode,
		CacheTime: cacheTime,
		LogPath:   logPath,
	})
}

// SendSummary sends summary data and activates the Summary pane.
func (c *Console) SendSummary(data *SummaryData) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.SummaryDataMsg{
		Data: (*console.SummaryData)(data),
	})
}

// SetInitSummary sends init summary data for structured display in the Init pane.
func (c *Console) SetInitSummary(summary *InitSummary) {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.InitSummaryMsg{
		Summary: (*console.InitSummary)(summary),
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

// SummaryData is an alias for console.SummaryData for public use.
type SummaryData = console.SummaryData

// LockStatus is an alias for console.LockStatus for public use.
type LockStatus = console.LockStatus

// InitSummary is an alias for console.InitSummary for public use.
type InitSummary = console.InitSummary

// InitSummaryFlags is an alias for console.InitSummaryFlags for public use.
type InitSummaryFlags = console.InitSummaryFlags

// ExecutionLayer is an alias for console.ExecutionLayer for public use.
type ExecutionLayer = console.ExecutionLayer

// ExecutionModule is an alias for console.ExecutionModule for public use.
type ExecutionModule = console.ExecutionModule

// Level constants for public use.
const (
	LevelInfo  = console.LevelInfo
	LevelWarn  = console.LevelWarn
	LevelError = console.LevelError
)

// Phase constants for public use.
const (
	PhaseInit    = console.PhaseInit
	PhaseRun     = console.PhaseRun
	PhaseSummary = console.PhaseSummary
)

// PhaseStatus constants for public use.
const (
	PhasePending  = console.PhasePending
	PhaseActive   = console.PhaseActive
	PhaseComplete = console.PhaseComplete
	PhaseFailed   = console.PhaseFailed
)

// printSummary prints a plain-text summary after the TUI exits.
func (c *Console) printSummary(m *console.Model) {
	// Use the console package's ViewFinal method to generate plain-text output
	summary := m.ViewFinal()
	fmt.Print(summary)
}

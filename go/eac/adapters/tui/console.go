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
	tuicontract "github.com/ready-to-release/eac/contracts/eac-adapters-tui"
	"github.com/ready-to-release/eac/go/eac/adapters/tui/console"
	"github.com/ready-to-release/eac/go/eac/adapters/tui/stream"
	"golang.org/x/term"
)

// isTerminal returns true if the given file descriptor is a terminal.
func isTerminal(fd int) bool {
	return term.IsTerminal(fd)
}

// ParallelConsole manages the TUI console window for build/test output.
type ParallelConsole struct {
	config Config

	// Channel for stream.MultiWriter to send lines (uses contract types)
	contractLineChan chan tuicontract.Line

	// Internal channels that the console.Model reads from
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

	// Async message queue to prevent blocking workers.
	// program.Send() is blocking - if the TUI event loop is slow/busy,
	// it blocks indefinitely, causing workers to hang. This decouples
	// worker completion from TUI rendering speed.
	msgChan chan tea.Msg
	msgWg   sync.WaitGroup // Track pending messages for clean shutdown
}

// NewParallelConsole creates a new console with the given configuration.
func NewParallelConsole(config Config) *ParallelConsole {
	if config.Height <= 0 {
		config.Height = 5
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	return &ParallelConsole{
		config:           config,
		contractLineChan: make(chan tuicontract.Line, 100),
		lineChan:         make(chan console.Line, 100),
		statusChan:       make(chan console.Status, 10),
		ready:            make(chan struct{}),
		done:             make(chan struct{}),
	}
}

// Start starts the TUI program. This blocks until the program exits.
// Use StartAsync for non-blocking start.
func (c *ParallelConsole) Start(ctx context.Context) error {
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

	// Start goroutine to convert contract lines to console lines
	go c.convertLines()

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

	// Build program options based on environment
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),         // Take over screen, restore on exit
		tea.WithoutBracketedPaste(),
		tea.WithoutSignalHandler(), // Let our custom signal handler catch Ctrl-C
	}

	// Check if stdin is a terminal - if not, disable input to prevent blocking
	// When stdin is not a terminal (e.g., /dev/null, pipe), bubbletea's default
	// behavior of opening a new TTY can block indefinitely
	stdinFd := int(os.Stdin.Fd())
	stdinIsTerminal := isTerminal(stdinFd)
	if stdinIsTerminal {
		// Normal terminal mode - enable mouse for scrolling and hover
		opts = append(opts, tea.WithMouseAllMotion())
	} else {
		// Non-interactive mode - disable input to prevent blocking
		// This happens when stdin is redirected (e.g., running in background)
		opts = append(opts, tea.WithInput(nil))
	}

	c.program = tea.NewProgram(model, opts...)

	// Start async message pump goroutine.
	// This decouples worker completion from TUI rendering - workers can
	// complete immediately without blocking on slow TUI updates.
	c.msgChan = make(chan tea.Msg, 100) // Buffered to absorb bursts
	c.msgWg.Add(1)
	go func() {
		defer c.msgWg.Done()
		for msg := range c.msgChan {
			c.program.Send(msg)
		}
	}()

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

			// Close contract line channel (this stops the converter goroutine)
			close(c.contractLineChan)

			// Close internal channels to trigger TUI exit
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

// convertLines reads from contractLineChan and converts to console.Line
func (c *ParallelConsole) convertLines() {
	for line := range c.contractLineChan {
		c.mu.Lock()
		stopped := c.stopped
		c.mu.Unlock()

		if stopped {
			return
		}

		// Convert from contract type to console type
		consoleLine := console.Line{
			Text:      line.Text,
			Source:    line.Source,
			Level:     console.Level(line.Level),
			Timestamp: line.Timestamp,
		}

		select {
		case c.lineChan <- consoleLine:
		default:
			// Drop if channel full
		}
	}
}

// StartAsync starts the TUI program in a goroutine.
// Waits for TUI to be ready before returning. Call Stop to stop the program.
func (c *ParallelConsole) StartAsync(ctx context.Context) {
	go func() {
		//nolint:errcheck // TUI errors are non-fatal in async mode
		_ = c.Start(ctx)
	}()

	// Wait for TUI to be ready
	<-c.ready
}

// Wait waits for the TUI program to fully complete, including printSummary().
// Does not force the program to quit. Use Stop() to force quit.
func (c *ParallelConsole) Wait() {
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
func (c *ParallelConsole) Stop() {
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
	close(c.contractLineChan)
	close(c.lineChan)
	close(c.statusChan)

	// Close async message channel and wait for pending messages.
	// This ensures all queued TUI updates are delivered before exit.
	if c.msgChan != nil {
		close(c.msgChan)
		c.msgWg.Wait()
	}

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
func (c *ParallelConsole) NewWriter(source string, logWriter io.Writer) io.Writer {
	c.mu.Lock()
	defer c.mu.Unlock()

	mw := stream.NewMultiWriter(c.contractLineChan, source, logWriter)
	c.writers = append(c.writers, mw)
	return mw
}

// UpdateStatus sends a status update to the console.
func (c *ParallelConsole) UpdateStatus(status tuicontract.Status) {
	c.mu.Lock()
	stopped := c.stopped
	c.mu.Unlock()

	if stopped {
		return
	}

	// Convert from contract type to internal console type
	locks := make([]console.LockStatus, len(status.Locks))
	for i, lock := range status.Locks {
		locks[i] = console.LockStatus{
			Name:     lock.Name,
			Type:     lock.Type,
			Capacity: lock.Capacity,
			Used:     lock.Used,
			Waiting:  lock.Waiting,
		}
	}

	consoleStatus := console.Status{
		Phase:             status.Phase,
		Running:           status.Running,
		Completed:         status.Completed,
		Total:             status.Total,
		Layer:             status.Layer,
		TotalLayers:       status.TotalLayers,
		Locks:             locks,
		ActiveContainers:  status.ActiveContainers,
		UsedContainers:    status.UsedContainers,
		ActiveSystemTools: status.ActiveSystemTools,
		UsedSystemTools:   status.UsedSystemTools,
	}

	select {
	case c.statusChan <- consoleStatus:
	default:
		// Drop if channel full
	}
}

// SendLine directly sends a line to the console.
func (c *ParallelConsole) SendLine(line tuicontract.Line) {
	c.mu.Lock()
	stopped := c.stopped
	c.mu.Unlock()

	if stopped {
		return
	}

	// Convert from contract type to internal console type
	consoleLine := console.Line{
		Text:      line.Text,
		Source:    line.Source,
		Level:     console.Level(line.Level),
		Timestamp: line.Timestamp,
	}

	select {
	case c.lineChan <- consoleLine:
	default:
		// Drop if channel full
	}
}

// SendInfo sends an info line to the console.
func (c *ParallelConsole) SendInfo(source, text string) {
	c.SendLine(tuicontract.Line{
		Text:   text,
		Source: source,
		Level:  tuicontract.LevelInfo,
	})
}

// SendWarn sends a warning line to the console.
func (c *ParallelConsole) SendWarn(source, text string) {
	c.SendLine(tuicontract.Line{
		Text:   text,
		Source: source,
		Level:  tuicontract.LevelWarn,
	})
}

// SendError sends an error line to the console.
func (c *ParallelConsole) SendError(source, text string) {
	c.SendLine(tuicontract.Line{
		Text:   text,
		Source: source,
		Level:  tuicontract.LevelError,
	})
}

// sendAsync queues a message for async delivery to the TUI.
// Non-blocking - returns immediately even if TUI is slow.
// This prevents workers from blocking on slow TUI rendering.
func (c *ParallelConsole) sendAsync(msg tea.Msg) {
	c.mu.Lock()
	if c.stopped || c.program == nil || c.msgChan == nil {
		c.mu.Unlock()
		return
	}
	msgChan := c.msgChan
	c.mu.Unlock()

	select {
	case msgChan <- msg:
		// Queued successfully
	default:
		// Buffer full - drop message (TUI updates are lossy)
	}
}

// SetPhase switches to a new phase (Init, Run, End).
func (c *ParallelConsole) SetPhase(phase Phase) {
	c.sendAsync(console.PhaseUpdateMsg{
		Phase:  console.Phase(phase),
		Status: console.PhaseActive,
	})
}

// SetPhaseSummary sets the summary text for a collapsed phase.
func (c *ParallelConsole) SetPhaseSummary(phase Phase, summary string) {
	c.sendAsync(console.PhaseUpdateMsg{
		Phase:   console.Phase(phase),
		Summary: summary,
	})
}

// CompletePhase marks a phase as complete with a summary.
func (c *ParallelConsole) CompletePhase(phase Phase, success bool, summary string) {
	status := console.PhaseComplete
	if !success {
		status = console.PhaseFailed
	}

	c.sendAsync(console.PhaseUpdateMsg{
		Phase:   console.Phase(phase),
		Status:  status,
		Summary: summary,
	})
}

// WriteToPhase writes a line to a specific phase's buffer.
func (c *ParallelConsole) WriteToPhase(phase Phase, text string) {
	c.sendAsync(console.PhaseLineMsg{
		Phase: console.Phase(phase),
		Line: console.Line{
			Text:   text,
			Source: phase.String(),
			Level:  console.LevelInfo,
		},
	})
}

// WriteResult writes a line to the results buffer (appears below Run pane).
func (c *ParallelConsole) WriteResult(text string) {
	c.sendAsync(console.ResultLineMsg{
		Line: console.Line{
			Text:   text,
			Source: "results",
			Level:  console.LevelInfo,
		},
	})
}

// StartModule notifies the TUI that a module has started with its weight.
// This creates the tab in pending state (scheduled but waiting for slot).
func (c *ParallelConsole) StartModule(moniker string, weight int) {
	c.sendAsync(console.ModuleStartMsg{
		Moniker: moniker,
		Weight:  weight,
	})
}

// MarkModuleRunning notifies the TUI that a module has acquired its execution slot.
// This transitions the tab from pending to running state.
func (c *ParallelConsole) MarkModuleRunning(moniker string) {
	c.sendAsync(console.ModuleRunningMsg{
		Moniker: moniker,
	})
}

// MarkModuleComplete notifies the TUI that a module has finished with the given exit code.
// Exit code 0 = success, non-zero = failure.
func (c *ParallelConsole) MarkModuleComplete(moniker string, exitCode int) {
	c.MarkModuleCompleteWithCacheInfo(moniker, exitCode, time.Time{}, "")
}

// MarkModuleCompleteWithCacheInfo marks a module as complete with optional cache info.
// For cached modules, cacheTime is when the artifact was built, logPath is the build log location.
func (c *ParallelConsole) MarkModuleCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	c.sendAsync(console.ModuleCompleteMsg{
		Moniker:   moniker,
		ExitCode:  exitCode,
		CacheTime: cacheTime,
		LogPath:   logPath,
	})
}

// SendSummary sends summary data and activates the Summary pane.
// NOTE: This intentionally uses blocking program.Send() (not sendAsync).
// The final summary must be delivered before TUI exits - it's only called
// once at the end and we need to guarantee it's displayed.
func (c *ParallelConsole) SendSummary(data *SummaryData) {
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
func (c *ParallelConsole) SetInitSummary(summary *InitSummary) {
	// Convert to console.InitSummary
	layers := make([]console.ExecutionLayer, len(summary.ExecutionTree))
	for i, layer := range summary.ExecutionTree {
		modules := make([]console.ExecutionModule, len(layer.Modules))
		for j, mod := range layer.Modules {
			modules[j] = console.ExecutionModule{
				Name:       mod.Name,
				Components: mod.Components,
			}
		}
		layers[i] = console.ExecutionLayer{Modules: modules}
	}

	c.sendAsync(console.InitSummaryMsg{
		Summary: &console.InitSummary{
			Command:               summary.Command,
			ExecutionContext:      summary.ExecutionContext,
			RequestedModules:      summary.RequestedModules,
			CalculatedModules:     summary.CalculatedModules,
			AddedDepm:             summary.AddedDepm,
			UoWCount:              summary.UoWCount,
			ExecutionTree:         layers,
			LayerCount:            summary.LayerCount,
			LayerSizes:            summary.LayerSizes,
			ComponentsPerModLayer: summary.ComponentsPerModLayer,
			FlatExecution:         summary.FlatExecution,
			ParallelismMode:       summary.ParallelismMode,
			EffectiveWorkers:      summary.EffectiveWorkers,
			TurboBoost:            summary.TurboBoost,
			WeightedCapacity:      summary.WeightedCapacity,
			Flags: console.InitSummaryFlags{
				TidyFirst:    summary.Flags.TidyFirst,
				ForceRebuild: summary.Flags.ForceRebuild,
				DryRun:       summary.Flags.DryRun,
				UseTUI:       summary.Flags.UseTUI,
				SkipDeps:     summary.Flags.SkipDeps,
				SkipDepm:     summary.Flags.SkipDepm,
			},
			DepsVerified:        summary.DepsVerified,
			DepsSkipped:         summary.DepsSkipped,
			DepsAvailable:       summary.DepsAvailable,
			DepsMissing:         summary.DepsMissing,
			DepmVerified:        summary.DepmVerified,
			DepmSkipped:         summary.DepmSkipped,
			DepmResolved:        summary.DepmResolved,
			DepmExisting:        summary.DepmExisting,
			DepmTotal:           summary.DepmTotal,
			DepmMissing:         summary.DepmMissing,
			IncrementalEnabled:  summary.IncrementalEnabled,
			IncrementalChanged:  summary.IncrementalChanged,
			IncrementalUpToDate: summary.IncrementalUpToDate,
			IncrementalFresh:    summary.IncrementalFresh,
			TestSuiteName:       summary.TestSuiteName,
			TestSelected:        summary.TestSelected,
			TestDiscovered:      summary.TestDiscovered,
			TestOSFiltered:      summary.TestOSFiltered,
			OutputDir:           summary.OutputDir,
		},
	})
}

// printSummary prints a plain-text summary after the TUI exits.
func (c *ParallelConsole) printSummary(m *console.Model) {
	// Use the console package's ViewFinal method to generate plain-text output
	summary := m.ViewFinal()
	fmt.Print(summary)
}

// Package tui provides a text user interface for displaying build and test output.
package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuicontract "github.com/ready-to-release/eac/contracts/tui/0.1.0"
	"github.com/ready-to-release/eac/go/adapters/tui/console"
	"github.com/ready-to-release/eac/go/adapters/tui/demo"
	"github.com/ready-to-release/eac/go/adapters/tui/demo/cells"
	"github.com/ready-to-release/eac/go/adapters/tui/stream"
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

	mu       sync.Mutex
	started  bool
	stopped  bool
	demoMode bool          // True when using experimental demo layout
	ready    chan struct{} // Signals when TUI is ready
	done     chan struct{} // Signals when Start() has fully completed (including printSummary)

	// Atomic flags for hot-path reads (avoids mutex in convertLines per-line loop)
	atomicStopped  int32 // 1 = stopped
	atomicDemoMode int32 // 1 = demo mode

	// Signals to Model that it should stop listening.
	// Close this to unblock listenForLines/listenForStatus.
	modelDone chan struct{}

	// Track multi-writers for cleanup
	writers []*stream.MultiWriter

	// Store final model state for post-exit summary
	finalModel *console.Model

	// Async message pump - delivers messages to the TUI program with priority ordering.
	pump *messagePump

	// Real terminal output - captured at creation time before any output buffering.
	// This ensures TUI writes directly to terminal even when stdout is redirected.
	realStdout *os.File
	realStderr *os.File
}

// NewParallelConsole creates a new console with the given configuration.
// IMPORTANT: This captures os.Stdout/os.Stderr at creation time. Create the console
// BEFORE any output buffering is started to ensure TUI writes directly to terminal.
func NewParallelConsole(config Config) *ParallelConsole {
	if config.Height <= 0 {
		config.Height = 5
	}
	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	return &ParallelConsole{
		config:           config,
		contractLineChan: make(chan tuicontract.Line, 500),
		lineChan:         make(chan console.Line, 500),
		statusChan:       make(chan console.Status, 10),
		ready:            make(chan struct{}),
		done:             make(chan struct{}),
		modelDone:        make(chan struct{}),
		// Capture real stdout/stderr before any buffering redirects them
		realStdout:       os.Stdout,
		realStderr:       os.Stderr,
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

	// Prevent lipgloss from querying terminal background color (causes OSC escape leaks)
	lipgloss.SetHasDarkBackground(true)

	// Create the appropriate model and program
	model := c.createModel()
	opts := c.buildProgramOptions()
	c.program = tea.NewProgram(model, opts...)

	// Start async message pump
	c.startMessagePump()

	// Signal that TUI is ready
	close(c.ready)

	// Run with signal handling and cleanup
	return c.runWithCleanup(ctx)
}

// createModel builds the bubbletea model based on configuration (standard vs demo).
func (c *ParallelConsole) createModel() tea.Model {
	if c.config.TUI3Demo {
		c.mu.Lock()
		c.demoMode = true
		atomic.StoreInt32(&c.atomicDemoMode, 1)
		c.mu.Unlock()

		width, height, _ := term.GetSize(int(os.Stdout.Fd()))
		if width <= 0 {
			width = 120
		}
		if height <= 0 {
			height = 40
		}
		return demo.NewModel(width, height, c.config.RunPhaseName, c.config.ASCIIMode)
	}

	return console.NewModel(console.ModelOptions{
		Height:       c.config.Height,
		RunPhaseName: c.config.RunPhaseName,
		LineChan:     c.lineChan,
		StatusChan:   c.statusChan,
		DoneChan:     c.modelDone,
		ASCIIMode:    c.config.ASCIIMode,
		SkipTUIDelay: c.config.SkipTUIDelay,
		TUIConfig:    c.config.TUIConfig,
	})
}

// buildProgramOptions constructs bubbletea program options based on the environment.
func (c *ParallelConsole) buildProgramOptions() []tea.ProgramOption {
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithoutBracketedPaste(),
		tea.WithoutSignalHandler(),
		tea.WithOutput(c.realStdout),
	}

	stdinFd := int(os.Stdin.Fd())
	if isTerminal(stdinFd) {
		opts = append(opts, tea.WithMouseAllMotion())
	} else {
		opts = append(opts, tea.WithInput(nil))
	}

	return opts
}

// startMessagePump initializes and starts the async message pump.
// TUI is a pure observer - workers never block on TUI rendering.
func (c *ParallelConsole) startMessagePump() {
	c.pump = newMessagePump(c.program)
	c.pump.Start()
}

// runWithCleanup runs the TUI program with signal handling and deferred cleanup.
func (c *ParallelConsole) runWithCleanup(ctx context.Context) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	defer c.cleanup()

	sigDone := make(chan struct{})
	defer close(sigDone)

	go func() {
		select {
		case <-sigChan:
			c.program.Quit()
		case <-ctx.Done():
			c.program.Quit()
		case <-sigDone:
			return
		}
	}()

	finalModel, err := c.program.Run()
	if err != nil {
		return err
	}

	c.handleFinalModel(finalModel)
	return nil
}

// cleanup closes channels and resets terminal state after the TUI exits.
func (c *ParallelConsole) cleanup() {
	c.mu.Lock()
	if !c.stopped {
		c.stopped = true
		atomic.StoreInt32(&c.atomicStopped, 1)

		for _, w := range c.writers {
			w.Close()
		}
		c.writers = nil

		close(c.contractLineChan)
		close(c.lineChan)
		close(c.statusChan)
	}
	c.mu.Unlock()

	if c.pump != nil {
		c.pump.Close()
	}

	fmt.Fprint(c.realStdout, "\033[0m\033[?25h")
}

// handleFinalModel stores and renders the final model state after program exit.
func (c *ParallelConsole) handleFinalModel(finalModel tea.Model) {
	switch m := finalModel.(type) {
	case console.Model:
		c.mu.Lock()
		c.finalModel = &m
		c.mu.Unlock()
		c.printSummary(&m)
	case demo.Model:
		fmt.Fprintln(c.realStdout, "Execution complete.")
	}
}

// convertLines reads from contractLineChan and converts to console.Line or tui3 messages
func (c *ParallelConsole) convertLines() {
	for line := range c.contractLineChan {
		// Hot path: use atomic reads to avoid mutex contention per line
		if atomic.LoadInt32(&c.atomicStopped) != 0 {
			return
		}

		if atomic.LoadInt32(&c.atomicDemoMode) != 0 {
			// Convert to tui3 output line message
			level := cells.LineLevelInfo
			if line.Level == tuicontract.LevelWarn {
				level = cells.LineLevelWarn
			} else if line.Level == tuicontract.LevelError {
				level = cells.LineLevelError
			}

			c.sendAsync(demo.OutputLineMsg{
				Line: cells.OutputLine{
					Text:   line.Text,
					Source: line.Source,
					Level:  level,
				},
			})
		} else {
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
	atomic.StoreInt32(&c.atomicStopped, 1)
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
	atomic.StoreInt32(&c.atomicStopped, 1)
	started := c.started
	c.mu.Unlock()

	// Signal model to stop listening FIRST - this unblocks the listeners
	c.closeModelDoneOnce()

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

	// Close async message pump and wait for pending messages.
	// This ensures all queued TUI updates are delivered before exit.
	if c.pump != nil {
		c.pump.Close()
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
	fmt.Fprint(c.realStdout, "\033[0m\033[?25h")
}

// closeModelDoneOnce safely closes modelDone channel once.
// Safe to call multiple times from different goroutines.
// This signals listenForLines/listenForStatus to stop blocking.
func (c *ParallelConsole) closeModelDoneOnce() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.modelDone:
		// Already closed
	default:
		close(c.modelDone)
	}
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
	demoMode := c.demoMode
	c.mu.Unlock()

	if stopped {
		return
	}

	if demoMode {
		// Send tool status updates for containers
		for _, name := range status.ActiveContainerTools {
			c.sendAsync(demo.ToolActiveMsg{
				Name:        name,
				IsContainer: true,
			})
		}

		// Send tool status updates for system tools
		for _, name := range status.ActiveSystemTools {
			c.sendAsync(demo.ToolActiveMsg{
				Name:        name,
				IsContainer: false,
			})
		}
		return
	}

	// Standard console mode
	consoleStatus := convertContractStatus(status)

	select {
	case c.statusChan <- consoleStatus:
	default:
		// Drop if channel full
	}
}

// StatusRefreshInterval returns how often the orchestrator should send status updates.
// This keeps the TUI responsive even during idle periods.
func (c *ParallelConsole) StatusRefreshInterval() time.Duration {
	if c.config.TUIConfig != nil && c.config.TUIConfig.StatusRefreshInterval > 0 {
		return c.config.TUIConfig.StatusRefreshInterval
	}
	return tuicontract.DefaultTUIConfig().StatusRefreshInterval
}

// SendLine directly sends a line to the console.
func (c *ParallelConsole) SendLine(line tuicontract.Line) {
	c.mu.Lock()
	stopped := c.stopped
	demoMode := c.demoMode
	c.mu.Unlock()

	if stopped {
		return
	}

	if demoMode {
		// Convert to tui3 output line message
		level := cells.LineLevelInfo
		if line.Level == tuicontract.LevelWarn {
			level = cells.LineLevelWarn
		} else if line.Level == tuicontract.LevelError {
			level = cells.LineLevelError
		}

		c.sendAsync(demo.OutputLineMsg{
			Line: cells.OutputLine{
				Text:   line.Text,
				Source: line.Source,
				Level:  level,
			},
		})
		return
	}

	// Standard console mode
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
// NOTE: Output lines may be dropped if buffer is full; use sendCritical for state changes.
func (c *ParallelConsole) sendAsync(msg tea.Msg) {
	c.mu.Lock()
	if c.stopped || c.program == nil || c.pump == nil {
		c.mu.Unlock()
		return
	}
	pump := c.pump
	c.mu.Unlock()

	pump.SendAsync(msg)
}

// sendCritical sends a message that must not be dropped (state changes).
// Uses a dedicated critical message channel with larger buffer.
// Never blocks - relies on large buffer to avoid drops.
func (c *ParallelConsole) sendCritical(msg tea.Msg) {
	c.mu.Lock()
	if c.stopped || c.program == nil || c.pump == nil {
		c.mu.Unlock()
		return
	}
	pump := c.pump
	c.mu.Unlock()

	pump.SendCritical(msg)
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

// StartUoW notifies the TUI that a module has started with its weight.
// This creates the tab in pending state (scheduled but waiting for slot).
// The displayName is used for tab labels; moniker is the globally unique ID for matching.
// Uses sendCritical because state changes must not be dropped.
func (c *ParallelConsole) StartUoW(moniker, displayName string, weight int) {
	c.mu.Lock()
	demoMode := c.demoMode
	c.mu.Unlock()

	if demoMode {
		c.sendCritical(demo.UnitStartMsg{
			Moniker:     moniker,
			DisplayName: displayName,
			Weight:      weight,
		})
		return
	}

	c.sendCritical(console.UoWStartMsg{
		Moniker:     moniker,
		DisplayName: displayName,
		Weight:      weight,
	})
}

// MarkUoWRunning notifies the TUI that a module has acquired its execution slot.
// This transitions the tab from pending to running state.
// Uses sendCritical because state changes must not be dropped.
func (c *ParallelConsole) MarkUoWRunning(moniker string) {
	c.mu.Lock()
	demoMode := c.demoMode
	c.mu.Unlock()

	if demoMode {
		c.sendCritical(demo.UnitRunningMsg{
			Moniker: moniker,
		})
		return
	}

	c.sendCritical(console.UoWRunningMsg{
		Moniker: moniker,
	})
}

// MarkUoWComplete notifies the TUI that a module has finished with the given exit code.
// Exit code 0 = success, non-zero = failure.
func (c *ParallelConsole) MarkUoWComplete(moniker string, exitCode int) {
	c.MarkUoWCompleteWithCacheInfo(moniker, exitCode, time.Time{}, "")
}

// MarkUoWCompleteWithCacheInfo marks a module as complete with optional cache info.
// For cached modules, cacheTime is when the artifact was built, logPath is the build log location.
// Uses sendCritical (priority channel) so completion events are processed immediately.
// This ensures the tab turns green/red as soon as the worker finishes.
func (c *ParallelConsole) MarkUoWCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
	c.mu.Lock()
	demoMode := c.demoMode
	c.mu.Unlock()

	if demoMode {
		c.sendCritical(demo.UnitCompleteMsg{
			Moniker:  moniker,
			ExitCode: exitCode,
		})
		return
	}

	c.sendCritical(console.UoWCompleteMsg{
		Moniker:   moniker,
		ExitCode:  exitCode,
		CacheTime: cacheTime,
		LogPath:   logPath,
	})
}

// SendSummary sends summary data and activates the Summary pane.
// Routes through criticalChan so the summary queues behind any pending
// completion events — this prevents tabs from appearing "running" when
// the summary arrives (completion events turn them green/red first).
// The critical channel has 10,000 capacity; blocking send guarantees delivery.
func (c *ParallelConsole) SendSummary(data *SummaryData) {
	c.mu.Lock()
	stopped := c.stopped
	demoMode := c.demoMode
	pump := c.pump
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	if demoMode {
		// Send summary through critical channel for ordering
		c.sendCritical(demo.SummaryMsg{
			Data: &cells.SummaryData{
				Duration: data.TotalTime,
			},
		})

		// Signal listeners to stop AFTER sending summary so remaining lines
		// in lineChan can be drained by listenForLines before it exits.
		c.closeModelDoneOnce()

		// Send quit after a short delay to allow final UI update
		go func() {
			time.Sleep(500 * time.Millisecond)
			program.Send(demo.QuitMsg{})
		}()
		return
	}

	// Send through critical channel for ordering after preceding completion events.
	if pump != nil {
		pump.SendCritical(console.SummaryDataMsg{
			Data: (*console.SummaryData)(data),
		})
	} else {
		// Fallback: direct send if pump not initialized
		program.Send(console.SummaryDataMsg{
			Data: (*console.SummaryData)(data),
		})
	}

	// Signal listeners to stop AFTER sending summary so remaining lines
	// in lineChan can be drained by listenForLines before it exits.
	// TUI exit is driven by SummaryDataMsg → tea.Quit, not by doneChan.
	c.closeModelDoneOnce()
}

// SendConfigReady delivers early configuration metadata for progressive display.
// Called after config loading but before module resolution.
func (c *ParallelConsole) SendConfigReady(commandName, executionContext, parallelismMode string,
	effectiveWorkers, weightedCapacity int, outputDir string) {
	c.sendAsync(console.ConfigReadyMsg{
		CommandName:      commandName,
		ExecutionContext: executionContext,
		ParallelismMode:  parallelismMode,
		EffectiveWorkers: effectiveWorkers,
		WeightedCapacity: weightedCapacity,
		OutputDir:        outputDir,
	})
}

// SetInitSummary sends init summary data for structured display in the Init pane.
func (c *ParallelConsole) SetInitSummary(summary *InitSummary) {
	if summary == nil {
		return
	}

	// Check if we're in tui3 mode
	c.mu.Lock()
	demoMode := c.demoMode
	c.mu.Unlock()

	if demoMode {
		// Pre-populate all units from ExecutionTree for tui3
		// Use sendCritical because unit registration must not be dropped
		for _, module := range summary.ExecutionTree {
			for _, uow := range module.UoWs {
				// Add unit in pending state with short display name for tabs
				c.sendCritical(demo.UnitStartMsg{
					Moniker:     uow.ID,
					DisplayName: uow.DisplayName,
					Weight:      uow.Weight,
				})
			}
		}
		return
	}

	// Standard console mode - convert to console.InitSummary
	modules := make([]console.ExecutionModule, len(summary.ExecutionTree))
	for i, mod := range summary.ExecutionTree {
		// Convert UoWs preserving ID (for matching), DisplayName (for display), and Weight
		uows := make([]console.UoWEntry, len(mod.UoWs))
		for k, uow := range mod.UoWs {
			uows[k] = console.UoWEntry{
				ID:            uow.ID,
				DisplayName:   uow.DisplayName,
				Weight:        uow.Weight,
				Module:        uow.Module,
				Component:     uow.Component,
				Tool:          uow.Tool,
				ComponentType: uow.ComponentType,
				Container:     uow.Container,
			}
		}
		modules[i] = console.ExecutionModule{
			Name: mod.Name,
			UoWs: uows,
		}
	}

	// Convert PlannedTools
	plannedTools := make([]console.PlannedTool, len(summary.PlannedTools))
	for i, tool := range summary.PlannedTools {
		plannedTools[i] = console.PlannedTool{
			Name:        tool.Name,
			IsContainer: tool.IsContainer,
		}
	}

	c.sendAsync(console.InitSummaryMsg{
		Summary: &console.InitSummary{
			Command:             summary.Command,
			ExecutionContext:    summary.ExecutionContext,
			RequestedModules:    summary.RequestedModules,
			CalculatedModules:   summary.CalculatedModules,
			AddedDepm:           summary.AddedDepm,
			UoWCount:            summary.UoWCount,
			ExecutionTree:       modules,
			ParallelismMode:     summary.ParallelismMode,
			EffectiveWorkers:    summary.EffectiveWorkers,
			TurboBoost:          summary.TurboBoost,
			WeightedCapacity:    summary.WeightedCapacity,
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
			PlannedTools:        plannedTools,
		},
	})
}

// SendPlannedWork delivers predicted work items for grey skeleton tabs.
func (c *ParallelConsole) SendPlannedWork(items []PlannedWorkItem) {
	consoleItems := make([]console.PlannedWorkItem, len(items))
	for i, item := range items {
		consoleItems[i] = console.PlannedWorkItem{
			ID:            item.ID,
			DisplayName:   item.DisplayName,
			Weight:        item.Weight,
			Module:        item.Module,
			Component:     item.Component,
			ComponentType: item.ComponentType,
		}
	}
	c.sendCritical(console.PlannedWorkMsg{Items: consoleItems})
}

// EnrichUoW delivers incremental enrichment data for an existing planned tab.
func (c *ParallelConsole) EnrichUoW(enrichment UoWEnrichment) {
	c.sendCritical(console.UoWEnrichMsg{
		ID:          enrichment.ID,
		Tool:        enrichment.Tool,
		Container:   enrichment.Container,
		Weight:      enrichment.Weight,
		CacheStatus: console.CacheHit(enrichment.CacheStatus),
		CacheTime:   enrichment.CacheTime,
		DependsOn:   enrichment.DependsOn,
	})
}

// SignalAllWorkDone signals that all work including AfterExecute hooks is complete.
// Uses blocking program.Send (same pattern as SendSummary) to guarantee delivery.
func (c *ParallelConsole) SignalAllWorkDone() {
	c.mu.Lock()
	stopped := c.stopped
	program := c.program
	c.mu.Unlock()

	if stopped || program == nil {
		return
	}

	program.Send(console.AllWorkDoneMsg{})
}

// convertContractStatus converts a contract Status to internal console.Status.
func convertContractStatus(status tuicontract.Status) console.Status {
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

	return console.Status{
		Phase:                 status.Phase,
		Running:               status.Running,
		Completed:             status.Completed,
		Total:                 status.Total,
		Roof:                  status.Roof,
		PressureTarget:        status.PressureTarget,
		Locks:                 locks,
		ActiveContainerTools:  status.ActiveContainerTools,
		UsedContainerTools:    status.UsedContainerTools,
		ActiveSystemTools:     status.ActiveSystemTools,
		UsedSystemTools:       status.UsedSystemTools,
		DockerMemPercent:      status.DockerMemPercent,
		DockerAvailable:       status.DockerAvailable,
		RunningContainerCount: status.RunningContainerCount,
		TotalContainerCount:   status.TotalContainerCount,
		RunningSystemCount:    status.RunningSystemCount,
		TotalSystemCount:      status.TotalSystemCount,
	}
}

// printSummary prints a plain-text summary after the TUI exits.
func (c *ParallelConsole) printSummary(m *console.Model) {
	// Use the console package's ViewFinal method to generate plain-text output
	summary := m.ViewFinal()
	fmt.Fprint(c.realStdout, summary)
}

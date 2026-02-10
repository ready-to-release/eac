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

	// Async message queues - TUI is pure observer, never blocks workers.
	// Two channels: critical (state changes) and regular (output lines).
	// Critical channel has larger buffer and is drained with priority.
	msgChan      chan tea.Msg     // Regular messages (output lines) - may drop
	criticalChan chan tea.Msg     // Critical messages (state changes) - large buffer
	msgWg        sync.WaitGroup  // Track pending messages for clean shutdown
	msgCloseOnce sync.Once       // Ensures criticalChan/msgChan closed exactly once

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

	// Create the appropriate model based on demo mode
	var model tea.Model
	if c.config.TUI3Demo {
		// Use experimental tui3 layout
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
		model = demo.NewModel(width, height, c.config.RunPhaseName, c.config.ASCIIMode)
	} else {
		// Use standard console model
		model = console.NewModel(
			c.config.Height,
			c.config.RunPhaseName,
			c.lineChan,
			c.statusChan,
			c.modelDone, // Termination signal for listeners
			c.config.ASCIIMode,
			c.config.SkipTUIDelay,
			c.config.TUIConfig, // Pass through TUI config (nil uses defaults)
		)
	}

	// Build program options based on environment
	opts := []tea.ProgramOption{
		tea.WithAltScreen(),         // Take over screen, restore on exit
		tea.WithoutBracketedPaste(),
		tea.WithoutSignalHandler(), // Let our custom signal handler catch Ctrl-C
		// CRITICAL: Use the real stdout captured at console creation time.
		// This ensures TUI output goes directly to terminal even when stdout
		// is redirected by the output buffer for capturing subprocess leaks.
		tea.WithOutput(c.realStdout),
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
	// TUI is a pure observer - workers must NEVER block on TUI.
	// Two channels: critical (state changes, large buffer) and regular (output, may drop).
	c.msgChan = make(chan tea.Msg, 500)        // Regular messages - output lines
	c.criticalChan = make(chan tea.Msg, 10000) // Critical messages - state changes (large buffer)
	c.msgWg.Add(1)
	go func() {
		defer c.msgWg.Done()
		for {
			// Priority drain: always check critical channel first
			select {
			case msg, ok := <-c.criticalChan:
				if !ok {
					// Critical channel closed, drain remaining regular messages
					for msg := range c.msgChan {
						c.program.Send(msg)
					}
					return
				}
				c.program.Send(msg)
			default:
				// No critical messages, check both channels
				select {
				case msg, ok := <-c.criticalChan:
					if !ok {
						for msg := range c.msgChan {
							c.program.Send(msg)
						}
						return
					}
					c.program.Send(msg)
				case msg, ok := <-c.msgChan:
					if !ok {
						// Regular channel closed, drain critical
						for msg := range c.criticalChan {
							c.program.Send(msg)
						}
						return
					}
					c.program.Send(msg)
				}
			}
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
			atomic.StoreInt32(&c.atomicStopped, 1)

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

		// Close async message channels and wait for pump goroutine.
		// This prevents Send() calls to a completed bubbletea program,
		// which would block indefinitely and leak the pump goroutine.
		c.closeMsgChannelsOnce()
		c.msgWg.Wait()

		// Reset terminal state - always run this to ensure clean terminal
		// \033[0m = reset attributes, \033[?25h = show cursor
		fmt.Fprint(c.realStdout, "\033[0m\033[?25h")
	}()

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

	// Run the TUI and capture final model state
	finalModel, err := c.program.Run()
	if err != nil {
		return err
	}

	// Store final model for post-exit summary
	// Handle both console.Model and demo.Model
	switch m := finalModel.(type) {
	case console.Model:
		c.mu.Lock()
		c.finalModel = &m
		c.mu.Unlock()
		// Print plain-text summary after alt screen is restored
		c.printSummary(&m)
	case demo.Model:
		// tui3 handles its own final view rendering in View()
		// For now, just print completion message
		fmt.Fprintln(c.realStdout, "Execution complete.")
	}

	return nil
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

	// Close async message channels and wait for pending messages.
	// This ensures all queued TUI updates are delivered before exit.
	c.closeMsgChannelsOnce()
	c.msgWg.Wait()

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

// closeMsgChannelsOnce safely closes the async message channels exactly once.
// Safe to call from both Start() defer and Stop() without double-close panic.
func (c *ParallelConsole) closeMsgChannelsOnce() {
	c.msgCloseOnce.Do(func() {
		close(c.criticalChan)
		close(c.msgChan)
	})
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
		Phase:                     status.Phase,
		Running:                   status.Running,
		Completed:                 status.Completed,
		Total:                     status.Total,
		Roof:                      status.Roof,
		PressureTarget:            status.PressureTarget,
		Locks:                     locks,
		ActiveContainerTools:      status.ActiveContainerTools,
		UsedContainerTools:        status.UsedContainerTools,
		ActiveSystemTools:         status.ActiveSystemTools,
		UsedSystemTools:           status.UsedSystemTools,
		DockerMemPercent:          status.DockerMemPercent,
		DockerAvailable:           status.DockerAvailable,
		RunningContainerCount:     status.RunningContainerCount,
		TotalContainerCount:       status.TotalContainerCount,
		RunningSystemCount:        status.RunningSystemCount,
		TotalSystemCount:          status.TotalSystemCount,
	}

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

// sendCritical sends a message that must not be dropped (state changes).
// Uses a dedicated critical message channel with larger buffer.
// Never blocks - relies on large buffer to avoid drops.
func (c *ParallelConsole) sendCritical(msg tea.Msg) {
	c.mu.Lock()
	if c.stopped || c.program == nil || c.criticalChan == nil {
		c.mu.Unlock()
		return
	}
	critChan := c.criticalChan
	c.mu.Unlock()

	select {
	case critChan <- msg:
		// Queued successfully
	default:
		// Buffer full - this should rarely happen with large buffer
		// Fall back to regular channel as last resort (may drop)
		c.sendAsync(msg)
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
	critChan := c.criticalChan
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

	// Blocking send to critical channel — guarantees delivery AND ordering
	// after all preceding completion events already in the channel.
	if critChan != nil {
		critChan <- console.SummaryDataMsg{
			Data: (*console.SummaryData)(data),
		}
	} else {
		// Fallback: direct send if channels not initialized
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

// printSummary prints a plain-text summary after the TUI exits.
func (c *ParallelConsole) printSummary(m *console.Model) {
	// Use the console package's ViewFinal method to generate plain-text output
	summary := m.ViewFinal()
	fmt.Fprint(c.realStdout, summary)
}

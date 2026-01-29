// Package defaulttui provides a simple interactive TUI for command selection.
// It is used as the fallback TUI for commands without specific TUI bindings.
package defaulttui

import (
	"context"
	"io"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ready-to-release/eac/go/eac/commands/tui"
)

// Console implements tui.Console with a simple interactive interface.
// It provides:
// - Command name and description display
// - Subcommand list with navigation
// - Enter to select and execute subcommand
type Console struct {
	config  tui.Config
	program *tea.Program

	mu      sync.Mutex
	started bool
	stopped bool
	ready   chan struct{}
	done    chan struct{}

	// Interactive state
	subcommands []tui.SubcommandInfo
	selected    int // -1 = no selection, 0+ = selected index
	params      map[string]string
}

// New creates a new default Console.
func New(config tui.Config) *Console {
	config = config.WithDefaults()

	return &Console{
		config:   config,
		ready:    make(chan struct{}),
		done:     make(chan struct{}),
		params:   make(map[string]string),
		selected: -1, // No selection initially
	}
}

// Factory returns a ConsoleFactory for the default TUI.
func Factory() tui.ConsoleFactory {
	return func(config tui.Config) tui.Console {
		return New(config)
	}
}

// Start starts the TUI program. This blocks until the program exits.
func (c *Console) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return nil
	}
	c.started = true
	c.mu.Unlock()

	defer close(c.done)

	model := NewModel(c.config, c.subcommands)
	c.program = tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	close(c.ready)

	finalModel, err := c.program.Run()
	if err != nil {
		return err
	}

	// Capture the selected command and arguments
	if m, ok := finalModel.(Model); ok {
		c.mu.Lock()
		if m.IsExecuting() {
			c.selected = m.selected
			if args := m.GetArguments(); args != "" {
				c.params["args"] = args
			}
		}
		c.mu.Unlock()
	}

	return nil
}

// StartAsync starts the TUI program in a goroutine.
func (c *Console) StartAsync(ctx context.Context) {
	go func() {
		_ = c.Start(ctx)
	}()
	<-c.ready
}

// Wait waits for the TUI program to fully complete.
func (c *Console) Wait() {
	c.mu.Lock()
	started := c.started
	c.mu.Unlock()

	if started {
		<-c.done
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

	if c.program != nil {
		c.program.Quit()
		c.program.Wait()
	}
	<-c.done
}

// NewWriter returns the log writer (no TUI streaming for default TUI).
func (c *Console) NewWriter(source string, logWriter io.Writer) io.Writer {
	return logWriter
}

// SendLine is a no-op for default TUI.
func (c *Console) SendLine(line tui.Line) {}

// WriteResult is a no-op for default TUI.
func (c *Console) WriteResult(text string) {}

// UpdateStatus is a no-op for default TUI.
func (c *Console) UpdateStatus(status tui.Status) {}

// SetPhase is a no-op for default TUI.
func (c *Console) SetPhase(phase tui.Phase) {}

// CompletePhase is a no-op for default TUI.
func (c *Console) CompletePhase(phase tui.Phase, success bool, summary string) {}

// WriteToPhase is a no-op for default TUI.
func (c *Console) WriteToPhase(phase tui.Phase, text string) {}

// SetPhaseSummary is a no-op for default TUI.
func (c *Console) SetPhaseSummary(phase tui.Phase, summary string) {}

// StartModule is a no-op for default TUI.
func (c *Console) StartModule(moniker string, weight int) {}

// MarkModuleRunning is a no-op for default TUI.
func (c *Console) MarkModuleRunning(moniker string) {}

// MarkModuleComplete is a no-op for default TUI.
func (c *Console) MarkModuleComplete(moniker string, exitCode int) {}

// MarkModuleCompleteWithCacheInfo is a no-op for default TUI.
func (c *Console) MarkModuleCompleteWithCacheInfo(moniker string, exitCode int, cacheTime time.Time, logPath string) {
}

// SendSummary is a no-op for default TUI.
func (c *Console) SendSummary(data *tui.SummaryData) {}

// SetInitSummary is a no-op for default TUI.
func (c *Console) SetInitSummary(summary *tui.InitSummary) {}

// SetSubcommands sets the available subcommands for selection.
func (c *Console) SetSubcommands(commands []tui.SubcommandInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.subcommands = commands
}

// GetSelectedCommand returns the selected command and parameters.
func (c *Console) GetSelectedCommand() (string, map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.selected >= 0 && c.selected < len(c.subcommands) {
		return c.subcommands[c.selected].Name, c.params
	}
	return "", nil
}

// ShowParameterEditor is not yet implemented.
func (c *Console) ShowParameterEditor(params []tui.ParameterInfo) {
	// Future: implement parameter editing
}

// Compile-time check that Console implements tui.Console.
var _ tui.Console = (*Console)(nil)

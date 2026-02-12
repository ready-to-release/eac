package console

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/shirou/gopsutil/v3/cpu"

	tui "github.com/ready-to-release/eac/contracts/tui/0.1.0"
	"github.com/ready-to-release/eac/go/core/logging"
)

var log = logging.C()

// Model is the Bubbletea model for the console window.
// Displays build/test output in a 3-pane view (Init/Run/Summary).
type Model struct {
	Boot        ModelBootState
	Display     DisplayConfig
	Execution   ExecutionState
	Interaction InteractionState
	Resources   ResourceState
}

// ModelOptions holds all parameters for creating a new console Model.
// This replaces the previous 8 positional parameters for clarity and extensibility.
type ModelOptions struct {
	Height       int            // Total height (0 = default 5)
	RunPhaseName string         // Custom name for Run phase (e.g., "building", "testing")
	LineChan     <-chan Line     // Incoming output lines
	StatusChan   <-chan Status   // Status updates
	DoneChan     <-chan struct{} // Termination signal - closing it signals listeners to stop
	ASCIIMode    bool           // Use ASCII-only characters
	SkipTUIDelay bool           // Skip user interaction tracking, exit immediately when done
	TUIConfig    *tui.TUIConfig // TUI configuration (nil uses defaults)
}

// NewModel creates a new console model.
// If opts.TUIConfig is nil, defaults are used.
// opts.DoneChan must be provided - closing it signals listeners to stop,
// preventing blocking reads on empty channels when work is complete.
func NewModel(opts ModelOptions) Model {
	height := opts.Height
	runPhaseName := opts.RunPhaseName
	lineChan := opts.LineChan
	statusChan := opts.StatusChan
	doneChan := opts.DoneChan
	asciiMode := opts.ASCIIMode
	skipTUIDelay := opts.SkipTUIDelay
	tuiConfig := opts.TUIConfig
	// Initialize zone manager for mouse click tracking
	zone.NewGlobal()

	if height <= 0 {
		height = 5
	}

	// Use default if no custom run phase name provided
	if runPhaseName == "" {
		runPhaseName = "Run"
	}

	// Apply TUI config defaults if nil
	if tuiConfig == nil {
		tuiConfig = tui.DefaultTUIConfig()
	}

	// Create panes with appropriate buffer sizes
	bufferSize := tuiConfig.BufferSizePane
	if bufferSize <= 0 {
		bufferSize = 500 // Fallback
	}
	panes := [3]*Pane{
		NewPane(PhaseInit, bufferSize),
		NewPane(PhaseRun, bufferSize),
		NewPane(PhaseSummary, bufferSize),
	}

	// Get values from config with fallbacks
	resultsBufferSize := tuiConfig.BufferSizeResults
	if resultsBufferSize <= 0 {
		resultsBufferSize = 100
	}
	maxTabs := tuiConfig.MaxTabs
	if maxTabs <= 0 {
		maxTabs = 36
	}
	paneWidthCols := tuiConfig.DefaultColumns
	if paneWidthCols <= 0 {
		paneWidthCols = 4
	}
	freezeTimeoutSecs := int(tuiConfig.ExitCountdown.Seconds())
	if freezeTimeoutSecs <= 0 {
		freezeTimeoutSecs = 10
	}
	metricsInterval := tuiConfig.MetricsInterval
	if metricsInterval <= 0 {
		metricsInterval = 500 * time.Millisecond
	}
	minDisplayTime := tuiConfig.MinDisplayTime
	if minDisplayTime <= 0 {
		minDisplayTime = 1500 * time.Millisecond
	}
	autoScrollResume := tuiConfig.AutoScrollResume
	if autoScrollResume <= 0 {
		autoScrollResume = 8 * time.Second
	}
	bufferSizeUoW := tuiConfig.BufferSizeUoW
	if bufferSizeUoW <= 0 {
		bufferSizeUoW = 200
	}

	// Prime CPU metrics - gopsutil needs a baseline sample for cpu.Percent(0, true)
	// to return meaningful data. Without this, the first call returns empty/zero.
	_, _ = cpu.Percent(0, true)

	// Initialize widget catalog
	catalog := NewWidgetCatalog()
	RegisterAllWidgets(catalog)

	return Model{
		Boot: ModelBootState{
			State: BootChrome, // Start with structural chrome only
		},
		Display: DisplayConfig{
			Height:                height,
			Width:                 80, // Default, will be updated on WindowSizeMsg
			RunPhaseName:          runPhaseName,
			AsciiMode:             asciiMode,
			MetricsUpdateInterval: metricsInterval,
			MinDisplayTime:        minDisplayTime,
			AutoScrollResume:      autoScrollResume,
			BufferSizeUoW:         bufferSizeUoW,
		},
		Execution: ExecutionState{
			Panes:         panes,
			ActivePhase:   PhaseInit, // Start with Init phase
			ResultsBuffer: NewRingBuffer(resultsBufferSize),
			LineChan:      lineChan,
			StatusChan:    statusChan,
			DoneChan:      doneChan,
			StartTime:     time.Now(),
			UoWStates:     make(map[string]*UoWState), // Per-module state tracking
			UoWOrder:      make([]string, 0),          // Tab ordering
		},
		Interaction: InteractionState{
			MouseMode:         true, // Start with mouse ON (scrolling enabled)
			ActiveTab:         "",   // Start with aggregate view
			MaxTabs:           maxTabs,
			TabWidth:          tabWidthDefault,
			PaneWidthCols:     paneWidthCols,
			SkipTUIDelay:      skipTUIDelay,
			FreezeTimeoutSecs: freezeTimeoutSecs,
		},
		Resources: ResourceState{
			Catalog: catalog,
		},
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.listenForLines(),
		m.listenForStatus(),
		m.tickCmd(),
	)
}

// listenForLines creates a command that waits for new output lines.
// Returns batchLineMsg with up to 50 lines to reduce Update+View cycles.
// Returns linesDoneMsg when lineChan is closed or doneChan is closed.
func (m Model) listenForLines() tea.Cmd {
	return func() tea.Msg {
		if m.Execution.LineChan == nil {
			return linesDoneMsg{}
		}
		// Block until at least one line is available
		select {
		case line, ok := <-m.Execution.LineChan:
			if !ok {
				return linesDoneMsg{}
			}
			batch := make([]Line, 1, 50)
			batch[0] = line
			// Drain up to 49 more lines non-blocking
			for i := 0; i < 49; i++ {
				select {
				case line, ok := <-m.Execution.LineChan:
					if !ok {
						return batchLineMsg(batch)
					}
					batch = append(batch, line)
				default:
					return batchLineMsg(batch)
				}
			}
			return batchLineMsg(batch)
		case <-m.Execution.DoneChan:
			// Drain any remaining buffered lines before signaling done.
			// Go's select randomly picks between ready cases, so doneChan
			// may fire while lineChan still has buffered lines. Without
			// draining, those lines are permanently lost.
			var remaining []Line
		drain:
			for {
				select {
				case line, ok := <-m.Execution.LineChan:
					if !ok {
						break drain
					}
					remaining = append(remaining, line)
				default:
					break drain
				}
			}
			if len(remaining) > 0 {
				return batchLineMsg(remaining)
			}
			return linesDoneMsg{}
		}
	}
}

// listenForStatus creates a command that waits for status updates.
// Returns statusDoneMsg when statusChan is closed or doneChan is closed.
func (m Model) listenForStatus() tea.Cmd {
	return func() tea.Msg {
		if m.Execution.StatusChan == nil {
			return statusDoneMsg{}
		}
		select {
		case status, ok := <-m.Execution.StatusChan:
			if !ok {
				return statusDoneMsg{}
			}
			return statusMsg(status)
		case <-m.Execution.DoneChan:
			return statusDoneMsg{}
		}
	}
}

// tickCmd returns a tick every 100ms for elapsed time updates.
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

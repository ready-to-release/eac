package shared

import "time"

// UnitStatus represents execution state of a unit (component).
type UnitStatus int

const (
	UnitPending UnitStatus = iota
	UnitQueued
	UnitRunning
	UnitSuccess
	UnitSkipped
	UnitFailed
)

// String returns a human-readable status name.
func (s UnitStatus) String() string {
	switch s {
	case UnitPending:
		return "Pending"
	case UnitQueued:
		return "Queued"
	case UnitRunning:
		return "Running"
	case UnitSuccess:
		return "Success"
	case UnitSkipped:
		return "Skipped"
	case UnitFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// Icon returns an ASCII-safe icon for the status.
func (s UnitStatus) Icon() string {
	switch s {
	case UnitPending:
		return "o"
	case UnitQueued:
		return "*"
	case UnitRunning:
		return ">"
	case UnitSuccess:
		return "V"
	case UnitSkipped:
		return "="
	case UnitFailed:
		return "X"
	default:
		return "?"
	}
}

// UnicodeIcon returns a Unicode icon for the status.
func (s UnitStatus) UnicodeIcon() string {
	switch s {
	case UnitPending:
		return "o"
	case UnitQueued:
		return "@"
	case UnitRunning:
		return ">"
	case UnitSuccess:
		return "V"
	case UnitSkipped:
		return ">>"
	case UnitFailed:
		return "X"
	default:
		return "?"
	}
}

// UnitState tracks individual unit execution state.
type UnitState struct {
	Moniker   string
	Module    string
	Component string
	Handler   string
	Status    UnitStatus
	Weight    int
	StartTime time.Time
	EndTime   time.Time
	ExitCode  int
}

// Duration returns the execution duration.
func (u *UnitState) Duration() time.Duration {
	if u.StartTime.IsZero() {
		return 0
	}
	if !u.EndTime.IsZero() {
		return u.EndTime.Sub(u.StartTime)
	}
	return time.Since(u.StartTime)
}

// UnitDetails provides detailed info for the selected unit panel.
type UnitDetails struct {
	Moniker      string
	Module       string
	Component    string
	Handler      string
	Status       UnitStatus
	Weight       int
	Duration     time.Duration
	Dependencies []string
}

// ExecutionLayer represents a single execution layer with its modules.
type ExecutionLayer struct {
	Index   int
	Modules []ExecutionModule
}

// ExecutionModule represents a module and its components within a layer.
type ExecutionModule struct {
	Name       string
	Components []string
}

// Direction represents navigation direction for keyboard input.
type Direction int

const (
	DirUp Direction = iota
	DirDown
	DirLeft
	DirRight
)

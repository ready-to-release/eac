// Package tui provides terminal user interface implementations.
package tui

import (
	tuicontract "github.com/ready-to-release/eac/contracts/tui-adapter/0.1.0/interfaces"
)

// Re-export types from contracts for convenience.
// This allows consumers to import just the adapter package.

// Console is the primary interface for all TUI implementations.
type Console = tuicontract.Console

// ConsoleFactory creates Console instances with the given configuration.
type ConsoleFactory = tuicontract.ConsoleFactory

// Registry holds command-to-TUI bindings.
type Registry = tuicontract.Registry

// Config configures TUI console behavior.
type Config = tuicontract.Config

// Line represents a line of output from a command.
type Line = tuicontract.Line

// Level represents the severity level of an output line.
type Level = tuicontract.Level

// Status represents a status update from the orchestrator.
type Status = tuicontract.Status

// LockStatus represents the state of a single lock.
type LockStatus = tuicontract.LockStatus

// Phase represents the three execution phases.
type Phase = tuicontract.Phase

// PhaseStatus represents the status of a phase.
type PhaseStatus = tuicontract.PhaseStatus

// SummaryData holds structured summary information.
type SummaryData = tuicontract.SummaryData

// InitSummary holds structured init summary data.
type InitSummary = tuicontract.InitSummary

// ExecutionModule represents a module and its UoWs.
type ExecutionModule = tuicontract.ExecutionModule

// UoWEntry represents a unit of work with its globally unique ID.
type UoWEntry = tuicontract.UoWEntry

// InitSummaryFlags captures relevant flags for display.
type InitSummaryFlags = tuicontract.InitSummaryFlags

// PlannedTool represents a tool that will be used during execution.
type PlannedTool = tuicontract.PlannedTool

// SubcommandInfo describes a subcommand for display/conversion purposes.
type SubcommandInfo = tuicontract.SubcommandInfo

// Binding represents a resolved command-to-TUI binding.
type Binding = tuicontract.Binding

// Level constants for convenience.
const (
	LevelInfo  = tuicontract.LevelInfo
	LevelWarn  = tuicontract.LevelWarn
	LevelError = tuicontract.LevelError
)

// Phase constants for convenience.
const (
	PhaseInit    = tuicontract.PhaseInit
	PhaseRun     = tuicontract.PhaseRun
	PhaseSummary = tuicontract.PhaseSummary
)

// PhaseStatus constants for convenience.
const (
	PhasePending  = tuicontract.PhasePending
	PhaseActive   = tuicontract.PhaseActive
	PhaseComplete = tuicontract.PhaseComplete
	PhaseFailed   = tuicontract.PhaseFailed
)

// Config defaults for convenience.
const (
	DefaultHeight     = tuicontract.DefaultHeight
	DefaultBufferSize = tuicontract.DefaultBufferSize
)

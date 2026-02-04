package cells

// CommandCell displays the synthetic command line with module list.
type CommandCell struct {
	command string
	modules []string
}

// NewCommandCell creates a new command cell.
func NewCommandCell() *CommandCell {
	return &CommandCell{}
}

// SetCommand sets the command name (e.g., "build", "test", "lint").
func (c *CommandCell) SetCommand(cmd string) {
	c.command = cmd
}

// SetModules sets the list of module names.
func (c *CommandCell) SetModules(modules []string) {
	c.modules = modules
}

// Render returns the command line display, truncated if necessary.
func (c *CommandCell) Render(width, height int) string {
	if c.command == "" {
		return ""
	}

	result := c.command
	truncated := false

	// Append modules until width exceeded
	for i, mod := range c.modules {
		candidate := result + " " + mod
		if len(candidate) > width-4 { // Leave room for " ..."
			truncated = true
			break
		}
		result = candidate
		// Mark as not truncated if we've included all modules
		if i == len(c.modules)-1 {
			truncated = false
		}
	}

	// Only add "..." if we didn't include all modules
	if truncated {
		return result + " ..."
	}

	return result
}

// ZoneID returns the mouse zone identifier.
func (c *CommandCell) ZoneID() string {
	return "res-command"
}

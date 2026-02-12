package registry

import (
	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

// Register adds a command to the registry. Returns ErrDuplicateCommand if the name is taken.
func (r *CommandRegistry) Register(cmd core.CommandPort) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		return &ErrDuplicateCommand{Name: name}
	}
	r.commands[name] = cmd
	return nil
}

// MustRegister registers a command, panicking on duplicate.
func (r *CommandRegistry) MustRegister(cmd core.CommandPort) {
	if err := r.Register(cmd); err != nil {
		panic(err)
	}
}

// RegisterAll registers multiple commands. Stops and returns the first error encountered.
func (r *CommandRegistry) RegisterAll(cmds ...core.CommandPort) error {
	for _, cmd := range cmds {
		if err := r.Register(cmd); err != nil {
			return err
		}
	}
	return nil
}

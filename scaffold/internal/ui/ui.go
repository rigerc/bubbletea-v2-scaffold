// Package ui provides the TUI entry point for scaffold.
package ui

import (
	tea "charm.land/bubbletea/v2"

	"scaffold/config"
)

// New creates a new root model from the config.
func New(cfg config.Config) rootModel {
	return newRootModel(cfg)
}

// Run starts the TUI program.
func Run(m rootModel) error {
	_, err := tea.NewProgram(m).Run()
	return err
}

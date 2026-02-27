// Package ui provides the TUI entry point for scaffold.
package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"scaffold/config"
)

// New creates a new root model from the config.
// configPath is the path to persist settings; empty means no file save.
// firstRun indicates that no config file existed before this launch.
func New(cfg config.Config, configPath string, firstRun bool) rootModel {
	ctx, cancel := context.WithCancel(context.Background())
	return newRootModel(ctx, cancel, cfg, configPath, firstRun)
}

// Run starts the TUI program. ctx is used to cancel background goroutines on quit.
func Run(ctx context.Context, m rootModel) error {
	_, err := tea.NewProgram(m, tea.WithContext(ctx)).Run()
	return err
}

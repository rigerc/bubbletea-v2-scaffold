// Package spinner provides a thin, theme-aware wrapper around bubbles/spinner.
// Screens that run background tasks embed this model and delegate Update/View.
package spinner

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/theme"
)

// Model wraps bubbles spinner with theme-aware styling.
type Model struct {
	s spinner.Model
}

// New creates a spinner styled with the given palette's primary colour.
func New(p theme.Palette) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(p.Primary)
	return Model{s: s}
}

// Init returns the command that starts the tick loop.
func (m Model) Init() tea.Cmd {
	return m.s.Tick
}

// Update forwards messages to the inner spinner and returns updated state.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.s, cmd = m.s.Update(msg)
	return m, cmd
}

// View renders the current spinner frame.
func (m Model) View() string {
	return m.s.View()
}

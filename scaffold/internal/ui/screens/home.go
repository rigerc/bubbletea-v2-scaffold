// Package screens provides individual screen components for the TUI.
package screens

import tea "charm.land/bubbletea/v2"

// Screen is the interface for screen components that can be composed.
type Screen interface {
	tea.Model
	Body() string // Returns body content for layout composition
}

// Home is the home screen.
type Home struct {
	width int
}

// NewHome creates a new Home screen.
func NewHome() Home {
	return Home{}
}

// SetWidth sets the screen width.
func (h Home) SetWidth(w int) Home {
	h.width = w
	return h
}

// Init initializes the home screen.
func (h Home) Init() tea.Cmd {
	return nil
}

// Update handles messages for the home screen.
func (h Home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

// View renders the home screen.
func (h Home) View() tea.View {
	return tea.NewView(h.Body())
}

// Body returns the body content for layout composition.
func (h Home) Body() string {
	return "Welcome to scaffold.\n\nReplace this placeholder with your application views."
}

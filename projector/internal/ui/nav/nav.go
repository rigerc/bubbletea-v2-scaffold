// Package nav provides the navigation/routing framework for the application.
// It defines the Screen interface and navigation messages for stack-based routing.
package nav

import (
	tea "charm.land/bubbletea/v2"
)

// Screen is implemented by every navigable view in the application.
// It follows the Elm Architecture pattern: Init, Update, View.
//
// Unlike the standard tea.Model interface, Screen.Update returns (Screen, tea.Cmd)
// instead of (tea.Model, tea.Cmd). This allows screens to be stored in a slice
// while maintaining type safety.
type Screen interface {
	// Init returns the initial command for the screen.
	Init() tea.Cmd

	// Update handles incoming messages and returns an updated screen and command.
	Update(tea.Msg) (Screen, tea.Cmd)

	// View renders the screen content as a string.
	// The returned string is wrapped by the root Model with theme styling.
	View() string
}

// Themeable is an optional interface that screens can implement.
// The router calls SetTheme on push and on BackgroundColorMsg to allow
// screens to adapt their styling to light/dark terminal backgrounds.
type Themeable interface {
	SetTheme(isDark bool)
}

// Navigation message types for stack-based routing.

// PushMsg adds a new screen to the top of the navigation stack.
type PushMsg struct {
	Screen Screen
}

// PopMsg removes the current screen from the navigation stack.
type PopMsg struct{}

// ReplaceMsg replaces the current screen with a new one.
type ReplaceMsg struct {
	Screen Screen
}

// Push returns a Cmd that sends PushMsg, adding a new screen to the stack.
func Push(s Screen) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Screen: s}
	}
}

// Pop returns a Cmd that sends PopMsg, removing the current screen from the stack.
func Pop() tea.Cmd {
	return func() tea.Msg {
		return PopMsg{}
	}
}

// Replace returns a Cmd that sends ReplaceMsg, replacing the current screen.
func Replace(s Screen) tea.Cmd {
	return func() tea.Msg {
		return ReplaceMsg{Screen: s}
	}
}

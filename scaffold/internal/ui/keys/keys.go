// Package keys provides global key bindings for the TUI.
package keys

import "charm.land/bubbles/v2/key"

// GlobalKeyMap holds global key bindings.
type GlobalKeyMap struct {
	Quit key.Binding
}

// DefaultGlobalKeyMap returns the default global key bindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

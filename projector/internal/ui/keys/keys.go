// Package keys defines application-wide key bindings.
package keys

import "charm.land/bubbles/v2/key"

// GlobalKeyMap defines the application-wide key bindings.
// It implements help.KeyMap so it can be passed directly to help.View().
type GlobalKeyMap struct {
	Back key.Binding // "esc"    — go to previous screen
	Quit key.Binding // "ctrl+c" — always quit (no conflict with list filter "q")
	Help key.Binding // "?"      — toggle help expansion
}

// New returns a GlobalKeyMap with default bindings.
func New() GlobalKeyMap {
	return GlobalKeyMap{
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}
}

// ShortHelp returns the short-form key bindings for the help bar.
// Implements help.KeyMap.
func (k GlobalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Quit}
}

// FullHelp returns the full-form key bindings for the expanded help view.
// Implements help.KeyMap.
func (k GlobalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Back, k.Help}, {k.Quit}}
}

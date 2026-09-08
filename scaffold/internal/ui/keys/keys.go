// Package keys provides global key bindings for the TUI.
package keys

import "charm.land/bubbles/v2/key"

// GlobalKeyMap holds global key bindings.
type GlobalKeyMap struct {
	Quit        key.Binding
	Back        key.Binding
	RandomTheme key.Binding // hidden
}

// DefaultGlobalKeyMap returns the default global key bindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q/ctrl+c", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		RandomTheme: key.NewBinding(
			key.WithKeys("ctrl+t"),
		),
	}
}

// ShortHelp returns a slice of bindings for short help view.
func (k GlobalKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Quit}
}

// FullHelp returns grouped bindings for full help view.
func (k GlobalKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Back, k.Quit}}
}

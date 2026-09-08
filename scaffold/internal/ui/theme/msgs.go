package theme

import tea "charm.land/bubbletea/v2"

// ThemeChangedMsg is broadcast when theme state changes.
type ThemeChangedMsg struct {
	State State
}

// RequestThemeUpdate returns a command that broadcasts the current theme.
func RequestThemeUpdate(state State) tea.Cmd {
	return func() tea.Msg {
		return ThemeChangedMsg{State: state}
	}
}

// Package ui — View rendering helpers for rootModel.
package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/keys"
	"scaffold/internal/ui/screens"
)

// helpView renders the persistent help box showing global and screen-specific keybindings.
func (m rootModel) helpView() string {
	combined := m.combinedKeys()
	return m.styles.Help.Render(m.help.View(combined))
}

// combinedKeys returns a key map that combines global keys with screen-specific keys.
func (m rootModel) combinedKeys() combinedKeyMap {
	return combinedKeyMap{
		global: m.keys,
		screen: m.current,
	}
}

// combinedKeyMap combines global and screen-specific key bindings.
type combinedKeyMap struct {
	global keys.GlobalKeyMap
	screen screens.Screen
}

// ShortHelp returns combined short help bindings.
func (c combinedKeyMap) ShortHelp() []key.Binding {
	bindings := c.global.ShortHelp()
	if kb, ok := c.screen.(screens.KeyBinder); ok {
		bindings = append(bindings, kb.ShortHelp()...)
	}
	return bindings
}

// FullHelp returns combined full help bindings.
func (c combinedKeyMap) FullHelp() [][]key.Binding {
	groups := c.global.FullHelp()
	if kb, ok := c.screen.(screens.KeyBinder); ok {
		groups = append(groups, kb.FullHelp()...)
	}
	return groups
}

// Layout constants document the fixed-height chrome components.
// Header and help heights are dynamic (banner height varies; help wraps at
// narrow terminals), so they are measured at runtime and cached in rootModel.bodyH.
const (
	// footerLines is the number of terminal lines the footer chrome occupies.
	footerLines = 1
	// minBodyLines is the minimum body height guaranteed by View().
	minBodyLines = 1
	// maxBodyPercent is the maximum percentage of terminal height the body can occupy.
	maxBodyPercent = 60
)

// bodyHeight estimates the available height for the body content area.
// It subtracts the header, help, and footer chrome from the terminal height,
// then caps the result at maxBodyPercent of the terminal height.
func (m rootModel) bodyHeight() int {
	if m.height == 0 {
		return 0
	}
	helpH := lipgloss.Height(m.helpView())
	body := m.height - m.header.Height() - helpH - footerLines

	// Cap at maxBodyPercent of terminal height
	maxBody := m.height * maxBodyPercent / 100
	if body > maxBody {
		body = maxBody
	}

	if body < minBodyLines {
		body = minBodyLines
	}
	return body
}

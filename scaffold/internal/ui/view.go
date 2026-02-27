// Package ui — View rendering helpers for rootModel.
package ui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/banner"
	"scaffold/internal/ui/keys"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/theme"
)

// renderBanner renders the ASCII art banner at its natural width and caches the result.
// Using a large fixed width lets lipgloss.Width(m.banner) reflect the font's true width,
// which headerView uses to decide whether the terminal is wide enough to display it.
func (m *rootModel) renderBanner() {
	state := m.themeMgr.State()
	p := state.Palette
	if p.Primary == nil {
		p = theme.NewPalette(m.cfg.UI.ThemeName, state.IsDark)
	}
	b, err := banner.Render(banner.Config{
		Text:          m.cfg.App.Name,
		Font:          "larry3d",
		Width:         100,
		Justification: 0,
		Gradient:      banner.GradientThemed(p.Primary, p.Secondary),
	})
	if err != nil {
		b = m.cfg.App.Name
	}
	m.banner = b
}

// headerView renders the header with either the ASCII banner or a styled plain-text title.
// The ASCII banner is shown only when ShowBanner is enabled, the banner has been rendered,
// and the terminal is wide enough to display it. In all other cases — including when
// ShowBanner is disabled or the terminal is too narrow — the plain-text title is shown.
func (m rootModel) headerView() string {
	if m.cfg.UI.ShowBanner && m.banner != "" && m.width > 0 && m.width >= lipgloss.Width(m.banner) {
		return m.styles.Header.Render(m.banner)
	}
	return m.styles.Header.Render(m.plainTitleView())
}

// plainTitleView renders a styled plain-text title used when ShowBanner is off.
func (m rootModel) plainTitleView() string {
	return m.styles.PlainTitle.Render(m.cfg.App.Name)
}

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

// bodyHeight estimates the available height for the body content area.
// It subtracts the header, help, and footer chrome from the terminal height.
func (m rootModel) bodyHeight() int {
	if m.height == 0 {
		return 0
	}
	header := lipgloss.Height(m.headerView())
	helpH := lipgloss.Height(m.helpView())
	footer := lipgloss.Height(m.footerView())
	body := m.height - header - helpH - footer
	if body < 1 {
		body = 1
	}
	return body
}

// footerView renders the status bar footer.
func (m rootModel) footerView() string {
	left := m.statusStyles.Render(m.status.Text, m.status.Kind)
	rightContent := " v" + m.cfg.App.Version
	if m.cfg.Debug {
		rightContent += " [DEBUG]"
	}
	right := m.styles.StatusRight.Render(rightContent + " ")

	// Account for footer border (2) and padding (1)
	innerWidth := m.styles.MaxWidth - 3

	gap := lipgloss.NewStyle().
		Width(innerWidth - lipgloss.Width(left) - lipgloss.Width(right)).
		Render("")
	footerContent := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	return m.styles.Footer.Render(footerContent)
}

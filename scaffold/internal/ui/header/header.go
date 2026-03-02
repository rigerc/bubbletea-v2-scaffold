// Package header provides the self-contained header component for the TUI.
// It renders either an ASCII art banner (when enabled and wide enough) or a
// styled plain-text title, and updates itself on theme and window-size changes.
package header

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/ui/banner"
	"scaffold/internal/ui/theme"
)

// Model is the header component. All fields are unexported; callers interact
// through New, Update, View, Height, and WithCfg.
type Model struct {
	cfg        config.Config
	banner     string
	headerSty  lipgloss.Style
	titleSty   lipgloss.Style
	descSty    lipgloss.Style
	width      int
	themeState theme.State // cached for banner re-renders after config changes
}

// New creates a header Model from the given config.
// Styles and banner are populated on the first ThemeChangedMsg.
func New(cfg config.Config) Model {
	return Model{cfg: cfg}
}

// WithCfg returns a new Model with an updated config.
// If ShowBanner was just disabled the cached banner is cleared.
// If ShowBanner was just enabled and a theme state is available the banner is
// re-rendered immediately so the caller does not need to trigger a theme update.
func (m Model) WithCfg(cfg config.Config) Model {
	m.cfg = cfg
	if !cfg.UI.ShowBanner {
		m.banner = ""
	} else if m.banner == "" && m.themeState.Palette.Primary != nil {
		m.banner = renderBannerStr(cfg, m.themeState)
	}
	return m
}

// Update handles messages relevant to the header.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case theme.ThemeChangedMsg:
		m.themeState = msg.State
		p := msg.State.Palette

		m.headerSty = lipgloss.NewStyle().
			Padding(2).
			MarginBottom(0).
			PaddingBottom(3)

		m.titleSty = lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(p.Secondary)

		m.descSty = lipgloss.NewStyle().
			Foreground(p.Secondary).
			Bold(true).
			MarginLeft(3)

		if m.cfg.UI.ShowBanner {
			m.banner = renderBannerStr(m.cfg, msg.State)
		} else {
			m.banner = ""
		}
	}

	return m, nil
}

// View renders the header.
// The ASCII banner is shown only when ShowBanner is enabled, the banner has
// been rendered, and the terminal is wide enough. Otherwise a plain title is
// shown.
func (m Model) View() tea.View {
	var heading string
	if m.cfg.UI.ShowBanner && m.banner != "" && m.width > 0 && m.width >= lipgloss.Width(m.banner) {
		heading = m.banner
	} else {
		heading = m.titleSty.Render(m.cfg.App.Name)
	}
	if m.cfg.UI.ShowDescription && m.cfg.App.Description != "" {
		heading += "\n" + m.descSty.Render(m.cfg.App.Description)
	}
	return tea.NewView(m.headerSty.Render(heading))
}

// Height returns the number of terminal lines the header occupies.
func (m Model) Height() int {
	return lipgloss.Height(m.View().Content)
}

// renderBannerStr renders the ASCII art banner at a fixed large width and
// returns the result. Using a large width lets lipgloss.Width(banner) reflect
// the font's true natural width, which View uses to decide whether the terminal
// is wide enough to display it.
func renderBannerStr(cfg config.Config, state theme.State) string {
	p := state.Palette
	if p.Primary == nil {
		p = theme.NewPalette(cfg.UI.ThemeName, state.IsDark)
	}
	b, err := banner.Render(banner.Config{
		Text:          cfg.App.Name,
		Font:          "larry3d",
		Width:         100,
		Justification: 0,
		Gradient:      banner.GradientThemed(p.Primary, p.Secondary),
	})
	if err != nil {
		return cfg.App.Name
	}
	return b
}

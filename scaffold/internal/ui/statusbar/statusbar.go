// Package statusbar provides the self-contained footer / status-bar component
// for the TUI. It owns the status state and renders the full footer line
// including the left status message and the right version/debug indicator.
package statusbar

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/ui/status"
	"scaffold/internal/ui/theme"
)

// Model is the statusbar component.
type Model struct {
	state     status.State
	statusSty status.Styles
	footerSty lipgloss.Style
	rightSty  lipgloss.Style
	cfg       config.Config
	maxW      int
}

// New creates a statusbar Model. Styles are populated on the first
// ThemeChangedMsg; until then View returns an unstyled empty string.
func New(cfg config.Config) Model {
	return Model{
		cfg:   cfg,
		state: status.State{Text: "Ready", Kind: status.KindNone},
	}
}

// Update handles messages relevant to the statusbar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case status.Msg:
		m.state = status.State{Text: msg.Text, Kind: msg.Kind}

	case status.ClearMsg:
		m.state = status.State{Text: "Ready", Kind: status.KindNone}

	case theme.ThemeChangedMsg:
		p := msg.State.Palette

		m.statusSty = status.NewStyles(p)

		m.footerSty = lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(p.Border).
			PaddingLeft(1)

		m.rightSty = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)

		// Mirror the MaxWidth calculation from theme.newStylesFromPalette so that
		// the gap arithmetic matches the rest of the layout.
		w := msg.State.Width
		maxWidth := w * 90 / 100
		if maxWidth < 40 {
			maxWidth = w - 4
		}
		m.maxW = maxWidth
	}

	return m, nil
}

// State returns the current status state. Exposed for tests.
func (m Model) State() status.State {
	return m.state
}

// View renders the full footer: left status badge + spacer + right version text.
func (m Model) View() tea.View {
	left := m.statusSty.Render(m.state.Text, m.state.Kind)

	rightContent := " v" + m.cfg.App.Version
	if m.cfg.Debug {
		rightContent += " [DEBUG]"
	}
	right := m.rightSty.Render(rightContent + " ")

	// Account for footer border (2) and padding (1).
	innerWidth := m.maxW - 3
	gapW := max(0, innerWidth-lipgloss.Width(left)-lipgloss.Width(right))
	gap := lipgloss.NewStyle().Width(gapW).Render("")

	footerContent := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	return tea.NewView(m.footerSty.Render(footerContent))
}

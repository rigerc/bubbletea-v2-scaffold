// Package theme provides styling for the TUI.
package theme

import "charm.land/lipgloss/v2"

// Styles holds all styled components for the UI.
type Styles struct {
	App         lipgloss.Style
	Header      lipgloss.Style
	Body        lipgloss.Style
	Footer      lipgloss.Style
	StatusLeft  lipgloss.Style
	StatusRight lipgloss.Style
}

// New creates a Styles struct with adaptive colors based on the background.
func New(isDark bool, width int) Styles {
	ld := lipgloss.LightDark(isDark)
	subtle := ld(lipgloss.Color("#555555"), lipgloss.Color("#999999"))
	accent := lipgloss.Color("#7D56F4")
	fg := ld(lipgloss.Color("#1a1a1a"), lipgloss.Color("#f1f1f1"))

	// Max width is 70% of terminal width
	maxWidth := width * 70 / 100
	if maxWidth < 60 {
		maxWidth = width - 4 // Minimum usable width
	}

	return Styles{
		App: lipgloss.NewStyle().
			Width(maxWidth).
			Padding(0, 0),
		Header: lipgloss.NewStyle().
			Padding(0),
		Body: lipgloss.NewStyle().
			Padding(0, 3).
			Foreground(fg),
		Footer: lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(subtle).
			PaddingLeft(1),
		StatusLeft: lipgloss.NewStyle().
			Background(accent).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true),
		StatusRight: lipgloss.NewStyle().
			Foreground(subtle),
	}
}

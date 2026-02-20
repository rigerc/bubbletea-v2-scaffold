package theme

import (
	lipgloss "charm.land/lipgloss/v2"
)

// Theme contains all lipgloss styles used throughout the application.
// The Palette field provides access to semantic colors for dynamic styling.
type Theme struct {
	Palette ThemePalette

	// Container styles
	App   lipgloss.Style // Outer container with margin
	Panel lipgloss.Style // Bordered panel for grouped content

	// Typography
	Title  lipgloss.Style // Header/title bar with primary background
	Status lipgloss.Style // Status/informational text
	Subtle lipgloss.Style // De-emphasized text

	// UI elements
	Border  lipgloss.Style // Horizontal dividers and borders
	Error   lipgloss.Style // Error messages
	Warning lipgloss.Style // Warning messages
	Success lipgloss.Style // Success messages
}

// New creates a Theme with adaptive colors for the given background.
// The isDark parameter should come from tea.BackgroundColorMsg.IsDark().
func New(isDark bool) Theme {
	p := NewPalette(isDark)
	return Theme{
		Palette: p,

		// Container styles
		App: lipgloss.NewStyle().Margin(1, 2),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Primary).
			Padding(1),

		// Typography - ALL colors from palette
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.PrimaryFg). // From palette, not hardcoded
			Background(p.Primary).
			Padding(0, 1),
		Status: lipgloss.NewStyle().Foreground(p.Primary),
		Subtle: lipgloss.NewStyle().Foreground(p.Subtle),

		// UI elements - ALL colors from palette
		Border:  lipgloss.NewStyle().Foreground(p.Border),
		Error:   lipgloss.NewStyle().Foreground(p.Alert).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(p.Warning).Bold(true),
		Success: lipgloss.NewStyle().Foreground(p.Success).Bold(true),
	}
}

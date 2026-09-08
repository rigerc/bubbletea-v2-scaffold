package status

import (
	"charm.land/lipgloss/v2"
	"scaffold/internal/ui/theme"
)

// Styles provides styled rendering for status messages.
type Styles struct {
	Base    lipgloss.Style // Base style (shared)
	Info    lipgloss.Style // Info message styling
	Success lipgloss.Style // Success message styling
	Warning lipgloss.Style // Warning message styling
	Error   lipgloss.Style // Error message styling
}

// NewStyles creates status styles from a theme palette.
// Uses background colors for clear visual distinction of status types.
func NewStyles(p theme.Palette) Styles {
	base := lipgloss.NewStyle().Bold(true).Padding(0, 1)

	return Styles{
		Base:    base.Background(p.Primary).Foreground(p.OnPrimary),
		Info:    base.Background(p.Info).Foreground(p.OnInfo),
		Success: base.Background(p.Success).Foreground(p.OnSuccess),
		Warning: base.Background(p.Warning).Foreground(p.OnWarning),
		Error:   base.Background(p.Error).Foreground(p.OnError),
	}
}

// StyleFor returns the appropriate style for a status kind.
func (s Styles) StyleFor(kind Kind) lipgloss.Style {
	switch kind {
	case KindInfo:
		return s.Info
	case KindSuccess:
		return s.Success
	case KindWarning:
		return s.Warning
	case KindError:
		return s.Error
	default:
		return s.Base
	}
}

// Render renders the status text with appropriate styling.
func (s Styles) Render(text string, kind Kind) string {
	return s.StyleFor(kind).Render(" " + text + " ")
}

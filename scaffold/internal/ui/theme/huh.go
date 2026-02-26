package theme

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// HuhTheme returns a huh.Theme that matches the application palette.
// Uses huh.ThemeFunc so huh drives isDark on every View() call.
func HuhTheme() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		p := NewPalette(isDark)
		s := huh.ThemeCharm(isDark)

		s.Group.Title = s.Group.Title.Foreground(p.Primary).Bold(true)
		s.Group.Description = s.Group.Description.Foreground(p.TextSecondary)

		s.Focused.Base = s.Focused.Base.BorderForeground(p.Primary)
		s.Focused.Title = s.Focused.Title.Foreground(p.Primary)
		s.Focused.Description = s.Focused.Description.Foreground(p.TextSecondary)
		s.Focused.SelectSelector = s.Focused.SelectSelector.Foreground(p.PrimaryHover)
		s.Focused.NextIndicator = s.Focused.NextIndicator.Foreground(p.PrimaryHover)
		s.Focused.PrevIndicator = s.Focused.PrevIndicator.Foreground(p.PrimaryHover)
		s.Focused.FocusedButton = s.Focused.FocusedButton.Background(p.Primary).Foreground(p.TextInverse)
		s.Focused.TextInput.Cursor = s.Focused.TextInput.Cursor.Foreground(p.PrimaryHover)
		s.Focused.TextInput.Prompt = s.Focused.TextInput.Prompt.Foreground(p.Primary)
		s.Focused.ErrorMessage = s.Focused.ErrorMessage.Foreground(p.Error)
		s.Focused.ErrorIndicator = s.Focused.ErrorIndicator.Foreground(p.Error)

		s.Blurred = s.Focused
		s.Blurred.Base = s.Focused.Base.BorderStyle(lipgloss.HiddenBorder())

		return s
	})
}

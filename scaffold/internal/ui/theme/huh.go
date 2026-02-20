// Package theme provides Huh form integration for the application theme.
package theme

import (
	"charm.land/huh/v2"
)

// HuhThemeFunc returns a Huh theme that matches the application's visual style.
// The theme function uses the isDark parameter provided by Huh, ensuring colors
// stay in sync even if the terminal background changes.
//
// CRITICAL: This function creates its own ThemePalette using Huh's isDark parameter
// rather than using a cached Theme. This ensures colors stay synchronized even if
// the terminal background changes after the application starts.
//
// Usage: form.WithTheme(theme.HuhThemeFunc())
func HuhThemeFunc() huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		// Create palette using Huh's isDark parameter (not from cached Theme)
		p := NewPalette(isDark)

		// Start with Charm's default theme as base
		s := huh.ThemeCharm(isDark)

		// Apply app's green theme to titles - ALL colors from palette
		s.Group.Title = s.Group.Title.
			Foreground(p.PrimaryFg). // From palette, not hardcoded
			Background(p.Primary)    // From palette, not hardcoded

		// Match status messages to app theme
		s.Focused.Description = s.Focused.Description.
			Foreground(p.Primary)

		// Remove borders for cleaner look
		s.Form.Base = s.Form.Base.
			BorderTop(false).BorderRight(false).
			BorderBottom(false).BorderLeft(false)
		s.Group.Base = s.Group.Base.
			BorderTop(false).BorderRight(false).
			BorderBottom(false).BorderLeft(false)
		s.Focused.Base = s.Focused.Base.
			BorderTop(false).BorderRight(false).
			BorderBottom(false).BorderLeft(false)
		s.Blurred.Base = s.Blurred.Base.
			BorderTop(false).BorderRight(false).
			BorderBottom(false).BorderLeft(false)

		// Increase spacing between options
		s.Focused.Option = s.Focused.Option.Margin(2, 0)
		s.Blurred.Option = s.Blurred.Option.Margin(2, 0)

		return s
	})
}

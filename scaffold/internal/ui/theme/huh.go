package theme

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// HuhTheme returns a huh.Theme that matches the application palette for the given theme name.
// Uses huh.ThemeFunc so huh drives isDark on every View() call.
// Focused elements use Primary, unfocused use Secondary, descriptions use ForegroundMuted.
// No background colors are applied.
func HuhTheme(name string) huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		p := NewPalette(name, isDark)
		t := huh.ThemeCharm(isDark)

		// Focused state - use Primary for interactive elements
		t.Focused.Base = t.Focused.Base.BorderForeground(p.Border)
		t.Focused.Card = t.Focused.Base
		t.Focused.Title = t.Focused.Title.Foreground(p.Primary)
		t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(p.Primary)
		t.Focused.Directory = t.Focused.Directory.Foreground(p.Primary)
		t.Focused.Description = t.Focused.Description.Foreground(p.ForegroundMuted)
		t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(p.Error)
		t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(p.Error)
		t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(p.Primary)
		t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(p.Primary)
		t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(p.Primary)
		t.Focused.Option = t.Focused.Option.Foreground(p.Foreground)
		t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(p.Primary)
		t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(p.Success)
		t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(p.Success)
		t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(p.ForegroundSubtle)
		t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(p.Foreground)
		t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(p.Primary)
		t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(p.Secondary)

		// Text input styles
		t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(p.Primary)
		t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(p.ForegroundSubtle)
		t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(p.Primary)

		// Blurred state - use Secondary for unfocused items
		t.Blurred = t.Focused
		t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
		t.Blurred.Card = t.Blurred.Base
		t.Blurred.Title = t.Blurred.Title.Foreground(p.Secondary)
		t.Blurred.NoteTitle = t.Blurred.NoteTitle.Foreground(p.Secondary)
		t.Blurred.Directory = t.Blurred.Directory.Foreground(p.Secondary)
		t.Blurred.SelectSelector = t.Blurred.SelectSelector.Foreground(p.Secondary)
		t.Blurred.MultiSelectSelector = t.Blurred.MultiSelectSelector.Foreground(p.Secondary)
		t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(p.Secondary)

		// Help styles - use muted colors
		t.Help.Ellipsis = t.Help.Ellipsis.Foreground(p.ForegroundMuted)
		t.Help.ShortKey = t.Help.ShortKey.Foreground(p.ForegroundMuted)
		t.Help.ShortDesc = t.Help.ShortDesc.Foreground(p.ForegroundSubtle)
		t.Help.ShortSeparator = t.Help.ShortSeparator.Foreground(p.ForegroundMuted)
		t.Help.FullKey = t.Help.FullKey.Foreground(p.ForegroundMuted)
		t.Help.FullDesc = t.Help.FullDesc.Foreground(p.ForegroundSubtle)
		t.Help.FullSeparator = t.Help.FullSeparator.Foreground(p.ForegroundMuted)

		// Group styles
		t.Group.Title = t.Focused.Title
		t.Group.Description = t.Focused.Description

		return t
	})
}

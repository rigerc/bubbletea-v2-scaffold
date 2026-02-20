package screens

import (
	"strings"

	"charm.land/bubbles/v2/help"
	lipgloss "charm.land/lipgloss/v2"

	appkeys "projector/internal/ui/keys"
	"projector/internal/ui/theme"
)

// Sizing constants for consistent view dimensions across all screens.
const (
	// MaxContentHeight is the maximum height for any screen's main content area.
	// Views will be capped at this height or 1/3 of terminal height, whichever is smaller.
	MaxContentHeight = 25

	// MinContentHeight is the minimum height for any screen's main content area.
	// This ensures UI elements remain usable even on small terminals.
	MinContentHeight = 10
)

// ScreenBase holds state shared by every screen: adaptive theme, terminal
// dimensions, global key bindings, and a help bar component.
// Embed it in your Screen struct and call its helpers to avoid repeating
// layout and theming boilerplate.
type ScreenBase struct {
	Theme   theme.Theme
	IsDark  bool
	Width   int
	Height  int
	Keys    appkeys.GlobalKeyMap
	Help    help.Model
	AppName string // application name shown in every screen's header badge
}

// NewBase initialises a ScreenBase for the given terminal background.
func NewBase(isDark bool, appName string) ScreenBase {
	b := ScreenBase{Keys: appkeys.New(), Help: help.New(), AppName: appName}
	b.ApplyTheme(isDark)
	return b
}

// ApplyTheme regenerates theme and help styles for the given background.
// Call at the start of SetTheme() before any component-specific updates.
func (b *ScreenBase) ApplyTheme(isDark bool) {
	b.IsDark = isDark
	b.Theme = theme.New(isDark)
	b.Help.Styles = help.DefaultStyles(isDark)
}

// ContentWidth returns the inner width after the App container's horizontal frame.
func (b *ScreenBase) ContentWidth() int {
	frameH, _ := b.Theme.App.GetFrameSize()
	return b.Width - frameH
}

// IsSized returns true once the screen has received a non-zero WindowSizeMsg.
func (b *ScreenBase) IsSized() bool {
	return b.Width > 0 && b.Height > 0
}

// HeaderView renders the app name badge followed by a horizontal rule that
// fills the remaining content width. Visible on every screen.
// The bottom margin creates consistent spacing between the header and the
// screen content below; all sizing calculations use lipgloss.Height() on
// this output so the space is automatically accounted for.
func (b *ScreenBase) HeaderView() string {
	t := b.Theme.Title.Padding(1, 2).Render(b.AppName)
	lineW := max(0, b.ContentWidth()-lipgloss.Width(t))
	line := b.Theme.Subtle.Render(strings.Repeat("─", lineW))
	return lipgloss.NewStyle().MarginBottom(1).Render(
		lipgloss.JoinHorizontal(lipgloss.Center, t, line),
	)
}

// RenderHelp renders the help bar from any help.KeyMap, with a top margin.
func (b *ScreenBase) RenderHelp(km help.KeyMap) string {
	return lipgloss.NewStyle().MarginTop(1).Render(b.Help.View(km))
}

// CalculateContentHeight returns the appropriate content height based on:
// 1. Available space after subtracting header and help heights
// 2. MaxContentHeight constant (or 1/3 of terminal height, whichever is smaller)
// 3. MinContentHeight minimum
//
// The headerHeight and helpHeight parameters should be calculated using
// lipgloss.Height() on the actual rendered header and help views.
func (b *ScreenBase) CalculateContentHeight(headerHeight, helpHeight int) int {
	if !b.IsSized() {
		return MinContentHeight
	}

	_, frameV := b.Theme.App.GetFrameSize()
	availableH := b.Height - frameV - headerHeight - helpHeight

	// Cap at MaxContentHeight or 1/3 of terminal height, whichever is smaller
	maxH := MaxContentHeight
	if thirdHeight := b.Height / 3; thirdHeight < maxH {
		maxH = thirdHeight
	}

	// Use the smaller of available space or max height
	contentH := availableH
	if contentH > maxH {
		contentH = maxH
	}

	// Enforce minimum
	if contentH < MinContentHeight {
		contentH = MinContentHeight
	}

	return contentH
}

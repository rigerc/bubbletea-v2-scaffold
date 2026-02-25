// Package screens provides the individual screen implementations for the application.
package screens

// NewDetailsExampleScreen creates a DetailScreen pre-loaded with example
// scrollable content. All copy for this view lives here so that model.go
// stays free of view-specific strings.
func NewDetailsExampleScreen(isDark bool, appName string) *DetailScreen {
	return NewDetailScreen("Details Example", detailsExampleContent, isDark, appName)
}

// detailsExampleContent is the scrollable demo text shown in the Details
// Example screen. Edit this constant to customise what the screen displays.
const detailsExampleContent = `Details Example

This screen demonstrates the scrollable viewport component.

Scroll controls:
  • j / ↓          — line down
  • k / ↑          — line up
  • d / page down   — half page down
  • u / page up     — half page up
  • g / home        — top
  • G / end         — bottom
  • mouse wheel     — scroll

Press ESC to return to the menu.

─────────────────────────────────────

Section 1 — Lorem Ipsum

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod
tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,
quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo
consequat.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident,
sunt in culpa qui officia deserunt mollit anim id est laborum.

─────────────────────────────────────

Section 2 — More Content

Curabitur pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius,
turpis molestie pretium placerat, arcu ante tincidunt purus, vel bibendum
nisi nunc a lectus.

Pellentesque habitant morbi tristique senectus et netus et malesuada fames
ac turpis egestas. Vestibulum tortor quam, feugiat vitae, ultricies eget,
tempor sit amet, ante. Donec eu libero sit amet quam egestas semper. Aenean
ultricies mi vitae est.

─────────────────────────────────────

Section 3 — Even More

Mauris placerat eleifend leo. Quisque sit amet est et sapien ullamcorper
pharetra. Vestibulum erat wisi, condimentum sed, commodo vitae, ornare sit
amet, wisi.

Aenean fermentum, elit eget tincidunt condimentum, eros ipsum rutrum orci,
sagittis tempus lacus enim ac dui. Donec non enim in turpis pulvinar
facilisis. Ut felis. Praesent dapibus, neque id cursus faucibus, tortor
neque egestas augue, eu vulputate magna eros eu erat.

─────────────────────────────────────

End of content.`

// Package spinner provides a thin, theme-aware wrapper around bubbles/spinner.
// Screens that run background tasks embed this model and delegate Update/View.
package spinner

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/theme"
)

// Model wraps bubbles spinner with theme-aware styling.
type Model struct {
	s spinner.Model
}

// New creates a spinner styled with the given palette's secondary colour.
func New(p theme.Palette) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(p.Secondary)
	return Model{s: s}
}

// Init returns the command that starts the tick loop.
func (m Model) Init() tea.Cmd {
	return m.s.Tick
}

// Update forwards messages to the inner spinner and returns updated state.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.s, cmd = m.s.Update(msg)
	return m, cmd
}

// View renders the current spinner frame.
func (m Model) View() string {
	return m.s.View()
}

// Loading combines a spinner with an active flag. Embed it in any screen that
// runs background tasks so all five concerns (field storage, palette updates,
// init, update forwarding, and view rendering) live in one place.
type Loading struct {
	spin   Model
	active bool
}

// NewLoading creates a Loading initialised with the given palette.
func NewLoading(p theme.Palette) Loading {
	return Loading{spin: New(p)}
}

// Start marks the loading as active and returns the spinner tick command.
// Call this inside the screen's Init (or alongside the task command).
func (l *Loading) Start() tea.Cmd {
	l.active = true
	return l.spin.Init()
}

// Stop marks the loading as inactive.
func (l *Loading) Stop() {
	l.active = false
}

// Active reports whether loading is in progress.
func (l *Loading) Active() bool {
	return l.active
}

// ApplyPalette re-creates the inner spinner with a new palette.
// Call this from the screen's ApplyTheme method.
func (l *Loading) ApplyPalette(p theme.Palette) {
	l.spin = New(p)
}

// Update forwards the message to the spinner and returns the updated Loading.
// Call this from the screen's Update method while Active() is true.
func (l Loading) Update(msg tea.Msg) (Loading, tea.Cmd) {
	var cmd tea.Cmd
	l.spin, cmd = l.spin.Update(msg)
	return l, cmd
}

// View renders the spinner dot and label as a single padded element.
// The dot uses the secondary colour (set on New); the label uses primary.
func (l Loading) View(label string, p theme.Palette) string {
	text := lipgloss.NewStyle().Foreground(p.Primary).Render(label)
	inner := l.spin.View() + text
	return lipgloss.NewStyle().Padding(2).Render(inner)
}

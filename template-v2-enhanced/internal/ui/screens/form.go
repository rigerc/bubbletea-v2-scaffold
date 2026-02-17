// Package screens provides the individual screen implementations for the application.
package screens

import (
	lipgloss "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	tea "charm.land/bubbletea/v2"

	huhadapter "template-v2-enhanced/internal/ui/huh"
	"template-v2-enhanced/internal/ui/nav"
)

// FormScreen wraps a huh.Form to implement nav.Screen.
// It provides integration between Huh forms and the app's navigation system,
// handling global keys (ESC, Ctrl+C, ?) alongside form-specific keys.
type FormScreen struct {
	ScreenBase
	form           *huh.Form
	onSubmit       func() tea.Cmd
	onAbort        func() tea.Cmd
	formBuilder    func() *huh.Form // Function to rebuild the form when needed
	needsReset     bool            // Flag indicating form needs reset on next Update
}

// NewFormScreen creates a form screen with theme and callbacks.
// The isDark parameter determines the initial theme; it will be updated
// via SetTheme when the screen is pushed onto the stack.
//
// onSubmit is called when the form is completed successfully.
// onAbort is called when the form is aborted (ESC or form abort).
func NewFormScreen(
	form *huh.Form,
	isDark bool,
	appName string,
	onSubmit func() tea.Cmd,
	onAbort func() tea.Cmd,
) *FormScreen {
	fs := &FormScreen{
		ScreenBase: NewBase(isDark, appName),
		form:       form,
		onSubmit:   onSubmit,
		onAbort:    onAbort,
	}

	// Apply theme and keymap
	form.WithTheme(huhadapter.ThemeFunc(fs.Theme))
	form.WithKeyMap(huhadapter.KeyMap(fs.Keys))

	return fs
}

// newFormScreenWithBuilder is like NewFormScreen but takes a form builder
// function that can rebuild the form when it needs to be reset (e.g., after
// navigation returns to this screen).
func newFormScreenWithBuilder(
	formBuilder func() *huh.Form,
	isDark bool,
	appName string,
	onSubmit func() tea.Cmd,
	onAbort func() tea.Cmd,
) *FormScreen {
	form := formBuilder()
	fs := &FormScreen{
		ScreenBase:  NewBase(isDark, appName),
		form:        form,
		onSubmit:    onSubmit,
		onAbort:     onAbort,
		formBuilder: formBuilder,
	}

	// Apply theme and keymap
	form.WithTheme(huhadapter.ThemeFunc(fs.Theme))
	form.WithKeyMap(huhadapter.KeyMap(fs.Keys))

	return fs
}

// Init returns the form's initial command.
func (s *FormScreen) Init() tea.Cmd {
	return s.form.Init()
}

// Update handles incoming messages and returns an updated screen and command.
// Global keys take precedence over form-specific keys:
//   - ? toggles help expansion
//   - ESC triggers abort callback (navigates back)
//   - Ctrl+C quits the application
func (s *FormScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case resetFormMsg:
		return s.handleResetMsg()

	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height

	case tea.KeyPressMsg:
		// Global keys take precedence
		switch {
		case key.Matches(msg, s.Keys.Help):
			s.Help.ShowAll = !s.Help.ShowAll
			return s, nil
		case key.Matches(msg, s.Keys.Back):
			return s, s.onAbort()
		case key.Matches(msg, s.Keys.Quit):
			return s, tea.Quit
		}
	}

	// Delegate to form
	form, cmd := s.form.Update(msg)
	s.form = form.(*huh.Form)

	// Check form state and trigger callbacks
	// After triggering callbacks, reset form state so it remains usable
	// when we navigate back to this screen (e.g., after ESC from a pushed screen)
	switch s.form.State {
	case huh.StateCompleted:
		cmd = tea.Batch(cmd, s.onSubmit(), s.resetFormCmd())
	case huh.StateAborted:
		cmd = tea.Batch(cmd, s.onAbort(), s.resetFormCmd())
	}

	return s, cmd
}

// View renders the form screen with header and help bar.
func (s *FormScreen) View() string {
	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			s.form.View(),
		),
	)
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *FormScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
	s.form.WithTheme(huhadapter.ThemeFunc(s.Theme))
}

// resetFormCmd returns a command that resets the form to its initial state.
// This ensures the form remains usable after navigation returns to this screen.
func (s *FormScreen) resetFormCmd() tea.Cmd {
	return func() tea.Msg {
		return resetFormMsg{}
	}
}

// resetFormMsg is a message that triggers a form reset.
type resetFormMsg struct{}

// handleResetMsg rebuilds the form if a builder is available, or re-inits it.
func (s *FormScreen) handleResetMsg() (nav.Screen, tea.Cmd) {
	if s.formBuilder != nil {
		// Rebuild the form entirely
		s.form = s.formBuilder()
		s.form.WithTheme(huhadapter.ThemeFunc(s.Theme))
		s.form.WithKeyMap(huhadapter.KeyMap(s.Keys))
	}
	// Re-initialize the form
	return s, s.form.Init()
}

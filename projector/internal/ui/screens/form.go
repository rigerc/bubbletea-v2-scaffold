// Package screens provides the individual screen implementations for the application.
package screens

import (
	lipgloss "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	tea "charm.land/bubbletea/v2"

	huhadapter "projector/internal/ui/huh"
	"projector/internal/ui/nav"
	"projector/internal/ui/theme"
)

// FormScreen wraps a huh.Form to implement nav.Screen.
// It provides integration between Huh forms and the app's navigation system,
// handling global keys (ESC, Ctrl+C, ?) alongside form-specific keys.
type FormScreen struct {
	ScreenBase
	form              *huh.Form
	onSubmit          func() tea.Cmd
	onAbort           func() tea.Cmd
	formBuilder       func() *huh.Form // Function to rebuild the form when needed
	needsReset        bool            // Flag indicating form needs reset on next Update
	maxContentHeight  int             // Optional override for max content height
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
	form.WithTheme(theme.HuhThemeFunc())
	form.WithKeyMap(huhadapter.KeyMap(fs.Keys))

	return fs
}

// newFormScreenWithBuilder is like NewFormScreen but takes a form builder
// function that can rebuild the form when it needs to be reset (e.g., after
// navigation returns to this screen).
//
// The maxContentHeight parameter allows overriding the default MaxContentHeight
// for forms that need more space (e.g., menus). Pass 0 to use the default.
func newFormScreenWithBuilder(
	formBuilder func() *huh.Form,
	isDark bool,
	appName string,
	onSubmit func() tea.Cmd,
	onAbort func() tea.Cmd,
	maxContentHeight int,
) *FormScreen {
	form := formBuilder()
	fs := &FormScreen{
		ScreenBase:     NewBase(isDark, appName),
		form:           form,
		onSubmit:       onSubmit,
		onAbort:        onAbort,
		formBuilder:    formBuilder,
		maxContentHeight: maxContentHeight,
	}

	// Apply theme and keymap
	form.WithTheme(theme.HuhThemeFunc())
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

// View renders the form screen with header.
// The form view is wrapped with a height constraint to ensure consistent sizing.
// Note: Huh forms render their own help internally, so we don't call RenderHelp.
func (s *FormScreen) View() string {
	// Calculate the max height for the form content
	headerH := lipgloss.Height(s.HeaderView())

	// Estimate form internal height (usually includes help at the bottom)
	formInternalHelpH := 4 // Approximate height for form's internal help

	// Use custom max height if set, otherwise use shared helper
	maxFormH := s.CalculateContentHeight(headerH, formInternalHelpH)
	if s.maxContentHeight > 0 {
		// Apply custom max height override
		_, frameV := s.Theme.App.GetFrameSize()
		availableH := s.Height - frameV - headerH - formInternalHelpH
		customMax := s.maxContentHeight
		if customMax > availableH {
			customMax = availableH
		}
		if customMax < MinContentHeight {
			customMax = MinContentHeight
		}
		maxFormH = customMax
	}

	// Wrap the form view with height constraint
	formView := lipgloss.NewStyle().
		Height(maxFormH).
		MaxHeight(maxFormH).
		Render(s.form.View())

	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			formView,
		),
	)
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *FormScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
	s.form.WithTheme(theme.HuhThemeFunc())
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
		s.form.WithTheme(theme.HuhThemeFunc())
		s.form.WithKeyMap(huhadapter.KeyMap(s.Keys))
	}
	// Re-initialize the form
	return s, s.form.Init()
}

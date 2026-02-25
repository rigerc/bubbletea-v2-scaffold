// Package screens provides the individual screen implementations for the application.
package screens

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"

	"scaffold/internal/ui/nav"
)

// homeOption pairs a menu label with the navigation command it triggers.
type homeOption struct {
	title  string
	action tea.Cmd
}

// HomeScreen is the root screen of the application. It displays a greeting
// and a Huh-powered Select menu linking to the main feature screens.
//
// It implements nav.Screen and nav.Themeable.
type HomeScreen struct {
	*FormScreen
	options     []homeOption
	selectedIdx *int
}

// NewHomeScreen constructs the root HomeScreen.
//
// appName is used for the ScreenBase (help bar, key bindings, etc.).
// isDark is the initial theme hint; the router will call SetTheme with the
// correct value once the terminal background colour is detected.
func NewHomeScreen(appName string, isDark bool) *HomeScreen {
	options := []homeOption{
		{
			title:  "Details example",
			action: nav.Push(NewDetailsExampleScreen(isDark, appName)),
		},
		{
			title:  "Settings",
			action: nav.Push(NewSettingsScreen(isDark, appName)),
		},
	}

	selectedIdx := new(int)

	formBuilder := func() *huh.Form {
		huhOptions := make([]huh.Option[int], len(options))
		for i, opt := range options {
			huhOptions[i] = huh.NewOption(opt.title, i)
		}
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Options(huhOptions...).
					Value(selectedIdx).
					Height(len(options)+2),
			),
		).WithShowHelp(true).WithShowErrors(true)
	}

	onSubmit := func() tea.Cmd {
		if *selectedIdx >= 0 && *selectedIdx < len(options) {
			return options[*selectedIdx].action
		}
		return nil
	}

	// ESC on the root menu quits the application.
	onAbort := func() tea.Cmd {
		return tea.Quit
	}

	fs := newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)

	return &HomeScreen{
		FormScreen:  fs,
		options:     options,
		selectedIdx: selectedIdx,
	}
}

// greetingView renders the "Hello there!" header that sits above the menu.
func (s *HomeScreen) greetingView() string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(s.Theme.Palette.Primary).
		MarginBottom(1).
		Render("Hello there!")
}

// View renders the home screen:
//
//	Hello there!
//
//	  > Details example
//	    Settings
//
//	  esc back  ctrl+c/q quit
func (s *HomeScreen) View() string {
	greeting := s.greetingView()
	helpView := s.RenderHelp(s.Keys)

	// Reserve space for Huh's built-in help bar so the form is never clipped.
	const formInternalHelpH = 4

	bodyH := s.Layout().
		Header(greeting).
		Help(helpView).
		BodyHeight()

	maxFormH := max(MinContentHeight, bodyH-formInternalHelpH)

	formView := lipgloss.NewStyle().
		Height(maxFormH).
		MaxHeight(maxFormH).
		Render(s.FormScreen.form.View())

	return s.Layout().
		Header(greeting).
		Body(formView).
		Help(helpView).
		Render()
}

// Update delegates to FormScreen and keeps s.FormScreen in sync.
func (s *HomeScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	screen, cmd := s.FormScreen.Update(msg)
	if fs, ok := screen.(*FormScreen); ok {
		s.FormScreen = fs
	}
	return s, cmd
}

// SetTheme propagates the theme change to the embedded FormScreen.
// Implements nav.Themeable.
func (s *HomeScreen) SetTheme(isDark bool) {
	s.FormScreen.SetTheme(isDark)
}

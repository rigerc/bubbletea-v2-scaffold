// Package screens provides the individual screen implementations for the application.
package screens

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lipgloss "charm.land/lipgloss/v2"

	"projector/internal/ui/banner"
	"projector/internal/ui/nav"
)

// HuhMenuOption represents a menu item with its navigation action for Huh-based menus.
type HuhMenuOption struct {
	Title       string
	Description string
	Action      tea.Cmd
}

// HuhMenuScreen uses Huh's Select field for navigation.
// It implements nav.Screen and nav.Themeable.
type HuhMenuScreen struct {
	*FormScreen
	options     []HuhMenuOption
	selectedIdx *int // Pointer to the form's bound value

	bannerCfg   banner.BannerConfig // fixed for the lifetime of this screen
	bannerCache string              // last rendered banner; re-rendered on width change
	bannerWidth int
}

// NewHuhMenuScreen creates a menu using Huh's Select field.
// The isDark parameter should be false initially; the correct value will be
// set via SetTheme when the screen is pushed onto the stack.
func NewHuhMenuScreen(options []HuhMenuOption, isDark bool, appName string) *HuhMenuScreen {
	// Capture the menu options and selected index
	menuOpts := options
	selectedIdx := new(int)

	// formBuilder is a function that can rebuild the form when needed
	formBuilder := func() *huh.Form {
		// Build Huh options from menu options
		huhOptions := make([]huh.Option[int], len(menuOpts))
		for i, opt := range menuOpts {
			huhOptions[i] = huh.NewOption(opt.Title, i)
		}

		// Create form with Select field
		// Set explicit height to show more menu items
		return huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[int]().
					Options(huhOptions...).
					Value(selectedIdx).
					Height(10),
			),
		).WithShowHelp(true).WithShowErrors(true)
	}

	onSubmit := func() tea.Cmd {
		// Use the current selectedIdx value
		if *selectedIdx >= 0 && *selectedIdx < len(menuOpts) {
			if menuOpts[*selectedIdx].Action != nil {
				return menuOpts[*selectedIdx].Action
			}
		}
		return nil
	}

	onAbort := func() tea.Cmd {
		return tea.Quit // ESC quits from main menu
	}

	fs := newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)

	return &HuhMenuScreen{
		FormScreen:  fs,
		options:     menuOpts,
		selectedIdx: selectedIdx,
		bannerCfg: banner.BannerConfig{
			Text:  appName,
			Font:  "smslant",
			Width: 90,
			Background: false,
		},
	}
}

// View renders the menu with a figlet ASCII banner above the form.
// On terminals shorter than 20 rows it falls back to the plain text header.
// The banner is cached and only re-rendered when the content width changes.
func (s *HuhMenuScreen) View() string {
	contentWidth := s.ContentWidth()

	if s.Height <= 20 {
		s.bannerCache = s.HeaderView()
		s.bannerWidth = contentWidth
	} else if s.bannerCache == "" || s.bannerWidth != contentWidth {
		rendered, err := banner.RenderBanner(s.bannerCfg, contentWidth)
		if err != nil {
			rendered = s.HeaderView()
		}
		s.bannerCache = rendered
		s.bannerWidth = contentWidth
	}

	bannerH := lipgloss.Height(s.bannerCache)
	formInternalHelpH := 4
	maxFormH := s.CalculateContentHeight(bannerH, formInternalHelpH)

	formView := lipgloss.NewStyle().
		Height(maxFormH).
		MaxHeight(maxFormH).
		Render(s.FormScreen.form.View())

	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.bannerCache,
			formView,
		),
	)
}

// Update handles incoming messages and returns an updated screen and command.
func (s *HuhMenuScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	// Delegate to FormScreen which handles the form update
	screen, cmd := s.FormScreen.Update(msg)

	// Update our reference if the FormScreen changed
	if fs, ok := screen.(*FormScreen); ok {
		s.FormScreen = fs
	}

	return s, cmd
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *HuhMenuScreen) SetTheme(isDark bool) {
	s.FormScreen.SetTheme(isDark)
}

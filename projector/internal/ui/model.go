// Package ui provides the BubbleTea UI model for the application.
// It implements a stack-based navigation router with theme support.
package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"projector/config"
	applogger "projector/internal/logger"
	"projector/internal/ui/nav"
	"projector/internal/ui/screens"
)

// Model represents the application state with a navigation stack.
type Model struct {
	// screens holds the navigation stack. The last element is the active screen.
	screens []nav.Screen

	// width and height store the current terminal dimensions.
	width, height int

	// isDark indicates if the terminal has a dark background.
	isDark bool

	// quitting is set to true when the app is about to exit.
	quitting bool

	// Config-derived fields (extracted from config.Config at construction).
	altScreen    bool
	mouseEnabled bool
	windowTitle  string
}

// New creates a new Model with the provided configuration.
// It accepts config.Config as a value type (main.go passes *cfg dereferenced).
func New(cfg config.Config) Model {
	// Define sample content for detail screens.
	detailContent := `This is a detail screen with scrollable content.

Scroll controls:
  • j / ↓        — line down
  • k / ↑        — line up
  • d / page down — half page down
  • u / page up   — half page up
  • g / home      — top
  • G / end       — bottom
  • mouse wheel   — scroll

Press ESC to return to the menu.

─────────────────────────────────────

Section 1 — Lorem Ipsum

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod
tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam,
quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore
eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.

─────────────────────────────────────

Section 2 — More Filler

Sunt in culpa qui officia deserunt mollit anim id est laborum. Curabitur
pretium tincidunt lacus. Nulla gravida orci a odio. Nullam varius, turpis
molestie pretium placerat, arcu ante tincidunt purus, vel bibendum nisi.

Pellentesque habitant morbi tristique senectus et netus et malesuada fames
ac turpis egestas. Vestibulum tortor quam, feugiat vitae, ultricies eget,
tempor sit amet, ante. Donec eu libero sit amet quam egestas semper.

─────────────────────────────────────

Section 3 — Even More

Aenean ultricies mi vitae est. Mauris placerat eleifend leo. Quisque sit
amet est et sapien ullamcorper pharetra. Vestibulum erat wisi, condimentum
sed, commodo vitae, ornare sit amet, wisi. Aenean fermentum, elit eget
tincidunt condimentum, eros ipsum rutrum orci.

Nullam venenatis felis eu purus vestibulum, nec malesuada nisl iaculis.
Fusce aliquet purus vel mauris pharetra, a condimentum lectus tincidunt.

─────────────────────────────────────

End of content.`

	aboutContent := `scaffold

A BubbleTea v2 skeleton application with:
  • Stack-based navigation
  • Adaptive light/dark theming
  • List-based menu navigation
  • Scrollable detail screens with capped height

Built with:
  • charm.land/bubbletea/v2
  • charm.land/bubbles/v2
  • charm.land/lipgloss/v2
  • github.com/spf13/cobra
  • github.com/knadh/koanf/v2
  • github.com/rs/zerolog

─────────────────────────────────────

Architecture

The application uses a stack-based navigator (internal/ui/nav). Each screen
implements the nav.Screen interface (Init / Update / View). The root model
holds the stack and fans messages out to the active screen.

Theme detection uses tea.RequestBackgroundColor, which fires a
tea.BackgroundColorMsg carrying the terminal's actual background colour.
Screens implement nav.Themeable to receive isDark updates.

Config is loaded via koanf: defaults → config file → env vars → flags.
Logging uses zerolog with a file sink so it doesn't interfere with the TUI.

─────────────────────────────────────

Press ESC to return to the menu.`

	// Create menu items using Huh-based menu.
	menuOptions := []screens.HuhMenuOption{
		{
			Title:       "Details",
			Description: "View a detail screen",
			Action:      nav.Push(screens.NewDetailScreen("Details", detailContent, false, cfg.App.Name)),
		},
		{
			Title:       "Browse Files",
			Description: "Browse the filesystem",
			Action:      nav.Push(screens.NewHuhFilePickerScreen(".", false, cfg.App.Name)),
		},
		{
			Title:       "Settings",
			Description: "Configure application",
			Action:      nav.Push(screens.NewSettingsScreen(false, cfg.App.Name)),
		},
		{
			Title:       "Banner Demo",
			Description: "Showcase ASCII fonts and gradients",
			Action:      nav.Push(screens.NewBannerDemoScreen(false, cfg.App.Name)),
		},
		{
			Title:       "About",
			Description: "About this application",
			Action:      nav.Push(screens.NewDetailScreen("About", aboutContent, false, cfg.App.Name)),
		},
	}

	root := screens.NewHuhMenuScreen(menuOptions, false, cfg.App.Name)

	return Model{
		screens:      []nav.Screen{root},
		altScreen:    cfg.UI.AltScreen,
		mouseEnabled: cfg.UI.MouseEnabled,
		windowTitle:  cfg.App.Title,
	}
}

// Init returns the initial command. It requests the terminal background color
// and initializes the root screen.
func (m Model) Init() tea.Cmd {
	applogger.Debug().Msg("Initializing UI model")
	cmds := []tea.Cmd{tea.RequestBackgroundColor}
	if len(m.screens) > 0 {
		cmds = append(cmds, m.screens[len(m.screens)-1].Init())
	}
	return tea.Batch(cmds...)
}

// Update handles incoming messages and returns an updated model and command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			applogger.Debug().Msg("Quit key pressed")
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		applogger.Debug().Msgf("Window resized: %dx%d", m.width, m.height)
		// fall through to delegate to active screen

	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()
		applogger.Debug().Msgf("Background color detected: isDark=%v", m.isDark)
		// Propagate theme to ALL screens in stack
		for i := range m.screens {
			if t, ok := m.screens[i].(nav.Themeable); ok {
				t.SetTheme(m.isDark)
			}
		}
		// fall through to deliver msg to active screen

	case nav.PushMsg:
		s := msg.Screen
		if cmd := s.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if t, ok := s.(nav.Themeable); ok {
			t.SetTheme(m.isDark)
		}
		s, cmd := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		cmds = append(cmds, cmd)
		m.screens = append(m.screens, s)
		return m, tea.Batch(cmds...)

	case nav.PopMsg:
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
			// Refresh the newly-exposed screen with current window size
			top := m.screens[len(m.screens)-1]
			updated, cmd := top.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m.screens[len(m.screens)-1] = updated
			return m, cmd
		}
		return m, nil

	case nav.ReplaceMsg:
		if len(m.screens) > 0 {
			s := msg.Screen
			if cmd := s.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
			if t, ok := s.(nav.Themeable); ok {
				t.SetTheme(m.isDark)
			}
			s, cmd := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			cmds = append(cmds, cmd)
			m.screens[len(m.screens)-1] = s
		}
		return m, tea.Batch(cmds...)

	case screens.SettingsAppliedMsg:
		// Settings were applied - log them and optionally update app config
		applogger.Debug().Msgf("Settings applied: %+v", msg.Data)
		// Pop the settings screen after successful submission
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
		}
		return m, nil
	}

	// Delegate to active screen
	if len(m.screens) > 0 {
		top := m.screens[len(m.screens)-1]
		updated, cmd := top.Update(msg)
		m.screens[len(m.screens)-1] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current model state as a tea.View.
func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	var content string
	if len(m.screens) > 0 {
		content = m.screens[len(m.screens)-1].View()
	}

	v := tea.NewView(content)
	v.AltScreen = m.altScreen   // from cfg.UI.AltScreen
	v.WindowTitle = m.windowTitle // from cfg.App.Title
	if m.mouseEnabled {          // from cfg.UI.MouseEnabled
		v.MouseMode = tea.MouseModeCellMotion
	}
	return v
}

// Run starts the BubbleTea program with the given model.
func Run(m Model) error {
	applogger.Info().Msg("Starting BubbleTea program")

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running program: %w", err)
	}

	applogger.Info().Msg("Program exited successfully")
	return nil
}

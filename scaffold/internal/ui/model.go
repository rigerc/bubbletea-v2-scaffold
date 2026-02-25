// Package ui provides the BubbleTea UI model for the application.
// It implements a stack-based navigation router with a persistent banner and
// theme support.
package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"scaffold/config"
	applogger "scaffold/internal/logger"
	"scaffold/internal/ui/banner"
	"scaffold/internal/ui/nav"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/theme"
)

// borderOverhead is the number of columns/rows consumed by the persistent
// outer border (1 char on each side → 2 per axis).
const borderOverhead = 2

// Model represents the application state with a navigation stack.
type Model struct {
	// screens holds the navigation stack. The last element is the active screen.
	screens []nav.Screen

	// width and height store the *inner* content dimensions (terminal size minus
	// borderOverhead on each axis), so every screen sizes itself to fit inside
	// the persistent outer border rendered by View.
	width, height int

	// bannerStr is the cached rendered ASCII art banner.
	// It is re-rendered whenever the terminal width changes.
	bannerStr string

	// bannerHeight is the row count of bannerStr. It is subtracted from the
	// height delivered to child screens via WindowSizeMsg so they know exactly
	// how much vertical space is available below the banner.
	bannerHeight int

	// bannerWidth tracks the content width at which bannerStr was last rendered
	// so we can invalidate the cache on resize.
	bannerWidth int

	// isDark indicates if the terminal has a dark background.
	isDark bool

	// th is the current application theme, rebuilt whenever isDark changes.
	th theme.Theme

	// quitting is set to true when the app is about to exit.
	quitting bool

	// appName is the short application name rendered in the persistent banner.
	appName string

	// Config-derived fields (extracted from config.Config at construction).
	altScreen    bool
	mouseEnabled bool
	windowTitle  string
}

// New creates a new Model with the provided configuration.
// It accepts config.Config as a value type (main.go passes *cfg dereferenced).
func New(cfg config.Config) Model {
	return Model{
		screens:      []nav.Screen{screens.NewHomeScreen(cfg.App.Name, false)},
		th:           theme.New(false),
		appName:      cfg.App.Name,
		altScreen:    cfg.UI.AltScreen,
		mouseEnabled: cfg.UI.MouseEnabled,
		windowTitle:  cfg.App.Title,
	}
}

// renderBanner renders the persistent ASCII art banner at the model's current
// content width. It returns the rendered string and its height in rows.
// If the width is not yet known (≤ 0), it returns empty values.
func (m Model) renderBanner() (string, int) {
	if m.width <= 0 {
		return "", 0
	}
	cfg := banner.BannerConfig{
		Text:          m.appName,
		Font:          "smslant",
		Width:         m.width,
		Justification: 1, // centered
	}
	s, err := banner.RenderBanner(cfg, m.width)
	if err != nil {
		// Graceful fallback: plain text line.
		s = m.appName + "\n"
	}
	return s, lipgloss.Height(s)
}

// screenHeight returns the content height available to the active screen after
// reserving space for the persistent banner.
func (m Model) screenHeight() int {
	return max(0, m.height-m.bannerHeight)
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

	// Shrink WindowSizeMsg by the border overhead so all screens and the stored
	// m.width / m.height refer to the inner content area (inside the border).
	// We also re-render and cache the banner here so screenHeight() is accurate
	// for the downstream message that screens receive.
	if wm, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = max(0, wm.Width-borderOverhead)
		m.height = max(0, wm.Height-borderOverhead)

		// Re-render banner only when the width actually changed.
		if m.width != m.bannerWidth {
			m.bannerStr, m.bannerHeight = m.renderBanner()
			m.bannerWidth = m.width
		}

		applogger.Debug().Msgf(
			"Window resized: terminal=%dx%d inner=%dx%d banner=%d rows screen=%dx%d",
			wm.Width, wm.Height,
			m.width, m.height,
			m.bannerHeight,
			m.width, m.screenHeight(),
		)

		// Replace the message with the screen-adjusted dimensions so that
		// every downstream handler (switch cases and active-screen delegation)
		// receives the correct available area.
		msg = tea.WindowSizeMsg{Width: m.width, Height: m.screenHeight()}
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			applogger.Debug().Msg("Quit key pressed")
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		// m.width and m.height are already updated in the block above.
		// This case is kept so the message falls through to the active-screen
		// delegation at the bottom of the function.
		_ = msg

	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()
		m.th = theme.New(m.isDark)
		applogger.Debug().Msgf("Background color detected: isDark=%v", m.isDark)
		// Propagate theme to ALL screens in the stack.
		for i := range m.screens {
			if t, ok := m.screens[i].(nav.Themeable); ok {
				t.SetTheme(m.isDark)
			}
		}
		// fall through to deliver msg to the active screen

	case nav.PushMsg:
		s := msg.Screen
		if cmd := s.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if t, ok := s.(nav.Themeable); ok {
			t.SetTheme(m.isDark)
		}
		s, cmd := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.screenHeight()})
		cmds = append(cmds, cmd)
		m.screens = append(m.screens, s)
		return m, tea.Batch(cmds...)

	case nav.PopMsg:
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
			// Refresh the newly-exposed screen with the current window size.
			top := m.screens[len(m.screens)-1]
			updated, cmd := top.Update(tea.WindowSizeMsg{Width: m.width, Height: m.screenHeight()})
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
			s, cmd := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.screenHeight()})
			cmds = append(cmds, cmd)
			m.screens[len(m.screens)-1] = s
		}
		return m, tea.Batch(cmds...)

	case screens.SettingsAppliedMsg:
		applogger.Debug().Msgf("Settings applied: %+v", msg.Data)
		// Pop the settings screen after a successful submission.
		if len(m.screens) > 1 {
			m.screens = m.screens[:len(m.screens)-1]
		}
		return m, nil
	}

	// Delegate to the active screen.
	if len(m.screens) > 0 {
		top := m.screens[len(m.screens)-1]
		updated, cmd := top.Update(msg)
		m.screens[len(m.screens)-1] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current model state as a tea.View.
//
// Layout (inside the outer rounded border):
//
//	┌───────────────────────────────────┐
//	│  ███████╗  ██████╗  ██████╗ ...  │  ← persistent ASCII art banner
//	│                                   │
//	│  [active screen content]          │  ← screen fills the remaining rows
//	└───────────────────────────────────┘
func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	// Render the active screen.
	var screenContent string
	if len(m.screens) > 0 {
		screenContent = m.screens[len(m.screens)-1].View()
	}

	// Stack banner above screen content. If the banner is not yet available
	// (width unknown on the very first frame), fall back to screen-only output.
	var content string
	if m.bannerStr != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, m.bannerStr, screenContent)
	} else {
		content = screenContent
	}

	// Wrap everything in the persistent rounded border whose colour tracks the
	// current theme. In alt-screen mode the border also fills the full height.
	var rendered string
	if m.width > 0 {
		bs := m.th.AppBorder.Width(m.width)
		if m.altScreen && m.height > 0 {
			bs = bs.Height(m.height)
		}
		rendered = bs.Render(content)
	} else {
		rendered = content
	}

	v := tea.NewView(rendered)
	v.AltScreen = m.altScreen
	v.WindowTitle = m.windowTitle
	if m.mouseEnabled {
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

package ui

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/ui/banner"
	"scaffold/internal/ui/keys"
	"scaffold/internal/ui/menu"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/theme"
)

// NavigateMsg is a message to navigate to a new screen.
type NavigateMsg struct {
	Screen screens.Screen
}

// statusMsg is sent to update the footer status text.
type statusMsg struct{ text string }

// clearStatusMsg is sent after a delay to reset the footer.
type clearStatusMsg struct{}

// setStatus returns a Cmd that sets a timed status message.
func setStatus(text string, duration time.Duration) tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return statusMsg{text: text} },
		tea.Tick(duration, func(time.Time) tea.Msg { return clearStatusMsg{} }),
	)
}

// screenStack holds the navigation history.
type screenStack struct {
	screens []screens.Screen
}

// Push adds a screen to the stack.
func (s *screenStack) Push(screen screens.Screen) {
	s.screens = append(s.screens, screen)
}

// Pop removes and returns the top screen.
func (s *screenStack) Pop() screens.Screen {
	if len(s.screens) == 0 {
		return nil
	}
	idx := len(s.screens) - 1
	screen := s.screens[idx]
	s.screens = s.screens[:idx]
	return screen
}

// Peek returns the top screen without removing it.
func (s *screenStack) Peek() screens.Screen {
	if len(s.screens) == 0 {
		return nil
	}
	return s.screens[len(s.screens)-1]
}

// Len returns the stack depth.
func (s *screenStack) Len() int {
	return len(s.screens)
}

// rootModel is the root tea.Model — owns routing, WindowSize, header/footer.
type rootModel struct {
	cfg        config.Config
	configPath string // empty = no persistent save
	status     string // footer status text
	width      int
	height     int
	banner     string
	themeMgr   *theme.Manager
	ready      bool
	styles     theme.Styles
	keys       keys.GlobalKeyMap
	help       help.Model
	current    screens.Screen
	stack      screenStack
}

// newRootModel creates a new root model.
func newRootModel(cfg config.Config, configPath string) rootModel {
	return rootModel{
		cfg:        cfg,
		configPath: configPath,
		status:     "Ready",
		themeMgr:   theme.GetManager(),
		current:    screens.NewHome(),
		keys:       keys.DefaultGlobalKeyMap(),
		help:       help.New(),
	}
}

// Init initializes the root model.
func (m rootModel) Init() tea.Cmd {
	return tea.Batch(
		tea.RequestBackgroundColor,
		m.themeMgr.Init(m.cfg.UI.ThemeName, false, m.width),
	)
}

// Update handles messages for the root model.
func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Propagate width and height to current screen
		if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
			m.current = setter.SetWidth(m.width)
		}
		if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
			m.current = setter.SetHeight(m.bodyHeight())
		}
		return m, m.themeMgr.SetWidth(m.width)

	case tea.BackgroundColorMsg:
		isDark := msg.IsDark()
		m.help.Styles = help.DefaultStyles(isDark)
		return m, m.themeMgr.SetDarkMode(isDark)

	case theme.ThemeChangedMsg:
		// Apply to self
		m.styles = theme.NewFromPalette(msg.State.Palette, msg.State.Width)
		m.help.SetWidth(m.styles.MaxWidth)

		// Render/re-render banner with current theme palette
		if m.cfg.UI.ShowBanner {
			m.renderBanner()
		}

		// Apply to current screen
		if t, ok := m.current.(theme.Themeable); ok {
			t.ApplyTheme(msg.State)
		}
		return m, nil

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m, tea.Quit
		}

	case NavigateMsg:
		// Push current screen to stack and navigate to new screen
		m.stack.Push(m.current)
		m.current = msg.Screen
		// Propagate width and height to new screen
		if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
			m.current = setter.SetWidth(m.width)
		}
		if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
			m.current = setter.SetHeight(m.bodyHeight())
		}
		// Apply current theme to new screen
		if t, ok := m.current.(theme.Themeable); ok {
			t.ApplyTheme(m.themeMgr.State())
		}
		return m, m.current.Init()

	case menu.SelectionMsg:
		switch msg.Item.ScreenID() {
		case "settings":
			return m.Update(NavigateMsg{Screen: screens.NewSettings(m.cfg)})
		default:
			detail := screens.NewDetail(
				msg.Item.Title(), msg.Item.Description(), msg.Item.ScreenID(),
			)
			return m.Update(NavigateMsg{Screen: detail})
		}

	case screens.SettingsSavedMsg:
		themeChanged := m.cfg.UI.ThemeName != msg.Cfg.UI.ThemeName
		m.cfg = msg.Cfg

		// Clear banner if disabled
		if !msg.Cfg.UI.ShowBanner {
			m.banner = ""
		}

		var saveCmd tea.Cmd
		if m.configPath != "" {
			if err := config.Save(&m.cfg, m.configPath); err != nil {
				saveCmd = setStatus("Save failed: "+err.Error(), 5*time.Second)
			} else {
				saveCmd = setStatus("Settings saved", 3*time.Second)
			}
		} else {
			saveCmd = setStatus("Settings applied (no config file)", 3*time.Second)
		}

		// Handle theme change via manager
		if themeChanged {
			if m.stack.Len() > 0 {
				m.current = m.stack.Pop()
			}
			return m, tea.Batch(
				saveCmd,
				m.themeMgr.SetThemeName(m.cfg.UI.ThemeName),
			)
		}

		if m.stack.Len() > 0 {
			m.current = m.stack.Pop()
		}
		return m, saveCmd

	case screens.BackMsg:
		if m.stack.Len() > 0 {
			m.current = m.stack.Pop()
		}
		return m, nil

	case statusMsg:
		m.status = msg.text
		return m, nil

	case clearStatusMsg:
		m.status = "Ready"
		return m, nil
	}

	// Delegate to current screen
	var cmd tea.Cmd
	updated, cmd := m.current.Update(msg)
	if s, ok := updated.(screens.Screen); ok {
		m.current = s
	}
	return m, cmd
}

// View renders the root model.
func (m rootModel) View() tea.View {
	if !m.ready {
		return tea.NewView("")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.headerView(),
		m.styles.Body.Render(m.current.Body()),
		m.helpView(),
		m.footerView(),
	)

	return tea.NewView(m.styles.App.Render(content))
}

// renderBanner renders the ASCII art banner at its natural width and caches the result.
// Using a large fixed width lets lipgloss.Width(m.banner) reflect the font's true width,
// which headerView uses to decide whether the terminal is wide enough to display it.
func (m *rootModel) renderBanner() {
	state := m.themeMgr.State()
	p := state.Palette
	if p.Primary == nil {
		p = theme.NewPalette(m.cfg.UI.ThemeName, state.IsDark)
	}
	b, err := banner.Render(banner.Config{
		Text:          m.cfg.App.Title,
		Font:          "larry3d",
		Width:         100,
		Justification: 0,
		Gradient:      banner.GradientThemed(p.Primary, p.Secondary),
	})
	if err != nil {
		b = m.cfg.App.Title
	}
	m.banner = b
}

// headerView renders the header with either the ASCII banner or a styled plain-text title.
// The ASCII banner is shown only when ShowBanner is enabled, the banner has been rendered,
// and the terminal is wide enough to display it. In all other cases — including when
// ShowBanner is disabled or the terminal is too narrow — the plain-text title is shown.
func (m rootModel) headerView() string {
	if m.cfg.UI.ShowBanner && m.banner != "" && m.width > 0 && m.width >= lipgloss.Width(m.banner) {
		return m.styles.Header.Render(m.banner)
	}
	return m.styles.Header.Render(m.plainTitleView())
}

// plainTitleView renders a styled plain-text title used when ShowBanner is off.
func (m rootModel) plainTitleView() string {
	return m.styles.PlainTitle.Render(m.cfg.App.Title)
}

// helpView renders the persistent help box showing global and screen-specific keybindings.
func (m rootModel) helpView() string {
	combined := m.combinedKeys()
	return m.styles.Help.Render(m.help.View(combined))
}

// combinedKeys returns a key map that combines global keys with screen-specific keys.
func (m rootModel) combinedKeys() combinedKeyMap {
	return combinedKeyMap{
		global: m.keys,
		screen: m.current,
	}
}

// combinedKeyMap combines global and screen-specific key bindings.
type combinedKeyMap struct {
	global keys.GlobalKeyMap
	screen screens.Screen
}

// ShortHelp returns combined short help bindings.
func (c combinedKeyMap) ShortHelp() []key.Binding {
	bindings := c.global.ShortHelp()
	if kb, ok := c.screen.(screens.KeyBinder); ok {
		bindings = append(bindings, kb.ShortHelp()...)
	}
	return bindings
}

// FullHelp returns combined full help bindings.
func (c combinedKeyMap) FullHelp() [][]key.Binding {
	groups := c.global.FullHelp()
	if kb, ok := c.screen.(screens.KeyBinder); ok {
		groups = append(groups, kb.FullHelp()...)
	}
	return groups
}

// bodyHeight estimates the available height for the body content area.
// It subtracts the header, help, and footer chrome from the terminal height.
func (m rootModel) bodyHeight() int {
	if m.height == 0 {
		return 0
	}
	header := lipgloss.Height(m.headerView())
	helpH := lipgloss.Height(m.helpView())
	footer := lipgloss.Height(m.footerView())
	body := m.height - header - helpH - footer
	if body < 1 {
		body = 1
	}
	return body
}

// footerView renders the status bar footer.
func (m rootModel) footerView() string {
	left := m.styles.StatusLeft.Render(" " + m.status + " ")
	rightContent := " v" + m.cfg.App.Version
	if m.cfg.Debug {
		rightContent += " [DEBUG]"
	}
	right := m.styles.StatusRight.Render(rightContent + " ")

	// Account for footer border (2) and padding (1)
	innerWidth := m.styles.MaxWidth - 3

	gap := lipgloss.NewStyle().
		Width(innerWidth - lipgloss.Width(left) - lipgloss.Width(right)).
		Render("")
	footerContent := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
	return m.styles.Footer.Render(footerContent)
}

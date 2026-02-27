package ui

import (
	"context"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/task"
	"scaffold/internal/ui/banner"
	"scaffold/internal/ui/keys"
	"scaffold/internal/ui/menu"
	"scaffold/internal/ui/modal"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/status"
	"scaffold/internal/ui/theme"
)

// NavigateMsg is a message to navigate to a new screen.
type NavigateMsg struct {
	Screen screens.Screen
}

// rootState represents the loading state of the root model.
type rootState int

const (
	rootStateLoading rootState = iota // waiting for first WindowSizeMsg
	rootStateReady                    // terminal dimensions known, UI renderable
	rootStateError                    // unrecoverable startup error
)

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
	ctx          context.Context
	cancel       context.CancelFunc // shutdown only; cancels all running tasks on quit
	cfg          config.Config
	configPath   string // empty = no persistent save
	firstRun     bool
	status       status.State
	statusStyles status.Styles
	width        int
	height       int
	banner       string
	themeMgr     *theme.Manager
	state        rootState
	styles       theme.Styles
	keys         keys.GlobalKeyMap
	help         help.Model
	modal        modal.Model
	current      screens.Screen
	stack        screenStack
}

// newRootModel creates a new root model.
func newRootModel(ctx context.Context, cancel context.CancelFunc, cfg config.Config, configPath string, firstRun bool) rootModel {
	return rootModel{
		ctx:        ctx,
		cancel:     cancel,
		cfg:        cfg,
		configPath: configPath,
		firstRun:   firstRun,
		status:     status.State{Text: "Ready", Kind: status.KindNone},
		themeMgr:   theme.GetManager(),
		current:    screens.NewHome(),
		keys:       keys.DefaultGlobalKeyMap(),
		help:       help.New(),
	}
}

// Init initializes the root model.
func (m rootModel) Init() tea.Cmd {
	cmds := tea.Batch(
		tea.RequestBackgroundColor,
		m.themeMgr.Init(m.cfg.UI.ThemeName, false, m.width),
	)
	if m.firstRun {
		return tea.Batch(cmds, func() tea.Msg {
			return NavigateMsg{Screen: screens.NewWelcome()}
		})
	}
	return cmds
}

// Update handles messages for the root model.
func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case tea.BackgroundColorMsg:
		return m.handleBgColor(msg)
	case theme.ThemeChangedMsg:
		return m.handleThemeChanged(msg)
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case modal.ShowMsg:
		return m.handleModalShow(msg)
	case modal.ConfirmedMsg, modal.CancelledMsg, modal.PromptSubmittedMsg:
		return m.handleModalDismiss(msg)
	case task.ErrMsg:
		return m.handleTaskErr(msg)
	case screens.WelcomeDoneMsg:
		return m.handleWelcomeDone(msg)
	case NavigateMsg:
		return m.handleNavigate(msg)
	case menu.SelectionMsg:
		return m.handleMenuSelection(msg)
	case screens.SettingsSavedMsg:
		return m.handleSettingsSaved(msg)
	case screens.BackMsg:
		return m.handleBack(msg)
	case status.Msg:
		return m.handleStatus(msg)
	case status.ClearMsg:
		return m.handleStatusClear(msg)
	}
	return m.forwardToScreen(msg)
}

// View renders the root model.
func (m rootModel) View() tea.View {
	if m.state != rootStateReady {
		return tea.NewView("")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.headerView(),
		m.styles.Body.Render(m.current.Body()),
		m.helpView(),
		m.footerView(),
	)

	base := m.styles.App.Render(content)
	if m.modal.Visible() {
		return tea.NewView(modal.Overlay(base, m.modal.View(), m.width, m.height))
	}
	return tea.NewView(base)
}

// --- Update handlers ---

func (m rootModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.state = rootStateReady

	if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
		m.current = setter.SetWidth(m.width)
	}
	if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
		m.current = setter.SetHeight(m.bodyHeight())
	}
	return m, m.themeMgr.SetWidth(m.width)
}

func (m rootModel) handleBgColor(msg tea.BackgroundColorMsg) (tea.Model, tea.Cmd) {
	isDark := msg.IsDark()
	m.help.Styles = help.DefaultStyles(isDark)
	return m, m.themeMgr.SetDarkMode(isDark)
}

func (m rootModel) handleThemeChanged(msg theme.ThemeChangedMsg) (tea.Model, tea.Cmd) {
	m.styles = theme.NewFromPalette(msg.State.Palette, msg.State.Width)
	m.statusStyles = status.NewStyles(msg.State.Palette)
	m.help.SetWidth(m.styles.MaxWidth)

	if m.cfg.UI.ShowBanner {
		m.renderBanner()
	}

	if t, ok := m.current.(theme.Themeable); ok {
		t.ApplyTheme(msg.State)
	}
	return m, nil
}

func (m rootModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modal.Visible() {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}
	if key.Matches(msg, m.keys.Quit) {
		m.cancel()
		return m, tea.Quit
	}
	return m.forwardToScreen(msg)
}

func (m rootModel) handleModalShow(msg modal.ShowMsg) (tea.Model, tea.Cmd) {
	m.modal = modal.New(msg, m.themeMgr.State().Palette)
	return m, nil
}

func (m rootModel) handleModalDismiss(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.modal = modal.Model{}
	updated, cmd := m.current.Update(msg)
	if s, ok := updated.(screens.Screen); ok {
		m.current = s
	}
	return m, cmd
}

func (m rootModel) handleTaskErr(msg task.ErrMsg) (tea.Model, tea.Cmd) {
	return m, status.SetError(msg.Err.Error(), 0)
}

func (m rootModel) handleWelcomeDone(_ screens.WelcomeDoneMsg) (tea.Model, tea.Cmd) {
	m.cfg.ConfigVersion = config.CurrentConfigVersion
	if m.configPath != "" {
		if err := config.Save(&m.cfg, m.configPath); err != nil {
			return m, status.SetError("Save failed: "+err.Error(), 0)
		}
	}
	if m.stack.Len() > 0 {
		m.current = m.stack.Pop()
	}
	return m, status.SetSuccess("Welcome! Config saved.", 0)
}

func (m rootModel) handleNavigate(msg NavigateMsg) (tea.Model, tea.Cmd) {
	m.stack.Push(m.current)
	m.current = msg.Screen
	if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
		m.current = setter.SetWidth(m.width)
	}
	if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
		m.current = setter.SetHeight(m.bodyHeight())
	}
	if t, ok := m.current.(theme.Themeable); ok {
		t.ApplyTheme(m.themeMgr.State())
	}
	return m, m.current.Init()
}

func (m rootModel) handleMenuSelection(msg menu.SelectionMsg) (tea.Model, tea.Cmd) {
	switch msg.Item.ScreenID() {
	case "settings":
		return m.Update(NavigateMsg{Screen: screens.NewSettings(m.cfg)})
	default:
		detail := screens.NewDetail(
			msg.Item.Title(), msg.Item.Description(), msg.Item.ScreenID(), m.ctx,
		)
		return m.Update(NavigateMsg{Screen: detail})
	}
}

func (m rootModel) handleSettingsSaved(msg screens.SettingsSavedMsg) (tea.Model, tea.Cmd) {
	themeChanged := m.cfg.UI.ThemeName != msg.Cfg.UI.ThemeName
	m.cfg = msg.Cfg

	if !msg.Cfg.UI.ShowBanner {
		m.banner = ""
	}

	var saveCmd tea.Cmd
	if m.configPath != "" {
		if err := config.Save(&m.cfg, m.configPath); err != nil {
			saveCmd = status.SetError("Save failed: "+err.Error(), 0)
		} else {
			saveCmd = status.SetSuccess("Settings saved", 0)
		}
	} else {
		saveCmd = status.SetInfo("Settings applied (no config file)", 0)
	}

	if themeChanged {
		if m.stack.Len() > 0 {
			m.current = m.stack.Pop()
		}
		return m, tea.Batch(saveCmd, m.themeMgr.SetThemeName(m.cfg.UI.ThemeName))
	}

	if m.stack.Len() > 0 {
		m.current = m.stack.Pop()
	}
	return m, saveCmd
}

func (m rootModel) handleBack(_ screens.BackMsg) (tea.Model, tea.Cmd) {
	if m.stack.Len() > 0 {
		m.current = m.stack.Pop()
	}
	return m, nil
}

func (m rootModel) handleStatus(msg status.Msg) (tea.Model, tea.Cmd) {
	m.status = status.State{Text: msg.Text, Kind: msg.Kind}
	return m, nil
}

func (m rootModel) handleStatusClear(_ status.ClearMsg) (tea.Model, tea.Cmd) {
	m.status = status.State{Text: "Ready", Kind: status.KindNone}
	return m, nil
}

// forwardToScreen delegates an unhandled message to the current screen.
func (m rootModel) forwardToScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	updated, cmd := m.current.Update(msg)
	if s, ok := updated.(screens.Screen); ok {
		m.current = s
	}
	return m, cmd
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
	left := m.statusStyles.Render(m.status.Text, m.status.Kind)
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

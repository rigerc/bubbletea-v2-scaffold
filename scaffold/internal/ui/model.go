package ui

import (
	"context"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/task"
	"scaffold/internal/ui/header"
	"scaffold/internal/ui/keys"
	"scaffold/internal/ui/menu"
	"scaffold/internal/ui/modal"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/statusbar"
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
	ctx        context.Context
	cancel     context.CancelFunc // shutdown only; cancels all running tasks on quit
	cfg        config.Config
	configPath string // empty = no persistent save
	firstRun   bool
	width      int
	height     int
	bodyH      int // cached body height, updated on resize/navigation/theme change
	themeMgr   *theme.Manager
	state      rootState
	styles     theme.Styles
	keys       keys.GlobalKeyMap
	help       help.Model
	modal      modal.Model
	header     header.Model
	statusbar  statusbar.Model
	current    screens.Screen
	stack      screenStack
}

// newRootModel creates a new root model.
func newRootModel(ctx context.Context, cancel context.CancelFunc, cfg config.Config, configPath string, firstRun bool) rootModel {
	return rootModel{
		ctx:        ctx,
		cancel:     cancel,
		cfg:        cfg,
		configPath: configPath,
		firstRun:   firstRun,
		themeMgr:   theme.GetManager(),
		current:    screens.NewHome(),
		keys:       keys.DefaultGlobalKeyMap(),
		help:       help.New(),
		header:     header.New(cfg),
		statusbar:  statusbar.New(cfg),
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
	}
	return m.broadcast(msg)
}

// View renders the root model.
func (m rootModel) View() tea.View {
	if m.state != rootStateReady {
		return tea.NewView("")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.header.View().Content,
		m.styles.Body.MaxHeight(m.bodyH).Render(m.current.Body()),
		m.helpView(),
		m.statusbar.View().Content,
	)

	base := m.styles.App.Render(content)

	if m.modal.Visible() {
		return tea.NewView(modal.Overlay(base, m.modal.View().Content, m.width, m.height))
	}
	return tea.NewView(base)
}

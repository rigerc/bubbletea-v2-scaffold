// Package ui — Update message handlers for rootModel.
package ui

import (
	"math/rand"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"scaffold/config"
	"scaffold/internal/task"
	"scaffold/internal/ui/menu"
	"scaffold/internal/ui/modal"
	"scaffold/internal/ui/screens"
	"scaffold/internal/ui/status"
	"scaffold/internal/ui/theme"
)

func (m rootModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.width = msg.Width
	m.height = msg.Height
	m.state = rootStateReady

	m.header, cmd = m.header.Update(msg)
	cmds = append(cmds, cmd)
	m.statusbar, cmd = m.statusbar.Update(msg)
	cmds = append(cmds, cmd)

	m.bodyH = m.bodyHeight()

	if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
		m.current = setter.SetWidth(m.width)
	}
	if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
		m.current = setter.SetHeight(m.bodyH)
	}
	return m, tea.Batch(append(cmds, m.themeMgr.SetWidth(m.width))...)
}

func (m rootModel) handleBgColor(msg tea.BackgroundColorMsg) (tea.Model, tea.Cmd) {
	isDark := msg.IsDark()
	m.help.Styles = help.DefaultStyles(isDark)
	return m, m.themeMgr.SetDarkMode(isDark)
}

func (m rootModel) handleThemeChanged(msg theme.ThemeChangedMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.styles = theme.NewFromPalette(msg.State.Palette, msg.State.Width)
	m.help.SetWidth(m.styles.MaxWidth)

	m.header, cmd = m.header.Update(msg)
	cmds = append(cmds, cmd)
	m.statusbar, cmd = m.statusbar.Update(msg)
	cmds = append(cmds, cmd)

	if t, ok := m.current.(theme.Themeable); ok {
		t.ApplyTheme(msg.State)
	}

	m.bodyH = m.bodyHeight()
	return m, tea.Batch(cmds...)
}

func (m rootModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modal.Visible() {
		var cmd tea.Cmd
		m.modal, cmd = m.modal.Update(msg)
		return m, cmd
	}
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}
	if key.Matches(msg, m.keys.RandomTheme) {
		return m.handleRandomTheme()
	}
	return m.broadcast(msg)
}

func (m rootModel) handleRandomTheme() (tea.Model, tea.Cmd) {
	themes := theme.AvailableThemes()
	if len(themes) == 0 {
		return m, nil
	}

	// Pick random theme different from current if possible
	currentTheme := m.cfg.UI.ThemeName
	var candidates []string
	for _, t := range themes {
		if t != currentTheme {
			candidates = append(candidates, t)
		}
	}

	// If only one theme or all same, use first
	if len(candidates) == 0 {
		candidates = themes
	}

	newTheme := candidates[rand.Intn(len(candidates))]
	m.cfg.UI.ThemeName = newTheme
	return m, tea.Batch(
		status.SetInfo("Theme: "+newTheme, 0),
		m.themeMgr.SetThemeName(newTheme),
	)
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
	m.bodyH = m.bodyHeight()
	if m.configPath != "" {
		return m, status.SetSuccess("Welcome! Config saved.", 0)
	}
	return m, status.SetSuccess("Welcome!", 0)
}

func (m rootModel) handleNavigate(msg NavigateMsg) (tea.Model, tea.Cmd) {
	m.stack.Push(m.current)
	m.current = msg.Screen
	// Recompute bodyH: the incoming screen may have different key bindings,
	// which changes help height and therefore available body height.
	m.bodyH = m.bodyHeight()
	if setter, ok := m.current.(interface{ SetWidth(int) screens.Screen }); ok {
		m.current = setter.SetWidth(m.width)
	}
	if setter, ok := m.current.(interface{ SetHeight(int) screens.Screen }); ok {
		m.current = setter.SetHeight(m.bodyH)
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

	// Propagate new config to the header component. WithCfg handles
	// clearing the banner when ShowBanner is disabled and re-rendering it
	// when ShowBanner is newly enabled (using the cached theme state).
	m.header = m.header.WithCfg(m.cfg)

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
		m.bodyH = m.bodyHeight()
		return m, tea.Batch(saveCmd, m.themeMgr.SetThemeName(m.cfg.UI.ThemeName))
	}

	if m.stack.Len() > 0 {
		m.current = m.stack.Pop()
	}
	m.bodyH = m.bodyHeight()
	return m, saveCmd
}

func (m rootModel) handleBack(_ screens.BackMsg) (tea.Model, tea.Cmd) {
	if m.stack.Len() > 0 {
		m.current = m.stack.Pop()
	}
	m.bodyH = m.bodyHeight()
	return m, nil
}

// broadcast sends msg to all chrome components (header, statusbar) and the
// current screen, collecting commands via tea.Batch. It is the fallback for
// all messages not explicitly handled by the root Update switch — this ensures
// status.Msg, status.ClearMsg, and any other unrecognised messages reach the
// components that care about them.
func (m rootModel) broadcast(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.header, cmd = m.header.Update(msg)
	cmds = append(cmds, cmd)

	m.statusbar, cmd = m.statusbar.Update(msg)
	cmds = append(cmds, cmd)

	updated, cmd := m.current.Update(msg)
	if s, ok := updated.(screens.Screen); ok {
		m.current = s
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

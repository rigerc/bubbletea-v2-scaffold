// Package ui — Update message handlers for rootModel.
package ui

import (
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

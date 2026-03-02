package screens

import (
	"scaffold/config"
	"scaffold/internal/ui/modal"
	"scaffold/internal/ui/theme"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// settingsKeyMap defines help-visible keybindings for the settings form.
type settingsKeyMap struct {
	Up      key.Binding
	Down    key.Binding
	Submit  key.Binding
	Reset   key.Binding
	NextTab key.Binding
	PrevTab key.Binding
}

func defaultSettingsKeyMap() settingsKeyMap {
	return settingsKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "shift+tab"),
			key.WithHelp("↑/shift+tab", "prev"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "tab"),
			key.WithHelp("↓/tab", "next"),
		),
		Submit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "submit"),
		),
		Reset: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reset defaults"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("}"),
			key.WithHelp("}", "next group"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("{"),
			key.WithHelp("{", "prev group"),
		),
	}
}

// Settings is the settings screen backed by a dynamic huh form.
type Settings struct {
	theme.ThemeAware

	cfg          *config.Config
	form         *huh.Form
	groups       []config.GroupMeta
	keys         settingsKeyMap
	huhKeys      *huh.KeyMap
	width        int
	height       int
	currentGroup int
	tabStyles    tabStyles
}

// NewSettings creates a Settings screen from a config snapshot.
// The config is value-copied so the form edits a working copy.
func NewSettings(cfg config.Config) *Settings {
	cfgCopy := cfg
	s := &Settings{
		cfg:          &cfgCopy,
		keys:         defaultSettingsKeyMap(),
		currentGroup: 0,
	}
	s.groups = config.Schema(s.cfg)

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	s.huhKeys = km

	// Note: initTabStyles() is called by ApplyTheme which is invoked by handleNavigate

	// Build single stacked form with all groups at a fixed height
	s.form = s.buildForm(cfg.UI.ThemeName)
	return s
}

// RequiredHeight returns the minimum height needed to display the form.
func (s *Settings) RequiredHeight() int {
	return requiredHeight(s.groups)
}

// SetWidth sets the screen width.
func (s *Settings) SetWidth(w int) Screen {
	s.width = w
	s.form = s.form.WithWidth(w)
	return s
}

// SetHeight sets the available body height. The form height is capped at
// RequiredHeight() so all fields are visible when space permits, and huh
// handles internal scrolling when the terminal is shorter.
func (s *Settings) SetHeight(h int) Screen {
	s.height = h
	formH := s.RequiredHeight()
	if h > 0 && h < formH {
		formH = h
	}
	s.form = s.form.WithHeight(formH)
	return s
}

// ApplyTheme implements theme.Themeable.
func (s *Settings) ApplyTheme(state theme.State) {
	s.ApplyThemeState(state)
	s.initTabStyles()
	// Rebuild the form so huh re-applies styles from the new theme.
	// WithTheme alone does not re-style already-initialized fields.
	// Accessor objects write directly to s.cfg, so current edits are preserved.
	s.form = s.buildForm(state.Name)
}

// buildForm constructs the settings form with the given theme applied.
func (s *Settings) buildForm(themeName string) *huh.Form {
	f := buildFormForAllGroups(s.groups).
		WithTheme(theme.HuhTheme(themeName, maxLabelWidth(s.groups), maxDescWidth(s.groups))).
		WithKeyMap(s.huhKeys).
		WithShowHelp(false).
		WithHeight(1)
	if s.width > 0 {
		f = f.WithWidth(s.width)
	}
	return f
}

// Init initializes the settings form.
func (s *Settings) Init() tea.Cmd {
	return s.form.Init()
}

// Update handles messages for the settings screen.
func (s *Settings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle modal response: confirmed reset → dispatch SettingsSavedMsg with defaults.
	if confirmed, ok := msg.(modal.ConfirmedMsg); ok {
		if confirmed.ID == "reset-settings" {
			defaults := config.DefaultConfig()
			return s, func() tea.Msg { return SettingsSavedMsg{Cfg: *defaults} }
		}
	}

	// Handle reset and submit keys
	if s.form.State == huh.StateNormal {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(keyMsg, s.keys.NextTab):
				// Cycle to next group
				if s.currentGroup < len(s.groups)-1 {
					s.currentGroup++
					return s, s.form.NextGroup()
				}
			case key.Matches(keyMsg, s.keys.PrevTab):
				// Cycle to previous group
				if s.currentGroup > 0 {
					s.currentGroup--
					return s, s.form.PrevGroup()
				}
			case key.Matches(keyMsg, s.keys.Reset):
				return s, modal.ShowConfirm(
					"reset-settings",
					"Reset Settings",
					"Restore all defaults and save? This cannot be undone.",
				)
			case keyMsg.String() == "enter":
				// Submit the form with Enter from any field
				form, formCmd := s.form.Update(msg)
				if f, ok := form.(*huh.Form); ok {
					s.form = f
				}
				saved := *s.cfg
				return s, tea.Sequence(formCmd, func() tea.Msg {
					return SettingsSavedMsg{Cfg: saved}
				})
			}
		}
	}

	form, cmd := s.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		s.form = f
	}
	cmds = append(cmds, cmd)

	switch s.form.State {
	case huh.StateCompleted:
		saved := *s.cfg
		return s, func() tea.Msg { return SettingsSavedMsg{Cfg: saved} }
	case huh.StateAborted:
		return s, func() tea.Msg { return BackMsg{} }
	}

	return s, tea.Batch(cmds...)
}

// View renders the settings screen.
func (s *Settings) View() tea.View {
	return tea.NewView(s.Body())
}

// Body returns the body content for layout composition.
func (s *Settings) Body() string {
	if s.form.State != huh.StateNormal {
		return "Applying settings..."
	}
	tabBar := s.renderTabBar()
	formView := s.form.View()
	if tabBar == "" {
		return formView
	}
	return tabBar + "\n" + formView
}

// ShortHelp returns short help key bindings for the global help bar.
func (s *Settings) ShortHelp() []key.Binding {
	if len(s.groups) > 1 {
		return []key.Binding{s.keys.Submit, s.keys.Reset, s.keys.NextTab}
	}
	return []key.Binding{s.keys.Submit, s.keys.Reset}
}

// FullHelp returns full help key bindings for the global help bar.
func (s *Settings) FullHelp() [][]key.Binding {
	if len(s.groups) > 1 {
		return [][]key.Binding{
			{s.keys.Submit, s.keys.Reset},
			{s.keys.NextTab, s.keys.PrevTab},
		}
	}
	return [][]key.Binding{{s.keys.Submit, s.keys.Reset}}
}

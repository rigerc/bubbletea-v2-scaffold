package screens

import (
	"fmt"
	"reflect"
	"strings"

	"scaffold/config"
	"scaffold/internal/ui/modal"
	"scaffold/internal/ui/theme"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// reflectAccessor bridges reflect.Value to huh.Accessor[T].
type reflectAccessor[T any] struct {
	v reflect.Value
}

func (a *reflectAccessor[T]) Get() T {
	return a.v.Interface().(T)
}

func (a *reflectAccessor[T]) Set(val T) {
	a.v.Set(reflect.ValueOf(val))
}

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
	s.form = buildFormForAllGroups(s.groups).
		WithTheme(theme.HuhTheme(cfg.UI.ThemeName)).
		WithKeyMap(km).
		WithShowHelp(false).
		WithHeight(s.RequiredHeight())
	return s
}

// RequiredHeight returns the minimum height needed to display the form.
func (s *Settings) RequiredHeight() int {
	const (
		fieldHeight  = 1
		groupHeader  = 2
		submitHeight = 2
		tabBarHeight = 2
	)

	total := submitHeight
	if len(s.groups) > 1 {
		total += tabBarHeight
	}
	for _, g := range s.groups {
		total += groupHeader
		total += len(g.Fields) * fieldHeight
	}
	return total
}

// initTabStyles creates tab styles from the current theme palette.
func (s *Settings) initTabStyles() {
	p := s.Palette()
	s.tabStyles = tabStyles{
		active: lipgloss.NewStyle().
			Foreground(p.TextInverse).
			Background(p.Primary).
			Bold(true).
			Padding(0, 1),
		inactive: lipgloss.NewStyle().
			Foreground(p.TextMuted).
			Padding(0, 1),
		tabBar: lipgloss.NewStyle().
			Padding(0, 1).
			MarginBottom(1),
	}
}

// renderTabBar renders the horizontal tab bar with group labels.
func (s *Settings) renderTabBar() string {
	if len(s.groups) <= 1 {
		return ""
	}
	// Sync currentGroup with form's actual state by checking focused field
	s.syncCurrentGroup()

	var tabs []string
	for i, g := range s.groups {
		if i == s.currentGroup {
			tabs = append(tabs, s.tabStyles.active.Render(g.Label))
		} else {
			tabs = append(tabs, s.tabStyles.inactive.Render(g.Label))
		}
	}
	return s.tabStyles.tabBar.Render(strings.Join(tabs, " "))
}

// syncCurrentGroup updates currentGroup to match the form's actual focused group.
func (s *Settings) syncCurrentGroup() {
	field := s.form.GetFocusedField()
	if field == nil {
		return
	}

	// Try to get the field's key
	keyer, ok := field.(interface{ GetKey() string })
	if !ok {
		return
	}
	key := keyer.GetKey()

	// Find which group contains this key
	for i, g := range s.groups {
		for _, f := range g.Fields {
			if f.Key == key {
				if s.currentGroup != i {
					s.currentGroup = i
				}
				return
			}
		}
	}
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
	// Rebuild tab styles and form with new theme
	s.initTabStyles()
	s.form = s.form.WithTheme(theme.HuhTheme(state.Name))
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

// buildFormForAllGroups constructs a huh.Form from all config groups.
// Uses LayoutDefault for pagination (one group per page) to handle many fields.
func buildFormForAllGroups(groups []config.GroupMeta) *huh.Form {
	huhGroups := make([]*huh.Group, 0, len(groups))
	for _, g := range groups {
		fields := make([]huh.Field, 0, len(g.Fields))
		for _, fm := range g.Fields {
			if f := buildField(fm); f != nil {
				fields = append(fields, f)
			}
		}
		if len(fields) > 0 {
			huhGroups = append(huhGroups, huh.NewGroup(fields...).Title(g.Label).WithHeight(5))
		}
	}
	if len(huhGroups) > 0 {
		return huh.NewForm(huhGroups...).WithLayout(huh.LayoutDefault)
	}
	return huh.NewForm()
}

// buildField maps a single FieldMeta to a huh.Field.
func buildField(m config.FieldMeta) huh.Field {
	switch m.Kind {
	case config.FieldSelect:
		options := m.Options
		if m.Key == "ui.themeName" {
			options = theme.AvailableThemes()
		}
		opts := make([]huh.Option[string], len(options))
		for i, o := range options {
			opts[i] = huh.NewOption(strings.ToUpper(o[:1])+o[1:], o)
		}
		return huh.NewSelect[string]().
			Key(m.Key).Title(m.Label).
			Inline(true).
			Options(opts...).
			Accessor(&reflectAccessor[string]{v: m.Value})
	case config.FieldConfirm:
		return huh.NewConfirm().
			Key(m.Key).Title(m.Label).
			Affirmative("Yes").Negative("No").
			Inline(true).
			Accessor(&reflectAccessor[bool]{v: m.Value})
	case config.FieldReadOnly:
		return huh.NewNote().
			Title(m.Label + ": " + fmt.Sprint(m.Value.Interface()))
	default: // FieldInput
		// Handle different types for input fields
		switch m.Value.Kind() {
		case reflect.Int:
			return huh.NewInput().
				Key(m.Key).Title(m.Label).
				Inline(true).
				Accessor(&intAccessor{v: m.Value})
		case reflect.Bool:
			return huh.NewConfirm().
				Key(m.Key).Title(m.Label).
				Affirmative("Yes").Negative("No").
				Inline(true).
				Accessor(&reflectAccessor[bool]{v: m.Value})
		default: // string and others
			return huh.NewInput().
				Key(m.Key).Title(m.Label).
				Inline(true).
				Accessor(&reflectAccessor[string]{v: m.Value})
		}
	}
}

// intAccessor bridges reflect.Value for int fields to huh.Accessor[string].
// It converts between int and string representation for huh.Input.
type intAccessor struct {
	v reflect.Value
}

func (a *intAccessor) Get() string {
	return fmt.Sprintf("%d", a.v.Int())
}

func (a *intAccessor) Set(val string) {
	var intVal int
	fmt.Sscanf(val, "%d", &intVal)
	a.v.SetInt(int64(intVal))
}

// tabStyles holds lipgloss styles for the group tab bar.
type tabStyles struct {
	active   lipgloss.Style
	inactive lipgloss.Style
	tabBar   lipgloss.Style
}

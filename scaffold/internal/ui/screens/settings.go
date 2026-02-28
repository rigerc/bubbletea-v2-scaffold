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
	Up     key.Binding
	Down   key.Binding
	Submit key.Binding
	Reset  key.Binding
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
	}
}

// Settings is the settings screen backed by a dynamic huh form.
type Settings struct {
	theme.ThemeAware

	cfg     *config.Config
	form    *huh.Form
	groups  []config.GroupMeta
	keys    settingsKeyMap
	huhKeys *huh.KeyMap
	width   int
	height  int
}

// NewSettings creates a Settings screen from a config snapshot.
// The config is value-copied so the form edits a working copy.
func NewSettings(cfg config.Config) *Settings {
	cfgCopy := cfg
	s := &Settings{
		cfg:  &cfgCopy,
		keys: defaultSettingsKeyMap(),
	}
	s.groups = config.Schema(s.cfg)

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
	s.huhKeys = km

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
	)

	total := submitHeight
	for _, g := range s.groups {
		total += groupHeader
		total += len(g.Fields) * fieldHeight
	}
	return total
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
	// Rebuild form with new theme
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
	return s.form.View()
}

// ShortHelp returns short help key bindings for the global help bar.
func (s *Settings) ShortHelp() []key.Binding {
	return []key.Binding{s.keys.Submit, s.keys.Reset}
}

// FullHelp returns full help key bindings for the global help bar.
func (s *Settings) FullHelp() [][]key.Binding {
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

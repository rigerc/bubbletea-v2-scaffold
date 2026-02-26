package screens

import (
	"fmt"
	"reflect"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/paginator"
	"charm.land/huh/v2"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/config"
	"scaffold/internal/ui/theme"
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
	Left   key.Binding
	Right  key.Binding
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
		Left: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "prev group"),
		),
		Right: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "next group"),
		),
	}
}

// Settings is the settings screen backed by a dynamic huh form.
type Settings struct {
	theme.ThemeAware

	cfg       *config.Config
	form      *huh.Form
	groups    []config.GroupMeta
	paginator paginator.Model
	keys      settingsKeyMap
	width     int
	height    int
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

	// Initialize paginator with one page per group
	s.paginator = paginator.New(
		paginator.WithTotalPages(len(s.groups)),
	)
	s.paginator.Type = paginator.Dots

	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

	// Build form with only the first group
	s.form = buildFormForGroup(s.groups[0]).
		WithTheme(theme.HuhTheme(cfg.UI.ThemeName)).
		WithKeyMap(km).
		WithShowHelp(false)
	return s
}

// RequiredHeight returns the minimum height needed to display the form.
func (s *Settings) RequiredHeight() int {
	const (
		fieldHeight  = 3
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

// SetHeight sets the available body height for scrolling.
func (s *Settings) SetHeight(h int) Screen {
	s.height = h
	if h > 0 {
		s.form = s.form.WithHeight(h)
	}
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

	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		s.width = ws.Width
		s.form = s.form.WithWidth(s.width)
		if ws.Height > 0 {
			s.height = ws.Height
			s.form = s.form.WithHeight(s.height)
		}
	}

	// Handle page navigation with left/right keys
	if s.form.State == huh.StateNormal {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(keyMsg, s.keys.Left):
				if !s.paginator.OnFirstPage() {
					s.paginator.PrevPage()
					return s, s.rebuildFormForCurrentPage()
				}
				return s, nil
			case key.Matches(keyMsg, s.keys.Right):
				if !s.paginator.OnLastPage() {
					s.paginator.NextPage()
					return s, s.rebuildFormForCurrentPage()
				}
				return s, nil
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

	// Show group title with page indicator, form, and paginator
	groupTitle := s.groups[s.paginator.Page].Label
	pageInfo := fmt.Sprintf("(%d/%d)", s.paginator.Page+1, len(s.groups))

	// Style the header with secondary color
	headerStyle := lipgloss.NewStyle().Foreground(s.Palette().Secondary).Bold(true)
	header := headerStyle.Render(groupTitle + " " + pageInfo)

	return header + "\n" + s.form.View() + "\n" + s.paginator.View()
}

// rebuildFormForCurrentPage rebuilds the form for the current page.
func (s *Settings) rebuildFormForCurrentPage() tea.Cmd {
	s.form = buildFormForGroup(s.groups[s.paginator.Page]).
		WithTheme(theme.HuhTheme(s.ThemeState().Name)).
		WithWidth(s.width).
		WithShowHelp(false)
	if s.height > 0 {
		s.form = s.form.WithHeight(s.height)
	}
	return s.form.Init()
}

// ShortHelp returns short help key bindings for the global help bar.
func (s *Settings) ShortHelp() []key.Binding {
	return []key.Binding{s.keys.Left, s.keys.Right, s.keys.Submit}
}

// FullHelp returns full help key bindings for the global help bar.
func (s *Settings) FullHelp() [][]key.Binding {
	return [][]key.Binding{{s.keys.Left, s.keys.Right, s.keys.Submit}}
}

// buildFormForGroup constructs a huh.Form for a single group.
func buildFormForGroup(g config.GroupMeta) *huh.Form {
	fields := make([]huh.Field, 0, len(g.Fields))
	for _, fm := range g.Fields {
		if f := buildField(fm); f != nil {
			fields = append(fields, f)
		}
	}
	if len(fields) > 0 {
		return huh.NewForm(huh.NewGroup(fields...).Title(g.Label))
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
			Key(m.Key).Title(m.Label).Description(m.Desc).
			Options(opts...).
			Accessor(&reflectAccessor[string]{v: m.Value})
	case config.FieldConfirm:
		return huh.NewConfirm().
			Key(m.Key).Title(m.Label).Description(m.Desc).
			Affirmative("Yes").Negative("No").
			Accessor(&reflectAccessor[bool]{v: m.Value})
	case config.FieldReadOnly:
		return huh.NewNote().
			Title(m.Label).
			Description(fmt.Sprint(m.Value.Interface()))
	default: // FieldInput
		return huh.NewInput().
			Key(m.Key).Title(m.Label).Description(m.Desc).
			Accessor(&reflectAccessor[string]{v: m.Value})
	}
}

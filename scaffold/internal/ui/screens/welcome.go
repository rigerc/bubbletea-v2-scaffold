package screens

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/theme"
)

// WelcomeDoneMsg is sent when the user completes the welcome screen.
// rootModel handles it by saving config and navigating back to Home.
type WelcomeDoneMsg struct{}

type welcomeKeyMap struct {
	Continue key.Binding
}

// Welcome is displayed on the first run of the application.
// It shows a brief introduction and waits for the user to press Enter.
type Welcome struct {
	theme.ThemeAware
	keys  welcomeKeyMap
	width int
}

// NewWelcome creates the first-run welcome screen.
func NewWelcome() *Welcome {
	return &Welcome{
		keys: welcomeKeyMap{
			Continue: key.NewBinding(
				key.WithKeys("enter", " "),
				key.WithHelp("enter", "get started"),
			),
		},
	}
}

// SetWidth sets the available render width.
func (w *Welcome) SetWidth(width int) Screen {
	w.width = width
	return w
}

// ApplyTheme implements theme.Themeable.
func (w *Welcome) ApplyTheme(state theme.State) {
	w.ApplyThemeState(state)
}

// Init is a no-op; no commands needed on enter.
func (w *Welcome) Init() tea.Cmd { return nil }

// Update handles key events.
func (w *Welcome) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		if key.Matches(keyMsg, w.keys.Continue) {
			return w, func() tea.Msg { return WelcomeDoneMsg{} }
		}
	}
	return w, nil
}

// View satisfies tea.Model.
func (w *Welcome) View() tea.View { return tea.NewView(w.Body()) }

// Body returns the renderable content for layout composition.
func (w *Welcome) Body() string {
	p := w.Palette()

	headingStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(p.Primary).
		MarginBottom(1)

	subStyle := lipgloss.NewStyle().
		Foreground(p.Secondary).
		Bold(true)

	textStyle := lipgloss.NewStyle().
		Foreground(p.TextPrimary)

	mutedStyle := lipgloss.NewStyle().
		Foreground(p.TextMuted).
		Italic(true)

	features := []string{
		"  • Context-aware async task runner",
		"  • Modal dialogs (confirm, alert, prompt)",
		"  • Theme system with 8 built-in palettes",
		"  • Persistent settings via config file",
	}

	featureLines := make([]string, len(features))
	for i, f := range features {
		featureLines[i] = textStyle.Render(f)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		headingStyle.Render("Welcome to Scaffold"),
		textStyle.Render("A production-ready BubbleTea v2 application template."),
		"",
		subStyle.Render("What's included:"),
		lipgloss.JoinVertical(lipgloss.Left, featureLines...),
		"",
		mutedStyle.Render("Press enter to get started →"),
	)
}

// ShortHelp returns key bindings for the help bar.
func (w *Welcome) ShortHelp() []key.Binding {
	return []key.Binding{w.keys.Continue}
}

// FullHelp returns grouped key bindings for the expanded help bar.
func (w *Welcome) FullHelp() [][]key.Binding {
	return [][]key.Binding{{w.keys.Continue}}
}

// Package screens provides individual screen components for the TUI.
package screens

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"scaffold/internal/ui/menu"
)

// Screen is the interface for screen components that can be composed.
type Screen interface {
	tea.Model
	Body() string // Returns body content for layout composition
}

// KeyBinder is an optional interface for screens that provide key bindings.
type KeyBinder interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// Home is the home screen with a menu.
type Home struct {
	width  int
	isDark bool
	menu   menu.Model
	ready  bool
}

// NewHome creates a new Home screen.
func NewHome() Home {
	m := menu.New()
	return Home{
		menu: m.SetItems([]menu.Item{
			menu.NewItem("Dashboard", "View application dashboard", "dashboard"),
			menu.NewItem("Settings", "Configure application settings", "settings"),
			menu.NewItem("Profile", "Manage your profile", "profile"),
			menu.NewItem("About", "About this application", "about"),
		}),
	}
}

// SetWidth sets the screen width.
func (h Home) SetWidth(w int) Screen {
	h.width = w
	// Calculate menu height dynamically based on number of items
	height := h.menu.RequiredHeight()
	if height == 0 {
		height = 10 // fallback
	}
	h.menu = h.menu.SetSize(w-6, height)
	return h
}

// SetStyles sets the screen styles based on theme name and dark/light mode.
func (h Home) SetStyles(name string, isDark bool) Screen {
	h.isDark = isDark
	h.menu = h.menu.SetStyles(name, isDark)
	return h
}

// Init initializes the home screen.
func (h Home) Init() tea.Cmd {
	return nil
}

// Update handles messages for the home screen.
func (h Home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	h.menu, cmd = h.menu.Update(msg)
	return h, cmd
}

// View renders the home screen.
func (h Home) View() tea.View {
	return tea.NewView(h.Body())
}

// Body returns the body content for layout composition.
func (h Home) Body() string {
	return h.menu.View()
}

// ShortHelp returns short help key bindings for the home screen.
func (h Home) ShortHelp() []key.Binding {
	return h.menu.Keys().ShortHelp()
}

// FullHelp returns full help key bindings for the home screen.
func (h Home) FullHelp() [][]key.Binding {
	return h.menu.Keys().FullHelp()
}

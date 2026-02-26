package screens

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/theme"
)

// Detail is a detail screen that shows information about a selected menu item.
type Detail struct {
	title       string
	description string
	screenID    string
	width       int
	isDark      bool
	themeName   string
	styles      theme.DetailStyles
}

// NewDetail creates a new Detail screen.
func NewDetail(title, description, screenID string) Detail {
	return Detail{
		title:       title,
		description: description,
		screenID:    screenID,
	}
}

// SetWidth sets the screen width.
func (d Detail) SetWidth(w int) Screen {
	d.width = w
	return d
}

// SetStyles sets the screen styles based on theme name and dark/light mode.
func (d Detail) SetStyles(name string, isDark bool) Screen {
	d.themeName = name
	d.isDark = isDark
	d.styles = theme.NewDetailStyles(name, isDark)
	return d
}

// Init initializes the detail screen.
func (d Detail) Init() tea.Cmd {
	return nil
}

// Update handles messages for the detail screen.
func (d Detail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return d, func() tea.Msg { return BackMsg{} }
		}
	}
	return d, nil
}

// View renders the detail screen.
func (d Detail) View() tea.View {
	return tea.NewView(d.Body())
}

// Body returns the body content for layout composition.
func (d Detail) Body() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		d.styles.Title.Render(d.title),
		d.styles.Desc.Render(d.description),
		d.styles.Content.Render(fmt.Sprintf("Screen ID: %s", d.screenID)),
		"",
		d.styles.Info.Render("Press Esc to go back to the menu"),
	)

	return content
}

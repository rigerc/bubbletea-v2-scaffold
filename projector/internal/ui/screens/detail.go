// Package screens provides the individual screen implementations for the application.
package screens

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	appkeys "projector/internal/ui/keys"
	"projector/internal/ui/nav"
)

// detailHelpKeys implements help.KeyMap by combining the viewport scroll
// bindings with the global app bindings (esc, ?) for the help bar.
type detailHelpKeys struct {
	vp  viewport.KeyMap
	app appkeys.GlobalKeyMap
}

func (k detailHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.vp.Up, k.vp.Down, k.app.Back, k.app.Help}
}

func (k detailHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.vp.Up, k.vp.Down, k.vp.HalfPageUp, k.vp.HalfPageDown},
		{k.vp.PageUp, k.vp.PageDown, k.app.Back, k.app.Help},
	}
}

// DetailScreen displays scrollable text content with a pager-style header and footer.
// It implements nav.Screen and nav.Themeable.
type DetailScreen struct {
	ScreenBase
	title, content string
	vp             viewport.Model
	ready          bool // false until first WindowSizeMsg
}

// NewDetailScreen creates a new DetailScreen with the given title and content.
func NewDetailScreen(title, content string, isDark bool, appName string) *DetailScreen {
	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.SoftWrap = true

	return &DetailScreen{
		ScreenBase: NewBase(isDark, appName),
		title:      title,
		content:    content,
		vp:         vp,
	}
}

// Init returns nil (no initial commands needed).
func (s *DetailScreen) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and returns an updated screen and command.
func (s *DetailScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height
		s.updateViewportSize()
		if !s.ready {
			s.applyGutter()
			s.vp.SetContent(s.content)
			s.ready = true
		}

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.Keys.Help):
			s.Help.ShowAll = !s.Help.ShowAll
			s.updateViewportSize()
			return s, nil
		case key.Matches(msg, s.Keys.Back):
			return s, nav.Pop()
		}
	}

	var cmd tea.Cmd
	s.vp, cmd = s.vp.Update(msg)
	return s, cmd
}

// View renders the detail screen: pager-style header, scrollable viewport, footer.
func (s *DetailScreen) View() string {
	if !s.ready {
		return "Loading..."
	}
	helpKeys := detailHelpKeys{vp: s.vp.KeyMap, app: s.Keys}
	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			s.vp.View(),
			s.footerView(),
			s.RenderHelp(helpKeys),
		),
	)
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *DetailScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
	s.applyGutter()
}

// footerView renders a horizontal rule with a scroll-percentage badge on the right.
//
//	──────────────────────────────────┤  42%  │
//	                                  ╰───────╯
func (s *DetailScreen) footerView() string {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	info := lipgloss.NewStyle().
		BorderStyle(b).
		BorderForeground(s.Theme.Palette.Primary).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", s.vp.ScrollPercent()*100))

	lineW := max(0, s.ContentWidth()-lipgloss.Width(info))
	line := s.Theme.Subtle.Render(strings.Repeat("─", lineW))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

// applyGutter sets the viewport's left gutter to show line numbers.
// Called on first render and whenever the theme changes.
func (s *DetailScreen) applyGutter() {
	gutterStyle := s.Theme.Subtle
	s.vp.LeftGutterFunc = func(info viewport.GutterContext) string {
		switch {
		case info.Soft:
			return gutterStyle.Render("     │ ")
		case info.Index >= info.TotalLines:
			return gutterStyle.Render("   ~ │ ")
		default:
			return gutterStyle.Render(fmt.Sprintf("%4d │ ", info.Index+1))
		}
	}
}

// updateViewportSize recalculates viewport dimensions from the window size,
// theme frame, header height, and footer height.
func (s *DetailScreen) updateViewportSize() {
	if !s.IsSized() {
		return
	}
	s.Help.SetWidth(s.ContentWidth())
	headerH := lipgloss.Height(s.HeaderView())
	footerH := lipgloss.Height(s.footerView())
	helpH := lipgloss.Height(s.RenderHelp(detailHelpKeys{vp: s.vp.KeyMap, app: s.Keys}))

	// Calculate content height using shared helper
	vpH := s.CalculateContentHeight(headerH+footerH, helpH)
	s.vp.SetWidth(s.ContentWidth())
	s.vp.SetHeight(vpH)
}

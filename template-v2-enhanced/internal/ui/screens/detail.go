// Package screens provides the individual screen implementations for the application.
package screens

import (
	"fmt"
	"strings"

	lipglossv2 "charm.land/lipgloss/v2"
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"

	appkeys "template-v2-enhanced/internal/ui/keys"
	"template-v2-enhanced/internal/ui/nav"
	"template-v2-enhanced/internal/ui/styles"
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
	title, content string
	keys           appkeys.GlobalKeyMap
	help           help.Model
	theme          styles.Theme
	isDark         bool
	width, height  int
	vp             viewport.Model
	ready          bool // false until first WindowSizeMsg
}

// NewDetailScreen creates a new DetailScreen with the given title and content.
func NewDetailScreen(title, content string, isDark bool) *DetailScreen {
	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.SoftWrap = true

	h := help.New()
	h.Styles = help.DefaultStyles(isDark)

	return &DetailScreen{
		title:   title,
		content: content,
		keys:    appkeys.New(),
		help:    h,
		theme:   styles.New(isDark),
		isDark:  isDark,
		vp:      vp,
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
		s.width, s.height = msg.Width, msg.Height
		s.updateViewportSize()
		if !s.ready {
			s.applyGutter()
			s.vp.SetContent(s.content)
			s.ready = true
		}

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keys.Help):
			s.help.ShowAll = !s.help.ShowAll
			s.updateViewportSize()
			return s, nil
		case key.Matches(msg, s.keys.Back):
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
	helpKeys := detailHelpKeys{vp: s.vp.KeyMap, app: s.keys}
	helpView := lipglossv2.NewStyle().MarginTop(1).Render(s.help.View(helpKeys))
	return s.theme.App.Render(
		lipglossv2.JoinVertical(lipglossv2.Left,
			s.headerView(),
			s.vp.View(),
			s.footerView(),
			helpView,
		),
	)
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *DetailScreen) SetTheme(isDark bool) {
	s.isDark = isDark
	s.theme = styles.New(isDark)
	s.help.Styles = help.DefaultStyles(isDark)
	s.applyGutter()
}

// SetContent updates the viewport content.
func (s *DetailScreen) SetContent(content string) {
	s.content = content
	if s.ready {
		s.vp.SetContent(content)
	}
}

// headerView renders the theme title badge with a horizontal rule extending to the right.
// Vertical padding is increased relative to the shared Title style to give the
// header more visual weight (terminals have no font-size; padding is the lever).
//
//	 Title  ────────────────────────────────
//	(green, tall)
func (s *DetailScreen) headerView() string {
	title := s.theme.Title.Padding(1, 2).Render(s.title)
	lineW := max(0, s.contentWidth()-lipglossv2.Width(title))
	line := s.theme.Subtle.Render(strings.Repeat("─", lineW))
	return lipglossv2.JoinHorizontal(lipglossv2.Center, title, line)
}

// footerView renders a horizontal rule with a scroll-percentage badge on the right.
//
//	──────────────────────────────────┤  42%  │
//	                                  ╰───────╯
func (s *DetailScreen) footerView() string {
	b := lipglossv2.RoundedBorder()
	b.Left = "┤"
	info := lipglossv2.NewStyle().
		BorderStyle(b).
		BorderForeground(lipglossv2.Color("#25A065")).
		Padding(0, 1).
		Render(fmt.Sprintf("%3.f%%", s.vp.ScrollPercent()*100))

	lineW := max(0, s.contentWidth()-lipglossv2.Width(info))
	line := s.theme.Subtle.Render(strings.Repeat("─", lineW))
	return lipglossv2.JoinHorizontal(lipglossv2.Center, line, info)
}

// applyGutter sets the viewport's left gutter to show line numbers.
// Called on first render and whenever the theme changes.
func (s *DetailScreen) applyGutter() {
	gutterStyle := s.theme.Subtle
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

// contentWidth returns the usable width inside the App frame.
func (s *DetailScreen) contentWidth() int {
	frameH, _ := s.theme.App.GetFrameSize()
	return s.width - frameH
}

// updateViewportSize recalculates viewport dimensions from the window size,
// theme frame, header height, and footer height.
func (s *DetailScreen) updateViewportSize() {
	if s.width == 0 || s.height == 0 {
		return
	}
	_, frameV := s.theme.App.GetFrameSize()
	s.help.SetWidth(s.contentWidth())
	headerH := lipglossv2.Height(s.headerView())
	footerH := lipglossv2.Height(s.footerView())

	// Help sits below the footer (outside the pager block) so is not subtracted here.
	vpH := s.height - frameV - headerH - footerH
	if cap := s.height / 3; vpH > cap {
		vpH = cap
	}
	s.vp.SetWidth(s.contentWidth())
	s.vp.SetHeight(vpH)
}

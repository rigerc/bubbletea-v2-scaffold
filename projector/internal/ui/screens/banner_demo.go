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
	"projector/internal/ui/banner"
	"projector/internal/ui/nav"
)

// bannerDemoHelpKeys implements help.KeyMap for the banner demo screen.
type bannerDemoHelpKeys struct {
	vp  viewport.KeyMap
	app appkeys.GlobalKeyMap
}

func (k bannerDemoHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.vp.Up, k.vp.Down, k.app.Back, k.app.Help}
}

func (k bannerDemoHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.vp.Up, k.vp.Down, k.vp.HalfPageUp, k.vp.HalfPageDown},
		{k.vp.PageUp, k.vp.PageDown, k.app.Back, k.app.Help},
	}
}

// BannerDemoScreen displays a scrollable showcase of all safe fonts and gradients.
// It implements nav.Screen and nav.Themeable.
type BannerDemoScreen struct {
	ScreenBase
	vp    viewport.Model
	ready bool
}

// NewBannerDemoScreen creates a banner demo screen.
func NewBannerDemoScreen(isDark bool, appName string) *BannerDemoScreen {
	vp := viewport.New()
	vp.MouseWheelEnabled = true
	vp.SoftWrap = false // ASCII art must not be reflowed

	return &BannerDemoScreen{
		ScreenBase: NewBase(isDark, appName),
		vp:         vp,
	}
}

// Init returns nil (no initial commands needed).
func (s *BannerDemoScreen) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and returns an updated screen and command.
func (s *BannerDemoScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height
		s.updateViewportSize()
		if !s.ready {
			s.vp.SetContent(s.generateContent(s.ContentWidth()))
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

// View renders the banner demo screen.
func (s *BannerDemoScreen) View() string {
	if !s.ready {
		return "Loading..."
	}
	helpKeys := bannerDemoHelpKeys{vp: s.vp.KeyMap, app: s.Keys}
	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			s.vp.View(),
			s.demofooterView(),
			s.RenderHelp(helpKeys),
		),
	)
}

// SetTheme updates the screen's theme based on the terminal background.
// Implements nav.Themeable.
func (s *BannerDemoScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
}

// updateViewportSize recalculates viewport dimensions from the window size.
func (s *BannerDemoScreen) updateViewportSize() {
	if !s.IsSized() {
		return
	}
	s.Help.SetWidth(s.ContentWidth())
	headerH := lipgloss.Height(s.HeaderView())
	footerH := lipgloss.Height(s.demofooterView())
	helpH := lipgloss.Height(s.RenderHelp(bannerDemoHelpKeys{vp: s.vp.KeyMap, app: s.Keys}))

	vpH := s.CalculateContentHeight(headerH+footerH, helpH)
	s.vp.SetWidth(s.ContentWidth())
	s.vp.SetHeight(vpH)
}

// demofooterView renders a horizontal rule with a scroll-percentage badge.
func (s *BannerDemoScreen) demofooterView() string {
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

// generateContent builds the full demo content at the given width.
// It renders each safe font with the aurora gradient, then each gradient with slant.
func (s *BannerDemoScreen) generateContent(width int) string {
	var sb strings.Builder
	ruler := strings.Repeat("─", width)

	// Section 1: fonts
	sb.WriteString("FONTS  ·  aurora gradient\n")
	sb.WriteString(ruler + "\n\n")

	auroraCfg, _ := banner.GradientByName("aurora")
	for _, fontName := range banner.SafeFonts {
		g := auroraCfg
		cfg := banner.BannerConfig{
			Text:     fontName,
			Font:     fontName,
			Gradient: &g,
		}
		result, err := banner.RenderBanner(cfg, width)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[%s: render error: %v]\n", fontName, err))
		} else {
			sb.WriteString(result)
		}
		sb.WriteString("\n")
	}

	// Section 2: gradients
	sb.WriteString("\nGRADIENTS  ·  slant font\n")
	sb.WriteString(ruler + "\n\n")

	for _, grad := range banner.AllGradients() {
		g := grad
		cfg := banner.BannerConfig{
			Text:     grad.Name,
			Font:     "slant",
			Gradient: &g,
		}
		result, err := banner.RenderBanner(cfg, width)
		if err != nil {
			sb.WriteString(fmt.Sprintf("[%s: render error: %v]\n", grad.Name, err))
		} else {
			sb.WriteString(result)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

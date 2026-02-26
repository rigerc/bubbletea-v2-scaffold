// Package theme provides styling for the TUI.
package theme

import (
	"image/color"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/exp/charmtone"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// desaturate returns c with its HSL saturation reduced to s (0–1).
// go-colorful is used here because lipgloss has no saturation adjuster.
func desaturate(c color.Color, s float64) color.Color {
	cf, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	h, _, l := cf.Hsl()
	return colorful.Hsl(h, s, l)
}

// Palette defines semantic colors for the application theme.
type Palette struct {
	// Brand
	Primary       color.Color // primary brand
	PrimaryHover  color.Color // primary hover state
	Secondary     color.Color // secondary brand
	SubtlePrimary color.Color // muted primary, unfocused primary items

	// Text (adaptive)
	TextPrimary   color.Color // primary text
	TextSecondary color.Color // secondary text
	TextMuted     color.Color // borders, subtle elements
	TextInverse   color.Color // text on brand-color backgrounds

	// Status (always visible)
	Success color.Color
	Error   color.Color
	Warning color.Color
	Info    color.Color
}

// AvailableThemes returns the list of built-in theme names.
func AvailableThemes() []string {
	return []string{"default", "ocean", "forest"}
}

// defaultPalette creates the default charmtone-based palette.
func defaultPalette(isDark bool) Palette {
	ld := lipgloss.LightDark(isDark)

	var primary, primaryHover, secondary color.Color
	if isDark {
		primary = lipgloss.Lighten(charmtone.Zinc, 0.12)
		primaryHover = lipgloss.Lighten(charmtone.Zinc, 0.22)
		secondary = lipgloss.Lighten(charmtone.Charple, 0.12)
	} else {
		primary = charmtone.Zinc
		primaryHover = lipgloss.Darken(charmtone.Zinc, 0.08)
		secondary = charmtone.Charple
	}

	return Palette{
		Primary:       primary,
		PrimaryHover:  primaryHover,
		Secondary:     secondary,
		SubtlePrimary: desaturate(charmtone.Zinc, 0.30),

		TextPrimary:   ld(charmtone.Pepper, charmtone.Salt),
		TextSecondary: ld(charmtone.Charcoal, charmtone.Ash),
		TextMuted:     ld(charmtone.Squid, charmtone.Oyster),
		TextInverse:   charmtone.Pepper,

		Error:   lipgloss.Complementary(charmtone.Zinc),
		Success: lipgloss.Alpha(charmtone.Julep, 0.85),
		Warning: lipgloss.Alpha(charmtone.Tang, 0.90),
		Info:    lipgloss.Alpha(charmtone.Thunder, 0.90),
	}
}

// oceanPalette creates a steel-blue / teal palette.
func oceanPalette(isDark bool) Palette {
	ld := lipgloss.LightDark(isDark)

	base := lipgloss.Color("#4A90D9")
	sec := lipgloss.Color("#2BC4C4")
	var primary, primaryHover, secondary color.Color
	if isDark {
		primary = lipgloss.Lighten(base, 0.12)
		primaryHover = lipgloss.Lighten(base, 0.22)
		secondary = lipgloss.Lighten(sec, 0.12)
	} else {
		primary = base
		primaryHover = lipgloss.Darken(base, 0.08)
		secondary = sec
	}

	return Palette{
		Primary:       primary,
		PrimaryHover:  primaryHover,
		Secondary:     secondary,
		SubtlePrimary: desaturate(base, 0.30),

		TextPrimary:   ld(charmtone.Pepper, charmtone.Salt),
		TextSecondary: ld(charmtone.Charcoal, charmtone.Ash),
		TextMuted:     ld(charmtone.Squid, charmtone.Oyster),
		TextInverse:   charmtone.Pepper,

		Error:   lipgloss.Complementary(base),
		Success: lipgloss.Alpha(charmtone.Julep, 0.85),
		Warning: lipgloss.Alpha(charmtone.Tang, 0.90),
		Info:    lipgloss.Alpha(charmtone.Thunder, 0.90),
	}
}

// forestPalette creates a forest-green / amber palette.
func forestPalette(isDark bool) Palette {
	ld := lipgloss.LightDark(isDark)

	base := lipgloss.Color("#4A7C59")
	sec := lipgloss.Color("#C9913D")
	var primary, primaryHover, secondary color.Color
	if isDark {
		primary = lipgloss.Lighten(base, 0.12)
		primaryHover = lipgloss.Lighten(base, 0.22)
		secondary = lipgloss.Lighten(sec, 0.12)
	} else {
		primary = base
		primaryHover = lipgloss.Darken(base, 0.08)
		secondary = sec
	}

	return Palette{
		Primary:       primary,
		PrimaryHover:  primaryHover,
		Secondary:     secondary,
		SubtlePrimary: desaturate(base, 0.30),

		TextPrimary:   ld(charmtone.Pepper, charmtone.Salt),
		TextSecondary: ld(charmtone.Charcoal, charmtone.Ash),
		TextMuted:     ld(charmtone.Squid, charmtone.Oyster),
		TextInverse:   charmtone.Pepper,

		Error:   lipgloss.Complementary(base),
		Success: lipgloss.Alpha(charmtone.Julep, 0.85),
		Warning: lipgloss.Alpha(charmtone.Tang, 0.90),
		Info:    lipgloss.Alpha(charmtone.Thunder, 0.90),
	}
}

// NewPalette creates a semantic color palette for the given theme name and background.
func NewPalette(name string, isDark bool) Palette {
	switch name {
	case "ocean":
		return oceanPalette(isDark)
	case "forest":
		return forestPalette(isDark)
	default:
		return defaultPalette(isDark)
	}
}

// AccentHex returns the primary accent color as a hex string (without '#').
func AccentHex() string {
	return charmtone.Zinc.Hex()[1:] // strip leading '#'
}

// Styles holds all styled components for the UI.
type Styles struct {
	App         lipgloss.Style
	Header      lipgloss.Style
	PlainTitle  lipgloss.Style
	Body        lipgloss.Style
	Help        lipgloss.Style
	Footer      lipgloss.Style
	StatusLeft  lipgloss.Style
	StatusRight lipgloss.Style
	MaxWidth    int
}

// newStylesFromPalette creates Styles from a Palette.
func newStylesFromPalette(p Palette, width int) Styles {
	maxWidth := width * 50 / 100
	if maxWidth < 40 {
		maxWidth = width - 4
	}

	return Styles{
		MaxWidth: maxWidth,
		App:      lipgloss.NewStyle().Width(maxWidth).Padding(0, 0),
		Header:   lipgloss.NewStyle().Padding(2).PaddingBottom(1),
		PlainTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(p.Secondary).
			PaddingBottom(1),
		Body: lipgloss.NewStyle().Padding(0, 3).Foreground(p.TextPrimary),
		Help: lipgloss.NewStyle().MarginTop(0).Padding(0, 3),
		Footer: lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(p.TextSecondary).
			PaddingLeft(1),
		StatusLeft: lipgloss.NewStyle().
			Background(p.Primary).
			Foreground(p.TextInverse).
			Bold(true),
		StatusRight: lipgloss.NewStyle().Foreground(p.TextMuted),
	}
}

// New creates Styles with adaptive colors for the given theme name.
func New(name string, isDark bool, width int) Styles {
	return newStylesFromPalette(NewPalette(name, isDark), width)
}

// DetailStyles holds styles for the detail screen.
type DetailStyles struct {
	Title   lipgloss.Style
	Desc    lipgloss.Style
	Content lipgloss.Style
	Info    lipgloss.Style
}

// newDetailStylesFromPalette creates DetailStyles from a Palette.
func newDetailStylesFromPalette(p Palette) DetailStyles {
	return DetailStyles{
		Title:   lipgloss.NewStyle().Bold(true).Foreground(p.Primary).MarginBottom(1),
		Desc:    lipgloss.NewStyle().Foreground(p.TextMuted).MarginBottom(2),
		Content: lipgloss.NewStyle().Foreground(p.TextPrimary),
		Info:    lipgloss.NewStyle().Foreground(p.TextSecondary).Italic(true),
	}
}

// NewDetailStyles creates detail styles with adaptive colors for the given theme name.
func NewDetailStyles(name string, isDark bool) DetailStyles {
	return newDetailStylesFromPalette(NewPalette(name, isDark))
}

// StatusStyles provides pre-built styles for status messages.
type StatusStyles struct {
	Success lipgloss.Style
	Error   lipgloss.Style
	Warning lipgloss.Style
	Info    lipgloss.Style
}

// NewStatusStyles creates status styles from a Palette for the given theme name.
func NewStatusStyles(name string, isDark bool) StatusStyles {
	p := NewPalette(name, isDark)
	return StatusStyles{
		Success: lipgloss.NewStyle().Foreground(p.Success).Bold(true),
		Error:   lipgloss.NewStyle().Foreground(p.Error).Bold(true),
		Warning: lipgloss.NewStyle().Foreground(p.Warning),
		Info:    lipgloss.NewStyle().Foreground(p.Info),
	}
}

// ListStyles creates list.Styles from a Palette.
func ListStyles(p Palette) list.Styles {
	s := list.DefaultStyles(false)

	s.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	s.Title = lipgloss.NewStyle().
		Background(p.PrimaryHover).
		Foreground(p.TextInverse).
		Padding(0, 1)
	s.Spinner = lipgloss.NewStyle().Foreground(p.Primary)
	s.PaginationStyle = lipgloss.NewStyle().Foreground(p.TextMuted).PaddingLeft(2)
	s.HelpStyle = lipgloss.NewStyle().Foreground(p.TextSecondary).Padding(1, 0, 0, 2)
	s.StatusBar = lipgloss.NewStyle().Foreground(p.TextSecondary).Padding(0, 0, 1, 2)
	s.StatusEmpty = lipgloss.NewStyle().Foreground(p.TextMuted)
	s.NoItems = lipgloss.NewStyle().Foreground(p.TextSecondary)
	s.ActivePaginationDot = lipgloss.NewStyle().Foreground(p.Primary).SetString("•")
	s.InactivePaginationDot = lipgloss.NewStyle().Foreground(p.TextMuted).SetString("•")
	s.DividerDot = lipgloss.NewStyle().Foreground(p.TextMuted).SetString(" • ")

	return s
}

// ListItemStyles creates list.DefaultItemStyles from a Palette.
func ListItemStyles(p Palette) list.DefaultItemStyles {
	s := list.NewDefaultItemStyles(false)

	// Normal state (unfocused items)
	s.NormalTitle = lipgloss.NewStyle().Foreground(p.SubtlePrimary)
	s.NormalDesc = lipgloss.NewStyle().Foreground(p.TextSecondary)

	// Selected state (focused item)
	s.SelectedTitle = lipgloss.NewStyle().
		Foreground(p.PrimaryHover).
		Bold(true)
	s.SelectedDesc = lipgloss.NewStyle().Foreground(p.TextMuted)

	// Dimmed state (when filter input is activated)
	s.DimmedTitle = lipgloss.NewStyle().Foreground(p.TextMuted)
	s.DimmedDesc = lipgloss.NewStyle().Foreground(p.TextMuted)

	// Filter match
	s.FilterMatch = lipgloss.NewStyle().Foreground(p.Primary)

	return s
}

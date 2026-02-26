package theme

import (
	"image/color"
	"sort"

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

// saturate returns c with its HSL saturation set to s (0–1).
// Used for hover states to provide saturation-based feedback.
func saturate(c color.Color, s float64) color.Color {
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

// -----------------------------------------------------------------------------
// Available Themes
// -----------------------------------------------------------------------------

type ThemeSpec struct {
	Name      string
	Base      color.Color
	Secondary color.Color

	// Optional override hook
	Modify func(p Palette, isDark bool) Palette
}

var themeRegistry = map[string]ThemeSpec{}

// -----------------------------------------------------------------------------
// Registration
// -----------------------------------------------------------------------------

func RegisterTheme(spec ThemeSpec) {
	themeRegistry[spec.Name] = spec
}

func AvailableThemes() []string {
	names := make([]string, 0, len(themeRegistry))
	for name := range themeRegistry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// -----------------------------------------------------------------------------
// Core Palette Builder
// -----------------------------------------------------------------------------

func buildPalette(base, sec color.Color, isDark bool) Palette {
	ld := lipgloss.LightDark(isDark)

	var primary, primaryHover, secondary color.Color

	if isDark {
		primary = lipgloss.Lighten(base, 0.12)
		// Hover: increase both lightness AND saturation for clear affordance
		primaryHover = saturate(lipgloss.Lighten(base, 0.22), 0.85)
		secondary = lipgloss.Lighten(sec, 0.12)
	} else {
		primary = base
		// Hover: darken slightly with saturation boost for clear affordance
		primaryHover = saturate(lipgloss.Darken(base, 0.08), 0.90)
		secondary = sec
	}

	// Fixed status colors for consistent UX - independent of brand
	// These are recognizable emotional anchors that users expect
	fixedError := ld(lipgloss.Color("#FF4444"), lipgloss.Color("#CC3333"))    // Red - always red
	fixedSuccess := ld(lipgloss.Color("#44DD66"), lipgloss.Color("#22AA44"))  // Green - always green
	fixedWarning := ld(lipgloss.Color("#FFAA22"), lipgloss.Color("#DD8800"))  // Amber - always amber
	fixedInfo := ld(lipgloss.Color("#44AAFF"), lipgloss.Color("#2277DD"))     // Blue - always blue

	return Palette{
		Primary:       primary,
		PrimaryHover:  primaryHover,
		Secondary:     secondary,
		SubtlePrimary: desaturate(base, 0.30),

		TextPrimary:   ld(charmtone.Pepper, charmtone.Salt),
		TextSecondary: ld(charmtone.Charcoal, charmtone.Ash),
		TextMuted:     ld(charmtone.Squid, charmtone.Oyster),
		TextInverse:   charmtone.Pepper,

		Error:   fixedError,
		Success: fixedSuccess,
		Warning: fixedWarning,
		Info:    fixedInfo,
	}
}

// -----------------------------------------------------------------------------
// Public Factory
// -----------------------------------------------------------------------------

func NewPalette(name string, isDark bool) Palette {
	spec, ok := themeRegistry[name]
	if !ok {
		spec = themeRegistry["default"]
	}

	p := buildPalette(spec.Base, spec.Secondary, isDark)

	if spec.Modify != nil {
		p = spec.Modify(p, isDark)
	}

	return p
}

// -----------------------------------------------------------------------------
// Theme Definitions
// -----------------------------------------------------------------------------

func init() {
	RegisterTheme(ThemeSpec{
		Name:      "default",
		Base:      charmtone.Zinc,
		Secondary: charmtone.Charple,
	})

	RegisterTheme(ThemeSpec{
		Name:      "ocean",
		Base:      lipgloss.Color("#4A90D9"),
		Secondary: lipgloss.Color("#2BC4C4"),
	})

	RegisterTheme(ThemeSpec{
		Name:      "forest",
		Base:      lipgloss.Color("#4A7C59"),
		Secondary: lipgloss.Color("#C9913D"),
	})

	RegisterTheme(ThemeSpec{
		Name:      "sunset",
		Base:      lipgloss.Color("#FF6B6B"),
		Secondary: lipgloss.Color("#5F4B8B"),
	})

	RegisterTheme(ThemeSpec{
		Name:      "aurora",
		Base:      lipgloss.Color("#7F5AF0"),
		Secondary: lipgloss.Color("#2CB67D"),
	})

	RegisterTheme(ThemeSpec{
		Name:      "ember",
		Base:      lipgloss.Color("#8B1E3F"),
		Secondary: lipgloss.Color("#CFAE70"),
		Modify: func(p Palette, _ bool) Palette {
			p.TextInverse = charmtone.Salt
			return p
		},
	})

	RegisterTheme(ThemeSpec{
		Name:      "neon",
		Base:      lipgloss.Color("#00F5D4"),
		Secondary: lipgloss.Color("#FF00C8"),
		Modify: func(p Palette, _ bool) Palette {
			// Neon theme uses brighter, more saturated status colors
			p.Error = lipgloss.Color("#FF3B3B")
			p.Success = lipgloss.Color("#00FF85")
			p.Warning = lipgloss.Color("#FFD60A")
			p.Info = lipgloss.Color("#FF00C8") // Use secondary as info for neon aesthetic
			return p
		},
	})

	RegisterTheme(ThemeSpec{
		Name:      "slate",
		Base:      lipgloss.Color("#3A506B"),
		Secondary: lipgloss.Color("#1C7ED6"),
	})
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
	maxWidth := width * 70 / 100
	if maxWidth < 40 {
		maxWidth = width - 4
	}

	return Styles{
		MaxWidth: maxWidth,
		App:      lipgloss.NewStyle().Width(maxWidth).Padding(0, 0),
		Header:   lipgloss.NewStyle().Padding(2).MarginBottom(0).PaddingBottom(0),
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

// NewFromPalette creates Styles from an existing Palette (avoids recalculation).
func NewFromPalette(p Palette, width int) Styles {
	return newStylesFromPalette(p, width)
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
	s.NormalTitle = lipgloss.NewStyle().Foreground(p.Primary)
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

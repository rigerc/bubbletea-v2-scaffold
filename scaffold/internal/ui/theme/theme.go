package theme

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
	colorful "github.com/lucasb-eyer/go-colorful"
)

// -----------------------------------------------------------------------------
// HCL-Based Color Manipulation (Perceptually Uniform)
// -----------------------------------------------------------------------------

// wrapHue normalizes hue to [0, 360) range.
func wrapHue(h float64) float64 {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	return h
}

// desaturateHcl returns c with its HCL chroma reduced to s (0-1).
// Uses HCL color space for perceptually uniform desaturation.
// Always returns a valid color via Clamped().
func desaturateHcl(c color.Color, s float64) color.Color {
	cf, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	h, _, l := cf.Hcl()
	return colorful.Hcl(h, s, l).Clamped()
}

// saturateHcl returns c with its HCL chroma set to s (0-1).
// Uses HCL color space for perceptually uniform saturation.
func saturateHcl(c color.Color, s float64) color.Color {
	cf, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	h, _, l := cf.Hcl()
	return colorful.Hcl(h, s, l).Clamped()
}

// lightenHcl lightens a color by delta in HCL lightness space.
// delta is additive to L (0-1 range). Positive = lighter.
func lightenHcl(c color.Color, delta float64) color.Color {
	cf, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	h, cVal, l := cf.Hcl()
	newL := math.Max(0, math.Min(1, l+delta))
	return colorful.Hcl(h, cVal, newL).Clamped()
}

// darkenHcl darkens a color by delta in HCL lightness space.
// delta is subtracted from L (0-1 range). Positive = darker.
func darkenHcl(c color.Color, delta float64) color.Color {
	return lightenHcl(c, -delta)
}

// colorDistance computes perceptual distance using CIEDE2000.
func colorDistance(c1, c2 color.Color) float64 {
	cf1, ok1 := colorful.MakeColor(c1)
	cf2, ok2 := colorful.MakeColor(c2)
	if !ok1 || !ok2 {
		return 1.0 // Max distance for invalid colors
	}
	return cf1.DistanceCIEDE2000(cf2)
}

// Backward-compatible aliases using HCL internally.
// These replace the original HSL-based implementations.

// desaturate returns c with its chroma reduced to s (0-1).
// Uses HCL for perceptually uniform desaturation.
func desaturate(c color.Color, s float64) color.Color {
	return desaturateHcl(c, s)
}

// saturate returns c with its chroma set to s (0-1).
// Uses HCL for perceptually uniform saturation.
func saturate(c color.Color, s float64) color.Color {
	return saturateHcl(c, s)
}

// -----------------------------------------------------------------------------
// Color Variant Generation
// -----------------------------------------------------------------------------

// ColorVariant represents a generated color variant.
type ColorVariant struct {
	Name  string
	Color color.Color
}

// GenerateVariants creates perceptually uniform color variants from a base color.
// Uses HCL hue rotation for harmonious color relationships.
func GenerateVariants(base color.Color) []ColorVariant {
	cf, ok := colorful.MakeColor(base)
	if !ok {
		return nil
	}
	h, c, l := cf.Hcl()

	return []ColorVariant{
		{Name: "complementary", Color: colorful.Hcl(wrapHue(h+180), c, l).Clamped()},
		{Name: "analogous1", Color: colorful.Hcl(wrapHue(h+30), c, l).Clamped()},
		{Name: "analogous2", Color: colorful.Hcl(wrapHue(h-30), c, l).Clamped()},
		{Name: "triadic1", Color: colorful.Hcl(wrapHue(h+120), c, l).Clamped()},
		{Name: "triadic2", Color: colorful.Hcl(wrapHue(h-120), c, l).Clamped()},
	}
}

// -----------------------------------------------------------------------------
// Palette Validation
// -----------------------------------------------------------------------------

// ValidatePalette checks that palette colors meet perceptual distance requirements.
// Returns warnings for colors that are too similar (confusion risk).
func ValidatePalette(p Palette) []string {
	const (
		minTextContrastDistance = 0.5
		minStatusColorDistance  = 0.15
	)

	var warnings []string

	// Check text contrast with primary
	if dist := colorDistance(p.TextPrimary, p.Primary); dist < minTextContrastDistance {
		warnings = append(warnings, "TextPrimary may have insufficient contrast with Primary")
	}

	// Check status color distinctness
	statusColors := []struct {
		name string
		col  color.Color
	}{
		{"Success", p.Success},
		{"Error", p.Error},
		{"Warning", p.Warning},
		{"Info", p.Info},
	}

	for i := 0; i < len(statusColors); i++ {
		for j := i + 1; j < len(statusColors); j++ {
			dist := colorDistance(statusColors[i].col, statusColors[j].col)
			if dist < minStatusColorDistance {
				warnings = append(warnings, fmt.Sprintf(
					"%s and %s are too similar (distance: %.2f)",
					statusColors[i].name, statusColors[j].name, dist))
			}
		}
	}

	return warnings
}

// Palette defines semantic colors for the application theme.
type Palette struct {
	// Brand
	Primary         color.Color // primary brand
	PrimaryHover    color.Color // primary hover state
	Secondary       color.Color // secondary brand
	SubtlePrimary   color.Color // muted primary, unfocused primary items
	SubtleSecondary color.Color // muted secondary, unfocused secondary items

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
		primary = lightenHcl(base, 0.12)
		// Hover: increase both lightness AND chroma for clear affordance
		primaryHover = saturateHcl(lightenHcl(base, 0.22), 0.85)
		secondary = lightenHcl(sec, 0.12)
	} else {
		primary = base
		// Hover: darken slightly with chroma boost for clear affordance
		primaryHover = saturateHcl(darkenHcl(base, 0.08), 0.90)
		secondary = sec
	}

	// Fixed status colors for consistent UX - independent of brand
	// These are recognizable emotional anchors that users expect
	fixedError := ld(lipgloss.Color("#FF4444"), lipgloss.Color("#CC3333"))   // Red - always red
	fixedSuccess := ld(lipgloss.Color("#44DD66"), lipgloss.Color("#22AA44")) // Green - always green
	fixedWarning := ld(lipgloss.Color("#FFAA22"), lipgloss.Color("#DD8800")) // Amber - always amber
	fixedInfo := ld(lipgloss.Color("#44AAFF"), lipgloss.Color("#2277DD"))    // Blue - always blue

	return Palette{
		Primary:         primary,
		PrimaryHover:    primaryHover,
		Secondary:       secondary,
		SubtlePrimary:   desaturate(base, 0.30),
		SubtleSecondary: desaturate(secondary, 0.30),

		TextPrimary:   ld(lipgloss.Color("#201F26"), lipgloss.Color("#F1EFEF")),
		TextSecondary: ld(lipgloss.Color("#3A3943"), lipgloss.Color("#DFDBDD")),
		TextMuted:     ld(lipgloss.Color("#858392"), lipgloss.Color("#605F6B")),
		TextInverse:   lipgloss.Color("#201F26"),

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
		Base:      lipgloss.Color("#10B1AE"),
		Secondary: lipgloss.Color("#6B50FF"),
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
			p.TextInverse = lipgloss.Color("#F1EFEF")
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
	maxWidth := width * 90 / 100
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
			BorderForeground(p.TextMuted).
			PaddingLeft(1),
		StatusLeft: lipgloss.NewStyle().
			Background(p.SubtlePrimary).
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
		Desc:    lipgloss.NewStyle().Foreground(p.SubtleSecondary).MarginBottom(2),
		Content: lipgloss.NewStyle().Foreground(p.TextPrimary),
		Info:    lipgloss.NewStyle().Foreground(p.Primary).Italic(true).MarginBottom(1),
	}
}

// NewDetailStyles creates detail styles with adaptive colors for the given theme name.
func NewDetailStyles(name string, isDark bool) DetailStyles {
	return newDetailStylesFromPalette(NewPalette(name, isDark))
}

// NewDetailStylesFromPalette creates DetailStyles from an existing Palette.
func NewDetailStylesFromPalette(p Palette) DetailStyles {
	return newDetailStylesFromPalette(p)
}

// ModalStyles holds styles for modal dialogs.
type ModalStyles struct {
	Title  lipgloss.Style
	Body   lipgloss.Style
	Hint   lipgloss.Style
	Dialog lipgloss.Style
}

// newModalStylesFromPalette creates ModalStyles from a Palette.
func newModalStylesFromPalette(p Palette) ModalStyles {
	return ModalStyles{
		Title: lipgloss.NewStyle().Bold(true).Foreground(p.Primary),
		Body:  lipgloss.NewStyle().Foreground(p.TextPrimary),
		Hint:  lipgloss.NewStyle().Foreground(p.TextMuted).Italic(true),
		Dialog: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(p.Primary).
			Padding(1, 2).
			Width(52),
	}
}

// NewModalStylesFromPalette creates ModalStyles from an existing Palette.
func NewModalStylesFromPalette(p Palette) ModalStyles {
	return newModalStylesFromPalette(p)
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
	s.NormalDesc = lipgloss.NewStyle().Foreground(p.TextMuted)

	// Selected state (focused item)
	s.SelectedTitle = lipgloss.NewStyle().
		Foreground(p.PrimaryHover).
		Bold(true)
	s.SelectedDesc = lipgloss.NewStyle().Foreground(p.SubtleSecondary)

	// Dimmed state (when filter input is activated)
	s.DimmedTitle = lipgloss.NewStyle().Foreground(p.TextMuted)
	s.DimmedDesc = lipgloss.NewStyle().Foreground(p.TextMuted)

	// Filter match
	s.FilterMatch = lipgloss.NewStyle().Foreground(p.Primary)

	return s
}

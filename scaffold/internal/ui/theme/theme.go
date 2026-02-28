// Package theme manages application color themes, including palette generation,
// perceptually uniform color manipulation, and lipgloss/huh style construction.
// Themes are registered via [RegisterTheme] and queried through [NewPalette].
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

// withAlpha returns c with alpha channel simulated (0-1 range).
// Since lipgloss doesn't support alpha, this desaturates and lightens/darkens
// to simulate transparency. Blends toward lightness 0.5 (gray midpoint), which
// is appropriate for borders and muted text where blending to the background
// would lose too much visibility.
func withAlpha(c color.Color, alpha float64) color.Color {
	cf, ok := colorful.MakeColor(c)
	if !ok {
		return c
	}
	h, chroma, l := cf.Hcl()
	newC := chroma * alpha
	newL := l + (0.5-l)*(1-alpha)
	return colorful.Hcl(h, newC, newL).Clamped()
}

// contrastingForeground returns white or black based on luminance to ensure
// high contrast text on the given background color.
// Uses the YIQ formula which is specifically designed for readability.
func contrastingForeground(bg color.Color) color.Color {
	cf, ok := colorful.MakeColor(bg)
	if !ok {
		return lipgloss.Color("#201F26") // default dark
	}
	// colorful.Color stores R,G,B as float64 in [0,1], so yiq is in [0,1]; threshold is 0.5.
	yiq := (cf.R*299 + cf.G*587 + cf.B*114) / 1000
	if yiq >= 0.5 {
		return lipgloss.Color("#201F26") // dark text on light bg
	}
	return lipgloss.Color("#F1EFEF") // light text on dark bg
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

	// Check text contrast with background
	if dist := colorDistance(p.Foreground, p.Background); dist < minTextContrastDistance {
		warnings = append(warnings, "Foreground may have insufficient contrast with Background")
	}

	// Check primary/on-primary contrast
	if dist := colorDistance(p.Primary, p.OnPrimary); dist < minTextContrastDistance {
		warnings = append(warnings, "Primary and OnPrimary may have insufficient contrast")
	}

	// Check secondary/on-secondary contrast
	if dist := colorDistance(p.Secondary, p.OnSecondary); dist < minTextContrastDistance {
		warnings = append(warnings, "Secondary and OnSecondary may have insufficient contrast")
	}

	// Check status/on-status contrast
	statusChecks := []struct {
		name string
		col  color.Color
		on   color.Color
	}{
		{"Success", p.Success, p.OnSuccess},
		{"Error", p.Error, p.OnError},
		{"Warning", p.Warning, p.OnWarning},
		{"Info", p.Info, p.OnInfo},
	}

	for _, check := range statusChecks {
		if dist := colorDistance(check.col, check.on); dist < minTextContrastDistance {
			warnings = append(warnings, fmt.Sprintf("%s and On%s may have insufficient contrast", check.name, check.name))
		}
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
	// ── Core Colors (explicitly set by each theme) ──
	Primary    color.Color // primary brand/action fill
	Secondary  color.Color // secondary brand/action fill
	Background color.Color // page/app background
	Surface    color.Color // card, panel, sheet
	Foreground color.Color // primary text/icons on Background/Surface

	// ── Computed from Surface ──────────────────────────
	SurfaceRaised color.Color // elevated surface (popover, dropdown)
	Overlay       color.Color // scrim behind modals (alpha 0.5 Foreground)
	Border        color.Color // default border (alpha 0.12 Foreground)
	BorderMuted   color.Color // subtle separators (alpha 0.06 Foreground)

	// ── Computed from Foreground ──────────────────────
	ForegroundMuted  color.Color // secondary text (alpha 0.6 Foreground)
	ForegroundSubtle color.Color // placeholder/disabled (alpha 0.38 Foreground)

	// ── Computed from Primary/Secondary ───────────────
	OnPrimary      color.Color // text on Primary (high contrast white/black)
	PrimaryMuted   color.Color // low-emphasis primary (alpha 0.12 Primary)
	OnSecondary    color.Color // text on Secondary (high contrast white/black)
	SecondaryMuted color.Color // low-emphasis secondary (alpha 0.12 Secondary)

	// ── Interactive ───────────────────────────────────
	Focus color.Color // focus ring (always = Primary unless overridden)

	// Status (defaults derived from isDark; can be overridden via Modify)
	Success color.Color
	Error   color.Color
	Warning color.Color
	Info    color.Color

	OnSuccess color.Color // high contrast text on Success
	OnError   color.Color // high contrast text on Error
	OnWarning color.Color // high contrast text on Warning
	OnInfo    color.Color // high contrast text on Info
}

// -----------------------------------------------------------------------------
// Available Themes
// -----------------------------------------------------------------------------

// ThemeSpec defines a named theme by its core colors.
// Register themes with [RegisterTheme] before calling [NewPalette].
// An optional Modify hook can adjust the generated Palette after derivation.
type ThemeSpec struct {
	Name       string
	Primary    color.Color // primary brand/action
	Secondary  color.Color // secondary brand/action
	Background color.Color // page/app background
	Surface    color.Color // card, panel, sheet
	Foreground color.Color // primary text/icons

	// Optional override hook
	Modify func(p Palette, isDark bool) Palette
}

var themeRegistry = map[string]ThemeSpec{}

// -----------------------------------------------------------------------------
// Registration
// -----------------------------------------------------------------------------

// RegisterTheme adds spec to the global theme registry.
// Call this in an init function before the TUI starts.
// RegisterTheme is not concurrency-safe; call only from init().
func RegisterTheme(spec ThemeSpec) {
	themeRegistry[spec.Name] = spec
}

// AvailableThemes returns the sorted names of all registered themes.
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

func buildPalette(spec ThemeSpec, isDark bool) Palette {
	// ── SurfaceRaised: lighten Surface in light mode, darken in dark mode
	var surfaceRaised color.Color
	if isDark {
		surfaceRaised = darkenHcl(spec.Surface, 0.08)
	} else {
		surfaceRaised = lightenHcl(spec.Surface, 0.08)
	}

	// ── Overlay, Border, BorderMuted from Foreground with simulated alpha
	overlay := withAlpha(spec.Foreground, 0.5)
	border := withAlpha(spec.Foreground, 0.12)
	borderMuted := withAlpha(spec.Foreground, 0.06)

	// ── ForegroundMuted, ForegroundSubtle from Foreground
	foregroundMuted := withAlpha(spec.Foreground, 0.6)
	foregroundSubtle := withAlpha(spec.Foreground, 0.38)

	// ── OnPrimary, PrimaryMuted from Primary
	onPrimary := contrastingForeground(spec.Primary)
	primaryMuted := withAlpha(spec.Primary, 0.12)

	// ── OnSecondary, SecondaryMuted from Secondary
	onSecondary := contrastingForeground(spec.Secondary)
	secondaryMuted := withAlpha(spec.Secondary, 0.12)

	// ── Status colors (defaults; can be overridden via Modify)
	var success, warning, info, errColor color.Color
	if isDark {
		success = lipgloss.Color("#44DD66")
		warning = lipgloss.Color("#FFAA22")
		info = lipgloss.Color("#44AAFF")
		errColor = lipgloss.Color("#FF4444")
	} else {
		success = lipgloss.Color("#22AA44")
		warning = lipgloss.Color("#DD8800")
		info = lipgloss.Color("#2277DD")
		errColor = lipgloss.Color("#CC3333")
	}

	return Palette{
		// Core colors (from spec)
		Primary:    spec.Primary,
		Secondary:  spec.Secondary,
		Background: spec.Background,
		Surface:    spec.Surface,
		Foreground: spec.Foreground,

		// Computed from Surface
		SurfaceRaised: surfaceRaised,
		Overlay:       overlay,
		Border:        border,
		BorderMuted:   borderMuted,

		// Computed from Foreground
		ForegroundMuted:  foregroundMuted,
		ForegroundSubtle: foregroundSubtle,

		// Computed from Primary
		OnPrimary:    onPrimary,
		PrimaryMuted: primaryMuted,

		// Computed from Secondary
		OnSecondary:    onSecondary,
		SecondaryMuted: secondaryMuted,

		// Interactive
		Focus: spec.Primary,

		// Status
		Success:   success,
		Error:     errColor,
		Warning:   warning,
		Info:      info,
		OnSuccess: contrastingForeground(success),
		OnError:   contrastingForeground(errColor),
		OnWarning: contrastingForeground(warning),
		OnInfo:    contrastingForeground(info),
	}
}

// -----------------------------------------------------------------------------
// Public Factory
// -----------------------------------------------------------------------------

// NewPalette generates a [Palette] for named theme.
// If name is unknown, it falls back to "default" theme.
// If "default" is also not registered, it uses hardcoded sentinel colors.
// isDark selects the dark or light variant.
func NewPalette(name string, isDark bool) Palette {
	spec, ok := themeRegistry[name]
	if !ok {
		spec, ok = themeRegistry["default"]
		if !ok {
			// Fallback sentinel colors
			spec = ThemeSpec{
				Name:       "default",
				Primary:    lipgloss.Color("#10B1AE"),
				Secondary:  lipgloss.Color("#6B50FF"),
				Background: lipgloss.Color("#16161A"),
				Surface:    lipgloss.Color("#1A1A1F"),
				Foreground: lipgloss.Color("#F1EFEF"),
			}
		}
	}

	p := buildPalette(spec, isDark)

	if spec.Modify != nil {
		p = spec.Modify(p, isDark)
	}

	return p
}

// -----------------------------------------------------------------------------
// Theme Definitions
// -----------------------------------------------------------------------------

func init() {
	// default — teal primary, purple secondary
	RegisterTheme(ThemeSpec{
		Name:       "default",
		Primary:    lipgloss.Color("#10B1AE"),
		Secondary:  lipgloss.Color("#6B50FF"),
		Background: lipgloss.Color("#16161A"),
		Surface:    lipgloss.Color("#1A1A1F"),
		Foreground: lipgloss.Color("#F1EFEF"),
	})

	// ocean — blue primary, cyan secondary
	RegisterTheme(ThemeSpec{
		Name:       "ocean",
		Primary:    lipgloss.Color("#4A90D9"),
		Secondary:  lipgloss.Color("#2BC4C4"),
		Background: lipgloss.Color("#0A1628"),
		Surface:    lipgloss.Color("#111D32"),
		Foreground: lipgloss.Color("#E8F4FD"),
	})

	// forest — green primary, amber secondary
	RegisterTheme(ThemeSpec{
		Name:       "forest",
		Primary:    lipgloss.Color("#4A7C59"),
		Secondary:  lipgloss.Color("#C9913D"),
		Background: lipgloss.Color("#0F1A14"),
		Surface:    lipgloss.Color("#16241D"),
		Foreground: lipgloss.Color("#F0F7ED"),
	})

	// sunset — pink primary, purple secondary
	RegisterTheme(ThemeSpec{
		Name:       "sunset",
		Primary:    lipgloss.Color("#FF6B6B"),
		Secondary:  lipgloss.Color("#5F4B8B"),
		Background: lipgloss.Color("#1F1419"),
		Surface:    lipgloss.Color("#2A1D23"),
		Foreground: lipgloss.Color("#FFF5F5"),
	})

	// aurora — purple primary, green secondary
	RegisterTheme(ThemeSpec{
		Name:       "aurora",
		Primary:    lipgloss.Color("#7F5AF0"),
		Secondary:  lipgloss.Color("#2CB67D"),
		Background: lipgloss.Color("#141420"),
		Surface:    lipgloss.Color("#1E1D2A"),
		Foreground: lipgloss.Color("#F5F0FF"),
	})

	// ember — red primary, gold secondary (custom OnPrimary/OnSecondary)
	RegisterTheme(ThemeSpec{
		Name:       "ember",
		Primary:    lipgloss.Color("#8B1E3F"),
		Secondary:  lipgloss.Color("#CFAE70"),
		Background: lipgloss.Color("#1A0F13"),
		Surface:    lipgloss.Color("#25161C"),
		Foreground: lipgloss.Color("#FFF8F0"),
		Modify: func(p Palette, _ bool) Palette {
			p.OnPrimary = lipgloss.Color("#F1EFEF")
			p.OnSecondary = lipgloss.Color("#201F26")
			return p
		},
	})

	// neon — cyan primary, magenta secondary (bright status, magenta focus)
	RegisterTheme(ThemeSpec{
		Name:       "neon",
		Primary:    lipgloss.Color("#00F5D4"),
		Secondary:  lipgloss.Color("#FF00C8"),
		Background: lipgloss.Color("#0A1A1C"),
		Surface:    lipgloss.Color("#12272A"),
		Foreground: lipgloss.Color("#F0FFFA"),
		Modify: func(p Palette, _ bool) Palette {
			// Brighter, more saturated status colors
			p.Error = lipgloss.Color("#FF3B3B")
			p.Success = lipgloss.Color("#00FF85")
			p.Warning = lipgloss.Color("#FFD60A")
			p.Info = lipgloss.Color("#FF00C8")
			p.Focus = lipgloss.Color("#FF00C8") // magenta focus
			p.OnPrimary = lipgloss.Color("#201F26")
			p.OnSecondary = lipgloss.Color("#F1EFEF")
			return p
		},
	})

	// slate — blue-grey primary, blue secondary
	RegisterTheme(ThemeSpec{
		Name:       "slate",
		Primary:    lipgloss.Color("#3A506B"),
		Secondary:  lipgloss.Color("#1C7ED6"),
		Background: lipgloss.Color("#0F141A"),
		Surface:    lipgloss.Color("#182029"),
		Foreground: lipgloss.Color("#E8EDF5"),
	})

	// sakura — cherry blossom pink, lavender secondary
	RegisterTheme(ThemeSpec{
		Name:       "sakura",
		Primary:    lipgloss.Color("#E87EA1"),
		Secondary:  lipgloss.Color("#9B72CF"),
		Background: lipgloss.Color("#1F151C"),
		Surface:    lipgloss.Color("#2C1E27"),
		Foreground: lipgloss.Color("#FFF5FA"),
	})

	// nord — arctic blue primary, frost cyan secondary
	RegisterTheme(ThemeSpec{
		Name:       "nord",
		Primary:    lipgloss.Color("#5E81AC"),
		Secondary:  lipgloss.Color("#88C0D0"),
		Background: lipgloss.Color("#10171E"),
		Surface:    lipgloss.Color("#19222D"),
		Foreground: lipgloss.Color("#ECEFF4"),
	})

	// mono — monochrome minimal
	RegisterTheme(ThemeSpec{
		Name:       "mono",
		Primary:    lipgloss.Color("#787878"),
		Secondary:  lipgloss.Color("#A8A8A8"),
		Background: lipgloss.Color("#121212"),
		Surface:    lipgloss.Color("#1C1C1C"),
		Foreground: lipgloss.Color("#E8E8E8"),
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
		App:    lipgloss.NewStyle().Width(maxWidth).Padding(0, 0),
		Header: lipgloss.NewStyle().Padding(2).MarginBottom(0).PaddingBottom(0),
		PlainTitle: lipgloss.NewStyle().
			Bold(true).
			Foreground(p.Primary).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(p.Secondary).
			PaddingBottom(1),
		Body: lipgloss.NewStyle().Padding(0, 3).Foreground(p.Foreground),
		Help: lipgloss.NewStyle().MarginTop(0).Padding(0, 3),
		Footer: lipgloss.NewStyle().
			MarginTop(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(p.Border).
			PaddingLeft(1),
		StatusLeft: lipgloss.NewStyle().
			Background(p.PrimaryMuted).
			Foreground(p.OnPrimary).
			Bold(true),
		StatusRight: lipgloss.NewStyle().Foreground(p.ForegroundSubtle),
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
		Desc:    lipgloss.NewStyle().Foreground(p.SecondaryMuted).MarginBottom(2),
		Content: lipgloss.NewStyle().Foreground(p.Foreground),
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
		Body:  lipgloss.NewStyle().Foreground(p.Foreground),
		Hint:  lipgloss.NewStyle().Foreground(p.ForegroundSubtle).Italic(true),
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
		Background(p.Primary).
		Foreground(p.OnPrimary).
		Padding(0, 1)
	s.Spinner = lipgloss.NewStyle().Foreground(p.Primary)
	s.PaginationStyle = lipgloss.NewStyle().Foreground(p.ForegroundSubtle).PaddingLeft(2)
	s.HelpStyle = lipgloss.NewStyle().Foreground(p.ForegroundMuted).Padding(1, 0, 0, 2)
	s.StatusBar = lipgloss.NewStyle().Foreground(p.ForegroundMuted).Padding(0, 0, 1, 2)
	s.StatusEmpty = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)
	s.NoItems = lipgloss.NewStyle().Foreground(p.ForegroundMuted)
	s.ActivePaginationDot = lipgloss.NewStyle().Foreground(p.Primary).SetString("•")
	s.InactivePaginationDot = lipgloss.NewStyle().Foreground(p.ForegroundSubtle).SetString("•")
	s.DividerDot = lipgloss.NewStyle().Foreground(p.ForegroundSubtle).SetString(" • ")

	return s
}

// ListItemStyles creates list.DefaultItemStyles from a Palette.
func ListItemStyles(p Palette) list.DefaultItemStyles {
	s := list.NewDefaultItemStyles(false)

	// Normal state (unfocused items)
	s.NormalTitle = lipgloss.NewStyle().Foreground(p.Primary)
	s.NormalDesc = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)

	// Selected state (focused item)
	s.SelectedTitle = lipgloss.NewStyle().
		Foreground(p.Primary).
		Bold(true)
	s.SelectedDesc = lipgloss.NewStyle().Foreground(p.SecondaryMuted)

	// Dimmed state (when filter input is activated)
	s.DimmedTitle = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)
	s.DimmedDesc = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)

	// Filter match
	s.FilterMatch = lipgloss.NewStyle().Foreground(p.Primary)

	return s
}

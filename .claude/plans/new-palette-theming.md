# Plan: Refactor Theme to New Palette Structure

## Context

The current theme palette uses a brand-focused structure where all colors are derived from just two seed colors (base and secondary). This refactor introduces a more robust, design-system-aligned palette where each theme explicitly defines 5 core colors (Primary, Secondary, Background, Surface, Foreground), with all other semantic colors computed from these values. This enables greater theme flexibility while maintaining consistency.

## New Palette Structure

```go
type Palette struct {
    // ── Core Colors (explicitly set by each theme) ──
    Primary    color.Color // primary brand/action fill
    Secondary  color.Color // secondary brand/action fill
    Background color.Color // page/app background
    Surface    color.Color // card, panel, sheet
    Foreground color.Color // primary text/icons on Background/Surface

    // ── Computed from Background ──────────────────────
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

    // ── Interactive ──────────────────────────────────
    Focus color.Color // focus ring (always = Primary)

    // Status (can be derived from hue or overridden)
    Success color.Color
    Error   color.Color
    Warning color.Color
    Info    color.Color
    OnSuccess color.Color // high contrast text on Success
    OnError   color.Color // high contrast text on Error
    OnWarning color.Color // high contrast text on Warning
    OnInfo    color.Color // high contrast text on Info
}
```

## Color Computation Rules

```go
// ── From Background ──────────────────────────────
SurfaceRaised    = Lighten(Surface, 4%)   // or Darken in dark mode
Overlay          = WithAlpha(Foreground, 0.5)
Border           = WithAlpha(Foreground, 0.12)
BorderMuted      = WithAlpha(Foreground, 0.06)

// ── From Foreground ──────────────────────────────
ForegroundMuted  = WithAlpha(Foreground, 0.6)
ForegroundSubtle = WithAlpha(Foreground, 0.38)

// ── From Primary / Secondary ─────────────────────
OnPrimary        = ContrastingForeground(Primary)   // white or black based on luminance
PrimaryMuted     = WithAlpha(Primary, 0.12)
OnSecondary      = ContrastingForeground(Secondary)
SecondaryMuted   = WithAlpha(Secondary, 0.12)

// ── From Status colors (optional - can override) ──
OnSuccess        = ContrastingForeground(Success)
OnError          = ContrastingForeground(Error)
OnWarning        = ContrastingForeground(Warning)
OnInfo           = ContrastingForeground(Info)

// ── Interactive ──────────────────────────────────
Focus            = Primary
```

## Implementation Steps

### Step 1: Add Helper Functions for Color Computation

**File:** `scaffold/internal/ui/theme/theme.go` (add after existing HCL functions, around line 70)

Add these new helper functions:

```go
// withAlpha returns c with alpha channel set (0-1 range).
// Since lipgloss doesn't support alpha, this desaturates
// and lightens/darkens to simulate transparency.
// Blends toward lightness 0.5 (gray midpoint) which is
// appropriate for borders and muted text where blending to
// background would lose too much visibility.
func withAlpha(c color.Color, alpha float64) color.Color {
    cf, ok := colorful.MakeColor(c)
    if !ok {
        return c
    }
    h, _, l := cf.Hcl()
    // Simulate alpha by reducing chroma and blending toward lightness 0.5
    newC := cf.C * alpha
    newL := l + (0.5 - l) * (1 - alpha)
    return colorful.Hcl(h, newC, newL).Clamped()
}

// contrastingForeground returns white or black based on luminance
// to ensure high contrast text on the given background color.
// Uses YIQ formula which is specifically designed for readability.
func contrastingForeground(bg color.Color) color.Color {
    cf, ok := colorful.MakeColor(bg)
    if !ok {
        return lipgloss.Color("#201F26") // default dark
    }
    // Use YIQ luminance for readability
    yiq := (cf.R*299 + cf.G*587 + cf.B*114) / 1000
    if yiq >= 128 {
        return lipgloss.Color("#201F26") // dark text on light bg
    }
    return lipgloss.Color("#F1EFEF") // light text on dark bg
}

// lightenPercent lightens by percentage (0-1 range).
func lightenPercent(c color.Color, percent float64) color.Color {
    cf, ok := colorful.MakeColor(c)
    if !ok {
        return c
    }
    _, _, l := cf.Hcl()
    newL := math.Min(1, l + percent)
    _, cVal, _ := cf.Hcl()
    return colorful.Hcl(cf.H, cVal, newL).Clamped()
}

// darkenPercent darkens by percentage (0-1 range).
func darkenPercent(c color.Color, percent float64) color.Color {
    return lightenPercent(c, -percent)
}
```

### Step 2: Update Palette Struct

**File:** `scaffold/internal/ui/theme/theme.go` (lines 154-174)

Replace with new Palette struct.

### Step 3: Update ThemeSpec Struct

**File:** `scaffold/internal/ui/theme/theme.go` (lines 180-191)

Update to include Background, Surface, Foreground:

```go
// ThemeSpec defines a named theme by its core colors.
// Register themes with [RegisterTheme] before calling [NewPalette].
// An optional Modify hook can adjust generated Palette after derivation.
type ThemeSpec struct {
    Name      string
    Primary   color.Color // primary brand/action
    Secondary color.Color // secondary brand/action
    Background color.Color // page/app background
    Surface    color.Color // card, panel, sheet
    Foreground color.Color // primary text/icons

    // Optional override hook
    Modify func(p Palette, isDark bool) Palette
}
```

### Step 4: Rewrite `buildPalette()` Function

**File:** `scaffold/internal/ui/theme/theme.go` (lines 219-261)

Replace with new computation logic:

```go
func buildPalette(spec ThemeSpec, isDark bool) Palette {
    // ── SurfaceRaised: light from Background by 8% or darken in dark mode
    // 0.08 in HCL lightness (0-1 range) ≈ 20% perceptual lightness
    var surfaceRaised color.Color
    if isDark {
        surfaceRaised = darkenPercent(spec.Surface, 0.08)
    } else {
        surfaceRaised = lightenPercent(spec.Surface, 0.08)
    }

    // ── Overlay, Border, BorderMuted from Foreground with alpha
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

    // ── Status colors (default derived from hue with light mode adjustment)
    // Using light mode base colors as reference
    var success, warning, info, error color.Color
    if isDark {
        // Dark mode: lighter variants
        success = lipgloss.Color("#44DD66")
        warning = lipgloss.Color("#FFAA22")
        info = lipgloss.Color("#44AAFF")
        error = lipgloss.Color("#FF4444")
    } else {
        // Light mode: darker variants for better contrast
        success = lipgloss.Color("#22AA44")
        warning = lipgloss.Color("#DD8800")
        info = lipgloss.Color("#2277DD")
        error = lipgloss.Color("#CC3333")
    }

    // OnStatus colors
    onSuccess := contrastingForeground(success)
    onWarning := contrastingForeground(warning)
    onInfo := contrastingForeground(info)
    onError := contrastingForeground(error)

    return Palette{
        // Core colors (from spec)
        Primary:    spec.Primary,
        Secondary:  spec.Secondary,
        Background: spec.Background,
        Surface:    spec.Surface,
        Foreground: spec.Foreground,

        // Computed from Background
        SurfaceRaised: surfaceRaised,
        Overlay:       overlay,
        Border:        border,
        BorderMuted:   borderMuted,

        // Computed from Foreground
        ForegroundMuted:  foregroundMuted,
        ForegroundSubtle: foregroundSubtle,

        // Computed from Primary
        OnPrimary:    onPrimary,
        PrimaryMuted:   primaryMuted,

        // Computed from Secondary
        OnSecondary:    onSecondary,
        SecondaryMuted: secondaryMuted,

        // Interactive
        Focus: spec.Primary,

        // Status
        Success:  success,
        Error:    error,
        Warning:  warning,
        Info:     info,
        OnSuccess: onSuccess,
        OnError:   onError,
        OnWarning: onWarning,
        OnInfo:    onInfo,
    }
}
```

### Step 5: Update `NewPalette()` Function

**File:** `scaffold/internal/ui/theme/theme.go` (lines 274-294)

Update to pass ThemeSpec instead of separate colors:

```go
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
                Name:      "default",
                Primary:   lipgloss.Color("#10B1AE"),
                Secondary: lipgloss.Color("#6B50FF"),
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
```

### Step 6: Update Theme Registry (All 10 Themes)

**File:** `scaffold/internal/ui/theme/theme.go` (lines 300-383)

Replace entire `init()` function with explicit 5-color definitions:

```go
func init() {
    // default — teal primary, purple secondary
    RegisterTheme(ThemeSpec{
        Name:      "default",
        Primary:   lipgloss.Color("#10B1AE"),
        Secondary: lipgloss.Color("#6B50FF"),
        Background: lipgloss.Color("#16161A"),
        Surface:    lipgloss.Color("#1A1A1F"),
        Foreground: lipgloss.Color("#F1EFEF"),
    })

    // ocean — blue primary, cyan secondary
    RegisterTheme(ThemeSpec{
        Name:      "ocean",
        Primary:   lipgloss.Color("#4A90D9"),
        Secondary: lipgloss.Color("#2BC4C4"),
        Background: lipgloss.Color("#0A1628"),
        Surface:    lipgloss.Color("#111D32"),
        Foreground: lipgloss.Color("#E8F4FD"),
    })

    // forest — green primary, amber secondary
    RegisterTheme(ThemeSpec{
        Name:      "forest",
        Primary:   lipgloss.Color("#4A7C59"),
        Secondary: lipgloss.Color("#C9913D"),
        Background: lipgloss.Color("#0F1A14"),
        Surface:    lipgloss.Color("#16241D"),
        Foreground: lipgloss.Color("#F0F7ED"),
    })

    // sunset — pink primary, purple secondary
    RegisterTheme(ThemeSpec{
        Name:      "sunset",
        Primary:   lipgloss.Color("#FF6B6B"),
        Secondary: lipgloss.Color("#5F4B8B"),
        Background: lipgloss.Color("#1F1419"),
        Surface:    lipgloss.Color("#2A1D23"),
        Foreground: lipgloss.Color("#FFF5F5"),
    })

    // aurora — purple primary, green secondary
    RegisterTheme(ThemeSpec{
        Name:      "aurora",
        Primary:   lipgloss.Color("#7F5AF0"),
        Secondary: lipgloss.Color("#2CB67D"),
        Background: lipgloss.Color("#141420"),
        Surface:    lipgloss.Color("#1E1D2A"),
        Foreground: lipgloss.Color("#F5F0FF"),
    })

    // ember — red primary, gold secondary (custom OnPrimary/OnSecondary)
    RegisterTheme(ThemeSpec{
        Name:      "ember",
        Primary:   lipgloss.Color("#8B1E3F"),
        Secondary: lipgloss.Color("#CFAE70"),
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
        Name:      "neon",
        Primary:   lipgloss.Color("#00F5D4"),
        Secondary: lipgloss.Color("#FF00C8"),
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
        Name:      "slate",
        Primary:   lipgloss.Color("#3A506B"),
        Secondary: lipgloss.Color("#1C7ED6"),
        Background: lipgloss.Color("#0F141A"),
        Surface:    lipgloss.Color("#182029"),
        Foreground: lipgloss.Color("#E8EDF5"),
    })

    // sakura — cherry blossom pink, lavender secondary
    RegisterTheme(ThemeSpec{
        Name:      "sakura",
        Primary:   lipgloss.Color("#E87EA1"),
        Secondary: lipgloss.Color("#9B72CF"),
        Background: lipgloss.Color("#1F151C"),
        Surface:    lipgloss.Color("#2C1E27"),
        Foreground: lipgloss.Color("#FFF5FA"),
    })

    // nord — arctic blue primary, frost cyan secondary
    RegisterTheme(ThemeSpec{
        Name:      "nord",
        Primary:   lipgloss.Color("#5E81AC"),
        Secondary: lipgloss.Color("#88C0D0"),
        Background: lipgloss.Color("#10171E"),
        Surface:    lipgloss.Color("#19222D"),
        Foreground: lipgloss.Color("#ECEFF4"),
    })

    // mono — monochrome minimal
    RegisterTheme(ThemeSpec{
        Name:      "mono",
        Primary:   lipgloss.Color("#787878"),
        Secondary: lipgloss.Color("#A8A8A8"),
        Background: lipgloss.Color("#121212"),
        Surface:    lipgloss.Color("#1C1C1C"),
        Foreground: lipgloss.Color("#E8E8E8"),
    })
}
```

### Step 7: Update `ValidatePalette()`

**File:** `scaffold/internal/ui/theme/theme.go` (lines 115-151)

Update for new field references and add status contrast checks:

```go
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
        on    color.Color
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
```

### Step 8: Update Style Providers

**File:** `scaffold/internal/ui/theme/theme.go`

**`newStylesFromPalette()` (lines 399-428):**
- Add `.Background(p.Background)` to App style
- Add `.Background(p.Surface)` to Header style
- `p.TextPrimary` → `p.Foreground`
- `p.TextMuted` → `p.BorderMuted`
- `p.SubtlePrimary` → `p.PrimaryMuted`
- `p.TextInverse` → `p.OnPrimary`

**`newDetailStylesFromPalette()` (lines 449-456):**
- `p.SubtleSecondary` → `p.SecondaryMuted`
- `p.TextPrimary` → `p.Foreground`

**`newModalStylesFromPalette()` (lines 477-488):**
- `p.TextPrimary` → `p.Foreground`
- `p.TextMuted` → `p.ForegroundSubtle`
- Add `.Background(p.SurfaceRaised)` to Dialog style

**`ListStyles()` (lines 515-534):**
- Remove `p.PrimaryHover` usage
- `p.TextSecondary` → `p.ForegroundMuted`
- `p.TextMuted` → `p.ForegroundSubtle`
- Keep `p.Primary` for key elements

**`ListItemStyles()` (lines 537-558):**
- `p.TextMuted` → `p.ForegroundSubtle`
- Remove `p.PrimaryHover` usage
- `p.SubtleSecondary` → `p.SecondaryMuted`

### Step 9: Update `HuhTheme()` Function

**File:** `scaffold/internal/ui/theme/huh.go`

Update all field references:
- `p.FocusBorder` → `p.Focus`
- `p.TextSecondary` → `p.ForegroundMuted`
- `p.TextMuted` → `p.ForegroundSubtle`
- `p.TextInverse` → `p.OnPrimary`
- `p.PrimaryHover` → `p.Primary`
- `p.SubtleSecondary` → `p.SecondaryMuted`
- `p.SubtlePrimary` → `p.PrimaryMuted`

### Step 10: Update `statusbar/statusbar.go`

**File:** `scaffold/internal/ui/statusbar/statusbar.go` (line 51)

```go
BorderForeground(p.ForegroundSubtle),  // was p.TextMuted
```

And line 54:
```go
m.rightSty = lipgloss.NewStyle().Foreground(p.ForegroundSubtle)  // was p.TextMuted
```

### Step 11: Update `status/styles.go`

**File:** `scaffold/internal/ui/status/styles.go` (line 23)

Use specific OnStatus colors for high contrast on each status color:

```go
Base:    base.Background(p.Primary).Foreground(p.OnPrimary),    // was p.TextInverse
Info:    base.Background(p.Info).Foreground(p.OnInfo),          // was p.TextInverse → use p.OnInfo
Success: base.Background(p.Success).Foreground(p.OnSuccess), // was p.TextInverse → use p.OnSuccess
Warning: base.Background(p.Warning).Foreground(p.OnWarning), // was p.TextInverse → use p.OnWarning
Error:   base.Background(p.Error).Foreground(p.OnError),   // was p.TextInverse → use p.OnError
```

## Files to Modify

| File | Changes |
|-------|----------|
| `scaffold/internal/ui/theme/theme.go` | Helper functions, Palette struct, ThemeSpec, buildPalette(), NewPalette(), ValidatePalette(), style providers, theme registry |
| `scaffold/internal/ui/theme/huh.go` | HuhTheme() field references |
| `scaffold/internal/ui/statusbar/statusbar.go` | TextMuted → ForegroundSubtle |
| `scaffold/internal/ui/status/styles.go` | TextInverse → OnPrimary |

## Verification

1. **Build check:**
   ```bash
   cd scaffold && go build ./...
   ```

2. **Run tests:**
   ```bash
   go test ./...
   ```

3. **Visual verification:**
   - Run application and cycle through all 10 themes
   - Verify dark/light mode switching
   - Check modal dialogs render correctly with new SurfaceRaised background
   - Verify status messages use correct OnPrimary text color
   - Test alpha-simulated colors (borders, muted text)

4. **Theme-specific test:**
   - `ember` theme should have custom OnPrimary/OnSecondary colors
   - `neon` theme should use magenta Focus and brighter status colors

## Notes

- **Backward compatibility:** The public API (State, Manager, ThemeAware) remains unchanged
- **5 explicit core colors:** Each theme defines Primary, Secondary, Background, Surface, Foreground
- **Computed colors:** All other semantic colors derived from core colors using the specified formulas
- **Status colors:** Default colors provided but can be overridden via Modify hook
- **Alpha simulation:** Since lipgloss doesn't support alpha, `withAlpha()` simulates transparency by desaturating and blending toward mid lightness

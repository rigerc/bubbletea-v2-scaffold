// ── Surfaces ──────────────────────────────────────
    Background        color.Color // page / app background
    Surface           color.Color // card, panel, sheet
    SurfaceRaised     color.Color // elevated surface (popover, dropdown)
    Overlay           color.Color // scrim behind modals (typically semi-transparent)

    // ── Borders ──────────────────────────────────────
    Border            color.Color // default border (dividers, inputs)
    BorderMuted       color.Color // subtle separators, hairlines

    // ── Content (text & icons) ───────────────────────
    Foreground        color.Color // primary text / icons on Background/Surface
    ForegroundMuted   color.Color // secondary / supporting text
    ForegroundSubtle  color.Color // placeholder, disabled-looking text

    // ── Brand / Interactive ──────────────────────────
    Primary           color.Color // primary action fill
    OnPrimary         color.Color // text/icon on Primary fill
    PrimaryMuted      color.Color // low-emphasis primary (selected row bg, tinted badge bg)

    Secondary         color.Color // secondary action fill
    OnSecondary       color.Color // text/icon on Secondary fill
    SecondaryMuted    color.Color // low-emphasis secondary

    // ── Interactive states ───────────────────────────
    Focus             color.Color // focus ring / outline


---
# Computed:

// ── From Background ──────────────────────────────
SurfaceRaised    = Lighten(Surface, 4%)        // or Darken in dark mode
Overlay          = WithAlpha(Foreground, 0.5)
Border           = WithAlpha(Foreground, 0.12)
BorderMuted      = WithAlpha(Foreground, 0.06)

// ── From Foreground ──────────────────────────────
ForegroundMuted  = WithAlpha(Foreground, 0.6)
ForegroundSubtle = WithAlpha(Foreground, 0.38)

// ── From Primary / Secondary ─────────────────────
OnPrimary        = ContrastingForeground(Primary)   // white or black
PrimaryMuted     = WithAlpha(Primary, 0.12)          // tinted background
OnSecondary      = ContrastingForeground(Secondary)
SecondaryMuted   = WithAlpha(Secondary, 0.12)

// ── From Status colors ───────────────────────────
OnError          = ContrastingForeground(Error)
Success          = DeriveFromHue(140, Background)    // or set manually
OnSuccess        = ContrastingForeground(Success)
Warning          = DeriveFromHue(38, Background)
OnWarning        = ContrastingForeground(Warning)
Info             = DeriveFromHue(210, Background)
OnInfo           = ContrastingForeground(Info)

// ── Interactive ──────────────────────────────────
Focus            = Primary                           // almost always right

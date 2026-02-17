# Theming/Styling Refactoring Plan

## Context

The current theming system in `template-v2-enhanced/internal/ui/` has several issues:
- **Hardcoded colors** scattered across files (`#25A065` in detail.go, huh/theme.go)
- **Incomplete theme**: Only 4 styles (App, Title, StatusMessage, Subtle) when screens need more
- **Dual implementations**: menu.go vs menu_huh.go, filepicker.go vs filepicker_huh.go
- **Inconsistent styling**: Some screens create styles directly instead of using theme

This refactoring creates a centralized `theme` package with semantic color names and comprehensive style definitions, eliminating all hardcoded colors and duplicate implementations.

## Design Approach

Inspired by production examples (bubbleMonitor-master, plural-main), the new system uses:

1. **Semantic Color Palette**: Primary, Secondary, Success, Warning, Alert, Text, Muted, Subtle, Border
2. **Adaptive Colors**: `lipgloss.LightDark()` for automatic light/dark mode
3. **Comprehensive Theme**: All UI element styles defined centrally
4. **Huh Integration**: Clean adapter pattern for form theming

## File Structure

```
internal/ui/
├── theme/                  # NEW - centralized theming package
│   ├── palette.go          # ThemePalette struct with predefined palettes
│   ├── theme.go            # Theme struct with all lipgloss styles
│   └── huh.go              # Huh theme adapter (moved from huh/theme.go)
├── screens/
│   ├── base.go             # UPDATE - use theme package
│   ├── detail.go           # UPDATE - remove hardcoded colors
│   ├── form.go             # UPDATE - use new theme adapter
│   ├── menu_huh.go         # UPDATE - use new theme adapter
│   ├── filepicker_huh.go   # UPDATE - use new theme adapter
│   └── settings.go         # UPDATE - use new theme adapter
└── (DELETE)
    ├── styles/             # DELETE - replaced by theme package
    ├── screens/menu.go     # DELETE - Huh-only approach
    ├── screens/filepicker.go # DELETE - Huh-only approach
    └── huh/theme.go        # DELETE - moved to theme/huh.go
```

## Implementation Steps

### Step 1: Create New Theme Package

**Create `theme/palette.go`:**
```go
// Package theme provides centralized theming for the application UI.
// It defines semantic color palettes and Lip Gloss styles that adapt
// to light and dark terminal backgrounds.
package theme

import (
    lipgloss "charm.land/lipgloss/v2"
    "image/color"
)

// ThemePalette defines semantic colors used throughout the application.
// Colors are adaptive and automatically adjust for light/dark backgrounds.
type ThemePalette struct {
    // Brand colors
    Primary   color.Color // Main brand color (green)
    PrimaryFg color.Color // Text color on primary background (#FFFDF5)

    // Accent colors
    Secondary color.Color // Secondary accent color (purple)

    // Semantic colors
    Success color.Color
    Warning color.Color
    Alert   color.Color

    // Text colors
    Text  color.Color // Primary text
    Muted color.Color // Secondary text
    Subtle color.Color // De-emphasized text

    // UI elements
    Border color.Color
}

// NewPalette creates a palette with adaptive colors using lipgloss.LightDark().
// The isDark parameter should come from tea.BackgroundColorMsg.IsDark().
func NewPalette(isDark bool) ThemePalette {
    ld := lipgloss.LightDark(isDark)
    return ThemePalette{
        // Brand colors - green theme
        Primary:   ld(lipgloss.Color("#04B575"), lipgloss.Color("#10CC85")),
        PrimaryFg: lipgloss.Color("#FFFDF5"), // Constant - white/cream text on green

        // Accent
        Secondary: ld(lipgloss.Color("#7D56F4"), lipgloss.Color("#9B7CFF")),

        // Semantic
        Success: ld(lipgloss.Color("#00CC66"), lipgloss.Color("#00FF9F")),
        Warning: ld(lipgloss.Color("#FFCC00"), lipgloss.Color("#FFD700")),
        Alert:   ld(lipgloss.Color("#FF5F87"), lipgloss.Color("#FF7AA3")),

        // Text
        Text:   ld(lipgloss.Color("#1A1A1A"), lipgloss.Color("#F0F0F0")),
        Muted:  ld(lipgloss.Color("#626262"), lipgloss.Color("#9B9B9B")),
        Subtle: ld(lipgloss.Color("#9B9B9B"), lipgloss.Color("#626262")),

        // UI
        Border: ld(lipgloss.Color("#D0D0D0"), lipgloss.Color("#3A3A3A")),
    }
}
```

**Create `theme/theme.go`:**
```go
package theme

import (
    lipgloss "charm.land/lipgloss/v2"
)

// Theme contains all lipgloss styles used throughout the application.
// The Palette field provides access to semantic colors for dynamic styling.
type Theme struct {
    Palette ThemePalette

    // Container styles
    App   lipgloss.Style // Outer container with margin
    Panel lipgloss.Style // Bordered panel for grouped content

    // Typography
    Title  lipgloss.Style // Header/title bar with primary background
    Status lipgloss.Style // Status/informational text
    Subtle lipgloss.Style // De-emphasized text

    // UI elements
    Border  lipgloss.Style // Horizontal dividers and borders
    Error   lipgloss.Style // Error messages
    Warning lipgloss.Style // Warning messages
    Success lipgloss.Style // Success messages
}

// New creates a Theme with adaptive colors for the given background.
// The isDark parameter should come from tea.BackgroundColorMsg.IsDark().
func New(isDark bool) Theme {
    p := NewPalette(isDark)
    return Theme{
        Palette: p,

        // Container styles
        App: lipgloss.NewStyle().Margin(1, 2),
        Panel: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(p.Primary).
            Padding(1),

        // Typography - ALL colors from palette
        Title: lipgloss.NewStyle().
            Bold(true).
            Foreground(p.PrimaryFg).  // From palette, not hardcoded
            Background(p.Primary).
            Padding(0, 1),
        Status: lipgloss.NewStyle().Foreground(p.Primary),
        Subtle: lipgloss.NewStyle().Foreground(p.Subtle),

        // UI elements - ALL colors from palette
        Border:  lipgloss.NewStyle().Foreground(p.Border),
        Error:   lipgloss.NewStyle().Foreground(p.Alert).Bold(true),
        Warning: lipgloss.NewStyle().Foreground(p.Warning).Bold(true),
        Success: lipgloss.NewStyle().Foreground(p.Success).Bold(true),
    }
}
```

**Create `theme/huh.go`:**
```go
// Package theme provides Huh form integration for the application theme.
package theme

import (
    "charm.land/huh/v2"
    lipgloss "charm.land/lipgloss/v2"
)

// HuhThemeFunc returns a Huh theme that matches the application's visual style.
// The theme function uses the isDark parameter provided by Huh, ensuring colors
// stay in sync even if the terminal background changes.
//
// CRITICAL: This function creates its own ThemePalette using Huh's isDark parameter
// rather than using a cached Theme. This ensures colors stay synchronized even if
// the terminal background changes after the application starts.
//
// Usage: form.WithTheme(theme.HuhThemeFunc())
func HuhThemeFunc() huh.Theme {
    return huh.ThemeFunc(func(isDark bool) *huh.Styles {
        // Create palette using Huh's isDark parameter (not from cached Theme)
        p := NewPalette(isDark)

        // Start with Charm's default theme as base
        s := huh.ThemeCharm(isDark)

        // Apply app's green theme to titles - ALL colors from palette
        s.Group.Title = s.Group.Title.
            Foreground(p.PrimaryFg).  // From palette, not hardcoded
            Background(p.Primary)     // From palette, not hardcoded

        // Match status messages to app theme
        s.Focused.Description = s.Focused.Description.
            Foreground(p.Primary)

        // Remove borders for cleaner look
        s.Form.Base = s.Form.Base.
            BorderTop(false).BorderRight(false).
            BorderBottom(false).BorderLeft(false)
        s.Group.Base = s.Group.Base.
            BorderTop(false).BorderRight(false).
            BorderBottom(false).BorderLeft(false)
        s.Focused.Base = s.Focused.Base.
            BorderTop(false).BorderRight(false).
            BorderBottom(false).BorderLeft(false)
        s.Blurred.Base = s.Blurred.Base.
            BorderTop(false).BorderRight(false).
            BorderBottom(false).BorderLeft(false)

        // Increase spacing between options
        s.Focused.Option = s.Focused.Option.Margin(2, 0)
        s.Blurred.Option = s.Blurred.Option.Margin(2, 0)

        return s
    })
}
```

**Note:** The Huh theme function creates its own `ThemePalette` using Huh's `isDark`
parameter rather than using a cached `Theme`. This is critical because:
1. Huh calls this function with its own detected background value
2. The terminal background may change after Theme creation
3. Creating the palette inside the closure ensures colors stay synchronized

### Step 2: Update ScreenBase

**Edit `screens/base.go`:**
- Change import from `internal/ui/styles` to `internal/ui/theme`
- Change `Theme` type from `styles.Theme` to `theme.Theme`
- Update `ApplyTheme()` to use `theme.New(isDark)`

### Step 3: Update Screens

**Edit `screens/detail.go`:**
- Change import to use `theme` package
- Replace hardcoded `#25A065` with `s.Theme.Palette.Primary`

**Edit `screens/form.go`:**
- Change import from `internal/ui/huh` to `internal/ui/theme`
- Update `form.WithTheme(huhadapter.ThemeFunc(s.Theme))` to `form.WithTheme(theme.HuhThemeFunc())`

**Edit `screens/menu_huh.go`, `screens/filepicker_huh.go`, `screens/settings.go`:**
- Update imports to use new theme package

### Step 4: Delete Deprecated Files

- Delete `internal/ui/styles/` directory
- Delete `internal/ui/screens/menu.go`
- Delete `internal/ui/screens/filepicker.go`
- Delete `internal/ui/huh/theme.go`

## Critical Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `theme/palette.go` | CREATE | Semantic color palette with adaptive colors |
| `theme/theme.go` | CREATE | Central theme with all lipgloss styles |
| `theme/huh.go` | CREATE | Huh form integration (uses isDark param directly) |
| `screens/base.go` | MODIFY | Change to use theme package |
| `screens/detail.go` | MODIFY | Remove hardcoded `#25A065` |
| `screens/form.go` | MODIFY | Update theme adapter usage |
| `screens/menu_huh.go` | MODIFY | Update to use theme package |
| `screens/settings.go` | MODIFY | Update to use theme package |
| `screens/filepicker_huh.go` | MODIFY | Update to use theme package |
| `styles/` | DELETE | Replaced by theme package |
| `screens/menu.go` | DELETE | Huh-only approach |
| `screens/filepicker.go` | DELETE | Huh-only approach |
| `huh/theme.go` | DELETE | Moved to theme/huh.go |

## Key Fixes from Code Review

The plan was reviewed by a BubbleTea development agent. Here are the issues that were identified and addressed:

### Round 1: Initial Fixes
1. **Fixed: Palette as single source of truth**
   - Added `PrimaryFg` field for title text color
   - All colors in `theme.go` reference palette fields
   - No hardcoded hex colors remain in style definitions

2. **Fixed: Huh theme `isDark` parameter handling**
   - `HuhThemeFunc()` creates its own `ThemePalette` using Huh's `isDark` parameter
   - All Huh theme colors reference palette fields (`p.Primary`, `p.PrimaryFg`)
   - Ensures synchronization even if terminal background changes

3. **Fixed: Added Go style guide documentation**
   - Package-level godoc for all new files
   - Struct documentation explaining purpose and usage
   - Function documentation with parameter descriptions

4. **Fixed: Consistent color definitions**
   - All brand colors defined in one place (ThemePalette)
   - Adaptive colors use `ld()` helper consistently
   - Constant colors clearly marked in comments

### Round 2: Additional Fixes (After Agent Review)
5. **Fixed: Huh theme now uses palette directly** (CRITICAL)
   - Previous version had hardcoded hex colors in `HuhThemeFunc()`
   - Now creates `p := NewPalette(isDark)` inside the closure
   - ALL Huh colors reference palette fields, no hex values remain

6. **Fixed: Added missing common styles** (MEDIUM)
   - Added `Error`, `Warning`, `Success` styles to Theme struct
   - These use `Alert`, `Warning`, `Success` palette colors
   - Makes semantic message styling readily available

7. **Fixed: Improved grep verification pattern** (LOW)
   - Changed from `#[0-9A-Fa-F]\{6\}` to `#[0-9A-Fa-f]{3,8}`
   - Now catches 3-char shorthand, 4-char ARGB, and 8-color AARRGGBB formats

8. **Note: Import alias consistency** (LOW)
   - Use `lipgloss "charm.land/lipgloss/v2"` consistently across all files

## Verification

1. **Build check**: `go build ./template-v2-enhanced/`
   - Verify no import errors

2. **Run the app**: `./template-v2-enhanced/template-v2-enhanced`
   - Navigate through all screens
   - Verify styling looks correct

3. **Test light/dark mode**:
   - Change terminal background color
   - Restart app
   - Verify colors adapt correctly

4. **Check for remaining hardcoded colors**:
   ```bash
   # Search for hex colors in theme package (should only be in palette.go)
   grep -rE "#[0-9A-Fa-f]{3,8}" template-v2-enhanced/internal/ui/theme/
   ```
   Expected: Only matches in `palette.go`

5. **Find all files that need import updates**:
   ```bash
   # Find files importing old packages
   grep -r "internal/ui/styles" template-v2-enhanced/
   grep -r "internal/ui/huh" template-v2-enhanced/
   ```

6. **Verify no hardcoded colors in screens**:
   ```bash
   # Should return no results after migration
   grep -rn "lipgloss.Color(\"#" template-v2-enhanced/internal/ui/screens/
   ```

7. **Test Huh forms**:
   - Navigate to Settings screen
   - Verify form styling matches app theme

## Migration Notes

- The refactoring maintains the existing green-based color scheme
- All existing functionality is preserved
- No API changes to screen interfaces
- Help styles continue to use `help.DefaultStyles(isDark)`

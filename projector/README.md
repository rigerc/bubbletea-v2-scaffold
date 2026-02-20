# projector

A production-ready BubbleTea v2 skeleton for building terminal user interface (TUI) applications in Go. Clone it, rename the module, and ship.

---

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Configuration](#configuration)
- [Architecture](#architecture)
- [Package Reference](#package-reference)
- [Built-in Screens](#built-in-screens)
- [Extending the Scaffold](#extending-the-scaffold)
- [Theming](#theming)
- [Logging](#logging)
- [CI / Release](#ci--release)

---

## Features

| Capability | Details |
|---|---|
| **BubbleTea v2** | Elm-architecture TUI with `tea.View`, `KeyPressMsg`, and per-frame rendering |
| **Stack-based navigation** | `nav.Push` / `nav.Pop` / `nav.Replace` router; no global state |
| **Adaptive theming** | Light/dark detection via `tea.BackgroundColorMsg`; every screen reacts |
| **Huh v2 forms** | Multi-page, dynamic forms with `WithHideFunc` / `OptionsFunc` / `TitleFunc` |
| **ASCII banner** | figlet-go rendering with 27 gradient presets and 15 safe fonts |
| **Cobra CLI** | `--config`, `--debug`, `--log-level` flags; `version` and `completion` subcommands |
| **Structured logging** | zerolog sink to `debug.log`; silent in normal mode so it never pollutes the TUI |
| **koanf configuration** | JSON file → CLI flags priority chain; embedded defaults |
| **GoReleaser** | Multi-platform binary releases with `CGO_ENABLED=0` |
| **GitHub Actions** | Build, lint, release, and Dependabot auto-merge workflows |

---

## Prerequisites

- Go 1.25+
- A terminal that supports ANSI color (virtually all modern terminals)

---

## Quick Start

```sh
# Clone and rename
git clone <repo-url> myapp
cd myapp
go mod edit -module myapp

# Run
go run . 

# Build
go build -o myapp .
```

### Shell Completions

```sh
# Bash
source <(myapp completion bash)

# Zsh
myapp completion zsh > "${fpath[1]}/_myapp"

# Fish
myapp completion fish | source

# PowerShell
myapp completion powershell | Out-String | Invoke-Expression
```

---

## Usage

```
scaffold [flags]
scaffold <command>

Flags:
  --config string      Path to configuration file (default: $HOME/.scaffold.json)
  --debug              Enable debug mode (forces log level to trace)
  --log-level string   Log verbosity: trace|debug|info|warn|error|fatal (default: info)

Commands:
  version              Print version
  completion           Generate shell completion script
  help                 Help for any command
```

### Examples

```sh
# Launch TUI with defaults
scaffold

# Use a custom config file
scaffold --config /etc/myapp/config.json

# Debug mode — logs written to debug.log
scaffold --debug

# Explicit log level (without forcing debug mode)
scaffold --log-level warn

# Print version only (TUI does not start)
scaffold version
```

> **Note:** Subcommands like `version` and `completion` set `runUI = false` before returning, so the TUI never starts for those invocations.

---

## Configuration

Configuration is resolved in priority order (highest wins):

```
defaults → config file (--config) → explicit CLI flags
```

### Schema

```json
{
  "logLevel": "info",
  "debug": false,
  "ui": {
    "altScreen": false,
    "mouseEnabled": true,
    "themeName": "default"
  },
  "app": {
    "name": "scaffold",
    "version": "1.0.0",
    "title": "Template V2 Enhanced"
  }
}
```

The embedded `assets/config.default.json` is used as a fallback when no config file is specified. Copy it to customize:

```sh
cp assets/config.default.json ~/.scaffold.json
```

### Fields

| Key | Type | Default | Description |
|---|---|---|---|
| `logLevel` | string | `"info"` | Minimum log level (`trace`…`fatal`) |
| `debug` | bool | `false` | When `true`, overrides `logLevel` to `trace` |
| `ui.altScreen` | bool | `false` | Run TUI in alternate screen buffer (fullscreen) |
| `ui.mouseEnabled` | bool | `true` | Enable mouse support (`CellMotion` mode) |
| `ui.themeName` | string | `"default"` | Reserved for future theme selection |
| `app.name` | string | `"scaffold"` | Shown in every screen's header badge |
| `app.title` | string | `"Template V2 Enhanced"` | Terminal window title |

---

## Architecture

```
main.go                   Entry point: Cobra → config → logger → ui.Run()
├── cmd/                  Cobra commands
│   ├── root.go           Root command, persistent flags, runUI gate
│   ├── version.go        `version` subcommand
│   └── completion.go     `completion` subcommand
├── config/               Configuration management
│   ├── config.go         Config struct, Load(), Validate(), koanf wiring
│   └── defaults.go       DefaultConfig(), DefaultConfigJSON()
├── assets/
│   └── config.default.json  Embedded default configuration
└── internal/
    ├── logger/           zerolog wrapper (global logger, convenience funcs)
    └── ui/               BubbleTea application
        ├── model.go      Root Model: navigation stack, theme detection, View()
        ├── nav/          Screen interface + Push/Pop/Replace messages
        ├── keys/         GlobalKeyMap (esc, ctrl+c, ?)
        ├── huh/          Huh keymap adapter (preserves global quit)
        ├── theme/        ThemePalette, Theme, HuhThemeFunc()
        └── screens/      Individual screen implementations
            ├── base.go             ScreenBase — shared state & layout helpers
            ├── form.go             FormScreen — huh.Form ↔ nav.Screen adapter
            ├── menu_huh.go         HuhMenuScreen — main menu with ASCII banner
            ├── detail.go           DetailScreen — scrollable viewport with gutter
            ├── settings.go         SettingsScreen — multi-page dynamic form demo
            ├── filepicker_huh.go   HuhFilePickerScreen — filesystem browser
            └── banner_demo.go      BannerDemoScreen — font & gradient showcase
```

### Data Flow

```
main()
  └─ cmd.Execute()            Cobra parses flags, sets runUI
  └─ loadConfig()             defaults → file → flags
  └─ initLogger()             zerolog → file (debug) or io.Discard (normal)
  └─ ui.Run(ui.New(cfg))      BubbleTea program starts
        │
        ├── Init()            tea.RequestBackgroundColor + root screen Init()
        │
        ├── Update(msg)
        │     ├── BackgroundColorMsg  → set isDark, call SetTheme on all screens
        │     ├── WindowSizeMsg       → store dims, delegate to active screen
        │     ├── PushMsg             → append screen to stack, init + size it
        │     ├── PopMsg              → pop stack, re-size newly exposed screen
        │     ├── ReplaceMsg          → swap top of stack
        │     └── *                   → delegate to screens[top]
        │
        └── View()            render screens[top].View(), set AltScreen/MouseMode/WindowTitle
```

---

## Package Reference

### `nav` — Navigation Router

The router lives entirely in messages; there is no singleton.

```go
// Screen is the interface every navigable view must implement.
type Screen interface {
    Init()   tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View()   string
}

// Themeable is optional — implement it to receive light/dark updates.
type Themeable interface {
    SetTheme(isDark bool)
}

// Navigation commands
nav.Push(screen)    // → PushMsg  — add screen on top of stack
nav.Pop()           // → PopMsg   — remove current screen
nav.Replace(screen) // → ReplaceMsg — swap current screen
```

All three helpers return `tea.Cmd`, so they compose naturally with `tea.Batch`.

---

### `screens.ScreenBase` — Shared Screen State

Embed `ScreenBase` in every new screen to avoid repeating layout and theming boilerplate.

```go
type MyScreen struct {
    screens.ScreenBase
    // your state here
}

func NewMyScreen(isDark bool, appName string) *MyScreen {
    return &MyScreen{
        ScreenBase: screens.NewBase(isDark, appName),
    }
}
```

**Provided helpers:**

| Method | Returns | Description |
|---|---|---|
| `ApplyTheme(isDark bool)` | — | Rebuild `Theme` and help styles |
| `ContentWidth() int` | int | Terminal width minus App container frame |
| `IsSized() bool` | bool | True after first `WindowSizeMsg` |
| `HeaderView() string` | string | App-name badge + horizontal rule |
| `RenderHelp(km) string` | string | Help bar from any `help.KeyMap` |
| `CalculateContentHeight(headerH, helpH int) int` | int | Capped content height respecting `MaxContentHeight` (25) and `MinContentHeight` (10) |

---

### `screens.FormScreen` — Huh Form Adapter

`FormScreen` wraps a `huh.Form` as a `nav.Screen`. Global keys take precedence over form keys.

```go
// Static form
fs := screens.NewFormScreen(form, isDark, appName, onSubmit, onAbort)

// Rebuilding form (e.g. menu that resets after navigation)
fs := screens.newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, maxContentHeight)
```

Key behaviour:
- `?` toggles help expansion
- `ESC` calls `onAbort`
- `ctrl+c` calls `tea.Quit`
- After `StateCompleted` or `StateAborted`, the form is automatically reset so it is reusable when navigating back

---

### `theme` — Adaptive Styling

```go
// Build a theme for the current terminal background
t := theme.New(isDark)

// Use semantic styles
t.Title.Render("My App")
t.Panel.Render(content)
t.Error.Render("something went wrong")

// Access raw palette colors
t.Palette.Primary   // brand green
t.Palette.Alert     // red/pink
t.Palette.Subtle    // de-emphasized text

// Apply to huh forms
form.WithTheme(theme.HuhThemeFunc())
```

`HuhThemeFunc()` returns a `huh.ThemeFunc` that re-evaluates `isDark` on every render, ensuring forms stay in sync even if the terminal background changes after startup.

---

### `banner` — ASCII Art Rendering

```go
cfg := banner.BannerConfig{
    Text:       "myapp",
    Font:       "slant",          // or "" for random
    Gradient:   &banner.GradientNeon,
    Background: true,             // fills background with gradient BG color
    Width:      90,               // override terminal width
}

rendered, err := banner.RenderBanner(cfg, terminalWidth)
```

**Convenience constructors:**

```go
banner.RandomBanner("myapp")                        // random safe font + gradient
banner.NamedBanner("myapp", "doom", "aurora")       // explicit font + gradient name
```

**Safe fonts** (render cleanly at 80–120 columns):
`slant`, `big`, `banner3`, `doom`, `epic`, `isometric1`, `larry3d`, `lean`, `ogre`, `roman`, `shadow`, `small`, `smslant`, `standard`, `straight`

**Gradient presets** (27 total): `sunset`, `ocean`, `forest`, `neon`, `aurora`, `fire`, `pastel`, `monochrome`, `vaporwave`, `matrix`, `mind`, `rainbow`, `galaxy`, `lunar`, `phoenix`, `spirit`, `cherry`, `waves`, `dreamy`, `magic`, `electric`, `venom`, `mirage`, `rebel`, `drift`, `bloom`, `atlas`

```go
// Look up by name
grad, ok := banner.GradientByName("aurora")

// All gradients
for _, g := range banner.AllGradients() { ... }

// Random selection
grad := banner.RandomGradient()
font := banner.RandomSafeFont()
```

---

### `logger` — Global zerolog Logger

During TUI mode the terminal is occupied, so all log output is redirected to `debug.log` (debug mode) or silenced entirely (normal mode).

```go
// Convenience functions (global logger)
logger.Trace().Msg("verbose trace")
logger.Debug().Str("key", val).Msg("debug")
logger.Info().Msg("started")
logger.Warn().Err(err).Msg("warning")
logger.Error().Err(err).Msg("error")
logger.Fatal().Err(err).Msg("fatal")   // calls os.Exit(1)

// With context fields
logger.With().Str("component", "ui").Logger()

// Dynamic level change
logger.SetLevel(logger.LevelDebug)
```

Log format is `console` by default; set `ENV=production` to switch to JSON.

---

## Built-in Screens

| Screen | Constructor | Description |
|---|---|---|
| `HuhMenuScreen` | `screens.NewHuhMenuScreen(options, isDark, appName)` | Main menu. Renders an ASCII banner above a `huh.Select`. ESC quits. |
| `DetailScreen` | `screens.NewDetailScreen(title, content, isDark, appName)` | Scrollable viewport with line-number gutter and scroll-percentage footer. |
| `SettingsScreen` | `screens.NewSettingsScreen(isDark, appName)` | 6-page dynamic form demo showing `WithHideFunc`, `OptionsFunc`, `TitleFunc`. |
| `HuhFilePickerScreen` | `screens.NewHuhFilePickerScreen(startDir, isDark, appName)` | Filesystem browser; pushes a `DetailScreen` with file content on selection. |
| `BannerDemoScreen` | `screens.NewBannerDemoScreen(isDark, appName)` | Scrollable showcase of all safe fonts and gradient presets. |

### Global Key Bindings

| Key | Action |
|---|---|
| `ESC` | Go back (pop screen / abort form) |
| `ctrl+c` | Quit application |
| `?` | Toggle help bar expansion |

---

## Extending the Scaffold

### 1. Rename the module

```sh
# Replace all occurrences of the module name
find . -type f -name "*.go" | xargs sed -i 's|scaffold|myapp|g'
go mod edit -module myapp
go mod tidy
```

### 2. Add a new screen

Create `internal/ui/screens/myscreeen.go`:

```go
package screens

import (
    tea "charm.land/bubbletea/v2"
    "scaffold/internal/ui/nav"
)

type MyScreen struct {
    ScreenBase
}

func NewMyScreen(isDark bool, appName string) *MyScreen {
    return &MyScreen{ScreenBase: NewBase(isDark, appName)}
}

func (s *MyScreen) Init() tea.Cmd { return nil }

func (s *MyScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        s.Width, s.Height = msg.Width, msg.Height
    case tea.KeyPressMsg:
        if msg.String() == "esc" {
            return s, nav.Pop()
        }
    }
    return s, nil
}

func (s *MyScreen) View() string {
    return s.Theme.App.Render(s.HeaderView() + "\nHello from MyScreen!")
}

// SetTheme implements nav.Themeable (optional but recommended).
func (s *MyScreen) SetTheme(isDark bool) {
    s.ApplyTheme(isDark)
}
```

### 3. Add the screen to the menu

In `internal/ui/model.go`, add a `HuhMenuOption`:

```go
menuOptions := []screens.HuhMenuOption{
    // ... existing options ...
    {
        Title:       "My Feature",
        Description: "Launch my custom screen",
        Action:      nav.Push(screens.NewMyScreen(false, cfg.App.Name)),
    },
}
```

### 4. Add a form screen

```go
func NewMyFormScreen(isDark bool, appName string) *FormScreen {
    name := ""

    formBuilder := func() *huh.Form {
        return huh.NewForm(
            huh.NewGroup(
                huh.NewInput().Title("Your name").Value(&name),
            ),
        )
    }

    onSubmit := func() tea.Cmd {
        return func() tea.Msg { return MyResultMsg{Name: name} }
    }

    onAbort := func() tea.Cmd { return nav.Pop() }

    return newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)
}
```

### 5. Add a Cobra subcommand

Create `cmd/mycommand.go`:

```go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "Does something without starting the TUI",
    PreRun: func(cmd *cobra.Command, args []string) {
        runUI = false // prevent TUI from launching
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("hello from mycommand")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

---

## Theming

### Color Palette

The palette is defined in `internal/ui/theme/palette.go` using `lipgloss.LightDark()`. Every color automatically adapts to the terminal background.

```go
// Customise the brand color
Primary: ld(lipgloss.Color("#04B575"), lipgloss.Color("#10CC85")),
```

Replace the hex values to apply your brand. Semantic roles:

| Token | Role |
|---|---|
| `Primary` | Brand color; used in titles, borders, status |
| `PrimaryFg` | Text color rendered on Primary backgrounds |
| `Secondary` | Accent / secondary interactive elements |
| `Success` / `Warning` / `Alert` | Semantic feedback |
| `Text` / `Muted` / `Subtle` | Text hierarchy |
| `Border` | Dividers and panel outlines |

### Huh Form Theme

`theme.HuhThemeFunc()` builds on top of `huh.ThemeCharm` and overrides the group title, description, and border visibility to match the app palette. Customise it in `internal/ui/theme/huh.go`.

---

## Logging

| Scenario | Behavior |
|---|---|
| Normal mode | All logs go to `io.Discard` — terminal is clean |
| `--debug` flag | Logs written to `debug.log` in CWD via `tea.LogToFile` |
| `ENV=production` | JSON format instead of colored console |

Tail logs while running:

```sh
# Terminal 1
scaffold --debug

# Terminal 2
tail -f debug.log
```

---

## CI / Release

### GitHub Actions Workflows

| Workflow | Trigger | Description |
|---|---|---|
| `build.yml` | push / PR | `go test -race -coverpkg=./...`; uploads coverage to Codecov |
| `lint.yml` | push / PR | `golangci-lint` with project `.golangci.yml` |
| `release.yml` | tag push | GoReleaser cross-platform binaries |
| `dependabot-sync.yml` | Dependabot PR | Auto-approve and squash-merge patch/minor updates |

### Creating a Release

```sh
git tag v1.2.3
git push origin v1.2.3
# GitHub Actions runs GoReleaser automatically
```

GoReleaser is configured in `.goreleaser.yaml`:
- `CGO_ENABLED=0` for fully static binaries
- Targets all first-class Go platforms
- Changelog grouped by `feat:` / `fix:` conventional commits

---

## Dependency Overview

| Package | Role |
|---|---|
| `charm.land/bubbletea/v2` | TUI event loop and Elm architecture |
| `charm.land/bubbles/v2` | Viewport, help, key binding components |
| `charm.land/lipgloss/v2` | Terminal styling and layout |
| `charm.land/huh/v2` | Interactive forms and prompts |
| `github.com/spf13/cobra` | CLI commands and flags |
| `github.com/knadh/koanf/v2` | Configuration loading (JSON + embedded defaults) |
| `github.com/rs/zerolog` | Structured, zero-allocation logging |
| `github.com/lsferreira42/figlet-go` | ASCII art banner rendering |

---

## License

See [LICENSE](LICENSE).
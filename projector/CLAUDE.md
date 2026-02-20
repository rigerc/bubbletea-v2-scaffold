# CLAUDE.md — bubbletea-v2-scaffold/scaffold

All code uses **v2 Charm libraries** with `charm.land` import paths.
Never use the old `github.com/charmbracelet/...` paths.

| Library | Import path | Purpose |
|---|---|---|
| BubbleTea v2 | `charm.land/bubbletea/v2` | TUI event loop, Elm architecture |
| Bubbles v2 | `charm.land/bubbles/v2` | Viewport, help, key bindings |
| Lip Gloss v2 | `charm.land/lipgloss/v2` | Terminal styling and layout |
| huh v2 | `charm.land/huh/v2` | Interactive forms and prompts |
| Cobra | `github.com/spf13/cobra` | CLI commands and flags |
| koanf v2 | `github.com/knadh/koanf/v2` | JSON config with priority merging |
| zerolog | `github.com/rs/zerolog` | Structured logging |
| figlet-go | `github.com/lsferreira42/figlet-go` | ASCII art banners |

---

## Critical v2 API Rules

These are the most common sources of bugs. Verify every time.

### Import paths

```go
// CORRECT
import tea "charm.land/bubbletea/v2"
import "charm.land/bubbles/v2/viewport"
import lipgloss "charm.land/lipgloss/v2"
import "charm.land/huh/v2"

// WRONG — do not use
import tea "github.com/charmbracelet/bubbletea"
```

### BubbleTea v2 breaking changes

| Concept | v1 (wrong) | v2 (correct) |
|---|---|---|
| Key press message type | `tea.KeyMsg` | `tea.KeyPressMsg` |
| Space bar | `case " ":` | `case "space":` |
| View return type | `string` | `tea.View` via `tea.NewView(s)` |
| Alt screen | `tea.WithAltScreen()` option | `view.AltScreen = true` in `View()` |
| Mouse mode | `tea.WithMouseCellMotion()` option | `view.MouseMode = tea.MouseModeCellMotion` |
| Program start | `p.Start()` | `p.Run()` |
| Component dimensions | `m.Width = 40` | `m.SetWidth(40)` |
| Spinner tick | `spinner.Tick()` pkg func | `model.Tick()` method |

### Light/dark theme — always required

Bubbles v2 components do **not** auto-detect terminal background. You must
request it and propagate `isDark` to every component manually:

```go
func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        isDark := msg.IsDark()
        m.list.Styles    = list.DefaultStyles(isDark)
        m.help.Styles    = help.DefaultStyles(isDark)
        m.textInput.SetStyles(textinput.DefaultStyles(isDark))
        m.styles         = newStyles(isDark)      // lipgloss styles
    }
    return m, nil
}
```

---

## Scaffold Architecture

### Navigation

The scaffold uses a **stack-based router** with no global state.
All navigation happens via BubbleTea messages:

```go
nav.Push(screen)    // add screen on top of stack  → PushMsg
nav.Pop()           // remove current screen        → PopMsg
nav.Replace(screen) // swap current screen          → ReplaceMsg
```

The `nav.Screen` interface every screen must implement:

```go
type Screen interface {
    Init()   tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)  // returns Screen, not tea.Model
    View()   string
}
```

Optionally implement `nav.Themeable` to receive light/dark updates:

```go
type Themeable interface {
    SetTheme(isDark bool)
}
```

### ScreenBase

Embed `screens.ScreenBase` in every new screen — it provides theme, dimensions,
global key bindings, header rendering, help bar, and content-height calculation:

```go
type MyScreen struct {
    screens.ScreenBase
    // screen-specific state
}

func NewMyScreen(isDark bool, appName string) *MyScreen {
    return &MyScreen{ScreenBase: screens.NewBase(isDark, appName)}
}

func (s *MyScreen) SetTheme(isDark bool) { s.ApplyTheme(isDark) }
```

### Forms (huh v2)

Use `FormScreen` or `newFormScreenWithBuilder` to wrap huh forms as nav screens.
Global keys (ESC, ctrl+c, ?) take precedence over form keys. Forms are
auto-reset after `StateCompleted` or `StateAborted` so they are reusable on
back-navigation.

```go
formBuilder := func() *huh.Form {
    return huh.NewForm(huh.NewGroup( /* fields */ ))
}
onSubmit := func() tea.Cmd { return func() tea.Msg { return MyResultMsg{} } }
onAbort  := func() tea.Cmd { return nav.Pop() }

fs := screens.newFormScreenWithBuilder(formBuilder, isDark, appName, onSubmit, onAbort, 0)
```

### Theming

All colors come from `internal/ui/theme`. Never hardcode hex values outside
`theme/palette.go`.

```go
t := theme.New(isDark)        // build theme for current terminal background
t.Title.Render("heading")     // use semantic styles
t.Palette.Primary             // access raw palette colors
form.WithTheme(theme.HuhThemeFunc())  // apply to huh forms
```

`HuhThemeFunc()` creates a `huh.ThemeFunc` that re-evaluates `isDark` on every
render — critical so forms stay in sync if the terminal background changes.

### Logging

The terminal is occupied during TUI mode. All logging must go to a file or
be silenced — never to stdout/stderr.

```go
// debug mode: tea.LogToFile("debug.log", "debug") in main.go
// normal mode: io.Discard

applogger.Info().Msg("started")
applogger.Debug().Str("key", val).Msg("detail")
applogger.Error().Err(err).Msg("failed")
```

### Configuration priority

```
defaults (config.DefaultConfig())
  → JSON file (--config flag or $HOME/.scaffold.json)
    → explicit CLI flags (only when Changed() == true)
```

---

## Adding a New Screen (checklist)

1. Create `internal/ui/screens/mysceen.go`
2. Embed `ScreenBase`, implement `nav.Screen` (`Init`, `Update`, `View`)
3. Implement `SetTheme(isDark bool)` to satisfy `nav.Themeable`
4. Handle `tea.WindowSizeMsg` to store `s.Width, s.Height`
5. Return `nav.Pop()` on ESC
6. Add a `HuhMenuOption` entry in `internal/ui/model.go`
7. `go build ./...` — confirm no errors

---

## Adding a Cobra Subcommand (checklist)

1. Create `cmd/mycommand.go`
2. Set `runUI = false` in `PreRun` if the subcommand should not start the TUI
3. Call `rootCmd.AddCommand(myCmd)` in `init()`
4. Add persistent flags with `rootCmd.PersistentFlags()` if shared,
   or local flags with `myCmd.Flags()` if subcommand-specific

---

## Code Style

Follow the Google Go Style Guide (load `go-styleguide` skill for details).

- `MixedCaps` for all identifiers — never `snake_case` or `ALL_CAPS` constants
- Error strings lowercase, no trailing punctuation
- Wrap errors with `%w`: `fmt.Errorf("loading config: %w", err)`
- Interfaces defined at the point of use, kept small (1–3 methods)
- No named return values except in short functions where they aid clarity
- `context.Context` as first parameter, named `ctx`
- Table-driven tests with `t.Run`

---

## Validation

Before finishing any change:

```sh
go build ./...          # must pass cleanly
go test ./...           # run all tests
go vet ./...            # no vet warnings
```

For the scaffold specifically:
```sh
cd scaffold
go build ./...
go test ./...
```

The CI pipeline runs `go test -race -coverpkg=./...` and `golangci-lint`.
Ensure your changes pass both locally before pushing.

---

## What NOT to do

- Do not use `github.com/charmbracelet/...` import paths anywhere
- Do not write to stdout/stderr from inside a running BubbleTea program
- Do not hardcode hex color values outside `internal/ui/theme/palette.go`
- Do not call `os.Exit` inside screens — return `tea.Quit` instead
- Do not add global mutable state; pass config/dependencies explicitly
- Do not simplify or remove existing code just to fix a lint warning;
  report the issue instead
- Do not commit secrets or API keys

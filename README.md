# bubbletea-v2-scaffold

A monorepo of **Claude Code skills** and a **production-ready BubbleTea v2 project scaffold** for building terminal user interface (TUI) applications in Go with the full Charm ecosystem.

---

## What's in this repo

| Path | Purpose |
|---|---|
| `.claude/skills/` | Claude Code skill definitions (MCP-served references) |
| `.claude/agents/bubbletea-dev.md` | Specialist agent for TUI development |
| `.claude/commands/bt-dev.md` | Slash command to load the TUI skill suite |
| `.claude/commands/new-app.md` | Slash command to create a new app from the scaffold |
| `scaffold/` | Production-ready BubbleTea v2 project template |

---

## Claude Skills

Skills are curated knowledge packages served to Claude Code via MCP. They give the AI accurate, up-to-date API references, patterns, and examples for libraries that change frequently or where hallucination risk is high.

### Available Skills

| Skill | Library | What it covers |
|---|---|---|
| `bubbletea-v2` | `charm.land/bubbletea/v2` | Elm architecture, Model/Update/View, async Cmds, `BackgroundColorMsg`, `tea.View`, mouse, keyboard, alternate screen |
| `bubbles-v2` | `charm.land/bubbles/v2` | Spinner, TextInput, Textarea, List, Table, Viewport, Progress, Help, FilePicker, Paginator, Timer, Stopwatch, KeyMap |
| `lipgloss-v2` | `charm.land/lipgloss/v2` | Style API, adaptive `LightDark` colors, borders, layout (`JoinHorizontal`/`JoinVertical`/`Place`), compositing, list/table/tree subpackages |
| `huh-v2` | `charm.land/huh/v2` | Forms, Groups, Input/Select/MultiSelect/Confirm/Note/FilePicker fields, dynamic `Func` variants, `WithHideFunc`, BubbleTea integration, themes |
| `cobra` | `github.com/spf13/cobra` | Commands, persistent flags, `PreRun`/`PostRun` hooks, shell completions, `ValidArgs` |
| `zerolog` | `github.com/rs/zerolog` | Zero-allocation JSON/console logging, log levels, context fields, hooks, `pkgerrors` stack traces |
| `go-styleguide` | — | Google Go Style Guide: naming, formatting, error handling, package design, testing patterns |

### Using Skills in Claude Code

Skills are loaded automatically by the `bubbletea-dev` agent or on demand via the `/bt-dev` command:

```
/bt-dev
```

This loads the core TUI skill suite (`go-styleguide`, `bubbletea-v2`, `bubbles-v2`, `lipgloss-v2`, `huh-v2`). If your work also involves configuration, CLI flags, or logging, the agent will additionally load `koanf`, `cobra`, and `zerolog` skills as needed.

### Slash Commands

| Command | Usage | Description |
|---|---|---|
| `/bt-dev` | `/bt-dev` | Load the full Charm TUI skill suite into context |
| `/new-app` | `/new-app <app-name>` | Create a new app from the scaffold (see below) |

### The `bubbletea-dev` Agent

Located at `.claude/agents/bubbletea-dev.md`, this specialist agent activates when you ask Claude to build TUI apps, terminal interfaces, or anything Charm-related. It:

- Enforces correct v2 import paths (`charm.land/...` not `github.com/charmbracelet/...`)
- Handles light/dark theme detection via `tea.BackgroundColorMsg`
- Applies Bubbles v2 component patterns (method-based setters, `isDark` parameters)
- Follows the Google Go Style Guide throughout

---

## The Scaffold

`scaffold/` is a batteries-included starting point for a production TUI application. Copy it, rename the module, and ship.

### Technology Stack

| Library | Version | Role |
|---|---|---|
| [BubbleTea v2](https://charm.land/bubbletea/v2) | v2.0.0-rc.2 | TUI event loop, Elm architecture |
| [Bubbles v2](https://charm.land/bubbles/v2) | v2.0.0-rc.1 | Viewport, help bar, key bindings |
| [Lip Gloss v2](https://charm.land/lipgloss/v2) | v2.0.0-beta.3 | Terminal styling and layout |
| [huh v2](https://charm.land/huh/v2) | v2.0.0 | Interactive forms and prompts |
| [Cobra](https://github.com/spf13/cobra) | v1.9.1 | CLI commands and flags |
| [koanf v2](https://github.com/knadh/koanf) | v2.1.2 | JSON configuration with priority merging |
| [zerolog](https://github.com/rs/zerolog) | v1.33.0 | Structured logging (file sink in TUI mode) |
| [figlet-go](https://github.com/lsferreira42/figlet-go) | v0.0.2-beta | ASCII art banners with gradient color |

### Scaffold Features

- **Stack-based navigation** — `nav.Push` / `nav.Pop` / `nav.Replace` with no global state
- **Adaptive theming** — light/dark palette via `tea.BackgroundColorMsg`; all screens react
- **`ScreenBase` embed** — shared header, help bar, and content-height calculation for every screen
- **`FormScreen` adapter** — bridges `huh.Form` to `nav.Screen` with global key precedence and auto-reset
- **ASCII banner** — 27 gradient presets, 15 curated safe fonts, background fill support
- **Cobra CLI** — `--config`, `--debug`, `--log-level` flags; `version` + `completion` subcommands
- **Logging** — zerolog to `debug.log` in debug mode; `io.Discard` in normal mode (TUI-safe)
- **koanf config** — `defaults → JSON file → CLI flags` priority chain with embedded defaults
- **GoReleaser** — cross-platform static binaries (`CGO_ENABLED=0`)
- **GitHub Actions** — build/test (race detector + coverage), lint, release, Dependabot auto-merge

### Quick Start

The easiest way to create a new app is with the `/new-app` Claude Code command:

```
/new-app myapp
```

This copies the scaffold, renames the Go module and all internal imports, updates
the CLI binary name, config defaults, and README, then validates the result with
`go build`. The new app is created as `myapp/` next to `scaffold/`.

Alternatively, do it manually:

```sh
# Copy the scaffold
cp -r scaffold myapp
cd myapp

# Rename the module
go mod edit -module myapp
find . -type f -name "*.go" | xargs sed -i 's|scaffold|myapp|g'
go mod tidy

# Run
go run .
```

See [`scaffold/README.md`](scaffold/README.md) for the full developer guide, including how to add screens, forms, Cobra subcommands, and custom themes.

### Scaffold Layout

```
scaffold/
├── main.go                    Entry point: Cobra → config → logger → ui.Run()
├── cmd/                       Cobra CLI commands
│   ├── root.go                Root command, persistent flags, runUI gate
│   ├── version.go             `version` subcommand
│   └── completion.go          `completion` subcommand (bash/zsh/fish/powershell)
├── config/                    Configuration management
│   ├── config.go              Config struct, Load(), Validate(), koanf wiring
│   └── defaults.go            DefaultConfig()
├── assets/
│   └── config.default.json    Embedded default configuration
└── internal/
    ├── logger/                zerolog global logger with convenience functions
    └── ui/
        ├── model.go           Root BubbleTea model: nav stack, theme, View()
        ├── nav/               Screen interface + Push/Pop/Replace nav messages
        ├── keys/              GlobalKeyMap (esc, ctrl+c, ?)
        ├── huh/               Huh keymap adapter
        ├── theme/             ThemePalette, Theme styles, HuhThemeFunc()
        └── screens/
            ├── base.go        ScreenBase — shared state & layout helpers
            ├── form.go        FormScreen — huh.Form ↔ nav.Screen bridge
            ├── menu_huh.go    Main menu with ASCII banner
            ├── detail.go      Scrollable viewport with line-number gutter
            ├── settings.go    Multi-page dynamic form demo
            ├── filepicker_huh.go  Filesystem browser → DetailScreen
            └── banner_demo.go Font & gradient showcase
```

---

## Repository Structure

```
bubbletea-v2-scaffold/
├── .claude/
│   ├── agents/
│   │   └── bubbletea-dev.md   Specialist TUI agent definition
│   ├── commands/
│   │   ├── bt-dev.md          /bt-dev — load TUI skill suite
│   │   └── new-app.md         /new-app <name> — create app from scaffold
│   ├── plans/                 Architecture planning documents
│   └── skills/                Skill definitions (MCP-served)
├── scaffold/                  BubbleTea v2 project scaffold
│   └── README.md              Full scaffold developer guide
├── AGENTS.md                  AI agent engineering rules for this repo
└── README.md                  This file
```

---

## Key Conventions

### v2 Import Paths

All Charm libraries in this repo use the new `charm.land` paths. **Do not use the old `github.com/charmbracelet/...` paths.**

| ❌ v1 (old) | ✅ v2 (correct) |
|---|---|
| `github.com/charmbracelet/bubbletea` | `charm.land/bubbletea/v2` |
| `github.com/charmbracelet/bubbles` | `charm.land/bubbles/v2` |
| `github.com/charmbracelet/lipgloss` | `charm.land/lipgloss/v2` |
| `github.com/charmbracelet/huh` | `charm.land/huh/v2` |

### Critical v2 API Differences

| Concept | v1 | v2 |
|---|---|---|
| Key press message | `tea.KeyMsg` | `tea.KeyPressMsg` |
| View return type | `string` | `tea.View` via `tea.NewView(s)` |
| Alt screen | `tea.WithAltScreen()` option | `view.AltScreen = true` in `View()` |
| Mouse mode | `tea.WithMouseCellMotion()` option | `view.MouseMode = tea.MouseModeCellMotion` in `View()` |
| Program start | `p.Start()` | `p.Run()` |
| Component width/height | `m.Width = 40` | `m.SetWidth(40)` |
| Spinner tick | `spinner.Tick()` (package func) | `m.Tick()` (method) |

### Light/Dark Theme Handling

Bubbles v2 components do not auto-detect terminal background. Always request it explicitly and propagate to all components:

```go
func (m model) Init() tea.Cmd {
    return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.BackgroundColorMsg:
        isDark := msg.IsDark()
        m.list.Styles = list.DefaultStyles(isDark)
        m.help.Styles = help.DefaultStyles(isDark)
        m.styles = newStyles(isDark)
    }
    return m, nil
}
```

---

## License

[Apache 2.0](LICENSE)

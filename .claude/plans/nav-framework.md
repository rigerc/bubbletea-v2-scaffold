# Plan: BubbleTea v2 Navigation/Routing Framework

## Context

`template-v2-enhanced/` has a working skeleton with logging (zerolog), CLI (cobra),
config (koanf), and **config now wired to the UI** (done in previous session — see
`main.go:loadConfig()`, `cmd/root.go:WasLogLevelSet()`). The `internal/ui/model.go`
is a 4-field skeleton. This plan adds a stack-based navigation router driven by
`charm.land/bubbles/v2/list` with adaptive lipgloss-v2 styling and the bubbles
`help`/`key` system for keybindings.

## Current State (Already Done — Do Not Re-implement)

| What | Where | Status |
|---|---|---|
| `loadConfig()` with proper CLI flag priority | `main.go` | ✓ done |
| `WasLogLevelSet()` | `cmd/root.go` | ✓ done |
| `ui.New(cfg config.Config)` signature | `internal/ui/model.go` | ✓ done (value type, not pointer) |
| `v.AltScreen`, `v.MouseMode` in `View()` | `internal/ui/model.go` | ✓ done |
| `charm.land/bubbles/v2` in go.mod | `go.mod` | ✓ present (indirect only) |

---

## Dependencies

`charm.land/lipgloss/v2` is not yet in `go.mod`. `charm.land/bubbles/v2` is indirect.
Run once before implementing:
```sh
cd template-v2-enhanced
go get charm.land/bubbles/v2 charm.land/lipgloss/v2
```

This promotes `bubbles/v2` to direct and adds `lipgloss/v2`.

---

## File Structure

```
internal/ui/
├── model.go              [REWRITE]  — Root router; owns stack, help, AltScreen
├── nav/
│   └── nav.go           [CREATE]   — Screen interface + navigation messages
├── styles/
│   └── styles.go        [CREATE]   — Adaptive lipgloss styles
├── keys/
│   └── keys.go          [CREATE]   — Global key bindings
└── screens/
    ├── menu.go          [CREATE]   — List-based navigation menu
    └── detail.go        [CREATE]   — Example scrollable content screen
```

**Build order:** `keys` → `nav` → `styles` → `screens/detail` → `screens/menu` → `model`

---

## File 1: `internal/ui/nav/nav.go`

**Package:** `nav`

```go
// Screen is implemented by every navigable view.
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)  // returns Screen, not tea.Model
    View() string                       // string, not tea.View — Model.View() builds the final View
}

// Themeable is optional. Router calls SetTheme on push and on BackgroundColorMsg.
type Themeable interface {
    SetTheme(isDark bool)
}

// ScreenKeyMap is optional. Router uses it to show screen-specific help keys.
type ScreenKeyMap interface {
    HelpKeys() []key.Binding
}

// Navigation message types:
type PushMsg    struct{ Screen Screen }
type PopMsg     struct{}
type ReplaceMsg struct{ Screen Screen }

// Push returns a Cmd that sends PushMsg.
func Push(s Screen) tea.Cmd { return func() tea.Msg { return PushMsg{Screen: s} } }

// Pop returns a Cmd that sends PopMsg.
func Pop() tea.Cmd { return func() tea.Msg { return PopMsg{} } }

// Replace returns a Cmd that sends ReplaceMsg.
func Replace(s Screen) tea.Cmd { return func() tea.Msg { return ReplaceMsg{Screen: s} } }
```

---

## File 2: `internal/ui/keys/keys.go`

**Package:** `keys`

Implements `help.KeyMap` so it can be passed directly to `help.View()`.

```go
type GlobalKeyMap struct {
    Back key.Binding  // "esc"    — go to previous screen
    Quit key.Binding  // "ctrl+c" — always quit (no conflict with list filter "q")
    Help key.Binding  // "?"      — toggle help expansion
}

func New() GlobalKeyMap

// Implements help.KeyMap:
func (k GlobalKeyMap) ShortHelp() []key.Binding  { return []key.Binding{k.Back, k.Quit} }
func (k GlobalKeyMap) FullHelp() [][]key.Binding { return [][]key.Binding{{k.Back, k.Help}, {k.Quit}} }
```

**Critical:** `"q"` is intentionally NOT in `Quit`. The list uses `"q"` for filtering.
Only `ctrl+c` is the guaranteed quit path.

---

## File 3: `internal/ui/styles/styles.go`

**Package:** `styles`

```go
type Theme struct {
    App           lipgloss.Style   // outer container: Margin(1, 2)
    Title         lipgloss.Style   // list title bar (overrides list default)
    StatusMessage lipgloss.Style   // list status bar message
    HelpBar       lipgloss.Style   // help bar at bottom
    Detail        lipgloss.Style   // body text in detail screens
    Subtle        lipgloss.Style   // de-emphasized text
}

// New constructs a Theme using lipgloss.LightDark(isDark) for adaptive colors.
func New(isDark bool) Theme {
    ld := lipgloss.LightDark(isDark)
    return Theme{
        App:   lipgloss.NewStyle().Margin(1, 2),
        Title: lipgloss.NewStyle().Bold(true).
                   Foreground(lipgloss.Color("#FFFDF5")).
                   Background(lipgloss.Color("#25A065")).
                   Padding(0, 1),
        StatusMessage: lipgloss.NewStyle().
                   Foreground(ld(lipgloss.Color("#04B575"), lipgloss.Color("#10CC85"))),
        HelpBar: lipgloss.NewStyle().
                   Foreground(ld(lipgloss.Color("#626262"), lipgloss.Color("#9B9B9B"))),
        Detail:  lipgloss.NewStyle().Margin(0, 2),
        Subtle:  lipgloss.NewStyle().Foreground(ld(lipgloss.Color("#9B9B9B"), lipgloss.Color("#626262"))),
    }
}

// HelpBarHeight returns lines consumed by the help bar (1 short, 3 full).
func HelpBarHeight(showAll bool) int {
    if showAll { return 3 }
    return 1
}
```

---

## File 4: `internal/ui/screens/menu.go`

**Package:** `screens`

### MenuItem

```go
type MenuItem struct {
    title, description string
    action             tea.Cmd  // typically nav.Push(NewDetailScreen(...))
}

func (i MenuItem) Title() string       { return i.title }
func (i MenuItem) Description() string { return i.description }
func (i MenuItem) FilterValue() string { return i.title }

func NewMenuItem(title, description string, action tea.Cmd) MenuItem
```

### MenuScreen struct

```go
type MenuScreen struct {
    list         list.Model
    delegateKeys delegateKeyMap
    keys         keys.GlobalKeyMap
    theme        styles.Theme
    isDark       bool
    width, height int
    helpExpanded  bool
}
```

### Critical initialization in NewMenuScreen

```go
func NewMenuScreen(title string, items []list.Item, isDark bool) *MenuScreen {
    dKeys := newDelegateKeyMap()
    d     := newMenuDelegate(dKeys, isDark)

    l := list.New(items, d, 0, 0)   // 0,0: WindowSizeMsg drives size
    l.Title = title
    l.Styles = list.DefaultStyles(isDark)
    l.Styles.Title = theme.Title          // branded title override
    l.DisableQuitKeybindings()            // REQUIRED: prevents list eating ctrl+c/q
    l.SetShowHelp(false)                  // we render help in Model, not list
    // ...
}
```

### MenuScreen.Update — ESC/filter conflict

```go
func (s *MenuScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.WindowSizeMsg:
        s.width, s.height = msg.Width, msg.Height
        s.updateListSize()

    case tea.BackgroundColorMsg:
        s.isDark = msg.IsDark()
        s.theme = styles.New(s.isDark)
        s.list.Styles = list.DefaultStyles(s.isDark)
        s.list.Styles.Title = s.theme.Title
        d := list.NewDefaultDelegate()
        d.Styles = list.NewDefaultItemStyles(s.isDark)
        d.UpdateFunc = ...   // same delegate logic, new styles
        s.list.SetDelegate(d)
        return s, nil

    case tea.KeyPressMsg:
        // ESC pops stack ONLY when not filtering
        if msg.String() == "esc" && s.list.FilterState() == list.Unfiltered {
            return s, nav.Pop()
        }
    }

    var cmd tea.Cmd
    s.list, cmd = s.list.Update(msg)
    return s, cmd
}
```

### updateListSize — dynamic height

```go
func (s *MenuScreen) updateListSize() {
    frameH, frameV := s.theme.App.GetFrameSize()
    helpH := styles.HelpBarHeight(s.helpExpanded)
    s.list.SetSize(s.width-frameH, s.height-frameV-helpH)
}
```

Called on every `WindowSizeMsg` and whenever `helpExpanded` toggles.

### Delegate

```go
func newMenuDelegate(dKeys delegateKeyMap, isDark bool) list.DefaultDelegate {
    d := list.NewDefaultDelegate()
    d.Styles = list.NewDefaultItemStyles(isDark)
    d.UpdateFunc = func(msg tea.Msg, m *list.Model) tea.Cmd {
        if kp, ok := msg.(tea.KeyPressMsg); ok {
            if key.Matches(kp, dKeys.choose) {
                if item, ok := m.SelectedItem().(MenuItem); ok && item.action != nil {
                    return item.action
                }
            }
        }
        return nil
    }
    d.ShortHelpFunc = func() []key.Binding { return []key.Binding{dKeys.choose} }
    d.FullHelpFunc  = func() [][]key.Binding { return [][]key.Binding{{dKeys.choose}} }
    return d
}
```

---

## File 5: `internal/ui/screens/detail.go`

**Package:** `screens`

```go
type DetailScreen struct {
    title, content string
    keys           keys.GlobalKeyMap
    theme          styles.Theme
    isDark         bool
    width, height  int
    vp             viewport.Model
    ready          bool  // false until first WindowSizeMsg
}

func NewDetailScreen(title, content string, isDark bool) *DetailScreen

func (s *DetailScreen) Init() tea.Cmd  // nil
func (s *DetailScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd)
func (s *DetailScreen) View() string

// Implements nav.Themeable:
func (s *DetailScreen) SetTheme(isDark bool)

func (s *DetailScreen) SetContent(content string)  // update viewport content
func (s *DetailScreen) headerView() string         // private: renders title bar
```

### WindowSizeMsg handling

```go
case tea.WindowSizeMsg:
    s.width, s.height = msg.Width, msg.Height
    headerH := lipgloss.Height(s.headerView())
    frameH, frameV := s.theme.App.GetFrameSize()
    helpH  := styles.HelpBarHeight(false)
    s.vp.SetWidth(s.width - frameH)
    s.vp.SetHeight(s.height - frameV - headerH - helpH)
    if !s.ready {
        s.vp.SetContent(s.content)
        s.ready = true
    }

case tea.KeyPressMsg:
    if key.Matches(msg, s.keys.Back) {
        return s, nav.Pop()
    }
```

Viewport receives remaining messages for scroll (j/k, PageUp/Down, etc.).

Viewport constructor (v2 changed from v1):
```go
s.vp = viewport.New()   // no positional args in v2
s.vp.MouseWheelEnabled = true
```

---

## File 6: `internal/ui/model.go` (Rewrite)

**Package:** `ui`

**Note on signature:** Current `New(cfg config.Config)` takes a **value type** (not `*config.Config`
as in the original plan). Keep the value type. In `main.go`, the call is already `ui.New(*cfg)`.

### Model struct

```go
type Model struct {
    screens      []nav.Screen
    width, height int
    isDark        bool
    quitting      bool
    help          help.Model
    keys          keys.GlobalKeyMap
    helpExpanded  bool
    // from config (extracted from config.Config at construction):
    altScreen    bool   // cfg.UI.AltScreen
    mouseEnabled bool   // cfg.UI.MouseEnabled
    windowTitle  string // cfg.App.Title
}
```

### New()

```go
// New accepts config.Config (value type — main.go passes *cfg dereferenced).
func New(cfg config.Config) Model {
    globalKeys := keys.New()

    items := []list.Item{
        screens.NewMenuItem("Details", "View a detail screen",
            nav.Push(screens.NewDetailScreen("Details", detailContent, false))),
        screens.NewMenuItem("About", "About this application",
            nav.Push(screens.NewDetailScreen("About", aboutContent, false))),
    }
    root := screens.NewMenuScreen(cfg.App.Title, items, false)

    h := help.New()
    h.Styles = help.DefaultStyles(false)  // corrected on BackgroundColorMsg

    return Model{
        screens:      []nav.Screen{root},
        help:         h,
        keys:         globalKeys,
        altScreen:    cfg.UI.AltScreen,
        mouseEnabled: cfg.UI.MouseEnabled,
        windowTitle:  cfg.App.Title,
    }
}
```

### Init()

```go
func (m Model) Init() tea.Cmd {
    cmds := []tea.Cmd{tea.RequestBackgroundColor}
    if len(m.screens) > 0 {
        cmds = append(cmds, m.screens[len(m.screens)-1].Init())
    }
    return tea.Batch(cmds...)
}
```

### Update() — routing logic

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {

    case tea.KeyPressMsg:
        if key.Matches(msg, m.keys.Quit) {
            m.quitting = true
            return m, tea.Quit
        }
        if key.Matches(msg, m.keys.Help) {
            m.helpExpanded = !m.helpExpanded
            m.help.ShowAll = m.helpExpanded
            m.notifyActiveScreenResize()
            return m, nil
        }

    case tea.WindowSizeMsg:
        m.width, m.height = msg.Width, msg.Height
        m.help.Width = msg.Width   // Width is a field, not a method
        // fall through to delegate to active screen

    case tea.BackgroundColorMsg:
        m.isDark = msg.IsDark()
        m.help.Styles = help.DefaultStyles(m.isDark)
        // Propagate theme to ALL screens in stack
        for i := range m.screens {
            if t, ok := m.screens[i].(nav.Themeable); ok {
                t.SetTheme(m.isDark)
            }
        }
        // fall through to deliver msg to active screen

    case nav.PushMsg:
        s := msg.Screen
        if cmd := s.Init(); cmd != nil { cmds = append(cmds, cmd) }
        if t, ok := s.(nav.Themeable); ok { t.SetTheme(m.isDark) }
        s, cmd := s.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
        cmds = append(cmds, cmd)
        m.screens = append(m.screens, s)
        return m, tea.Batch(cmds...)

    case nav.PopMsg:
        if len(m.screens) > 1 {
            m.screens = m.screens[:len(m.screens)-1]
        }
        return m, nil

    case nav.ReplaceMsg:
        if len(m.screens) > 0 {
            s := msg.Screen
            if cmd := s.Init(); cmd != nil { cmds = append(cmds, cmd) }
            if t, ok := s.(nav.Themeable); ok { t.SetTheme(m.isDark) }
            m.screens[len(m.screens)-1] = s
        }
        return m, tea.Batch(cmds...)
    }

    // Delegate to active screen
    if len(m.screens) > 0 {
        top := m.screens[len(m.screens)-1]
        updated, cmd := top.Update(msg)
        m.screens[len(m.screens)-1] = updated
        cmds = append(cmds, cmd)
    }

    return m, tea.Batch(cmds...)
}
```

### View()

```go
func (m Model) View() tea.View {
    if m.quitting {
        return tea.NewView("")
    }

    var screenContent string
    if len(m.screens) > 0 {
        screenContent = m.screens[len(m.screens)-1].View()
    }

    helpView := m.help.View(m.buildKeyMap())
    content  := lipgloss.JoinVertical(lipgloss.Left, screenContent, helpView)

    v := tea.NewView(content)
    v.AltScreen   = m.altScreen     // from cfg.UI.AltScreen
    v.WindowTitle = m.windowTitle   // from cfg.App.Title
    if m.mouseEnabled {             // from cfg.UI.MouseEnabled
        v.MouseMode = tea.MouseModeCellMotion
    }
    return v
}

// buildKeyMap merges global keys + active screen's extra bindings for the help bar.
func (m Model) buildKeyMap() help.KeyMap {
    var extra []key.Binding
    if len(m.screens) > 0 {
        if sk, ok := m.screens[len(m.screens)-1].(nav.ScreenKeyMap); ok {
            extra = sk.HelpKeys()
        }
    }
    return combinedKeyMap{global: m.keys, extra: extra}
}

// notifyActiveScreenResize re-sends WindowSizeMsg so active screen recalculates
// height after help bar height changes (? toggle).
func (m *Model) notifyActiveScreenResize() {
    if len(m.screens) == 0 { return }
    top := m.screens[len(m.screens)-1]
    updated, _ := top.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
    m.screens[len(m.screens)-1] = updated
}

// combinedKeyMap merges GlobalKeyMap + active screen's extra bindings.
type combinedKeyMap struct {
    global keys.GlobalKeyMap
    extra  []key.Binding
}
func (c combinedKeyMap) ShortHelp() []key.Binding  { return append(c.global.ShortHelp(), c.extra...) }
func (c combinedKeyMap) FullHelp() [][]key.Binding {
    full := c.global.FullHelp()
    if len(c.extra) > 0 { full = append(full, c.extra) }
    return full
}
```

---

## Files to Create/Modify

| File | Action | Notes |
|---|---|---|
| `go.mod` / `go.sum` | MODIFY | `go get charm.land/bubbles/v2 charm.land/lipgloss/v2` |
| `internal/ui/model.go` | REWRITE | Stack router replaces skeleton |
| `internal/ui/nav/nav.go` | CREATE | Screen interface + nav messages |
| `internal/ui/styles/styles.go` | CREATE | Adaptive lipgloss theme |
| `internal/ui/keys/keys.go` | CREATE | Global key bindings (help.KeyMap) |
| `internal/ui/screens/menu.go` | CREATE | List-based navigation menu |
| `internal/ui/screens/detail.go` | CREATE | Scrollable text detail screen |
| `main.go` | **UNCHANGED** | Already complete with loadConfig() |
| `cmd/root.go` | **UNCHANGED** | Already has WasLogLevelSet() |
| `config/` | **UNCHANGED** | Already complete |

---

## Critical Implementation Notes

| Issue | Resolution |
|---|---|
| `list.DisableQuitKeybindings()` | MUST be called or list intercepts `ctrl+c`/`q` |
| `l.SetShowHelp(false)` | MUST be set — we render help in Model, not list |
| `list.New(items, d, 0, 0)` | Always 0,0 — `WindowSizeMsg` drives size via `updateListSize()` |
| `list.SetItems` returns `tea.Cmd` | Never ignore — contains status message animation |
| `help.Width` is a field, not a method | `m.help.Width = msg.Width`, not `m.help.SetWidth()` |
| `list.DefaultStyles(isDark bool)` | Must pass correct `isDark`; `false` is OK as initial |
| `list.NewDefaultItemStyles(isDark)` | Same — update in `SetTheme()` by recreating delegate |
| ESC in menu | Check `list.FilterState() == list.Unfiltered` before calling `nav.Pop()` |
| Theme on new screens | Inject via `nav.Themeable.SetTheme()` + send `WindowSizeMsg` on push |
| Viewport in v2 | `viewport.New()` with no positional args (changed from v1) |
| `ui.New()` signature | Value type: `New(cfg config.Config)`, not `New(cfg *config.Config)` |

---

## Verification

```sh
cd template-v2-enhanced

# 1. Add dependencies
go get charm.land/bubbles/v2 charm.land/lipgloss/v2

# 2. Build all packages
go build ./...

# 3. Run the app
go run .

# Manual tests:
# a) Arrow keys / j/k navigate the list
# b) Enter on "Details" pushes DetailScreen
# c) ESC from DetailScreen pops back to menu
# d) "/" enters filter mode; ESC exits filter (does NOT pop)
# e) "?" toggles full help — list height adjusts
# f) Resize terminal — list height adjusts dynamically
# g) ctrl+c quits from any screen

# 4. Run with debug logging
go run . --debug
tail -f debug.log

# 5. Config flag still works (no regression from previous session)
go run . --config assets/config.default.json
go run . version
```

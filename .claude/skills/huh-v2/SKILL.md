---
name: huh-v2
description: >-
  Generates, explains, and debugs interactive terminal form and prompt code using
  Charm's huh v2 library for Go (charm.land/huh/v2). Use when building CLI tools
  that need user input via forms, single-field prompts, multi-step wizards,
  dynamic forms that react to earlier answers, accessible prompts for screen
  readers, or loading spinners. Covers all field types (Input, Text, Select,
  MultiSelect, Confirm, Note, FilePicker), Form/Group composition, dynamic Func
  variants, BubbleTea integration, themes, layouts, accessibility mode, keymaps,
  and the standalone spinner package. Import path: charm.land/huh/v2.
argument-hint: "[describe the form or field you need]"
metadata:
  category: go-tui
  version: "2.0"
---

# huh v2 — Terminal Forms & Prompts

**Import:** `charm.land/huh/v2`
**Spinner subpackage:** `charm.land/huh/v2/spinner`
**Docs:** https://pkg.go.dev/charm.land/huh/v2
`go get charm.land/huh/v2@v2.0.0-20260105203756-d8977490d20c`

huh? builds interactive terminal forms from fields grouped into pages. Each field is a Bubble Tea Model — usable standalone via `.Run()` or embedded in a larger Bubble Tea app.

## Quick Start

```go
var name string
err := huh.NewInput().Title("Name?").Value(&name).Run()
```

Or a full multi-page form:

```go
err := huh.NewForm(
    huh.NewGroup(
        huh.NewSelect[string]().
            Title("Language").
            Options(huh.NewOptions("Go", "Rust", "Zig")...).
            Value(&lang),
    ),
    huh.NewGroup(
        huh.NewInput().Title("Name").Value(&name).
            Validate(func(s string) error {
                if s == "" { return errors.New("required") }
                return nil
            }),
        huh.NewConfirm().Title("Proceed?").Value(&ok),
    ),
).Run()
if errors.Is(err, huh.ErrUserAborted) { os.Exit(0) }
```

## Field Types

| Field | Constructor | Value type |
|-------|-------------|------------|
| Single-line text | `huh.NewInput()` | `*string` |
| Multi-line text | `huh.NewText()` | `*string` |
| Single choice | `huh.NewSelect[T]()` | `*T` |
| Multiple choices | `huh.NewMultiSelect[T]()` | `*[]T` |
| Yes/No | `huh.NewConfirm()` | `*bool` |
| Display text | `huh.NewNote()` | none |
| File chooser | `huh.NewFilePicker()` | `*string` |

## Common Field Methods (all fields)

```go
.Title("label")                   // static title
.TitleFunc(func() string, &bind)  // dynamic — re-evaluated when bind changes
.Description("hint text")
.DescriptionFunc(func() string, &bind)
.Value(&variable)                 // pointer to result variable
.Key("fieldKey")                  // key for form.GetString/GetInt/GetBool
.Validate(func(v T) error)        // return nil to pass, error to show message
.Run()                            // run field in isolation (blocking)
```

## Input Field

```go
huh.NewInput().
    Title("Username").
    Placeholder("johndoe").
    Prompt("> ").             // prefix character, default ">"
    CharLimit(50).
    EchoMode(huh.EchoModePassword). // hide input (passwords)
    Suggestions([]string{"alice", "bob"}). // tab-complete
    SuggestionsFunc(func() []string { ... }, &bind).
    Inline(true).             // title and input on same line
    Value(&s)
```

`EchoMode` values: `EchoModeNormal`, `EchoModePassword`, `EchoModeNone`

## Text Field (multi-line)

```go
huh.NewText().
    Title("Bio").
    Placeholder("Tell us about yourself...").
    Lines(5).                 // visible rows
    CharLimit(500).
    ShowLineNumbers(true).
    ExternalEditor(true).     // ctrl+e opens $EDITOR
    Editor("vim").            // override editor
    EditorExtension("md").    // temp file extension
    Value(&bio)
```

## Select Field

```go
huh.NewSelect[string]().
    Title("Color").
    Options(
        huh.NewOption("Red", "red"),
        huh.NewOption("Blue", "blue").Selected(true), // pre-select
    ).
    Height(5).                // scroll if more options
    Inline(false).            // true = horizontal carousel
    Validate(func(v string) error { ... }).
    Value(&color)

// Shorthand when key == value:
huh.NewSelect[string]().Options(huh.NewOptions("Go", "Rust", "Zig")...)

// Dynamic options (fetched/computed from previous answers):
huh.NewSelect[string]().
    OptionsFunc(func() []huh.Option[string] {
        return fetchOptions(country) // can do network calls; huh caches results
    }, &country).  // re-run only when country changes
    TitleFunc(func() string { return "State in " + country }, &country)
```

`huh.Hovered()` → returns the value currently under the cursor before submit.

## MultiSelect Field

```go
huh.NewMultiSelect[string]().
    Title("Features").
    Options(
        huh.NewOption("Auth", "auth").Selected(true),
        huh.NewOption("Logging", "logging"),
    ).
    Limit(3).             // max selections (0 = unlimited)
    Filterable(true).     // show "/" filter input
    Height(8).
    Validate(func(vs []string) error {
        if len(vs) == 0 { return errors.New("pick at least one") }
        return nil
    }).
    Value(&features)
```

Key bindings: `space`/`x` toggle, `ctrl+a` select-all/none, `/` filter.

## Confirm Field

```go
huh.NewConfirm().
    Title("Deploy?").
    Description("This will restart the server.").
    Affirmative("Yes!").    // default "Yes"
    Negative("No.").        // set "" for single-button (no toggle)
    Inline(false).
    Value(&confirmed)
```

Key bindings: `←`/`→` or `h`/`l` toggle, `y` yes, `n` no.

## Note Field (display only)

```go
huh.NewNote().
    Title("Welcome").
    Description("_italic_ *bold* `code`"). // basic markdown
    Next(true).             // show a Next button
    NextLabel("Continue").
    Height(10)
// Notes are skipped by default (auto-advance). Next(true) makes user confirm.
// DescriptionFunc can create live markdown preview panels.
```

## FilePicker Field

```go
huh.NewFilePicker().
    Title("Config file").
    CurrentDirectory(".").
    AllowedTypes([]string{".json", ".yaml"}).
    Value(&path)
```

## Options Helpers

```go
// From explicit key/value pairs:
huh.NewOption("Display Label", actualValue)
huh.NewOption("Active", true).Selected(true)

// Auto-generate options where key == fmt.Sprint(value):
huh.NewOptions("Go", "Rust", "Zig")        // []Option[string]
huh.NewOptions(1, 2, 3)                     // []Option[int]
```

## Form & Group

```go
form := huh.NewForm(group1, group2, ...)

// Form options
form.WithTheme(huh.ThemeFunc(huh.ThemeCharm))
form.WithWidth(80)
form.WithHeight(24)
form.WithAccessible(os.Getenv("ACCESSIBLE") != "")
form.WithShowHelp(true)
form.WithShowErrors(true)
form.WithKeyMap(huh.NewDefaultKeyMap())
form.WithLayout(huh.LayoutDefault)
form.WithTimeout(30 * time.Second)
form.WithOutput(os.Stderr)          // default output

// Group options
group := huh.NewGroup(field1, field2).
    Title("Page title").
    Description("Page subtitle").
    WithHide(true).           // static skip
    WithHideFunc(func() bool { return !showGroup }) // dynamic skip
```

## Dynamic Forms (Func variants)

All fields support `XxxFunc(fn, &binding)` for every static setter. The function
re-runs only when the value at `&binding` changes (pointer equality tracking).
Results are automatically cached.

```go
var country string
huh.NewSelect[string]().
    TitleFunc(func() string {
        if country == "US" { return "State" }
        return "Province"
    }, &country).
    OptionsFunc(func() []huh.Option[string] {
        return fetchRegions(country) // may be slow — huh shows spinner
    }, &country).
    Value(&region)
```

## Group Hiding (Conditional Pages)

```go
var wantExtras bool
huh.NewForm(
    huh.NewGroup(
        huh.NewConfirm().Title("Want extras?").Value(&wantExtras),
    ),
    huh.NewGroup(
        huh.NewText().Title("Describe extras").Value(&extras),
    ).WithHideFunc(func() bool { return !wantExtras }),
)
```

## Layouts

```go
form.WithLayout(huh.LayoutDefault)          // one group at a time (default)
form.WithLayout(huh.LayoutStack)            // all groups stacked
form.WithLayout(huh.LayoutColumns(2))       // groups in N columns
form.WithLayout(huh.LayoutGrid(rows, cols)) // groups in a grid
```

## Themes

Built-in themes (all accept `isDark bool`):
- `huh.ThemeCharm` (default)
- `huh.ThemeDracula`
- `huh.ThemeCatppuccin`
- `huh.ThemeBase16`
- `huh.ThemeBase`

```go
form.WithTheme(huh.ThemeFunc(huh.ThemeDracula))
form.WithTheme(huh.ThemeFunc(huh.ThemeCatppuccin))

// Custom theme — build on an existing one:
myTheme := huh.ThemeFunc(func(isDark bool) *huh.Styles {
    t := huh.ThemeCharm(isDark)
    t.Focused.Title = t.Focused.Title.Foreground(lipgloss.Color("#FF0000"))
    return t
})
form.WithTheme(myTheme)
```

## Accessibility Mode

```go
accessible := os.Getenv("ACCESSIBLE") != ""
form.WithAccessible(accessible)
// Drops TUI rendering; uses plain line prompts instead.
// TERM=dumb triggers accessible mode automatically.
```

## Standalone Spinner

```go
import "charm.land/huh/v2/spinner"

// Action style (blocking):
err := spinner.New().
    Title("Processing...").
    Action(func() { doWork() }).
    Run()

// Action with context and error:
err := spinner.New().
    Title("Fetching data...").
    ActionWithErr(func(ctx context.Context) error {
        return fetchData(ctx)
    }).
    Run()

// Context style (non-blocking action):
go doWork()
err := spinner.New().
    Type(spinner.Line).  // Dots, MiniDot, Jump, Points, Pulse, Globe, Moon, Monkey, Meter, Hamburger, Ellipsis
    Title("Working...").
    Context(ctx).
    Run()
```

## BubbleTea Integration

`huh.Form` implements `tea.Model`. Use `.Init()`, `.Update()`, `.View()` directly:

```go
type Model struct {
    form *huh.Form
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    form, cmd := m.form.Update(msg)
    if f, ok := form.(*huh.Form); ok {
        m.form = f
    }
    if m.form.State == huh.StateCompleted {
        // read results here
        class := m.form.GetString("class")
        level := m.form.GetInt("level")
    }
    return m, cmd
}

func (m Model) View() string {
    if m.form.State == huh.StateCompleted {
        return "Done!"
    }
    return m.form.View()
}
```

Use `.Key("name")` on fields to read via `form.Get("name")`, `form.GetString()`, `form.GetInt()`, `form.GetBool()`. Use `.Value(&var)` when you don't need the key-based API.

## Form State & Errors

```go
// States: huh.StateNormal, huh.StateCompleted, huh.StateAborted
// Errors from Run():
huh.ErrUserAborted  // user pressed ctrl+c
huh.ErrTimeout      // WithTimeout exceeded
// Check in BubbleTea: m.form.State == huh.StateCompleted

// Get current field errors:
errs := m.form.Errors()

// Custom submit/cancel callbacks (BubbleTea mode):
form.SubmitCmd = mySubmitCmd
form.CancelCmd = myQuitCmd
```

## Keymap Customization

```go
km := huh.NewDefaultKeyMap()
km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "q"))
km.Select.Filter = key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl+f", "filter"))
form.WithKeyMap(km)
```

## Examples Index

| Example | File | Demonstrates |
|---------|------|--------------|
| Burger ordering | `_examples/burger/main.go` | Full multi-group form, Note, Select, MultiSelect, Input, Text, Confirm, Spinner |
| BubbleTea embed | `_examples/bubbletea/main.go` | huh.Form as tea.Model, Key-based results, side panel |
| Dynamic country | `_examples/dynamic-country/main.go` | OptionsFunc with async API call, TitleFunc |
| Conditional pages | `_examples/conditional/main.go` | OptionsFunc reacting to earlier selection |
| Accessibility | `_examples/accessibility/main.go` | WithAccessible(true) |
| Spinner | `_examples/spinner/main.go` | Standalone spinner with action |
| Group hide | `_examples/hide/main.go` | WithHide, WithHideFunc |
| Layout columns | `_examples/layout/main.go` | LayoutColumns |
| File picker | `_examples/filepicker/main.go` | NewFilePicker |
| Theme switcher | `_examples/theme/main.go` | Multiple built-in themes |

See `references/API.md` for the complete API surface.

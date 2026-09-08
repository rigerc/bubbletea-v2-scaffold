# Settings Screen — Dynamic huh-v2 Form Plan

---

## Architecture

```/dev/null/arch.txt#L1-18
config/
  config.go      MODIFY — add cfg_label/cfg_desc/cfg_options/cfg_readonly tags
  schema.go      NEW    — reflection → []GroupMeta, no huh dependency
  save.go        NEW    — Save(*Config, path) via koanf round-trip

internal/ui/
  theme/
    huh.go       NEW    — HuhTheme() palette adapter (ThemeFunc closure)
  screens/
    settings.go  NEW    — SettingsScreen + ReflectAccessor + buildForm
  model.go       MODIFY — configPath field, wire settings nav + SettingsSavedMsg
  ui.go          MODIFY — New() accepts configPath string

main.go          MODIFY — pass configPath to ui.New()
go.mod           MODIFY — add charm.land/huh/v2
```

The `config` package has **zero** UI dependencies. The form builder in `screens/settings.go` has **zero** references to `cfg.LogLevel`, `cfg.UI.ThemeName`, etc — it only sees `[]GroupMeta`.

---

## File 1: `config/config.go` — MODIFY

Add four struct tags to every settable field. The schema reflector reads these to produce form metadata; everything else about the config is unchanged.

```/dev/null/config_tags.go#L1-33
type Config struct {
    LogLevel string `json:"logLevel" mapstructure:"logLevel" koanf:"logLevel"
        cfg_label:"Log Level"
        cfg_desc:"Logging verbosity (effective level shown in footer)"
        cfg_options:"trace,debug,info,warn,error,fatal"`

    Debug bool `json:"debug" mapstructure:"debug" koanf:"debug"
        cfg_label:"Debug Mode"
        cfg_desc:"Forces log level to trace; writes debug.log"`

    UI  UIConfig  `json:"ui"  mapstructure:"ui"  koanf:"ui"  cfg_label:"UI Settings"`
    App AppConfig `json:"app" mapstructure:"app" koanf:"app" cfg_label:"Application"`
}

type UIConfig struct {
    MouseEnabled bool   `json:"mouseEnabled" mapstructure:"mouseEnabled" koanf:"mouseEnabled"
        cfg_label:"Mouse Support"
        cfg_desc:"Enable mouse click and scroll events"`

    ThemeName string `json:"themeName" mapstructure:"themeName" koanf:"themeName"
        cfg_label:"Color Theme"
        cfg_desc:"Visual theme for the application"
        cfg_options:"default,dark,light"`
}

type AppConfig struct {
    Name    string `json:"name"    mapstructure:"name"    koanf:"name"
        cfg_label:"App Name"  cfg_desc:"Displayed in the banner"`

    Version string `json:"version" mapstructure:"version" koanf:"version"
        cfg_label:"Version"   cfg_readonly:"true"`

    Title string `json:"title" mapstructure:"title" koanf:"title"
        cfg_label:"Window Title" cfg_desc:"Default title bar text"`
}
```

> **Note on tag syntax:** Go struct tags are a single back-quoted string. Each `key:"value"` pair is space-separated. The line breaks above are for readability in the plan; the actual source will have them on one line each.

---

## File 2: `config/schema.go` — NEW

Pure reflection, no huh import. This is the schema-to-metadata bridge.

```/dev/null/schema.go#L1-90
package config

import (
    "reflect"
    "slices"
    "strings"
)

// FieldKind determines which huh widget represents this field.
type FieldKind uint8

const (
    FieldInput    FieldKind = iota // string  → huh.Input
    FieldSelect                    // string + cfg_options → huh.Select[string]
    FieldConfirm                   // bool    → huh.Confirm
    FieldReadOnly                  // any     + cfg_readonly:"true" → huh.Note
)

// FieldMeta describes a single leaf config field as a UI field.
type FieldMeta struct {
    Key      string        // dot-path koanf key, e.g. "ui.themeName"
    Label    string        // from cfg_label tag, falls back to field name
    Desc     string        // from cfg_desc tag
    Kind     FieldKind     // derived from Go kind + tags
    Options  []string      // non-nil only for FieldSelect; values from cfg_options
    ReadOnly bool          // true when cfg_readonly:"true"
    Value    reflect.Value // settable Value pointing into the working *Config
}

// GroupMeta groups related FieldMeta entries into one huh.Group page.
type GroupMeta struct {
    Label  string
    Fields []FieldMeta
}

// Schema reflects over cfg and returns ordered groups of field metadata.
//
// IMPORTANT: cfg must be a pointer to the working copy that will live for the
// duration of the form session. All FieldMeta.Value fields point directly into

Good — I have everything I need. Here's the complete plan continuing exactly where it cut off:

---

## File 2: `config/schema.go` — NEW (continued)

```/dev/null/schema.go#L1-105
package config

import (
    "reflect"
    "slices"
    "strings"
)

// FieldKind determines which huh widget represents this field.
type FieldKind uint8

const (
    FieldInput    FieldKind = iota // string              → huh.Input
    FieldSelect                    // string + cfg_options → huh.Select[string]
    FieldConfirm                   // bool                → huh.Confirm
    FieldReadOnly                  // cfg_readonly:"true"  → huh.Note
)

// FieldMeta describes a single leaf config field as a UI form field.
type FieldMeta struct {
    Key      string        // dot-path koanf key, e.g. "ui.themeName"
    Label    string        // cfg_label tag, falls back to struct field name
    Desc     string        // cfg_desc tag
    Kind     FieldKind     // derived from Go kind + presence of cfg_* tags
    Options  []string      // non-nil only for FieldSelect; values from cfg_options
    ReadOnly bool          // true when cfg_readonly:"true"
    Value    reflect.Value // settable Value pointing into the working *Config
}

// GroupMeta groups related FieldMeta for one huh.Group page.
type GroupMeta struct {
    Label  string
    Fields []FieldMeta
}

// Schema reflects over cfg and returns ordered groups of field metadata.
//
// cfg MUST be a pointer to the working copy that lives for the entire
// form session. All FieldMeta.Value fields point directly into *cfg memory.
// The caller is responsible for keeping cfg alive while the form runs.
func Schema(cfg *Config) []GroupMeta {
    rv := reflect.ValueOf(cfg).Elem() // Config value
    rt := rv.Type()

    var groups  []GroupMeta
    var topFields []FieldMeta

    for i := range rt.NumField() {
        sf := rt.Field(i)
        fv := rv.Field(i)
        koanfKey := sf.Tag.Get("koanf")
        if koanfKey == "" {
            continue
        }

        if fv.Kind() == reflect.Struct {
            // Nested struct → its own group
            groups = append(groups, GroupMeta{
                Label:  tagOrName(sf, "cfg_label"),
                Fields: nestedFields(fv, koanfKey),
            })
        } else {
            topFields = append(topFields, leafField(sf, fv, koanfKey))
        }
    }

    if len(topFields) > 0 {
        // Prepend top-level scalar fields as a "General" group
        groups = slices.Insert(groups, 0, GroupMeta{
            Label:  "General",
            Fields: topFields,
        })
    }

    return groups
}

// nestedFields reflects over a struct value, prefixing keys with parent.
func nestedFields(rv reflect.Value, prefix string) []FieldMeta {
    rt := rv.Type()
    fields := make([]FieldMeta, 0, rt.NumField())
    for i := range rt.NumField() {
        sf := rt.Field(i)
        fv := rv.Field(i)
        key := sf.Tag.Get("koanf")
        if key == "" {
            continue
        }
        fields = append(fields, leafField(sf, fv, prefix+"."+key))
    }
    return fields
}

// leafField builds a FieldMeta from a single struct field and its reflect.Value.
func leafField(sf reflect.StructField, fv reflect.Value, key string) FieldMeta {
    readOnly := sf.Tag.Get("cfg_readonly") == "true"
    options  := parseOptions(sf.Tag.Get("cfg_options"))
    return FieldMeta{
        Key:      key,
        Label:    tagOrName(sf, "cfg_label"),
        Desc:     sf.Tag.Get("cfg_desc"),
        ReadOnly: readOnly,
        Options:  options,
        Kind:     deriveKind(fv.Kind(), options, readOnly),
        Value:    fv,
    }
}

func deriveKind(k reflect.Kind, options []string, readOnly bool) FieldKind {
    if readOnly {
        return FieldReadOnly
    }
    switch k {
    case reflect.Bool:
        return FieldConfirm
    case reflect.String:
        if len(options) > 0 {
            return FieldSelect
        }
        return FieldInput
    default:
        return FieldInput
    }
}

func tagOrName(sf reflect.StructField, tag string) string {
    if v := sf.Tag.Get(tag); v != "" {
        return v
    }
    return sf.Name
}

func parseOptions(s string) []string {
    if s == "" {
        return nil
    }
    parts := strings.Split(s, ",")
    for i, p := range parts {
        parts[i] = strings.TrimSpace(p)
    }
    return parts
}
```

**What this gives you for free:** add a new field to `AppConfig` with `cfg_label` and `cfg_desc` tags — it appears in the form on next build with no changes to `settings.go`, `buildForm`, or anything in the UI layer.

---

## File 3: `config/save.go` — NEW

This is the koanf integration. Every write goes through koanf's parse-validate-marshal pipeline. The atomic write (temp file + rename) prevents a partial write from corrupting the config file.

```/dev/null/save.go#L1-52
package config

import (
    "fmt"
    "os"

    koanfjson "github.com/knadh/koanf/parsers/json"
    "github.com/knadh/koanf/providers/rawbytes"
    "github.com/knadh/koanf/v2"
)

// Save persists cfg to the file at path using koanf as the write pipeline.
//
// The save process:
//  1. Validates cfg (same rules as Load)
//  2. Marshals to JSON bytes via cfg.ToJSON()
//  3. Parses those bytes back through koanf to produce a canonical key-value map
//  4. Marshals that map back to JSON via koanf's own marshaler
//  5. Writes atomically via a temp file + rename
//
// Using koanf as the round-trip ensures the output format is identical to
// what Load() expects, and catches any structural inconsistencies early.
func Save(cfg *Config, path string) error {
    if err := cfg.Validate(); err != nil {
        return fmt.Errorf("config: save validation: %w", err)
    }

    // Step 1: struct → JSON bytes
    raw, err := cfg.ToJSON()
    if err != nil {
        return fmt.Errorf("config: encoding for save: %w", err)
    }

    // Step 2: JSON bytes → koanf key-value map (validates structure)
    k := koanf.New(".")
    if err := k.Load(rawbytes.Provider(raw), koanfjson.Parser()); err != nil {
        return fmt.Errorf("config: koanf parse during save: %w", err)
    }

    // Step 3: koanf key-value map → canonical JSON bytes
    out, err := k.Marshal(koanfjson.Parser())
    if err != nil {
        return fmt.Errorf("config: koanf marshal during save: %w", err)
    }

    // Step 4: atomic write — temp file + rename prevents corruption
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, out, 0o644); err != nil {
        return fmt.Errorf("config: writing temp file: %w", err)
    }
    if err := os.Rename(tmp, path); err != nil {
        _ = os.Remove(tmp) // best-effort cleanup
        return fmt.Errorf("config: atomic rename: %w", err)
    }
    return nil
}
```

**Why the koanf round-trip matters:** if someone later adds a koanf middleware (e.g. environment variable overrides, secret expansion), those transformations live in koanf's pipeline. The save path going through koanf means the written file always reflects what koanf actually knows about the config, not a raw struct dump.

---

## File 4: `internal/ui/theme/huh.go` — NEW

```/dev/null/theme_huh.go#L1-46
package theme

import (
    "charm.land/huh/v2"
    "charm.land/lipgloss/v2"
)

// HuhTheme returns a huh.Theme that matches the application's visual palette.
//
// IMPORTANT: this uses huh.ThemeFunc — a closure that receives huh's own
// isDark bool on every View() call. Do NOT capture the app's isDark state
// here; let huh drive it. This keeps form colours correct even if the
// terminal background changes mid-session.
//
// Usage:
//
//	form.WithTheme(theme.HuhTheme())
func HuhTheme() huh.Theme {
    return huh.ThemeFunc(func(isDark bool) *huh.Styles {
        p := NewPalette(isDark) // always fresh from huh's own isDark
        s := huh.ThemeCharm(isDark)

        // Group chrome
        s.Group.Title       = s.Group.Title.Foreground(p.Accent).Bold(true)
        s.Group.Description = s.Group.Description.Foreground(p.Muted)

        // Focused field chrome
        s.Focused.Base             = s.Focused.Base.BorderForeground(p.Accent)
        s.Focused.Title            = s.Focused.Title.Foreground(p.Accent)
        s.Focused.Description      = s.Focused.Description.Foreground(p.Muted)
        s.Focused.SelectSelector   = s.Focused.SelectSelector.Foreground(p.AccentHover)
        s.Focused.NextIndicator    = s.Focused.NextIndicator.Foreground(p.AccentHover)
        s.Focused.PrevIndicator    = s.Focused.PrevIndicator.Foreground(p.AccentHover)
        s.Focused.FocusedButton    = s.Focused.FocusedButton.Background(p.Accent).Foreground(p.Inverse)
        s.Focused.TextInput.Cursor = s.Focused.TextInput.Cursor.Foreground(p.AccentHover)
        s.Focused.TextInput.Prompt = s.Focused.TextInput.Prompt.Foreground(p.Accent)
        s.Focused.ErrorMessage     = s.Focused.ErrorMessage.Foreground(p.Error)
        s.Focused.ErrorIndicator   = s.Focused.ErrorIndicator.Foreground(p.Error)

        // Blurred field: hidden border, everything else inherited
        s.Blurred      = s.Focused
        s.Blurred.Base = s.Focused.Base.BorderStyle(lipgloss.HiddenBorder())

        return s
    })
}
```

---

## File 5: `internal/ui/screens/settings.go` — NEW

This is the largest file. Three distinct responsibilities: ownership model, `ReflectAccessor[T]`, and `buildForm`.

### Ownership model — why `cfg *config.Config`

```/dev/null/settings_ownership.txt#L1-15
NewSettings(cfg config.Config)
  │
  ├── cfgCopy := cfg          — value copy on the heap via &cfgCopy
  ├── s.cfg = &cfgCopy        — Settings holds *Config pointer (stable address)
  ├── config.Schema(s.cfg)    — reflect.Values point into *s.cfg
  └── return s                — Settings value returned; pointer inside is stable

When rootModel stores s in m.current (interface value) and later copies it
via value receivers, ALL copies share the same *config.Config on the heap.
reflect.Values remain valid for the lifetime of the form session.

On StateCompleted: SettingsSavedMsg{Cfg: *s.cfg}
  — dereferences the heap copy to get the final mutated Config value.
```

### `ReflectAccessor[T]`

Generic accessor backed by a `reflect.Value`. Uses `reflect.Value.Set` which works for both `string` and `bool` without type-specific methods. The type parameter `T` drives the interface satisfaction for `huh.Accessor[T]`.

```/dev/null/reflect_accessor.go#L1-15
// reflectAccessor[T] implements huh.Accessor[T] via a reflect.Value.
// The Value must be settable (i.e. addressable) — guaranteed when obtained
// from config.Schema which reflects over a *Config pointer.
type reflectAccessor[T any] struct {
    v reflect.Value
}

func (a *reflectAccessor[T]) Get() T {
    return a.v.Interface().(T)
}

func (a *reflectAccessor[T]) Set(val T) {
    a.v.Set(reflect.ValueOf(val))
}
```

### `buildForm` and `buildField`

```/dev/null/build_form.go#L1-65
// buildForm constructs a *huh.Form from schema groups.
// No field names, koanf keys, or config types appear here —
// all information comes from GroupMeta/FieldMeta.
func buildForm(groups []config.GroupMeta) *huh.Form {
    huhGroups := make([]*huh.Group, 0, len(groups))
    for _, g := range groups {
        fields := make([]huh.Field, 0, len(g.Fields))
        for _, fm := range g.Fields {
            if f := buildField(fm); f != nil {
                fields = append(fields, f)
            }
        }
        if len(fields) > 0 {
            huhGroups = append(huhGroups,
                huh.NewGroup(fields...).Title(g.Label),
            )
        }
    }
    return huh.NewForm(huhGroups...)
}

// buildField dispatches on FieldKind to produce the correct huh.Field.
func buildField(m config.FieldMeta) huh.Field {
    switch m.Kind {

    case config.FieldSelect:
        // Build typed Option[string] slice from the tag-derived option strings.
        opts := make([]huh.Option[string], len(m.Options))
        for i, o := range m.Options {
            // Key is title-cased for display; Value is the raw config string.
            opts[i] = huh.NewOption(cases.Title(language.Und).String(o), o)
        }
        return huh.NewSelect[string]().
            Key(m.Key).
            Title(m.Label).
            Description(m.Desc).
            Options(opts...).
            Accessor(&reflectAccessor[string]{v: m.Value})

    case config.FieldConfirm:
        return huh.NewConfirm().
            Key(m.Key).
            Title(m.Label).
            Description(m.Desc).
            Affirmative("Yes").
            Negative("No").
            Accessor(&reflectAccessor[bool]{v: m.Value})

    case config.FieldReadOnly:
        // Note fields are non-interactive; skip=true by default so they render
        // inline but don't block keyboard navigation (see huh Note.Skip()).
        return huh.NewNote().
            Title(m.Label).
            Description(fmt.Sprint(m.Value.Interface()))

    default: // FieldInput
        return huh.NewInput().
            Key(m.Key).
            Title(m.Label).
            Description(m.Desc).
            Accessor(&reflectAccessor[string]{v: m.Value})
    }
}
```

**Import note:** `cases.Title(language.Und).String(o)` requires `golang.org/x/text/cases` and `golang.org/x/text/language`. If you'd rather avoid the dependency, a simple `strings.ToUpper(o[:1]) + o[1:]` is a fine substitute since option values are always lowercase ASCII.

### Full struct and BubbleTea methods

```/dev/null/settings_full.go#L1-95
// SettingsSavedMsg is emitted when the user submits the settings form.
// Root model handles this: updates m.cfg, persists to disk, pops screen.
type SettingsSavedMsg struct {
    Cfg config.Config
}

// Settings is the settings screen — a huh.Form driven by config reflection.
type Settings struct {
    cfg    *config.Config // working copy; reflect.Values in form point here
    form   *huh.Form
    width  int
    isDark bool
}

// NewSettings creates a Settings screen from a copy of the live config.
// The copy is heap-allocated so its address is stable across value copies
// of the Settings struct (needed for reflect.Value validity).
func NewSettings(cfg config.Config) Settings {
    cfgCopy := cfg
    s := Settings{cfg: &cfgCopy}
    schema := config.Schema(s.cfg)
    s.form = buildForm(schema).
        WithTheme(theme.HuhTheme())
    return s
}

// SetWidth implements the SetWidth(int) Screen interface used by rootModel.
func (s Settings) SetWidth(w int) Screen {
    s.width = w
    s.form.WithWidth(w)
    return s
}

// SetStyles implements the SetStyles(bool) Screen interface used by rootModel.
func (s Settings) SetStyles(isDark bool) Screen {
    s.isDark = isDark
    // huh re-derives theme per-render via ThemeFunc — nothing else needed.
    return s
}

// Init satisfies tea.Model. Width must be set before Init so groups
// lay out correctly; rootModel calls SetWidth via NavigateMsg first.
func (s Settings) Init() tea.Cmd {
    return s.form.Init()
}

// Update satisfies tea.Model.
func (s Settings) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Propagate resize to form before delegating.
    if ws, ok := msg.(tea.WindowSizeMsg); ok {
        s.width = ws.Width
        s.form.WithWidth(s.width)
    }

    form, cmd := s.form.Update(msg)
    s.form = form.(*huh.Form)

    switch s.form.State {
    case huh.StateCompleted:
        // huh has already written all final values into *s.cfg via the
        // reflectAccessor.Set calls. Dereference once for the message value.
        saved := *s.cfg
        return s, func() tea.Msg { return SettingsSavedMsg{Cfg: saved} }

    case huh.StateAborted:
        // User pressed Escape / q inside huh — treat as Back.
        return s, func() tea.Msg { return BackMsg{} }
    }

    return s, cmd
}

// View satisfies tea.Model.
func (s Settings) View() tea.View {
    return tea.NewView(s.Body())
}

// Body returns the form content for layout composition by rootModel.
func (s Settings) Body() string {
    // huh.Form.View() returns "" once quitting=true (after StateCompleted/
    // StateAborted). Show a brief transitional string to avoid a blank frame
    // while the rootModel processes the message in the next update cycle.
    if s.form.State != huh.StateNormal {
        return "Applying settings…"
    }
    return s.form.View()
}

// ShortHelp satisfies KeyBinder — returns nil so huh owns its own help bar.
// The global root help bar will still show esc/back and q/quit.
func (s Settings) ShortHelp() []key.Binding { return nil }

// FullHelp satisfies KeyBinder.
func (s Settings) FullHelp() [][]key.Binding { return nil }
```

**Why `ShortHelp` returns nil:** huh renders its own compact help inside each group's footer viewport. Surfacing it again in the root help bar would duplicate it. The global bindings (`esc`, `q`) remain visible via `GlobalKeyMap`.

---

## File 6: `internal/ui/model.go` — MODIFY

Three targeted changes. No existing logic is removed or restructured.

### a) Add `configPath` to the struct

```/dev/null/model_struct_patch.go#L1-8
type rootModel struct {
    cfg        config.Config
    configPath string      // ← NEW: empty string = no persistent save
    width      int
    // ... all other fields unchanged
}
```

### b) Update `newRootModel`

```/dev/null/model_ctor_patch.go#L1-6
func newRootModel(cfg config.Config, configPath string) rootModel {
    return rootModel{
        cfg:        cfg,
        configPath: configPath,
        current:    screens.NewHome(),
        keys:       keys.DefaultGlobalKeyMap(),
        help:       help.New(),
    }
}
```

### c) Extend `Update` — two new cases

Replace the existing `menu.SelectionMsg` handler and add `SettingsSavedMsg`:

```/dev/null/model_update_patch.go#L1-26
case menu.SelectionMsg:
    switch msg.Item.ScreenID() {
    case "settings":
        // Pass a copy of the live config; Settings makes its own heap copy.
        return m.Update(NavigateMsg{Screen: screens.NewSettings(m.cfg)})
    default:
        detail := screens.NewDetail(
            msg.Item.Title(),
            msg.Item.Description(),
            msg.Item.ScreenID(),
        )
        return m.Update(NavigateMsg{Screen: detail})
    }

case screens.SettingsSavedMsg:
    // 1. Apply the new config to live state.
    m.cfg = msg.Cfg
    // 2. Persist via koanf round-trip (no-op if no config file was loaded).
    if m.configPath != "" {
        if err := config.Save(&m.cfg, m.configPath); err != nil {
            // Non-fatal: log and continue; the in-memory config is already updated.
            // (logger is available if the app has initialised it)
            _ = err // TODO: surface via a status message
        }
    }
    // 3. Pop the settings screen.
    if m.stack.Len() > 0 {
        m.current = m.stack.Pop()
    }
    return m, nil
```

---

## File 7: `internal/ui/ui.go` — MODIFY

```/dev/null/ui_patch.go#L1-12
package ui

import (
    tea "charm.land/bubbletea/v2"
    "scaffold/config"
)

// New creates a new root model from the config and an optional config file path.
// configPath may be empty if running with defaults (no file on disk to save to).
func New(cfg config.Config, configPath string) rootModel {
    return newRootModel(cfg, configPath)
}

// Run starts the TUI program. Unchanged.
func Run(m rootModel) error {
    _, err := tea.NewProgram(m).Run()
    return err
}
```

---

## File 8: `main.go` — MODIFY

One change: pass `configPath` through to `ui.New`. The existing `loadConfig` already knows the path via `cmd.GetConfigFile()` — we just need to surface it.

```/dev/null/main_patch.go#L1-18
func main() {
    // ... unchanged up to loadConfig() ...

    cfg, configPath := loadConfig()   // ← returns (config, path) now

    // ... logging setup unchanged ...

    if err := ui.Run(ui.New(*cfg, configPath)); err != nil {   // ← pass path
        applogger.Fatal().Err(err).Msg("UI failed")
    }
}

// loadConfig returns both the effective Config and the path that was loaded
// (empty string if no file was used, meaning saves should be skipped).
func loadConfig() (*config.Config, string) {
    cfg := config.DefaultConfig()
    configPath := ""

    if path := cmd.GetConfigFile(); path != "" {
        fileCfg, err := config.Load(path)
        if err == nil {
            cfg = fileCfg
            configPath = path   // ← only set when a file was actually loaded
        }
    }

    if cmd.IsDebugMode() {
        cfg.Debug = true
    }
    if cmd.WasLogLevelSet() {
        cfg.LogLevel = cmd.GetLogLevel()
    }

    return cfg, configPath
}
```

---

## File 9: `go.mod` — MODIFY

```/dev/null/gomod_patch.txt#L1-1
charm.land/huh/v2 v2.0.0
```

Run after adding the import:

```/dev/null/goget.sh#L1-2
go get charm.land/huh/v2
go mod tidy
```

If `golang.org/x/text` (for `cases.Title`) isn't already an indirect dependency, `go mod tidy` will add it. Alternatively, replace `cases.Title(language.Und).String(o)` with the inline `strings.ToUpper(o[:1]) + o[1:]` to avoid the extra dependency entirely.

---

## Complete File Map

```/dev/null/file_map.txt#L1-20
scaffold/
├── config/
│   ├── config.go      MODIFY  add cfg_label/cfg_desc/cfg_options/cfg_readonly tags
│   ├── schema.go      NEW     Schema(*Config) []GroupMeta via reflection
│   ├── save.go        NEW     Save(*Config, path) koanf round-trip + atomic write
│   └── defaults.go    no change
│
├── internal/ui/
│   ├── theme/
│   │   ├── theme.go   no change
│   │   └── huh.go     NEW     HuhTheme() ThemeFunc closure
│   ├── screens/
│   │   ├── settings.go  NEW   Settings screen + ReflectAccessor + buildForm
│   │   ├── home.go      no change
│   │   └── detail.go    no change
│   ├── model.go       MODIFY  configPath field + settings nav + SettingsSavedMsg
│   └── ui.go          MODIFY  New() signature adds configPath string
│
├── main.go            MODIFY  loadConfig() returns (cfg, path); pass to ui.New
└── go.mod             MODIFY  add charm.land/huh/v2
```

---

## Implementation Order

| # | File | Reason for ordering |
|---|------|---------------------|
| 1 | `config/config.go` — add tags | Foundation; schema depends on these |
| 2 | `config/schema.go` | No new deps; can be built + tested standalone |
| 3 | `config/save.go` | Only needs existing koanf deps |
| 4 | `go get charm.land/huh/v2` + `go mod tidy` | Must exist before UI files compile |
| 5 | `internal/ui/theme/huh.go` | Only needs existing `theme` palette |
| 6 | `internal/ui/screens/settings.go` | Needs schema + theme/huh + huh dep |
| 7 | `internal/ui/model.go` | Needs settings.go for `SettingsSavedMsg` |
| 8 | `internal/ui/ui.go` | Needs updated `newRootModel` signature |
| 9 | `main.go` | Needs updated `ui.New` signature |
| 10 | `go build ./...` + `go vet ./...` | Verify |

---

## Critical Notes

### 1. `reflect.Value` settability
`config.Schema` must receive `cfg *Config` (a pointer). `reflect.ValueOf(cfg).Elem()` gives an addressable `reflect.Value`. Fields obtained from it are also addressable and thus **settable**. If you pass a non-pointer `Config`, `reflect.Value.CanSet()` will be false and `reflectAccessor.Set` will panic.

### 2. `form.Update` return type assertion
`huh.Form.Update` returns `(huh.Model, tea.Cmd)`. You must assert back to `*huh.Form`:
```/dev/null/assert_note.go#L1-3
form, cmd := s.form.Update(msg)
s.form = form.(*huh.Form)   // safe: huh.Form always returns itself
```

### 3. Never call `form.Run()` inside a BubbleTea program
`Run()` spawns its own `tea.Program`. It would create a nested program and deadlock. The embedded model pattern (`Init`/`Update`/`View` delegation) is the correct path for in-app forms.

### 4. `huh.Note` skip behaviour
`Note.skip` is `true` by default. When a Note sits alongside other fields in a group, `huh` renders it visually but skips it during keyboard navigation — the user never "focuses" it, which is exactly right for a read-only `Version` field. No special handling is needed.

### 5. Width before Init
`rootModel.Update` handles `NavigateMsg` by calling `SetWidth` on the new screen **before** it's stored in `m.current`. That width propagates via `s.form.WithWidth(w)`. Then on the next `WindowSizeMsg` (or if the form is `Init`'d), the layout is already correct. Do not reverse this order.

### 6. Growing the config is free
Adding a new field to `UIConfig` or `AppConfig` with appropriate `cfg_*` tags requires zero changes in `settings.go`, `buildForm`, or anywhere in the UI layer. The schema reflector picks it up automatically and `buildField` dispatches on the Go type.
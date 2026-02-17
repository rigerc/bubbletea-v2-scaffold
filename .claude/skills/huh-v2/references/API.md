# huh v2 — Complete API Reference

Import: `charm.land/huh/v2`
Spinner: `charm.land/huh/v2/spinner`

---

## Package-Level Types

### FormState

```go
type FormState int
const (
    StateNormal    FormState = iota // user is filling out the form
    StateCompleted                  // form submitted successfully
    StateAborted                    // user pressed quit key
)
```

### Errors

```go
var ErrUserAborted = errors.New("user aborted")    // ctrl+c or quit key
var ErrTimeout     = errors.New("timeout")          // WithTimeout exceeded
var ErrTimeoutUnsupported = errors.New("timeout is not supported in accessible mode")
```

### Field Interface

Every field type implements `Field`:

```go
type Field interface {
    Model                              // Init(), Update(), View() — tea.Model
    Blur() tea.Cmd
    Focus() tea.Cmd
    Error() error
    Run() error
    RunAccessible(w io.Writer, r io.Reader) error
    Skip() bool                        // true = auto-advance (e.g. Note)
    Zoom() bool                        // true = take full group height
    KeyBinds() []key.Binding
    WithTheme(Theme) Field
    WithKeyMap(*KeyMap) Field
    WithWidth(int) Field
    WithHeight(int) Field
    WithPosition(FieldPosition) Field
    GetKey() string
    GetValue() any
}
```

### FieldPosition

```go
type FieldPosition struct {
    Group, Field             int
    FirstField, LastField    int
    GroupCount               int
    FirstGroup, LastGroup    int
}
func (p FieldPosition) IsFirst() bool
func (p FieldPosition) IsLast() bool
```

---

## Form

### Constructor

```go
func NewForm(groups ...*Group) *Form
```

### Form Methods

```go
// Configuration — call before Run()
func (f *Form) WithAccessible(accessible bool) *Form
func (f *Form) WithShowHelp(v bool) *Form
func (f *Form) WithShowErrors(v bool) *Form
func (f *Form) WithTheme(theme Theme) *Form
func (f *Form) WithKeyMap(keymap *KeyMap) *Form
func (f *Form) WithWidth(width int) *Form       // 0 = auto from terminal
func (f *Form) WithHeight(height int) *Form     // 0 = auto from terminal
func (f *Form) WithOutput(w io.Writer) *Form    // default stderr
func (f *Form) WithInput(r io.Reader) *Form     // default stdin
func (f *Form) WithTimeout(t time.Duration) *Form
func (f *Form) WithProgramOptions(opts ...tea.ProgramOption) *Form
func (f *Form) WithViewHook(hook compat.ViewHook) *Form
func (f *Form) WithLayout(layout Layout) *Form

// Callbacks (BubbleTea mode)
SubmitCmd tea.Cmd   // called on submit; default tea.Quit
CancelCmd tea.Cmd   // called on abort; default tea.Interrupt
State FormState     // current state (read in Update)

// Running
func (f *Form) Run() error
func (f *Form) RunWithContext(ctx context.Context) error

// BubbleTea Model implementation
func (f *Form) Init() tea.Cmd
func (f *Form) Update(msg tea.Msg) (Model, tea.Cmd)
func (f *Form) View() string

// Reading results (after StateCompleted or Run)
func (f *Form) Get(key string) any
func (f *Form) GetString(key string) string
func (f *Form) GetInt(key string) int
func (f *Form) GetBool(key string) bool

// Inspection
func (f *Form) Errors() []error
func (f *Form) Help() help.Model
func (f *Form) KeyBinds() []key.Binding
func (f *Form) GetFocusedField() Field

// Manual navigation (BubbleTea mode)
func (f *Form) NextGroup() tea.Cmd
func (f *Form) PrevGroup() tea.Cmd
func (f *Form) NextField() tea.Cmd
func (f *Form) PrevField() tea.Cmd
func (f *Form) UpdateFieldPositions() *Form
```

---

## Group

### Constructor

```go
func NewGroup(fields ...Field) *Group
```

### Group Methods

```go
func (g *Group) Title(title string) *Group
func (g *Group) Description(description string) *Group
func (g *Group) WithShowHelp(show bool) *Group
func (g *Group) WithShowErrors(show bool) *Group
func (g *Group) WithTheme(t Theme) *Group
func (g *Group) WithKeyMap(k *KeyMap) *Group
func (g *Group) WithWidth(width int) *Group
func (g *Group) WithHeight(height int) *Group
func (g *Group) WithHide(hide bool) *Group              // static
func (g *Group) WithHideFunc(f func() bool) *Group      // dynamic
func (g *Group) Errors() []error
func (g *Group) Header() string
func (g *Group) Footer() string
func (g *Group) Content() string
func (g *Group) View() string
```

---

## Input Field

```go
func NewInput() *Input

func (i *Input) Title(title string) *Input
func (i *Input) TitleFunc(f func() string, bindings any) *Input
func (i *Input) Description(description string) *Input
func (i *Input) DescriptionFunc(f func() string, bindings any) *Input
func (i *Input) Placeholder(str string) *Input
func (i *Input) PlaceholderFunc(f func() string, bindings any) *Input
func (i *Input) Prompt(prompt string) *Input              // prefix character
func (i *Input) CharLimit(charlimit int) *Input
func (i *Input) EchoMode(mode EchoMode) *Input
func (i *Input) Password(password bool) *Input            // deprecated; use EchoMode
func (i *Input) Suggestions(suggestions []string) *Input
func (i *Input) SuggestionsFunc(f func() []string, bindings any) *Input
func (i *Input) Inline(inline bool) *Input                // title+input on same line
func (i *Input) Validate(validate func(string) error) *Input
func (i *Input) Value(value *string) *Input
func (i *Input) Accessor(accessor Accessor[string]) *Input
func (i *Input) Key(key string) *Input
func (i *Input) Run() error

// EchoMode constants
const (
    EchoModeNormal   EchoMode // default
    EchoModePassword EchoMode // show mask character
    EchoModeNone     EchoMode // show nothing
)
```

---

## Text Field

```go
func NewText() *Text

func (t *Text) Title(title string) *Text
func (t *Text) TitleFunc(f func() string, bindings any) *Text
func (t *Text) Description(description string) *Text
func (t *Text) DescriptionFunc(f func() string, bindings any) *Text
func (t *Text) Placeholder(str string) *Text
func (t *Text) PlaceholderFunc(f func() string, bindings any) *Text
func (t *Text) Lines(lines int) *Text                    // visible row count
func (t *Text) CharLimit(charlimit int) *Text
func (t *Text) ShowLineNumbers(show bool) *Text
func (t *Text) ExternalEditor(enabled bool) *Text        // ctrl+e integration
func (t *Text) Editor(editor ...string) *Text            // cmd + optional args
func (t *Text) EditorExtension(extension string) *Text   // e.g. "md", "go"
func (t *Text) Validate(validate func(string) error) *Text
func (t *Text) Value(value *string) *Text
func (t *Text) Accessor(accessor Accessor[string]) *Text
func (t *Text) Key(key string) *Text
func (t *Text) Run() error
// Editor defaults to $EDITOR env var, fallback "nano"
// EditorExtension defaults to "md"
```

---

## Select Field

```go
func NewSelect[T comparable]() *Select[T]

func (s *Select[T]) Title(title string) *Select[T]
func (s *Select[T]) TitleFunc(f func() string, bindings any) *Select[T]
func (s *Select[T]) Description(description string) *Select[T]
func (s *Select[T]) DescriptionFunc(f func() string, bindings any) *Select[T]
func (s *Select[T]) Options(options ...Option[T]) *Select[T]
func (s *Select[T]) OptionsFunc(f func() []Option[T], bindings any) *Select[T]
func (s *Select[T]) Height(height int) *Select[T]         // enables scrolling
func (s *Select[T]) Inline(v bool) *Select[T]             // horizontal carousel
func (s *Select[T]) Filtering(filtering bool) *Select[T]  // start in filter mode
func (s *Select[T]) Validate(validate func(T) error) *Select[T]
func (s *Select[T]) Value(value *T) *Select[T]
func (s *Select[T]) Accessor(accessor Accessor[T]) *Select[T]
func (s *Select[T]) Key(key string) *Select[T]
func (s *Select[T]) Hovered() (T, bool)                   // value under cursor
func (s *Select[T]) GetFiltering() bool
func (s *Select[T]) Run() error
// OptionsFunc: bindings is a pointer to the variable whose change triggers re-evaluation.
// huh automatically caches results by bindings value hash.
// If options take time to load, huh shows a loading spinner automatically.
```

---

## MultiSelect Field

```go
func NewMultiSelect[T comparable]() *MultiSelect[T]

func (m *MultiSelect[T]) Title(title string) *MultiSelect[T]
func (m *MultiSelect[T]) TitleFunc(f func() string, bindings any) *MultiSelect[T]
func (m *MultiSelect[T]) Description(description string) *MultiSelect[T]
func (m *MultiSelect[T]) DescriptionFunc(f func() string, bindings any) *MultiSelect[T]
func (m *MultiSelect[T]) Options(options ...Option[T]) *MultiSelect[T]
func (m *MultiSelect[T]) OptionsFunc(f func() []Option[T], bindings any) *MultiSelect[T]
func (m *MultiSelect[T]) Limit(limit int) *MultiSelect[T]       // 0 = unlimited
func (m *MultiSelect[T]) Height(height int) *MultiSelect[T]
func (m *MultiSelect[T]) Width(width int) *MultiSelect[T]
func (m *MultiSelect[T]) Filterable(filterable bool) *MultiSelect[T]  // default true
func (m *MultiSelect[T]) Filtering(filtering bool) *MultiSelect[T]    // start in filter mode
func (m *MultiSelect[T]) Validate(validate func([]T) error) *MultiSelect[T]
func (m *MultiSelect[T]) Value(value *[]T) *MultiSelect[T]
func (m *MultiSelect[T]) Accessor(accessor Accessor[[]T]) *MultiSelect[T]
func (m *MultiSelect[T]) Key(key string) *MultiSelect[T]
func (m *MultiSelect[T]) Hovered() (T, bool)
func (m *MultiSelect[T]) GetFiltering() bool
func (m *MultiSelect[T]) Run() error
// Default key bindings: space/x=toggle, ctrl+a=select-all/none, /=filter
```

---

## Confirm Field

```go
func NewConfirm() *Confirm

func (c *Confirm) Title(title string) *Confirm
func (c *Confirm) TitleFunc(f func() string, bindings any) *Confirm
func (c *Confirm) Description(description string) *Confirm
func (c *Confirm) DescriptionFunc(f func() string, bindings any) *Confirm
func (c *Confirm) Affirmative(affirmative string) *Confirm  // default "Yes"
func (c *Confirm) Negative(negative string) *Confirm        // default "No"; "" = no toggle
func (c *Confirm) Inline(inline bool) *Confirm
func (c *Confirm) WithButtonAlignment(p lipgloss.Position) *Confirm
func (c *Confirm) Validate(validate func(bool) error) *Confirm
func (c *Confirm) Value(value *bool) *Confirm
func (c *Confirm) Accessor(accessor Accessor[bool]) *Confirm
func (c *Confirm) Key(key string) *Confirm
func (c *Confirm) String() string  // returns affirmative or negative label
func (c *Confirm) Run() error
// Default key bindings: h/l/←/→ toggle, y=yes, n=no, enter=confirm
```

---

## Note Field

```go
func NewNote() *Note

func (n *Note) Title(title string) *Note
func (n *Note) TitleFunc(f func() string, bindings any) *Note
func (n *Note) Description(description string) *Note       // supports _italic_ *bold* `code`
func (n *Note) DescriptionFunc(f func() string, bindings any) *Note
func (n *Note) Height(height int) *Note
func (n *Note) Next(show bool) *Note                       // show Next button
func (n *Note) NextLabel(label string) *Note               // default "Next"
func (n *Note) Run() error
// Notes auto-advance (Skip()=true) unless they are the only field in a group,
// or Next(true) is set (then user must press enter).
```

---

## FilePicker Field

```go
func NewFilePicker() *FilePicker

func (f *FilePicker) Title(title string) *FilePicker
func (f *FilePicker) Description(description string) *FilePicker
func (f *FilePicker) CurrentDirectory(directory string) *FilePicker
func (f *FilePicker) Cursor(cursor string) *FilePicker
func (f *FilePicker) Picking(v bool) *FilePicker          // start in pick mode
func (f *FilePicker) AllowedTypes(types []string) *FilePicker  // e.g. []string{".go", ".json"}
func (f *FilePicker) ShowHidden(v bool) *FilePicker
func (f *FilePicker) DirAllowed(v bool) *FilePicker
func (f *FilePicker) FileAllowed(v bool) *FilePicker
func (f *FilePicker) Validate(validate func(string) error) *FilePicker
func (f *FilePicker) Value(value *string) *FilePicker
func (f *FilePicker) Accessor(accessor Accessor[string]) *FilePicker
func (f *FilePicker) Key(key string) *FilePicker
func (f *FilePicker) Run() error
```

---

## Option Type

```go
type Option[T comparable] struct {
    Key   string  // display label
    Value T       // actual value
}

func NewOption[T comparable](key string, value T) Option[T]
func NewOptions[T comparable](values ...T) []Option[T]  // key = fmt.Sprint(value)

// Methods on Option (value receiver — must reassign or use in-place):
func (o Option[T]) Selected(selected bool) Option[T]  // pre-select in MultiSelect
func (o Option[T]) String() string                     // returns Key
```

---

## Layout

```go
type Layout interface {
    View(f *Form) string
    GroupWidth(f *Form, g *Group, w int) int
}

var LayoutDefault Layout          // one group at a time (default)
var LayoutStack Layout            // all groups stacked vertically
func LayoutColumns(columns int) Layout  // N equal-width columns (advances in segments)
func LayoutGrid(rows, columns int) Layout  // R×C grid layout
```

---

## Theme

```go
type Theme interface {
    Theme(isDark bool) *Styles
}
type ThemeFunc func(isDark bool) *Styles

// Built-in theme constructors (isDark bool) -> *Styles
func ThemeCharm(isDark bool) *Styles      // default
func ThemeDracula(isDark bool) *Styles
func ThemeCatppuccin(isDark bool) *Styles
func ThemeBase16(isDark bool) *Styles
func ThemeBase(isDark bool) *Styles       // minimal base

// Styles structure
type Styles struct {
    Form           FormStyles
    Group          GroupStyles
    FieldSeparator lipgloss.Style
    Blurred        FieldStyles
    Focused        FieldStyles
    Help           help.Styles
}
type FormStyles struct { Base lipgloss.Style }
type GroupStyles struct {
    Base, Title, Description lipgloss.Style
}
type FieldStyles struct {
    Base, Title, Description           lipgloss.Style
    ErrorIndicator, ErrorMessage       lipgloss.Style
    SelectSelector                     lipgloss.Style
    Option, NextIndicator, PrevIndicator lipgloss.Style
    Directory, File                    lipgloss.Style
    MultiSelectSelector                lipgloss.Style
    SelectedOption, SelectedPrefix     lipgloss.Style
    UnselectedOption, UnselectedPrefix lipgloss.Style
    TextInput                          TextInputStyles
    FocusedButton, BlurredButton       lipgloss.Style
    Card, NoteTitle, Next              lipgloss.Style
}
type TextInputStyles struct {
    Cursor, CursorText, Placeholder, Prompt, Text lipgloss.Style
}

// Usage:
form.WithTheme(huh.ThemeFunc(huh.ThemeDracula))

// Custom theme:
form.WithTheme(huh.ThemeFunc(func(isDark bool) *huh.Styles {
    t := huh.ThemeCharm(isDark)
    t.Focused.Title = t.Focused.Title.Foreground(lipgloss.Color("#FF00FF"))
    return t
}))
```

---

## KeyMap

```go
type KeyMap struct {
    Quit        key.Binding
    Confirm     ConfirmKeyMap
    FilePicker  FilePickerKeyMap
    Input       InputKeyMap
    MultiSelect MultiSelectKeyMap
    Note        NoteKeyMap
    Select      SelectKeyMap
    Text        TextKeyMap
}

func NewDefaultKeyMap() *KeyMap

// Default bindings summary:
// Global:  ctrl+c = quit
// Input:   enter/tab = next, shift+tab = back, ctrl+e = accept suggestion
// Text:    enter/tab = next, shift+tab = back, alt+enter/ctrl+j = newline, ctrl+e = open editor
// Select:  j/k/↑/↓ = navigate, / = filter, enter/tab = select, shift+tab = back
// Multi:   j/k/↑/↓ = navigate, space/x = toggle, ctrl+a = select-all, / = filter
// Confirm: h/l/←/→ = toggle, y = yes, n = no, enter = confirm
// Note:    enter = next, shift+tab = back
```

---

## Accessor Interface

For advanced use — wire a field to a custom getter/setter instead of a pointer:

```go
type Accessor[T any] interface {
    Get() T
    Set(T)
}

func NewPointerAccessor[T any](value *T) Accessor[T]  // wraps a *T

// EmbeddedAccessor is used internally when no Value() is set.
// Use Accessor() method instead of Value() to supply a custom one.
```

---

## Commands (BubbleTea navigation)

```go
func NextField() tea.Msg   // advance to next field
func PrevField() tea.Msg   // go back to previous field
// Used by fields internally; can also be dispatched from custom Update
```

---

## Spinner Package (`charm.land/huh/v2/spinner`)

```go
func New() *Spinner

func (s *Spinner) Title(title string) *Spinner
func (s *Spinner) Type(t Type) *Spinner
func (s *Spinner) Action(action func()) *Spinner                         // blocking
func (s *Spinner) ActionWithErr(action func(context.Context) error) *Spinner
func (s *Spinner) Context(ctx context.Context) *Spinner                  // for context-style
func (s *Spinner) WithAccessible(accessible bool) *Spinner
func (s *Spinner) WithTheme(theme Theme) *Spinner
func (s *Spinner) WithOutput(w io.Writer) *Spinner
func (s *Spinner) WithInput(r io.Reader) *Spinner
func (s *Spinner) WithViewHook(hook compat.ViewHook) *Spinner
func (s *Spinner) Run() error

// Spinner Types:
var (
    Line, Dots, MiniDot, Jump, Points, Pulse Type
    Globe, Moon, Monkey, Meter, Hamburger, Ellipsis Type
)

// Theme
type Styles struct { Spinner, Title lipgloss.Style }
func ThemeDefault(isDark bool) *Styles

// Note: Spinner is also a tea.Model (Init/Update/View) for BubbleTea embedding.
```

---

## Eval / Dynamic Pattern (internals)

The `Func` variants use an internal `Eval[T]` struct with:
- `val T` — current cached value
- `fn func() T` — recomputation function
- `bindings any` — pointer whose hash is tracked
- `cache map[uint64]T` — cache keyed by hash

When `bindings` value changes (by pointer hash), the fn is re-executed
asynchronously via tea.Cmd, and a spinner shows if loading takes >500ms.
All Func variants follow the same pattern:

```go
.XxxFunc(fn func() ReturnType, bindings any) *Field
// bindings is typically &someVar — the address of the variable that when
// changed should trigger recomputation.
```

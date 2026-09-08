# Research: plural-main BubbleTea v2 TUI Architecture

**Date:** 2026-02-28
**Source:** `docs/_go-project-examples/plural-main/`
**Model:** Claude Opus 4.6 (glm-5)

---

## Component Architecture

### Core Model Structure

The main model is defined in `internal/app/app.go` (~1738 lines). It follows the classic BubbleTea pattern with a composite model:

```go
type Model struct {
    // UI Components
    header  *ui.Header
    footer  *ui.Footer
    sidebar *ui.Sidebar
    chat    *ui.Chat
    modal   *ui.Modal

    // Layout state
    width  int
    height int
    focus  Focus  // FocusSidebar | FocusChat

    // Services (dependency injection)
    gitService     *git.GitService
    sessionService *session.SessionService
    issueRegistry  *issues.ProviderRegistry

    // State machine
    state AppState  // StateIdle | StateStreamingClaude
}
```

### Key Architecture Patterns

#### 1. Focus Management (`internal/app/app.go:26-32`)

```go
type Focus int
const (
    FocusSidebar Focus = iota
    FocusChat
)
```

- Tab key switches focus between panels
- Each panel has focused/unfocused styling

#### 2. State Machine (`internal/app/app.go:37-54`)

```go
type AppState int
const (
    StateIdle AppState = iota
    StateStreamingClaude
)

func (s AppState) String() string {
    switch s {
    case StateIdle:
        return "Idle"
    case StateStreamingClaude:
        return "StreamingClaude"
    default:
        return "Unknown"
    }
}
```

- Explicit state prevents invalid combinations
- State transitions are traceable
- String() method aids debugging

#### 3. Service Pattern - Dependency injection with interfaces:

- `GitService` - all git operations
- `SessionService` - worktree creation/management
- `IssueProvider` interface with GitHub, Asana, Linear implementations
- `CommandExecutor` interface for testability

#### 4. Message-Driven Communication - 40+ message types:

| Message Type | Purpose |
|-------------|---------|
| `ClaudeResponseMsg` | Streaming response chunks |
| `PermissionRequestMsg` | Permission prompts |
| `QuestionRequestMsg` | AskUserQuestion prompts |
| `PlanApprovalRequestMsg` | ExitPlanMode requests |
| `IssuesFetchedMsg` | Async data fetch results |
| `tea.WindowSizeMsg` | Terminal resize |

### View Method Composition

The View method (`internal/app/view.go:21-81`) uses lipgloss composition:

```go
func (m *Model) View() tea.View {
    var v tea.View
    v.AltScreen = true
    v.MouseMode = tea.MouseModeCellMotion
    v.ReportFocus = true

    // Render components
    header := m.header.View()
    footer := m.footer.View()
    sidebarView := m.sidebar.View()
    chatView := m.chat.View()

    // Horizontal: side-by-side panels
    panels := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, chatView)

    // Vertical: stack header, panels, footer
    view := lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)

    // Modal overlay (if visible)
    if m.modal.IsVisible() {
        v.SetContent(lipgloss.Place(
            m.width, m.height,
            lipgloss.Center, lipgloss.Center,
            m.modal.View(m.width, m.height),
        ))
        return v
    }

    v.SetContent(view)
    return v
}
```

---

## Layout System

### Layout Structure

```
┌──────────────────────────────────────────────────────────────────┐
│                        Header (1 line)                           │
├──────────────────────┬───────────────────────────────────────────┤
│    Sidebar (1/5)     │           Chat Panel (4/5)                │
│                      │                                           │
│    - Repo list       │    ┌─────────────────────────────────┐    │
│    - Session tree    │    │  Message viewport               │    │
│    - Search          │    │  (scrollable)                   │    │
│                      │    ├─────────────────────────────────┤    │
│                      │    │  Input textarea (3 lines)       │    │
│                      │    └─────────────────────────────────┘    │
├──────────────────────┴───────────────────────────────────────────┤
│                        Footer (1 line)                           │
└──────────────────────────────────────────────────────────────────┘
```

### Centralized Layout Calculations

`internal/ui/context.go` - Singleton ViewContext:

```go
type ViewContext struct {
    TerminalWidth  int
    TerminalHeight int
    HeaderHeight   int
    FooterHeight   int
    ContentHeight  int
    SidebarWidth   int
    ChatWidth      int
    mu sync.Mutex
}

func (v *ViewContext) UpdateTerminalSize(width, height int) {
    v.mu.Lock()
    defer v.mu.Unlock()

    // Validate dimensions
    if width < MinTerminalWidth {
        width = MinTerminalWidth
    }
    if height < MinTerminalHeight {
        height = MinTerminalHeight
    }

    v.TerminalWidth = width
    v.TerminalHeight = height
    v.ContentHeight = height - v.HeaderHeight - v.FooterHeight
    v.SidebarWidth = width / SidebarWidthRatio  // 5
    v.ChatWidth = width - v.SidebarWidth
}

func (v *ViewContext) InnerWidth(panelWidth int) int {
    return panelWidth - BorderSize
}
```

### Layout Constants

`internal/ui/constants.go`:

| Constant | Value | Purpose |
|----------|-------|---------|
| `HeaderHeight` | 1 | Fixed header lines |
| `FooterHeight` | 1 | Fixed footer lines |
| `BorderSize` | 2 | Border padding (1px each side) |
| `SidebarWidthRatio` | 5 | Sidebar gets 1/5, chat gets 4/5 |
| `MinTerminalWidth` | 40 | Minimum supported width |
| `MinTerminalHeight` | 10 | Minimum supported height |
| `TextareaHeight` | 3 | Input area height |

### Theme System

`internal/ui/theme.go` - 8 built-in themes:

```go
type Theme struct {
    Name        string
    Primary     string  // Accent color
    Secondary   string  // Secondary accent
    Bg          string  // Background
    BgSelected  string  // Selection background
    Text        string  // Primary text
    TextMuted   string  // Muted text
    TextInverse string  // Text on colored backgrounds

    // Semantic colors
    User      string // User message labels
    Assistant string // Assistant message labels
    Warning   string // Permission prompts, warnings
    Error     string // Error messages
    Info      string // Information, questions
    Success   string // Success messages, confirmations

    // Border colors
    Border      string // Default borders
    BorderFocus string // Focused element borders

    // Diff colors
    DiffAdded   string
    DiffRemoved string
    DiffHeader  string
    DiffHunk    string

    // Markdown colors
    MarkdownH1       string
    MarkdownH2       string
    MarkdownH3       string
    MarkdownCode     string
    MarkdownCodeBg   string
    MarkdownLink     string
    MarkdownListItem string

    // Text selection
    TextSelectionBg string
    TextSelectionFg string

    // Syntax highlighting
    SyntaxStyle string // Chroma style name
}
```

**Available themes:** `dark-purple`, `nord`, `dracula`, `gruvbox`, `tokyo-night`, `catppuccin`, `science-fiction`, `light`

### Style Registry

`internal/ui/styles.go` - Centralized style definitions:

```go
var (
    ColorPrimary     = lipgloss.Color("#7C3AED") // Purple
    ColorSecondary   = lipgloss.Color("#06B6D4") // Cyan
    ColorMuted       = lipgloss.Color("#6B7280") // Gray
    ColorBorder      = lipgloss.Color("#374151") // Dark gray
    ColorBorderFocus = lipgloss.Color("#7C3AED") // Purple when focused
    ColorBg          = lipgloss.Color("#1F2937") // Dark background
    ColorText        = lipgloss.Color("#F9FAFB") // Light text
    ColorUser        = lipgloss.Color("#A78BFA") // Light purple
    ColorAssistant   = lipgloss.Color("#22D3EE") // Bright cyan
    ColorWarning     = lipgloss.Color("#F59E0B") // Amber
    ColorError       = lipgloss.Color("#EF4444") // Red
    ColorSuccess     = lipgloss.Color("#10B981") // Green
)

var (
    HeaderStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(ColorText).
        Background(ColorPrimary).
        Padding(0, 1)

    PanelStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorBorder)

    PanelFocusedStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(ColorBorderFocus)
)
```

---

## Key Files Summary

| File | Purpose | Lines |
|------|---------|-------|
| `internal/app/app.go` | Main model, Init, Update | ~1738 |
| `internal/app/view.go` | View composition | ~100 |
| `internal/ui/context.go` | Centralized layout calculations | ~99 |
| `internal/ui/constants.go` | Layout constants | ~50 |
| `internal/ui/styles.go` | Style definitions | ~200 |
| `internal/ui/theme.go` | Theme management, 8 themes | ~300 |
| `internal/ui/sidebar.go` | Session tree, focus management | ~600 |
| `internal/ui/chat.go` | Messages, input, viewport | ~800 |
| `internal/ui/header.go` | Gradient header with session info | ~200 |
| `internal/ui/footer.go` | Context-aware shortcuts | ~150 |
| `internal/ui/modal.go` | Modal dialog factory | ~100 |

---

## Comparison with Scaffold

### Architecture Differences

| Aspect | Plural | Scaffold |
|--------|--------|----------|
| **Layout Model** | Panel-based (sidebar/chat split) | Screen-based (page navigation) |
| **Navigation** | Focus enum + Tab switching | screenStack + NavigateMsg |
| **Layout Calc** | ViewContext singleton | bodyHeight() method in rootModel |
| **Theme System** | Global style variables + regenerateStyles() | theme.Manager + ThemeAware interface |
| **State Machine** | AppState enum (Idle/Streaming) | rootState enum (Loading/Ready/Error) |
| **Message Types** | 40+ in app.go | ~15 in model.go + screen-specific |

### What Scaffold Already Does Well

1. **Theme System** - `theme.Manager` with `ThemeAware` interface is more modular than Plural's global styles
2. **Screen Navigation** - `screenStack` pattern is cleaner for multi-page apps
3. **Dynamic Layout** - `bodyHeight()` measures actual rendered content vs Plural's fixed constants
4. **huh Integration** - Settings screen uses huh forms with theme integration

---

## Patterns Worth Extracting

### 1. ViewContext Singleton (Recommended)

**Current Scaffold** (`scaffold/internal/ui/view.go:100-113`):
```go
func (m rootModel) bodyHeight() int {
    if m.height == 0 {
        return 0
    }
    header := lipgloss.Height(m.headerView())
    helpH := lipgloss.Height(m.helpView())
    body := m.height - header - helpH - footerLines
    if body < minBodyLines {
        body = minBodyLines
    }
    return body
}
```

**Plural Pattern** (`plural-main/internal/ui/context.go`):
```go
type ViewContext struct {
    TerminalWidth  int
    TerminalHeight int
    HeaderHeight   int
    FooterHeight   int
    ContentHeight  int
    mu sync.Mutex
}

func (v *ViewContext) UpdateTerminalSize(width, height int) {
    v.ContentHeight = height - v.HeaderHeight - v.FooterHeight
}

func (v *ViewContext) InnerWidth(panelWidth int) int {
    return panelWidth - BorderSize
}
```

**Benefit**: Centralized layout calculations, thread-safe, reusable across components.

### 2. Focus Management (For Future Multi-Panel Layouts)

**Plural Pattern** (`internal/app/app.go:26-32`):
```go
type Focus int
const (
    FocusSidebar Focus = iota
    FocusChat
)
```

**When to Use**: If scaffold adds split-pane views (e.g., file tree + editor), this pattern provides explicit focus tracking.

### 3. State Machine Enum with String()

**Current Scaffold** (`scaffold/internal/ui/model.go:26-32`):
```go
type rootState int
const (
    rootStateLoading rootState = iota
    rootStateReady
    rootStateError
)
```

**Plural Enhancement**:
```go
func (s AppState) String() string {
    switch s {
    case StateIdle: return "Idle"
    case StateStreamingClaude: return "StreamingClaude"
    default: return "Unknown"
    }
}
```

**Benefit**: Better logging and debugging.

### 4. Message Type Organization

**Plural**: All message types defined in one place (`app.go`) with clear naming:
- `ClaudeResponseMsg`
- `PermissionRequestMsg`
- `IssuesFetchedMsg`

**Scaffold**: Messages scattered across files. Consider consolidating.

---

## Extraction Plan

### Phase 1: Add ViewContext (Low Risk)

1. Create `scaffold/internal/ui/layout/context.go`
2. Define ViewContext singleton with thread-safe access
3. Refactor `bodyHeight()` to use ViewContext
4. Add constants for fixed heights

### Phase 2: Enhance State Machine (Low Risk)

1. Add `String()` method to `rootState`
2. Consider renaming to `AppState` for consistency

### Phase 3: Add Focus Management (Future)

1. Add when implementing split-pane views
2. Create `Focus` enum with Tab switching
3. Add focused/unfocused styles to theme

---

## Verification

After extraction:
1. Run `go test ./...` in scaffold directory
2. Run `go run .` to verify UI renders correctly
3. Test terminal resize behavior
4. Verify theme switching still works

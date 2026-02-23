# Plan: Refactor ralphio to Align with Ralph Workflow

## Context

The current `ralphio` implementation is a **task runner** that deterministically executes pre-defined tasks from `tasks.json`. The Ralph workflow (per the playbook) is fundamentally different: an agent autonomously manages its own plan, selects tasks, and evolves the implementation through iteration.

This plan refactors ralphio to align with the Ralph methodology while preserving the multi-adapter architecture and TUI.

---

## Key Differences (Current vs Ralph)

| Aspect | Current ralphio | Ralph Playbook |
|--------|-----------------|----------------|
| Task selection | Orchestrator picks by priority | Agent picks "most important" |
| Modes | BUILD only | PLANNING + BUILDING |
| Plan format | `tasks.json` (JSON, orchestrator-mutated) | `tasks.json` (JSON, agent-mutated) |
| Context per loop | Single task fields | Full specs + plan + AGENTS.md |
| Plan mutation | Status only | Add/remove tasks, discoveries, notes |

---

## Phase 1: File Structure Alignment

### 1a: New Required Files

Create project structure expected by Ralph:

```
project-root/
├── PROMPT_build.md   # Build mode instructions
├── PROMPT_plan.md    # Plan mode instructions
├── AGENTS.md         # Operational guide (build/test commands)
├── tasks.json        # Agent-maintained task list (existing schema)
└── specs/            # Requirement specifications
    └── *.md
```

### 1b: Retain `tasks.json` as the Plan Format

- `tasks.json` is the single source of truth for the plan — no format change.
- Agent reads and rewrites `tasks.json` directly (add, update, remove tasks).
- Schema per existing file:

```json
[
  {
    "id": "1",
    "title": "Short task title",
    "description": "What to do and acceptance criteria",
    "priority": 1,
    "status": "pending",
    "retryCount": 0,
    "maxRetries": 3,
    "validationCommand": "go test ./..."
  }
]
```

Valid `status` values: `"pending"`, `"in_progress"`, `"completed"`, `"failed"`.

### 1c: State File Update

Update `.ralph/state.json`:

```json
{
  "currentIteration": 14,
  "currentTaskId": "T-14",
  "loopStatus": "running",
  "loopMode": "building",
  "activeAdapter": "claude",
  "lastUpdated": "..."
}
```

Add `loopMode`: `"planning"` or `"building"`.

---

## Phase 2: Loop Mode System

### 2a: Mode Definition — `internal/plan/mode.go`

```go
type LoopMode string

const (
    ModePlanning LoopMode = "planning"
    ModeBuilding LoopMode = "building"
)
```

### 2b: Mode-Aware Orchestrator

Update `orchestrator.go` to:

1. Check current mode from state
2. Pass appropriate prompt file to adapter
3. PLANNING mode: agent generates/updates plan only
4. BUILDING mode: agent implements from plan

### 2c: Mode Selection in TUI

- Add `m` key binding to toggle mode
- Dashboard shows current mode indicator
- Mode change persists to state file

---

## Phase 3: Agent-Driven Task Selection

### 3a: Remove Orchestrator Task Selection

Current (remove):

```go
next := plan.NextTask(tasks)
if next == nil {
    o.send(LoopDoneMsg{})
    return
}
```

New approach:

```go
// Agent receives full context and decides what to do
prompt := o.buildPrompt()
execErr := o.currentAdapter.Execute(ctx, prompt, func(text string) {
    o.send(AgentOutputMsg{Text: text})
})
```

### 3b: Prompt Composition — `internal/prompt/builder.go`

Build prompt from project files each iteration:

```go
type PromptBuilder struct {
    projectDir string
    mode       plan.LoopMode
}

func (b *PromptBuilder) Build() (string, error) {
    // 1. Read PROMPT_build.md or PROMPT_plan.md based on mode
    // 2. Return full prompt text
    return promptContent, nil
}
```

The prompt files themselves instruct the agent to:
- Study `specs/*`
- Read `tasks.json`
- Study `src/lib/*`
- Select the highest-priority pending task
- Implement or plan accordingly

### 3c: Agent Commits, Not Orchestrator

The agent (via prompt instructions) handles:
- Updating `tasks.json` (status, new tasks, discoveries)
- Running validation via `AGENTS.md` commands
- Git commit when done

Orchestrator just:
1. Builds prompt
2. Executes adapter
3. Streams output to TUI
4. Waits for completion
5. Loops

---

## Phase 4: Plan Manager Refactor

### 4a: JSON Plan Parser — `internal/plan/parser.go`

Parse `tasks.json` for TUI display (schema is already the existing one):

```go
type Task struct {
    ID                string `json:"id"`
    Title             string `json:"title"`
    Description       string `json:"description"`
    Priority          int    `json:"priority"`
    Status            string `json:"status"`
    RetryCount        int    `json:"retryCount"`
    MaxRetries        int    `json:"maxRetries"`
    ValidationCommand string `json:"validationCommand"`
}

type ParsedPlan struct {
    Tasks []Task
}

func ParsePlan(content []byte) (*ParsedPlan, error)
```

No special parsing logic needed — standard `encoding/json` unmarshal.

### 4b: Plan Watcher

Watch `tasks.json` for file changes to update TUI in real-time:

```go
func (m *Manager) Watch(ctx context.Context, onChange func(tasks []Task)) error
```

Use `fsnotify` or polling.

### 4c: Remove `plan.Manager` CRUD

Agent manages plan directly via file writes. Remove:
- `SaveTasks()`
- `UpdateTask()`
- `NextTask()`

Keep:
- File watching
- Parsing for TUI display

---

## Phase 5: Prompt Templates

### 5a: `PROMPT_plan.md` Template

Port from playbook, adapted for `tasks.json`:

```markdown
0a. Study `specs/*` with up to 250 parallel Sonnet subagents to learn the application specifications.
0b. Study @tasks.json (if present) to understand the plan so far.
0c. Study `src/lib/*` with up to 250 parallel Sonnet subagents to understand shared utilities & components.
0d. For reference, the application source code is in `src/*`.

1. Study @tasks.json (if present; it may be incorrect) and use up to 500 Sonnet subagents to study existing source code in `src/*` and compare it against `specs/*`. Use an Opus subagent to analyze findings, prioritize tasks, and create/update @tasks.json as a JSON array sorted by priority (1 = highest) of items yet to be implemented. Each task must follow the schema: id, title, description, priority, status ("pending"), retryCount (0), maxRetries, validationCommand. Ultrathink. Consider searching for TODO, minimal implementations, placeholders, skipped/flaky tests, and inconsistent patterns.

IMPORTANT: Plan only. Do NOT implement anything. Do NOT assume functionality is missing; confirm with code search first.
```

### 5b: `PROMPT_build.md` Template

Port from playbook, adapted for `tasks.json`:

```markdown
0a. Study `specs/*` with up to 500 parallel Sonnet subagents to learn the application specifications.
0b. Study @tasks.json.
0c. For reference, the application source code is in `src/*`.

1. Your task is to implement functionality per the specifications using parallel subagents. Follow @tasks.json and choose the highest-priority pending task to address. Set its status to "in_progress" in @tasks.json before starting. Before making changes, search the codebase (don't assume not implemented) using Sonnet subagents.
2. After implementing functionality or resolving problems, run the tests for that unit of code that was improved.
3. When you discover issues, immediately add them as new tasks in @tasks.json with appropriate priority.
4. When the tests pass, set the task status to "completed" in @tasks.json, then `git add -A` then `git commit` with a message describing the changes.

99999. Important: When authoring documentation, capture the why.
999999. Single sources of truth, no migrations/adapters.
9999999. As soon as there are no build or test errors create a git tag.
9999999999. Keep @tasks.json current with learnings.
```

### 5c: Configurable Prompt Paths

Allow users to specify custom prompt files:

```bash
ralphio --prompt-dir ./my-prompts/
ralphio --build-prompt ./PROMPT_custom.md
```

---

## Phase 6: AGENTS.md Integration

### 6a: AGENTS.md Discovery

Orchestrator checks for `AGENTS.md` existence:

```go
func (b *PromptBuilder) hasAgentsMD() bool
```

### 6b: Include in Context

The prompt instructs agent to study `@AGENTS.md` for:
- Build commands
- Test commands
- Project-specific patterns

### 6c: TUI Display

Show AGENTS.md status in dashboard:
- "AGENTS.md: ✓ found" or "✗ missing"
- Quick view on key press

---

## Phase 7: Validation Strategy

### 7a: Agent-Driven Validation

Remove orchestrator's `validationCommand` execution. Agent runs validation per prompt instructions.

### 7b: Validation Status Detection

Parse agent output to detect validation results:

```go
type ValidationSignal struct {
    Passed bool
    Output string
}

func DetectValidation(output string) *ValidationSignal
```

Look for patterns:
- "tests passed" / "all tests pass"
- Exit code indicators
- Build success/failure messages

### 7c: Backpressure Visualization

TUI shows validation status detected from agent stream:
- Green: validation passed
- Red: validation failed, agent retrying
- Yellow: validation running

---

## Phase 8: TUI Updates

### 8a: Dashboard Mode Indicator

```
╔══════════════════════════════════════════════════════╗
║ RALPHIO | Project: auth-service | Mode: BUILDING    ║
╠══════════════════════════════════════════════════════╣
```

### 8b: Plan View from tasks.json

Parse and display `tasks.json` sorted by priority:

```
╠══════════════════════ TASK PLAN ══════════════════════════╣
║ [✓] #1  Setup project scaffold                          ║
║ [✓] #2  Add linting                                     ║
║ [→] #3  Implement auth middleware  ← IN PROGRESS        ║
║ [ ] #4  Add integration tests                           ║
╚══════════════════════════════════════════════════════════╝
```

Status icons: `✓` completed, `→` in_progress, ` ` pending, `✗` failed.

### 8c: File Status Indicators

Show which Ralph files exist:

```
╠════════════════════════ RALPH FILES ═════════════════════╣
║ PROMPT_build.md ✓  PROMPT_plan.md ✓  AGENTS.md ✓       ║
║ tasks.json ✓ (4 tasks)  specs/ (3 files) ✓              ║
╚══════════════════════════════════════════════════════════╝
```

### 8d: Key Binding Updates

| Key | Action |
|-----|--------|
| `m` | Toggle mode (planning/building) |
| `e` | Edit tasks.json in $EDITOR |
| `R` | Regenerate plan (switch to planning mode) |
| `a` | View AGENTS.md |

---

## Phase 9: Crash Recovery

### 9a: State Recovery

On startup, check:
1. `.ralph/state.json` exists
2. `tasks.json` exists
3. Git status (uncommitted changes?)

### 9b: Resume Logic

- If mode was `building` and task in progress → agent will see it in plan
- If mode was `planning` → resume planning iteration
- No manual task state management needed

### 9c: Clean State Option

```bash
ralphio --reset-state
```

Clears `.ralph/` directory for fresh start.

---

## Implementation Order

| Order | Phase | Files | Dependencies |
|-------|-------|-------|--------------|
| 1 | 1c | `internal/state/state.go` | None |
| 2 | 2a | `internal/plan/mode.go` | None |
| 3 | 5a, 5b | `assets/PROMPT_plan.md`, `assets/PROMPT_build.md` | None |
| 4 | 3b | `internal/prompt/builder.go` | Phase 2, 3 |
| 5 | 4a | `internal/plan/parser.go` (JSON) | None |
| 6 | 3a | `internal/orchestrator/orchestrator.go` refactor | Phases 1-5 |
| 7 | 7a, 7b | Remove validation from orchestrator | Phase 6 |
| 8 | 8a-8d | TUI updates | Phases 4, 6 |
| 9 | 6a-6c | AGENTS.md integration | Phase 3 |
| 10 | 9a-9c | Crash recovery | Phase 1 |
| 11 | 10a-10e | UI Refactor — banner, chat panel, status bar | Phase 8 |

---

## File Tree (changes)

```
internal/
  plan/
    mode.go           # NEW: LoopMode constants
    parser.go         # NEW: JSON plan parser (encoding/json over tasks.json)
    plan.go           # MODIFY: Remove NextTask, UpdateTask, SaveTasks
  prompt/
    builder.go        # NEW: Prompt composition from files
  orchestrator/
    orchestrator.go   # MAJOR REFACTOR: Agent-driven loop
    messages.go       # MODIFY: Add mode-related messages
  state/
    state.go          # MODIFY: Add LoopMode field
  validator/
    validator.go      # DEPRECATE or keep for optional use
  ui/
    screens/
      dashboard.go    # MODIFY: Mode indicator, plan view, file status
config/
  config.go           # MODIFY: Add promptDir, ralphMode config
  defaults.go         # MODIFY: Default prompt paths
assets/
  PROMPT_plan.md      # NEW: Default planning prompt
  PROMPT_build.md     # NEW: Default building prompt
cmd/
  root.go             # MODIFY: Add --mode, --reset-state flags
```

---

## Verification

1. **Mode switching**: `ralphio --mode planning` starts in planning mode
2. **Agent task selection**: Orchestrator passes full prompt, agent picks highest-priority pending task from `tasks.json`
3. **Plan mutation**: Agent can add/update/remove tasks in `tasks.json`
4. **TUI reflects plan**: `tasks.json` changes appear in dashboard within one polling cycle
5. **Crash recovery**: Restart resumes from state, agent sees `in_progress` task in `tasks.json`
6. **Two modes work**: Planning regenerates `tasks.json`, building implements from it

---

## Migration Path

No format migration required — `tasks.json` schema is unchanged. Behavioural changes only:

1. Agent now mutates `tasks.json` directly (add/update/remove tasks).
2. Orchestrator no longer selects the next task or sets `retryCount` — the agent manages those fields.
3. Existing `tasks.json` files from previous versions are fully compatible.

---

## Risks

| Risk | Mitigation |
|------|------------|
| Agent doesn't follow prompt | Well-tested prompt templates; user can customize |
| Plan parsing fragile | Standard `encoding/json`; fallback to raw display if invalid JSON |
| Validation detection unreliable | Best-effort detection; agent handles actual validation |
| Breaking change for users | Migration tool; deprecation period |

---

## Success Criteria

- [ ] Agent selects tasks autonomously from `tasks.json`
- [ ] PLANNING mode generates/updates `tasks.json`
- [ ] BUILDING mode implements from `tasks.json`
- [ ] TUI shows live task list from `tasks.json`
- [ ] AGENTS.md loaded for project-specific commands
- [ ] Agent can add/update/remove tasks in `tasks.json`
- [ ] State recovery works after crash

---

## Phase 10: UI Refactor

This phase replaces the dashboard's flat log output with a structured, visually
distinct layout: a persistent ASCII-art banner at the top, a scrollable
chat-style output panel in the middle, and a persistent key-binding status bar
at the bottom. All three regions are composed in a single `layout()` helper and
respond correctly to `tea.WindowSizeMsg`.

---

### 10a: Persistent RALPHIO Banner

The banner occupies the top of the screen at all times, regardless of scroll
position in the output panel below it.

**Rendering**

Use `github.com/lsferreira42/figlet-go` to render the text `"RALPHIO"` into a
large ASCII-art block at startup. Store the rendered string as a field on the
dashboard model so it is computed once, not on every frame.

```go
import figlet "github.com/lsferreira42/figlet-go"

type dashboardModel struct {
    // ...
    bannerText string // pre-rendered figlet output, set in Init
    bannerHeight int  // lipgloss.Height(bannerText) + 1 for sub-line
}

func (m dashboardModel) initBanner() dashboardModel {
    rendered, _ := figlet.RenderStr("RALPHIO", "standard")
    m.bannerText = rendered
    m.bannerHeight = lipgloss.Height(rendered) + 1 // +1 for sub-line
    return m
}
```

**Sub-line**

A single line immediately below the figlet block, composed from live state:

```
Project: auth-service  |  Mode: BUILDING  |  Adapter: claude  |  Iteration: 14
```

The sub-line is formatted in `View()` using the current model fields:

```go
subLine := fmt.Sprintf(
    "Project: %s  |  Mode: %s  |  Adapter: %s  |  Iteration: %d",
    m.projectName,
    strings.ToUpper(string(m.mode)),
    m.adapterName,
    m.iteration,
)
```

**Styling**

All banner colors come from `internal/ui/theme`. Use `charm.land/lipgloss/v2`:

```go
import "charm.land/lipgloss/v2"

bannerStyle := lipgloss.NewStyle().
    Foreground(theme.BannerFg).
    Bold(true)

subLineStyle := lipgloss.NewStyle().
    Foreground(theme.SubLineFg)

dividerStyle := lipgloss.NewStyle().
    Foreground(theme.DividerFg)
```

**Height accounting**

`bannerHeight` is exposed via the `ScreenBase` helper (see 10e). The chat panel
and layout helper read this value so the remaining vertical space is allocated
correctly.

---

### 10b: Chat-like Output Panel

The output panel replaces the current flat log/output area. Agent output and
system messages are rendered as clearly separated turns inside a
`charm.land/bubbles/v2/viewport` scroll region.

**Turn types**

```go
// TurnKind identifies the visual style of an output turn.
type TurnKind int

const (
    TurnAgent  TurnKind = iota // numbered iteration output from the LLM
    TurnSystem                 // validation results, mode changes, loop events
)

type OutputTurn struct {
    Kind      TurnKind
    Iteration int       // 0 for system turns
    Timestamp time.Time
    Lines     []string  // accumulated lines; last may be partial (streaming)
    Streaming bool      // true while this turn is still receiving content
}
```

**Rendering a single turn**

```go
func renderTurn(t OutputTurn, isDark bool, width int) string {
    ld := lipgloss.LightDark(isDark)

    switch t.Kind {
    case TurnAgent:
        label := fmt.Sprintf("[#%d]", t.Iteration)
        ts := t.Timestamp.Format("15:04:05")
        header := lipgloss.NewStyle().
            Foreground(ld(theme.AgentLabelLight, theme.AgentLabelDark)).
            Bold(true).
            Render(label + "  " + ts)
        body := strings.Join(t.Lines, "\n")
        if t.Streaming {
            body += " \u258c" // block cursor indicator
        }
        return lipgloss.JoinVertical(lipgloss.Left, header, body)

    case TurnSystem:
        prefix := "[sys]"
        text := strings.Join(t.Lines, "\n")
        return lipgloss.NewStyle().
            Foreground(ld(theme.SystemMsgLight, theme.SystemMsgDark)).
            Italic(true).
            Render(prefix + "  " + text)
    }
    return ""
}
```

**Viewport integration**

```go
import "charm.land/bubbles/v2/viewport"

type dashboardModel struct {
    // ...
    chatViewport viewport.Model
    turns        []OutputTurn
    autoScroll   bool // reset to true on new content; user scroll sets false
}
```

Rebuild viewport content whenever `turns` changes:

```go
func (m *dashboardModel) rebuildViewport() {
    var sb strings.Builder
    for _, t := range m.turns {
        sb.WriteString(renderTurn(t, m.isDark, m.chatViewport.Width()))
        sb.WriteString("\n\n")
    }
    m.chatViewport.SetContent(sb.String())
    if m.autoScroll {
        m.chatViewport.GotoBottom()
    }
}
```

**Auto-scroll behaviour**

- New `AgentOutputMsg` or `SystemOutputMsg` received: call `rebuildViewport()`
  with `autoScroll = true`.
- User presses `up`, `pgup`, or scrolls the mouse wheel up: set
  `autoScroll = false`.
- User presses `end` or scrolls to the bottom: set `autoScroll = true`.

**Message routing**

```go
case tea.KeyPressMsg:
    switch msg.String() {
    case "up", "pgup":
        m.autoScroll = false
        m.chatViewport, cmd = m.chatViewport.Update(msg)
    case "end":
        m.autoScroll = true
        m.chatViewport.GotoBottom()
    default:
        m.chatViewport, cmd = m.chatViewport.Update(msg)
    }
```

---

### 10c: Status Bar (bottom)

A persistent single-line bar at the very bottom of the terminal showing key
bindings relevant to the current mode.

**Help component**

```go
import "charm.land/bubbles/v2/help"
import "charm.land/bubbles/v2/key"

type dashboardKeyMap struct {
    ToggleMode key.Binding
    EditPlan   key.Binding
    Run        key.Binding
    Help       key.Binding
    Quit       key.Binding
}

func (k dashboardKeyMap) ShortHelp() []key.Binding {
    return []key.Binding{k.ToggleMode, k.EditPlan, k.Run, k.Help, k.Quit}
}

func (k dashboardKeyMap) FullHelp() [][]key.Binding {
    return [][]key.Binding{
        {k.ToggleMode, k.EditPlan, k.Run},
        {k.Help, k.Quit},
    }
}

var defaultKeyMap = dashboardKeyMap{
    ToggleMode: key.NewBinding(key.WithKeys("m"),        key.WithHelp("m", "mode")),
    EditPlan:   key.NewBinding(key.WithKeys("e"),        key.WithHelp("e", "edit plan")),
    Run:        key.NewBinding(key.WithKeys("r"),        key.WithHelp("r", "run")),
    Help:       key.NewBinding(key.WithKeys("?"),        key.WithHelp("?", "help")),
    Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}
```

The `help.Model` is stored on the dashboard model. Its styles are set when
`tea.BackgroundColorMsg` arrives:

```go
case tea.BackgroundColorMsg:
    m.isDark = msg.IsDark()
    m.helpBar.Styles = help.DefaultStyles(m.isDark)
```

The status bar line is always 1 row tall:

```go
const statusBarHeight = 1
```

---

### 10d: Layout Composition

All three regions are composed in a single `layout()` helper method. It is
called from `View()` and whenever `tea.WindowSizeMsg` is received.

**Height allocation**

```
terminalHeight = bannerHeight + chatPanelHeight + statusBarHeight
chatPanelHeight = terminalHeight - bannerHeight - statusBarHeight
```

**`layout()` helper**

```go
// layout recomputes component dimensions from the current terminal size.
// Call this in Update whenever tea.WindowSizeMsg is received, and also
// once during Init after the banner is rendered.
func (m *dashboardModel) layout() {
    chatH := m.height - m.bannerHeight - statusBarHeight
    if chatH < 1 {
        chatH = 1
    }
    m.chatViewport.SetWidth(m.width)
    m.chatViewport.SetHeight(chatH)
    m.helpBar.SetWidth(m.width)
    m.rebuildViewport()
}
```

**Handling resize**

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    m.layout()
```

**`View()` composition**

```go
func (m dashboardModel) View() tea.View {
    banner := lipgloss.JoinVertical(lipgloss.Left,
        bannerStyle.Render(m.bannerText),
        subLineStyle.Render(m.subLine()),
        dividerStyle.Render(strings.Repeat("─", m.width)),
    )

    chat := m.chatViewport.View()

    statusBar := lipgloss.NewStyle().
        Foreground(theme.StatusBarFg).
        Background(theme.StatusBarBg).
        Width(m.width).
        Render(m.helpBar.View(m.keys))

    v := tea.NewView(lipgloss.JoinVertical(lipgloss.Left,
        banner,
        chat,
        statusBar,
    ))
    v.AltScreen = true
    return v
}
```

**Target layout (ASCII mockup)**

```
╔══════════════════════════════════════════════════╗
║  ██████╗  █████╗ ██╗     ██████╗ ██╗  ██╗       ║
║  ██╔══██╗██╔══██╗██║     ██╔══██╗██║  ██║       ║
║  ██████╔╝███████║██║     ██████╔╝███████║       ║
║  ██╔══██╗██╔══██╗██║     ██╔══██╗██╔══██║       ║
║  ██║  ██║██║  ██║███████╗██║  ██║██║  ██║       ║
║  ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═╝       ║
║  Project: auth-service | Mode: BUILDING | #14   ║
╠══════════════════════════════════════════════════╣
║  [sys]  Loop started — iteration 14             ║
║                                                  ║
║  [#14]  Reading specs/auth.md...                ║
║         Implementing JWT middleware...           ║
║         Running go test ./...                   ║
║                                                  ║
║  [sys]  Tests passed                            ║
║                                                  ║
║  [#15]  Selecting next task from tasks.json...  ║
║         (streaming)                             ║
╠══════════════════════════════════════════════════╣
║  m: mode  e: edit plan  r: run  ?: help  q: quit║
╚══════════════════════════════════════════════════╝
```

---

### 10e: Implementation Files

| File | Change type | Summary |
|------|-------------|---------|
| `internal/ui/screens/dashboard.go` | MAJOR MODIFY | Add `bannerText`, `bannerHeight`, `turns`, `chatViewport`, `helpBar`, `keys`, `autoScroll`, `isDark` fields; implement `initBanner()`, `layout()`, `rebuildViewport()`, `renderTurn()`, `subLine()`; replace existing View output with three-region composition |
| `internal/ui/theme/palette.go` | MODIFY | Add color constants: `BannerFg`, `SubLineFg`, `DividerFg`, `AgentLabelLight`, `AgentLabelDark`, `SystemMsgLight`, `SystemMsgDark`, `StatusBarFg`, `StatusBarBg` |
| `internal/ui/screens/screen_base.go` | MODIFY | `ContentHeight()` must subtract `bannerHeight` in addition to any existing fixed chrome; accept banner height as a parameter or read it from a shared layout struct |

**New dependency** — add to `go.mod`:

```
github.com/lsferreira42/figlet-go
```

**Import paths used in this phase**

```go
import (
    tea      "charm.land/bubbletea/v2"
    "charm.land/bubbles/v2/viewport"
    "charm.land/bubbles/v2/help"
    "charm.land/bubbles/v2/key"
    "charm.land/lipgloss/v2"
    figlet   "github.com/lsferreira42/figlet-go"
)
```

# Plan: Create Ralphio App from PRD

## Context

The PRD at `.claude/plans/ralphio.md` describes a "Multi-Client Ralph Workflow TUI" — an autonomous execution engine that orchestrates task selection, AI agent execution, validation (backpressure), plan mutation, and iterative loops. We need to scaffold this as a BubbleTea v2 app called `ralphio` using the `/new-app` command, then implement the full multi-adapter system with JSON stream parsing and model selection.

Reference implementation: `../ralph/src/lib/` (TypeScript) provides patterns for agent commands, stream JSON parsing, and model fetching that we port to Go.

---

## Step 1: Scaffold the App

Run `/new-app ralphio` to create `ralphio/` from the scaffold. This handles:
- Copy scaffold → `ralphio/`
- Rewrite `go.mod` module to `ralphio`
- Update all import paths `scaffold/` → `ralphio/`
- Update CLI name, config, README
- Run `go mod tidy` + `go build`

---

## Step 2: Remove Demo Screens

Delete scaffold demo screens that won't be used:
- `internal/ui/screens/settings.go`
- `internal/ui/screens/filepicker_huh.go`
- `internal/ui/screens/banner_demo.go`

Keep: `base.go`, `form.go`, `detail.go`, `menu_huh.go` (reference patterns).

---

## Step 3: Domain Layer (No BubbleTea dependency)

### 3a: State Management — `internal/state/state.go`
```go
type State struct {
    CurrentIteration int
    CurrentTaskID    string
    LoopStatus       string    // "running", "paused", "stopped", "error"
    ActiveAdapter    string
    ActiveModel      string    // selected model for the adapter
    LastUpdated      time.Time
}
```
- `Load(path)` / `Save(path)` with crash-safe write (tmp + rename)
- Detects interrupted writes on startup

### 3b: Plan Manager — `internal/plan/plan.go`
```go
type Task struct {
    ID, Title, Description string
    Priority               int
    Status                 string // "pending", "in_progress", "completed", "failed", "skipped"
    RetryCount, MaxRetries int
    ValidationCommand      string
}
```
- `LoadTasks()` / `SaveTasks()` from `tasks.json`
- `NextTask()` — first pending task by priority
- `UpdateTask()` — update status and retry count

### 3c: Adapter System — `internal/adapter/`

#### Core Interface — `internal/adapter/adapter.go`
```go
type AgentType string

const (
    AgentClaude   AgentType = "claude"
    AgentCursor   AgentType = "cursor"
    AgentCodex    AgentType = "codex"
    AgentOpencode AgentType = "opencode"
    AgentKilo     AgentType = "kilo"
    AgentPi       AgentType = "pi"
)

type Adapter interface {
    Name() AgentType
    Execute(ctx context.Context, prompt string, onOutput func(text string)) error
    SupportsModelSelection() bool
}

type ModelFetcher interface {
    FetchModels(ctx context.Context) ([]string, error)
}

// Adapters that support model selection implement both Adapter + ModelFetcher
```

Key design: `Execute` takes an `onOutput` callback for streaming text to the TUI in real-time (not just a final result string). This matches how ralph streams agent output.

#### Agent Command Config — `internal/adapter/commands.go`

Port from `ralph/src/lib/services/config/constants.ts`:

```go
type AgentCommandConfig struct {
    Command []string
    Env     map[string]string
}

var AgentCommands = map[AgentType]AgentCommandConfig{
    AgentCursor: {
        Command: []string{"agent", "-p", "--force", "--output-format", "stream-json", "--stream-partial-output"},
    },
    AgentClaude: {
        Command: []string{"claude", "-p", "--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"},
    },
    AgentCodex: {
        Command: []string{"codex", "exec", "--full-auto", "--json"},
    },
    AgentOpencode: {
        Command: []string{"opencode", "run", "--format", "json"},
        Env: map[string]string{"OPENCODE_PERMISSION": `{"*":"allow"}`},
    },
    AgentKilo: {
        Command: []string{"kilo", "run", "--format", "json"},
        Env: map[string]string{"KILO_PERMISSION": `{"*":"allow"}`},
    },
    AgentPi: {
        Command: []string{"pi", "--mode", "json", "-p"},
    },
}

var ValidAgents = []AgentType{AgentCursor, AgentClaude, AgentCodex, AgentOpencode, AgentKilo, AgentPi}
var AgentsSupportingModel = []AgentType{AgentOpencode, AgentKilo, AgentPi}
```

#### JSON Stream Parser — `internal/adapter/stream.go`

Port from `ralph/src/lib/agent-stream.ts`. Parses NDJSON output from agents:

```go
type StreamMessage struct {
    Type    string          `json:"type"`
    Subtype string          `json:"subtype,omitempty"`
    Text    string          `json:"text,omitempty"`
    Result  string          `json:"result,omitempty"`
    Message *MessageContent `json:"message,omitempty"`
    Part    *PartContent    `json:"part,omitempty"`
    AssistantMessageEvent *AssistantEvent `json:"assistantMessageEvent,omitempty"`
}

func ParseStreamLine(line string) string  // returns extracted text or ""
```

Handles these message types (matching the TS implementation):
1. `type:"assistant"` with `message.content[].text` — Claude/Cursor format
2. `type:"result", subtype:"success"` with `result` — final result
3. `type:"text"` with `part.text` — opencode/kilo format
4. `type:"message_update"` with `assistantMessageEvent.delta` — streaming deltas
5. `type:"step_finish"` — ignored (no text)
6. Non-JSON lines returned as-is (plain text output)

#### Model Fetcher — `internal/adapter/models.go`

Port from `ralph/src/lib/model-fetcher.ts`:

```go
func FetchModels(ctx context.Context, agent AgentType) ([]string, error)
```

- `opencode` / `kilo`: run `<agent> models`, split lines
- `pi`: run `pi --list-models`, parse `provider model` table format into `provider/model` strings
- In-memory cache (map keyed by agent type)
- Only callable for agents in `AgentsSupportingModel`

#### Concrete Adapters

Each adapter in its own file, all follow the same pattern:
- **`internal/adapter/claude.go`** — `claude -p --output-format stream-json ...`, appends prompt as final arg
- **`internal/adapter/cursor.go`** — `agent -p --output-format stream-json ...`
- **`internal/adapter/codex.go`** — `codex exec --full-auto --json ...`
- **`internal/adapter/opencode.go`** — `opencode run --format json ...`, implements `ModelFetcher`, accepts `--model` flag
- **`internal/adapter/kilo.go`** — `kilo run --format json ...`, implements `ModelFetcher`, accepts `--model` flag
- **`internal/adapter/pi.go`** — `pi --mode json -p ...`, implements `ModelFetcher`, accepts `--model` flag

Common execution pattern (shared helper):
```go
func runAgentProcess(ctx context.Context, cfg AgentCommandConfig, prompt string, model string, onOutput func(string)) error {
    // 1. Build command args (append prompt, optionally --model)
    // 2. Set env vars from cfg.Env
    // 3. Start process with stdout pipe
    // 4. Scan stdout line by line
    // 5. For each line: ParseStreamLine() → if text != "", call onOutput(text)
    // 6. Wait for process exit, check exit code
}
```

### 3d: Validator — `internal/validator/validator.go`
```go
type Result struct {
    Command  string
    Passed   bool
    Output   string
    Duration time.Duration
}
func RunValidation(ctx context.Context, command string) Result
```
- Runs `exec.CommandContext`, captures output, checks exit code

---

## Step 4: Orchestrator — `internal/orchestrator/`

### 4a: Messages — `internal/orchestrator/messages.go`
Typed `tea.Msg` types: `IterationStartMsg`, `IterationCompleteMsg`, `LogEntryMsg`, `LoopStateMsg`, `LoopErrorMsg`, `LoopPausedMsg`, `LoopResumedMsg`, `AgentOutputMsg` (streaming text from adapter)

### 4b: Core Loop — `internal/orchestrator/orchestrator.go`
- Runs in a goroutine, communicates with TUI via channels
- `msgCh chan<- tea.Msg` — sends state snapshots + streaming output to TUI
- `cmdCh <-chan any` — receives user commands (retry, skip, pause, change adapter, change model)
- Loop: Load state → Select task → Generate prompt → Execute adapter (streaming output via `onOutput` → `AgentOutputMsg`) → Validate → Update plan → Persist → Send snapshot → Repeat
- Checks `ctx.Done()` for clean shutdown
- Supports `ChangeAdapterCmd{Agent, Model}` to hot-swap the active adapter

---

## Step 5: Configuration & CLI

### 5a: Extend Config — `config/config.go`, `config/defaults.go`, `assets/config.default.json`
```json
{
  "ralph": {
    "projectDir": ".",
    "agent": "claude",
    "agentModel": "",
    "maxRetries": 3,
    "retryDelayMs": 5000,
    "agentTimeoutMs": 1800000,
    "iterationDelayMs": 2000,
    "iterations": 10,
    "validation": {
      "enabled": false,
      "commands": ["go build ./...", "go test ./..."],
      "failOnWarning": false
    }
  }
}
```

### 5b: CLI — `cmd/root.go`, `cmd/run.go`
- Add `--project-dir`, `--agent`, `--model` flags to root
- New `run` subcommand that starts the orchestrator + TUI together

---

## Step 6: TUI Screens

### 6a: Key Bindings — `internal/ui/keys/keys.go`
Add: `r` (retry), `s` (skip), `v` (detail), `c` (client), `h` (history), `p` (pause), `m` (change model)

### 6b: Theme — `internal/ui/theme/palette.go`
Add status colors: running (blue), paused (yellow), failed (red), completed (green), skipped (gray)

### 6c: DashboardScreen — `internal/ui/screens/dashboard.go`
Root screen replacing the demo menu. Multi-pane layout:
- Header: iteration count, loop status, task progress (e.g. "3/12"), active agent + model
- Left pane: task list with status icons (`[x]` done, `[>]` active, `[ ]` pending, `[!]` failed)
- Right pane: validation results
- Bottom pane: agent output stream (viewport, auto-scroll) — shows real-time streaming text from adapter
- Help bar with key bindings
- Responsive: collapses to single-column on narrow terminals

### 6d: TaskDetailScreen — `internal/ui/screens/task_detail.go`
Follows `DetailScreen` pattern. Scrollable viewport showing task fields. ESC pops back.

### 6e: HistoryScreen — `internal/ui/screens/history.go`
Scrollable viewport showing iteration history: number, task ID, pass/fail, duration. ESC pops back.

### 6f: AdapterScreen — `internal/ui/screens/adapter_mgmt.go`
Interactive screen using huh form:
- Select adapter from `ValidAgents` list
- If adapter supports model selection, fetch models via `FetchModels()` and show model selector
- Show current config (timeout, etc.)
- On submit: send `ChangeAdapterCmd` to orchestrator
- ESC pops back without changes

### Navigation:
```
DashboardScreen (root)
  ├── v → TaskDetailScreen → ESC → back
  ├── h → HistoryScreen    → ESC → back
  ├── c → AdapterScreen    → ESC → back
  └── m → AdapterScreen (model tab) → ESC → back
```

---

## Step 7: Wire Integration

### 7a: Root Model — `internal/ui/model.go`
- Add `orchCh` and `orchCmdCh` channel fields
- Replace demo menu with DashboardScreen as initial screen
- Handle orchestrator messages in `Update()`, re-subscribe after each
- Forward `AgentOutputMsg` to dashboard for streaming display
- Forward user commands (retry/skip/pause/change adapter) to orchestrator via `orchCmdCh`

### 7b: Entry Point — `main.go`
- Create orchestrator with config
- Create channels for bidirectional communication
- Pass channels to both orchestrator and Model
- Start orchestrator goroutine in `cmd/run.go` handler
- Cancel context on `tea.Quit`

---

## File Tree (new/modified files under `ralphio/`)

```
internal/
  adapter/
    adapter.go          # Adapter + ModelFetcher interfaces, AgentType constants
    commands.go         # AgentCommandConfig map (ported from ralph constants.ts)
    stream.go           # ParseStreamLine — NDJSON parser (ported from agent-stream.ts)
    models.go           # FetchModels — cached model listing (ported from model-fetcher.ts)
    process.go          # runAgentProcess — shared exec helper with streaming
    claude.go           # Claude adapter
    cursor.go           # Cursor adapter
    codex.go            # Codex adapter
    opencode.go         # Opencode adapter (+ ModelFetcher)
    kilo.go             # Kilo adapter (+ ModelFetcher)
    pi.go               # Pi adapter (+ ModelFetcher)
  state/
    state.go            # State struct, crash-safe Load/Save
  plan/
    plan.go             # Task struct, PlanManager CRUD
  validator/
    validator.go        # Run validation commands
  orchestrator/
    messages.go         # All tea.Msg types
    orchestrator.go     # Core loop with channel communication
  ui/
    model.go            # (modify) Wire orchestrator channels, DashboardScreen as root
    keys/keys.go        # (modify) Add ralphio key bindings
    theme/palette.go    # (modify) Add status colors
    screens/
      dashboard.go      # Main monitoring view
      task_detail.go    # Single task view
      history.go        # Iteration history
      adapter_mgmt.go   # Adapter + model selection form
config/
  config.go             # (modify) Add Ralph config section
  defaults.go           # (modify) Add Ralph defaults
assets/
  config.default.json   # (modify) Add Ralph defaults
cmd/
  root.go               # (modify) Add --project-dir, --agent, --model flags
  run.go                # (new) Run subcommand
```

---

## Implementation Order

| Order | Files | Dependencies |
|-------|-------|-------------|
| 1 | `/new-app ralphio` | None |
| 2 | Remove demo screens | Step 1 |
| 3 | `internal/state/state.go` | None |
| 4 | `internal/plan/plan.go` | None |
| 5 | `internal/adapter/adapter.go`, `commands.go` | None |
| 6 | `internal/adapter/stream.go` | Step 5 |
| 7 | `internal/adapter/models.go` | Step 5 |
| 8 | `internal/adapter/process.go` | Steps 5-6 |
| 9 | `internal/adapter/claude.go`, `cursor.go`, `codex.go`, `opencode.go`, `kilo.go`, `pi.go` | Step 8 |
| 10 | `internal/validator/validator.go` | None |
| 11 | `internal/orchestrator/messages.go` | Steps 3-5 |
| 12 | `internal/orchestrator/orchestrator.go` | Steps 9-11 |
| 13 | Config + CLI extensions | Step 1 |
| 14 | Keys + theme extensions | Step 1 |
| 15 | DashboardScreen | Steps 11-14 |
| 16 | TaskDetailScreen, HistoryScreen | Step 14 |
| 17 | AdapterScreen (with model selection form) | Steps 7, 14 |
| 18 | Root model + main.go wiring | Steps 12, 15-17 |

---

## Verification

1. `cd ralphio && go build ./...` — compiles cleanly
2. `go vet ./...` — no issues
3. `golangci-lint run` — passes
4. Create a test `tasks.json` with 2-3 simple tasks and a mock validation command (`echo ok`)
5. Run `./ralphio run --project-dir ./test-project --agent claude` — dashboard renders, loop processes tasks
6. Run `./ralphio run --agent opencode --model anthropic/claude-sonnet-4-20250514` — model flag passed to adapter
7. Test keyboard shortcuts: `v` (detail), `h` (history), `c` (change adapter), `p` (pause/resume), `q` (quit)
8. Test adapter screen: select different agent, verify model list fetches for opencode/kilo/pi
9. Verify streaming: agent output appears line-by-line in dashboard log pane during execution
10. Kill process mid-iteration, restart — verify state recovery from `.ralph/state.json`

A well-structured **Bubble Tea** TUI should make strict separations between:

* **State** (`model`)
* **Events** (`tea.Msg`)
* **Side effects / async work** (`tea.Cmd`)
* **Rendering** (`View()`)

The framework is essentially an Elm Architecture implementation in Go. Below is a production-oriented structure that scales cleanly.

---

# 1. Core Architectural Pattern

Bubble Tea enforces this flow:

```
User Input / Async Result
        ↓
      Msg
        ↓
     Update()
        ↓
 (mutate model, return Cmd)
        ↓
  Cmd runs async
        ↓
  Cmd returns Msg
        ↓
     Update()
```

Key principles:

* `Update()` must be pure state transition logic.
* Async work must happen in `tea.Cmd`.
* Never block inside `Update()`.

---

# 2. Recommended Project Structure

For non-trivial applications:

```
/cmd/app/main.go
/internal/app/model.go
/internal/app/update.go
/internal/app/view.go
/internal/app/msgs.go
/internal/app/cmds.go
/internal/components/...
/internal/services/...
```

You separate:

* domain state
* message definitions
* command factories
* UI components
* external service logic

---

# 3. Model Design

Your `model` should:

* Represent *only* UI state.
* Never directly perform I/O.
* Be deterministic.

Example:

```go
type model struct {
    loading     bool
    err         error
    data        []Item
    spinner     spinner.Model
    viewport    viewport.Model
}
```

Guidelines:

* Keep service clients OUT of the model.
* Keep large mutable caches OUT unless they are UI-visible.
* Compose sub-models for complex UIs.

---

# 4. Message Design (Critical for Scale)

Define explicit message types.

```go
type dataLoadedMsg struct {
    items []Item
    err   error
}

type tickMsg time.Time
```

Patterns:

### ✔ Use typed structs for domain messages

Avoid generic `interface{}` blobs.

### ✔ Separate internal vs external messages

Example:

* `tea.KeyMsg` → external
* `dataLoadedMsg` → internal async result

---

# 5. Commands (Async Tasks Done Correctly)

All async work must be wrapped in `tea.Cmd`.

### Pattern: Service Call

```go
func fetchDataCmd(client *Client) tea.Cmd {
    return func() tea.Msg {
        items, err := client.Fetch()
        return dataLoadedMsg{
            items: items,
            err:   err,
        }
    }
}
```

Key properties:

* Runs in separate goroutine.
* Returns exactly one `Msg`.
* Must not mutate model directly.

---

# 6. Update Function Structure

Production-grade update typically follows:

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case tea.KeyMsg:
        return handleKey(m, msg)

    case dataLoadedMsg:
        m.loading = false
        if msg.err != nil {
            m.err = msg.err
            return m, nil
        }
        m.data = msg.items
        return m, nil

    case tickMsg:
        var cmd tea.Cmd
        m.spinner, cmd = m.spinner.Update(msg)
        return m, cmd
    }

    return m, nil
}
```

Best practice:

* Delegate to smaller functions.
* Keep cases small.
* Avoid 500-line Update functions.

---

# 7. Async Task Patterns

## A. Startup Task

```go
func (m model) Init() tea.Cmd {
    return tea.Batch(
        fetchDataCmd(apiClient),
        spinner.Tick,
    )
}
```

## B. Fire-and-Forget Action

User presses key → trigger async task:

```go
case "r":
    m.loading = true
    return m, fetchDataCmd(apiClient)
```

---

## C. Periodic Tasks

Use `tea.Tick`:

```go
func tickCmd() tea.Cmd {
    return tea.Tick(time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

Return it again inside handler to keep looping.

---

## D. Parallel Tasks

```go
return m, tea.Batch(
    fetchUsersCmd(),
    fetchStatsCmd(),
)
```

---

## E. Long-running Background Worker Pattern

For streaming or subscription-style work:

```go
func listenCmd(ch <-chan Event) tea.Cmd {
    return func() tea.Msg {
        return eventMsg(<-ch)
    }
}
```

Then re-issue the command in handler to keep listening.

---

# 8. Component Architecture (Scaling Properly)

For large apps:

```
root model
 ├── sidebar component
 ├── table component
 ├── status bar component
```

Each component:

* Has its own `Model`
* Has its own `Update`
* Has its own `View`
* Emits its own messages

Pattern:

```go
m.sidebar, cmd = m.sidebar.Update(msg)
cmds = append(cmds, cmd)
```

Use `tea.Batch(cmds...)` to combine.

---

# 9. Error Handling Strategy

Never panic inside a TUI.

Recommended:

* Store `error` in model
* Render error view
* Allow retry

Example:

```go
if m.err != nil {
    return errorView(m.err)
}
```

---

# 10. State Machines (For Complex Flows)

For non-trivial apps:

```go
type viewState int

const (
    stateLoading viewState = iota
    stateReady
    stateError
)
```

This prevents implicit UI transitions.

---

# 11. Cancellation Pattern (Advanced)

Bubble Tea does not automatically cancel commands.

Use `context.Context` manually:

```go
ctx, cancel := context.WithCancel(context.Background())
```

Store `cancel` in model only if absolutely required.

Better approach:

* Move cancellation to service layer
* Emit `cancelledMsg`

---

# 12. Clean Testing Strategy

You can test `Update()` deterministically:

```go
m := initialModel()
newModel, cmd := m.Update(dataLoadedMsg{items: fakeData})
```

You don’t test commands directly — you test the resulting message handling.

---

# 13. Anti-Patterns

Avoid:

* Blocking I/O inside `Update`
* Goroutines that mutate model directly
* Global state
* 1 massive model with 50 fields
* Commands returning multiple messages (not supported)

---

# 14. Minimal Production Template

```go
type model struct {
    loading bool
    data    []Item
    err     error
}

func initialModel() model {
    return model{loading: true}
}

func (m model) Init() tea.Cmd {
    return fetchDataCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    case dataLoadedMsg:
        m.loading = false
        m.data = msg.items
        m.err = msg.err
        return m, nil

    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }

    return m, nil
}

func (m model) View() string {
    if m.loading {
        return "Loading..."
    }
    if m.err != nil {
        return fmt.Sprintf("Error: %v", m.err)
    }
    return renderData(m.data)
}
```

---

# 15. What “Good” Looks Like

A well-designed Bubble Tea app:

* Has a small, deterministic `Update()`
* Treats commands as effect boundaries
* Uses explicit message types
* Composes sub-models
* Avoids shared mutable state
* Is testable without terminal interaction

  ---
  Component Architecture

  The application follows a clear separation of concerns with three main architectural
  layers:

  1. Main Model (internal/app/app.go)

  The Model struct acts as the central coordinator holding all UI components:

  type Model struct {
      // Core UI components
      header  *ui.Header    // Top bar
      footer  *ui.Footer    // Bottom status bar
      sidebar *ui.Sidebar  // Session list (left)
      chat    *ui.Chat     // Chat interface (right)
      modal   *ui.Modal    // Overlay dialogs

      // Layout state
      width, height int
      focus Focus  // FocusSidebar or FocusChat

      // Services (dependency injection)
      activeSession *config.Session
      claudeRunner  claude.RunnerInterface
      sessionMgr    *SessionManager
      gitService    *git.GitService
      sessionService *session.SessionService
      issueRegistry *issues.ProviderRegistry

      // App state machine
      state AppState  // StateIdle or StateStreamingClaude
  }

  2. UI Components (internal/ui/)

  Each component is a standalone tea.Model with its own Update() and View():

  ┌───────────┬────────────────────┬───────────────────────────────────┬─────────────┐
  │ Component │        File        │              Purpose              │    Size     │
  ├───────────┼────────────────────┼───────────────────────────────────┼─────────────┤
  │ Header    │ header.go          │ App title, session name, diff     │ 1 line      │
  │           │                    │ stats, status indicators          │ height      │
  ├───────────┼────────────────────┼───────────────────────────────────┼─────────────┤
  │ Footer    │ footer.go          │ Context-aware keyboard shortcuts, │ 1 line      │
  │           │                    │  flash messages                   │ height      │
  ├───────────┼────────────────────┼───────────────────────────────────┼─────────────┤
  │ Sidebar   │ sidebar.go         │ Session list with tree structure, │ 1/5 of      │
  │           │                    │  search, multi-select             │ width       │
  ├───────────┼────────────────────┼───────────────────────────────────┼─────────────┤
  │ Chat      │ chat.go +          │ Message viewport, input textarea, │ 4/5 of      │
  │           │ chat_render.go     │  typing animation                 │ width       │
  ├───────────┼────────────────────┼───────────────────────────────────┼─────────────┤
  │ Modal     │ modal.go           │ Base modal with state             │ Centered    │
  │           │                    │ polymorphism                      │ overlay     │
  └───────────┴────────────────────┴───────────────────────────────────┴─────────────┘

  3. Modal System (internal/ui/modals/)

  The modal package uses type-safe state polymorphism:

  // Base interface all modals implement
  type ModalState interface {
      Update(tea.Msg) (ModalState, tea.Cmd)
      Render() string
  }

  // Optional interfaces for size negotiation
  type ModalWithPreferredWidth interface {
      PreferredWidth() int
  }

  type ModalWithSize interface {
      SetSize(width, height int)
  }

  The ui.Modal wrapper delegates to the current state:
  type Modal struct {
      State ModalState  // nil = hidden
      error string
  }

  func (m *Modal) View(w, h int) string {
      // Check for preferred width
      if mw, ok := m.State.(ModalWithPreferredWidth); ok {
          width = mw.PreferredWidth()
      }
      // Notify modal of actual size
      if ms, ok := m.State.(ModalWithSize); ok {
          ms.SetSize(width, screenHeight)
      }
      // Center overlay
      return lipgloss.Place(w, h, Center, Center, content)
  }

  4. Message Routing

  The main model routes messages to appropriate components:

  func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      var cmds []tea.Cmd

      // Route to focused panel
      if m.focus == FocusSidebar {
          m.sidebar, cmd = m.sidebar.Update(msg)
      } else {
          m.chat, cmd = m.chat.Update(msg)
      }
      cmds = append(cmds, cmd)

      // Modals always get messages first
      m.modal, cmd = m.modal.Update(msg)
      cmds = append(cmds, cmd)

      return m, tea.Batch(cmds...)
  }

  ---
  Layout System

  The layout uses a centralized calculation system to ensure consistency.

  Visual Structure

  ┌──────────────────────────────────────────────────────────────────┐
  │  Header (fixed: 1 line)                                    │
  ├──────────────────────┬───────────────────────────────────────────┤
  │                      │                                           │
  │    Sidebar           │           Chat Panel                      │
  │    (1/5 width)       │           (4/5 width)                     │
  │                      │                                           │
  │  - Session list     │  ┌─────────────────────────────────┐   │
  │  - Tree structure  │  │  Message viewport               │   │
  │  - Status icons     │  │  (scrollable)                   │   │
  │                      │  │                                 │   │
  │                      │  ├─────────────────────────────────┤   │
  │                      │  │  Input textarea (3 lines)       │   │
  │                      │  └─────────────────────────────────┘   │
  │                      │                                           │
  ├──────────────────────┴───────────────────────────────────────────┤
  │  Footer (fixed: 1 line)                                    │
  └──────────────────────────────────────────────────────────────────┘

  ViewContext Singleton (internal/ui/context.go)

  A singleton provides centralized dimension calculations:

  type ViewContext struct {
      TerminalWidth, TerminalHeight int  // From terminal resize
      HeaderHeight, FooterHeight    int  // Fixed constants (1 line each)
      ContentHeight                  int  // Calculated: height - header - footer
      SidebarWidth, ChatWidth        int  // Calculated: width/5, width - sidebar
  }

  func (v *ViewContext) UpdateTerminalSize(w, h int) {
      v.TerminalWidth = w
      v.TerminalHeight = h
      v.HeaderHeight = HeaderHeight   // = 1
      v.FooterHeight = FooterHeight   // = 1
      v.ContentHeight = h - HeaderHeight - FooterHeight
      v.SidebarWidth = w / SidebarWidthRatio  // = 1/5
      v.ChatWidth = w - v.SidebarWidth          // = 4/5
  }

  Main View Assembly (internal/app/view.go)

  func (m *Model) View() tea.View {
      // 1. Configure view properties
      v := tea.NewView("")
      v.AltScreen = true
      v.MouseMode = tea.MouseModeCellMotion
      v.ReportFocus = true

      // 2. Update component sizes
      ctx := ui.GetViewContext()
      m.header.SetWidth(ctx.TerminalWidth)
      m.footer.SetWidth(ctx.TerminalWidth)
      m.sidebar.SetSize(ctx.SidebarWidth, ctx.ContentHeight)
      m.chat.SetSize(ctx.ChatWidth, ctx.ContentHeight)

      // 3. Render each component
      header := m.header.View()
      footer := m.footer.View()
      sidebarView := m.sidebar.View()
      chatView := m.chat.View()

      // 4. Assemble panels horizontally
      panels := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, chatView)

      // 5. Assemble final view vertically
      view := lipgloss.JoinVertical(lipgloss.Left, header, panels, footer)

      // 6. Overlay modal if visible (centered)
      if m.modal.IsVisible() {
          modalView := m.modal.View(m.width, m.height)
          v.SetContent(lipgloss.Place(m.width, m.height, Center, Center, modalView))
      } else {
          v.SetContent(view)
      }

      return v
  }

  Layout Constants (internal/ui/constants.go)

  All magic numbers have documented rationale:

  const (
      HeaderHeight = 1          // Single line for app title
      FooterHeight = 1          // Single line for shortcuts
      SidebarWidthRatio = 5      // Sidebar gets 1/5, chat gets 4/5
      TextareaHeight = 3          // Input area height in lines
      BorderSize = 2              // Top + bottom border padding
      MinTerminalWidth = 40       // Prevent negative calculations
      MinTerminalHeight = 10      // Minimum usable height

      // Modal widths
      ModalWidth = 80            // Traditional terminal width
      ModalWidthWide = 120         // For complex content like issue lists
      ModalInputWidth = 72         // Fits within 80-char modal with padding

      // Visibility limits for scrolling
      HelpModalMaxVisible = 18
      IssuesModalMaxVisible = 10
      SearchModalMaxVisible = 8
  )

  Focus System

  Focus switches between FocusSidebar and FocusChat via Tab key:

  type Focus int
  const (
      FocusSidebar Focus = iota
      FocusChat
  )

  // Footer context updates based on focus
  m.footer.SetContext(
      hasSession,              // Any session selected?
      sidebarFocused,           // Which panel has focus?
      hasPendingPermission,     // Permission prompt active?
      hasPendingQuestion,      // Question prompt active?
      isStreaming,             // Claude currently streaming?
      viewChangesMode,         // Preview mode active?
      searchMode,              // Sidebar search active?
      multiSelectMode,          // Multi-select active?
      hasDetectedOptions,      // Numbered options detected?
  )

  ---
  Key Architectural Patterns

  1. Type-safe modal polymorphism - All modals implement ModalState interface; optional
  PreferredWidth()/SetSize() for size negotiation
  2. Centralized layout - ViewContext singleton ensures consistent sizing across all
  components
  3. Focus routing - Main model delegates keyboard events to focused panel only
  4. Gradient rendering - Header manually computes per-character gradient from theme
  colors
  5. Hash-based change detection - Sidebar uses FNV hash to avoid expensive tree rebuilds
  when data hasn't changed
  6. Mouse coordinate translation - Mouse events are transformed for panel-relative
  coordinates (adjustMouseForChat)
  7. Named constants - All layout values have documented rationale, no magic numbers
  8. Service injection - Dependencies passed as interfaces for testability

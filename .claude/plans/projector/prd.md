
# Product Requirements Document: Projector

Version: 1.0  
Date: 2026-02-20  
Status: Draft

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 1. Overview

Projector is a terminal user interface (TUI) application for developers to manage and monitor multiple Git-based projects from a single dashboard. Built on the BubbleTea v2 scaffold, it scans a configured directory for Git repositories and provides real-time status, metadata, and quick access to external tools. A scaffold is setup in ./projector/

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 2. Goals & Objectives

### Primary Goals
- Provide instant visibility into all active projects in a developer's workspace
- Display comprehensive Git status (branch, changes, push/pull state, last commit)
- Enable quick navigation and filtering across projects
- Support launching projects in terminal or editor with minimal friction

### Success Metrics
- Scan and display 50+ projects in <2 seconds
- Support filtering/grouping with <100ms response time
- Zero-config startup with sensible defaults

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 3. User Personas

Primary: Solo developer or tech lead managing 10-100 local repositories across multiple languages/frameworks.

Use Cases:
- Morning standup: quickly check which projects have uncommitted work
- Context switching: find and open a specific project by name or language
- Status overview: see which projects are ahead/behind remote branches

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 4. Functional Requirements

### 4.1 Project Discovery

FR-1.1: Scan immediate subdirectories (depth 1) of a configured root directory  
FR-1.2: Identify projects by presence of .git directory  
FR-1.3: Support both:
- Settings screen to change scan directory (persistent to config)
- Command-line flag --projects-dir <path> to override on launch

### 4.2 Git Status Information

FR-2.1: Display for each project:
- Current branch name
- Uncommitted changes count (staged + unstaged)
- Unpushed commits count
- Unpulled commits count (ahead/behind remote)
- Last commit message (truncated to 60 chars)
- Last commit timestamp (relative, e.g., "2 hours ago")
- Remote tracking status (up-to-date, ahead, behind, diverged, no remote)

FR-2.2: Visual indicators:
- Clean status (green checkmark)
- Dirty status (yellow dot)
- Unpushed commits (orange up arrow)
- Unpulled commits (blue down arrow)
- Diverged (red warning icon)

### 4.3 Project Metadata

FR-3.1: Detect language/framework from:
- package.json → Node.js
- go.mod → Go
- Cargo.toml → Rust
- pyproject.toml / requirements.txt → Python
- pom.xml / build.gradle → Java
- Gemfile → Ruby
- .csproj / .sln → C#

FR-3.2: Display:
- Last modified timestamp (most recent file change in project root)
- Project size (disk usage, human-readable: KB/MB/GB)

### 4.4 Navigation & UI

FR-4.1: Primary view: List + Dashboard hybrid
- Scrollable list of projects (arrow keys)
- Each row shows: name, language icon, git status icons, branch, last commit
- Press Enter on a project → detail panel (dashboard card) overlays or replaces list

FR-4.2: Detail view shows:
- Full project path
- All git status details (FR-2.1)
- Language/framework
- Last modified, size
- Available actions (FR-5)

FR-4.3: Search/filter bar (toggle with / key):
- Fuzzy search by project name
- Filter by language (e.g., /lang:go)
- Filter by git status (e.g., /dirty, /unpushed)

FR-4.4: Grouping modes (toggle with g key, cycle through):
- Flat list (default)
- Group by parent directory
- Group by language
- Group by git status (clean, dirty, unpushed, unpulled, diverged)

### 4.5 Actions

FR-5.1: Read-only status display (default view)

FR-5.2: Open in external tools (from detail view or via hotkey):
- o → Open in terminal (spawn $SHELL in project directory)
- e → Open in editor (spawn $EDITOR or configured command)

FR-5.3: Command configuration:
- Detect $EDITOR and $SHELL from environment
- Allow override in config file:
 json
  {
    "commands": {
      "editor": "code .",
      "terminal": "wezterm start --cwd"
    }
  }
  

### 4.6 Refresh Behavior

FR-6.1: Auto-refresh when TUI gains focus (detect terminal focus events if supported, fallback to manual)  
FR-6.2: Manual refresh with r key  
FR-6.3: Show "Scanning..." loading state during background scan

### 4.7 Performance

FR-7.1: Asynchronous scanning:
- Display cached results immediately on startup
- Scan in background, update UI incrementally as projects are discovered
- Show spinner/progress indicator during scan

FR-7.2: Git operations run concurrently (goroutine pool, max 10 concurrent)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 5. Non-Functional Requirements

### 5.1 Performance
- Initial render: <500ms (with cached data)
- Full scan of 100 projects: <3 seconds
- UI responsiveness: <16ms frame time (60 FPS)

### 5.2 Compatibility
- Linux, macOS, Windows (via WSL or native)
- Terminals: iTerm2, Alacritty, WezTerm, Windows Terminal, Kitty

### 5.3 Configuration
- Config file: ~/.projector.json (extends scaffold's existing config schema)
- New fields:
 json
  {
    "projectsDir": "~/projects",
    "commands": {
      "editor": "$EDITOR",
      "terminal": "$SHELL"
    },
    "scan": {
      "autoRefreshOnFocus": true
    }
  }
  

### 5.4 Logging
- Debug mode (--debug) logs:
  - Scan duration per project
  - Git command execution time
  - Errors (missing remotes, detached HEAD, etc.)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 6. User Interface Specification

### 6.1 Main Screen (List View)

┌─ Projector ────────────────────────────────────────────────────────┐
│ 📁 ~/projects (42 projects)                    [Filter: /unpushed] │
├────────────────────────────────────────────────────────────────────┤
│ ● myapp              🟢 main ↑2        "Fix login bug"    2h ago   │
│ ● api-server      🐹 🟡 dev  ↓1 ↑3    "Add metrics"      1d ago   │
│ ● frontend        ⚛️ 🔴 feat ⚠️       "WIP: redesign"    3h ago   │
│   ...                                                               │
├────────────────────────────────────────────────────────────────────┤
│ ↑/↓: navigate  Enter: details  /: filter  g: group  o: open  q: quit│
└────────────────────────────────────────────────────────────────────┘


### 6.2 Detail View (Dashboard Card)

┌─ api-server ───────────────────────────────────────────────────────┐
│ Path: ~/projects/api-server                                        │
│ Language: Go (go.mod)                                              │
│ Size: 45.2 MB  Modified: 6 hours ago                               │
│                                                                     │
│ Git Status:                                                        │
│   Branch: dev                                                      │
│   Remote: origin/dev (behind 1, ahead 3)                           │
│   Uncommitted: 5 files (3 staged, 2 unstaged)                     │
│   Last commit: "Add metrics endpoint" (John Doe, 1 day ago)       │
│                                                                     │
│ Actions:                                                           │
│   [o] Open in terminal                                             │
│   [e] Open in editor                                               │
├────────────────────────────────────────────────────────────────────┤
│ Esc: back  o: terminal  e: editor                                  │
└────────────────────────────────────────────────────────────────────┘


### 6.3 Settings Screen

Extend existing scaffold settings with:
- Projects Directory (text input with path validation)
- Editor Command (text input, default: $EDITOR)
- Terminal Command (text input, default: $SHELL)
- Auto-refresh on focus (toggle)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 7. Technical Architecture

### 7.1 New Packages

internal/
  projector/
    scanner.go          # Directory scanning, .git detection
    git.go              # Git status queries (branch, diff, log)
    metadata.go         # Language detection, size calculation
    project.go          # Project struct, aggregated data
  ui/screens/
    projects_list.go    # Main list view (FR-4.1)
    project_detail.go   # Detail card (FR-4.2)
    projects_settings.go # Extend settings screen


### 7.2 Data Model

go
type Project struct {
    Name          string
    Path          string
    Language      string        // "Go", "Node.js", etc.
    Size          int64         // bytes
    LastModified  time.Time
    Git           GitStatus
}

type GitStatus struct {
    Branch          string
    Remote          string
    Uncommitted     int         // staged + unstaged
    Unpushed        int
    Unpulled        int
    LastCommitMsg   string
    LastCommitTime  time.Time
    LastCommitAuthor string
    Status          StatusType  // Clean, Dirty, Ahead, Behind, Diverged
}


### 7.3 Concurrency

- Scanner spawns goroutines per project (pooled, max 10)
- Each goroutine:
  1. Checks .git existence
  2. Runs git status --porcelain, git log -1, git rev-list @{u}..HEAD, etc.
  3. Detects language files
  4. Calculates size
  5. Sends result to channel → UI updates incrementally

### 7.4 Caching

- Cache scan results in memory (map[string]Project)
- Persist to ~/.cache/projector/projects.json on exit
- Load cache on startup for instant display

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 8. Implementation Phases

### Phase 1: Core Scanning (MVP)
- FR-1.1, FR-1.2: Directory scanning, .git detection
- FR-2.1: Basic git status (branch, uncommitted, unpushed)
- FR-4.1: Simple list view
- FR-6.2: Manual refresh

### Phase 2: Enhanced Status & Metadata
- FR-2.1: Full git status (unpulled, last commit, remote tracking)
- FR-3.1, FR-3.2: Language detection, size, last modified
- FR-4.2: Detail view

### Phase 3: Filtering & Grouping
- FR-4.3: Search/filter bar
- FR-4.4: Grouping modes

### Phase 4: Actions & Configuration
- FR-5.2, FR-5.3: Open in terminal/editor
- FR-1.3: Settings screen + CLI flag for projects directory
- FR-6.1: Auto-refresh on focus

### Phase 5: Performance & Polish
- FR-7.1, FR-7.2: Async scanning, caching
- Visual refinements (icons, colors, animations)
- Error handling (missing git, permission errors)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 9. Open Questions

1. Submodules: Should projects with submodules show nested status?  
 Recommendation: Phase 2+ feature, show indicator only in Phase 1

2. Non-Git projects: Should non-Git directories be shown (e.g., "Not a repo")?  
 Recommendation: No, focus on Git projects only

3. Remote operations: Should we fetch from remote to get accurate ahead/behind?  
 Recommendation: No auto-fetch (slow), show last known state, add manual "Fetch All" action in Phase 4

4. Multi-root support: Allow multiple scan directories simultaneously?  
 Recommendation: Phase 4+, single directory for MVP

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 10. Success Criteria

- [ ] Scans 50 projects in <2 seconds
- [ ] Displays all FR-2.1 git status fields accurately
- [ ] Filter by language and git status works with <100ms latency
- [ ] Opens project in terminal/editor without errors on Linux/macOS
- [ ] Config file changes persist across restarts
- [ ] Zero crashes on malformed Git repos or permission errors

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 11. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Slow git operations on large repos | High | Timeout git commands (5s), show "Timeout" status |
| Missing git binary | High | Check git --version on startup, show error screen |
| Permission errors on scan | Medium | Skip inaccessible directories, log to debug.log |
| Terminal focus detection unsupported | Low | Fallback to manual refresh only |

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━


## 12. Appendix

### A. Language Detection Rules

| File | Language |
|------|----------|
| package.json | Node.js |
| go.mod | Go |
| Cargo.toml | Rust |
| pyproject.toml, requirements.txt | Python |
| pom.xml, build.gradle | Java |
| Gemfile | Ruby |
| .csproj, .sln | C# |

Priority: First match wins (check in order above).

### B. Git Commands

bash
# Branch name
git rev-parse --abbrev-ref HEAD

# Uncommitted changes
git status --porcelain | wc -l

# Unpushed commits
git rev-list @{u}..HEAD --count

# Unpulled commits
git rev-list HEAD..@{u} --count

# Last commit
git log -1 --format="%s|%ar|%an"

# Remote tracking branch
git rev-parse --abbrev-ref @{u}


### C. Config Schema Extension

json
{
  "logLevel": "info",
  "debug": false,
  "ui": { ... },
  "app": { ... },
  "projector": {
    "projectsDir": "~/projects",
    "commands": {
      "editor": "$EDITOR",
      "terminal": "$SHELL"
    },
    "scan": {
      "autoRefreshOnFocus": true,
      "concurrency": 10,
      "gitTimeout": 5
    }
  }
}

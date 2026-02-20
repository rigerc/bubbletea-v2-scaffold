# Ralph Tasks

## Phase 1: Core Scanning (MVP)

- [x] Create internal/projector package with Project and GitStatus structs (data model)
- [x] Implement scanner.go - directory scanning and .git detection
- [x] Implement git.go - basic git status queries (branch, uncommitted, unpushed counts)
- [x] Create projects_list.go screen with scrollable list of projects
- [x] Add manual refresh with 'r' key
- [x] Integrate scanner with UI - display projects with name, branch, status icons
- [x] Add --projects-dir CLI flag to override scan directory
- [x] Extend config with projector settings (projectsDir, commands, scan options)

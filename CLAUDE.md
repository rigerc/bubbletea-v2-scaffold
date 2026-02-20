# CLAUDE.md — bubbletea-v2-scaffold

This file is read automatically by Claude Code at session start. It provides
project context, conventions, and working instructions.

---

## Project Overview

This repository contains two things:

1. **Claude Code skills** (`.claude/`) — curated knowledge packages for the
   Charm TUI ecosystem, served via MCP to give Claude accurate API references
   and patterns for libraries where hallucination risk is high.

2. **A project scaffold** (`scaffold/`) — a production-ready BubbleTea v2
   application template that demonstrates the correct usage of every library
   covered by the skills.

---

## Repository Layout

```
.claude/
  agents/bubbletea-dev.md   Specialist TUI agent (auto-activates on TUI tasks)
  commands/bt-dev.md        /bt-dev slash command — loads the TUI skill suite
  plans/                    Archived architecture planning docs (read-only)
  skills/                   Skill definitions served via MCP
scaffold/                   BubbleTea v2 project scaffold
  README.md                 Full scaffold developer guide
AGENTS.md                   Engineering rules for AI agents in this repo
CLAUDE.md                   This file
README.md                   Repository overview
```

---

## Skills

Load skills before working on any code in this repo:

| Task | Command |
|---|---|
| Any TUI / BubbleTea work | `/bt-dev` (loads go-styleguide, bubbletea-v2, bubbles-v2, lipgloss-v2, huh-v2) |
| CLI flags / subcommands | also load `cobra` skill |
| Configuration (koanf) | also load `koanf` skill |
| Logging (zerolog) | also load `zerolog` skill |

The `bubbletea-dev` agent loads these automatically. If working outside that
agent, invoke `/bt-dev` manually at the start of the session.

---

- Do not simplify or remove existing code just to fix a lint warning;
  report the issue instead
- Do not commit secrets or API keys

@AGENTS.md

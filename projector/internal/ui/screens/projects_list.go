package screens

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"projector/internal/projector"
	appkeys "projector/internal/ui/keys"
	"projector/internal/ui/nav"
	"projector/internal/ui/theme"
)

type projectsHelpKeys struct {
	app     appkeys.GlobalKeyMap
	refresh key.Binding
	enter   key.Binding
}

func (k projectsHelpKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.refresh, k.enter, k.app.Back, k.app.Help}
}

func (k projectsHelpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.refresh, k.enter, k.app.Back},
		{k.app.Help, k.app.Quit},
	}
}

type ScanCompleteMsg struct {
	projects []projector.Project
	err      error
}

type ProjectsListScreen struct {
	ScreenBase
	projects    []projector.Project
	selectedIdx int
	scanning    bool
	ready       bool
	projectsDir string
	scanner     *projector.Scanner
	filterText  string
	filtering   bool
}

func NewProjectsListScreen(projectsDir string, isDark bool, appName string) *ProjectsListScreen {
	return &ProjectsListScreen{
		ScreenBase:  NewBase(isDark, appName),
		projectsDir: projectsDir,
	}
}

func (s *ProjectsListScreen) Init() tea.Cmd {
	if s.scanner != nil {
		s.scanning = true
		return s.scanCmd()
	}
	return nil
}

func (s *ProjectsListScreen) scanCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := s.scanner.Scan()
		return ScanCompleteMsg{projects: projects, err: err}
	}
}

func (s *ProjectsListScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height
		s.ready = true

	case tea.KeyPressMsg:
		if s.filtering {
			return s.handleFilterInput(msg)
		}
		return s.handleNormalInput(msg)

	case ScanCompleteMsg:
		s.scanning = false
		if msg.err != nil {
			s.projects = nil
		} else {
			s.projects = msg.projects
		}
		if s.selectedIdx >= len(s.filteredProjects()) {
			s.selectedIdx = max(0, len(s.filteredProjects())-1)
		}
	}

	return s, nil
}

func (s *ProjectsListScreen) handleFilterInput(msg tea.KeyPressMsg) (nav.Screen, tea.Cmd) {
	switch msg.String() {
	case "esc":
		s.filtering = false
		s.filterText = ""
	case "enter":
		s.filtering = false
	case "backspace":
		if len(s.filterText) > 0 {
			s.filterText = s.filterText[:len(s.filterText)-1]
		}
	default:
		if len(msg.String()) == 1 {
			s.filterText += msg.String()
		}
	}
	return s, nil
}

func (s *ProjectsListScreen) handleNormalInput(msg tea.KeyPressMsg) (nav.Screen, tea.Cmd) {
	filtered := s.filteredProjects()

	switch msg.String() {
	case "up", "k":
		if s.selectedIdx > 0 {
			s.selectedIdx--
		}
	case "down", "j":
		if s.selectedIdx < len(filtered)-1 {
			s.selectedIdx++
		}
	case "r":
		if !s.scanning && s.scanner != nil {
			s.scanning = true
			return s, s.scanCmd()
		}
	case "/":
		s.filtering = true
		s.filterText = ""
	case "enter":
		if len(filtered) > 0 && s.selectedIdx < len(filtered) {
			return s, nav.Push(NewProjectDetailScreen(filtered[s.selectedIdx], s.IsDark, s.AppName))
		}
	case "esc":
		return s, nav.Pop()
	case "?":
		s.Help.ShowAll = !s.Help.ShowAll
	}

	return s, nil
}

func (s *ProjectsListScreen) filteredProjects() []projector.Project {
	if s.filterText == "" {
		return s.projects
	}
	var filtered []projector.Project
	for _, p := range s.projects {
		if strings.Contains(strings.ToLower(p.Name), strings.ToLower(s.filterText)) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (s *ProjectsListScreen) View() string {
	if !s.ready {
		return "Loading..."
	}

	var header string
	if s.scanning {
		header = s.Theme.Title.Render(s.AppName + " (scanning...)")
	} else {
		header = s.HeaderView()
	}

	filtered := s.filteredProjects()
	var content strings.Builder

	if s.filtering {
		content.WriteString(s.Theme.Subtle.Render("Filter: ") + s.filterText + "█\n\n")
	} else if s.filterText != "" {
		content.WriteString(s.Theme.Subtle.Render("Filter: ") + s.filterText + "\n\n")
	}

	if len(filtered) == 0 {
		if s.scanning {
			content.WriteString(s.Theme.Subtle.Render("Scanning for projects..."))
		} else {
			content.WriteString(s.Theme.Subtle.Render("No projects found."))
		}
	} else {
		for i, p := range filtered {
			line := s.renderProjectLine(p, i == s.selectedIdx)
			content.WriteString(line)
			content.WriteString("\n")
		}
	}

	helpKeys := projectsHelpKeys{
		app: s.Keys,
		refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "details"),
		),
	}

	footer := s.Theme.Subtle.Padding(0, 1).Render(fmt.Sprintf("%d project(s)", len(filtered)))

	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			header,
			content.String(),
			footer,
			s.RenderHelp(helpKeys),
		),
	)
}

func (s *ProjectsListScreen) renderProjectLine(p projector.Project, selected bool) string {
	var parts []string

	if selected {
		parts = append(parts, s.Theme.Status.Render("▸"))
	} else {
		parts = append(parts, " ")
	}

	parts = append(parts, p.Name)

	if p.Git.Branch != "" {
		branchStyle := s.Theme.Subtle
		parts = append(parts, branchStyle.Render("["+p.Git.Branch+"]"))
	}

	status := formatGitStatus(p.Git, s.Theme)
	if status != "" {
		parts = append(parts, status)
	}

	if p.Git.LastCommitMsg != "" {
		msg := p.Git.LastCommitMsg
		if len(msg) > 30 {
			msg = msg[:27] + "..."
		}
		parts = append(parts, s.Theme.Subtle.Render(`"`+msg+`"`))
	}

	if !p.Git.LastCommitTime.IsZero() {
		parts = append(parts, s.Theme.Subtle.Render(formatTimeAgo(p.Git.LastCommitTime)))
	}

	return strings.Join(parts, " ")
}

func (s *ProjectsListScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
}

func (s *ProjectsListScreen) SetScanner(scanner *projector.Scanner) {
	s.scanner = scanner
}

func formatGitStatus(g projector.GitStatus, t theme.Theme) string {
	var parts []string

	switch g.Status {
	case projector.StatusClean:
		parts = append(parts, t.Success.Render("✓"))
	case projector.StatusDirty:
		parts = append(parts, t.Warning.Render("●"))
	case projector.StatusAhead:
		parts = append(parts, t.Status.Render("↑"))
	case projector.StatusBehind:
		parts = append(parts, t.Status.Render("↓"))
	case projector.StatusDiverged:
		parts = append(parts, t.Error.Render("⚠"))
	case projector.StatusNoRemote:
		parts = append(parts, t.Subtle.Render("○"))
	}

	if g.Uncommitted > 0 {
		parts = append(parts, t.Warning.Render(fmt.Sprintf("±%d", g.Uncommitted)))
	}
	if g.Unpushed > 0 {
		parts = append(parts, t.Status.Render(fmt.Sprintf("↑%d", g.Unpushed)))
	}
	if g.Unpulled > 0 {
		parts = append(parts, t.Status.Render(fmt.Sprintf("↓%d", g.Unpulled)))
	}

	return strings.Join(parts, " ")
}

func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 min ago"
		}
		return fmt.Sprintf("%d mins ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

type ProjectDetailScreen struct {
	ScreenBase
	project projector.Project
	ready   bool
}

func NewProjectDetailScreen(project projector.Project, isDark bool, appName string) *ProjectDetailScreen {
	return &ProjectDetailScreen{
		ScreenBase: NewBase(isDark, appName),
		project:    project,
	}
}

func (s *ProjectDetailScreen) Init() tea.Cmd {
	return nil
}

func (s *ProjectDetailScreen) Update(msg tea.Msg) (nav.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.Width, s.Height = msg.Width, msg.Height
		s.ready = true

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			return s, nav.Pop()
		case "?":
			s.Help.ShowAll = !s.Help.ShowAll
		}
	}
	return s, nil
}

func (s *ProjectDetailScreen) View() string {
	if !s.ready {
		return "Loading..."
	}

	var lines []string
	lines = append(lines, s.Theme.Title.Render(s.project.Name))
	lines = append(lines, "")
	lines = append(lines, s.Theme.Subtle.Render("Path: ")+s.project.Path)

	if s.project.Language != "" {
		lines = append(lines, s.Theme.Subtle.Render("Language: ")+s.project.Language)
	}

	lines = append(lines, "")
	lines = append(lines, s.Theme.Status.Render("Git Status:"))
	lines = append(lines, fmt.Sprintf("  Branch: %s", s.project.Git.Branch))

	if s.project.Git.Remote != "" {
		lines = append(lines, fmt.Sprintf("  Remote: %s", s.project.Git.Remote))
	} else {
		lines = append(lines, "  Remote: (none)")
	}

	if s.project.Git.Uncommitted > 0 {
		lines = append(lines, fmt.Sprintf("  Uncommitted: %d file(s)", s.project.Git.Uncommitted))
	}

	if s.project.Git.Unpushed > 0 {
		lines = append(lines, fmt.Sprintf("  Unpushed: %d commit(s)", s.project.Git.Unpushed))
	}

	if s.project.Git.Unpulled > 0 {
		lines = append(lines, fmt.Sprintf("  Unpulled: %d commit(s)", s.project.Git.Unpulled))
	}

	if s.project.Git.LastCommitMsg != "" {
		lines = append(lines, fmt.Sprintf("  Last commit: %q", s.project.Git.LastCommitMsg))
		if s.project.Git.LastCommitAuthor != "" {
			lines = append(lines, fmt.Sprintf("  Author: %s", s.project.Git.LastCommitAuthor))
		}
	}

	content := strings.Join(lines, "\n")

	return s.Theme.App.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			s.HeaderView(),
			content,
			s.RenderHelp(s.Keys),
		),
	)
}

func (s *ProjectDetailScreen) SetTheme(isDark bool) {
	s.ApplyTheme(isDark)
}

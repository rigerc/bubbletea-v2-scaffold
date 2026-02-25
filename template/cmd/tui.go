package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"test-app/banner"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

type phase int

const (
	phaseSelect phase = iota
	phaseStream
	phaseDone
)

type streamMsg struct{}

type model struct {
	phase     phase
	options   []string
	cursor    int
	selected  string
	received  []string
	pending   []string
	isDark    bool
	width     int
	bannerStr string
}

var incomingMessages = []string{
	"Connecting to server...",
	"Authentication successful.",
	"Loading user profile...",
	"Fetching latest data...",
	"Processing chunk 1 / 5",
	"Processing chunk 2 / 5",
	"Processing chunk 3 / 5",
	"Processing chunk 4 / 5",
	"Processing chunk 5 / 5",
	"All done! Stream complete.",
}

func tickCmd() tea.Cmd {
	return tea.Tick(280*time.Millisecond, func(time.Time) tea.Msg {
		return streamMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.isDark = msg.IsDark()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		cfg := banner.Config{
			Text:          "TEST-APP",
			Font:          "basic",
			Color:         "blue",
			Width:         m.width,
			Justification: 0,
		}
		rendered, err := banner.Render(cfg)
		if err != nil {
			rendered = "test-app\n"
		}
		m.bannerStr = rendered

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		if m.phase == phaseSelect {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			case "enter", "space":
				m.selected = m.options[m.cursor]
				m.phase = phaseStream
				m.pending = make([]string, len(incomingMessages))
				copy(m.pending, incomingMessages)
				return m, tickCmd()
			}
		}

	case streamMsg:
		if len(m.pending) > 0 {
			m.received = append(m.received, m.pending[0])
			m.pending = m.pending[1:]
			if len(m.pending) > 0 {
				return m, tickCmd()
			}
			m.phase = phaseDone
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A78BFA"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6EE7B7")).Bold(true)
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FCD34D")).Bold(true)
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#F9FAFB")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	streamStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#93C5FD"))
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#34D399")).Bold(true)
	boxStyle := lipgloss.NewStyle().
		Padding(2).
		Width(m.width * 80 / 100).
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("#4361EE"))

	var b strings.Builder

	b.WriteString(m.bannerStr)
	b.WriteString("\n")

	switch m.phase {
	case phaseSelect:
		b.WriteString(titleStyle.Render("What would you like to do?") + "\n\n")
		for i, opt := range m.options {
			if i == m.cursor {
				b.WriteString(cursorStyle.Render("▶ ") + activeStyle.Render(opt) + "\n")
			} else {
				b.WriteString("  " + dimStyle.Render(opt) + "\n")
			}
		}
		b.WriteString("\n" + dimStyle.Render("↑/↓ to navigate  •  enter to select  •  q to quit") + "\n")

	case phaseStream:
		b.WriteString(titleStyle.Render("Option selected: ") + labelStyle.Render(m.selected) + "\n\n")
		for _, line := range m.received {
			b.WriteString(streamStyle.Render("  › "+line) + "\n")
		}
		b.WriteString(dimStyle.Render("  …") + "\n")

	case phaseDone:
		b.WriteString(titleStyle.Render("Option selected: ") + labelStyle.Render(m.selected) + "\n\n")
		for _, line := range m.received {
			b.WriteString(streamStyle.Render("  › "+line) + "\n")
		}
		b.WriteString("\n" + doneStyle.Render("  ✓ Stream complete") + "\n")
		b.WriteString(dimStyle.Render("\nq to quit") + "\n")
	}

	return tea.NewView(boxStyle.Render(b.String()))
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI",
	Long:  `Launch the interactive terminal user interface with menu selection and streaming display.`,
	Run: func(cmd *cobra.Command, args []string) {
		m := model{
			phase: phaseSelect,
			options: []string{
				"Fetch user data",
				"Run diagnostics",
				"Sync configuration",
				"Generate report",
			},
		}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

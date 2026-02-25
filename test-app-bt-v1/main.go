package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ralphio/test-app-3/banner"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.Color("170")).
			Bold(true)

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("170"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginBottom(1)

	// stream log level badge styles
	badgeInfo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75")).   // sky blue
			Width(7).Align(lipgloss.Right)

	badgeOK = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82")).   // green
		Width(7).Align(lipgloss.Right)

	badgeWarn = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")). // amber
			Width(7).Align(lipgloss.Right)

	badgeErr = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196")). // red
			Width(7).Align(lipgloss.Right)

	tsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	msgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("222")).
			Bold(true)

	streamDoneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82")).
			MarginTop(1)
)

// ── Stream script ─────────────────────────────────────────────────────────────

type logLevel int

const (
	lvlInfo logLevel = iota
	lvlOK
	lvlWarn
	lvlErr
)

type logLine struct {
	level   logLevel
	text    string
	delay   time.Duration // pause before this line is sent
}

func scriptFor(choice string) []logLine {
	scripts := map[string][]logLine{
		"Ramen": {
			{lvlInfo, "Order received: Ramen", 80 * time.Millisecond},
			{lvlInfo, "Locating nearest restaurant...", 200 * time.Millisecond},
			{lvlOK, "Found \"Tokyo Ramen House\" (0.3 mi)", 300 * time.Millisecond},
			{lvlInfo, "Sending order to kitchen...", 150 * time.Millisecond},
			{lvlWarn, "Kitchen busy — estimated wait: 12 min", 400 * time.Millisecond},
			{lvlInfo, "Chef Kenji accepted your order", 500 * time.Millisecond},
			{lvlInfo, "Preparing tonkotsu broth...", 600 * time.Millisecond},
			{lvlInfo, "Noodles selected: hakata-style thin", 300 * time.Millisecond},
			{lvlOK, "Broth temperature: 92°C ✓", 700 * time.Millisecond},
			{lvlInfo, "Adding toppings: chashu, nori, soft egg", 400 * time.Millisecond},
			{lvlWarn, "Low stock on soft-boiled eggs — substituting jammy egg", 300 * time.Millisecond},
			{lvlOK, "Toppings applied", 500 * time.Millisecond},
			{lvlInfo, "Packaging order...", 300 * time.Millisecond},
			{lvlOK, "Order ready! Driver en route...", 600 * time.Millisecond},
			{lvlOK, "Delivered. Enjoy your Ramen! 🍜", 400 * time.Millisecond},
		},
		"Sushi": {
			{lvlInfo, "Order received: Sushi", 80 * time.Millisecond},
			{lvlInfo, "Checking fish freshness index...", 250 * time.Millisecond},
			{lvlOK, "Freshness: 98/100 — excellent", 300 * time.Millisecond},
			{lvlInfo, "Itamae-san selecting today's catch...", 400 * time.Millisecond},
			{lvlOK, "Bluefin tuna: available", 200 * time.Millisecond},
			{lvlOK, "Yellowtail: available", 150 * time.Millisecond},
			{lvlWarn, "Uni: seasonal — limited quantity", 200 * time.Millisecond},
			{lvlInfo, "Pressing shari rice...", 600 * time.Millisecond},
			{lvlInfo, "Slicing nigiri: 8 pieces", 500 * time.Millisecond},
			{lvlOK, "Wasabi applied (house-grown)", 300 * time.Millisecond},
			{lvlInfo, "Plating on cedar board...", 400 * time.Millisecond},
			{lvlOK, "Ready. Itadakimasu! 🍣", 350 * time.Millisecond},
		},
		"Pizza": {
			{lvlInfo, "Order received: Pizza", 80 * time.Millisecond},
			{lvlInfo, "Preheating stone oven to 485°C...", 500 * time.Millisecond},
			{lvlOK, "Oven ready", 800 * time.Millisecond},
			{lvlInfo, "Proofing dough: 72-hour cold ferment", 300 * time.Millisecond},
			{lvlInfo, "Stretching dough — Neapolitan style", 400 * time.Millisecond},
			{lvlWarn, "Dough tore slightly — patching", 250 * time.Millisecond},
			{lvlOK, "Dough recovered", 200 * time.Millisecond},
			{lvlInfo, "Applying San Marzano tomato base...", 300 * time.Millisecond},
			{lvlInfo, "Adding fresh mozzarella di bufala...", 300 * time.Millisecond},
			{lvlInfo, "Into the oven: 90 seconds", 1200 * time.Millisecond},
			{lvlOK, "Leopard char achieved ✓", 200 * time.Millisecond},
			{lvlInfo, "Finishing with fresh basil + EVO drizzle", 300 * time.Millisecond},
			{lvlOK, "Pizza is served. Buon appetito! 🍕", 350 * time.Millisecond},
		},
		"Tacos": {
			{lvlInfo, "Order received: Tacos", 80 * time.Millisecond},
			{lvlInfo, "Sourcing corn tortillas — masa harina batch", 300 * time.Millisecond},
			{lvlOK, "Tortillas fresh from comal", 400 * time.Millisecond},
			{lvlInfo, "Marinating al pastor pork...", 500 * time.Millisecond},
			{lvlInfo, "Grilling on trompo at 260°C...", 700 * time.Millisecond},
			{lvlOK, "Caramelization achieved", 300 * time.Millisecond},
			{lvlInfo, "Slicing meat to order...", 400 * time.Millisecond},
			{lvlInfo, "Adding pineapple, onion, cilantro...", 300 * time.Millisecond},
			{lvlWarn, "Salsa verde running low — refilling", 200 * time.Millisecond},
			{lvlOK, "Salsa verde: restocked", 150 * time.Millisecond},
			{lvlInfo, "Lime wedges and radishes on the side", 200 * time.Millisecond},
			{lvlOK, "¡Buen provecho! 🌮", 350 * time.Millisecond},
		},
		"Salad": {
			{lvlInfo, "Order received: Salad", 80 * time.Millisecond},
			{lvlInfo, "Checking garden bed sensors...", 250 * time.Millisecond},
			{lvlOK, "Soil moisture: optimal", 200 * time.Millisecond},
			{lvlInfo, "Harvesting arugula + mixed greens...", 400 * time.Millisecond},
			{lvlInfo, "Washing with triple rinse protocol...", 500 * time.Millisecond},
			{lvlOK, "Bacteria scan: clean", 300 * time.Millisecond},
			{lvlInfo, "Spinning dry...", 300 * time.Millisecond},
			{lvlInfo, "Toasting pine nuts at 160°C...", 600 * time.Millisecond},
			{lvlWarn, "Pine nuts slightly over-toasted — acceptable", 200 * time.Millisecond},
			{lvlInfo, "Shaving parmesan...", 300 * time.Millisecond},
			{lvlInfo, "Emulsifying lemon-tahini dressing...", 400 * time.Millisecond},
			{lvlOK, "Dressed and tossed", 200 * time.Millisecond},
			{lvlOK, "Fresh, crisp, and ready. Enjoy! 🥗", 350 * time.Millisecond},
		},
	}

	lines, ok := scripts[choice]
	if !ok {
		return []logLine{{lvlOK, "Order placed for " + choice, 200 * time.Millisecond}}
	}
	return lines
}

// ── Messages ──────────────────────────────────────────────────────────────────

type bannerRenderedMsg string
type streamLineMsg int // carries the index of the line just delivered
type streamDoneMsg struct{}

// ── Model ─────────────────────────────────────────────────────────────────────

type model struct {
	// selection phase
	title      string
	choices    []string
	cursor     int
	chosen     int
	bannerText string

	// stream phase
	streaming   bool
	script      []logLine
	streamIndex int
	received    []string // rendered log lines accumulated so far
	streamDone  bool
}

func initialModel() model {
	return model{
		title: "What would you like for lunch?",
		choices: []string{
			"Ramen",
			"Sushi",
			"Pizza",
			"Tacos",
			"Salad",
		},
		cursor: 0,
		chosen: -1,
	}
}

// ── Commands ──────────────────────────────────────────────────────────────────

func renderBanner() tea.Msg {
	cfg := banner.RandomBanner("ralphio")
	cfg.Width = 80
	result, err := banner.Render(cfg)
	if err != nil {
		return bannerRenderedMsg("ralphio\n")
	}
	return bannerRenderedMsg(result)
}

// deliverLine sleeps for the line's delay then signals its arrival.
func deliverLine(script []logLine, index int) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(script[index].delay)
		return streamLineMsg(index)
	}
}

// ── Rendering helpers ─────────────────────────────────────────────────────────

var startTime = time.Now()

func renderLogLine(line logLine) string {
	ts := tsStyle.Render(fmt.Sprintf("[%6.2fs]", time.Since(startTime).Seconds()))

	var badge string
	switch line.level {
	case lvlOK:
		badge = badgeOK.Render("OK")
	case lvlWarn:
		badge = badgeWarn.Render("WARN")
	case lvlErr:
		badge = badgeErr.Render("ERR")
	default:
		badge = badgeInfo.Render("INFO")
	}

	// highlight quoted strings within the message
	text := line.text
	styled := highlightQuotes(text)

	return fmt.Sprintf("%s %s  %s", ts, badge, styled)
}

// highlightQuotes wraps "quoted" substrings in highlightStyle.
func highlightQuotes(s string) string {
	var b strings.Builder
	for {
		open := strings.Index(s, "\"")
		if open == -1 {
			b.WriteString(msgStyle.Render(s))
			break
		}
		close := strings.Index(s[open+1:], "\"")
		if close == -1 {
			b.WriteString(msgStyle.Render(s))
			break
		}
		close += open + 1
		b.WriteString(msgStyle.Render(s[:open]))
		b.WriteString(highlightStyle.Render(s[open : close+1]))
		s = s[close+1:]
	}
	return b.String()
}

// ── Init / Update / View ──────────────────────────────────────────────────────

func (m model) Init() tea.Cmd {
	return renderBanner
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case bannerRenderedMsg:
		m.bannerText = string(msg)
		return m, nil

	case streamLineMsg:
		idx := int(msg)
		line := m.script[idx]
		m.received = append(m.received, renderLogLine(line))
		m.streamIndex = idx + 1

		if m.streamIndex >= len(m.script) {
			m.streamDone = true
			return m, nil
		}
		return m, deliverLine(m.script, m.streamIndex)

	case tea.KeyMsg:
		if m.streaming {
			// only allow quit once stream is done
			if m.streamDone && (msg.String() == "q" || msg.String() == "ctrl+c") {
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.chosen = m.cursor
			m.streaming = true
			m.script = scriptFor(m.choices[m.chosen])
			startTime = time.Now()
			return m, deliverLine(m.script, 0)
		}
	}
	return m, nil
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("\n")

	if m.bannerText != "" {
		b.WriteString(m.bannerText)
		b.WriteString("\n")
	}

	// ── Selection phase ───────────────────────────────────────────────────────
	if !m.streaming {
		b.WriteString(titleStyle.Render(m.title))
		b.WriteString("\n")
		b.WriteString(subtitleStyle.Render("↑/↓ to move • enter to select • q to quit"))
		b.WriteString("\n")

		for i, choice := range m.choices {
			cursor := "  "
			if m.cursor == i {
				cursor = cursorStyle.Render("▸ ")
			}
			if m.cursor == i {
				b.WriteString(cursor + selectedStyle.Render(choice) + "\n")
			} else {
				b.WriteString(cursor + itemStyle.Render(choice) + "\n")
			}
		}
		return b.String()
	}

	// ── Stream phase ──────────────────────────────────────────────────────────
	choiceName := m.choices[m.chosen]

	b.WriteString(titleStyle.Render("Ordering " + choiceName + "..."))
	b.WriteString("\n\n")

	for _, line := range m.received {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	if m.streamDone {
		b.WriteString("\n")
		b.WriteString("  " + streamDoneStyle.Render("✓ Done — press q to exit"))
		b.WriteString("\n")
	}

	return b.String()
}

// ── Entry point ───────────────────────────────────────────────────────────────

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package screens

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/task"
	"scaffold/internal/ui/spinner"
	"scaffold/internal/ui/theme"
)

// Detail is a detail screen that shows information about a selected menu item.
// It demonstrates the async task + spinner pattern: content is "loaded" via
// task.RunWithTimeout, with a spinner displayed while the task runs.
// A tea.Tick command (§7C) counts elapsed seconds during loading.
type Detail struct {
	theme.ThemeAware

	ctx         context.Context
	title       string
	description string
	screenID    string
	width       int
	load        spinner.Loading
	elapsed     int // seconds elapsed since loading started
	styles      theme.DetailStyles
}

// NewDetail creates a new Detail screen. ctx is used to cancel the load task
// if the user navigates away or quits before it completes.
func NewDetail(title, description, screenID string, ctx context.Context) *Detail {
	return &Detail{
		ctx:         ctx,
		title:       title,
		description: description,
		screenID:    screenID,
		load:        spinner.NewLoading(theme.Palette{}),
	}
}

// SetWidth sets the screen width.
func (d *Detail) SetWidth(w int) Screen {
	d.width = w
	return d
}

// ApplyTheme implements theme.Themeable.
func (d *Detail) ApplyTheme(state theme.State) {
	d.ApplyThemeState(state)
	d.load.ApplyPalette(state.Palette)
	d.styles = theme.NewDetailStylesFromPalette(state.Palette)
}

// tickCmd returns a command that fires detailTickMsg after one second,
// demonstrating the canonical periodic-task pattern with tea.Tick.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return detailTickMsg(t)
	})
}

// Init starts the simulated background load, the spinner tick loop, and the
// elapsed-time ticker that counts seconds while loading is active.
func (d *Detail) Init() tea.Cmd {
	return tea.Batch(
		d.load.Start(),
		tickCmd(),
		task.Run(d.ctx, "detail-load",
			func(ctx context.Context) (string, error) {
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(1500 * time.Millisecond):
					return "loaded", nil
				}
			},
		),
	)
}

// Update handles messages for the detail screen.
func (d *Detail) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Resolve task results first so we catch them even if loading changes state.
	switch msg := msg.(type) {
	case task.DoneMsg[string]:
		if msg.Label == "detail-load" {
			d.load.Stop()
			return d, nil
		}
	case task.ErrMsg:
		if msg.Label == "detail-load" {
			d.load.Stop()
			return d, nil
		}
	case detailTickMsg:
		// Advance elapsed counter and reschedule while loading is active.
		if d.load.Active() {
			d.elapsed++
			return d, tickCmd()
		}
		return d, nil
	}

	// While loading, advance the spinner on every message.
	if d.load.Active() {
		var cmd tea.Cmd
		d.load, cmd = d.load.Update(msg)
		return d, cmd
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.String() {
		case "esc":
			return d, func() tea.Msg { return BackMsg{} }
		}
	}
	return d, nil
}

// View renders the detail screen.
func (d *Detail) View() tea.View {
	return tea.NewView(d.Body())
}

// Body returns the body content for layout composition.
func (d *Detail) Body() string {
	if d.load.Active() {
		label := fmt.Sprintf("Loading… %ds", d.elapsed)
		return d.load.View(label, d.Palette())
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		d.styles.Title.Render(d.title),
		d.styles.Desc.Render(d.description),
		d.styles.Content.Render(fmt.Sprintf("Screen ID: %s", d.screenID)),
		"",
		d.styles.Info.Render("Press Esc to go back to the menu"),
	)

	return content
}

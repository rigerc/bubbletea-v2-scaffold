package modal

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"scaffold/internal/ui/theme"
)

// Kind controls which buttons and actions are available.
type Kind int

const (
	KindConfirm Kind = iota // Yes / No
	KindAlert               // OK only
	KindPrompt              // single-line text input + Submit / Cancel
)

type keyMap struct {
	Confirm key.Binding
	Cancel  key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Confirm: key.NewBinding(
			key.WithKeys("y", "Y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "N", "esc"),
			key.WithHelp("n/esc", "cancel"),
		),
	}
}

// Model is a self-contained modal dialog rendered by rootModel over the
// current screen. The zero value is invisible (Visible() returns false).
type Model struct {
	id      string
	kind    Kind
	title   string
	body    string
	input   textinput.Model
	visible bool
	keys    keyMap
	styles  theme.ModalStyles
}

// New creates a visible modal from a ShowMsg.
func New(msg ShowMsg, p theme.Palette) Model {
	m := Model{
		id:      msg.ID,
		kind:    msg.Kind,
		title:   msg.Title,
		body:    msg.Body,
		visible: true,
		keys:    defaultKeyMap(),
		styles:  theme.NewModalStylesFromPalette(p),
	}
	if msg.Kind == KindPrompt {
		ti := textinput.New()
		ti.Focus()
		m.input = ti
	}
	return m
}

// Visible reports whether the modal is currently displayed.
func (m Model) Visible() bool { return m.visible }

// Update handles key presses, routing to ConfirmedMsg, CancelledMsg, or
// PromptSubmittedMsg depending on the modal Kind.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch m.kind {
		case KindConfirm:
			if key.Matches(keyMsg, m.keys.Confirm) {
				m.visible = false
				id := m.id
				return m, func() tea.Msg { return ConfirmedMsg{ID: id} }
			}
			if key.Matches(keyMsg, m.keys.Cancel) {
				m.visible = false
				id := m.id
				return m, func() tea.Msg { return CancelledMsg{ID: id} }
			}
		case KindAlert:
			// Any confirm or cancel key dismisses the alert
			if key.Matches(keyMsg, m.keys.Confirm) || key.Matches(keyMsg, m.keys.Cancel) {
				m.visible = false
				id := m.id
				return m, func() tea.Msg { return CancelledMsg{ID: id} }
			}
		case KindPrompt:
			if key.Matches(keyMsg, m.keys.Cancel) {
				m.visible = false
				id := m.id
				return m, func() tea.Msg { return CancelledMsg{ID: id} }
			}
			if keyMsg.String() == "enter" {
				val := m.input.Value()
				m.visible = false
				id := m.id
				return m, func() tea.Msg { return PromptSubmittedMsg{ID: id, Value: val} }
			}
		}
	}

	if m.kind == KindPrompt {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

// View renders the dialog box.
func (m Model) View() string {
	var rows []string
	rows = append(rows, m.styles.Title.Render(m.title))
	if m.body != "" {
		rows = append(rows, "")
		rows = append(rows, m.styles.Body.Render(m.body))
	}
	rows = append(rows, "")

	switch m.kind {
	case KindConfirm:
		rows = append(rows, m.styles.Hint.Render("[y] Yes   [n] No"))
	case KindAlert:
		rows = append(rows, m.styles.Hint.Render("[enter] OK"))
	case KindPrompt:
		rows = append(rows, m.input.View())
		rows = append(rows, "")
		rows = append(rows, m.styles.Hint.Render("[enter] Submit   [esc] Cancel"))
	}

	inner := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return m.styles.Dialog.Render(inner)
}

// ShowConfirm returns a Cmd that triggers a confirm (Yes/No) modal.
func ShowConfirm(id, title, body string) tea.Cmd {
	return func() tea.Msg {
		return ShowMsg{ID: id, Kind: KindConfirm, Title: title, Body: body}
	}
}

// ShowAlert returns a Cmd that triggers an alert (OK) modal.
func ShowAlert(id, title, body string) tea.Cmd {
	return func() tea.Msg {
		return ShowMsg{ID: id, Kind: KindAlert, Title: title, Body: body}
	}
}

// ShowPrompt returns a Cmd that triggers a text-input prompt modal.
func ShowPrompt(id, title, body string) tea.Cmd {
	return func() tea.Msg {
		return ShowMsg{ID: id, Kind: KindPrompt, Title: title, Body: body}
	}
}

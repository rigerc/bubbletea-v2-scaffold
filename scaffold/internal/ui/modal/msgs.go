// Package modal provides a reusable overlay dialog for confirmations, alerts,
// and single-line text prompts rendered on top of the current screen.
package modal

// ShowMsg is dispatched via tea.Cmd to display a modal dialog.
type ShowMsg struct {
	ID    string
	Kind  Kind
	Title string
	Body  string
}

// ConfirmedMsg is sent when the user accepts a KindConfirm modal.
type ConfirmedMsg struct{ ID string }

// CancelledMsg is sent when the user dismisses any modal.
type CancelledMsg struct{ ID string }

// PromptSubmittedMsg is sent when the user submits a KindPrompt modal.
type PromptSubmittedMsg struct {
	ID    string
	Value string
}

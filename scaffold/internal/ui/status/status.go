// Package status provides a typed, theme-aware status message system for TUIs.
package status

import "time"

// Kind represents the type of status message.
type Kind int

const (
	// KindNone is the default/reset state.
	KindNone Kind = iota
	// KindInfo is for general informational messages.
	KindInfo
	// KindSuccess is for successful operation messages.
	KindSuccess
	// KindWarning is for warning condition messages.
	KindWarning
	// KindError is for error condition messages.
	KindError
)

// Msg is a message to update the footer status with a typed message.
type Msg struct {
	Text     string
	Kind     Kind
	Duration time.Duration // 0 = persistent until cleared
}

// ClearMsg is sent to reset the footer to default state.
type ClearMsg struct{}

// State holds the current status state for rendering.
type State struct {
	Text string
	Kind Kind
}
